package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateQARun(ctx context.Context, projectID string, input QARunRequest) (*QARun, error) {
	run, err := scanQARun(s.db.QueryRow(ctx, `
INSERT INTO qa_runs (id, project_id, status, mode, credential_profile_id, api_auth_profile_id)
VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, '')::uuid)
RETURNING id, project_id, status, mode, discovery_run_id::text, quality_check_run_id::text, api_smoke_run_id::text, test_plan_id::text,
	test_plan_execution_id::text, credential_profile_id::text, api_auth_profile_id::text, error_message, summary_json,
	started_at, completed_at, created_at, updated_at
`, uuid.NewString(), projectID, StatusQueued, input.Mode, input.CredentialProfileID, input.APIAuthProfileID))
	if err != nil {
		return nil, fmt.Errorf("insert QA run: %w", err)
	}
	return &run, nil
}

func (s *Store) ListQARuns(ctx context.Context, projectID string) ([]QARun, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, status, mode, discovery_run_id::text, quality_check_run_id::text, api_smoke_run_id::text, test_plan_id::text,
	test_plan_execution_id::text, credential_profile_id::text, api_auth_profile_id::text, error_message, summary_json,
	started_at, completed_at, created_at, updated_at
FROM qa_runs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query QA runs: %w", err)
	}
	defer rows.Close()

	runs := make([]QARun, 0)
	for rows.Next() {
		run, err := scanQARun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate QA runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetQARun(ctx context.Context, id string) (*QARun, error) {
	run, err := scanQARun(s.db.QueryRow(ctx, `
SELECT id, project_id, status, mode, discovery_run_id::text, quality_check_run_id::text, api_smoke_run_id::text, test_plan_id::text,
	test_plan_execution_id::text, credential_profile_id::text, api_auth_profile_id::text, error_message, summary_json,
	started_at, completed_at, created_at, updated_at
FROM qa_runs
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get QA run: %w", err)
	}
	return &run, nil
}

func (s *Store) UpdateQARunStatus(ctx context.Context, id string, status string) (*QARun, error) {
	run, err := scanQARun(s.db.QueryRow(ctx, `
UPDATE qa_runs
SET status = $2,
	started_at = COALESCE(started_at, now()),
	updated_at = now()
WHERE id = $1
RETURNING id, project_id, status, mode, discovery_run_id::text, quality_check_run_id::text, api_smoke_run_id::text, test_plan_id::text,
	test_plan_execution_id::text, credential_profile_id::text, api_auth_profile_id::text, error_message, summary_json,
	started_at, completed_at, created_at, updated_at
`, id, status))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update QA run status: %w", err)
	}
	return &run, nil
}

func (s *Store) AttachQARunDiscovery(ctx context.Context, id string, discoveryRunID string) (*QARun, error) {
	return s.updateQARunLink(ctx, id, "discovery_run_id", discoveryRunID)
}

func (s *Store) AttachQARunQualityCheck(ctx context.Context, id string, qualityCheckRunID string) (*QARun, error) {
	return s.updateQARunLink(ctx, id, "quality_check_run_id", qualityCheckRunID)
}

func (s *Store) AttachQARunAPISmoke(ctx context.Context, id string, apiSmokeRunID string) (*QARun, error) {
	return s.updateQARunLink(ctx, id, "api_smoke_run_id", apiSmokeRunID)
}

func (s *Store) AttachQARunTestPlan(ctx context.Context, id string, testPlanID string) (*QARun, error) {
	return s.updateQARunLink(ctx, id, "test_plan_id", testPlanID)
}

func (s *Store) AttachQARunExecution(ctx context.Context, id string, executionID string) (*QARun, error) {
	return s.updateQARunLink(ctx, id, "test_plan_execution_id", executionID)
}

func (s *Store) updateQARunLink(ctx context.Context, id string, column string, value string) (*QARun, error) {
	run, err := scanQARun(s.db.QueryRow(ctx, fmt.Sprintf(`
UPDATE qa_runs
SET %s = NULLIF($2, '')::uuid, updated_at = now()
WHERE id = $1
RETURNING id, project_id, status, mode, discovery_run_id::text, quality_check_run_id::text, api_smoke_run_id::text, test_plan_id::text,
	test_plan_execution_id::text, credential_profile_id::text, api_auth_profile_id::text, error_message, summary_json,
	started_at, completed_at, created_at, updated_at
`, column), id, value))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update QA run link: %w", err)
	}
	return &run, nil
}

func (s *Store) CompleteQARun(ctx context.Context, id string, summary map[string]any) (*QARun, error) {
	return s.finishQARun(ctx, id, StatusCompleted, "", summary)
}

func (s *Store) FailQARun(ctx context.Context, id string, message string, summary map[string]any) (*QARun, error) {
	return s.finishQARun(ctx, id, StatusFailed, RedactSecrets(message), summary)
}

func (s *Store) finishQARun(ctx context.Context, id string, status string, message string, summary map[string]any) (*QARun, error) {
	if summary == nil {
		summary = map[string]any{}
	}
	rawSummary, err := json.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("marshal QA run summary: %w", err)
	}
	run, err := scanQARun(s.db.QueryRow(ctx, `
UPDATE qa_runs
SET status = $2,
	error_message = $3,
	summary_json = $4,
	started_at = COALESCE(started_at, now()),
	completed_at = now(),
	updated_at = now()
WHERE id = $1
RETURNING id, project_id, status, mode, discovery_run_id::text, quality_check_run_id::text, api_smoke_run_id::text, test_plan_id::text,
	test_plan_execution_id::text, credential_profile_id::text, api_auth_profile_id::text, error_message, summary_json,
	started_at, completed_at, created_at, updated_at
`, id, status, message, rawSummary))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("finish QA run: %w", err)
	}
	return &run, nil
}

func scanQARun(row scanRow) (QARun, error) {
	var run QARun
	var discoveryRunID sql.NullString
	var qualityCheckRunID sql.NullString
	var apiSmokeRunID sql.NullString
	var testPlanID sql.NullString
	var executionID sql.NullString
	var credentialProfileID sql.NullString
	var apiAuthProfileID sql.NullString
	var summaryRaw []byte
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&run.Status,
		&run.Mode,
		&discoveryRunID,
		&qualityCheckRunID,
		&apiSmokeRunID,
		&testPlanID,
		&executionID,
		&credentialProfileID,
		&apiAuthProfileID,
		&run.ErrorMessage,
		&summaryRaw,
		&run.StartedAt,
		&run.CompletedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return QARun{}, fmt.Errorf("scan QA run: %w", err)
	}
	if discoveryRunID.Valid {
		run.DiscoveryRunID = discoveryRunID.String
	}
	if qualityCheckRunID.Valid {
		run.QualityCheckRunID = qualityCheckRunID.String
	}
	if apiSmokeRunID.Valid {
		run.APISmokeRunID = apiSmokeRunID.String
	}
	if testPlanID.Valid {
		run.TestPlanID = testPlanID.String
	}
	if executionID.Valid {
		run.TestPlanExecutionID = executionID.String
	}
	if credentialProfileID.Valid {
		run.CredentialProfileID = credentialProfileID.String
	}
	if apiAuthProfileID.Valid {
		run.APIAuthProfileID = apiAuthProfileID.String
	}
	run.Summary = map[string]any{}
	if len(summaryRaw) > 0 {
		if err := json.Unmarshal(summaryRaw, &run.Summary); err != nil {
			return QARun{}, fmt.Errorf("unmarshal QA run summary: %w", err)
		}
	}
	return run, nil
}

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

type CompleteCIRunInput struct {
	QARunID           string
	BaselineID        string
	Status            string
	ExitCode          int
	GateStatus        string
	ComparisonStatus  string
	ReportURL         string
	HTMLReportURL     string
	IssueExportStatus string
	Summary           map[string]any
}

func (s *Store) CreateCIRun(ctx context.Context, projectID string, summary map[string]any) (*CIRun, error) {
	if summary == nil {
		summary = map[string]any{}
	}
	rawSummary, err := json.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("marshal CI run summary: %w", err)
	}
	run, err := scanCIRun(s.db.QueryRow(ctx, `
INSERT INTO ci_runs (id, project_id, status, exit_code, summary_json, started_at)
VALUES ($1, $2, $3, $4, $5, now())
RETURNING id, project_id, qa_run_id::text, baseline_id::text, status, exit_code, gate_status,
	comparison_status, report_url, html_report_url, issue_export_status, summary_json,
	started_at, completed_at, created_at, updated_at, error_message
`, uuid.NewString(), projectID, CIRunStatusRunning, 1, rawSummary))
	if err != nil {
		return nil, fmt.Errorf("insert CI run: %w", err)
	}
	return &run, nil
}

func (s *Store) CompleteCIRun(ctx context.Context, id string, input CompleteCIRunInput) (*CIRun, error) {
	if input.Summary == nil {
		input.Summary = map[string]any{}
	}
	rawSummary, err := json.Marshal(input.Summary)
	if err != nil {
		return nil, fmt.Errorf("marshal CI run summary: %w", err)
	}
	run, err := scanCIRun(s.db.QueryRow(ctx, `
UPDATE ci_runs
SET qa_run_id = NULLIF($2, '')::uuid,
	baseline_id = NULLIF($3, '')::uuid,
	status = $4,
	exit_code = $5,
	gate_status = $6,
	comparison_status = $7,
	report_url = $8,
	html_report_url = $9,
	issue_export_status = $10,
	summary_json = $11,
	completed_at = now(),
	updated_at = now(),
	error_message = ''
WHERE id = $1
RETURNING id, project_id, qa_run_id::text, baseline_id::text, status, exit_code, gate_status,
	comparison_status, report_url, html_report_url, issue_export_status, summary_json,
	started_at, completed_at, created_at, updated_at, error_message
`, id, input.QARunID, input.BaselineID, input.Status, input.ExitCode, input.GateStatus, input.ComparisonStatus, input.ReportURL, input.HTMLReportURL, input.IssueExportStatus, rawSummary))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("complete CI run: %w", err)
	}
	return &run, nil
}

func (s *Store) FailCIRun(ctx context.Context, id string, message string, summary map[string]any) (*CIRun, error) {
	if summary == nil {
		summary = map[string]any{}
	}
	message = RedactSecrets(message)
	rawSummary, err := json.Marshal(summary)
	if err != nil {
		return nil, fmt.Errorf("marshal CI run summary: %w", err)
	}
	run, err := scanCIRun(s.db.QueryRow(ctx, `
UPDATE ci_runs
SET status = $2,
	exit_code = 1,
	summary_json = $3,
	completed_at = now(),
	updated_at = now(),
	error_message = $4
WHERE id = $1
RETURNING id, project_id, qa_run_id::text, baseline_id::text, status, exit_code, gate_status,
	comparison_status, report_url, html_report_url, issue_export_status, summary_json,
	started_at, completed_at, created_at, updated_at, error_message
`, id, CIRunStatusError, rawSummary, message))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fail CI run: %w", err)
	}
	return &run, nil
}

func (s *Store) GetCIRun(ctx context.Context, id string) (*CIRun, error) {
	run, err := scanCIRun(s.db.QueryRow(ctx, `
SELECT id, project_id, qa_run_id::text, baseline_id::text, status, exit_code, gate_status,
	comparison_status, report_url, html_report_url, issue_export_status, summary_json,
	started_at, completed_at, created_at, updated_at, error_message
FROM ci_runs
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (s *Store) ListCIRuns(ctx context.Context, projectID string) ([]CIRun, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, qa_run_id::text, baseline_id::text, status, exit_code, gate_status,
	comparison_status, report_url, html_report_url, issue_export_status, summary_json,
	started_at, completed_at, created_at, updated_at, error_message
FROM ci_runs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query CI runs: %w", err)
	}
	defer rows.Close()
	runs := []CIRun{}
	for rows.Next() {
		run, err := scanCIRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate CI runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetLatestCompletedQARun(ctx context.Context, projectID string) (*QARun, error) {
	run, err := scanQARun(s.db.QueryRow(ctx, `
SELECT id, project_id, status, mode, discovery_run_id::text, quality_check_run_id::text, api_smoke_run_id::text, test_plan_id::text,
	test_plan_execution_id::text, credential_profile_id::text, api_auth_profile_id::text, error_message, summary_json,
	started_at, completed_at, created_at, updated_at
FROM qa_runs
WHERE project_id = $1 AND status = $2
ORDER BY created_at DESC
LIMIT 1
`, projectID, StatusCompleted))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get latest completed QA run: %w", err)
	}
	return &run, nil
}

func scanCIRun(row scanRow) (CIRun, error) {
	var run CIRun
	var qaRunID sql.NullString
	var baselineID sql.NullString
	var summaryRaw []byte
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&qaRunID,
		&baselineID,
		&run.Status,
		&run.ExitCode,
		&run.GateStatus,
		&run.ComparisonStatus,
		&run.ReportURL,
		&run.HTMLReportURL,
		&run.IssueExportStatus,
		&summaryRaw,
		&run.StartedAt,
		&run.CompletedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
		&run.ErrorMessage,
	); err != nil {
		return CIRun{}, fmt.Errorf("scan CI run: %w", err)
	}
	if qaRunID.Valid {
		run.QARunID = qaRunID.String
	}
	if baselineID.Valid {
		run.BaselineID = baselineID.String
	}
	run.Summary = map[string]any{}
	if len(summaryRaw) > 0 {
		if err := json.Unmarshal(summaryRaw, &run.Summary); err != nil {
			return CIRun{}, fmt.Errorf("unmarshal CI run summary: %w", err)
		}
	}
	return run, nil
}

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListAuthorizationChecks(ctx context.Context, projectID string) ([]AuthorizationCheck, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, name, description, type, resource_label,
	owner_credential_profile_id::text, actor_credential_profile_id::text, expected_outcome,
	target_url, api_spec_id::text, api_operation_id::text, method, path,
	expected_statuses_json, success_text_contains, denied_statuses_json, denied_text_contains,
	enabled, created_at, updated_at
FROM authorization_checks
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query authorization checks: %w", err)
	}
	defer rows.Close()

	checks := make([]AuthorizationCheck, 0)
	for rows.Next() {
		check, err := scanAuthorizationCheck(rows)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate authorization checks: %w", err)
	}
	return checks, nil
}

func (s *Store) GetAuthorizationCheck(ctx context.Context, id string) (*AuthorizationCheck, error) {
	check, err := scanAuthorizationCheck(s.db.QueryRow(ctx, `
SELECT id, project_id, name, description, type, resource_label,
	owner_credential_profile_id::text, actor_credential_profile_id::text, expected_outcome,
	target_url, api_spec_id::text, api_operation_id::text, method, path,
	expected_statuses_json, success_text_contains, denied_statuses_json, denied_text_contains,
	enabled, created_at, updated_at
FROM authorization_checks
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &check, nil
}

func (s *Store) CreateAuthorizationCheck(ctx context.Context, check AuthorizationCheck) (*AuthorizationCheck, error) {
	created, err := scanAuthorizationCheck(s.db.QueryRow(ctx, `
INSERT INTO authorization_checks (
	id, project_id, name, description, type, resource_label,
	owner_credential_profile_id, actor_credential_profile_id, expected_outcome,
	target_url, api_spec_id, api_operation_id, method, path,
	expected_statuses_json, success_text_contains, denied_statuses_json, denied_text_contains, enabled
) VALUES (
	$1, $2, $3, $4, $5, $6,
	NULLIF($7, '')::uuid, $8, $9,
	$10, NULLIF($11, '')::uuid, NULLIF($12, '')::uuid, $13, $14,
	$15, $16, $17, $18, $19
)
RETURNING id, project_id, name, description, type, resource_label,
	owner_credential_profile_id::text, actor_credential_profile_id::text, expected_outcome,
	target_url, api_spec_id::text, api_operation_id::text, method, path,
	expected_statuses_json, success_text_contains, denied_statuses_json, denied_text_contains,
	enabled, created_at, updated_at
`, uuid.NewString(), check.ProjectID, check.Name, check.Description, check.Type, check.ResourceLabel,
		check.OwnerCredentialProfileID, check.ActorCredentialProfileID, check.ExpectedOutcome,
		check.TargetURL, check.APISpecID, check.APIOperationID, check.Method, check.Path,
		mustMarshalJSON(check.ExpectedStatuses), check.SuccessTextContains, mustMarshalJSON(check.DeniedStatuses),
		check.DeniedTextContains, check.Enabled))
	if err != nil {
		return nil, fmt.Errorf("insert authorization check: %w", err)
	}
	return &created, nil
}

func (s *Store) UpdateAuthorizationCheck(ctx context.Context, id string, check AuthorizationCheck) (*AuthorizationCheck, error) {
	updated, err := scanAuthorizationCheck(s.db.QueryRow(ctx, `
UPDATE authorization_checks
SET name = $2,
	description = $3,
	type = $4,
	resource_label = $5,
	owner_credential_profile_id = NULLIF($6, '')::uuid,
	actor_credential_profile_id = $7,
	expected_outcome = $8,
	target_url = $9,
	api_spec_id = NULLIF($10, '')::uuid,
	api_operation_id = NULLIF($11, '')::uuid,
	method = $12,
	path = $13,
	expected_statuses_json = $14,
	success_text_contains = $15,
	denied_statuses_json = $16,
	denied_text_contains = $17,
	enabled = $18,
	updated_at = now()
WHERE id = $1
RETURNING id, project_id, name, description, type, resource_label,
	owner_credential_profile_id::text, actor_credential_profile_id::text, expected_outcome,
	target_url, api_spec_id::text, api_operation_id::text, method, path,
	expected_statuses_json, success_text_contains, denied_statuses_json, denied_text_contains,
	enabled, created_at, updated_at
`, id, check.Name, check.Description, check.Type, check.ResourceLabel,
		check.OwnerCredentialProfileID, check.ActorCredentialProfileID, check.ExpectedOutcome,
		check.TargetURL, check.APISpecID, check.APIOperationID, check.Method, check.Path,
		mustMarshalJSON(check.ExpectedStatuses), check.SuccessTextContains, mustMarshalJSON(check.DeniedStatuses),
		check.DeniedTextContains, check.Enabled))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update authorization check: %w", err)
	}
	return &updated, nil
}

func (s *Store) DeleteAuthorizationCheck(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM authorization_checks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete authorization check: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateAuthorizationCheckRun(ctx context.Context, projectID string, input AuthorizationCheckRunRequest) (*AuthorizationCheckRun, error) {
	selectedChecks, err := s.selectAuthorizationChecksForRun(ctx, projectID, input)
	if err != nil {
		return nil, err
	}
	if len(selectedChecks) == 0 {
		return nil, fmt.Errorf("no enabled authorization checks matched the request")
	}
	checkIDs := make([]string, 0, len(selectedChecks))
	for _, check := range selectedChecks {
		checkIDs = append(checkIDs, check.ID)
	}

	run, err := scanAuthorizationCheckRun(s.db.QueryRow(ctx, `
INSERT INTO authorization_check_runs (id, project_id, status, check_ids_json, max_checks, total_checks)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, status, check_ids_json, max_checks, total_checks, passed_checks,
	failed_checks, skipped_checks, error_message, started_at, completed_at, created_at, updated_at
`, uuid.NewString(), projectID, StatusQueued, mustMarshalJSON(checkIDs), input.MaxChecks, len(selectedChecks)))
	if err != nil {
		return nil, fmt.Errorf("insert authorization check run: %w", err)
	}
	return &run, nil
}

func (s *Store) ListAuthorizationCheckRuns(ctx context.Context, projectID string) ([]AuthorizationCheckRun, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, status, check_ids_json, max_checks, total_checks, passed_checks,
	failed_checks, skipped_checks, error_message, started_at, completed_at, created_at, updated_at
FROM authorization_check_runs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query authorization check runs: %w", err)
	}
	defer rows.Close()

	runs := make([]AuthorizationCheckRun, 0)
	for rows.Next() {
		run, err := scanAuthorizationCheckRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate authorization check runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetAuthorizationCheckRun(ctx context.Context, id string) (*AuthorizationCheckRun, error) {
	run, err := scanAuthorizationCheckRun(s.db.QueryRow(ctx, `
SELECT id, project_id, status, check_ids_json, max_checks, total_checks, passed_checks,
	failed_checks, skipped_checks, error_message, started_at, completed_at, created_at, updated_at
FROM authorization_check_runs
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

func (s *Store) GetAuthorizationCheckDetail(ctx context.Context, id string) (*AuthorizationCheckDetail, error) {
	run, err := s.GetAuthorizationCheckRun(ctx, id)
	if err != nil {
		return nil, err
	}
	results, err := s.ListAuthorizationCheckResults(ctx, id)
	if err != nil {
		return nil, err
	}
	checks, err := s.listAuthorizationChecksByIDs(ctx, run.ProjectID, run.CheckIDs)
	if err != nil {
		return nil, err
	}
	return &AuthorizationCheckDetail{Run: *run, Checks: checks, Results: results}, nil
}

func (s *Store) GetAuthorizationCheckReport(ctx context.Context, id string) (*AuthorizationCheckReport, error) {
	detail, err := s.GetAuthorizationCheckDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, detail.Run.ProjectID)
	if err != nil {
		return nil, err
	}
	findings, err := s.ListFindingsForAuthorizationCheckRun(ctx, id)
	if err != nil {
		return nil, err
	}
	evidence, err := s.ListEvidenceForAuthorizationCheckRun(ctx, id)
	if err != nil {
		return nil, err
	}
	return &AuthorizationCheckReport{
		Run:      detail.Run,
		Project:  *project,
		Checks:   detail.Checks,
		Results:  detail.Results,
		Summary:  summarizeFindings(findings),
		Findings: findings,
		Evidence: evidence,
		Metadata: map[string]any{
			"authorization_checks":   len(detail.Checks),
			"authorization_results":  len(detail.Results),
			"browser_url_checks":     countAuthorizationChecksByType(detail.Checks, AuthorizationCheckTypeBrowserURL),
			"api_get_checks_skipped": countAuthorizationChecksByType(detail.Checks, AuthorizationCheckTypeAPIGet),
			"safe_methods_only":      true,
			"destructive_actions":    false,
		},
		GeneratedAt: time.Now().UTC(),
	}, nil
}

func (s *Store) ListAuthorizationCheckResults(ctx context.Context, runID string) ([]AuthorizationCheckResult, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, check_id, status, expected_outcome, actual_outcome,
	actor_credential_profile_id::text, actor_role_name, target_url, final_url,
	http_status, page_title, duration_ms, evidence_id::text, finding_id::text,
	skip_reason, error_message, created_at
FROM authorization_check_results
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query authorization check results: %w", err)
	}
	defer rows.Close()

	results := make([]AuthorizationCheckResult, 0)
	for rows.Next() {
		result, err := scanAuthorizationCheckResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate authorization check results: %w", err)
	}
	return results, nil
}

func (s *Store) ListFindingsForAuthorizationCheckRun(ctx context.Context, runID string) ([]Finding, error) {
	rows, err := s.db.Query(ctx, `
SELECT f.id, f.run_id::text, f.test_plan_execution_id::text, f.authorization_check_run_id::text,
	f.scenario_execution_id::text, f.step_execution_id::text,
	f.title, f.severity, f.category, f.confidence, f.description, f.recommendation, f.evidence_ids, f.created_at
FROM findings f
WHERE f.authorization_check_run_id = $1
ORDER BY f.created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query authorization check findings: %w", err)
	}
	defer rows.Close()

	findings := make([]Finding, 0)
	for rows.Next() {
		finding, err := scanFinding(rows)
		if err != nil {
			return nil, err
		}
		findings = append(findings, finding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate authorization check findings: %w", err)
	}
	return findings, nil
}

func (s *Store) ListEvidenceForAuthorizationCheckRun(ctx context.Context, runID string) ([]Evidence, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	type, uri, metadata, created_at
FROM evidence
WHERE authorization_check_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query authorization check evidence: %w", err)
	}
	defer rows.Close()

	records := make([]Evidence, 0)
	for rows.Next() {
		record, err := scanEvidence(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate authorization check evidence: %w", err)
	}
	return records, nil
}

func (s *Store) MarkAuthorizationCheckRunFailed(ctx context.Context, id string, message string) error {
	tag, err := s.db.Exec(ctx, `
UPDATE authorization_check_runs
SET status = $2, error_message = $3, completed_at = COALESCE(completed_at, now()), updated_at = now()
WHERE id = $1
`, id, StatusFailed, message)
	if err != nil {
		return fmt.Errorf("mark authorization check run failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) selectAuthorizationChecksForRun(ctx context.Context, projectID string, input AuthorizationCheckRunRequest) ([]AuthorizationCheck, error) {
	allChecks, err := s.ListAuthorizationChecks(ctx, projectID)
	if err != nil {
		return nil, err
	}
	wanted := map[string]struct{}{}
	for _, id := range input.CheckIDs {
		wanted[id] = struct{}{}
	}
	selected := make([]AuthorizationCheck, 0, min(len(allChecks), input.MaxChecks))
	for _, check := range allChecks {
		if !check.Enabled {
			continue
		}
		if len(wanted) > 0 {
			if _, ok := wanted[check.ID]; !ok {
				continue
			}
		}
		selected = append(selected, check)
		if len(selected) >= input.MaxChecks {
			break
		}
	}
	return selected, nil
}

func (s *Store) listAuthorizationChecksByIDs(ctx context.Context, projectID string, ids []string) ([]AuthorizationCheck, error) {
	allChecks, err := s.ListAuthorizationChecks(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return allChecks, nil
	}
	index := map[string]int{}
	for i, id := range ids {
		index[id] = i
	}
	checks := make([]AuthorizationCheck, 0, len(ids))
	for _, check := range allChecks {
		if _, ok := index[check.ID]; ok {
			checks = append(checks, check)
		}
	}
	return checks, nil
}

func scanAuthorizationCheck(row scanRow) (AuthorizationCheck, error) {
	var check AuthorizationCheck
	var ownerID sql.NullString
	var apiSpecID sql.NullString
	var apiOperationID sql.NullString
	var expectedStatusesRaw []byte
	var deniedStatusesRaw []byte
	if err := row.Scan(
		&check.ID,
		&check.ProjectID,
		&check.Name,
		&check.Description,
		&check.Type,
		&check.ResourceLabel,
		&ownerID,
		&check.ActorCredentialProfileID,
		&check.ExpectedOutcome,
		&check.TargetURL,
		&apiSpecID,
		&apiOperationID,
		&check.Method,
		&check.Path,
		&expectedStatusesRaw,
		&check.SuccessTextContains,
		&deniedStatusesRaw,
		&check.DeniedTextContains,
		&check.Enabled,
		&check.CreatedAt,
		&check.UpdatedAt,
	); err != nil {
		return AuthorizationCheck{}, fmt.Errorf("scan authorization check: %w", err)
	}
	if ownerID.Valid {
		check.OwnerCredentialProfileID = ownerID.String
	}
	if apiSpecID.Valid {
		check.APISpecID = apiSpecID.String
	}
	if apiOperationID.Valid {
		check.APIOperationID = apiOperationID.String
	}
	_ = json.Unmarshal(expectedStatusesRaw, &check.ExpectedStatuses)
	_ = json.Unmarshal(deniedStatusesRaw, &check.DeniedStatuses)
	return check, nil
}

func scanAuthorizationCheckRun(row scanRow) (AuthorizationCheckRun, error) {
	var run AuthorizationCheckRun
	var checkIDsRaw []byte
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&run.Status,
		&checkIDsRaw,
		&run.MaxChecks,
		&run.TotalChecks,
		&run.PassedChecks,
		&run.FailedChecks,
		&run.SkippedChecks,
		&run.ErrorMessage,
		&run.StartedAt,
		&run.CompletedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return AuthorizationCheckRun{}, fmt.Errorf("scan authorization check run: %w", err)
	}
	_ = json.Unmarshal(checkIDsRaw, &run.CheckIDs)
	return run, nil
}

func scanAuthorizationCheckResult(row scanRow) (AuthorizationCheckResult, error) {
	var result AuthorizationCheckResult
	var status sql.NullInt64
	var duration sql.NullInt64
	var evidenceID sql.NullString
	var findingID sql.NullString
	if err := row.Scan(
		&result.ID,
		&result.RunID,
		&result.CheckID,
		&result.Status,
		&result.ExpectedOutcome,
		&result.ActualOutcome,
		&result.ActorCredentialProfileID,
		&result.ActorRoleName,
		&result.TargetURL,
		&result.FinalURL,
		&status,
		&result.PageTitle,
		&duration,
		&evidenceID,
		&findingID,
		&result.SkipReason,
		&result.ErrorMessage,
		&result.CreatedAt,
	); err != nil {
		return AuthorizationCheckResult{}, fmt.Errorf("scan authorization check result: %w", err)
	}
	if status.Valid {
		value := int(status.Int64)
		result.HTTPStatus = &value
	}
	if duration.Valid {
		value := int(duration.Int64)
		result.DurationMS = &value
	}
	if evidenceID.Valid {
		result.EvidenceID = evidenceID.String
	}
	if findingID.Valid {
		result.FindingID = findingID.String
	}
	return result, nil
}

func countAuthorizationChecksByType(checks []AuthorizationCheck, checkType string) int {
	count := 0
	for _, check := range checks {
		if check.Type == checkType {
			count++
		}
	}
	return count
}

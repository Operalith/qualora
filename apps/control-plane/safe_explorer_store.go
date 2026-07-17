package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateSafeExplorerRun(ctx context.Context, project Project, input SafeExplorerRunRequest) (*SafeExplorerRun, error) {
	run, err := scanSafeExplorerRun(s.db.QueryRow(ctx, `
INSERT INTO safe_explorer_runs (
	id, project_id, credential_profile_id, status, start_url, max_steps, max_depth, same_origin_only, allow_get_forms
) VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9)
RETURNING id, project_id, credential_profile_id::text, status, start_url, max_steps, max_depth,
	same_origin_only, allow_get_forms, started_at, completed_at, total_steps, total_pages_observed,
	total_actions_detected, total_actions_executed, total_actions_skipped, total_findings,
	error_message, created_at, updated_at
`, uuid.NewString(), project.ID, input.CredentialProfileID, StatusQueued, input.StartURL, input.MaxSteps, input.MaxDepth, *input.SameOriginOnly, input.AllowGetForms))
	if err != nil {
		return nil, fmt.Errorf("insert safe explorer run: %w", err)
	}
	return &run, nil
}

func (s *Store) ListSafeExplorerRuns(ctx context.Context, projectID string) ([]SafeExplorerRun, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, credential_profile_id::text, status, start_url, max_steps, max_depth,
	same_origin_only, allow_get_forms, started_at, completed_at, total_steps, total_pages_observed,
	total_actions_detected, total_actions_executed, total_actions_skipped, total_findings,
	error_message, created_at, updated_at
FROM safe_explorer_runs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query safe explorer runs: %w", err)
	}
	defer rows.Close()

	runs := make([]SafeExplorerRun, 0)
	for rows.Next() {
		run, err := scanSafeExplorerRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate safe explorer runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetSafeExplorerRun(ctx context.Context, id string) (*SafeExplorerRun, error) {
	run, err := scanSafeExplorerRun(s.db.QueryRow(ctx, `
SELECT id, project_id, credential_profile_id::text, status, start_url, max_steps, max_depth,
	same_origin_only, allow_get_forms, started_at, completed_at, total_steps, total_pages_observed,
	total_actions_detected, total_actions_executed, total_actions_skipped, total_findings,
	error_message, created_at, updated_at
FROM safe_explorer_runs
WHERE id = $1
`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get safe explorer run: %w", err)
	}
	return &run, nil
}

func (s *Store) MarkSafeExplorerRunFailed(ctx context.Context, id string, message string) error {
	tag, err := s.db.Exec(ctx, `
UPDATE safe_explorer_runs
SET status = $2, error_message = $3, completed_at = now(), updated_at = now()
WHERE id = $1
`, id, StatusFailed, message)
	if err != nil {
		return fmt.Errorf("mark safe explorer run failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetSafeExplorerTrace(ctx context.Context, id string) (*SafeExplorerTrace, error) {
	run, err := s.GetSafeExplorerRun(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, run.ProjectID)
	if err != nil {
		return nil, err
	}
	steps, err := s.ListSafeExplorerSteps(ctx, id)
	if err != nil {
		return nil, err
	}
	actions, err := s.ListSafeExplorerActions(ctx, id)
	if err != nil {
		return nil, err
	}
	findings, err := s.ListFindingsForSafeExplorerRun(ctx, id)
	if err != nil {
		return nil, err
	}
	evidence, err := s.ListEvidenceForSafeExplorerRun(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SafeExplorerTrace{
		Run:      *run,
		Project:  *project,
		Summary:  summarizeSafeExplorer(*run, steps, actions, findings),
		Steps:    steps,
		Actions:  actions,
		Findings: findings,
		Evidence: evidence,
	}, nil
}

func (s *Store) GetSafeExplorerReport(ctx context.Context, id string) (*SafeExplorerReport, error) {
	trace, err := s.GetSafeExplorerTrace(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SafeExplorerReport{
		GeneratedAt: time.Now().UTC(),
		Run:         trace.Run,
		Project:     trace.Project,
		Settings:    sanitizeSafeExplorerSettings(trace.Run),
		Summary:     trace.Summary,
		Steps:       trace.Steps,
		Actions:     trace.Actions,
		Findings:    trace.Findings,
		Evidence:    trace.Evidence,
		SafetyNotes: safeExplorerSafetyNotes(),
		Limitations: safeExplorerLimitations(),
		Metadata: map[string]any{
			"run_type":                      RunTypeSafeExplorer,
			"safe_explorer_schema_version":  "v1",
			"safe_actions_only":             true,
			"forms_submitted":               false,
			"destructive_actions":           false,
			"autonomous_ai_browser_control": false,
			"credentials_sent_to_ai":        false,
			"browser_storage_exposed_to_ai": false,
		},
	}, nil
}

func (s *Store) ListSafeExplorerSteps(ctx context.Context, runID string) ([]SafeExplorerStep, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, project_id, step_index, page_url, normalized_url, page_title, depth,
	action_id::text, action_type, action_label, action_selector_hint, action_target_url,
	action_safety, action_decision, skip_reason, result_status, http_status, final_url,
	screenshot_evidence_id::text, console_error_count, failed_request_count, duration_ms, created_at
FROM safe_explorer_steps
WHERE run_id = $1
ORDER BY step_index ASC, created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query safe explorer steps: %w", err)
	}
	defer rows.Close()

	steps := make([]SafeExplorerStep, 0)
	for rows.Next() {
		step, err := scanSafeExplorerStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate safe explorer steps: %w", err)
	}
	return steps, nil
}

func (s *Store) ListSafeExplorerActions(ctx context.Context, runID string) ([]SafeExplorerAction, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, step_id, source_url, action_type, label, text, selector_hint, href,
	target_url, method, same_origin, safety, decision, skip_reason, created_at
FROM safe_explorer_actions
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query safe explorer actions: %w", err)
	}
	defer rows.Close()

	actions := make([]SafeExplorerAction, 0)
	for rows.Next() {
		action, err := scanSafeExplorerAction(rows)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate safe explorer actions: %w", err)
	}
	return actions, nil
}

func (s *Store) ListFindingsForSafeExplorerRun(ctx context.Context, runID string) ([]Finding, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, scenario_execution_id::text, step_execution_id::text,
	title, severity, category, confidence, description, recommendation, evidence_ids, created_at
FROM findings
WHERE safe_explorer_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query safe explorer findings: %w", err)
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
		return nil, fmt.Errorf("iterate safe explorer findings: %w", err)
	}
	return findings, nil
}

func (s *Store) ListEvidenceForSafeExplorerRun(ctx context.Context, runID string) ([]Evidence, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, type, uri, metadata, created_at
FROM evidence
WHERE safe_explorer_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query safe explorer evidence: %w", err)
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
		return nil, fmt.Errorf("iterate safe explorer evidence: %w", err)
	}
	return records, nil
}

func scanSafeExplorerRun(row scanRow) (SafeExplorerRun, error) {
	var run SafeExplorerRun
	var credentialProfileID sql.NullString
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&credentialProfileID,
		&run.Status,
		&run.StartURL,
		&run.MaxSteps,
		&run.MaxDepth,
		&run.SameOriginOnly,
		&run.AllowGetForms,
		&run.StartedAt,
		&run.CompletedAt,
		&run.TotalSteps,
		&run.TotalPagesObserved,
		&run.TotalActionsDetected,
		&run.TotalActionsExecuted,
		&run.TotalActionsSkipped,
		&run.TotalFindings,
		&run.ErrorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return SafeExplorerRun{}, fmt.Errorf("scan safe explorer run: %w", err)
	}
	if credentialProfileID.Valid {
		run.CredentialProfileID = credentialProfileID.String
	}
	return run, nil
}

func scanSafeExplorerStep(row scanRow) (SafeExplorerStep, error) {
	var step SafeExplorerStep
	var actionID sql.NullString
	var httpStatus sql.NullInt32
	var screenshotEvidenceID sql.NullString
	var durationMS sql.NullInt32
	if err := row.Scan(
		&step.ID,
		&step.RunID,
		&step.ProjectID,
		&step.StepIndex,
		&step.PageURL,
		&step.NormalizedURL,
		&step.PageTitle,
		&step.Depth,
		&actionID,
		&step.ActionType,
		&step.ActionLabel,
		&step.ActionSelectorHint,
		&step.ActionTargetURL,
		&step.ActionSafety,
		&step.ActionDecision,
		&step.SkipReason,
		&step.ResultStatus,
		&httpStatus,
		&step.FinalURL,
		&screenshotEvidenceID,
		&step.ConsoleErrorCount,
		&step.FailedRequestCount,
		&durationMS,
		&step.CreatedAt,
	); err != nil {
		return SafeExplorerStep{}, fmt.Errorf("scan safe explorer step: %w", err)
	}
	if actionID.Valid {
		step.ActionID = actionID.String
	}
	if httpStatus.Valid {
		value := int(httpStatus.Int32)
		step.HTTPStatus = &value
	}
	if screenshotEvidenceID.Valid {
		step.ScreenshotEvidenceID = screenshotEvidenceID.String
	}
	if durationMS.Valid {
		value := int(durationMS.Int32)
		step.DurationMS = &value
	}
	return step, nil
}

func scanSafeExplorerAction(row scanRow) (SafeExplorerAction, error) {
	var action SafeExplorerAction
	if err := row.Scan(
		&action.ID,
		&action.RunID,
		&action.StepID,
		&action.SourceURL,
		&action.ActionType,
		&action.Label,
		&action.Text,
		&action.SelectorHint,
		&action.Href,
		&action.TargetURL,
		&action.Method,
		&action.SameOrigin,
		&action.Safety,
		&action.Decision,
		&action.SkipReason,
		&action.CreatedAt,
	); err != nil {
		return SafeExplorerAction{}, fmt.Errorf("scan safe explorer action: %w", err)
	}
	return action, nil
}

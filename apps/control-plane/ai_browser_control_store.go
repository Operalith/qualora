package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateAIBrowserControlRun(ctx context.Context, project Project, input AIBrowserControlRunRequest) (*AIBrowserControlRun, error) {
	id := uuid.NewString()
	if _, err := s.db.Exec(ctx, `
INSERT INTO ai_browser_control_runs (
	id, project_id, provider_id, credential_profile_id, status, start_url, goal, max_steps, max_depth, same_origin_only
) VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7, $8, $9, $10)
`, id, project.ID, input.ProviderID, input.CredentialProfileID, StatusQueued, input.StartURL, input.Goal, input.MaxSteps, input.MaxDepth, *input.SameOriginOnly); err != nil {
		return nil, fmt.Errorf("insert AI Browser Control run: %w", err)
	}
	return s.GetAIBrowserControlRun(ctx, id)
}

func (s *Store) ListAIBrowserControlRuns(ctx context.Context, projectID string) ([]AIBrowserControlRun, error) {
	rows, err := s.db.Query(ctx, `
SELECT r.id, r.project_id, r.provider_id, COALESCE(p.name, ''), r.credential_profile_id::text,
	r.status, r.start_url, r.goal, r.max_steps, r.max_depth, r.same_origin_only,
	r.policy_version, r.execution_mode,
	r.started_at, r.completed_at, r.total_steps, r.total_ai_suggestions,
	r.total_actions_approved, r.total_actions_executed, r.total_actions_skipped,
	r.total_policy_blocks, r.total_findings, r.error_message, r.created_at, r.updated_at
FROM ai_browser_control_runs r
LEFT JOIN ai_providers p ON p.id = r.provider_id
WHERE r.project_id = $1
ORDER BY r.created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query AI Browser Control runs: %w", err)
	}
	defer rows.Close()

	runs := make([]AIBrowserControlRun, 0)
	for rows.Next() {
		run, err := scanAIBrowserControlRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI Browser Control runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetAIBrowserControlRun(ctx context.Context, id string) (*AIBrowserControlRun, error) {
	run, err := scanAIBrowserControlRun(s.db.QueryRow(ctx, `
SELECT r.id, r.project_id, r.provider_id, COALESCE(p.name, ''), r.credential_profile_id::text,
	r.status, r.start_url, r.goal, r.max_steps, r.max_depth, r.same_origin_only,
	r.policy_version, r.execution_mode,
	r.started_at, r.completed_at, r.total_steps, r.total_ai_suggestions,
	r.total_actions_approved, r.total_actions_executed, r.total_actions_skipped,
	r.total_policy_blocks, r.total_findings, r.error_message, r.created_at, r.updated_at
FROM ai_browser_control_runs r
LEFT JOIN ai_providers p ON p.id = r.provider_id
WHERE r.id = $1
`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get AI Browser Control run: %w", err)
	}
	return &run, nil
}

func (s *Store) MarkAIBrowserControlRunFailed(ctx context.Context, id string, message string) error {
	tag, err := s.db.Exec(ctx, `
UPDATE ai_browser_control_runs
SET status = $2, error_message = $3, completed_at = now(), updated_at = now()
WHERE id = $1
`, id, StatusFailed, message)
	if err != nil {
		return fmt.Errorf("mark AI Browser Control run failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetAIBrowserControlTrace(ctx context.Context, id string) (*AIBrowserControlTrace, error) {
	run, err := s.GetAIBrowserControlRun(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, run.ProjectID)
	if err != nil {
		return nil, err
	}
	steps, err := s.ListAIBrowserControlSteps(ctx, id)
	if err != nil {
		return nil, err
	}
	findings, err := s.ListFindingsForAIBrowserControlRun(ctx, id)
	if err != nil {
		return nil, err
	}
	evidence, err := s.ListEvidenceForAIBrowserControlRun(ctx, id)
	if err != nil {
		return nil, err
	}
	return &AIBrowserControlTrace{
		Run:      *run,
		Project:  *project,
		Summary:  summarizeAIBrowserControl(*run, steps, findings, evidence),
		Steps:    steps,
		Findings: findings,
		Evidence: evidence,
	}, nil
}

func (s *Store) GetAIBrowserControlReport(ctx context.Context, id string) (*AIBrowserControlReport, error) {
	trace, err := s.GetAIBrowserControlTrace(ctx, id)
	if err != nil {
		return nil, err
	}
	report := &AIBrowserControlReport{
		GeneratedAt: time.Now().UTC(),
		Run:         trace.Run,
		Project:     trace.Project,
		Settings:    sanitizeAIBrowserControlSettings(trace.Run),
		Summary:     trace.Summary,
		Steps:       trace.Steps,
		Findings:    trace.Findings,
		Evidence:    trace.Evidence,
		SafetyNotes: aiBrowserControlSafetyNotes(),
		Limitations: aiBrowserControlLimitations(),
		Metadata: map[string]any{
			"run_type":                      RunTypeAIBrowserControl,
			"ai_browser_control_schema":     "v1",
			"policy_gate_required":          true,
			"ai_direct_browser_control":     false,
			"forms_submitted":               false,
			"destructive_actions":           false,
			"credentials_sent_to_ai":        false,
			"browser_storage_exposed_to_ai": false,
		},
	}
	report.ReportIntelligence = BuildReportIntelligence(ReportIntelligenceInput{
		ReportType:        RunTypeAIBrowserControl,
		ReportID:          trace.Run.ID,
		Status:            trace.Run.Status,
		Project:           &trace.Project,
		Findings:          trace.Findings,
		Evidence:          trace.Evidence,
		ChecksCompleted:   []string{"AI Browser Control"},
		ChecksSkipped:     []string{"Unsafe AI suggestions", "Unsupported actions", "Mutating forms", "External navigation by default"},
		WhatWasTested:     []string{"AI-proposed typed browser actions", "Deterministic policy validation", "Policy-approved safe browser actions", "Sanitized observations and trace metadata"},
		WhatWasNotTested:  defaultWhatWasNotTested(RunTypeAIBrowserControl),
		SafetyLimitations: report.Limitations,
	})
	return report, nil
}

func (s *Store) ListAIBrowserControlSteps(ctx context.Context, runID string) ([]AIBrowserControlStep, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, project_id, step_index, page_url, normalized_url, page_title, depth,
	sanitized_observation_json, COALESCE(ai_suggestion_json, '{}'::jsonb),
	action_type, action_label, action_target_url, action_selector_hint,
	policy_decision, policy_reason, execution_status, final_url, http_status,
	screenshot_evidence_id::text, console_error_count, failed_request_count, duration_ms, created_at
FROM ai_browser_control_steps
WHERE run_id = $1
ORDER BY step_index ASC, created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query AI Browser Control steps: %w", err)
	}
	defer rows.Close()

	steps := make([]AIBrowserControlStep, 0)
	for rows.Next() {
		step, err := scanAIBrowserControlStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI Browser Control steps: %w", err)
	}
	return steps, nil
}

func (s *Store) ListFindingsForAIBrowserControlRun(ctx context.Context, runID string) ([]Finding, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, ai_browser_control_run_id::text,
	scenario_execution_id::text, step_execution_id::text,
	title, severity, category, confidence, description, recommendation, evidence_ids, created_at
FROM findings
WHERE ai_browser_control_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query AI Browser Control findings: %w", err)
	}
	defer rows.Close()
	findings := make([]Finding, 0)
	for rows.Next() {
		finding, err := scanAIBrowserControlFinding(rows)
		if err != nil {
			return nil, err
		}
		findings = append(findings, finding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI Browser Control findings: %w", err)
	}
	return findings, nil
}

func (s *Store) ListEvidenceForAIBrowserControlRun(ctx context.Context, runID string) ([]Evidence, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, ai_browser_control_run_id::text,
	type, uri, metadata, created_at
FROM evidence
WHERE ai_browser_control_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query AI Browser Control evidence: %w", err)
	}
	defer rows.Close()
	evidence := make([]Evidence, 0)
	for rows.Next() {
		record, err := scanAIBrowserControlEvidence(rows)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI Browser Control evidence: %w", err)
	}
	return evidence, nil
}

func scanAIBrowserControlRun(row scanRow) (AIBrowserControlRun, error) {
	var run AIBrowserControlRun
	var credentialProfileID sql.NullString
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&run.ProviderID,
		&run.ProviderName,
		&credentialProfileID,
		&run.Status,
		&run.StartURL,
		&run.Goal,
		&run.MaxSteps,
		&run.MaxDepth,
		&run.SameOriginOnly,
		&run.PolicyVersion,
		&run.ExecutionMode,
		&run.StartedAt,
		&run.CompletedAt,
		&run.TotalSteps,
		&run.TotalAISuggestions,
		&run.TotalActionsApproved,
		&run.TotalActionsExecuted,
		&run.TotalActionsSkipped,
		&run.TotalPolicyBlocks,
		&run.TotalFindings,
		&run.ErrorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return AIBrowserControlRun{}, fmt.Errorf("scan AI Browser Control run: %w", err)
	}
	if credentialProfileID.Valid {
		run.CredentialProfileID = credentialProfileID.String
	}
	return run, nil
}

func scanAIBrowserControlStep(row scanRow) (AIBrowserControlStep, error) {
	var step AIBrowserControlStep
	var observationRaw []byte
	var suggestionRaw []byte
	var screenshotEvidenceID sql.NullString
	if err := row.Scan(
		&step.ID,
		&step.RunID,
		&step.ProjectID,
		&step.StepIndex,
		&step.PageURL,
		&step.NormalizedURL,
		&step.PageTitle,
		&step.Depth,
		&observationRaw,
		&suggestionRaw,
		&step.ActionType,
		&step.ActionLabel,
		&step.ActionTargetURL,
		&step.ActionSelectorHint,
		&step.PolicyDecision,
		&step.PolicyReason,
		&step.ExecutionStatus,
		&step.FinalURL,
		&step.HTTPStatus,
		&screenshotEvidenceID,
		&step.ConsoleErrorCount,
		&step.FailedRequestCount,
		&step.DurationMS,
		&step.CreatedAt,
	); err != nil {
		return AIBrowserControlStep{}, fmt.Errorf("scan AI Browser Control step: %w", err)
	}
	if len(observationRaw) > 0 {
		if err := json.Unmarshal(observationRaw, &step.SanitizedObservation); err != nil {
			return AIBrowserControlStep{}, fmt.Errorf("unmarshal sanitized observation: %w", err)
		}
	}
	if len(suggestionRaw) > 0 && string(suggestionRaw) != "{}" {
		if err := json.Unmarshal(suggestionRaw, &step.AISuggestion); err != nil {
			return AIBrowserControlStep{}, fmt.Errorf("unmarshal AI suggestion: %w", err)
		}
	}
	if screenshotEvidenceID.Valid {
		step.ScreenshotEvidenceID = screenshotEvidenceID.String
	}
	if step.SanitizedObservation == nil {
		step.SanitizedObservation = map[string]any{}
	}
	return step, nil
}

func scanAIBrowserControlEvidence(row scanRow) (Evidence, error) {
	var record Evidence
	var runID sql.NullString
	var executionID sql.NullString
	var authorizationRunID sql.NullString
	var discoveryRunID sql.NullString
	var safeExplorerRunID sql.NullString
	var aiBrowserRunID sql.NullString
	var metadataRaw []byte
	if err := row.Scan(
		&record.ID,
		&runID,
		&executionID,
		&authorizationRunID,
		&discoveryRunID,
		&safeExplorerRunID,
		&aiBrowserRunID,
		&record.Type,
		&record.URI,
		&metadataRaw,
		&record.CreatedAt,
	); err != nil {
		return Evidence{}, fmt.Errorf("scan AI Browser Control evidence: %w", err)
	}
	if err := json.Unmarshal(metadataRaw, &record.Metadata); err != nil {
		return Evidence{}, fmt.Errorf("unmarshal evidence metadata: %w", err)
	}
	if runID.Valid {
		record.RunID = runID.String
	}
	if executionID.Valid {
		record.TestPlanExecutionID = executionID.String
	}
	if authorizationRunID.Valid {
		record.AuthorizationRunID = authorizationRunID.String
	}
	if discoveryRunID.Valid {
		record.DiscoveryRunID = discoveryRunID.String
	}
	if safeExplorerRunID.Valid {
		record.SafeExplorerRunID = safeExplorerRunID.String
	}
	if aiBrowserRunID.Valid {
		record.AIBrowserControlRunID = aiBrowserRunID.String
	}
	return record, nil
}

func scanAIBrowserControlFinding(row scanRow) (Finding, error) {
	var finding Finding
	var runID sql.NullString
	var executionID sql.NullString
	var authorizationRunID sql.NullString
	var discoveryRunID sql.NullString
	var safeExplorerRunID sql.NullString
	var aiBrowserRunID sql.NullString
	var scenarioExecutionID sql.NullString
	var stepExecutionID sql.NullString
	var evidenceIDsRaw []byte
	if err := row.Scan(
		&finding.ID,
		&runID,
		&executionID,
		&authorizationRunID,
		&discoveryRunID,
		&safeExplorerRunID,
		&aiBrowserRunID,
		&scenarioExecutionID,
		&stepExecutionID,
		&finding.Title,
		&finding.Severity,
		&finding.Category,
		&finding.Confidence,
		&finding.Description,
		&finding.Recommendation,
		&evidenceIDsRaw,
		&finding.CreatedAt,
	); err != nil {
		return Finding{}, fmt.Errorf("scan AI Browser Control finding: %w", err)
	}
	if err := json.Unmarshal(evidenceIDsRaw, &finding.EvidenceIDs); err != nil {
		return Finding{}, fmt.Errorf("unmarshal finding evidence ids: %w", err)
	}
	if runID.Valid {
		finding.RunID = runID.String
	}
	if executionID.Valid {
		finding.TestPlanExecutionID = executionID.String
	}
	if authorizationRunID.Valid {
		finding.AuthorizationRunID = authorizationRunID.String
	}
	if discoveryRunID.Valid {
		finding.DiscoveryRunID = discoveryRunID.String
	}
	if safeExplorerRunID.Valid {
		finding.SafeExplorerRunID = safeExplorerRunID.String
	}
	if aiBrowserRunID.Valid {
		finding.AIBrowserControlRunID = aiBrowserRunID.String
	}
	if scenarioExecutionID.Valid {
		finding.ScenarioExecutionID = scenarioExecutionID.String
	}
	if stepExecutionID.Valid {
		finding.StepExecutionID = stepExecutionID.String
	}
	return finding, nil
}

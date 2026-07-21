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

func (s *Store) CreateFormTestRun(ctx context.Context, project Project, input FormTestRunRequest) (*FormTestRun, error) {
	run, err := scanFormTestRun(s.db.QueryRow(ctx, `
INSERT INTO form_test_runs (
	id, project_id, discovery_run_id, credential_profile_id, status, target_url,
	max_forms, max_tests_per_form, safe_get_only
) VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6, $7, $8, $9)
RETURNING id, project_id, discovery_run_id::text, credential_profile_id::text, status,
	target_url, max_forms, max_tests_per_form, safe_get_only, started_at, completed_at,
	total_forms_detected, total_forms_classified_safe, total_forms_tested, total_forms_skipped,
	total_findings, error_message, created_at, updated_at
`, uuid.NewString(), project.ID, input.DiscoveryRunID, input.CredentialProfileID, StatusQueued, input.TargetURL, input.MaxForms, input.MaxTestsPerForm, *input.SafeGetOnly))
	if err != nil {
		return nil, fmt.Errorf("insert form test run: %w", err)
	}
	return &run, nil
}

func (s *Store) ListFormTestRuns(ctx context.Context, projectID string) ([]FormTestRun, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, discovery_run_id::text, credential_profile_id::text, status,
	target_url, max_forms, max_tests_per_form, safe_get_only, started_at, completed_at,
	total_forms_detected, total_forms_classified_safe, total_forms_tested, total_forms_skipped,
	total_findings, error_message, created_at, updated_at
FROM form_test_runs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query form test runs: %w", err)
	}
	defer rows.Close()

	runs := make([]FormTestRun, 0)
	for rows.Next() {
		run, err := scanFormTestRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate form test runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetFormTestRun(ctx context.Context, id string) (*FormTestRun, error) {
	run, err := scanFormTestRun(s.db.QueryRow(ctx, `
SELECT id, project_id, discovery_run_id::text, credential_profile_id::text, status,
	target_url, max_forms, max_tests_per_form, safe_get_only, started_at, completed_at,
	total_forms_detected, total_forms_classified_safe, total_forms_tested, total_forms_skipped,
	total_findings, error_message, created_at, updated_at
FROM form_test_runs
WHERE id = $1
`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get form test run: %w", err)
	}
	return &run, nil
}

func (s *Store) MarkFormTestRunFailed(ctx context.Context, id string, message string) error {
	tag, err := s.db.Exec(ctx, `
UPDATE form_test_runs
SET status = $2, error_message = $3, completed_at = now(), updated_at = now()
WHERE id = $1
`, id, StatusFailed, RedactSecrets(message))
	if err != nil {
		return fmt.Errorf("mark form test run failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListFormTestResults(ctx context.Context, runID string) ([]FormTestResult, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, project_id, page_url, form_action, form_method, classification,
	safety, decision, skip_reason, submitted_url, final_url, http_status, page_title,
	test_values_summary, screenshot_evidence_id::text, console_error_count,
	failed_request_count, duration_ms, finding_id::text, created_at
FROM form_test_results
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query form test results: %w", err)
	}
	defer rows.Close()

	results := make([]FormTestResult, 0)
	for rows.Next() {
		result, err := scanFormTestResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate form test results: %w", err)
	}
	return results, nil
}

func (s *Store) GetFormTestReport(ctx context.Context, id string) (*FormTestReport, error) {
	run, err := s.GetFormTestRun(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, run.ProjectID)
	if err != nil {
		return nil, err
	}
	results, err := s.ListFormTestResults(ctx, id)
	if err != nil {
		return nil, err
	}
	findings, err := s.ListFindingsForFormTestRun(ctx, id)
	if err != nil {
		return nil, err
	}
	evidence, err := s.ListEvidenceForFormTestRun(ctx, id)
	if err != nil {
		return nil, err
	}
	var discoveryRun *DiscoveryRun
	if run.DiscoveryRunID != "" {
		discoveryRun, _ = s.GetDiscoveryRun(ctx, run.DiscoveryRunID)
	}
	report := &FormTestReport{
		GeneratedAt:  time.Now().UTC(),
		Run:          *run,
		Project:      *project,
		DiscoveryRun: discoveryRun,
		Settings:     sanitizeFormTestSettings(*run),
		Summary:      summarizeFormTest(*run, results, findings, evidence),
		Results:      results,
		Findings:     findings,
		Evidence:     evidence,
		SafetyNotes:  formTestSafetyNotes(),
		Limitations:  formTestLimitations(),
		Metadata: map[string]any{
			"run_type":                      RunTypeFormTest,
			"form_test_schema_version":      "v1",
			"safe_get_forms_only":           true,
			"forms_submitted":               true,
			"mutating_forms_submitted":      false,
			"arbitrary_form_submission":     false,
			"destructive_actions":           false,
			"autonomous_ai_browser_control": false,
			"credentials_sent_to_ai":        false,
			"browser_storage_exposed_to_ai": false,
		},
	}
	report.ReportIntelligence = BuildReportIntelligence(ReportIntelligenceInput{
		ReportType:        RunTypeFormTest,
		ReportID:          run.ID,
		Status:            run.Status,
		Project:           project,
		Findings:          findings,
		Evidence:          evidence,
		ChecksCompleted:   []string{"Safe Form Testing"},
		ChecksSkipped:     []string{"Mutating forms", "Sensitive forms", "Unsafe workflows", "Payload attacks", "Fuzzing"},
		WhatWasTested:     []string{"Deterministic form classification", "Simple same-origin GET search/filter/navigation forms", "Bounded benign form values", "Screenshots, console error counts, and failed network request counts"},
		WhatWasNotTested:  defaultWhatWasNotTested(RunTypeFormTest),
		SafetyLimitations: report.Limitations,
	})
	return report, nil
}

func (s *Store) ListFindingsForFormTestRun(ctx context.Context, runID string) ([]Finding, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, ai_browser_control_run_id::text,
	form_test_run_id::text, scenario_execution_id::text, step_execution_id::text,
	title, severity, category, confidence, description, recommendation, evidence_ids, created_at
FROM findings
WHERE form_test_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query form test findings: %w", err)
	}
	defer rows.Close()

	findings := make([]Finding, 0)
	for rows.Next() {
		finding, err := scanFormTestFinding(rows)
		if err != nil {
			return nil, err
		}
		findings = append(findings, finding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate form test findings: %w", err)
	}
	return findings, nil
}

func (s *Store) ListEvidenceForFormTestRun(ctx context.Context, runID string) ([]Evidence, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, ai_browser_control_run_id::text,
	form_test_run_id::text, type, uri, metadata, created_at
FROM evidence
WHERE form_test_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query form test evidence: %w", err)
	}
	defer rows.Close()

	evidence := make([]Evidence, 0)
	for rows.Next() {
		record, err := scanFormTestEvidence(rows)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate form test evidence: %w", err)
	}
	return evidence, nil
}

func scanFormTestRun(row scanRow) (FormTestRun, error) {
	var run FormTestRun
	var discoveryRunID sql.NullString
	var credentialProfileID sql.NullString
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&discoveryRunID,
		&credentialProfileID,
		&run.Status,
		&run.TargetURL,
		&run.MaxForms,
		&run.MaxTestsPerForm,
		&run.SafeGetOnly,
		&run.StartedAt,
		&run.CompletedAt,
		&run.TotalFormsDetected,
		&run.TotalFormsClassifiedSafe,
		&run.TotalFormsTested,
		&run.TotalFormsSkipped,
		&run.TotalFindings,
		&run.ErrorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return FormTestRun{}, fmt.Errorf("scan form test run: %w", err)
	}
	if discoveryRunID.Valid {
		run.DiscoveryRunID = discoveryRunID.String
	}
	if credentialProfileID.Valid {
		run.CredentialProfileID = credentialProfileID.String
	}
	return run, nil
}

func scanFormTestResult(row scanRow) (FormTestResult, error) {
	var result FormTestResult
	var valuesRaw []byte
	var screenshotEvidenceID sql.NullString
	var findingID sql.NullString
	if err := row.Scan(
		&result.ID,
		&result.RunID,
		&result.ProjectID,
		&result.PageURL,
		&result.FormAction,
		&result.FormMethod,
		&result.Classification,
		&result.Safety,
		&result.Decision,
		&result.SkipReason,
		&result.SubmittedURL,
		&result.FinalURL,
		&result.HTTPStatus,
		&result.PageTitle,
		&valuesRaw,
		&screenshotEvidenceID,
		&result.ConsoleErrorCount,
		&result.FailedRequestCount,
		&result.DurationMS,
		&findingID,
		&result.CreatedAt,
	); err != nil {
		return FormTestResult{}, fmt.Errorf("scan form test result: %w", err)
	}
	result.TestValuesSummary = map[string]any{}
	if len(valuesRaw) > 0 {
		if err := json.Unmarshal(valuesRaw, &result.TestValuesSummary); err != nil {
			return FormTestResult{}, fmt.Errorf("unmarshal form test values summary: %w", err)
		}
	}
	if screenshotEvidenceID.Valid {
		result.ScreenshotEvidenceID = screenshotEvidenceID.String
	}
	if findingID.Valid {
		result.FindingID = findingID.String
	}
	return result, nil
}

func scanFormTestEvidence(row scanRow) (Evidence, error) {
	var record Evidence
	var runID sql.NullString
	var executionID sql.NullString
	var authorizationRunID sql.NullString
	var discoveryRunID sql.NullString
	var safeExplorerRunID sql.NullString
	var aiBrowserRunID sql.NullString
	var formTestRunID sql.NullString
	var metadataRaw []byte
	if err := row.Scan(
		&record.ID,
		&runID,
		&executionID,
		&authorizationRunID,
		&discoveryRunID,
		&safeExplorerRunID,
		&aiBrowserRunID,
		&formTestRunID,
		&record.Type,
		&record.URI,
		&metadataRaw,
		&record.CreatedAt,
	); err != nil {
		return Evidence{}, fmt.Errorf("scan form test evidence: %w", err)
	}
	if err := json.Unmarshal(metadataRaw, &record.Metadata); err != nil {
		return Evidence{}, fmt.Errorf("unmarshal form test evidence metadata: %w", err)
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
	if formTestRunID.Valid {
		record.FormTestRunID = formTestRunID.String
	}
	return record, nil
}

func scanFormTestFinding(row scanRow) (Finding, error) {
	var finding Finding
	var runID sql.NullString
	var executionID sql.NullString
	var authorizationRunID sql.NullString
	var discoveryRunID sql.NullString
	var safeExplorerRunID sql.NullString
	var aiBrowserRunID sql.NullString
	var formTestRunID sql.NullString
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
		&formTestRunID,
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
		return Finding{}, fmt.Errorf("scan form test finding: %w", err)
	}
	if err := json.Unmarshal(evidenceIDsRaw, &finding.EvidenceIDs); err != nil {
		return Finding{}, fmt.Errorf("unmarshal form test finding evidence ids: %w", err)
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
	if formTestRunID.Valid {
		finding.FormTestRunID = formTestRunID.String
	}
	if scenarioExecutionID.Valid {
		finding.ScenarioExecutionID = scenarioExecutionID.String
	}
	if stepExecutionID.Valid {
		finding.StepExecutionID = stepExecutionID.String
	}
	return finding, nil
}

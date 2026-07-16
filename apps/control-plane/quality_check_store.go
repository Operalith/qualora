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

func (s *Store) CreateQualityCheckRun(ctx context.Context, projectID string, input QualityCheckRunRequest) (*QualityCheckRun, error) {
	run, err := scanQualityCheckRun(s.db.QueryRow(ctx, `
INSERT INTO quality_check_runs (
	id, project_id, discovery_run_id, credential_profile_id, status, target_url,
	max_pages, include_security, include_accessibility, include_performance
) VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6, $7, $8, $9, $10)
RETURNING id, project_id, discovery_run_id::text, credential_profile_id::text, status,
	target_url, max_pages, include_security, include_accessibility, include_performance,
	started_at, completed_at, total_pages, total_findings, critical_findings,
	high_findings, medium_findings, low_findings, info_findings, error_message,
	summary_json, created_at, updated_at
`, uuid.NewString(), projectID, input.DiscoveryRunID, input.CredentialProfileID, StatusQueued, input.TargetURL, input.MaxPages, *input.IncludeSecurity, *input.IncludeAccessibility, *input.IncludePerformance))
	if err != nil {
		return nil, fmt.Errorf("insert quality check run: %w", err)
	}
	return &run, nil
}

func (s *Store) ListQualityCheckRuns(ctx context.Context, projectID string) ([]QualityCheckRun, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, discovery_run_id::text, credential_profile_id::text, status,
	target_url, max_pages, include_security, include_accessibility, include_performance,
	started_at, completed_at, total_pages, total_findings, critical_findings,
	high_findings, medium_findings, low_findings, info_findings, error_message,
	summary_json, created_at, updated_at
FROM quality_check_runs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query quality check runs: %w", err)
	}
	defer rows.Close()

	runs := make([]QualityCheckRun, 0)
	for rows.Next() {
		run, err := scanQualityCheckRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quality check runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetQualityCheckRun(ctx context.Context, id string) (*QualityCheckRun, error) {
	run, err := scanQualityCheckRun(s.db.QueryRow(ctx, `
SELECT id, project_id, discovery_run_id::text, credential_profile_id::text, status,
	target_url, max_pages, include_security, include_accessibility, include_performance,
	started_at, completed_at, total_pages, total_findings, critical_findings,
	high_findings, medium_findings, low_findings, info_findings, error_message,
	summary_json, created_at, updated_at
FROM quality_check_runs
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get quality check run: %w", err)
	}
	return &run, nil
}

func (s *Store) MarkQualityCheckRunFailed(ctx context.Context, id string, message string) error {
	tag, err := s.db.Exec(ctx, `
UPDATE quality_check_runs
SET status = $2, error_message = $3, completed_at = now(), updated_at = now()
WHERE id = $1
`, id, StatusFailed, RedactSecrets(message))
	if err != nil {
		return fmt.Errorf("mark quality check run failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListQualityCheckResults(ctx context.Context, runID string) ([]QualityCheckResult, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, project_id, category, rule_id, severity, title, description,
	recommendation, url, evidence_json, created_at
FROM quality_check_results
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query quality check results: %w", err)
	}
	defer rows.Close()

	results := make([]QualityCheckResult, 0)
	for rows.Next() {
		result, err := scanQualityCheckResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quality check results: %w", err)
	}
	return results, nil
}

func (s *Store) GetQualityCheckReport(ctx context.Context, id string) (*QualityCheckReport, error) {
	run, err := s.GetQualityCheckRun(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, run.ProjectID)
	if err != nil {
		return nil, err
	}
	results, err := s.ListQualityCheckResults(ctx, id)
	if err != nil {
		return nil, err
	}
	var discoveryRun *DiscoveryRun
	if run.DiscoveryRunID != "" {
		discoveryRun, _ = s.GetDiscoveryRun(ctx, run.DiscoveryRunID)
	}
	return &QualityCheckReport{
		GeneratedAt:  time.Now().UTC(),
		Run:          *run,
		Project:      *project,
		DiscoveryRun: discoveryRun,
		Summary:      summarizeQualityCheckResults(*run, results),
		Results:      results,
		SafetyNotes:  qualityCheckSafetyNotes(),
		Limitations:  qualityCheckLimitations(),
		Metadata: map[string]any{
			"run_type":                      RunTypeQualityCheck,
			"quality_check_schema_version":  "v1",
			"forms_submitted":               false,
			"destructive_actions":           false,
			"autonomous_ai_browser_control": false,
			"credentials_sent_to_ai":        false,
			"browser_storage_exposed_to_ai": false,
		},
	}, nil
}

func scanQualityCheckRun(row scanRow) (QualityCheckRun, error) {
	var run QualityCheckRun
	var discoveryRunID sql.NullString
	var credentialProfileID sql.NullString
	var summaryRaw []byte
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&discoveryRunID,
		&credentialProfileID,
		&run.Status,
		&run.TargetURL,
		&run.MaxPages,
		&run.IncludeSecurity,
		&run.IncludeAccessibility,
		&run.IncludePerformance,
		&run.StartedAt,
		&run.CompletedAt,
		&run.TotalPages,
		&run.TotalFindings,
		&run.CriticalFindings,
		&run.HighFindings,
		&run.MediumFindings,
		&run.LowFindings,
		&run.InfoFindings,
		&run.ErrorMessage,
		&summaryRaw,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return QualityCheckRun{}, fmt.Errorf("scan quality check run: %w", err)
	}
	if discoveryRunID.Valid {
		run.DiscoveryRunID = discoveryRunID.String
	}
	if credentialProfileID.Valid {
		run.CredentialProfileID = credentialProfileID.String
	}
	run.Summary = map[string]any{}
	if len(summaryRaw) > 0 {
		if err := json.Unmarshal(summaryRaw, &run.Summary); err != nil {
			return QualityCheckRun{}, fmt.Errorf("unmarshal quality check run summary: %w", err)
		}
	}
	return run, nil
}

func scanQualityCheckResult(row scanRow) (QualityCheckResult, error) {
	var result QualityCheckResult
	var evidenceRaw []byte
	if err := row.Scan(
		&result.ID,
		&result.RunID,
		&result.ProjectID,
		&result.Category,
		&result.RuleID,
		&result.Severity,
		&result.Title,
		&result.Description,
		&result.Recommendation,
		&result.URL,
		&evidenceRaw,
		&result.CreatedAt,
	); err != nil {
		return QualityCheckResult{}, fmt.Errorf("scan quality check result: %w", err)
	}
	result.Evidence = map[string]any{}
	if len(evidenceRaw) > 0 {
		if err := json.Unmarshal(evidenceRaw, &result.Evidence); err != nil {
			return QualityCheckResult{}, fmt.Errorf("unmarshal quality check result evidence: %w", err)
		}
	}
	return result, nil
}

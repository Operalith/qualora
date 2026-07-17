package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateDiscoveryRun(ctx context.Context, project Project, input DiscoveryRunRequest) (*DiscoveryRun, error) {
	run, err := scanDiscoveryRun(s.db.QueryRow(ctx, `
INSERT INTO discovery_runs (
	id, project_id, credential_profile_id, status, start_url, max_pages, max_depth, same_origin_only
) VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8)
RETURNING id, project_id, credential_profile_id::text, status, start_url, max_pages, max_depth,
	same_origin_only, started_at, completed_at, total_pages, total_links, total_forms,
	total_console_errors, total_failed_requests, total_findings, error_message, created_at, updated_at
`, uuid.NewString(), project.ID, input.CredentialProfileID, StatusQueued, input.StartURL, input.MaxPages, input.MaxDepth, *input.SameOriginOnly))
	if err != nil {
		return nil, fmt.Errorf("insert discovery run: %w", err)
	}
	return &run, nil
}

func (s *Store) ListDiscoveryRuns(ctx context.Context, projectID string) ([]DiscoveryRun, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, credential_profile_id::text, status, start_url, max_pages, max_depth,
	same_origin_only, started_at, completed_at, total_pages, total_links, total_forms,
	total_console_errors, total_failed_requests, total_findings, error_message, created_at, updated_at
FROM discovery_runs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query discovery runs: %w", err)
	}
	defer rows.Close()

	runs := make([]DiscoveryRun, 0)
	for rows.Next() {
		run, err := scanDiscoveryRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discovery runs: %w", err)
	}
	return runs, nil
}

func (s *Store) GetDiscoveryRun(ctx context.Context, id string) (*DiscoveryRun, error) {
	run, err := scanDiscoveryRun(s.db.QueryRow(ctx, `
SELECT id, project_id, credential_profile_id::text, status, start_url, max_pages, max_depth,
	same_origin_only, started_at, completed_at, total_pages, total_links, total_forms,
	total_console_errors, total_failed_requests, total_findings, error_message, created_at, updated_at
FROM discovery_runs
WHERE id = $1
`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get discovery run: %w", err)
	}
	return &run, nil
}

func (s *Store) GetLatestCompletedDiscoveryRun(ctx context.Context, projectID string) (*DiscoveryRun, error) {
	run, err := scanDiscoveryRun(s.db.QueryRow(ctx, `
SELECT id, project_id, credential_profile_id::text, status, start_url, max_pages, max_depth,
	same_origin_only, started_at, completed_at, total_pages, total_links, total_forms,
	total_console_errors, total_failed_requests, total_findings, error_message, created_at, updated_at
FROM discovery_runs
WHERE project_id = $1 AND status = $2
ORDER BY completed_at DESC NULLS LAST, created_at DESC
LIMIT 1
`, projectID, StatusCompleted))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get latest completed discovery run: %w", err)
	}
	return &run, nil
}

func (s *Store) MarkDiscoveryRunFailed(ctx context.Context, id string, message string) error {
	tag, err := s.db.Exec(ctx, `
UPDATE discovery_runs
SET status = $2, error_message = $3, completed_at = now(), updated_at = now()
WHERE id = $1
`, id, StatusFailed, message)
	if err != nil {
		return fmt.Errorf("mark discovery run failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetDiscoveryMap(ctx context.Context, id string) (*DiscoveryMap, error) {
	run, err := s.GetDiscoveryRun(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, run.ProjectID)
	if err != nil {
		return nil, err
	}
	pages, err := s.ListDiscoveredPages(ctx, id)
	if err != nil {
		return nil, err
	}
	links, err := s.ListDiscoveredLinks(ctx, id)
	if err != nil {
		return nil, err
	}
	forms, err := s.ListDiscoveredForms(ctx, id)
	if err != nil {
		return nil, err
	}
	findings, err := s.ListFindingsForDiscoveryRun(ctx, id)
	if err != nil {
		return nil, err
	}
	evidence, err := s.ListEvidenceForDiscoveryRun(ctx, id)
	if err != nil {
		return nil, err
	}
	return &DiscoveryMap{
		Run:      *run,
		Project:  *project,
		Summary:  summarizeDiscoveryMap(*run, pages, links, forms, findings),
		Pages:    pages,
		Links:    links,
		Forms:    forms,
		Findings: findings,
		Evidence: evidence,
	}, nil
}

func (s *Store) GetDiscoveryReport(ctx context.Context, id string) (*DiscoveryReport, error) {
	discoveryMap, err := s.GetDiscoveryMap(ctx, id)
	if err != nil {
		return nil, err
	}
	return &DiscoveryReport{
		GeneratedAt: time.Now().UTC(),
		Run:         discoveryMap.Run,
		Project:     discoveryMap.Project,
		Settings:    sanitizeDiscoverySettings(discoveryMap.Run),
		Summary:     discoveryMap.Summary,
		Pages:       discoveryMap.Pages,
		Links:       discoveryMap.Links,
		Forms:       discoveryMap.Forms,
		Findings:    discoveryMap.Findings,
		Evidence:    discoveryMap.Evidence,
		SafetyNotes: discoverySafetyNotes(),
		Limitations: discoveryLimitations(),
		Metadata: map[string]any{
			"run_type":                       RunTypeAppDiscovery,
			"application_map_schema_version": "v1",
			"forms_submitted":                false,
			"autonomous_ai_browser_control":  false,
			"destructive_actions":            false,
		},
	}, nil
}

func (s *Store) ListDiscoveredPages(ctx context.Context, runID string) ([]DiscoveredPage, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, discovery_run_id, project_id, url, normalized_url, path, title, http_status,
	content_type, body_text_length, load_duration_ms, depth, screenshot_evidence_id::text,
	console_error_count, failed_request_count, discovered_at, created_at
FROM discovered_pages
WHERE discovery_run_id = $1
ORDER BY depth ASC, created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query discovered pages: %w", err)
	}
	defer rows.Close()

	pages := make([]DiscoveredPage, 0)
	for rows.Next() {
		page, err := scanDiscoveredPage(rows)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discovered pages: %w", err)
	}
	return pages, nil
}

func (s *Store) ListDiscoveredLinks(ctx context.Context, runID string) ([]DiscoveredLink, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, discovery_run_id, source_page_id, href, normalized_url, link_text, same_origin,
	skipped, skip_reason, created_at
FROM discovered_links
WHERE discovery_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query discovered links: %w", err)
	}
	defer rows.Close()

	links := make([]DiscoveredLink, 0)
	for rows.Next() {
		link, err := scanDiscoveredLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discovered links: %w", err)
	}
	return links, nil
}

func (s *Store) ListDiscoveredForms(ctx context.Context, runID string) ([]DiscoveredForm, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, discovery_run_id, page_id, form_name, form_action, form_method, field_count,
	password_field_count, submit_button_count, classification, skipped_reason, created_at
FROM discovered_forms
WHERE discovery_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query discovered forms: %w", err)
	}
	defer rows.Close()

	forms := make([]DiscoveredForm, 0)
	for rows.Next() {
		form, err := scanDiscoveredForm(rows)
		if err != nil {
			return nil, err
		}
		forms = append(forms, form)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discovered forms: %w", err)
	}
	if len(forms) == 0 {
		return forms, nil
	}
	fields, err := s.listDiscoveredFormFields(ctx, runID)
	if err != nil {
		return nil, err
	}
	fieldsByFormID := make(map[string][]DiscoveredFormField, len(forms))
	for _, field := range fields {
		fieldsByFormID[field.FormID] = append(fieldsByFormID[field.FormID], field)
	}
	for index := range forms {
		forms[index].Fields = sortedFormFields(fieldsByFormID[forms[index].ID])
	}
	return forms, nil
}

func (s *Store) listDiscoveredFormFields(ctx context.Context, runID string) ([]DiscoveredFormField, error) {
	rows, err := s.db.Query(ctx, `
SELECT ff.id, ff.form_id, ff.field_name, ff.field_type, ff.placeholder, ff.label, ff.required, ff.created_at
FROM discovered_form_fields ff
JOIN discovered_forms f ON f.id = ff.form_id
WHERE f.discovery_run_id = $1
ORDER BY ff.created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query discovered form fields: %w", err)
	}
	defer rows.Close()

	fields := make([]DiscoveredFormField, 0)
	for rows.Next() {
		field, err := scanDiscoveredFormField(rows)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discovered form fields: %w", err)
	}
	return fields, nil
}

func (s *Store) ListFindingsForDiscoveryRun(ctx context.Context, runID string) ([]Finding, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, scenario_execution_id::text, step_execution_id::text,
	title, severity, category, confidence, description, recommendation, evidence_ids, created_at
FROM findings
WHERE discovery_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query discovery findings: %w", err)
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
		return nil, fmt.Errorf("iterate discovery findings: %w", err)
	}
	return findings, nil
}

func (s *Store) ListEvidenceForDiscoveryRun(ctx context.Context, runID string) ([]Evidence, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	discovery_run_id::text, safe_explorer_run_id::text, type, uri, metadata, created_at
FROM evidence
WHERE discovery_run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query discovery evidence: %w", err)
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
		return nil, fmt.Errorf("iterate discovery evidence: %w", err)
	}
	return records, nil
}

func scanDiscoveryRun(row scanRow) (DiscoveryRun, error) {
	var run DiscoveryRun
	var credentialProfileID sql.NullString
	if err := row.Scan(
		&run.ID,
		&run.ProjectID,
		&credentialProfileID,
		&run.Status,
		&run.StartURL,
		&run.MaxPages,
		&run.MaxDepth,
		&run.SameOriginOnly,
		&run.StartedAt,
		&run.CompletedAt,
		&run.TotalPages,
		&run.TotalLinks,
		&run.TotalForms,
		&run.TotalConsoleErrors,
		&run.TotalFailedRequests,
		&run.TotalFindings,
		&run.ErrorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return DiscoveryRun{}, fmt.Errorf("scan discovery run: %w", err)
	}
	if credentialProfileID.Valid {
		run.CredentialProfileID = credentialProfileID.String
	}
	return run, nil
}

func scanDiscoveredPage(row scanRow) (DiscoveredPage, error) {
	var page DiscoveredPage
	var httpStatus sql.NullInt32
	var bodyTextLength sql.NullInt32
	var loadDurationMS sql.NullInt32
	var screenshotEvidenceID sql.NullString
	if err := row.Scan(
		&page.ID,
		&page.DiscoveryRunID,
		&page.ProjectID,
		&page.URL,
		&page.NormalizedURL,
		&page.Path,
		&page.Title,
		&httpStatus,
		&page.ContentType,
		&bodyTextLength,
		&loadDurationMS,
		&page.Depth,
		&screenshotEvidenceID,
		&page.ConsoleErrorCount,
		&page.FailedRequestCount,
		&page.DiscoveredAt,
		&page.CreatedAt,
	); err != nil {
		return DiscoveredPage{}, fmt.Errorf("scan discovered page: %w", err)
	}
	if httpStatus.Valid {
		value := int(httpStatus.Int32)
		page.HTTPStatus = &value
	}
	if bodyTextLength.Valid {
		value := int(bodyTextLength.Int32)
		page.BodyTextLength = &value
	}
	if loadDurationMS.Valid {
		value := int(loadDurationMS.Int32)
		page.LoadDurationMS = &value
	}
	if screenshotEvidenceID.Valid {
		page.ScreenshotEvidenceID = screenshotEvidenceID.String
	}
	return page, nil
}

func scanDiscoveredLink(row scanRow) (DiscoveredLink, error) {
	var link DiscoveredLink
	if err := row.Scan(
		&link.ID,
		&link.DiscoveryRunID,
		&link.SourcePageID,
		&link.Href,
		&link.NormalizedURL,
		&link.LinkText,
		&link.SameOrigin,
		&link.Skipped,
		&link.SkipReason,
		&link.CreatedAt,
	); err != nil {
		return DiscoveredLink{}, fmt.Errorf("scan discovered link: %w", err)
	}
	return link, nil
}

func scanDiscoveredForm(row scanRow) (DiscoveredForm, error) {
	var form DiscoveredForm
	if err := row.Scan(
		&form.ID,
		&form.DiscoveryRunID,
		&form.PageID,
		&form.FormName,
		&form.FormAction,
		&form.FormMethod,
		&form.FieldCount,
		&form.PasswordFieldCount,
		&form.SubmitButtonCount,
		&form.Classification,
		&form.SkippedReason,
		&form.CreatedAt,
	); err != nil {
		return DiscoveredForm{}, fmt.Errorf("scan discovered form: %w", err)
	}
	return form, nil
}

func scanDiscoveredFormField(row scanRow) (DiscoveredFormField, error) {
	var field DiscoveredFormField
	if err := row.Scan(
		&field.ID,
		&field.FormID,
		&field.FieldName,
		&field.FieldType,
		&field.Placeholder,
		&field.Label,
		&field.Required,
		&field.CreatedAt,
	); err != nil {
		return DiscoveredFormField{}, fmt.Errorf("scan discovered form field: %w", err)
	}
	return field, nil
}

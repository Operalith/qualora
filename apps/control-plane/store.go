package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *Store) CreateProject(ctx context.Context, input CreateProjectRequest) (*Project, error) {
	allowedHosts, err := json.Marshal(input.AllowedHosts)
	if err != nil {
		return nil, fmt.Errorf("marshal allowed hosts: %w", err)
	}

	project := &Project{}
	var allowedHostsRaw []byte
	err = s.db.QueryRow(ctx, `
INSERT INTO projects (
	id, name, frontend_url, api_base_url, openapi_url, allowed_hosts,
	security_mode, destructive_actions, allow_private_targets
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, name, frontend_url, api_base_url, openapi_url, allowed_hosts,
	security_mode, destructive_actions, allow_private_targets, created_at, updated_at
`,
		uuid.NewString(),
		input.Name,
		input.FrontendURL,
		input.APIBaseURL,
		input.OpenAPIURL,
		allowedHosts,
		input.SecurityMode,
		input.DestructiveActions,
		input.AllowPrivateTargets,
	).Scan(
		&project.ID,
		&project.Name,
		&project.FrontendURL,
		&project.APIBaseURL,
		&project.OpenAPIURL,
		&allowedHostsRaw,
		&project.SecurityMode,
		&project.DestructiveActions,
		&project.AllowPrivateTargets,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}
	if err := json.Unmarshal(allowedHostsRaw, &project.AllowedHosts); err != nil {
		return nil, fmt.Errorf("unmarshal allowed hosts: %w", err)
	}
	return project, nil
}

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, name, frontend_url, api_base_url, openapi_url, allowed_hosts,
	security_mode, destructive_actions, allow_private_targets, created_at, updated_at
FROM projects
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (s *Store) GetProject(ctx context.Context, id string) (*Project, error) {
	row := s.db.QueryRow(ctx, `
SELECT id, name, frontend_url, api_base_url, openapi_url, allowed_hosts,
	security_mode, destructive_actions, allow_private_targets, created_at, updated_at
FROM projects
WHERE id = $1
`, id)

	project, err := scanProject(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (s *Store) CreateRun(ctx context.Context, project Project) (*TestRun, []RunJob, error) {
	kinds := jobKindsForProject(project)
	if len(kinds) == 0 {
		return nil, nil, fmt.Errorf("project has no runnable targets")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin create run: %w", err)
	}
	defer tx.Rollback(ctx)

	run := &TestRun{}
	err = tx.QueryRow(ctx, `
INSERT INTO test_runs (id, project_id, status)
VALUES ($1, $2, $3)
RETURNING id, project_id, status, error_message, page_title, started_at, completed_at, created_at, updated_at
`, uuid.NewString(), project.ID, StatusPending).Scan(
		&run.ID,
		&run.ProjectID,
		&run.Status,
		&run.ErrorMessage,
		&run.PageTitle,
		&run.StartedAt,
		&run.CompletedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("insert test run: %w", err)
	}

	jobs := make([]RunJob, 0, len(kinds))
	for _, kind := range kinds {
		job := RunJob{}
		err := tx.QueryRow(ctx, `
INSERT INTO run_jobs (id, run_id, kind, status)
VALUES ($1, $2, $3, $4)
RETURNING id, run_id, kind, status, error_message, started_at, completed_at, created_at, updated_at
`, uuid.NewString(), run.ID, kind, StatusPending).Scan(
			&job.ID,
			&job.RunID,
			&job.Kind,
			&job.Status,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("insert run job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit create run: %w", err)
	}
	return run, jobs, nil
}

func (s *Store) MarkRunFailed(ctx context.Context, runID string, message string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin mark run failed: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
UPDATE run_jobs
SET status = $2, error_message = $3, completed_at = COALESCE(completed_at, now()), updated_at = now()
WHERE run_id = $1 AND status IN ('pending', 'running')
`, runID, StatusFailed, message); err != nil {
		return fmt.Errorf("mark run jobs failed: %w", err)
	}

	if _, err := tx.Exec(ctx, `
UPDATE test_runs
SET status = $2, error_message = $3, completed_at = now(), updated_at = now()
WHERE id = $1
`, runID, StatusFailed, message); err != nil {
		return fmt.Errorf("mark run failed: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit mark run failed: %w", err)
	}
	return nil
}

func (s *Store) GetRun(ctx context.Context, id string) (*TestRun, error) {
	run := &TestRun{}
	err := s.db.QueryRow(ctx, `
SELECT id, project_id, status, error_message, page_title, started_at, completed_at, created_at, updated_at
FROM test_runs
WHERE id = $1
`, id).Scan(
		&run.ID,
		&run.ProjectID,
		&run.Status,
		&run.ErrorMessage,
		&run.PageTitle,
		&run.StartedAt,
		&run.CompletedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get test run: %w", err)
	}
	return run, nil
}

func (s *Store) GetReport(ctx context.Context, runID string) (*Report, error) {
	run, err := s.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	findings, err := s.ListFindings(ctx, runID)
	if err != nil {
		return nil, err
	}
	evidence, err := s.ListEvidence(ctx, runID)
	if err != nil {
		return nil, err
	}
	jobs, err := s.ListRunJobs(ctx, runID)
	if err != nil {
		return nil, err
	}

	report := &Report{
		RunID:     run.ID,
		ProjectID: run.ProjectID,
		Status:    run.Status,
		Summary:   summarizeFindings(findings),
		Findings:  findings,
		Evidence:  evidence,
		Metadata: map[string]any{
			"page_title": run.PageTitle,
			"created_at": run.CreatedAt.Format(time.RFC3339),
			"jobs":       jobs,
		},
	}
	if run.ErrorMessage != "" {
		report.Metadata["error_message"] = run.ErrorMessage
	}
	return report, nil
}

func (s *Store) ListRunJobs(ctx context.Context, runID string) ([]RunJob, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, kind, status, error_message, started_at, completed_at, created_at, updated_at
FROM run_jobs
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query run jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]RunJob, 0)
	for rows.Next() {
		var job RunJob
		if err := rows.Scan(
			&job.ID,
			&job.RunID,
			&job.Kind,
			&job.Status,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan run job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run jobs: %w", err)
	}
	return jobs, nil
}

func (s *Store) ListFindings(ctx context.Context, runID string) ([]Finding, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, title, severity, category, confidence, description, recommendation, evidence_ids, created_at
FROM findings
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query findings: %w", err)
	}
	defer rows.Close()

	findings := make([]Finding, 0)
	for rows.Next() {
		var finding Finding
		var evidenceIDsRaw []byte
		if err := rows.Scan(
			&finding.ID,
			&finding.RunID,
			&finding.Title,
			&finding.Severity,
			&finding.Category,
			&finding.Confidence,
			&finding.Description,
			&finding.Recommendation,
			&evidenceIDsRaw,
			&finding.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan finding: %w", err)
		}
		if err := json.Unmarshal(evidenceIDsRaw, &finding.EvidenceIDs); err != nil {
			return nil, fmt.Errorf("unmarshal finding evidence ids: %w", err)
		}
		findings = append(findings, finding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate findings: %w", err)
	}
	return findings, nil
}

func (s *Store) ListEvidence(ctx context.Context, runID string) ([]Evidence, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, type, uri, metadata, created_at
FROM evidence
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query evidence: %w", err)
	}
	defer rows.Close()

	records := make([]Evidence, 0)
	for rows.Next() {
		var record Evidence
		var metadataRaw []byte
		if err := rows.Scan(
			&record.ID,
			&record.RunID,
			&record.Type,
			&record.URI,
			&metadataRaw,
			&record.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan evidence: %w", err)
		}
		if err := json.Unmarshal(metadataRaw, &record.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal evidence metadata: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evidence: %w", err)
	}
	return records, nil
}

func scanProject(row pgx.Row) (Project, error) {
	var project Project
	var allowedHostsRaw []byte
	if err := row.Scan(
		&project.ID,
		&project.Name,
		&project.FrontendURL,
		&project.APIBaseURL,
		&project.OpenAPIURL,
		&allowedHostsRaw,
		&project.SecurityMode,
		&project.DestructiveActions,
		&project.AllowPrivateTargets,
		&project.CreatedAt,
		&project.UpdatedAt,
	); err != nil {
		return Project{}, err
	}
	if err := json.Unmarshal(allowedHostsRaw, &project.AllowedHosts); err != nil {
		return Project{}, fmt.Errorf("unmarshal allowed hosts: %w", err)
	}
	return project, nil
}

func summarizeFindings(findings []Finding) ReportSummary {
	summary := ReportSummary{TotalFindings: len(findings)}
	for _, finding := range findings {
		switch finding.Severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		case "info":
			summary.Info++
		}
	}
	return summary
}

func jobKindsForProject(project Project) []string {
	kinds := make([]string, 0, 2)
	if project.FrontendURL != "" {
		kinds = append(kinds, JobKindBrowser)
	}
	if project.APIBaseURL != "" || project.OpenAPIURL != "" {
		kinds = append(kinds, JobKindAPI)
	}
	return kinds
}

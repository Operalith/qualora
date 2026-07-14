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

func (s *Store) CreateTestPlan(ctx context.Context, projectID string, runID string, providerID string, model string) (*TestPlan, error) {
	var nullableRunID any
	if runID != "" {
		nullableRunID = runID
	}
	var nullableProviderID any
	if providerID != "" {
		nullableProviderID = providerID
	}
	plan, err := scanTestPlan(s.db.QueryRow(ctx, `
INSERT INTO test_plans (id, project_id, run_id, provider_id, model, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, run_id::text, provider_id::text, '', model, status, title, summary,
	plan_json, risk_level, total_scenarios, error_message, created_at, updated_at
`, uuid.NewString(), projectID, nullableRunID, nullableProviderID, model, StatusRunning))
	if err != nil {
		return nil, fmt.Errorf("insert test plan: %w", err)
	}
	return &plan, nil
}

func (s *Store) CompleteTestPlan(ctx context.Context, id string, payload *TestPlanPayload, planJSON map[string]any) (*TestPlan, error) {
	rawJSON, err := json.Marshal(planJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal test plan json: %w", err)
	}
	plan, err := scanTestPlan(s.db.QueryRow(ctx, `
UPDATE test_plans
SET status = $2,
	title = $3,
	summary = $4,
	plan_json = $5,
	risk_level = $6,
	total_scenarios = $7,
	error_message = '',
	updated_at = now()
WHERE id = $1
RETURNING id, project_id, run_id::text, provider_id::text, '', model, status, title, summary,
	plan_json, risk_level, total_scenarios, error_message, created_at, updated_at
`, id, StatusCompleted, payload.Title, payload.Summary, rawJSON, testPlanRiskLevel(payload), len(payload.Scenarios)))
	if err != nil {
		return nil, fmt.Errorf("complete test plan: %w", err)
	}
	return &plan, nil
}

func (s *Store) FailTestPlan(ctx context.Context, id string, message string) (*TestPlan, error) {
	plan, err := scanTestPlan(s.db.QueryRow(ctx, `
UPDATE test_plans
SET status = $2, error_message = $3, updated_at = now()
WHERE id = $1
RETURNING id, project_id, run_id::text, provider_id::text, '', model, status, title, summary,
	plan_json, risk_level, total_scenarios, error_message, created_at, updated_at
`, id, StatusFailed, message))
	if err != nil {
		return nil, fmt.Errorf("fail test plan: %w", err)
	}
	return &plan, nil
}

func (s *Store) ListTestPlans(ctx context.Context, projectID string) ([]TestPlan, error) {
	rows, err := s.db.Query(ctx, `
SELECT t.id, t.project_id, t.run_id::text, t.provider_id::text, COALESCE(p.name, ''), t.model,
	t.status, t.title, t.summary, t.plan_json, t.risk_level, t.total_scenarios,
	t.error_message, t.created_at, t.updated_at
FROM test_plans t
LEFT JOIN ai_providers p ON p.id = t.provider_id
WHERE t.project_id = $1
ORDER BY t.created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query test plans: %w", err)
	}
	defer rows.Close()

	plans := make([]TestPlan, 0)
	for rows.Next() {
		plan, err := scanTestPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate test plans: %w", err)
	}
	return plans, nil
}

func (s *Store) GetTestPlan(ctx context.Context, id string) (*TestPlan, error) {
	plan, err := scanTestPlan(s.db.QueryRow(ctx, `
SELECT t.id, t.project_id, t.run_id::text, t.provider_id::text, COALESCE(p.name, ''), t.model,
	t.status, t.title, t.summary, t.plan_json, t.risk_level, t.total_scenarios,
	t.error_message, t.created_at, t.updated_at
FROM test_plans t
LEFT JOIN ai_providers p ON p.id = t.provider_id
WHERE t.id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (s *Store) DeleteTestPlan(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM test_plans WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete test plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListTestPlanRefsForRun(ctx context.Context, runID string) ([]TestPlanRef, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, title, status, risk_level, total_scenarios, created_at
FROM test_plans
WHERE run_id = $1 AND status = $2
ORDER BY created_at DESC
`, runID, StatusCompleted)
	if err != nil {
		return nil, fmt.Errorf("query test plan refs: %w", err)
	}
	defer rows.Close()

	refs := make([]TestPlanRef, 0)
	for rows.Next() {
		var ref TestPlanRef
		if err := rows.Scan(&ref.ID, &ref.Title, &ref.Status, &ref.RiskLevel, &ref.TotalScenarios, &ref.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan test plan ref: %w", err)
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate test plan refs: %w", err)
	}
	return refs, nil
}

func (s *Store) GetLatestRunForProject(ctx context.Context, projectID string) (*TestRun, error) {
	var run TestRun
	err := s.db.QueryRow(ctx, `
SELECT id, project_id, status, error_message, page_title, started_at, completed_at, created_at, updated_at
FROM test_runs
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT 1
`, projectID).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get latest run: %w", err)
	}
	return &run, nil
}

func scanTestPlan(row scanRow) (TestPlan, error) {
	var plan TestPlan
	var runID sql.NullString
	var providerID sql.NullString
	var planRaw []byte
	if err := row.Scan(
		&plan.ID,
		&plan.ProjectID,
		&runID,
		&providerID,
		&plan.ProviderName,
		&plan.Model,
		&plan.Status,
		&plan.Title,
		&plan.Summary,
		&planRaw,
		&plan.RiskLevel,
		&plan.TotalScenarios,
		&plan.ErrorMessage,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	); err != nil {
		return TestPlan{}, fmt.Errorf("scan test plan: %w", err)
	}
	if runID.Valid {
		plan.RunID = runID.String
	}
	if providerID.Valid {
		plan.ProviderID = providerID.String
	}
	if len(planRaw) == 0 {
		plan.PlanJSON = map[string]any{}
	} else if err := json.Unmarshal(planRaw, &plan.PlanJSON); err != nil {
		return TestPlan{}, fmt.Errorf("unmarshal test plan json: %w", err)
	}
	return plan, nil
}

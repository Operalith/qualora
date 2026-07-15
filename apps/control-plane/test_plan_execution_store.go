package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateTestPlanExecution(ctx context.Context, plan TestPlan, preview TestPlanExecutionPreview) (*TestPlanExecutionDetail, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create test plan execution: %w", err)
	}
	defer tx.Rollback(ctx)

	sourceRunID := any(nil)
	if plan.RunID != "" {
		sourceRunID = plan.RunID
	}

	executionID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
INSERT INTO test_plan_executions (id, test_plan_id, project_id, source_run_id, status)
VALUES ($1, $2, $3, $4, $5)
`, executionID, plan.ID, plan.ProjectID, sourceRunID, StatusQueued); err != nil {
		return nil, fmt.Errorf("insert test plan execution: %w", err)
	}

	for _, scenario := range preview.Scenarios {
		scenarioID := uuid.NewString()
		if _, err := tx.Exec(ctx, `
INSERT INTO test_plan_execution_scenarios (
	id, execution_id, scenario_id_from_plan, name, type, priority, status, skip_reason
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, scenarioID, executionID, scenario.ScenarioIDFromPlan, scenario.Name, scenario.Type, scenario.Priority, scenario.Status, scenario.SkipReason); err != nil {
			return nil, fmt.Errorf("insert test plan execution scenario: %w", err)
		}

		for _, step := range scenario.Steps {
			if _, err := tx.Exec(ctx, `
INSERT INTO test_plan_execution_steps (
	id, execution_id, scenario_execution_id, step_order, original_action, mapped_action,
	target, expected_result, status, skip_reason
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, uuid.NewString(), executionID, scenarioID, step.StepOrder, step.OriginalAction, step.MappedAction, step.Target, step.ExpectedResult, step.Status, step.SkipReason); err != nil {
				return nil, fmt.Errorf("insert test plan execution step: %w", err)
			}
		}
	}

	if _, err := tx.Exec(ctx, `SELECT refresh_test_plan_execution_status($1)`, executionID); err != nil {
		return nil, fmt.Errorf("refresh test plan execution status: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create test plan execution: %w", err)
	}

	return s.GetTestPlanExecution(ctx, executionID)
}

func (s *Store) ListTestPlanExecutions(ctx context.Context, testPlanID string) ([]TestPlanExecution, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, test_plan_id, project_id, source_run_id::text, status, total_scenarios,
	passed_scenarios, failed_scenarios, skipped_scenarios, total_steps, passed_steps,
	failed_steps, skipped_steps, error_message, started_at, completed_at, created_at, updated_at
FROM test_plan_executions
WHERE test_plan_id = $1
ORDER BY created_at DESC
`, testPlanID)
	if err != nil {
		return nil, fmt.Errorf("query test plan executions: %w", err)
	}
	defer rows.Close()

	executions := make([]TestPlanExecution, 0)
	for rows.Next() {
		execution, err := scanTestPlanExecution(rows)
		if err != nil {
			return nil, err
		}
		executions = append(executions, execution)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate test plan executions: %w", err)
	}
	return executions, nil
}

func (s *Store) GetTestPlanExecution(ctx context.Context, id string) (*TestPlanExecutionDetail, error) {
	execution, err := scanTestPlanExecution(s.db.QueryRow(ctx, `
SELECT id, test_plan_id, project_id, source_run_id::text, status, total_scenarios,
	passed_scenarios, failed_scenarios, skipped_scenarios, total_steps, passed_steps,
	failed_steps, skipped_steps, error_message, started_at, completed_at, created_at, updated_at
FROM test_plan_executions
WHERE id = $1
`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	scenarios, err := s.listTestPlanExecutionScenarios(ctx, id)
	if err != nil {
		return nil, err
	}
	return &TestPlanExecutionDetail{Execution: execution, Scenarios: scenarios}, nil
}

func (s *Store) MarkTestPlanExecutionFailed(ctx context.Context, id string, message string) error {
	tag, err := s.db.Exec(ctx, `
UPDATE test_plan_executions
SET status = $2, error_message = $3, completed_at = COALESCE(completed_at, now()), updated_at = now()
WHERE id = $1
`, id, StatusFailed, message)
	if err != nil {
		return fmt.Errorf("mark test plan execution failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetTestPlanExecutionReport(ctx context.Context, id string) (*TestPlanExecutionReport, error) {
	detail, err := s.GetTestPlanExecution(ctx, id)
	if err != nil {
		return nil, err
	}
	plan, err := s.GetTestPlan(ctx, detail.Execution.TestPlanID)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, detail.Execution.ProjectID)
	if err != nil {
		return nil, err
	}
	findings, err := s.ListFindingsForTestPlanExecution(ctx, id)
	if err != nil {
		return nil, err
	}
	evidence, err := s.ListEvidenceForTestPlanExecution(ctx, id)
	if err != nil {
		return nil, err
	}

	return &TestPlanExecutionReport{
		Execution:     detail.Execution,
		TestPlan:      *plan,
		Project:       *project,
		Scenarios:     detail.Scenarios,
		Findings:      findings,
		Evidence:      evidence,
		SafetySummary: summarizeTestPlanExecutionSafety(detail.Scenarios),
		GeneratedAt:   time.Now().UTC(),
	}, nil
}

func (s *Store) ListFindingsForTestPlanExecution(ctx context.Context, executionID string) ([]Finding, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	scenario_execution_id::text, step_execution_id::text,
	title, severity, category, confidence, description, recommendation, evidence_ids, created_at
FROM findings
WHERE test_plan_execution_id = $1
ORDER BY created_at ASC
`, executionID)
	if err != nil {
		return nil, fmt.Errorf("query test plan execution findings: %w", err)
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
		return nil, fmt.Errorf("iterate test plan execution findings: %w", err)
	}
	return findings, nil
}

func (s *Store) ListEvidenceForTestPlanExecution(ctx context.Context, executionID string) ([]Evidence, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id::text, test_plan_execution_id::text, authorization_check_run_id::text,
	type, uri, metadata, created_at
FROM evidence
WHERE test_plan_execution_id = $1
ORDER BY created_at ASC
`, executionID)
	if err != nil {
		return nil, fmt.Errorf("query test plan execution evidence: %w", err)
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
		return nil, fmt.Errorf("iterate test plan execution evidence: %w", err)
	}
	return records, nil
}

func (s *Store) listTestPlanExecutionScenarios(ctx context.Context, executionID string) ([]TestPlanExecutionScenario, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, execution_id, scenario_id_from_plan, name, type, priority, status,
	skip_reason, started_at, completed_at, created_at, updated_at
FROM test_plan_execution_scenarios
WHERE execution_id = $1
ORDER BY created_at ASC
`, executionID)
	if err != nil {
		return nil, fmt.Errorf("query test plan execution scenarios: %w", err)
	}
	defer rows.Close()

	scenarios := make([]TestPlanExecutionScenario, 0)
	scenarioIndex := map[string]int{}
	for rows.Next() {
		scenario, err := scanTestPlanExecutionScenario(rows)
		if err != nil {
			return nil, err
		}
		scenarioIndex[scenario.ID] = len(scenarios)
		scenarios = append(scenarios, scenario)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate test plan execution scenarios: %w", err)
	}

	stepRows, err := s.db.Query(ctx, `
SELECT id, execution_id, scenario_execution_id, step_order, original_action, mapped_action,
	target, expected_result, status, skip_reason, actual_result, error_message, duration_ms,
	evidence_id::text, created_at, updated_at
FROM test_plan_execution_steps
WHERE execution_id = $1
ORDER BY created_at ASC, step_order ASC
`, executionID)
	if err != nil {
		return nil, fmt.Errorf("query test plan execution steps: %w", err)
	}
	defer stepRows.Close()

	for stepRows.Next() {
		step, err := scanTestPlanExecutionStep(stepRows)
		if err != nil {
			return nil, err
		}
		index, ok := scenarioIndex[step.ScenarioExecutionID]
		if !ok {
			continue
		}
		scenarios[index].Steps = append(scenarios[index].Steps, step)
	}
	if err := stepRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate test plan execution steps: %w", err)
	}
	return scenarios, nil
}

func scanTestPlanExecution(row scanRow) (TestPlanExecution, error) {
	var execution TestPlanExecution
	var sourceRunID sql.NullString
	if err := row.Scan(
		&execution.ID,
		&execution.TestPlanID,
		&execution.ProjectID,
		&sourceRunID,
		&execution.Status,
		&execution.TotalScenarios,
		&execution.PassedScenarios,
		&execution.FailedScenarios,
		&execution.SkippedScenarios,
		&execution.TotalSteps,
		&execution.PassedSteps,
		&execution.FailedSteps,
		&execution.SkippedSteps,
		&execution.ErrorMessage,
		&execution.StartedAt,
		&execution.CompletedAt,
		&execution.CreatedAt,
		&execution.UpdatedAt,
	); err != nil {
		return TestPlanExecution{}, fmt.Errorf("scan test plan execution: %w", err)
	}
	if sourceRunID.Valid {
		execution.SourceRunID = sourceRunID.String
	}
	return execution, nil
}

func scanTestPlanExecutionScenario(row scanRow) (TestPlanExecutionScenario, error) {
	var scenario TestPlanExecutionScenario
	if err := row.Scan(
		&scenario.ID,
		&scenario.ExecutionID,
		&scenario.ScenarioIDFromPlan,
		&scenario.Name,
		&scenario.Type,
		&scenario.Priority,
		&scenario.Status,
		&scenario.SkipReason,
		&scenario.StartedAt,
		&scenario.CompletedAt,
		&scenario.CreatedAt,
		&scenario.UpdatedAt,
	); err != nil {
		return TestPlanExecutionScenario{}, fmt.Errorf("scan test plan execution scenario: %w", err)
	}
	return scenario, nil
}

func scanTestPlanExecutionStep(row scanRow) (TestPlanExecutionStep, error) {
	var step TestPlanExecutionStep
	var duration sql.NullInt64
	var evidenceID sql.NullString
	if err := row.Scan(
		&step.ID,
		&step.ExecutionID,
		&step.ScenarioExecutionID,
		&step.StepOrder,
		&step.OriginalAction,
		&step.MappedAction,
		&step.Target,
		&step.ExpectedResult,
		&step.Status,
		&step.SkipReason,
		&step.ActualResult,
		&step.ErrorMessage,
		&duration,
		&evidenceID,
		&step.CreatedAt,
		&step.UpdatedAt,
	); err != nil {
		return TestPlanExecutionStep{}, fmt.Errorf("scan test plan execution step: %w", err)
	}
	if duration.Valid {
		durationMS := int(duration.Int64)
		step.DurationMS = &durationMS
	}
	if evidenceID.Valid {
		step.EvidenceID = evidenceID.String
	}
	return step, nil
}

func summarizeTestPlanExecutionSafety(scenarios []TestPlanExecutionScenario) TestPlanExecutionSafetyReport {
	var summary TestPlanExecutionSafetyReport
	for _, scenario := range scenarios {
		if scenario.Status == StatusSkipped {
			summary.SkippedScenarios++
		}
		for _, step := range scenario.Steps {
			if step.Status == StatusSkipped {
				if isUnsafeExecutionSkipReason(step.SkipReason) || isUnsafeExecutionSkipReason(scenario.SkipReason) {
					summary.SkippedUnsafeSteps++
				} else {
					summary.SkippedUnsupportedSteps++
				}
				continue
			}
			summary.ExecutedSteps++
		}
	}
	return summary
}

func isUnsafeExecutionSkipReason(reason string) bool {
	reason = strings.ToLower(reason)
	return strings.Contains(reason, "unsafe") ||
		strings.Contains(reason, "destructive") ||
		strings.Contains(reason, "authentication") ||
		strings.Contains(reason, "out-of-scope")
}

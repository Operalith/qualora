package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const qaRunTimeout = 10 * time.Minute

func NormalizeQARunRequest(project Project, input QARunRequest) (QARunRequest, error) {
	input.Mode = strings.ToLower(strings.TrimSpace(input.Mode))
	if input.Mode == "" {
		input.Mode = "safe"
	}
	if input.Mode != "safe" {
		return input, fmt.Errorf("mode must be safe")
	}
	input.StartURL = strings.TrimSpace(input.StartURL)
	input.CredentialProfileID = strings.TrimSpace(input.CredentialProfileID)
	input.UseExistingDiscoveryRunID = strings.TrimSpace(input.UseExistingDiscoveryRunID)
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.ProductContext = strings.TrimSpace(sanitizeText(RedactSecrets(limitString(input.ProductContext, 4000))))
	if input.MaxPages == 0 {
		input.MaxPages = defaultDiscoveryMaxPages
	}
	if input.MaxPages < 1 || input.MaxPages > maxDiscoveryMaxPages {
		return input, fmt.Errorf("max_pages must be between 1 and %d", maxDiscoveryMaxPages)
	}
	if input.MaxDepth == 0 {
		input.MaxDepth = defaultDiscoveryMaxDepth
	}
	if input.MaxDepth < 0 || input.MaxDepth > maxDiscoveryMaxDepth {
		return input, fmt.Errorf("max_depth must be between 0 and %d", maxDiscoveryMaxDepth)
	}
	if input.MaxScenarios == 0 {
		input.MaxScenarios = defaultMaxTestPlanScenarios
	}
	if input.MaxScenarios < 1 || input.MaxScenarios > maxTestPlanScenarios {
		return input, fmt.Errorf("max_scenarios must be between 1 and %d", maxTestPlanScenarios)
	}
	planInput, err := NormalizeAITestPlanRequest(AITestPlanRequest{
		FocusAreas:   input.FocusAreas,
		MaxScenarios: input.MaxScenarios,
	})
	if err != nil {
		return input, err
	}
	input.FocusAreas = planInput.FocusAreas
	if project.FrontendURL == "" {
		return input, fmt.Errorf("project frontend_url is required for safe QA runs")
	}
	return input, nil
}

func (a *App) runSafeQARun(qaRunID string, project Project, input QARunRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), qaRunTimeout)
	defer cancel()

	if err := a.executeSafeQARun(ctx, qaRunID, project, input); err != nil {
		summary := map[string]any{
			"safe_execution":                true,
			"autonomous_ai_browser_control": false,
			"destructive_actions":           false,
			"error":                         RedactSecrets(err.Error()),
		}
		if _, failErr := a.store.FailQARun(ctx, qaRunID, err.Error(), summary); failErr != nil {
			a.logger.Error("mark QA run failed failed", "qa_run_id", qaRunID, "error", failErr)
		}
		a.logger.Error("safe QA run failed", "qa_run_id", qaRunID, "error", RedactSecrets(err.Error()))
	}
}

func (a *App) executeSafeQARun(ctx context.Context, qaRunID string, project Project, input QARunRequest) error {
	if _, err := a.store.UpdateQARunStatus(ctx, qaRunID, QARunStatusRunningDiscovery); err != nil {
		return err
	}
	discoveryRun, err := a.resolveQADiscoveryRun(ctx, qaRunID, project, input)
	if err != nil {
		return err
	}

	if _, err := a.store.UpdateQARunStatus(ctx, qaRunID, QARunStatusGeneratingPlan); err != nil {
		return err
	}
	includeDiscovery := true
	plan, err := a.generateAITestPlan(ctx, project, AITestPlanRequest{
		ProviderID:            input.ProviderID,
		DiscoveryRunID:        discoveryRun.ID,
		IncludeDiscoveryMap:   &includeDiscovery,
		ExecutionMode:         AITestPlanExecutionModeSafeExecutable,
		MaxPagesFromDiscovery: input.MaxPages,
		ProductContext:        input.ProductContext,
		FocusAreas:            input.FocusAreas,
		MaxScenarios:          input.MaxScenarios,
	})
	if err != nil {
		return err
	}
	if _, err := a.store.AttachQARunTestPlan(ctx, qaRunID, plan.ID); err != nil {
		return err
	}

	if _, err := a.store.UpdateQARunStatus(ctx, qaRunID, QARunStatusPreviewingExecution); err != nil {
		return err
	}
	preview, err := BuildTestPlanExecutionPreview(*plan, project, TestPlanExecutionRequest{
		DryRun:       true,
		MaxScenarios: input.MaxScenarios,
	})
	if err != nil {
		return err
	}
	summary := qaRunSummary(discoveryRun, plan, nil, preview, input.Execute, input.MaxScenarios)
	if !input.Execute {
		_, err := a.store.CompleteQARun(ctx, qaRunID, summary)
		return err
	}

	if _, err := a.store.UpdateQARunStatus(ctx, qaRunID, QARunStatusExecutingPlan); err != nil {
		return err
	}
	execution, err := a.store.CreateTestPlanExecution(ctx, *plan, *preview)
	if err != nil {
		return err
	}
	if _, err := a.store.AttachQARunExecution(ctx, qaRunID, execution.Execution.ID); err != nil {
		return err
	}
	if preview.ExecutableSteps > 0 {
		if err := a.queue.EnqueueTestPlanExecution(ctx, TestPlanExecutionJob{ExecutionID: execution.Execution.ID}); err != nil {
			_ = a.store.MarkTestPlanExecutionFailed(ctx, execution.Execution.ID, "test plan execution could not be queued")
			return err
		}
		execution, err = a.waitForTestPlanExecution(ctx, execution.Execution.ID)
		if err != nil {
			return err
		}
	}
	summary = qaRunSummary(discoveryRun, plan, execution, preview, input.Execute, input.MaxScenarios)
	if execution.Execution.Status != StatusCompleted {
		_, err := a.store.FailQARun(ctx, qaRunID, execution.Execution.ErrorMessage, summary)
		return err
	}
	_, err = a.store.CompleteQARun(ctx, qaRunID, summary)
	return err
}

func (a *App) executeExistingQARun(ctx context.Context, qaRunID string) error {
	run, err := a.store.GetQARun(ctx, qaRunID)
	if err != nil {
		return err
	}
	if run.TestPlanID == "" {
		return fmt.Errorf("QA run has no test plan to execute")
	}
	if run.TestPlanExecutionID != "" {
		return fmt.Errorf("QA run already has a test plan execution")
	}
	project, err := a.store.GetProject(ctx, run.ProjectID)
	if err != nil {
		return err
	}
	plan, err := a.store.GetTestPlan(ctx, run.TestPlanID)
	if err != nil {
		return err
	}
	if _, err := a.store.UpdateQARunStatus(ctx, qaRunID, QARunStatusExecutingPlan); err != nil {
		return err
	}
	preview, err := BuildTestPlanExecutionPreview(*plan, *project, TestPlanExecutionRequest{
		DryRun:       true,
		MaxScenarios: summaryInt(run.Summary, "max_scenarios", defaultMaxTestPlanScenarios),
	})
	if err != nil {
		return err
	}
	execution, err := a.store.CreateTestPlanExecution(ctx, *plan, *preview)
	if err != nil {
		return err
	}
	if _, err := a.store.AttachQARunExecution(ctx, qaRunID, execution.Execution.ID); err != nil {
		return err
	}
	if preview.ExecutableSteps > 0 {
		if err := a.queue.EnqueueTestPlanExecution(ctx, TestPlanExecutionJob{ExecutionID: execution.Execution.ID}); err != nil {
			_ = a.store.MarkTestPlanExecutionFailed(ctx, execution.Execution.ID, "test plan execution could not be queued")
			return err
		}
		execution, err = a.waitForTestPlanExecution(ctx, execution.Execution.ID)
		if err != nil {
			return err
		}
	}
	var discoveryRun *DiscoveryRun
	if run.DiscoveryRunID != "" {
		discoveryRun, _ = a.store.GetDiscoveryRun(ctx, run.DiscoveryRunID)
	}
	summary := qaRunSummary(discoveryRun, plan, execution, preview, true, summaryInt(run.Summary, "max_scenarios", defaultMaxTestPlanScenarios))
	if execution.Execution.Status != StatusCompleted {
		_, err := a.store.FailQARun(ctx, qaRunID, execution.Execution.ErrorMessage, summary)
		return err
	}
	_, err = a.store.CompleteQARun(ctx, qaRunID, summary)
	return err
}

func (a *App) resolveQADiscoveryRun(ctx context.Context, qaRunID string, project Project, input QARunRequest) (*DiscoveryRun, error) {
	if input.UseExistingDiscoveryRunID != "" {
		run, err := a.store.GetDiscoveryRun(ctx, input.UseExistingDiscoveryRunID)
		if err != nil {
			return nil, err
		}
		if run.ProjectID != project.ID {
			return nil, fmt.Errorf("selected discovery run does not belong to the project")
		}
		if run.Status != StatusCompleted {
			return nil, fmt.Errorf("selected discovery run must be completed before a safe QA run")
		}
		if _, err := a.store.AttachQARunDiscovery(ctx, qaRunID, run.ID); err != nil {
			return nil, err
		}
		return run, nil
	}

	if input.UseLatestDiscovery {
		run, err := a.store.GetLatestCompletedDiscoveryRun(ctx, project.ID)
		if err == nil {
			if _, err := a.store.AttachQARunDiscovery(ctx, qaRunID, run.ID); err != nil {
				return nil, err
			}
			return run, nil
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	sameOriginOnly := true
	normalized, err := NormalizeDiscoveryRunRequest(project, DiscoveryRunRequest{
		StartURL:            input.StartURL,
		CredentialProfileID: input.CredentialProfileID,
		MaxPages:            input.MaxPages,
		MaxDepth:            input.MaxDepth,
		SameOriginOnly:      &sameOriginOnly,
	})
	if err != nil {
		return nil, err
	}
	run, err := a.store.CreateDiscoveryRun(ctx, project, normalized)
	if err != nil {
		return nil, err
	}
	if _, err := a.store.AttachQARunDiscovery(ctx, qaRunID, run.ID); err != nil {
		return nil, err
	}
	if err := a.queue.EnqueueDiscoveryRun(ctx, DiscoveryRunJob{DiscoveryRunID: run.ID, ProjectID: project.ID}); err != nil {
		_ = a.store.MarkDiscoveryRunFailed(ctx, run.ID, "discovery run could not be queued")
		return nil, err
	}
	return a.waitForDiscoveryRun(ctx, run.ID)
}

func (a *App) waitForDiscoveryRun(ctx context.Context, runID string) (*DiscoveryRun, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		run, err := a.store.GetDiscoveryRun(ctx, runID)
		if err != nil {
			return nil, err
		}
		switch run.Status {
		case StatusCompleted:
			return run, nil
		case StatusFailed, StatusError, StatusCanceled:
			return run, fmt.Errorf("discovery run ended with status %s: %s", run.Status, RedactSecrets(run.ErrorMessage))
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (a *App) waitForTestPlanExecution(ctx context.Context, executionID string) (*TestPlanExecutionDetail, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		detail, err := a.store.GetTestPlanExecution(ctx, executionID)
		if err != nil {
			return nil, err
		}
		switch detail.Execution.Status {
		case StatusCompleted, StatusFailed, StatusError, StatusCanceled:
			return detail, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func qaRunSummary(discoveryRun *DiscoveryRun, plan *TestPlan, execution *TestPlanExecutionDetail, preview *TestPlanExecutionPreview, execute bool, maxScenarios int) map[string]any {
	summary := map[string]any{
		"execute_requested":             execute,
		"max_scenarios":                 maxScenarios,
		"safe_execution":                true,
		"forms_submitted":               false,
		"destructive_actions":           false,
		"autonomous_ai_browser_control": false,
		"credentials_sent_to_ai":        false,
		"browser_storage_exposed_to_ai": false,
		"execution_coverage":            TestPlanCoverageFromPreview(preview),
	}
	if discoveryRun != nil {
		summary["discovery_run_id"] = discoveryRun.ID
		summary["discovery_status"] = discoveryRun.Status
		summary["discovery_total_pages"] = discoveryRun.TotalPages
		summary["discovery_total_links"] = discoveryRun.TotalLinks
		summary["discovery_total_forms"] = discoveryRun.TotalForms
	}
	if plan != nil {
		summary["test_plan_id"] = plan.ID
		summary["test_plan_status"] = plan.Status
		summary["test_plan_total_scenarios"] = plan.TotalScenarios
		summary["test_plan_risk_level"] = plan.RiskLevel
	}
	if execution != nil {
		summary["test_plan_execution_id"] = execution.Execution.ID
		summary["test_plan_execution_status"] = execution.Execution.Status
		summary["passed_scenarios"] = execution.Execution.PassedScenarios
		summary["failed_scenarios"] = execution.Execution.FailedScenarios
		summary["skipped_scenarios"] = execution.Execution.SkippedScenarios
	}
	return summary
}

func summaryInt(summary map[string]any, key string, fallback int) int {
	if summary == nil {
		return fallback
	}
	switch value := summary[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return fallback
	}
}

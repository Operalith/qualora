package main

import (
	"context"
	"encoding/json"
	"fmt"
)

func (a *App) generateAITestPlan(ctx context.Context, project Project, input AITestPlanRequest) (*TestPlan, error) {
	normalized, err := NormalizeAITestPlanRequest(input)
	if err != nil {
		return nil, err
	}

	report, err := a.reportForTestPlan(ctx, project.ID, normalized.RunID)
	if err != nil {
		return nil, err
	}

	discoveryRunID, discoveryReport, err := a.discoveryReportForTestPlan(ctx, project.ID, normalized)
	if err != nil {
		return nil, err
	}

	provider, err := a.providerForAnalysis(ctx, normalized.ProviderID)
	if err != nil {
		return nil, err
	}

	runID := ""
	if report != nil {
		runID = report.RunID
	}
	sourceType := TestPlanSourceRunReport
	if discoveryRunID != "" {
		sourceType = TestPlanSourceDiscovery
	}
	plan, err := a.store.CreateTestPlan(ctx, project.ID, runID, discoveryRunID, sourceType, provider.ID, provider.Model)
	if err != nil {
		return nil, err
	}

	userPrompt, err := BuildAITestPlanUserPrompt(project, report, discoveryReport, normalized)
	if err != nil {
		failed, _ := a.store.FailTestPlan(ctx, plan.ID, RedactSecrets(err.Error()))
		return failed, err
	}
	clientRequest, err := a.clientRequestForProvider(*provider, []AIChatMessage{
		{Role: "system", Content: AITestPlanSystemPrompt()},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		failed, _ := a.store.FailTestPlan(ctx, plan.ID, RedactSecrets(err.Error()))
		return failed, err
	}

	clientResponse, err := a.aiClient.ChatCompletion(ctx, clientRequest)
	if err != nil {
		failed, _ := a.store.FailTestPlan(ctx, plan.ID, RedactSecrets(err.Error()))
		return failed, err
	}
	payload, planJSON, err := ParseTestPlanPayload(clientResponse.Content, normalized.MaxScenarios)
	if err != nil {
		failed, _ := a.store.FailTestPlan(ctx, plan.ID, RedactSecrets(err.Error()))
		return failed, err
	}
	if discoveryRunID != "" {
		TagDiscoveryGeneratedTestPlan(payload, discoveryRunID, normalized.ExecutionMode)
		rawJSON, err := json.Marshal(payload)
		if err != nil {
			failed, _ := a.store.FailTestPlan(ctx, plan.ID, RedactSecrets(err.Error()))
			return failed, err
		}
		if err := json.Unmarshal(rawJSON, &planJSON); err != nil {
			failed, _ := a.store.FailTestPlan(ctx, plan.ID, RedactSecrets(err.Error()))
			return failed, err
		}
	}

	completed, err := a.store.CompleteTestPlan(ctx, plan.ID, payload, planJSON)
	if err != nil {
		return nil, err
	}
	return a.updateGeneratedTestPlanCoverage(ctx, *completed, project, normalized)
}

func (a *App) discoveryReportForTestPlan(ctx context.Context, projectID string, input AITestPlanRequest) (string, *DiscoveryReport, error) {
	if input.DiscoveryRunID == "" && !input.UseLatestDiscovery {
		return "", nil, nil
	}

	var run *DiscoveryRun
	var err error
	if input.DiscoveryRunID != "" {
		run, err = a.store.GetDiscoveryRun(ctx, input.DiscoveryRunID)
	} else {
		run, err = a.store.GetLatestCompletedDiscoveryRun(ctx, projectID)
	}
	if err != nil {
		return "", nil, err
	}
	if run.ProjectID != projectID {
		return "", nil, fmt.Errorf("selected discovery run does not belong to the project")
	}
	if run.Status != StatusCompleted {
		return "", nil, fmt.Errorf("selected discovery run must be completed before AI test generation")
	}
	if input.IncludeDiscoveryMap == nil || !*input.IncludeDiscoveryMap {
		return run.ID, nil, nil
	}
	report, err := a.store.GetDiscoveryReport(ctx, run.ID)
	if err != nil {
		return "", nil, err
	}
	return run.ID, report, nil
}

func (a *App) updateGeneratedTestPlanCoverage(ctx context.Context, plan TestPlan, project Project, input AITestPlanRequest) (*TestPlan, error) {
	if project.FrontendURL == "" {
		return &plan, nil
	}
	preview, err := BuildTestPlanExecutionPreview(plan, project, TestPlanExecutionRequest{
		DryRun:       true,
		MaxScenarios: input.MaxScenarios,
	})
	if err != nil {
		return &plan, nil
	}
	updated, err := a.store.UpdateTestPlanCoverage(ctx, plan.ID, TestPlanCoverageFromPreview(preview))
	if err != nil {
		return &plan, nil
	}
	return updated, nil
}

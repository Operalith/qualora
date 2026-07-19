package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	CIRunStatusPassed  = "passed"
	CIRunStatusFailed  = "failed"
	CIRunStatusWarning = "warning"
	CIRunStatusRunning = "running"
	CIRunStatusError   = "error"

	defaultCIRunTimeoutSeconds = 900
	maxCIRunTimeoutSeconds     = 3600
)

func NormalizeCIRunRequest(project Project, input CIRunRequest) (CIRunRequest, error) {
	input.Mode = strings.ToLower(strings.TrimSpace(input.Mode))
	if input.Mode == "" {
		input.Mode = ReportTypeSafeQA
	}
	if input.Mode != ReportTypeSafeQA {
		return input, fmt.Errorf("mode must be safe_qa")
	}
	input.BaselineID = strings.TrimSpace(input.BaselineID)
	input.StartURL = strings.TrimSpace(input.StartURL)
	input.CredentialProfileID = strings.TrimSpace(input.CredentialProfileID)
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.IssueExportConfigID = strings.TrimSpace(input.IssueExportConfigID)

	if input.UseLatestBaseline == nil {
		value := true
		input.UseLatestBaseline = &value
	}
	if input.RunSafeQA == nil {
		value := true
		input.RunSafeQA = &value
	}
	if input.UseLatestDiscovery == nil {
		value := true
		input.UseLatestDiscovery = &value
	}
	if input.IncludeQualityChecks == nil {
		value := true
		input.IncludeQualityChecks = &value
	}
	if input.ExecuteSafePlan == nil {
		value := true
		input.ExecuteSafePlan = &value
	}
	if input.Wait == nil {
		value := true
		input.Wait = &value
	}
	if input.IssueExportDryRun == nil {
		value := true
		input.IssueExportDryRun = &value
	}
	if input.TimeoutSeconds == 0 {
		input.TimeoutSeconds = defaultCIRunTimeoutSeconds
	}
	if input.TimeoutSeconds < 30 || input.TimeoutSeconds > maxCIRunTimeoutSeconds {
		return input, fmt.Errorf("timeout_seconds must be between 30 and %d", maxCIRunTimeoutSeconds)
	}
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
	if *input.RunSafeQA && project.FrontendURL == "" {
		return input, fmt.Errorf("project frontend_url is required when run_safe_qa is true")
	}
	return input, nil
}

func (a *App) executeCIRun(ctx context.Context, ciRunID string, project Project, input CIRunRequest) (*CIRunResponse, error) {
	now := time.Now().UTC()
	summary := map[string]any{
		"mode":                          input.Mode,
		"run_safe_qa":                   *input.RunSafeQA,
		"use_latest_baseline":           *input.UseLatestBaseline,
		"include_quality_checks":        *input.IncludeQualityChecks,
		"include_safe_explorer":         input.IncludeSafeExplorer,
		"execute_safe_plan":             *input.ExecuteSafePlan,
		"ai_required_for_new_safe_qa":   *input.RunSafeQA,
		"autonomous_ai_browser_control": false,
		"destructive_actions":           false,
		"issue_export_requested":        input.ExportIssues,
		"issue_export_dry_run":          *input.IssueExportDryRun,
	}

	var qaRun *QARun
	var err error
	if *input.RunSafeQA {
		qaInput := QARunRequest{
			Mode:                 "safe",
			StartURL:             input.StartURL,
			CredentialProfileID:  input.CredentialProfileID,
			ProviderID:           input.ProviderID,
			UseLatestDiscovery:   *input.UseLatestDiscovery,
			MaxPages:             input.MaxPages,
			MaxDepth:             input.MaxDepth,
			MaxScenarios:         input.MaxScenarios,
			IncludeQualityChecks: input.IncludeQualityChecks,
			QualityMaxPages:      minInt(input.MaxPages, defaultQualityMaxPages),
			Execute:              *input.ExecuteSafePlan,
			FocusAreas:           []string{"smoke", "functional", "regression"},
			ProductContext:       "Qualora CI run. Secrets, cookies, storage, auth headers, screenshots, full HTML, request bodies, and response bodies are excluded from AI inputs.",
		}
		qualityEnabled := true
		qaInput.QualityIncludeSecurity = &qualityEnabled
		qaInput.QualityIncludeAccessibility = &qualityEnabled
		qaInput.QualityIncludePerformance = &qualityEnabled
		if _, err := a.providerForAnalysis(ctx, input.ProviderID); err != nil {
			return a.failCIRun(ctx, ciRunID, summary, "safe QA CI run requires a configured OpenAI-compatible provider when run_safe_qa is true", err)
		}
		normalized, err := NormalizeQARunRequest(project, qaInput)
		if err != nil {
			return a.failCIRun(ctx, ciRunID, summary, err.Error(), err)
		}
		if normalized.CredentialProfileID != "" {
			profile, err := a.store.GetCredentialProfile(ctx, normalized.CredentialProfileID)
			if err != nil {
				return a.failCIRun(ctx, ciRunID, summary, "credential profile could not be loaded", err)
			}
			if profile.ProjectID != project.ID {
				return a.failCIRun(ctx, ciRunID, summary, "credential profile does not belong to the project", fmt.Errorf("credential profile project mismatch"))
			}
		}
		qaRun, err = a.store.CreateQARun(ctx, project.ID, normalized)
		if err != nil {
			return a.failCIRun(ctx, ciRunID, summary, "QA run could not be created", err)
		}
		summary["qa_run_id"] = qaRun.ID
		if err := a.executeSafeQARun(ctx, qaRun.ID, project, normalized); err != nil {
			_, _ = a.store.FailQARun(ctx, qaRun.ID, err.Error(), map[string]any{"error": RedactSecrets(err.Error())})
			return a.failCIRun(ctx, ciRunID, summary, "safe QA run failed", err)
		}
		qaRun, err = a.store.GetQARun(ctx, qaRun.ID)
		if err != nil {
			return a.failCIRun(ctx, ciRunID, summary, "completed QA run could not be loaded", err)
		}
	} else {
		qaRun, err = a.store.GetLatestCompletedQARun(ctx, project.ID)
		if err != nil {
			return a.failCIRun(ctx, ciRunID, summary, "no completed Safe QA report is available for AI-free CI gate evaluation", err)
		}
		summary["qa_run_id"] = qaRun.ID
		summary["reused_latest_safe_qa"] = true
	}

	snapshot, err := a.store.GetReportSnapshot(ctx, ReportTypeSafeQA, qaRun.ID)
	if err != nil {
		return a.failCIRun(ctx, ciRunID, summary, "Safe QA report snapshot could not be built", err)
	}

	reportURL := fmt.Sprintf("/api/v1/qa-runs/%s/report", qaRun.ID)
	htmlReportURL := fmt.Sprintf("/api/v1/qa-runs/%s/report.html", qaRun.ID)
	summary["report_url"] = reportURL
	summary["html_report_url"] = htmlReportURL

	var comparison *ReportComparison
	var gate QualityGateResult
	var baselineID string
	baseline, err := a.ciRunBaseline(ctx, project.ID, input)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return a.failCIRun(ctx, ciRunID, summary, "baseline could not be loaded", err)
		}
		gate = MissingBaselineQualityGateResult(input.GateConfig, now)
	} else {
		baselineID = baseline.ID
		created := CompareReportToBaseline(snapshot, *baseline, now)
		comparison = &created
		gate = EvaluateQualityGate(created, snapshot.Intelligence.SeverityCounts, snapshot.Status, input.GateConfig, now)
		summary["baseline_id"] = baseline.ID
		summary["comparison_status"] = created.Status
	}
	summary["quality_gate_status"] = gate.Status
	summary["quality_gate_exit_code"] = gate.CIExitCode
	summary["new_critical"] = gate.ComparisonSummary.NewCritical
	summary["new_high"] = gate.ComparisonSummary.NewHigh
	summary["failed_rules"] = gate.FailedRules

	var issueResult *IssueExportResult
	issueStatus := ""
	if input.ExportIssues {
		exportRequest := IssueExportRequest{
			IssueExportConfigID: input.IssueExportConfigID,
			SeverityThreshold:   "high",
			MaxIssues:           10,
			TitlePrefix:         "[Qualora]",
		}
		exportRequest.DryRun = input.IssueExportDryRun
		result, err := a.exportIssuesForSnapshot(ctx, ReportTypeSafeQA, qaRun.ID, snapshot, exportRequest)
		if err != nil {
			return a.failCIRun(ctx, ciRunID, summary, "issue export failed", err)
		}
		issueResult = result
		issueStatus = result.Status
		summary["issue_export_status"] = result.Status
		summary["issues_created"] = result.CreatedCount
	}

	status := ciStatusFromGate(gate.Status)
	completed, err := a.store.CompleteCIRun(ctx, ciRunID, CompleteCIRunInput{
		QARunID:           qaRun.ID,
		BaselineID:        baselineID,
		Status:            status,
		ExitCode:          gate.CIExitCode,
		GateStatus:        gate.Status,
		ComparisonStatus:  comparisonStatus(comparison),
		ReportURL:         reportURL,
		HTMLReportURL:     htmlReportURL,
		IssueExportStatus: issueStatus,
		Summary:           summary,
	})
	if err != nil {
		return nil, err
	}
	response := ciRunResponse(*completed, comparison, &gate, issueResult)
	return &response, nil
}

func (a *App) ciRunBaseline(ctx context.Context, projectID string, input CIRunRequest) (*ReportBaseline, error) {
	if input.BaselineID != "" {
		return a.store.GetReportBaseline(ctx, input.BaselineID)
	}
	if input.UseLatestBaseline != nil && *input.UseLatestBaseline {
		return a.store.GetDefaultReportBaseline(ctx, projectID, ReportTypeSafeQA)
	}
	return nil, ErrNotFound
}

func (a *App) failCIRun(ctx context.Context, ciRunID string, summary map[string]any, message string, cause error) (*CIRunResponse, error) {
	if cause != nil {
		summary["error"] = RedactSecrets(cause.Error())
	}
	failed, err := a.store.FailCIRun(ctx, ciRunID, message, summary)
	if err != nil {
		return nil, err
	}
	response := ciRunResponse(*failed, nil, nil, nil)
	return &response, cause
}

func ciStatusFromGate(status string) string {
	switch status {
	case QualityGateStatusPassed:
		return CIRunStatusPassed
	case QualityGateStatusWarning:
		return CIRunStatusWarning
	case QualityGateStatusFailed:
		return CIRunStatusFailed
	default:
		return CIRunStatusError
	}
}

func comparisonStatus(comparison *ReportComparison) string {
	if comparison == nil {
		return ""
	}
	return comparison.Status
}

func ciRunResponse(run CIRun, comparison *ReportComparison, gate *QualityGateResult, issueResult *IssueExportResult) CIRunResponse {
	summaryText := run.Status
	if gate != nil {
		summaryText = fmt.Sprintf("%s: %d new critical, %d new high, %d failed rules", gate.Status, gate.ComparisonSummary.NewCritical, gate.ComparisonSummary.NewHigh, len(gate.FailedRules))
	}
	response := CIRunResponse{
		CIRunID:            run.ID,
		ProjectID:          run.ProjectID,
		Status:             run.Status,
		QARunID:            run.QARunID,
		ReportURL:          run.ReportURL,
		HTMLReportURL:      run.HTMLReportURL,
		BaselineID:         run.BaselineID,
		QualityGateResult:  gate,
		IssueExportSummary: issueResult,
		ExitCode:           run.ExitCode,
		Summary:            summaryText,
		CreatedAt:          run.CreatedAt,
		CompletedAt:        run.CompletedAt,
		ErrorMessage:       run.ErrorMessage,
	}
	if comparison != nil {
		response.ComparisonSummary = &comparison.Summary
	}
	return response
}

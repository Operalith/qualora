package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *Store) GetQARunReport(ctx context.Context, id string) (*QARunReport, error) {
	run, err := s.GetQARun(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.GetProject(ctx, run.ProjectID)
	if err != nil {
		return nil, err
	}

	var discoveryRun *DiscoveryRun
	var discoverySummary *DiscoverySummary
	findings := []Finding{}
	evidence := []Evidence{}
	if run.DiscoveryRunID != "" {
		discoveryReport, err := s.GetDiscoveryReport(ctx, run.DiscoveryRunID)
		if err != nil {
			return nil, err
		}
		discoveryRun = &discoveryReport.Run
		discoverySummary = &discoveryReport.Summary
		findings = append(findings, discoveryReport.Findings...)
		evidence = append(evidence, discoveryReport.Evidence...)
	}

	var qualityRun *QualityCheckRun
	var qualitySummary *QualityCheckSummary
	qualityResults := []QualityCheckResult{}
	if run.QualityCheckRunID != "" {
		qualityReport, err := s.GetQualityCheckReport(ctx, run.QualityCheckRunID)
		if err != nil {
			return nil, err
		}
		qualityRun = &qualityReport.Run
		qualitySummary = &qualityReport.Summary
		qualityResults = qualityReport.Results
		findings = append(findings, qualityReport.Findings...)
	}

	var apiSmokeRun *TestRun
	var apiSpec *APISpec
	var apiAuth *APIAuthSummary
	var apiSummary *APISmokeSummary
	apiResults := []APICheckResult{}
	if run.APISmokeRunID != "" {
		apiReport, err := s.GetReport(ctx, run.APISmokeRunID)
		if err != nil {
			return nil, err
		}
		apiSmokeRun, err = s.GetRun(ctx, run.APISmokeRunID)
		if err != nil {
			return nil, err
		}
		apiSpec = apiReport.APISpec
		apiAuth = apiReport.APIAuth
		apiSummary = apiReport.APISummary
		apiResults = apiReport.APIResults
		findings = append(findings, apiReport.Findings...)
		evidence = append(evidence, apiReport.Evidence...)
	}

	var plan *TestPlan
	var preview *TestPlanExecutionPreview
	if run.TestPlanID != "" {
		plan, err = s.GetTestPlan(ctx, run.TestPlanID)
		if err != nil {
			return nil, err
		}
		preview, _ = BuildTestPlanExecutionPreview(*plan, *project, TestPlanExecutionRequest{
			DryRun:       true,
			MaxScenarios: summaryInt(run.Summary, "max_scenarios", defaultMaxTestPlanScenarios),
		})
	}

	var executionReport *TestPlanExecutionReport
	if run.TestPlanExecutionID != "" {
		executionReport, err = s.GetTestPlanExecutionReport(ctx, run.TestPlanExecutionID)
		if err != nil {
			return nil, err
		}
		findings = append(findings, executionReport.Findings...)
		evidence = append(evidence, executionReport.Evidence...)
	}

	report := &QARunReport{
		Run:              *run,
		Project:          *project,
		DiscoveryRun:     discoveryRun,
		DiscoverySummary: discoverySummary,
		QualityCheckRun:  qualityRun,
		QualitySummary:   qualitySummary,
		QualityResults:   qualityResults,
		APISmokeRun:      apiSmokeRun,
		APISpec:          apiSpec,
		APIAuth:          apiAuth,
		APISummary:       apiSummary,
		APIResults:       apiResults,
		TestPlan:         plan,
		ExecutionPreview: preview,
		ExecutionReport:  executionReport,
		Findings:         findings,
		Evidence:         evidence,
		SafetyNotes:      qaRunSafetyNotes(),
		Limitations:      qaRunLimitations(),
		GeneratedAt:      time.Now().UTC(),
	}
	report.ReportIntelligence = BuildReportIntelligence(ReportIntelligenceInput{
		ReportType:        "safe_qa_run",
		ReportID:          run.ID,
		Status:            run.Status,
		Project:           project,
		Findings:          findings,
		Evidence:          evidence,
		ChecksCompleted:   qaRunCompletedChecks(report),
		ChecksSkipped:     qaRunSkippedChecks(report),
		WhatWasTested:     qaRunWhatWasTested(report),
		WhatWasNotTested:  defaultWhatWasNotTested("safe_qa_run"),
		SafetyLimitations: report.Limitations,
	})
	if baseline, err := s.GetDefaultReportBaseline(ctx, project.ID, ReportTypeSafeQA); err == nil {
		snapshot := ReportSnapshot{
			ProjectID:    project.ID,
			ReportType:   ReportTypeSafeQA,
			ReportID:     run.ID,
			SourceRunID:  run.ID,
			Status:       run.Status,
			Intelligence: report.ReportIntelligence,
		}
		comparison := CompareReportToBaseline(snapshot, *baseline, time.Now().UTC())
		gate := EvaluateQualityGate(comparison, report.ReportIntelligence.SeverityCounts, run.Status, QualityGateConfig{}, time.Now().UTC())
		report.Baseline = baseline
		report.Comparison = &comparison
		report.QualityGate = &gate
	} else if errors.Is(err, ErrNotFound) {
		report.BaselineMessage = "No baseline configured yet. Mark this report as baseline to enable regression tracking."
	} else {
		return nil, err
	}
	return report, nil
}

func qaRunSafetyNotes() []string {
	return []string{
		"Safe QA runs use deterministic discovery, sanitized AI test generation, and the approved safe browser DSL only.",
		"AI receives sanitized metadata only; credentials, cookies, local storage, session storage, authorization headers, full HTML, screenshots, request bodies, and response bodies are not sent.",
		"Generated plans are previewed and filtered before execution. Unsupported, unsafe, authenticated, destructive, mutating, submit/upload/admin, exploit, and out-of-scope steps are skipped.",
		"Discovery defaults to same-origin navigation and allowed-host enforcement, and does not submit forms or click arbitrary buttons.",
		"Included quality checks are passive, deterministic, metadata-only checks and do not submit forms, fuzz inputs, or run active security payloads.",
	}
}

func qaRunLimitations() []string {
	return []string{
		"Safe QA runs are alpha coverage and are not exhaustive crawling or full regression automation.",
		"Only the supported deterministic browser DSL subset can execute from AI-generated plans.",
		"Authenticated page discovery can use configured credential profiles, but credentials are not sent to AI and login actions are not generated by AI.",
		"Included API smoke checks are read-only, contract validation is lightweight alpha validation, and request/response bodies are not stored.",
		"Authorization checks and browser smoke remain separate flows; this QA run focuses on discovery-aware browser plan generation, safe plan execution, and optional safe API smoke checks.",
		"Quality checks are alpha heuristics for obvious passive security, accessibility, and performance issues, not a full audit.",
	}
}

func qaRunWhatWasTested(report *QARunReport) []string {
	items := []string{"Application discovery", "Optional passive quality checks", "AI-assisted test plan generation from sanitized metadata", "Approved safe browser DSL execution when enabled"}
	if report != nil && report.APISmokeRun != nil {
		items = append(items, "Optional safe API smoke and lightweight OpenAPI contract validation")
	}
	return items
}

func qaRunHTMLTitle(report *QARunReport) string {
	if report == nil {
		return "Qualora safe QA report"
	}
	return fmt.Sprintf("Qualora safe QA report - %s", report.Project.Name)
}

package main

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	defaultAIBrowserMaxSteps = 8
	defaultAIBrowserMaxDepth = 2
	maxAIBrowserMaxSteps     = 30
	maxAIBrowserMaxDepth     = 5
)

func NormalizeAIBrowserControlRunRequest(project Project, input AIBrowserControlRunRequest) (AIBrowserControlRunRequest, error) {
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.CredentialProfileID = strings.TrimSpace(input.CredentialProfileID)
	input.StartURL = strings.TrimSpace(input.StartURL)
	input.Goal = boundedCleanText(input.Goal, 800)
	if input.ProviderID == "" {
		return input, fmt.Errorf("provider_id is required")
	}
	if input.StartURL == "" {
		input.StartURL = project.FrontendURL
	}
	if input.StartURL == "" {
		return input, fmt.Errorf("project must have frontend_url or request start_url")
	}
	if input.SameOriginOnly == nil {
		value := true
		input.SameOriginOnly = &value
	}
	if input.MaxSteps == 0 {
		input.MaxSteps = defaultAIBrowserMaxSteps
	}
	if input.MaxDepth == 0 {
		input.MaxDepth = defaultAIBrowserMaxDepth
	}
	if input.MaxSteps < 1 || input.MaxSteps > maxAIBrowserMaxSteps {
		return input, fmt.Errorf("max_steps must be between 1 and %d", maxAIBrowserMaxSteps)
	}
	if input.MaxDepth < 0 || input.MaxDepth > maxAIBrowserMaxDepth {
		return input, fmt.Errorf("max_depth must be between 0 and %d", maxAIBrowserMaxDepth)
	}

	target, err := ValidateTargetURL(input.StartURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return input, fmt.Errorf("start_url: %w", err)
	}
	target.Fragment = ""
	if *input.SameOriginOnly {
		if project.FrontendURL == "" {
			return input, fmt.Errorf("same_origin_only AI Browser Control requires project frontend_url")
		}
		frontend, err := url.Parse(project.FrontendURL)
		if err != nil {
			return input, fmt.Errorf("project frontend_url is invalid")
		}
		if target.Scheme != frontend.Scheme || !sameHostPort(target, frontend) {
			return input, fmt.Errorf("start_url must stay on the project frontend origin when same_origin_only is true")
		}
	}
	input.StartURL = RedactSensitiveURLQuery(target).String()
	return input, nil
}

func summarizeAIBrowserControl(run AIBrowserControlRun, steps []AIBrowserControlStep, findings []Finding, evidence []Evidence) AIBrowserControlSummary {
	summary := AIBrowserControlSummary{
		TotalSteps:         run.TotalSteps,
		TotalAISuggestions: run.TotalAISuggestions,
		ActionsApproved:    run.TotalActionsApproved,
		ActionsExecuted:    run.TotalActionsExecuted,
		ActionsSkipped:     run.TotalActionsSkipped,
		PolicyBlocks:       run.TotalPolicyBlocks,
		Findings:           run.TotalFindings,
	}
	if summary.TotalSteps == 0 {
		summary.TotalSteps = len(steps)
	}
	if summary.Findings == 0 {
		summary.Findings = len(findings)
	}
	for _, step := range steps {
		if step.ScreenshotEvidenceID != "" {
			summary.Screenshots++
		}
		summary.ConsoleErrors += step.ConsoleErrorCount
		summary.FailedRequests += step.FailedRequestCount
		if run.TotalAISuggestions == 0 && len(step.AISuggestion) > 0 {
			summary.TotalAISuggestions++
		}
		if run.TotalActionsApproved == 0 && step.PolicyDecision == "approved" {
			summary.ActionsApproved++
		}
		if run.TotalActionsExecuted == 0 && step.ExecutionStatus == "executed" {
			summary.ActionsExecuted++
		}
		if run.TotalActionsSkipped == 0 && step.ExecutionStatus == "skipped" {
			summary.ActionsSkipped++
		}
		if run.TotalPolicyBlocks == 0 && step.PolicyDecision == "blocked" {
			summary.PolicyBlocks++
		}
	}
	if summary.Screenshots == 0 {
		for _, item := range evidence {
			if item.Type == "screenshot" {
				summary.Screenshots++
			}
		}
	}
	return summary
}

func aiBrowserControlSafetyNotes() []string {
	return []string{
		"AI Browser Control is policy-gated: AI proposes one typed action, Qualora validates it, and Playwright executes only approved safe actions.",
		"AI never directly controls the browser and cannot execute arbitrary Playwright commands.",
		"Credentials, cookies, tokens, authorization headers, browser storage, full HTML, screenshots, request bodies, and response bodies are never sent to AI.",
		"Form submission, destructive actions, mutating actions, payloads, fuzzing, active scanning, and external navigation are blocked by default.",
	}
}

func aiBrowserControlLimitations() []string {
	return []string{
		"AI Browser Control is alpha and conservative.",
		"Only supported safe actions are executed: safe navigation, simple assertions, screenshot capture, signal collection, and stop.",
		"Coverage is bounded by max_steps, max_depth, allowed hosts, same-origin policy, and observed safe candidates.",
		"Policy-gated AI Browser Control does not replace human review, deterministic regression checks, or security testing.",
	}
}

func sanitizeAIBrowserControlSettings(run AIBrowserControlRun) map[string]any {
	return map[string]any{
		"start_url":                     run.StartURL,
		"goal":                          run.Goal,
		"provider_id":                   run.ProviderID,
		"credential_profile_id":         run.CredentialProfileID,
		"max_steps":                     run.MaxSteps,
		"max_depth":                     run.MaxDepth,
		"same_origin_only":              run.SameOriginOnly,
		"ai_direct_browser_control":     false,
		"policy_gate_required":          true,
		"forms_submitted":               false,
		"destructive_actions":           false,
		"credentials_sent_to_ai":        false,
		"browser_storage_exposed_to_ai": false,
	}
}

func boundedCleanText(value string, max int) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(RedactSecrets(value))), " ")
	if len(value) > max {
		value = value[:max]
	}
	return value
}

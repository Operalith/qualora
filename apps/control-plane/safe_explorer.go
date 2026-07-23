package main

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	defaultSafeExplorerMaxSteps = 10
	defaultSafeExplorerMaxDepth = 2
	maxSafeExplorerMaxSteps     = 50
	maxSafeExplorerMaxDepth     = 5
)

func NormalizeSafeExplorerRunRequest(project Project, input SafeExplorerRunRequest) (SafeExplorerRunRequest, error) {
	input.StartURL = strings.TrimSpace(input.StartURL)
	input.CredentialProfileID = strings.TrimSpace(input.CredentialProfileID)
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
		input.MaxSteps = defaultSafeExplorerMaxSteps
	}
	if input.MaxDepth == 0 {
		input.MaxDepth = defaultSafeExplorerMaxDepth
	}
	if input.MaxSteps < 1 || input.MaxSteps > maxSafeExplorerMaxSteps {
		return input, fmt.Errorf("max_steps must be between 1 and %d", maxSafeExplorerMaxSteps)
	}
	if input.MaxDepth < 0 || input.MaxDepth > maxSafeExplorerMaxDepth {
		return input, fmt.Errorf("max_depth must be between 0 and %d", maxSafeExplorerMaxDepth)
	}

	target, err := ValidateTargetURL(input.StartURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return input, fmt.Errorf("start_url: %w", err)
	}
	target.Fragment = ""
	if *input.SameOriginOnly {
		if project.FrontendURL == "" {
			return input, fmt.Errorf("same_origin_only safe explorer requires project frontend_url")
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

func safeExplorerSafetyNotes() []string {
	return []string{
		"Safe Explorer is deterministic and executes only classified safe same-origin navigation actions by default.",
		"POST, PUT, PATCH, DELETE, password, file, hidden sensitive, logout, delete, transfer, payment, and reset actions are skipped with reasons.",
		"GET forms are skipped unless allow_get_forms is explicitly enabled, and even then only simple non-sensitive GET forms are eligible.",
		"Credentials, cookies, local storage, session storage, authorization headers, full HTML, and response bodies are not stored or sent to AI.",
		"AI does not choose or execute Safe Explorer actions.",
	}
}

func safeExplorerLimitations() []string {
	return []string{
		"Safe Explorer is alpha coverage and is not a full crawler or scanner.",
		"Buttons without deterministic safe navigation targets are observed but not clicked.",
		"Client-side flows that require arbitrary form submission, drag-and-drop, uploads, or custom widgets are not executed.",
		"Only metadata, screenshots, actions, skip reasons, and findings are recorded; full page HTML is not stored.",
		"Safe QA Run integration for Safe Explorer summaries is planned for a later release; v0.24.0-alpha exposes Safe Explorer as a standalone workflow.",
	}
}

func sanitizeSafeExplorerSettings(run SafeExplorerRun) map[string]any {
	return map[string]any{
		"start_url":             run.StartURL,
		"credential_profile_id": run.CredentialProfileID,
		"max_steps":             run.MaxSteps,
		"max_depth":             run.MaxDepth,
		"same_origin_only":      run.SameOriginOnly,
		"allow_get_forms":       run.AllowGetForms,
		"safe_actions_only":     true,
		"forms_submitted":       false,
		"destructive_actions":   false,
		"ai_action_selection":   false,
	}
}

func summarizeSafeExplorer(run SafeExplorerRun, steps []SafeExplorerStep, actions []SafeExplorerAction, findings []Finding) SafeExplorerSummary {
	summary := SafeExplorerSummary{
		TotalSteps:           run.TotalSteps,
		TotalPagesObserved:   run.TotalPagesObserved,
		TotalActionsDetected: run.TotalActionsDetected,
		TotalActionsExecuted: run.TotalActionsExecuted,
		TotalActionsSkipped:  run.TotalActionsSkipped,
		TotalFindings:        run.TotalFindings,
	}
	if summary.TotalSteps == 0 {
		summary.TotalSteps = len(steps)
	}
	if summary.TotalActionsDetected == 0 {
		summary.TotalActionsDetected = len(actions)
	}
	if summary.TotalFindings == 0 {
		summary.TotalFindings = len(findings)
	}
	for _, step := range steps {
		if step.ActionDecision == "observed" {
			summary.TotalPagesObserved++
		}
		if step.ScreenshotEvidenceID != "" {
			summary.PagesWithScreenshots++
		}
	}
	if run.TotalPagesObserved > 0 {
		summary.TotalPagesObserved = run.TotalPagesObserved
	}
	for _, action := range actions {
		if action.Safety == "safe" {
			summary.SafeActions++
		}
		if action.Decision == "execute" && run.TotalActionsExecuted == 0 {
			summary.TotalActionsExecuted++
		}
		if action.Decision == "skip" && run.TotalActionsSkipped == 0 {
			summary.TotalActionsSkipped++
		}
		switch action.SkipReason {
		case "external_action_skipped", "host_not_allowed":
			summary.ExternalActions++
		case "unsafe_action_skipped", "sensitive_query_skipped", "duplicate_url":
			summary.UnsafeActionsSkipped++
		case "unsupported_action", "unsupported_scheme", "button_without_safe_navigation", "form_method_not_safe", "get_forms_disabled":
			summary.UnsupportedActions++
		}
	}
	return summary
}

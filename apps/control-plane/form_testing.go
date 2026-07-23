package main

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	defaultFormTestMaxForms        = 10
	maxFormTestMaxForms            = 50
	defaultFormTestMaxTestsPerForm = 1
	maxFormTestMaxTestsPerForm     = 5
)

func NormalizeFormTestRunRequest(project Project, input FormTestRunRequest) (FormTestRunRequest, error) {
	input.DiscoveryRunID = strings.TrimSpace(input.DiscoveryRunID)
	input.CredentialProfileID = strings.TrimSpace(input.CredentialProfileID)
	input.TargetURL = strings.TrimSpace(input.TargetURL)
	if input.TargetURL == "" {
		input.TargetURL = project.FrontendURL
	}
	if input.TargetURL == "" {
		return input, fmt.Errorf("project must have frontend_url or request target_url")
	}
	if input.MaxForms == 0 {
		input.MaxForms = defaultFormTestMaxForms
	}
	if input.MaxTestsPerForm == 0 {
		input.MaxTestsPerForm = defaultFormTestMaxTestsPerForm
	}
	if input.MaxForms < 1 || input.MaxForms > maxFormTestMaxForms {
		return input, fmt.Errorf("max_forms must be between 1 and %d", maxFormTestMaxForms)
	}
	if input.MaxTestsPerForm < 1 || input.MaxTestsPerForm > maxFormTestMaxTestsPerForm {
		return input, fmt.Errorf("max_tests_per_form must be between 1 and %d", maxFormTestMaxTestsPerForm)
	}
	if input.SafeGetOnly == nil {
		value := true
		input.SafeGetOnly = &value
	}

	target, err := ValidateTargetURL(input.TargetURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return input, fmt.Errorf("target_url: %w", err)
	}
	target.Fragment = ""
	if project.FrontendURL != "" {
		frontend, err := url.Parse(project.FrontendURL)
		if err != nil {
			return input, fmt.Errorf("project frontend_url is invalid")
		}
		if target.Scheme != frontend.Scheme || !sameHostPort(target, frontend) {
			return input, fmt.Errorf("target_url must stay on the project frontend origin")
		}
	}
	if hasSensitiveFormTestQuery(target) {
		return input, fmt.Errorf("target_url query contains sensitive parameter names")
	}
	input.TargetURL = RedactSensitiveURLQuery(target).String()
	return input, nil
}

func hasSensitiveFormTestQuery(target *url.URL) bool {
	for name := range target.Query() {
		if isSensitiveQueryName(name) {
			return true
		}
	}
	return false
}

func summarizeFormTest(run FormTestRun, results []FormTestResult, findings []Finding, evidence []Evidence) FormTestSummary {
	summary := FormTestSummary{
		FormsDetected:       run.TotalFormsDetected,
		FormsClassifiedSafe: run.TotalFormsClassifiedSafe,
		FormsTested:         run.TotalFormsTested,
		FormsSkipped:        run.TotalFormsSkipped,
		Findings:            run.TotalFindings,
	}
	if summary.FormsDetected == 0 {
		summary.FormsDetected = len(results)
	}
	if summary.Findings == 0 {
		summary.Findings = len(findings)
	}
	for _, result := range results {
		if run.TotalFormsClassifiedSafe == 0 && result.Safety == "safe" {
			summary.FormsClassifiedSafe++
		}
		if run.TotalFormsTested == 0 && result.Decision == "tested" {
			summary.FormsTested++
		}
		if run.TotalFormsSkipped == 0 && result.Decision == "skipped" {
			summary.FormsSkipped++
		}
		summary.ConsoleErrors += result.ConsoleErrorCount
		summary.FailedRequests += result.FailedRequestCount
		if result.ScreenshotEvidenceID != "" {
			summary.Screenshots++
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

func formTestSafetyNotes() []string {
	return []string{
		"Safe Form Testing is deterministic and only submits simple same-origin GET forms with non-sensitive fields.",
		"POST, PUT, PATCH, DELETE, password, file, hidden sensitive, login, logout, payment, checkout, transfer, delete, reset, admin, upload, and profile forms are skipped with reasons.",
		"Only bounded benign values are used for eligible GET fields; no fuzzing, payload attacks, active scanning, or destructive testing is performed.",
		"Credentials may be used only for selector-based login before testing, and credentials, cookies, tokens, browser storage, auth headers, full HTML, request bodies, and response bodies are not stored or sent to AI.",
		"AI Browser Control may propose a safe GET form action, but the deterministic policy gate is final and Playwright executes only approved actions.",
	}
}

func formTestLimitations() []string {
	return []string{
		"Safe Form Testing is alpha coverage and is not a full form, workflow, accessibility, security, or contract test suite.",
		"Only simple GET search/filter/sort/navigation forms are submitted by default.",
		"Forms requiring JavaScript-only custom widgets, POST requests, authentication secrets, file uploads, CAPTCHAs, payments, or mutating workflows are skipped.",
		"Safe QA Run aggregation for form-test summaries is planned for a later release; v0.24.0-alpha exposes Safe Form Testing as a standalone workflow.",
	}
}

func sanitizeFormTestSettings(run FormTestRun) map[string]any {
	return map[string]any{
		"target_url":                      run.TargetURL,
		"discovery_run_id":                run.DiscoveryRunID,
		"credential_profile_id":           run.CredentialProfileID,
		"max_forms":                       run.MaxForms,
		"max_tests_per_form":              run.MaxTestsPerForm,
		"safe_get_only":                   run.SafeGetOnly,
		"safe_get_forms_only":             true,
		"arbitrary_form_submission":       false,
		"mutating_forms_submitted":        false,
		"destructive_actions":             false,
		"payload_attacks":                 false,
		"credentials_sent_to_ai":          false,
		"browser_storage_exposed_to_ai":   false,
		"request_or_response_bodies_kept": false,
	}
}

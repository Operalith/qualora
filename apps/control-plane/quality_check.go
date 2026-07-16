package main

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	defaultQualityMaxPages = 10
	maxQualityMaxPages     = 50
)

func NormalizeQualityCheckRunRequest(project Project, input QualityCheckRunRequest) (QualityCheckRunRequest, error) {
	input.TargetURL = strings.TrimSpace(input.TargetURL)
	input.CredentialProfileID = strings.TrimSpace(input.CredentialProfileID)
	input.DiscoveryRunID = strings.TrimSpace(input.DiscoveryRunID)
	if input.TargetURL == "" {
		input.TargetURL = project.FrontendURL
	}
	if input.TargetURL == "" {
		return input, fmt.Errorf("project frontend_url or target_url is required for quality checks")
	}
	if input.MaxPages == 0 {
		input.MaxPages = defaultQualityMaxPages
	}
	if input.MaxPages < 1 || input.MaxPages > maxQualityMaxPages {
		return input, fmt.Errorf("max_pages must be between 1 and %d", maxQualityMaxPages)
	}
	if input.IncludeSecurity == nil {
		value := true
		input.IncludeSecurity = &value
	}
	if input.IncludeAccessibility == nil {
		value := true
		input.IncludeAccessibility = &value
	}
	if input.IncludePerformance == nil {
		value := true
		input.IncludePerformance = &value
	}
	if !*input.IncludeSecurity && !*input.IncludeAccessibility && !*input.IncludePerformance {
		return input, fmt.Errorf("at least one quality category must be enabled")
	}
	target, err := ValidateTargetURL(input.TargetURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return input, fmt.Errorf("target_url: %w", err)
	}
	if project.FrontendURL != "" {
		frontend, err := url.Parse(project.FrontendURL)
		if err != nil {
			return input, fmt.Errorf("project frontend_url is invalid")
		}
		if target.Scheme != frontend.Scheme || !sameHostPort(target, frontend) {
			return input, fmt.Errorf("target_url must stay on the project frontend origin")
		}
	}
	if hasSensitiveQualityQuery(target) {
		return input, fmt.Errorf("target_url query contains sensitive parameter names")
	}
	target.Fragment = ""
	input.TargetURL = RedactSensitiveURLQuery(target).String()
	return input, nil
}

func hasSensitiveQualityQuery(target *url.URL) bool {
	for name := range target.Query() {
		if isSensitiveQueryName(name) {
			return true
		}
	}
	return false
}

func summarizeQualityCheckResults(run QualityCheckRun, results []QualityCheckResult) QualityCheckSummary {
	summary := QualityCheckSummary{
		TotalFindings: run.TotalFindings,
		Critical:      run.CriticalFindings,
		High:          run.HighFindings,
		Medium:        run.MediumFindings,
		Low:           run.LowFindings,
		Info:          run.InfoFindings,
		TotalPages:    run.TotalPages,
	}
	if summary.TotalFindings == 0 && len(results) > 0 {
		summary.TotalFindings = len(results)
	}
	for _, result := range results {
		switch result.Category {
		case "security":
			summary.SecurityFindings++
		case "accessibility":
			summary.AccessibilityFindings++
		case "performance":
			summary.PerformanceFindings++
		}
		if run.TotalFindings > 0 {
			continue
		}
		switch result.Severity {
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

func qualityCheckSafetyNotes() []string {
	return []string{
		"Quality checks are passive, read-only, and deterministic.",
		"The browser worker visits only project allowed hosts and same-origin frontend URLs.",
		"Quality checks do not submit forms, click arbitrary buttons, fuzz inputs, run payloads, perform destructive actions, or use AI browser control.",
		"Authenticated quality checks can use a configured credential profile for selector-based login, but credentials, cookies, browser storage, auth headers, and tokens are not stored or sent to AI.",
	}
}

func qualityCheckLimitations() []string {
	return []string{
		"Quality checks are alpha heuristics and are not full security, accessibility, or performance audits.",
		"Security checks are passive header, cookie-flag, mixed-content, source-map, and obvious exposure checks only.",
		"Accessibility checks cover simple document, image, form, link, button, and landmark heuristics; they are not WCAG certification.",
		"Performance checks use page-load metadata and resource observations; Lighthouse/Core Web Vitals are not implemented yet.",
	}
}

package main

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const (
	defaultDiscoveryMaxPages = 20
	defaultDiscoveryMaxDepth = 2
	maxDiscoveryMaxPages     = 100
	maxDiscoveryMaxDepth     = 5
)

func NormalizeDiscoveryRunRequest(project Project, input DiscoveryRunRequest) (DiscoveryRunRequest, error) {
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
	if input.MaxPages == 0 {
		input.MaxPages = defaultDiscoveryMaxPages
	}
	if input.MaxDepth == 0 {
		input.MaxDepth = defaultDiscoveryMaxDepth
	}
	if input.MaxPages < 1 || input.MaxPages > maxDiscoveryMaxPages {
		return input, fmt.Errorf("max_pages must be between 1 and %d", maxDiscoveryMaxPages)
	}
	if input.MaxDepth < 0 || input.MaxDepth > maxDiscoveryMaxDepth {
		return input, fmt.Errorf("max_depth must be between 0 and %d", maxDiscoveryMaxDepth)
	}

	target, err := ValidateTargetURL(input.StartURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return input, fmt.Errorf("start_url: %w", err)
	}
	target.Fragment = ""
	if *input.SameOriginOnly {
		if project.FrontendURL == "" {
			return input, fmt.Errorf("same_origin_only discovery requires project frontend_url")
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

func RedactSensitiveURLQuery(parsed *url.URL) *url.URL {
	copyURL := *parsed
	query := copyURL.Query()
	if len(query) == 0 {
		return &copyURL
	}
	for name, values := range query {
		if isSensitiveQueryName(name) {
			if len(values) == 0 {
				query.Set(name, "[REDACTED]")
			} else {
				query[name] = []string{"[REDACTED]"}
			}
		}
	}
	copyURL.RawQuery = query.Encode()
	return &copyURL
}

func isSensitiveQueryName(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, marker := range []string{
		"access_token",
		"api_key",
		"apikey",
		"auth",
		"authorization",
		"credential",
		"jwt",
		"key",
		"password",
		"passwd",
		"secret",
		"session",
		"token",
	} {
		if normalized == marker || strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func sameHostPort(a *url.URL, b *url.URL) bool {
	return strings.EqualFold(normalizedURLHostPort(a), normalizedURLHostPort(b))
}

func normalizedURLHostPort(parsed *url.URL) string {
	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	if port == "" {
		switch parsed.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}
	return host + ":" + port
}

func discoverySafetyNotes() []string {
	return []string{
		"Discovery visits only allowed hosts and defaults to same-origin navigation.",
		"Discovery does not submit forms, click arbitrary buttons, run payloads, or perform destructive actions.",
		"External, unsupported-scheme, non-HTML, and mutating-looking links are skipped and recorded with skip reasons.",
		"Credentials, cookies, local storage, session storage, authorization headers, full HTML, and screenshot pixels are not sent to AI by default.",
	}
}

func discoveryLimitations() []string {
	return []string{
		"Discovery is deterministic alpha coverage, not exhaustive crawling.",
		"Client-side routes that require arbitrary button clicks or form submissions are not explored.",
		"Authenticated discovery uses configured selector-based login only when a credential profile is selected.",
		"Discovery records metadata and screenshots; it does not save full page HTML or response bodies.",
	}
}

func sanitizeDiscoverySettings(run DiscoveryRun) map[string]any {
	return map[string]any{
		"start_url":             run.StartURL,
		"credential_profile_id": run.CredentialProfileID,
		"max_pages":             run.MaxPages,
		"max_depth":             run.MaxDepth,
		"same_origin_only":      run.SameOriginOnly,
		"safe_links_only":       true,
		"forms_submitted":       false,
		"destructive_actions":   false,
	}
}

func summarizeDiscoveryMap(run DiscoveryRun, pages []DiscoveredPage, links []DiscoveredLink, forms []DiscoveredForm, findings []Finding) DiscoverySummary {
	summary := DiscoverySummary{
		TotalPages:          run.TotalPages,
		TotalLinks:          run.TotalLinks,
		TotalForms:          run.TotalForms,
		TotalConsoleErrors:  run.TotalConsoleErrors,
		TotalFailedRequests: run.TotalFailedRequests,
		TotalFindings:       run.TotalFindings,
	}
	if summary.TotalPages == 0 {
		summary.TotalPages = len(pages)
	}
	if summary.TotalLinks == 0 {
		summary.TotalLinks = len(links)
	}
	if summary.TotalForms == 0 {
		summary.TotalForms = len(forms)
	}
	if summary.TotalFindings == 0 {
		summary.TotalFindings = len(findings)
	}
	for _, page := range pages {
		if page.ScreenshotEvidenceID != "" {
			summary.PagesWithScreenshots++
		}
		if run.TotalConsoleErrors == 0 {
			summary.TotalConsoleErrors += page.ConsoleErrorCount
		}
		if run.TotalFailedRequests == 0 {
			summary.TotalFailedRequests += page.FailedRequestCount
		}
	}
	for _, link := range links {
		if !link.Skipped {
			continue
		}
		summary.SkippedLinks++
		switch link.SkipReason {
		case "external_link_skipped":
			summary.ExternalLinksSkipped++
		case "unsafe_link_skipped":
			summary.UnsafeLinksSkipped++
		}
	}
	return summary
}

func sortedFormFields(fields []DiscoveredFormField) []DiscoveredFormField {
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].CreatedAt.Before(fields[j].CreatedAt)
	})
	return fields
}

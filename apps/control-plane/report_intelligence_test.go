package main

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeFindingSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		category string
		title    string
		ruleID   string
		want     string
	}{
		{name: "preserves critical authorization bypass", severity: "critical", category: "authorization_bypass", want: "critical"},
		{name: "missing csp is medium", category: "security", title: "Missing Content-Security-Policy header", ruleID: "missing_csp", want: "medium"},
		{name: "5xx is high", category: "server_error", title: "Server error while loading page", want: "high"},
		{name: "broken internal link is medium", category: "not_found", title: "404 page found", want: "medium"},
		{name: "external skipped is info", severity: "low", category: "explorer_external_action_skipped", want: "info"},
		{name: "minor accessibility is low", category: "accessibility", ruleID: "images_missing_alt", want: "low"},
		{name: "unknown is low", category: "unexpected_widget", want: "low"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeFindingSeverity(tt.severity, tt.category, tt.title, tt.ruleID)
			if got != tt.want {
				t.Fatalf("NormalizeFindingSeverity() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeFindingURLRedactsSensitiveQuery(t *testing.T) {
	safeURL, path := NormalizeFindingURL("https://example.test/account?token=secret&tab=billing#section")
	if strings.Contains(safeURL, "secret") || strings.Contains(path, "secret") {
		t.Fatalf("expected secret query value to be redacted, got url=%q path=%q", safeURL, path)
	}
	if !strings.Contains(safeURL, "token=%5BREDACTED%5D") || !strings.Contains(path, "token=%5BREDACTED%5D") {
		t.Fatalf("expected redacted token marker, got url=%q path=%q", safeURL, path)
	}
}

func TestBuildReportIntelligenceGroupsRepeatedFindings(t *testing.T) {
	now := time.Now().UTC()
	input := ReportIntelligenceInput{
		ReportType: RunTypeAppDiscovery,
		ReportID:   "discovery-1",
		Status:     StatusCompleted,
		Findings: []Finding{
			{
				ID:          "f1",
				Title:       "Out-of-scope action skipped by Safe Explorer",
				Severity:    "low",
				Category:    "explorer_external_action_skipped",
				Confidence:  "high",
				Description: "External link was skipped.",
				CreatedAt:   now,
			},
			{
				ID:          "f2",
				Title:       "Out-of-scope action skipped by Safe Explorer",
				Severity:    "low",
				Category:    "explorer_external_action_skipped",
				Confidence:  "medium",
				Description: "External link was skipped.",
				CreatedAt:   now,
			},
			{
				ID:          "f3",
				Title:       "Server error while loading page",
				Severity:    "high",
				Category:    "server_error",
				Confidence:  "high",
				Description: "The page returned HTTP 500.",
				CreatedAt:   now,
			},
		},
	}

	intelligence := BuildReportIntelligence(input)
	if intelligence.RawFindingsCount != 3 {
		t.Fatalf("raw count = %d, want 3", intelligence.RawFindingsCount)
	}
	if len(intelligence.GroupedFindings) != 2 {
		t.Fatalf("grouped count = %d, want 2: %#v", len(intelligence.GroupedFindings), intelligence.GroupedFindings)
	}
	if intelligence.DeduplicationSummary.DuplicateFindingsReduced != 1 {
		t.Fatalf("duplicate reduction = %d, want 1", intelligence.DeduplicationSummary.DuplicateFindingsReduced)
	}
	if intelligence.GroupedFindings[0].NormalizedSeverity != "high" {
		t.Fatalf("top group severity = %q, want high", intelligence.GroupedFindings[0].NormalizedSeverity)
	}
	if intelligence.ExecutiveSummary.OverallStatus != "fail" {
		t.Fatalf("overall status = %q, want fail", intelligence.ExecutiveSummary.OverallStatus)
	}
	if intelligence.NoiseSummary.HighNoise == 0 {
		t.Fatalf("expected repeated external skipped group to be high noise: %#v", intelligence.NoiseSummary)
	}
}

func TestBuildReportIntelligenceMergesQualityResults(t *testing.T) {
	now := time.Now().UTC()
	input := ReportIntelligenceInput{
		ReportType: RunTypeQualityCheck,
		ReportID:   "quality-1",
		Status:     StatusCompleted,
		QualityResults: []QualityCheckResult{
			{
				ID:             "r1",
				RunID:          "quality-1",
				Category:       "security",
				RuleID:         "missing_csp",
				Severity:       "",
				Title:          "Missing Content-Security-Policy header",
				Description:    "CSP was not observed.",
				Recommendation: "Add a restrictive CSP.",
				URL:            "https://example.test/?password=secret",
				CreatedAt:      now,
			},
			{
				ID:             "r2",
				RunID:          "quality-1",
				Category:       "security",
				RuleID:         "missing_csp",
				Severity:       "",
				Title:          "Missing Content-Security-Policy header",
				Description:    "CSP was not observed.",
				Recommendation: "Add a restrictive CSP.",
				URL:            "https://example.test/settings?password=secret",
				CreatedAt:      now,
			},
		},
	}

	intelligence := BuildReportIntelligence(input)
	if len(intelligence.GroupedFindings) != 1 {
		t.Fatalf("grouped count = %d, want 1", len(intelligence.GroupedFindings))
	}
	group := intelligence.GroupedFindings[0]
	if group.NormalizedSeverity != "medium" {
		t.Fatalf("severity = %q, want medium", group.NormalizedSeverity)
	}
	if group.OccurrencesCount != 2 {
		t.Fatalf("occurrences = %d, want 2", group.OccurrencesCount)
	}
	for _, raw := range group.RawOccurrenceRefs {
		if strings.Contains(raw.AffectedURL, "secret") {
			t.Fatalf("raw occurrence leaked sensitive query value: %#v", raw)
		}
	}
}

func TestRenderReportIntelligenceHTML(t *testing.T) {
	intelligence := BuildReportIntelligence(ReportIntelligenceInput{
		ReportType: RunTypeBrowserSmoke,
		ReportID:   "run-1",
		Status:     StatusCompleted,
		Findings: []Finding{{
			ID:         "f1",
			Title:      "Console error detected",
			Severity:   "medium",
			Category:   "frontend",
			Confidence: "medium",
		}},
	})
	html := string(reportIntelligenceHTML(intelligence))
	for _, expected := range []string{"Executive Summary", "Grouped Findings", "Noise / Repeated Findings", "Raw findings"} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected HTML to include %q, got %s", expected, html)
		}
	}
}

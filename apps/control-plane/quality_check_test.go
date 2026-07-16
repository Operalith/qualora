package main

import "testing"

func TestNormalizeQualityCheckRunRequestDefaultsToSafePassiveChecks(t *testing.T) {
	project := Project{
		ID:                  "project-1",
		FrontendURL:         "https://example.com",
		AllowedHosts:        []string{"example.com"},
		AllowPrivateTargets: true,
	}
	req, err := NormalizeQualityCheckRunRequest(project, QualityCheckRunRequest{})
	if err != nil {
		t.Fatalf("normalize quality check request: %v", err)
	}
	if req.TargetURL != "https://example.com" {
		t.Fatalf("unexpected target URL: %q", req.TargetURL)
	}
	if req.MaxPages != defaultQualityMaxPages {
		t.Fatalf("unexpected max pages: %d", req.MaxPages)
	}
	if req.IncludeSecurity == nil || !*req.IncludeSecurity {
		t.Fatal("expected security checks enabled")
	}
	if req.IncludeAccessibility == nil || !*req.IncludeAccessibility {
		t.Fatal("expected accessibility checks enabled")
	}
	if req.IncludePerformance == nil || !*req.IncludePerformance {
		t.Fatal("expected performance checks enabled")
	}
}

func TestNormalizeQualityCheckRunRequestRejectsSensitiveQuery(t *testing.T) {
	project := Project{
		ID:                  "project-1",
		FrontendURL:         "https://example.com",
		AllowedHosts:        []string{"example.com"},
		AllowPrivateTargets: true,
	}
	_, err := NormalizeQualityCheckRunRequest(project, QualityCheckRunRequest{
		TargetURL: "https://example.com/dashboard?token=abc",
	})
	if err == nil {
		t.Fatal("expected sensitive query to be rejected")
	}
}

func TestNormalizeQualityCheckRunRequestRejectsDifferentOrigin(t *testing.T) {
	project := Project{
		ID:                  "project-1",
		FrontendURL:         "https://example.com",
		AllowedHosts:        []string{"example.com", "app.example.net"},
		AllowPrivateTargets: true,
	}
	_, err := NormalizeQualityCheckRunRequest(project, QualityCheckRunRequest{
		TargetURL: "https://app.example.net",
	})
	if err == nil {
		t.Fatal("expected target outside frontend origin to be rejected")
	}
}

func TestSummarizeQualityCheckResultsCountsSeverityAndCategory(t *testing.T) {
	run := QualityCheckRun{TotalPages: 2}
	results := []QualityCheckResult{
		{Severity: "high", Category: "security"},
		{Severity: "medium", Category: "accessibility"},
		{Severity: "low", Category: "performance"},
		{Severity: "info", Category: "security"},
	}
	summary := summarizeQualityCheckResults(run, results)
	if summary.TotalFindings != 4 || summary.High != 1 || summary.Medium != 1 || summary.Low != 1 || summary.Info != 1 {
		t.Fatalf("unexpected severity summary: %#v", summary)
	}
	if summary.SecurityFindings != 2 || summary.AccessibilityFindings != 1 || summary.PerformanceFindings != 1 {
		t.Fatalf("unexpected category summary: %#v", summary)
	}
	if summary.TotalPages != 2 {
		t.Fatalf("unexpected page count: %d", summary.TotalPages)
	}
}

package main

import (
	"strings"
	"testing"
	"time"
)

func TestRenderHTMLReportEscapesContentAndIncludesSummary(t *testing.T) {
	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	project := &Project{
		ID:   "project-1",
		Name: "Example <App>",
	}
	run := &TestRun{
		ID:        "run-1",
		ProjectID: "project-1",
		Status:    StatusCompleted,
		CreatedAt: now,
	}
	report := &Report{
		RunID:     "run-1",
		ProjectID: "project-1",
		Status:    StatusCompleted,
		Summary: ReportSummary{
			TotalFindings: 1,
			High:          1,
		},
		Findings: []Finding{
			{
				Title:          "API endpoint returned 5xx",
				Severity:       "high",
				Category:       "api",
				Confidence:     "high",
				Description:    "GET /status returned HTTP 500.",
				Recommendation: "Inspect service logs.",
			},
		},
		Evidence: []Evidence{
			{
				Type: "screenshot",
				URI:  "s3://qualora-evidence/runs/run-1/screenshots/screen.png",
				Metadata: map[string]any{
					"filename":     "screen.png",
					"content_type": "image/png",
					"size_bytes":   12345,
					"final_url":    "http://demo-web:8080/",
				},
			},
			{
				Type: "browser_observations",
				URI:  "inline://browser-observations",
				Metadata: map[string]any{
					"page_title":       "Qualora Demo Web",
					"body_text_length": 42,
				},
			},
			{
				Type: "api_observations",
				URI:  "inline://api-observations",
				Metadata: map[string]any{
					"checked_endpoints": 2,
				},
			},
		},
		Metadata: map[string]any{
			"jobs": []map[string]string{{"kind": "api", "status": "completed"}},
		},
	}

	var output strings.Builder
	if err := RenderHTMLReport(&output, project, run, report, now); err != nil {
		t.Fatalf("render html report: %v", err)
	}

	html := output.String()
	for _, expected := range []string{
		"Qualora HTML report",
		"Example &lt;App&gt;",
		"API endpoint returned 5xx",
		"checked_endpoints",
		"browser_observations",
		"screen.png",
		"image/png",
		"Inspect service logs.",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected HTML report to contain %q, got:\n%s", expected, html)
		}
	}
	if strings.Contains(html, "Example <App>") {
		t.Fatalf("expected project name to be escaped")
	}
}

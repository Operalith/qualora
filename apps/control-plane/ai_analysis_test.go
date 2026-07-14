package main

import (
	"strings"
	"testing"
)

func TestRedactSecretsMasksCommonSecretPatterns(t *testing.T) {
	input := strings.Join([]string{
		"Authorization: Bearer sk-live-token",
		"password=hunter2",
		"api_key=sk-test-value",
		"refresh_token=refresh-secret",
		"cookie=session=abc123",
		"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.signature",
	}, " ")

	output := RedactSecrets(input)
	for _, leaked := range []string{"sk-live-token", "hunter2", "sk-test-value", "refresh-secret", "abc123", "eyJhbGciOiJIUzI1NiJ9"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("expected %q to be redacted from %q", leaked, output)
		}
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Fatalf("expected redaction marker in %q", output)
	}
}

func TestBuildSafeAIInputStripsSecretsBodiesAndQueryValues(t *testing.T) {
	report := &Report{
		RunID:     "run-1",
		ProjectID: "project-1",
		Status:    StatusCompleted,
		Summary:   ReportSummary{TotalFindings: 1, High: 1},
		Metadata: map[string]any{
			"page_title":    "Demo",
			"authorization": "Bearer should-not-pass",
			"jobs":          []any{map[string]any{"kind": "browser", "status": "completed", "token": "secret-token"}},
		},
		Findings: []Finding{
			{
				Title:       "Secret should not leak",
				Severity:    "high",
				Category:    "browser",
				Confidence:  "high",
				Description: "GET https://example.com/path?access_token=secret returned 500\nfull body omitted",
			},
		},
		Evidence: []Evidence{
			{
				Type: "browser_observations",
				Metadata: map[string]any{
					"target_url":      "https://example.com/path?api_key=secret#frag",
					"final_url":       "https://example.com/final?session_id=secret",
					"response_body":   "very secret body",
					"html":            "<html>secret</html>",
					"console_errors":  []any{"Bearer secret-token"},
					"failed_requests": []any{"https://api.example.com/data?password=secret"},
				},
			},
		},
	}

	input := BuildSafeAIInput(report)
	rendered := prettyJSON(input)
	for _, leaked := range []string{"api_key=secret", "access_token=secret", "session_id=secret", "password=secret", "secret-token", "very secret body", "<html>secret</html>", "should-not-pass"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("safe AI input leaked %q in:\n%s", leaked, rendered)
		}
	}
	if !strings.Contains(rendered, "https://example.com/path") {
		t.Fatalf("expected sanitized target URL to remain, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "?") || strings.Contains(rendered, "#frag") {
		t.Fatalf("expected URL query and fragment to be stripped, got:\n%s", rendered)
	}
}

func TestParseAIAnalysisPayloadValidatesAndSanitizesJSON(t *testing.T) {
	payload, normalized, err := ParseAIAnalysisPayload(`{
		"executive_summary": "Run completed. Bearer secret-token",
		"technical_summary": "One high finding was observed.",
		"risk_level": "HIGH",
		"likely_causes": ["Server error"],
		"recommended_actions": ["Inspect logs for api_key=secret"],
		"suggested_next_tests": ["Retest GET /health"],
		"confidence": 1.5,
		"limitations": ["No screenshots or full bodies were provided"]
	}`)
	if err != nil {
		t.Fatalf("parse AI analysis: %v", err)
	}
	if payload.RiskLevel != "high" {
		t.Fatalf("expected normalized risk level, got %q", payload.RiskLevel)
	}
	if payload.Confidence != 1 {
		t.Fatalf("expected confidence to be clamped to 1, got %f", payload.Confidence)
	}
	if strings.Contains(payload.ExecutiveSummary, "secret-token") {
		t.Fatalf("expected executive summary to be redacted, got %q", payload.ExecutiveSummary)
	}
	if strings.Contains(prettyJSON(normalized), "api_key=secret") {
		t.Fatalf("expected normalized JSON to be redacted, got %s", prettyJSON(normalized))
	}
}

func TestParseAIAnalysisPayloadRejectsInvalidRiskLevel(t *testing.T) {
	_, _, err := ParseAIAnalysisPayload(`{"executive_summary":"x","technical_summary":"y","risk_level":"unknown","likely_causes":[],"recommended_actions":[],"suggested_next_tests":[],"confidence":0.5,"limitations":[]}`)
	if err == nil {
		t.Fatal("expected invalid risk level to be rejected")
	}
}

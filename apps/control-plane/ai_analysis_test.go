package main

import (
	"fmt"
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

func TestBuildSafeDiscoveryAIInputStripsSecretsAndKeepsMapSummary(t *testing.T) {
	status := 200
	report := &DiscoveryReport{
		Run: DiscoveryRun{
			ID:             "discovery-run-id",
			ProjectID:      "project-id",
			Status:         StatusCompleted,
			StartURL:       "https://example.com/?token=should-not-pass",
			MaxPages:       20,
			MaxDepth:       2,
			SameOriginOnly: true,
		},
		Project: Project{ID: "project-id", Name: "Example"},
		Summary: DiscoverySummary{TotalPages: 1, TotalLinks: 1, TotalForms: 1},
		Pages: []DiscoveredPage{{
			Path:                 "/dashboard",
			NormalizedURL:        "https://example.com/dashboard?token=should-not-pass",
			Title:                "Dashboard",
			HTTPStatus:           &status,
			ConsoleErrorCount:    0,
			FailedRequestCount:   0,
			ScreenshotEvidenceID: "evidence-id",
		}},
		Links: []DiscoveredLink{{
			NormalizedURL: "https://example.com/logout?session=should-not-pass",
			LinkText:      "Logout",
			Skipped:       true,
			SkipReason:    "unsafe_link_skipped",
		}},
		Forms: []DiscoveredForm{{
			FormAction:         "/login",
			FormMethod:         "post",
			FieldCount:         2,
			PasswordFieldCount: 1,
			Classification:     "password_form",
			Fields: []DiscoveredFormField{{
				FieldName: "password",
				FieldType: "password",
				Label:     "Password",
				Required:  true,
			}},
		}},
		Evidence: []Evidence{{
			Type: "screenshot",
			Metadata: map[string]any{
				"key":    "discovery-runs/run/screenshots/file.png",
				"cookie": "should-not-pass",
			},
		}},
	}

	input := BuildSafeDiscoveryAIInput(report)
	rendered := fmt.Sprintf("%v", input)
	for _, leaked := range []string{"should-not-pass", "cookie"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("safe discovery AI input leaked %q in:\n%s", leaked, rendered)
		}
	}
	if !strings.Contains(rendered, "unsafe_link_skipped") || !strings.Contains(rendered, "password_form") {
		t.Fatalf("safe discovery AI input missed expected map metadata:\n%s", rendered)
	}
}

func TestBuildSafeQualityAIInputStripsSecretsAndKeepsQualityContext(t *testing.T) {
	report := &QualityCheckReport{
		Run: QualityCheckRun{
			ID:                   "quality-run-id",
			ProjectID:            "project-id",
			Status:               StatusCompleted,
			TargetURL:            "https://example.com/dashboard?token=should-not-pass",
			MaxPages:             10,
			IncludeSecurity:      true,
			IncludeAccessibility: true,
			IncludePerformance:   true,
		},
		Project: Project{ID: "project-id", Name: "Example"},
		Summary: QualityCheckSummary{TotalFindings: 3, SecurityFindings: 1, AccessibilityFindings: 1, PerformanceFindings: 1},
		Results: []QualityCheckResult{{
			Category:       "security",
			RuleID:         "cookie_flags_incomplete",
			Severity:       "medium",
			Title:          "Cookie security flags are incomplete",
			Description:    "A cookie was visible without flags.",
			Recommendation: "Set HttpOnly, Secure, and SameSite where appropriate.",
			URL:            "https://example.com/dashboard?session=should-not-pass",
			Evidence: map[string]any{
				"cookies":          []any{map[string]any{"name": "session", "value": "cookie-secret"}},
				"authorization":    "Bearer should-not-pass",
				"local_storage":    "storage-secret",
				"response_body":    "body-secret",
				"safe_observation": "metadata kept",
			},
		}},
		SafetyNotes: []string{"Quality checks are passive."},
		Limitations: []string{"Not a full audit."},
	}

	input := BuildSafeQualityAIInput(report)
	rendered := prettyJSON(input)
	for _, leaked := range []string{"should-not-pass", "cookie-secret", "storage-secret", "body-secret", "Bearer"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("safe quality AI input leaked %q in:\n%s", leaked, rendered)
		}
	}
	if !strings.Contains(rendered, "cookie_flags_incomplete") || !strings.Contains(rendered, "metadata kept") {
		t.Fatalf("safe quality AI input missed expected quality context:\n%s", rendered)
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

package main

import (
	"strings"
	"testing"
)

func TestNormalizeAITestPlanRequestDefaultsAndValidates(t *testing.T) {
	req, err := NormalizeAITestPlanRequest(AITestPlanRequest{})
	if err != nil {
		t.Fatalf("normalize request: %v", err)
	}
	if req.MaxScenarios != defaultMaxTestPlanScenarios {
		t.Fatalf("unexpected default max scenarios %d", req.MaxScenarios)
	}
	if len(req.FocusAreas) == 0 {
		t.Fatal("expected default focus areas")
	}

	if _, err := NormalizeAITestPlanRequest(AITestPlanRequest{MaxScenarios: 31}); err == nil {
		t.Fatal("expected max_scenarios above limit to be rejected")
	}
	if _, err := NormalizeAITestPlanRequest(AITestPlanRequest{FocusAreas: []string{"exploit"}}); err == nil {
		t.Fatal("expected unsupported focus area to be rejected")
	}
}

func TestBuildSafeTestPlanInputRedactsSecretsAndIncludesAnalysisSummary(t *testing.T) {
	project := Project{
		ID:          "project-1",
		Name:        "Checkout App",
		FrontendURL: "https://example.com/path?api_key=secret#frag",
		AllowedHosts: []string{
			"example.com",
		},
	}
	report := &Report{
		RunID:     "run-1",
		ProjectID: "project-1",
		Status:    StatusCompleted,
		Summary:   ReportSummary{High: 1, TotalFindings: 1},
		Findings: []Finding{{
			Title:       "Console error",
			Severity:    "medium",
			Category:    "browser",
			Confidence:  "high",
			Description: "GET https://example.com/api?access_token=secret failed",
		}},
		Evidence: []Evidence{{
			Type: "browser_observations",
			Metadata: map[string]any{
				"final_url":       "https://example.com/dashboard?session_id=secret",
				"console_errors":  []any{"Bearer should-not-leak"},
				"response_body":   "private body",
				"html":            "<html>secret</html>",
				"failed_requests": []any{"https://example.com/api?password=secret"},
			},
		}},
		Metadata: map[string]any{"page_title": "Checkout"},
		AIAnalysis: &AIAnalysis{
			Status:           StatusCompleted,
			ExecutiveSummary: "Review the run. api_key=secret",
			TechnicalSummary: "One browser issue.",
			RiskLevel:        "medium",
			AnalysisJSON: map[string]any{
				"recommended_actions": []any{"Fix error"},
			},
		},
	}
	input, err := NormalizeAITestPlanRequest(AITestPlanRequest{
		ProductContext: "Use password=hunter2 for admin.",
		FocusAreas:     []string{"smoke", "regression"},
		MaxScenarios:   5,
	})
	if err != nil {
		t.Fatalf("normalize request: %v", err)
	}

	safeInput := BuildSafeTestPlanInput(project, report, input)
	rendered := prettyJSON(safeInput)
	for _, leaked := range []string{"api_key=secret", "access_token=secret", "session_id=secret", "password=hunter2", "should-not-leak", "private body", "<html>secret</html>"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("safe test plan input leaked %q in:\n%s", leaked, rendered)
		}
	}
	if !strings.Contains(rendered, `"ai_analysis"`) {
		t.Fatalf("expected AI analysis summary in safe input:\n%s", rendered)
	}
	if strings.Contains(rendered, "?") || strings.Contains(rendered, "#frag") {
		t.Fatalf("expected URL query and fragment to be stripped:\n%s", rendered)
	}
}

func TestParseTestPlanPayloadValidatesAndNormalizes(t *testing.T) {
	payload, planJSON, err := ParseTestPlanPayload(`{
		"title": "Checkout smoke plan",
		"summary": "Covers visible checkout smoke behavior.",
		"assumptions": ["No credentials were provided"],
		"coverage_goals": ["Landing page loads"],
		"scenarios": [
			{
				"id": "checkout-smoke",
				"name": "Checkout page loads",
				"type": "smoke",
				"priority": "HIGH",
				"risk": "medium",
				"description": "Verify the observed page still loads.",
				"preconditions": [],
				"steps": [{"order": 0, "action": "Open page with Bearer secret-token", "target": "https://example.com/?token=secret", "data": "", "expected_result": "Page loads"}],
				"assertions": ["Page title is visible"],
				"test_data_needed": [],
				"automation_candidate": true,
				"destructive": false,
				"requires_authentication": false,
				"related_findings": [],
				"tags": ["smoke"]
			}
		],
		"suggested_next_instrumentation": ["Capture trace later"],
		"limitations": ["No auth flows were provided"]
	}`, 10)
	if err != nil {
		t.Fatalf("parse test plan: %v", err)
	}
	if payload.Scenarios[0].Priority != "high" {
		t.Fatalf("expected priority to normalize, got %q", payload.Scenarios[0].Priority)
	}
	if payload.Scenarios[0].Steps[0].Order != 1 {
		t.Fatalf("expected missing step order to normalize to 1")
	}
	if strings.Contains(payload.Scenarios[0].Steps[0].Action, "secret-token") {
		t.Fatalf("expected step action to be redacted")
	}
	if _, ok := planJSON["scenarios"]; !ok {
		t.Fatalf("expected normalized plan JSON to include scenarios")
	}
}

func TestParseTestPlanPayloadRejectsUnsupportedScenarioType(t *testing.T) {
	_, _, err := ParseTestPlanPayload(`{
		"title": "Bad plan",
		"summary": "Bad plan",
		"assumptions": [],
		"coverage_goals": [],
		"scenarios": [{
			"id": "bad",
			"name": "Bad",
			"type": "exploit",
			"priority": "high",
			"risk": "high",
			"description": "Bad",
			"preconditions": [],
			"steps": [{"order": 1, "action": "Do thing", "target": "App", "data": "", "expected_result": "Works"}],
			"assertions": [],
			"test_data_needed": [],
			"automation_candidate": false,
			"destructive": false,
			"requires_authentication": false,
			"related_findings": [],
			"tags": []
		}],
		"suggested_next_instrumentation": [],
		"limitations": []
	}`, 10)
	if err == nil {
		t.Fatal("expected unsupported scenario type to be rejected")
	}
}

func TestParseTestPlanPayloadRespectsMaxScenarios(t *testing.T) {
	raw := `{
		"title": "Plan",
		"summary": "Summary",
		"assumptions": [],
		"coverage_goals": [],
		"scenarios": [
			{"id":"a","name":"A","type":"smoke","priority":"low","risk":"low","description":"A","preconditions":[],"steps":[{"order":1,"action":"A","target":"A","data":"","expected_result":"A"}],"assertions":[],"test_data_needed":[],"automation_candidate":true,"destructive":false,"requires_authentication":false,"related_findings":[],"tags":[]},
			{"id":"b","name":"B","type":"functional","priority":"medium","risk":"medium","description":"B","preconditions":[],"steps":[{"order":1,"action":"B","target":"B","data":"","expected_result":"B"}],"assertions":[],"test_data_needed":[],"automation_candidate":true,"destructive":false,"requires_authentication":false,"related_findings":[],"tags":[]}
		],
		"suggested_next_instrumentation": [],
		"limitations": []
	}`
	payload, _, err := ParseTestPlanPayload(raw, 1)
	if err != nil {
		t.Fatalf("parse plan: %v", err)
	}
	if len(payload.Scenarios) != 1 {
		t.Fatalf("expected scenarios to be capped to 1, got %d", len(payload.Scenarios))
	}
}

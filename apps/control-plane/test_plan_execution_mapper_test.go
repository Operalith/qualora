package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildTestPlanExecutionPreviewMapsSafeDSL(t *testing.T) {
	plan := testPlanForExecution(t, []TestPlanScenario{{
		ID:                  "homepage",
		Name:                "Homepage smoke",
		Type:                "smoke",
		Priority:            "high",
		Risk:                "medium",
		Description:         "Verify public homepage content.",
		AutomationCandidate: true,
		Steps: []TestPlanStep{
			{Order: 1, Action: "goto", Target: "/"},
			{Order: 2, Action: "assert_title_contains", Target: "Qualora Demo Web"},
			{Order: 3, Action: "assert_text_visible", Target: "Self-hosted QA automation demo"},
			{Order: 4, Action: "assert_link_exists", Target: "/status"},
			{Order: 5, Action: "capture_screenshot"},
		},
	}})
	preview, err := BuildTestPlanExecutionPreview(plan, executionTestProject(), TestPlanExecutionRequest{DryRun: true})
	if err != nil {
		t.Fatalf("build preview: %v", err)
	}
	if !preview.DryRun {
		t.Fatal("expected dry_run flag to be preserved")
	}
	if preview.ExecutableScenarios != 1 || preview.ExecutableSteps != 5 || preview.SkippedSteps != 0 {
		t.Fatalf("unexpected preview counts: %#v", preview)
	}
	linkStep := preview.Scenarios[0].Steps[3]
	if linkStep.Target != "https://example.com/status" {
		t.Fatalf("expected same-origin URL to normalize, got %q", linkStep.Target)
	}
}

func TestBuildTestPlanExecutionPreviewSkipsUnsafeScenario(t *testing.T) {
	plan := testPlanForExecution(t, []TestPlanScenario{{
		ID:                     "login",
		Name:                   "Login with credentials",
		Type:                   "functional",
		Priority:               "high",
		Risk:                   "high",
		Description:            "Log in with a password.",
		AutomationCandidate:    true,
		RequiresAuthentication: true,
		Steps: []TestPlanStep{
			{Order: 1, Action: "goto", Target: "/login"},
		},
	}})
	preview, err := BuildTestPlanExecutionPreview(plan, executionTestProject(), TestPlanExecutionRequest{})
	if err != nil {
		t.Fatalf("build preview: %v", err)
	}
	if preview.ExecutableSteps != 0 || preview.SkippedScenarios != 1 {
		t.Fatalf("expected scenario to be skipped: %#v", preview)
	}
	if !strings.Contains(preview.Scenarios[0].SkipReason, "authentication") {
		t.Fatalf("expected authentication skip reason, got %q", preview.Scenarios[0].SkipReason)
	}
}

func TestBuildTestPlanExecutionPreviewSkipsUnsupportedAction(t *testing.T) {
	plan := testPlanForExecution(t, []TestPlanScenario{{
		ID:                  "unsupported",
		Name:                "Unsupported browser action",
		Type:                "smoke",
		Priority:            "medium",
		Risk:                "low",
		Description:         "Verify unsupported actions are not executed.",
		AutomationCandidate: true,
		Steps: []TestPlanStep{
			{Order: 1, Action: "click_button", Target: "#start"},
		},
	}})
	preview, err := BuildTestPlanExecutionPreview(plan, executionTestProject(), TestPlanExecutionRequest{})
	if err != nil {
		t.Fatalf("build preview: %v", err)
	}
	if preview.ExecutableSteps != 0 || preview.SkippedSteps != 1 {
		t.Fatalf("expected unsupported step to be skipped: %#v", preview)
	}
	if preview.Scenarios[0].Steps[0].SkipReason != "unsupported safe execution action" {
		t.Fatalf("unexpected skip reason: %q", preview.Scenarios[0].Steps[0].SkipReason)
	}
}

func TestBuildTestPlanExecutionPreviewRejectsOutOfScopeURL(t *testing.T) {
	plan := testPlanForExecution(t, []TestPlanScenario{{
		ID:                  "external",
		Name:                "External navigation",
		Type:                "smoke",
		Priority:            "medium",
		Risk:                "low",
		Description:         "Verify external URLs are skipped.",
		AutomationCandidate: true,
		Steps: []TestPlanStep{
			{Order: 1, Action: "goto", Target: "https://evil.example.net/"},
		},
	}})
	preview, err := BuildTestPlanExecutionPreview(plan, executionTestProject(), TestPlanExecutionRequest{})
	if err != nil {
		t.Fatalf("build preview: %v", err)
	}
	step := preview.Scenarios[0].Steps[0]
	if step.Status != StatusSkipped || !strings.Contains(step.SkipReason, "project frontend origin") {
		t.Fatalf("expected out-of-scope URL skip, got %#v", step)
	}
}

func TestBuildTestPlanExecutionPreviewRejectsSensitiveQueryURL(t *testing.T) {
	plan := testPlanForExecution(t, []TestPlanScenario{{
		ID:                  "query",
		Name:                "Sensitive query navigation",
		Type:                "smoke",
		Priority:            "medium",
		Risk:                "low",
		Description:         "Verify sensitive query URLs are skipped.",
		AutomationCandidate: true,
		Steps: []TestPlanStep{
			{Order: 1, Action: "check_link_status", Target: "/status?token=secret"},
		},
	}})
	preview, err := BuildTestPlanExecutionPreview(plan, executionTestProject(), TestPlanExecutionRequest{})
	if err != nil {
		t.Fatalf("build preview: %v", err)
	}
	step := preview.Scenarios[0].Steps[0]
	if step.Status != StatusSkipped || !strings.Contains(step.SkipReason, "sensitive parameter") {
		t.Fatalf("expected sensitive query skip, got %#v", step)
	}
}

func TestNormalizeTestPlanExecutionRequestDefaultsAndCaps(t *testing.T) {
	req := NormalizeTestPlanExecutionRequest(TestPlanExecutionRequest{
		MaxScenarios:        100,
		MaxStepsPerScenario: 100,
		ScenarioIDs:         []string{" a ", "a", "", "b"},
	})
	if req.MaxScenarios != maxExecutionScenarios {
		t.Fatalf("expected max_scenarios cap, got %d", req.MaxScenarios)
	}
	if req.MaxStepsPerScenario != maxExecutionStepsPerScenario {
		t.Fatalf("expected max_steps_per_scenario cap, got %d", req.MaxStepsPerScenario)
	}
	if len(req.ScenarioIDs) != 2 || req.ScenarioIDs[0] != "a" || req.ScenarioIDs[1] != "b" {
		t.Fatalf("expected scenario ids to dedupe, got %#v", req.ScenarioIDs)
	}
}

func executionTestProject() Project {
	return Project{
		ID:           "project-1",
		Name:         "Demo",
		FrontendURL:  "https://example.com/",
		AllowedHosts: []string{"example.com"},
	}
}

func testPlanForExecution(t *testing.T, scenarios []TestPlanScenario) TestPlan {
	t.Helper()
	payload := TestPlanPayload{
		Title:     "Safe execution test plan",
		Summary:   "Used by execution mapper tests.",
		Scenarios: scenarios,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, planJSON, err := ParseTestPlanPayload(string(raw), len(scenarios))
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	return TestPlan{
		ID:        "plan-1",
		ProjectID: "project-1",
		Status:    StatusCompleted,
		PlanJSON:  planJSON,
	}
}

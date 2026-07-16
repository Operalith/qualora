package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

const (
	defaultMaxExecutionScenarios        = 5
	maxExecutionScenarios               = 20
	defaultMaxExecutionStepsPerScenario = 10
	maxExecutionStepsPerScenario        = 30
)

var supportedExecutionActions = map[string]bool{
	"goto":                      true,
	"assert_title_contains":     true,
	"assert_url_contains":       true,
	"assert_text_visible":       true,
	"assert_element_visible":    true,
	"assert_link_exists":        true,
	"check_link_status":         true,
	"capture_screenshot":        true,
	"collect_browser_signals":   true,
	"wait_for_load_state":       true,
	"assert_no_console_errors":  true,
	"assert_no_failed_requests": true,
}

var unsafeExecutionPhrases = []string{
	"login",
	"log in",
	"sign in",
	"signin",
	"authenticate",
	"authentication",
	"password",
	"credential",
	"payment",
	"checkout payment",
	"delete",
	"update",
	"create",
	"submit",
	"upload",
	"admin mutation",
	"sql injection",
	"sqli",
	"xss",
	"ssrf",
	"brute force",
	"bruteforce",
	"exploit",
	"destructive",
}

var sensitiveExecutionQueryNames = []string{
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
}

func NormalizeTestPlanExecutionRequest(input TestPlanExecutionRequest) TestPlanExecutionRequest {
	if input.MaxScenarios == 0 {
		input.MaxScenarios = defaultMaxExecutionScenarios
	}
	if input.MaxScenarios < 1 {
		input.MaxScenarios = 1
	}
	if input.MaxScenarios > maxExecutionScenarios {
		input.MaxScenarios = maxExecutionScenarios
	}
	if input.MaxStepsPerScenario == 0 {
		input.MaxStepsPerScenario = defaultMaxExecutionStepsPerScenario
	}
	if input.MaxStepsPerScenario < 1 {
		input.MaxStepsPerScenario = 1
	}
	if input.MaxStepsPerScenario > maxExecutionStepsPerScenario {
		input.MaxStepsPerScenario = maxExecutionStepsPerScenario
	}

	seen := map[string]bool{}
	scenarioIDs := make([]string, 0, len(input.ScenarioIDs))
	for _, id := range input.ScenarioIDs {
		id = strings.TrimSpace(sanitizeText(RedactSecrets(id)))
		if id == "" || seen[id] {
			continue
		}
		scenarioIDs = append(scenarioIDs, limitString(id, 120))
		seen[id] = true
	}
	input.ScenarioIDs = scenarioIDs
	return input
}

func BuildTestPlanExecutionPreview(plan TestPlan, project Project, input TestPlanExecutionRequest) (*TestPlanExecutionPreview, error) {
	input = NormalizeTestPlanExecutionRequest(input)
	payload, err := payloadFromStoredTestPlan(plan)
	if err != nil {
		return nil, err
	}
	if project.FrontendURL == "" {
		return nil, fmt.Errorf("project frontend_url is required for safe test plan execution")
	}
	if _, err := parseExecutionRootURL(project); err != nil {
		return nil, err
	}

	includeScenarioID := map[string]bool{}
	for _, id := range input.ScenarioIDs {
		includeScenarioID[id] = true
	}

	preview := &TestPlanExecutionPreview{
		DryRun:              input.DryRun,
		TestPlanID:          plan.ID,
		ProjectID:           project.ID,
		MaxScenarios:        input.MaxScenarios,
		MaxStepsPerScenario: input.MaxStepsPerScenario,
		TotalScenarios:      len(payload.Scenarios),
		Scenarios:           make([]MappedExecutionScenario, 0, min(len(payload.Scenarios), input.MaxScenarios)),
	}

	included := 0
	for _, scenario := range payload.Scenarios {
		if len(includeScenarioID) > 0 && !includeScenarioID[scenario.ID] {
			continue
		}
		if included >= input.MaxScenarios {
			break
		}
		included++
		mapped := mapExecutionScenario(project, scenario, input.MaxStepsPerScenario)
		preview.Scenarios = append(preview.Scenarios, mapped)
		preview.TotalSteps += len(mapped.Steps)
		if mapped.Status == StatusSkipped {
			preview.SkippedScenarios++
			preview.SafetySummary.SkippedScenarios++
		} else {
			preview.ExecutableScenarios++
		}
		for _, step := range mapped.Steps {
			if step.Status == StatusQueued {
				preview.ExecutableSteps++
				preview.SafetySummary.ExecutedSteps++
			} else {
				preview.SkippedSteps++
				if strings.Contains(step.SkipReason, "unsafe") || strings.Contains(step.SkipReason, "authentication") || strings.Contains(step.SkipReason, "destructive") {
					preview.SafetySummary.SkippedUnsafeSteps++
				} else {
					preview.SafetySummary.SkippedUnsupportedSteps++
				}
			}
		}
	}
	return preview, nil
}

func TestPlanCoverageFromPreview(preview *TestPlanExecutionPreview) TestPlanExecutableCoverage {
	if preview == nil {
		return TestPlanExecutableCoverage{}
	}
	return TestPlanExecutableCoverage{
		TotalScenarios:          preview.TotalScenarios,
		ExecutableScenarios:     preview.ExecutableScenarios,
		SkippedScenarios:        preview.SkippedScenarios,
		TotalSteps:              preview.TotalSteps,
		ExecutableSteps:         preview.ExecutableSteps,
		SkippedSteps:            preview.SkippedSteps,
		UnsafeSkippedSteps:      preview.SafetySummary.SkippedUnsafeSteps,
		UnsupportedSkippedSteps: preview.SafetySummary.SkippedUnsupportedSteps,
	}
}

func payloadFromStoredTestPlan(plan TestPlan) (*TestPlanPayload, error) {
	raw, err := json.Marshal(plan.PlanJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal stored test plan: %w", err)
	}
	var payload TestPlanPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("stored test plan is not a valid Qualora test plan: %w", err)
	}
	if payload.Title == "" || len(payload.Scenarios) == 0 {
		return nil, fmt.Errorf("stored test plan is incomplete")
	}
	return &payload, nil
}

func mapExecutionScenario(project Project, scenario TestPlanScenario, maxSteps int) MappedExecutionScenario {
	mapped := MappedExecutionScenario{
		ScenarioIDFromPlan: scenario.ID,
		Name:               scenario.Name,
		Type:               scenario.Type,
		Priority:           scenario.Priority,
		Status:             StatusQueued,
		Steps:              make([]MappedExecutionStep, 0, min(len(scenario.Steps), maxSteps)),
	}

	if reason := unsafeScenarioReason(scenario); reason != "" {
		mapped.Status = StatusSkipped
		mapped.SkipReason = reason
		for i, step := range limitedPlanSteps(scenario.Steps, maxSteps) {
			mapped.Steps = append(mapped.Steps, skippedExecutionStep(step, i+1, reason))
		}
		return mapped
	}

	executableSteps := 0
	for i, step := range limitedPlanSteps(scenario.Steps, maxSteps) {
		mappedStep := mapExecutionStep(project, step, i+1)
		if mappedStep.Status == StatusQueued {
			executableSteps++
		}
		mapped.Steps = append(mapped.Steps, mappedStep)
	}
	if executableSteps == 0 {
		mapped.Status = StatusSkipped
		mapped.SkipReason = "scenario has no supported safe executable steps"
	}
	return mapped
}

func limitedPlanSteps(steps []TestPlanStep, maxSteps int) []TestPlanStep {
	if len(steps) <= maxSteps {
		return steps
	}
	return steps[:maxSteps]
}

func unsafeScenarioReason(scenario TestPlanScenario) string {
	if !scenario.AutomationCandidate {
		return "scenario is not marked as an automation candidate"
	}
	if scenario.Destructive {
		return "scenario is marked destructive"
	}
	if scenario.RequiresAuthentication {
		return "scenario requires authentication"
	}
	if containsUnsafeExecutionIntent(strings.Join(scenarioTextParts(scenario), " ")) {
		return "scenario contains unsafe or unsupported intent"
	}
	return ""
}

func scenarioTextParts(scenario TestPlanScenario) []string {
	parts := []string{
		scenario.ID,
		scenario.Name,
		scenario.Type,
		scenario.Priority,
		scenario.Risk,
		scenario.Description,
	}
	parts = append(parts, scenario.Preconditions...)
	parts = append(parts, scenario.Assertions...)
	parts = append(parts, scenario.TestDataNeeded...)
	parts = append(parts, scenario.RelatedFindings...)
	parts = append(parts, scenario.Tags...)
	for _, step := range scenario.Steps {
		parts = append(parts, step.Action, step.Target, step.Data, step.ExpectedResult)
	}
	return parts
}

func mapExecutionStep(project Project, step TestPlanStep, fallbackOrder int) MappedExecutionStep {
	order := step.Order
	if order <= 0 {
		order = fallbackOrder
	}
	mapped := MappedExecutionStep{
		StepOrder:      order,
		OriginalAction: limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.Action))), 120),
		MappedAction:   strings.ToLower(strings.TrimSpace(step.Action)),
		Target:         limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.Target))), 1200),
		ExpectedResult: limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.ExpectedResult))), 1200),
		Status:         StatusQueued,
	}

	if mapped.MappedAction == "" {
		mapped.Status = StatusSkipped
		mapped.SkipReason = "step action is empty"
		return mapped
	}
	if !supportedExecutionActions[mapped.MappedAction] {
		mapped.Status = StatusSkipped
		mapped.SkipReason = "unsupported safe execution action"
		return mapped
	}
	if containsUnsafeExecutionIntent(strings.Join([]string{step.Action, step.Target, step.Data, step.ExpectedResult}, " ")) {
		mapped.Status = StatusSkipped
		mapped.SkipReason = "step contains unsafe or unsupported intent"
		return mapped
	}

	switch mapped.MappedAction {
	case "goto", "check_link_status", "assert_link_exists":
		normalized, err := normalizeExecutionTargetURL(project, firstNonEmpty(step.Target, step.Data))
		if err != nil {
			mapped.Status = StatusSkipped
			mapped.SkipReason = "unsafe or out-of-scope URL target: " + err.Error()
			return mapped
		}
		mapped.Target = normalized
	case "assert_title_contains", "assert_url_contains", "assert_text_visible", "assert_element_visible":
		mapped.Target = firstNonEmpty(mapped.Target, sanitizeText(RedactSecrets(step.Data)), mapped.ExpectedResult)
		if strings.TrimSpace(mapped.Target) == "" {
			mapped.Status = StatusSkipped
			mapped.SkipReason = "assertion target is empty"
			return mapped
		}
	case "capture_screenshot", "collect_browser_signals", "wait_for_load_state", "assert_no_console_errors", "assert_no_failed_requests":
		mapped.Target = ""
	}

	return mapped
}

func skippedExecutionStep(step TestPlanStep, fallbackOrder int, reason string) MappedExecutionStep {
	order := step.Order
	if order <= 0 {
		order = fallbackOrder
	}
	return MappedExecutionStep{
		StepOrder:      order,
		OriginalAction: limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.Action))), 120),
		MappedAction:   strings.ToLower(strings.TrimSpace(step.Action)),
		Target:         limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.Target))), 1200),
		ExpectedResult: limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.ExpectedResult))), 1200),
		Status:         StatusSkipped,
		SkipReason:     reason,
	}
}

func containsUnsafeExecutionIntent(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}
	normalized = strings.NewReplacer("_", " ", "-", " ").Replace(normalized)
	for _, phrase := range unsafeExecutionPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func normalizeExecutionTargetURL(project Project, raw string) (string, error) {
	root, err := parseExecutionRootURL(project)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("URL target is required")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("URL target is invalid")
	}
	if !parsed.IsAbs() {
		parsed = root.ResolveReference(parsed)
	}
	parsed.Fragment = ""
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are supported")
	}
	if !sameOrigin(root, parsed) {
		return "", fmt.Errorf("target must stay on the project frontend origin")
	}
	if !HostAllowed(parsed.Hostname(), project.AllowedHosts) {
		return "", fmt.Errorf("target host is not present in allowed_hosts")
	}
	if hasSensitiveExecutionQuery(parsed.Query()) {
		return "", fmt.Errorf("target query contains sensitive parameter names")
	}
	return parsed.String(), nil
}

func parseExecutionRootURL(project Project) (*url.URL, error) {
	root, err := url.Parse(strings.TrimSpace(project.FrontendURL))
	if err != nil || root.Scheme == "" || root.Host == "" {
		return nil, fmt.Errorf("project frontend_url is invalid")
	}
	if root.Scheme != "http" && root.Scheme != "https" {
		return nil, fmt.Errorf("project frontend_url must use http or https")
	}
	return root, nil
}

func sameOrigin(left *url.URL, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) && strings.EqualFold(left.Host, right.Host)
}

func hasSensitiveExecutionQuery(values url.Values) bool {
	for name := range values {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized == "" {
			continue
		}
		if slices.Contains(sensitiveExecutionQueryNames, normalized) {
			return true
		}
		for _, sensitive := range sensitiveExecutionQueryNames {
			if strings.Contains(normalized, sensitive) {
				return true
			}
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

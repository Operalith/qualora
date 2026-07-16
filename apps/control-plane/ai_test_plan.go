package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	defaultMaxTestPlanScenarios = 10
	maxTestPlanScenarios        = 30
	defaultMaxDiscoveryPages    = 20
	maxDiscoveryPages           = 100
)

var allowedFocusAreas = map[string]bool{
	"smoke":            true,
	"functional":       true,
	"negative":         true,
	"accessibility":    true,
	"performance":      true,
	"security-passive": true,
	"authorization":    true,
	"api":              true,
	"visual":           true,
	"regression":       true,
}

var allowedScenarioTypes = map[string]bool{
	"smoke":            true,
	"functional":       true,
	"negative":         true,
	"accessibility":    true,
	"performance":      true,
	"security-passive": true,
	"authorization":    true,
	"api":              true,
	"visual":           true,
	"regression":       true,
}

type TestPlanPayload struct {
	Title                        string             `json:"title"`
	Summary                      string             `json:"summary"`
	Assumptions                  []string           `json:"assumptions"`
	CoverageGoals                []string           `json:"coverage_goals"`
	Scenarios                    []TestPlanScenario `json:"scenarios"`
	SuggestedNextInstrumentation []string           `json:"suggested_next_instrumentation"`
	Limitations                  []string           `json:"limitations"`
}

type TestPlanScenario struct {
	ID                     string         `json:"id"`
	Name                   string         `json:"name"`
	Type                   string         `json:"type"`
	Priority               string         `json:"priority"`
	Risk                   string         `json:"risk"`
	Description            string         `json:"description"`
	Preconditions          []string       `json:"preconditions"`
	Steps                  []TestPlanStep `json:"steps"`
	Assertions             []string       `json:"assertions"`
	TestDataNeeded         []string       `json:"test_data_needed"`
	AutomationCandidate    bool           `json:"automation_candidate"`
	Destructive            bool           `json:"destructive"`
	RequiresAuthentication bool           `json:"requires_authentication"`
	RelatedFindings        []string       `json:"related_findings"`
	Tags                   []string       `json:"tags"`
}

type TestPlanStep struct {
	Order          int    `json:"order"`
	Action         string `json:"action"`
	Target         string `json:"target"`
	Data           string `json:"data"`
	ExpectedResult string `json:"expected_result"`
}

func AITestPlanSystemPrompt() string {
	return strings.Join([]string{
		"You are Qualora's AI test planning assistant.",
		"Create reviewable software QA test plans only.",
		"Use only the provided project, run, and discovery-map data.",
		"Do not invent discovered pages, credentials, private APIs, user roles, database access, or evidence.",
		"When discovery_map is present, base page-navigation scenarios only on pages and links included in that map.",
		"Never suggest credential use, login, arbitrary form submission, destructive actions, active scanning, fuzzing, exploit payloads, or autonomous browser control.",
		"Clearly mark assumptions when data is insufficient.",
		"Prefer safe, non-destructive tests.",
		"Do not generate active exploit payloads or destructive security steps.",
		"Passive security checks are allowed.",
		"Authorization tests may be suggested only at a conceptual reviewable level unless roles or accounts appear in the input.",
		"If execution_mode is safe_executable, prefer these exact deterministic browser DSL actions only: goto, assert_title_contains, assert_url_contains, assert_text_visible, assert_element_visible, assert_link_exists, check_link_status, capture_screenshot, collect_browser_signals, wait_for_load_state, assert_no_console_errors, assert_no_failed_requests.",
		"Generated plans are suggestions and must not be executed automatically without an explicit Qualora execution request.",
		"Return strict JSON only.",
		`Use this JSON shape exactly: {"title":"string","summary":"string","assumptions":["string"],"coverage_goals":["string"],"scenarios":[{"id":"string","name":"string","type":"smoke|functional|negative|accessibility|performance|security-passive|authorization|api|visual|regression","priority":"low|medium|high|critical","risk":"low|medium|high|critical","description":"string","preconditions":["string"],"steps":[{"order":1,"action":"string","target":"string","data":"string","expected_result":"string"}],"assertions":["string"],"test_data_needed":["string"],"automation_candidate":true,"destructive":false,"requires_authentication":false,"related_findings":["string"],"tags":["string"]}],"suggested_next_instrumentation":["string"],"limitations":["string"]}.`,
	}, " ")
}

func NormalizeAITestPlanRequest(input AITestPlanRequest) (AITestPlanRequest, error) {
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.RunID = strings.TrimSpace(input.RunID)
	input.DiscoveryRunID = strings.TrimSpace(input.DiscoveryRunID)
	input.ExecutionMode = strings.ToLower(strings.TrimSpace(input.ExecutionMode))
	input.ProductContext = strings.TrimSpace(sanitizeText(RedactSecrets(limitString(input.ProductContext, 4000))))
	if input.IncludeDiscoveryMap == nil {
		value := true
		input.IncludeDiscoveryMap = &value
	}
	if input.ExecutionMode == "" {
		input.ExecutionMode = AITestPlanExecutionModeReviewOnly
	}
	if input.ExecutionMode != AITestPlanExecutionModeReviewOnly && input.ExecutionMode != AITestPlanExecutionModeSafeExecutable {
		return input, fmt.Errorf("execution_mode must be review_only or safe_executable")
	}
	if input.MaxPagesFromDiscovery == 0 {
		input.MaxPagesFromDiscovery = defaultMaxDiscoveryPages
	}
	if input.MaxPagesFromDiscovery < 1 || input.MaxPagesFromDiscovery > maxDiscoveryPages {
		return input, fmt.Errorf("max_pages_from_discovery must be between 1 and 100")
	}
	if input.MaxScenarios == 0 {
		input.MaxScenarios = defaultMaxTestPlanScenarios
	}
	if input.MaxScenarios < 1 || input.MaxScenarios > maxTestPlanScenarios {
		return input, fmt.Errorf("max_scenarios must be between 1 and 30")
	}
	if len(input.FocusAreas) == 0 {
		input.FocusAreas = []string{"smoke", "functional", "negative", "accessibility", "regression"}
	}
	seen := map[string]bool{}
	focusAreas := make([]string, 0, len(input.FocusAreas))
	for _, area := range input.FocusAreas {
		area = strings.ToLower(strings.TrimSpace(area))
		if area == "" {
			continue
		}
		if !allowedFocusAreas[area] {
			return input, fmt.Errorf("focus_areas contains unsupported value %q", area)
		}
		if !seen[area] {
			focusAreas = append(focusAreas, area)
			seen[area] = true
		}
	}
	if len(focusAreas) == 0 {
		return input, fmt.Errorf("focus_areas must include at least one supported value")
	}
	input.FocusAreas = focusAreas
	return input, nil
}

func BuildAITestPlanUserPrompt(project Project, report *Report, discoveryReport *DiscoveryReport, input AITestPlanRequest) (string, error) {
	safeInput := BuildSafeTestPlanInput(project, report, discoveryReport, input)
	raw, err := json.MarshalIndent(safeInput, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal safe AI test plan input: %w", err)
	}
	return "Create a reviewable Qualora AI test plan from this sanitized JSON input:\n" + string(raw), nil
}

func BuildSafeTestPlanInput(project Project, report *Report, discoveryReport *DiscoveryReport, input AITestPlanRequest) map[string]any {
	payload := map[string]any{
		"project": map[string]any{
			"id":            project.ID,
			"name":          project.Name,
			"frontend_url":  project.FrontendURL,
			"api_base_url":  project.APIBaseURL,
			"openapi_url":   project.OpenAPIURL,
			"allowed_hosts": project.AllowedHosts,
		},
		"product_context":          input.ProductContext,
		"focus_areas":              input.FocusAreas,
		"max_scenarios":            input.MaxScenarios,
		"execution_mode":           input.ExecutionMode,
		"include_discovery_map":    input.IncludeDiscoveryMap != nil && *input.IncludeDiscoveryMap,
		"max_pages_from_discovery": input.MaxPagesFromDiscovery,
		"safe_execution_contract": map[string]any{
			"actions":                       safeExecutionActions(),
			"forms_submitted":               false,
			"destructive_actions":           false,
			"autonomous_ai_browser_control": false,
			"execution_is_allowed_only_after_explicit_user_request": true,
		},
	}
	if report != nil {
		payload["run_report"] = BuildSafeAIInput(report)
		if report.AIAnalysis != nil {
			payload["ai_analysis"] = map[string]any{
				"status":               report.AIAnalysis.Status,
				"executive_summary":    report.AIAnalysis.ExecutiveSummary,
				"technical_summary":    report.AIAnalysis.TechnicalSummary,
				"risk_level":           report.AIAnalysis.RiskLevel,
				"recommended_actions":  jsonArrayField(report.AIAnalysis.AnalysisJSON, "recommended_actions"),
				"suggested_next_tests": jsonArrayField(report.AIAnalysis.AnalysisJSON, "suggested_next_tests"),
				"limitations":          jsonArrayField(report.AIAnalysis.AnalysisJSON, "limitations"),
			}
		}
	}
	if discoveryReport != nil && input.IncludeDiscoveryMap != nil && *input.IncludeDiscoveryMap {
		payload["discovery_map"] = BuildSafeDiscoveryTestPlanInput(discoveryReport, input.MaxPagesFromDiscovery)
	}
	return sanitizeValue(payload).(map[string]any)
}

func BuildSafeDiscoveryTestPlanInput(report *DiscoveryReport, maxPages int) map[string]any {
	if maxPages <= 0 {
		maxPages = defaultMaxDiscoveryPages
	}
	if maxPages > maxDiscoveryPages {
		maxPages = maxDiscoveryPages
	}
	pages := report.Pages
	if len(pages) > maxPages {
		pages = pages[:maxPages]
	}
	input := BuildSafeDiscoveryAIInput(&DiscoveryReport{
		Run:         report.Run,
		Project:     report.Project,
		Settings:    report.Settings,
		Summary:     report.Summary,
		Pages:       pages,
		Links:       report.Links,
		Forms:       report.Forms,
		Findings:    report.Findings,
		Evidence:    report.Evidence,
		SafetyNotes: report.SafetyNotes,
		Limitations: report.Limitations,
		Metadata:    report.Metadata,
	})
	input["input_limits"] = map[string]any{
		"max_pages_from_discovery": maxPages,
		"included_pages":           len(pages),
		"total_discovered_pages":   len(report.Pages),
	}
	return sanitizeValue(input).(map[string]any)
}

func safeExecutionActions() []string {
	return []string{
		"goto",
		"assert_title_contains",
		"assert_url_contains",
		"assert_text_visible",
		"assert_element_visible",
		"assert_link_exists",
		"check_link_status",
		"capture_screenshot",
		"collect_browser_signals",
		"wait_for_load_state",
		"assert_no_console_errors",
		"assert_no_failed_requests",
	}
}

func TagDiscoveryGeneratedTestPlan(payload *TestPlanPayload, discoveryRunID string, executionMode string) {
	if payload == nil {
		return
	}
	for index := range payload.Scenarios {
		payload.Scenarios[index].Tags = appendUniqueTags(payload.Scenarios[index].Tags, "generated_from_discovery")
		if discoveryRunID != "" {
			payload.Scenarios[index].Tags = appendUniqueTags(payload.Scenarios[index].Tags, "discovery_run:"+discoveryRunID)
		}
		if executionMode == AITestPlanExecutionModeSafeExecutable {
			payload.Scenarios[index].Tags = appendUniqueTags(payload.Scenarios[index].Tags, "safe_executable_candidate")
		}
	}
}

func appendUniqueTags(tags []string, values ...string) []string {
	seen := map[string]bool{}
	output := make([]string, 0, len(tags)+len(values))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[tag] {
			continue
		}
		output = append(output, tag)
		seen[tag] = true
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		output = append(output, value)
		seen[value] = true
	}
	if len(output) > 20 {
		return output[:20]
	}
	return output
}

func ParseTestPlanPayload(raw string, maxScenarios int) (*TestPlanPayload, map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}
	var payload TestPlanPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, nil, fmt.Errorf("AI test plan was not valid JSON: %w", err)
	}
	if maxScenarios <= 0 || maxScenarios > maxTestPlanScenarios {
		maxScenarios = defaultMaxTestPlanScenarios
	}
	normalized, err := normalizeTestPlanPayload(payload, maxScenarios)
	if err != nil {
		return nil, nil, err
	}
	rawJSON, err := json.Marshal(normalized)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal parsed AI test plan: %w", err)
	}
	var planJSON map[string]any
	if err := json.Unmarshal(rawJSON, &planJSON); err != nil {
		return nil, nil, fmt.Errorf("normalize parsed AI test plan: %w", err)
	}
	return normalized, planJSON, nil
}

func normalizeTestPlanPayload(payload TestPlanPayload, maxScenarios int) (*TestPlanPayload, error) {
	payload.Title = strings.TrimSpace(sanitizeText(RedactSecrets(payload.Title)))
	payload.Summary = strings.TrimSpace(sanitizeText(RedactSecrets(payload.Summary)))
	if payload.Title == "" {
		return nil, fmt.Errorf("AI test plan title is required")
	}
	if payload.Summary == "" {
		return nil, fmt.Errorf("AI test plan summary is required")
	}
	payload.Title = limitString(payload.Title, 200)
	payload.Summary = limitString(payload.Summary, 1200)
	payload.Assumptions = sanitizeStringSlice(payload.Assumptions, 20)
	payload.CoverageGoals = sanitizeStringSlice(payload.CoverageGoals, 20)
	payload.SuggestedNextInstrumentation = sanitizeStringSlice(payload.SuggestedNextInstrumentation, 20)
	payload.Limitations = sanitizeStringSlice(payload.Limitations, 20)
	if len(payload.Scenarios) == 0 {
		return nil, fmt.Errorf("AI test plan must include at least one scenario")
	}
	if len(payload.Scenarios) > maxScenarios {
		payload.Scenarios = payload.Scenarios[:maxScenarios]
	}
	for i := range payload.Scenarios {
		scenario, err := normalizeTestPlanScenario(payload.Scenarios[i], i+1)
		if err != nil {
			return nil, err
		}
		payload.Scenarios[i] = scenario
	}
	return &payload, nil
}

func normalizeTestPlanScenario(scenario TestPlanScenario, index int) (TestPlanScenario, error) {
	scenario.ID = strings.TrimSpace(sanitizeText(RedactSecrets(scenario.ID)))
	if scenario.ID == "" {
		scenario.ID = fmt.Sprintf("scenario-%02d", index)
	}
	scenario.Name = strings.TrimSpace(sanitizeText(RedactSecrets(scenario.Name)))
	scenario.Type = strings.ToLower(strings.TrimSpace(scenario.Type))
	scenario.Priority = strings.ToLower(strings.TrimSpace(scenario.Priority))
	scenario.Risk = strings.ToLower(strings.TrimSpace(scenario.Risk))
	scenario.Description = strings.TrimSpace(sanitizeText(RedactSecrets(scenario.Description)))
	if scenario.Name == "" {
		return TestPlanScenario{}, fmt.Errorf("AI test plan scenario %d name is required", index)
	}
	if !allowedScenarioTypes[scenario.Type] {
		return TestPlanScenario{}, fmt.Errorf("AI test plan scenario %q has unsupported type %q", scenario.ID, scenario.Type)
	}
	if !validRiskLevel(scenario.Priority) {
		return TestPlanScenario{}, fmt.Errorf("AI test plan scenario %q priority must be low, medium, high, or critical", scenario.ID)
	}
	if !validRiskLevel(scenario.Risk) {
		return TestPlanScenario{}, fmt.Errorf("AI test plan scenario %q risk must be low, medium, high, or critical", scenario.ID)
	}
	if scenario.Description == "" {
		return TestPlanScenario{}, fmt.Errorf("AI test plan scenario %q description is required", scenario.ID)
	}
	scenario.ID = limitString(scenario.ID, 80)
	scenario.Name = limitString(scenario.Name, 200)
	scenario.Description = limitString(scenario.Description, 1200)
	scenario.Preconditions = sanitizeStringSlice(scenario.Preconditions, 20)
	scenario.Assertions = sanitizeStringSlice(scenario.Assertions, 20)
	scenario.TestDataNeeded = sanitizeStringSlice(scenario.TestDataNeeded, 20)
	scenario.RelatedFindings = sanitizeStringSlice(scenario.RelatedFindings, 20)
	scenario.Tags = sanitizeStringSlice(scenario.Tags, 20)
	if len(scenario.Steps) == 0 {
		return TestPlanScenario{}, fmt.Errorf("AI test plan scenario %q must include at least one step", scenario.ID)
	}
	if len(scenario.Steps) > 30 {
		scenario.Steps = scenario.Steps[:30]
	}
	for i := range scenario.Steps {
		scenario.Steps[i] = normalizeTestPlanStep(scenario.Steps[i], i+1)
	}
	return scenario, nil
}

func normalizeTestPlanStep(step TestPlanStep, order int) TestPlanStep {
	if step.Order <= 0 {
		step.Order = order
	}
	step.Action = limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.Action))), 800)
	step.Target = limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.Target))), 500)
	step.Data = limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.Data))), 500)
	step.ExpectedResult = limitString(strings.TrimSpace(sanitizeText(RedactSecrets(step.ExpectedResult))), 800)
	return step
}

func testPlanRiskLevel(payload *TestPlanPayload) string {
	rank := map[string]int{"": 0, "low": 1, "medium": 2, "high": 3, "critical": 4}
	result := ""
	for _, scenario := range payload.Scenarios {
		if rank[scenario.Risk] > rank[result] {
			result = scenario.Risk
		}
	}
	return result
}

func jsonArrayField(fields map[string]any, key string) []any {
	if fields == nil {
		return []any{}
	}
	value, ok := fields[key]
	if !ok {
		return []any{}
	}
	items, ok := value.([]any)
	if !ok {
		return []any{}
	}
	return items
}

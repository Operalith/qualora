package main

import "time"

const (
	StatusQueued    = "queued"
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
	StatusPassed    = "passed"
	StatusError     = "error"
	StatusSkipped   = "skipped"
)

const (
	JobKindBrowser = "browser"
	JobKindAPI     = "api"
)

const (
	AIProviderOpenAICompatible = "openai-compatible"
)

type CreateProjectRequest struct {
	Name                string   `json:"name"`
	FrontendURL         string   `json:"frontend_url"`
	APIBaseURL          string   `json:"api_base_url"`
	OpenAPIURL          string   `json:"openapi_url"`
	AllowedHosts        []string `json:"allowed_hosts"`
	SecurityMode        string   `json:"security_mode"`
	DestructiveActions  bool     `json:"destructive_actions"`
	AllowPrivateTargets bool     `json:"allow_private_targets,omitempty"`
}

type Project struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	FrontendURL         string    `json:"frontend_url"`
	APIBaseURL          string    `json:"api_base_url"`
	OpenAPIURL          string    `json:"openapi_url"`
	AllowedHosts        []string  `json:"allowed_hosts"`
	SecurityMode        string    `json:"security_mode"`
	DestructiveActions  bool      `json:"destructive_actions"`
	AllowPrivateTargets bool      `json:"allow_private_targets"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type TestRun struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	PageTitle    string     `json:"page_title,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RunJob struct {
	ID           string     `json:"id"`
	RunID        string     `json:"run_id"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Finding struct {
	ID                  string    `json:"id"`
	RunID               string    `json:"run_id,omitempty"`
	TestPlanExecutionID string    `json:"test_plan_execution_id,omitempty"`
	ScenarioExecutionID string    `json:"scenario_execution_id,omitempty"`
	StepExecutionID     string    `json:"step_execution_id,omitempty"`
	Title               string    `json:"title"`
	Severity            string    `json:"severity"`
	Category            string    `json:"category"`
	Confidence          string    `json:"confidence"`
	Description         string    `json:"description"`
	Recommendation      string    `json:"recommendation"`
	EvidenceIDs         []string  `json:"evidence_ids"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
}

type Evidence struct {
	ID                  string         `json:"id"`
	RunID               string         `json:"run_id,omitempty"`
	TestPlanExecutionID string         `json:"test_plan_execution_id,omitempty"`
	Type                string         `json:"type"`
	URI                 string         `json:"uri"`
	Metadata            map[string]any `json:"metadata"`
	CreatedAt           time.Time      `json:"created_at,omitempty"`
}

type Report struct {
	RunID      string         `json:"run_id"`
	ProjectID  string         `json:"project_id"`
	Status     string         `json:"status"`
	Summary    ReportSummary  `json:"summary"`
	Findings   []Finding      `json:"findings"`
	Evidence   []Evidence     `json:"evidence"`
	Metadata   map[string]any `json:"metadata"`
	AIAnalysis *AIAnalysis    `json:"ai_analysis"`
	TestPlans  []TestPlanRef  `json:"test_plans"`
}

type ReportSummary struct {
	TotalFindings int `json:"total_findings"`
	Critical      int `json:"critical"`
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Info          int `json:"info"`
}

type AIProvider struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Preset                 string    `json:"preset"`
	Type                   string    `json:"type"`
	BaseURL                string    `json:"base_url"`
	Model                  string    `json:"model"`
	APIKeyEncrypted        string    `json:"-"`
	ExtraHeadersEncrypted  string    `json:"-"`
	Temperature            float64   `json:"temperature"`
	MaxOutputTokens        int       `json:"max_output_tokens"`
	TimeoutSeconds         int       `json:"timeout_seconds"`
	SendScreenshots        bool      `json:"send_screenshots"`
	SendHTML               bool      `json:"send_html"`
	SendNetworkBodies      bool      `json:"send_network_bodies"`
	RedactionEnabled       bool      `json:"redaction_enabled"`
	IsDefault              bool      `json:"is_default"`
	APIKeyConfigured       bool      `json:"api_key_configured"`
	ExtraHeadersConfigured bool      `json:"extra_headers_configured"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type AIProviderRequest struct {
	Name              string            `json:"name"`
	Preset            string            `json:"preset"`
	Type              string            `json:"type"`
	BaseURL           string            `json:"base_url"`
	Model             string            `json:"model"`
	APIKey            string            `json:"api_key"`
	ExtraHeaders      map[string]string `json:"extra_headers"`
	Temperature       float64           `json:"temperature"`
	MaxOutputTokens   int               `json:"max_output_tokens"`
	TimeoutSeconds    int               `json:"timeout_seconds"`
	SendScreenshots   *bool             `json:"send_screenshots"`
	SendHTML          *bool             `json:"send_html"`
	SendNetworkBodies *bool             `json:"send_network_bodies"`
	RedactionEnabled  *bool             `json:"redaction_enabled"`
	IsDefault         bool              `json:"is_default"`
}

type AIProviderTestResult struct {
	Success      bool   `json:"success"`
	ProviderName string `json:"provider_name"`
	Model        string `json:"model"`
	LatencyMS    int64  `json:"latency_ms"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type AIAnalysisRequest struct {
	ProviderID string `json:"provider_id"`
}

type AIAnalysis struct {
	ID               string         `json:"id"`
	RunID            string         `json:"run_id"`
	ProviderID       string         `json:"provider_id,omitempty"`
	ProviderName     string         `json:"provider_name,omitempty"`
	Model            string         `json:"model"`
	Status           string         `json:"status"`
	ExecutiveSummary string         `json:"executive_summary"`
	TechnicalSummary string         `json:"technical_summary"`
	RiskLevel        string         `json:"risk_level"`
	AnalysisJSON     map[string]any `json:"analysis_json"`
	PromptTokens     int            `json:"prompt_tokens"`
	CompletionTokens int            `json:"completion_tokens"`
	TotalTokens      int            `json:"total_tokens"`
	ErrorMessage     string         `json:"error_message,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type AITestPlanRequest struct {
	ProviderID     string   `json:"provider_id"`
	RunID          string   `json:"run_id"`
	ProductContext string   `json:"product_context"`
	FocusAreas     []string `json:"focus_areas"`
	MaxScenarios   int      `json:"max_scenarios"`
}

type TestPlan struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	RunID          string         `json:"run_id,omitempty"`
	ProviderID     string         `json:"provider_id,omitempty"`
	ProviderName   string         `json:"provider_name,omitempty"`
	Model          string         `json:"model,omitempty"`
	Status         string         `json:"status"`
	Title          string         `json:"title"`
	Summary        string         `json:"summary"`
	PlanJSON       map[string]any `json:"plan_json"`
	RiskLevel      string         `json:"risk_level"`
	TotalScenarios int            `json:"total_scenarios"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type TestPlanRef struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	RiskLevel      string    `json:"risk_level"`
	TotalScenarios int       `json:"total_scenarios"`
	CreatedAt      time.Time `json:"created_at"`
}

type TestPlanExecutionRequest struct {
	MaxScenarios        int      `json:"max_scenarios"`
	MaxStepsPerScenario int      `json:"max_steps_per_scenario"`
	ScenarioIDs         []string `json:"scenario_ids"`
	DryRun              bool     `json:"dry_run"`
}

type TestPlanExecutionPreview struct {
	DryRun              bool                          `json:"dry_run"`
	TestPlanID          string                        `json:"test_plan_id"`
	ProjectID           string                        `json:"project_id"`
	MaxScenarios        int                           `json:"max_scenarios"`
	MaxStepsPerScenario int                           `json:"max_steps_per_scenario"`
	TotalScenarios      int                           `json:"total_scenarios"`
	ExecutableScenarios int                           `json:"executable_scenarios"`
	SkippedScenarios    int                           `json:"skipped_scenarios"`
	TotalSteps          int                           `json:"total_steps"`
	ExecutableSteps     int                           `json:"executable_steps"`
	SkippedSteps        int                           `json:"skipped_steps"`
	Scenarios           []MappedExecutionScenario     `json:"scenarios"`
	SafetySummary       TestPlanExecutionSafetyReport `json:"safety_summary"`
}

type MappedExecutionScenario struct {
	ScenarioIDFromPlan string                `json:"scenario_id_from_plan"`
	Name               string                `json:"name"`
	Type               string                `json:"type"`
	Priority           string                `json:"priority"`
	Status             string                `json:"status"`
	SkipReason         string                `json:"skip_reason,omitempty"`
	Steps              []MappedExecutionStep `json:"steps"`
}

type MappedExecutionStep struct {
	StepOrder      int    `json:"step_order"`
	OriginalAction string `json:"original_action"`
	MappedAction   string `json:"mapped_action"`
	Target         string `json:"target"`
	ExpectedResult string `json:"expected_result"`
	Status         string `json:"status"`
	SkipReason     string `json:"skip_reason,omitempty"`
}

type TestPlanExecution struct {
	ID               string     `json:"id"`
	TestPlanID       string     `json:"test_plan_id"`
	ProjectID        string     `json:"project_id"`
	SourceRunID      string     `json:"source_run_id,omitempty"`
	Status           string     `json:"status"`
	TotalScenarios   int        `json:"total_scenarios"`
	PassedScenarios  int        `json:"passed_scenarios"`
	FailedScenarios  int        `json:"failed_scenarios"`
	SkippedScenarios int        `json:"skipped_scenarios"`
	TotalSteps       int        `json:"total_steps"`
	PassedSteps      int        `json:"passed_steps"`
	FailedSteps      int        `json:"failed_steps"`
	SkippedSteps     int        `json:"skipped_steps"`
	ErrorMessage     string     `json:"error_message,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TestPlanExecutionScenario struct {
	ID                 string                  `json:"id"`
	ExecutionID        string                  `json:"execution_id"`
	ScenarioIDFromPlan string                  `json:"scenario_id_from_plan"`
	Name               string                  `json:"name"`
	Type               string                  `json:"type"`
	Priority           string                  `json:"priority"`
	Status             string                  `json:"status"`
	SkipReason         string                  `json:"skip_reason,omitempty"`
	StartedAt          *time.Time              `json:"started_at,omitempty"`
	CompletedAt        *time.Time              `json:"completed_at,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
	Steps              []TestPlanExecutionStep `json:"steps,omitempty"`
}

type TestPlanExecutionStep struct {
	ID                  string    `json:"id"`
	ExecutionID         string    `json:"execution_id"`
	ScenarioExecutionID string    `json:"scenario_execution_id"`
	StepOrder           int       `json:"step_order"`
	OriginalAction      string    `json:"original_action"`
	MappedAction        string    `json:"mapped_action"`
	Target              string    `json:"target"`
	ExpectedResult      string    `json:"expected_result"`
	Status              string    `json:"status"`
	SkipReason          string    `json:"skip_reason,omitempty"`
	ActualResult        string    `json:"actual_result,omitempty"`
	ErrorMessage        string    `json:"error_message,omitempty"`
	DurationMS          *int      `json:"duration_ms,omitempty"`
	EvidenceID          string    `json:"evidence_id,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type TestPlanExecutionDetail struct {
	Execution TestPlanExecution           `json:"execution"`
	Scenarios []TestPlanExecutionScenario `json:"scenarios"`
}

type TestPlanExecutionSafetyReport struct {
	ExecutedSteps           int `json:"executed_steps"`
	SkippedUnsafeSteps      int `json:"skipped_unsafe_steps"`
	SkippedUnsupportedSteps int `json:"skipped_unsupported_steps"`
	SkippedScenarios        int `json:"skipped_scenarios"`
}

type TestPlanExecutionReport struct {
	Execution     TestPlanExecution             `json:"execution"`
	TestPlan      TestPlan                      `json:"test_plan"`
	Project       Project                       `json:"project"`
	Scenarios     []TestPlanExecutionScenario   `json:"scenarios"`
	Findings      []Finding                     `json:"findings"`
	Evidence      []Evidence                    `json:"evidence"`
	SafetySummary TestPlanExecutionSafetyReport `json:"safety_summary"`
	GeneratedAt   time.Time                     `json:"generated_at"`
}

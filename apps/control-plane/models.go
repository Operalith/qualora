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
	RunTypeFull                      = "full"
	RunTypeBrowserSmoke              = "browser_smoke"
	RunTypeAPISmoke                  = "api_smoke"
	RunTypeLoginCheck                = "login_check"
	RunTypeAuthenticatedBrowserSmoke = "authenticated_browser_smoke"
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
	ID                  string     `json:"id"`
	ProjectID           string     `json:"project_id"`
	RunType             string     `json:"run_type"`
	APISpecID           string     `json:"api_spec_id,omitempty"`
	CredentialProfileID string     `json:"credential_profile_id,omitempty"`
	TargetPath          string     `json:"target_path,omitempty"`
	CaptureScreenshot   bool       `json:"capture_screenshot"`
	MaxDurationSeconds  int        `json:"max_duration_seconds"`
	Status              string     `json:"status"`
	ErrorMessage        string     `json:"error_message,omitempty"`
	PageTitle           string     `json:"page_title,omitempty"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
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
	RunID        string           `json:"run_id"`
	ProjectID    string           `json:"project_id"`
	RunType      string           `json:"run_type"`
	Status       string           `json:"status"`
	Summary      ReportSummary    `json:"summary"`
	Findings     []Finding        `json:"findings"`
	Evidence     []Evidence       `json:"evidence"`
	Metadata     map[string]any   `json:"metadata"`
	AIAnalysis   *AIAnalysis      `json:"ai_analysis"`
	TestPlans    []TestPlanRef    `json:"test_plans"`
	APISpec      *APISpec         `json:"api_spec,omitempty"`
	APISummary   *APISmokeSummary `json:"api_summary,omitempty"`
	APIResults   []APICheckResult `json:"api_results,omitempty"`
	LoginSummary *LoginSummary    `json:"login_summary,omitempty"`
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

type CredentialProfile struct {
	ID                  string    `json:"id"`
	ProjectID           string    `json:"project_id"`
	Name                string    `json:"name"`
	Type                string    `json:"type"`
	UsernameEncrypted   string    `json:"-"`
	PasswordEncrypted   string    `json:"-"`
	UsernameConfigured  bool      `json:"username_configured"`
	PasswordConfigured  bool      `json:"password_configured"`
	UsernameDisplayHint string    `json:"username_display_hint,omitempty"`
	LoginURL            string    `json:"login_url"`
	UsernameSelector    string    `json:"username_selector"`
	PasswordSelector    string    `json:"password_selector"`
	SubmitSelector      string    `json:"submit_selector"`
	SuccessURLContains  string    `json:"success_url_contains,omitempty"`
	SuccessTextContains string    `json:"success_text_contains,omitempty"`
	FailureTextContains string    `json:"failure_text_contains,omitempty"`
	PostLoginWaitMS     int       `json:"post_login_wait_ms"`
	IsDefault           bool      `json:"is_default"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CredentialProfileRequest struct {
	Name                string `json:"name"`
	Type                string `json:"type"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	LoginURL            string `json:"login_url"`
	UsernameSelector    string `json:"username_selector"`
	PasswordSelector    string `json:"password_selector"`
	SubmitSelector      string `json:"submit_selector"`
	SuccessURLContains  string `json:"success_url_contains"`
	SuccessTextContains string `json:"success_text_contains"`
	FailureTextContains string `json:"failure_text_contains"`
	PostLoginWaitMS     int    `json:"post_login_wait_ms"`
	IsDefault           bool   `json:"is_default"`
}

type AuthenticatedBrowserSmokeRequest struct {
	CredentialProfileID string `json:"credential_profile_id"`
	TargetPath          string `json:"target_path"`
	CaptureScreenshot   *bool  `json:"capture_screenshot"`
	MaxDurationSeconds  int    `json:"max_duration_seconds"`
}

type LoginSummary struct {
	CredentialProfileID    string `json:"credential_profile_id,omitempty"`
	CredentialProfileName  string `json:"credential_profile_name,omitempty"`
	LoginStatus            string `json:"login_status,omitempty"`
	LoginURL               string `json:"login_url,omitempty"`
	LoginFinalURL          string `json:"login_final_url,omitempty"`
	PageTitle              string `json:"page_title,omitempty"`
	LoginDurationMS        int    `json:"login_duration_ms,omitempty"`
	AuthenticatedTargetURL string `json:"authenticated_target_url,omitempty"`
	FailureReason          string `json:"failure_reason,omitempty"`
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

type APISpecImportRequest struct {
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
	SourceURL  string `json:"source_url"`
	RawSpec    string `json:"raw_spec"`
}

type APISpec struct {
	ID                    string    `json:"id"`
	ProjectID             string    `json:"project_id"`
	Name                  string    `json:"name"`
	SourceType            string    `json:"source_type"`
	SourceURL             string    `json:"source_url,omitempty"`
	ParsedTitle           string    `json:"parsed_title,omitempty"`
	ParsedVersion         string    `json:"parsed_version,omitempty"`
	ServerURL             string    `json:"server_url,omitempty"`
	OperationCount        int       `json:"operation_count"`
	SafeOperationCount    int       `json:"safe_operation_count"`
	SkippedOperationCount int       `json:"skipped_operation_count"`
	Status                string    `json:"status"`
	ErrorMessage          string    `json:"error_message,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type APISpecDetail struct {
	Spec       APISpec        `json:"spec"`
	Operations []APIOperation `json:"operations,omitempty"`
}

type APIOperation struct {
	ID                     string    `json:"id"`
	APISpecID              string    `json:"api_spec_id"`
	ProjectID              string    `json:"project_id"`
	Method                 string    `json:"method"`
	Path                   string    `json:"path"`
	ResolvedPath           string    `json:"resolved_path,omitempty"`
	QueryString            string    `json:"query_string,omitempty"`
	OperationID            string    `json:"operation_id,omitempty"`
	Summary                string    `json:"summary,omitempty"`
	Description            string    `json:"description,omitempty"`
	Tags                   []string  `json:"tags"`
	ExpectedStatuses       []string  `json:"expected_statuses"`
	ExpectedContentTypes   []string  `json:"expected_content_types"`
	RequiresAuthentication *bool     `json:"requires_authentication,omitempty"`
	SafeToExecute          bool      `json:"safe_to_execute"`
	SkipReason             string    `json:"skip_reason,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type APICheckResult struct {
	ID                  string    `json:"id"`
	RunID               string    `json:"run_id"`
	APISpecID           string    `json:"api_spec_id"`
	OperationID         string    `json:"operation_id,omitempty"`
	Method              string    `json:"method"`
	Path                string    `json:"path"`
	ResolvedURL         string    `json:"resolved_url,omitempty"`
	Status              string    `json:"status"`
	HTTPStatus          *int      `json:"http_status,omitempty"`
	DurationMS          *int      `json:"duration_ms,omitempty"`
	ResponseContentType string    `json:"response_content_type,omitempty"`
	ResponseSizeBytes   *int      `json:"response_size_bytes,omitempty"`
	ErrorMessage        string    `json:"error_message,omitempty"`
	SkippedReason       string    `json:"skipped_reason,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type APISmokeSummary struct {
	TotalOperations    int `json:"total_operations"`
	ExecutedOperations int `json:"executed_operations"`
	SkippedOperations  int `json:"skipped_operations"`
	PassedOperations   int `json:"passed_operations"`
	FailedOperations   int `json:"failed_operations"`
	ErroredOperations  int `json:"errored_operations"`
}

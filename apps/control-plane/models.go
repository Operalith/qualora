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
	RunTypeAppDiscovery              = "app_discovery"
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

type SetupStatusResponse struct {
	SetupRequired bool   `json:"setup_required"`
	Version       string `json:"version"`
}

type SetupAdminRequest struct {
	Email           string `json:"email"`
	DisplayName     string `json:"display_name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LocalUser struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type AuthUser struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Role        string     `json:"role"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type UserSession struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	TokenHash     string     `json:"-"`
	CSRFTokenHash string     `json:"-"`
	UserAgent     string     `json:"user_agent,omitempty"`
	IPAddress     string     `json:"ip_address,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	LastSeenAt    time.Time  `json:"last_seen_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

type AuthResponse struct {
	User        AuthUser  `json:"user"`
	ExpiresAt   time.Time `json:"expires_at"`
	CSRFToken   string    `json:"csrf_token,omitempty"`
	SetupStatus string    `json:"setup_status,omitempty"`
}

type MeResponse struct {
	Authenticated bool       `json:"authenticated"`
	User          *AuthUser  `json:"user,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
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
	AuthorizationRunID  string    `json:"authorization_check_run_id,omitempty"`
	DiscoveryRunID      string    `json:"discovery_run_id,omitempty"`
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
	AuthorizationRunID  string         `json:"authorization_check_run_id,omitempty"`
	DiscoveryRunID      string         `json:"discovery_run_id,omitempty"`
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

type DiscoveryRunRequest struct {
	StartURL            string `json:"start_url,omitempty"`
	CredentialProfileID string `json:"credential_profile_id,omitempty"`
	MaxPages            int    `json:"max_pages,omitempty"`
	MaxDepth            int    `json:"max_depth,omitempty"`
	SameOriginOnly      *bool  `json:"same_origin_only,omitempty"`
}

type DiscoveryRun struct {
	ID                  string     `json:"id"`
	ProjectID           string     `json:"project_id"`
	CredentialProfileID string     `json:"credential_profile_id,omitempty"`
	Status              string     `json:"status"`
	StartURL            string     `json:"start_url"`
	MaxPages            int        `json:"max_pages"`
	MaxDepth            int        `json:"max_depth"`
	SameOriginOnly      bool       `json:"same_origin_only"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	TotalPages          int        `json:"total_pages"`
	TotalLinks          int        `json:"total_links"`
	TotalForms          int        `json:"total_forms"`
	TotalConsoleErrors  int        `json:"total_console_errors"`
	TotalFailedRequests int        `json:"total_failed_requests"`
	TotalFindings       int        `json:"total_findings"`
	ErrorMessage        string     `json:"error_message,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type DiscoveredPage struct {
	ID                   string    `json:"id"`
	DiscoveryRunID       string    `json:"discovery_run_id"`
	ProjectID            string    `json:"project_id"`
	URL                  string    `json:"url"`
	NormalizedURL        string    `json:"normalized_url"`
	Path                 string    `json:"path"`
	Title                string    `json:"title,omitempty"`
	HTTPStatus           *int      `json:"http_status,omitempty"`
	ContentType          string    `json:"content_type,omitempty"`
	BodyTextLength       *int      `json:"body_text_length,omitempty"`
	LoadDurationMS       *int      `json:"load_duration_ms,omitempty"`
	Depth                int       `json:"depth"`
	ScreenshotEvidenceID string    `json:"screenshot_evidence_id,omitempty"`
	ConsoleErrorCount    int       `json:"console_error_count"`
	FailedRequestCount   int       `json:"failed_request_count"`
	DiscoveredAt         time.Time `json:"discovered_at"`
	CreatedAt            time.Time `json:"created_at"`
}

type DiscoveredLink struct {
	ID             string    `json:"id"`
	DiscoveryRunID string    `json:"discovery_run_id"`
	SourcePageID   string    `json:"source_page_id"`
	Href           string    `json:"href"`
	NormalizedURL  string    `json:"normalized_url,omitempty"`
	LinkText       string    `json:"link_text,omitempty"`
	SameOrigin     bool      `json:"same_origin"`
	Skipped        bool      `json:"skipped"`
	SkipReason     string    `json:"skip_reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type DiscoveredForm struct {
	ID                 string                `json:"id"`
	DiscoveryRunID     string                `json:"discovery_run_id"`
	PageID             string                `json:"page_id"`
	FormName           string                `json:"form_name,omitempty"`
	FormAction         string                `json:"form_action,omitempty"`
	FormMethod         string                `json:"form_method,omitempty"`
	FieldCount         int                   `json:"field_count"`
	PasswordFieldCount int                   `json:"password_field_count"`
	SubmitButtonCount  int                   `json:"submit_button_count"`
	Classification     string                `json:"classification,omitempty"`
	SkippedReason      string                `json:"skipped_reason,omitempty"`
	Fields             []DiscoveredFormField `json:"fields,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
}

type DiscoveredFormField struct {
	ID          string    `json:"id"`
	FormID      string    `json:"form_id"`
	FieldName   string    `json:"field_name,omitempty"`
	FieldType   string    `json:"field_type,omitempty"`
	Placeholder string    `json:"placeholder,omitempty"`
	Label       string    `json:"label,omitempty"`
	Required    bool      `json:"required"`
	CreatedAt   time.Time `json:"created_at"`
}

type DiscoverySummary struct {
	TotalPages           int `json:"total_pages"`
	TotalLinks           int `json:"total_links"`
	TotalForms           int `json:"total_forms"`
	TotalConsoleErrors   int `json:"total_console_errors"`
	TotalFailedRequests  int `json:"total_failed_requests"`
	TotalFindings        int `json:"total_findings"`
	SkippedLinks         int `json:"skipped_links"`
	ExternalLinksSkipped int `json:"external_links_skipped"`
	UnsafeLinksSkipped   int `json:"unsafe_links_skipped"`
	PagesWithScreenshots int `json:"pages_with_screenshots"`
}

type DiscoveryMap struct {
	Run      DiscoveryRun     `json:"run"`
	Project  Project          `json:"project"`
	Summary  DiscoverySummary `json:"summary"`
	Pages    []DiscoveredPage `json:"pages"`
	Links    []DiscoveredLink `json:"links"`
	Forms    []DiscoveredForm `json:"forms"`
	Findings []Finding        `json:"findings"`
	Evidence []Evidence       `json:"evidence"`
}

type DiscoveryReport struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Run         DiscoveryRun     `json:"run"`
	Project     Project          `json:"project"`
	Settings    map[string]any   `json:"settings"`
	Summary     DiscoverySummary `json:"summary"`
	Pages       []DiscoveredPage `json:"pages"`
	Links       []DiscoveredLink `json:"links"`
	Forms       []DiscoveredForm `json:"forms"`
	Findings    []Finding        `json:"findings"`
	Evidence    []Evidence       `json:"evidence"`
	SafetyNotes []string         `json:"safety_notes"`
	Limitations []string         `json:"limitations"`
	Metadata    map[string]any   `json:"metadata"`
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
	RoleName            string    `json:"role_name,omitempty"`
	RoleDescription     string    `json:"role_description,omitempty"`
	SubjectLabel        string    `json:"subject_label,omitempty"`
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
	RoleName            string `json:"role_name"`
	RoleDescription     string `json:"role_description"`
	SubjectLabel        string `json:"subject_label"`
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
	RoleName               string `json:"role_name,omitempty"`
	LoginStatus            string `json:"login_status,omitempty"`
	LoginURL               string `json:"login_url,omitempty"`
	LoginFinalURL          string `json:"login_final_url,omitempty"`
	PageTitle              string `json:"page_title,omitempty"`
	LoginDurationMS        int    `json:"login_duration_ms,omitempty"`
	AuthenticatedTargetURL string `json:"authenticated_target_url,omitempty"`
	FailureReason          string `json:"failure_reason,omitempty"`
}

const (
	AuthorizationCheckTypeBrowserURL = "browser_url"
	AuthorizationCheckTypeAPIGet     = "api_get"

	AuthorizationExpectedAllowed = "allowed"
	AuthorizationExpectedDenied  = "denied"

	AuthorizationActualAllowed = "allowed"
	AuthorizationActualDenied  = "denied"
	AuthorizationActualUnknown = "unknown"
)

type AuthorizationCheck struct {
	ID                       string    `json:"id"`
	ProjectID                string    `json:"project_id"`
	Name                     string    `json:"name"`
	Description              string    `json:"description,omitempty"`
	Type                     string    `json:"type"`
	ResourceLabel            string    `json:"resource_label,omitempty"`
	OwnerCredentialProfileID string    `json:"owner_credential_profile_id,omitempty"`
	ActorCredentialProfileID string    `json:"actor_credential_profile_id"`
	ExpectedOutcome          string    `json:"expected_outcome"`
	TargetURL                string    `json:"target_url,omitempty"`
	APISpecID                string    `json:"api_spec_id,omitempty"`
	APIOperationID           string    `json:"api_operation_id,omitempty"`
	Method                   string    `json:"method,omitempty"`
	Path                     string    `json:"path,omitempty"`
	ExpectedStatuses         []int     `json:"expected_statuses,omitempty"`
	SuccessTextContains      string    `json:"success_text_contains,omitempty"`
	DeniedStatuses           []int     `json:"denied_statuses,omitempty"`
	DeniedTextContains       string    `json:"denied_text_contains,omitempty"`
	Enabled                  bool      `json:"enabled"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type AuthorizationCheckRequest struct {
	Name                     string `json:"name"`
	Description              string `json:"description"`
	Type                     string `json:"type"`
	ResourceLabel            string `json:"resource_label"`
	OwnerCredentialProfileID string `json:"owner_credential_profile_id"`
	ActorCredentialProfileID string `json:"actor_credential_profile_id"`
	ExpectedOutcome          string `json:"expected_outcome"`
	TargetURL                string `json:"target_url"`
	APISpecID                string `json:"api_spec_id"`
	APIOperationID           string `json:"api_operation_id"`
	Method                   string `json:"method"`
	Path                     string `json:"path"`
	ExpectedStatuses         []int  `json:"expected_statuses"`
	SuccessTextContains      string `json:"success_text_contains"`
	DeniedStatuses           []int  `json:"denied_statuses"`
	DeniedTextContains       string `json:"denied_text_contains"`
	Enabled                  *bool  `json:"enabled"`
}

type AuthorizationCheckRunRequest struct {
	CheckIDs  []string `json:"check_ids"`
	MaxChecks int      `json:"max_checks"`
}

type AuthorizationCheckRun struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Status        string     `json:"status"`
	CheckIDs      []string   `json:"check_ids,omitempty"`
	MaxChecks     int        `json:"max_checks"`
	TotalChecks   int        `json:"total_checks"`
	PassedChecks  int        `json:"passed_checks"`
	FailedChecks  int        `json:"failed_checks"`
	SkippedChecks int        `json:"skipped_checks"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AuthorizationCheckResult struct {
	ID                       string    `json:"id"`
	RunID                    string    `json:"run_id"`
	CheckID                  string    `json:"check_id"`
	Status                   string    `json:"status"`
	ExpectedOutcome          string    `json:"expected_outcome"`
	ActualOutcome            string    `json:"actual_outcome"`
	ActorCredentialProfileID string    `json:"actor_credential_profile_id"`
	ActorRoleName            string    `json:"actor_role_name,omitempty"`
	TargetURL                string    `json:"target_url,omitempty"`
	FinalURL                 string    `json:"final_url,omitempty"`
	HTTPStatus               *int      `json:"http_status,omitempty"`
	PageTitle                string    `json:"page_title,omitempty"`
	DurationMS               *int      `json:"duration_ms,omitempty"`
	EvidenceID               string    `json:"evidence_id,omitempty"`
	FindingID                string    `json:"finding_id,omitempty"`
	SkipReason               string    `json:"skip_reason,omitempty"`
	ErrorMessage             string    `json:"error_message,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
}

type AuthorizationCheckDetail struct {
	Run     AuthorizationCheckRun      `json:"run"`
	Checks  []AuthorizationCheck       `json:"checks,omitempty"`
	Results []AuthorizationCheckResult `json:"results"`
}

type AuthorizationCheckReport struct {
	Run         AuthorizationCheckRun      `json:"run"`
	Project     Project                    `json:"project"`
	Checks      []AuthorizationCheck       `json:"checks"`
	Results     []AuthorizationCheckResult `json:"results"`
	Summary     ReportSummary              `json:"summary"`
	Findings    []Finding                  `json:"findings"`
	Evidence    []Evidence                 `json:"evidence"`
	Metadata    map[string]any             `json:"metadata"`
	GeneratedAt time.Time                  `json:"generated_at"`
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

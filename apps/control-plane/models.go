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
	QARunStatusRunningDiscovery    = "running_discovery"
	QARunStatusRunningQuality      = "running_quality_checks"
	QARunStatusGeneratingPlan      = "generating_plan"
	QARunStatusPreviewingExecution = "previewing_execution"
	QARunStatusExecutingPlan       = "executing_plan"
)

const (
	AITestPlanExecutionModeReviewOnly     = "review_only"
	AITestPlanExecutionModeSafeExecutable = "safe_executable"
	TestPlanSourceRunReport               = "run_report"
	TestPlanSourceDiscovery               = "discovery"
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
	RunTypeQualityCheck              = "quality_check"
	RunTypeSafeExplorer              = "safe_explorer"
)

const (
	ReportTypeSafeQA        = "safe_qa"
	ReportTypeQualityCheck  = "quality_check"
	ReportTypeDiscovery     = "discovery"
	ReportTypeSafeExplorer  = "safe_explorer"
	ReportTypeAPISmoke      = "api_smoke"
	ReportTypeBrowserSmoke  = "browser_smoke"
	ReportTypeAuthorization = "authorization"
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

type ProjectSetupRequest struct {
	Project    CreateProjectRequest          `json:"project"`
	AI         ProjectSetupAIConfig          `json:"ai,omitempty"`
	Credential ProjectSetupCredentialConfig  `json:"credential,omitempty"`
	APISpec    ProjectSetupAPISpecConfig     `json:"api_spec,omitempty"`
	Workflow   ProjectSetupWorkflowSelection `json:"workflow,omitempty"`
}

type ProjectSetupAIConfig struct {
	Mode       string             `json:"mode,omitempty"`
	ProviderID string             `json:"provider_id,omitempty"`
	Provider   *AIProviderRequest `json:"provider,omitempty"`
}

type ProjectSetupCredentialConfig struct {
	Mode    string                    `json:"mode,omitempty"`
	Profile *CredentialProfileRequest `json:"profile,omitempty"`
}

type ProjectSetupAPISpecConfig struct {
	Mode string                `json:"mode,omitempty"`
	Spec *APISpecImportRequest `json:"spec,omitempty"`
}

type ProjectSetupWorkflowSelection struct {
	BrowserSmoke       bool `json:"browser_smoke,omitempty"`
	Discovery          bool `json:"discovery,omitempty"`
	QualityChecks      bool `json:"quality_checks,omitempty"`
	SafeQARun          bool `json:"safe_qa_run,omitempty"`
	ExecuteSafeQA      bool `json:"execute_safe_qa,omitempty"`
	APISmoke           bool `json:"api_smoke,omitempty"`
	AuthenticatedSmoke bool `json:"authenticated_smoke,omitempty"`
	UseDefaults        bool `json:"use_defaults,omitempty"`
}

type ProjectSetupResponse struct {
	Project           Project                    `json:"project"`
	AIProvider        *AIProvider                `json:"ai_provider,omitempty"`
	CredentialProfile *CredentialProfile         `json:"credential_profile,omitempty"`
	APISpec           *APISpec                   `json:"api_spec,omitempty"`
	Started           ProjectSetupStartedActions `json:"started"`
	Skipped           []ProjectSetupSkipped      `json:"skipped"`
	Timeline          []ProjectSetupTimelineItem `json:"timeline"`
	NextLinks         map[string]string          `json:"next_links"`
}

type ProjectSetupStartedActions struct {
	BrowserSmokeRunID       string `json:"browser_smoke_run_id,omitempty"`
	AuthenticatedSmokeRunID string `json:"authenticated_smoke_run_id,omitempty"`
	DiscoveryRunID          string `json:"discovery_run_id,omitempty"`
	QualityCheckRunID       string `json:"quality_check_run_id,omitempty"`
	SafeQARunID             string `json:"safe_qa_run_id,omitempty"`
	APISmokeRunID           string `json:"api_smoke_run_id,omitempty"`
	AIProviderID            string `json:"ai_provider_id,omitempty"`
	CredentialProfileID     string `json:"credential_profile_id,omitempty"`
	APISpecID               string `json:"api_spec_id,omitempty"`
}

type ProjectSetupSkipped struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type ProjectSetupTimelineItem struct {
	Step     string `json:"step"`
	Status   string `json:"status"`
	Resource string `json:"resource,omitempty"`
	Reason   string `json:"reason,omitempty"`
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
	SafeExplorerRunID   string    `json:"safe_explorer_run_id,omitempty"`
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
	SafeExplorerRunID   string         `json:"safe_explorer_run_id,omitempty"`
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
	ReportIntelligence
}

type ReportSummary struct {
	TotalFindings int `json:"total_findings"`
	Critical      int `json:"critical"`
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Info          int `json:"info"`
}

type ReportBaseline struct {
	ID                   string           `json:"id"`
	ProjectID            string           `json:"project_id"`
	Name                 string           `json:"name"`
	Description          string           `json:"description,omitempty"`
	ReportType           string           `json:"report_type"`
	ReportID             string           `json:"report_id"`
	SourceRunID          string           `json:"source_run_id,omitempty"`
	FingerprintSet       []GroupedFinding `json:"fingerprint_set"`
	SeverityCounts       ReportSummary    `json:"severity_counts"`
	GroupedFindingsCount int              `json:"grouped_findings_count"`
	RawFindingsCount     int              `json:"raw_findings_count"`
	CreatedByUserID      string           `json:"created_by_user_id,omitempty"`
	IsDefault            bool             `json:"is_default"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type ReportBaselineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ReportType  string `json:"report_type"`
	ReportID    string `json:"report_id"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

type ReportBaselineUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsDefault   *bool  `json:"is_default,omitempty"`
}

type ReportComparisonRequest struct {
	ReportType         string `json:"report_type"`
	CurrentReportID    string `json:"current_report_id"`
	BaselineID         string `json:"baseline_id,omitempty"`
	UseDefaultBaseline bool   `json:"use_default_baseline,omitempty"`
}

type SeverityChange struct {
	Fingerprint      string `json:"fingerprint"`
	Title            string `json:"title"`
	PreviousSeverity string `json:"previous_severity"`
	CurrentSeverity  string `json:"current_severity"`
}

type AffectedScopeChange struct {
	Fingerprint           string `json:"fingerprint"`
	Title                 string `json:"title"`
	PreviousAffectedURLs  int    `json:"previous_affected_urls"`
	CurrentAffectedURLs   int    `json:"current_affected_urls"`
	PreviousAffectedPaths int    `json:"previous_affected_paths"`
	CurrentAffectedPaths  int    `json:"current_affected_paths"`
}

type ReportComparisonSummary struct {
	NewFindingsCount       int              `json:"new_findings_count"`
	FixedFindingsCount     int              `json:"fixed_findings_count"`
	UnchangedFindingsCount int              `json:"unchanged_findings_count"`
	SeverityChanges        []SeverityChange `json:"severity_changes"`
	NewCritical            int              `json:"new_critical"`
	NewHigh                int              `json:"new_high"`
	NewMedium              int              `json:"new_medium"`
	FixedCritical          int              `json:"fixed_critical"`
	FixedHigh              int              `json:"fixed_high"`
	FixedMedium            int              `json:"fixed_medium"`
}

type ReportComparison struct {
	ComparisonID       string                  `json:"comparison_id,omitempty"`
	ProjectID          string                  `json:"project_id"`
	ReportType         string                  `json:"report_type"`
	BaselineID         string                  `json:"baseline_id,omitempty"`
	CurrentReportID    string                  `json:"current_report_id"`
	Status             string                  `json:"status"`
	Summary            ReportComparisonSummary `json:"summary"`
	NewFindings        []GroupedFinding        `json:"new_findings"`
	FixedFindings      []GroupedFinding        `json:"fixed_findings"`
	UnchangedFindings  []GroupedFinding        `json:"unchanged_findings"`
	SeverityDelta      ReportSummary           `json:"severity_delta"`
	AffectedPagesDelta []AffectedScopeChange   `json:"affected_pages_delta"`
	Recommendation     string                  `json:"recommendation"`
	GeneratedAt        time.Time               `json:"generated_at"`
}

type QualityGateConfig struct {
	FailOnNewCritical   *bool `json:"fail_on_new_critical,omitempty"`
	FailOnNewHigh       *bool `json:"fail_on_new_high,omitempty"`
	FailOnNewMedium     *bool `json:"fail_on_new_medium,omitempty"`
	MaxNewHigh          *int  `json:"max_new_high,omitempty"`
	MaxNewMedium        *int  `json:"max_new_medium,omitempty"`
	MaxTotalCritical    *int  `json:"max_total_critical,omitempty"`
	MaxTotalHigh        *int  `json:"max_total_high,omitempty"`
	FailOnRunError      *bool `json:"fail_on_run_error,omitempty"`
	FailOnMissingReport *bool `json:"fail_on_missing_report,omitempty"`
	IgnoreInfo          *bool `json:"ignore_info,omitempty"`
	IgnoreNoisy         *bool `json:"ignore_noisy,omitempty"`
}

type QualityGateEvaluationRequest struct {
	ReportType         string            `json:"report_type"`
	CurrentReportID    string            `json:"current_report_id"`
	BaselineID         string            `json:"baseline_id,omitempty"`
	UseDefaultBaseline bool              `json:"use_default_baseline,omitempty"`
	GateConfig         QualityGateConfig `json:"gate_config,omitempty"`
	Format             string            `json:"format,omitempty"`
}

type QualityGateResult struct {
	Status            string                  `json:"status"`
	FailedRules       []string                `json:"failed_rules"`
	Warnings          []string                `json:"warnings"`
	ComparisonSummary ReportComparisonSummary `json:"comparison_summary"`
	SeverityCounts    ReportSummary           `json:"severity_counts"`
	Recommendation    string                  `json:"recommendation"`
	CIExitCode        int                     `json:"ci_exit_code"`
	GeneratedAt       time.Time               `json:"generated_at"`
}

type CIQualityGateResult struct {
	Status        string   `json:"status"`
	ExitCode      int      `json:"exit_code"`
	Summary       string   `json:"summary"`
	ReportURL     string   `json:"report_url"`
	ComparisonURL string   `json:"comparison_url,omitempty"`
	FailedRules   []string `json:"failed_rules"`
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
	ReportIntelligence
}

type SafeExplorerRunRequest struct {
	StartURL            string `json:"start_url,omitempty"`
	CredentialProfileID string `json:"credential_profile_id,omitempty"`
	MaxSteps            int    `json:"max_steps,omitempty"`
	MaxDepth            int    `json:"max_depth,omitempty"`
	SameOriginOnly      *bool  `json:"same_origin_only,omitempty"`
	AllowGetForms       bool   `json:"allow_get_forms,omitempty"`
}

type SafeExplorerRun struct {
	ID                   string     `json:"id"`
	ProjectID            string     `json:"project_id"`
	CredentialProfileID  string     `json:"credential_profile_id,omitempty"`
	Status               string     `json:"status"`
	StartURL             string     `json:"start_url"`
	MaxSteps             int        `json:"max_steps"`
	MaxDepth             int        `json:"max_depth"`
	SameOriginOnly       bool       `json:"same_origin_only"`
	AllowGetForms        bool       `json:"allow_get_forms"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	TotalSteps           int        `json:"total_steps"`
	TotalPagesObserved   int        `json:"total_pages_observed"`
	TotalActionsDetected int        `json:"total_actions_detected"`
	TotalActionsExecuted int        `json:"total_actions_executed"`
	TotalActionsSkipped  int        `json:"total_actions_skipped"`
	TotalFindings        int        `json:"total_findings"`
	ErrorMessage         string     `json:"error_message,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type SafeExplorerStep struct {
	ID                   string    `json:"id"`
	RunID                string    `json:"run_id"`
	ProjectID            string    `json:"project_id"`
	StepIndex            int       `json:"step_index"`
	PageURL              string    `json:"page_url"`
	NormalizedURL        string    `json:"normalized_url"`
	PageTitle            string    `json:"page_title,omitempty"`
	Depth                int       `json:"depth"`
	ActionID             string    `json:"action_id,omitempty"`
	ActionType           string    `json:"action_type,omitempty"`
	ActionLabel          string    `json:"action_label,omitempty"`
	ActionSelectorHint   string    `json:"action_selector_hint,omitempty"`
	ActionTargetURL      string    `json:"action_target_url,omitempty"`
	ActionSafety         string    `json:"action_safety"`
	ActionDecision       string    `json:"action_decision"`
	SkipReason           string    `json:"skip_reason,omitempty"`
	ResultStatus         string    `json:"result_status"`
	HTTPStatus           *int      `json:"http_status,omitempty"`
	FinalURL             string    `json:"final_url,omitempty"`
	ScreenshotEvidenceID string    `json:"screenshot_evidence_id,omitempty"`
	ConsoleErrorCount    int       `json:"console_error_count"`
	FailedRequestCount   int       `json:"failed_request_count"`
	DurationMS           *int      `json:"duration_ms,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

type SafeExplorerAction struct {
	ID           string    `json:"id"`
	RunID        string    `json:"run_id"`
	StepID       string    `json:"step_id"`
	SourceURL    string    `json:"source_url"`
	ActionType   string    `json:"action_type"`
	Label        string    `json:"label,omitempty"`
	Text         string    `json:"text,omitempty"`
	SelectorHint string    `json:"selector_hint,omitempty"`
	Href         string    `json:"href,omitempty"`
	TargetURL    string    `json:"target_url,omitempty"`
	Method       string    `json:"method,omitempty"`
	SameOrigin   bool      `json:"same_origin"`
	Safety       string    `json:"safety"`
	Decision     string    `json:"decision"`
	SkipReason   string    `json:"skip_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type SafeExplorerSummary struct {
	TotalSteps           int `json:"total_steps"`
	TotalPagesObserved   int `json:"total_pages_observed"`
	TotalActionsDetected int `json:"total_actions_detected"`
	TotalActionsExecuted int `json:"total_actions_executed"`
	TotalActionsSkipped  int `json:"total_actions_skipped"`
	TotalFindings        int `json:"total_findings"`
	SafeActions          int `json:"safe_actions"`
	UnsafeActionsSkipped int `json:"unsafe_actions_skipped"`
	ExternalActions      int `json:"external_actions_skipped"`
	UnsupportedActions   int `json:"unsupported_actions"`
	PagesWithScreenshots int `json:"pages_with_screenshots"`
}

type SafeExplorerTrace struct {
	Run      SafeExplorerRun      `json:"run"`
	Project  Project              `json:"project"`
	Summary  SafeExplorerSummary  `json:"summary"`
	Steps    []SafeExplorerStep   `json:"steps"`
	Actions  []SafeExplorerAction `json:"actions"`
	Findings []Finding            `json:"findings"`
	Evidence []Evidence           `json:"evidence"`
}

type SafeExplorerReport struct {
	GeneratedAt time.Time            `json:"generated_at"`
	Run         SafeExplorerRun      `json:"run"`
	Project     Project              `json:"project"`
	Settings    map[string]any       `json:"settings"`
	Summary     SafeExplorerSummary  `json:"summary"`
	Steps       []SafeExplorerStep   `json:"steps"`
	Actions     []SafeExplorerAction `json:"actions"`
	Findings    []Finding            `json:"findings"`
	Evidence    []Evidence           `json:"evidence"`
	SafetyNotes []string             `json:"safety_notes"`
	Limitations []string             `json:"limitations"`
	Metadata    map[string]any       `json:"metadata"`
	ReportIntelligence
}

type QualityCheckRunRequest struct {
	TargetURL            string `json:"target_url,omitempty"`
	CredentialProfileID  string `json:"credential_profile_id,omitempty"`
	DiscoveryRunID       string `json:"discovery_run_id,omitempty"`
	UseLatestDiscovery   bool   `json:"use_latest_discovery,omitempty"`
	MaxPages             int    `json:"max_pages,omitempty"`
	IncludeSecurity      *bool  `json:"include_security,omitempty"`
	IncludeAccessibility *bool  `json:"include_accessibility,omitempty"`
	IncludePerformance   *bool  `json:"include_performance,omitempty"`
}

type QualityCheckRun struct {
	ID                   string         `json:"id"`
	ProjectID            string         `json:"project_id"`
	DiscoveryRunID       string         `json:"discovery_run_id,omitempty"`
	CredentialProfileID  string         `json:"credential_profile_id,omitempty"`
	Status               string         `json:"status"`
	TargetURL            string         `json:"target_url"`
	MaxPages             int            `json:"max_pages"`
	IncludeSecurity      bool           `json:"include_security"`
	IncludeAccessibility bool           `json:"include_accessibility"`
	IncludePerformance   bool           `json:"include_performance"`
	StartedAt            *time.Time     `json:"started_at,omitempty"`
	CompletedAt          *time.Time     `json:"completed_at,omitempty"`
	TotalPages           int            `json:"total_pages"`
	TotalFindings        int            `json:"total_findings"`
	CriticalFindings     int            `json:"critical_findings"`
	HighFindings         int            `json:"high_findings"`
	MediumFindings       int            `json:"medium_findings"`
	LowFindings          int            `json:"low_findings"`
	InfoFindings         int            `json:"info_findings"`
	ErrorMessage         string         `json:"error_message,omitempty"`
	Summary              map[string]any `json:"summary"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type QualityCheckResult struct {
	ID             string         `json:"id"`
	RunID          string         `json:"run_id"`
	ProjectID      string         `json:"project_id"`
	Category       string         `json:"category"`
	RuleID         string         `json:"rule_id"`
	Severity       string         `json:"severity"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Recommendation string         `json:"recommendation"`
	URL            string         `json:"url"`
	Evidence       map[string]any `json:"evidence"`
	CreatedAt      time.Time      `json:"created_at"`
}

type QualityCheckSummary struct {
	TotalFindings         int `json:"total_findings"`
	Critical              int `json:"critical"`
	High                  int `json:"high"`
	Medium                int `json:"medium"`
	Low                   int `json:"low"`
	Info                  int `json:"info"`
	TotalPages            int `json:"total_pages"`
	SecurityFindings      int `json:"security_findings"`
	AccessibilityFindings int `json:"accessibility_findings"`
	PerformanceFindings   int `json:"performance_findings"`
}

type QualityCheckReport struct {
	GeneratedAt  time.Time            `json:"generated_at"`
	Run          QualityCheckRun      `json:"run"`
	Project      Project              `json:"project"`
	DiscoveryRun *DiscoveryRun        `json:"discovery_run,omitempty"`
	Summary      QualityCheckSummary  `json:"summary"`
	Results      []QualityCheckResult `json:"results"`
	Findings     []Finding            `json:"findings"`
	SafetyNotes  []string             `json:"safety_notes"`
	Limitations  []string             `json:"limitations"`
	Metadata     map[string]any       `json:"metadata"`
	ReportIntelligence
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
	ReportIntelligence
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
	ProviderID            string   `json:"provider_id"`
	RunID                 string   `json:"run_id"`
	DiscoveryRunID        string   `json:"discovery_run_id,omitempty"`
	UseLatestDiscovery    bool     `json:"use_latest_discovery,omitempty"`
	IncludeDiscoveryMap   *bool    `json:"include_discovery_map,omitempty"`
	ExecutionMode         string   `json:"execution_mode,omitempty"`
	MaxPagesFromDiscovery int      `json:"max_pages_from_discovery,omitempty"`
	ProductContext        string   `json:"product_context"`
	FocusAreas            []string `json:"focus_areas"`
	MaxScenarios          int      `json:"max_scenarios"`
}

type TestPlan struct {
	ID                string                     `json:"id"`
	ProjectID         string                     `json:"project_id"`
	RunID             string                     `json:"run_id,omitempty"`
	DiscoveryRunID    string                     `json:"discovery_run_id,omitempty"`
	SourceType        string                     `json:"source_type,omitempty"`
	ProviderID        string                     `json:"provider_id,omitempty"`
	ProviderName      string                     `json:"provider_name,omitempty"`
	Model             string                     `json:"model,omitempty"`
	Status            string                     `json:"status"`
	Title             string                     `json:"title"`
	Summary           string                     `json:"summary"`
	PlanJSON          map[string]any             `json:"plan_json"`
	RiskLevel         string                     `json:"risk_level"`
	TotalScenarios    int                        `json:"total_scenarios"`
	ExecutionCoverage TestPlanExecutableCoverage `json:"execution_coverage"`
	ErrorMessage      string                     `json:"error_message,omitempty"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

type TestPlanRef struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	RiskLevel      string    `json:"risk_level"`
	TotalScenarios int       `json:"total_scenarios"`
	CreatedAt      time.Time `json:"created_at"`
}

type TestPlanExecutableCoverage struct {
	TotalScenarios          int `json:"total_scenarios"`
	ExecutableScenarios     int `json:"executable_scenarios"`
	SkippedScenarios        int `json:"skipped_scenarios"`
	TotalSteps              int `json:"total_steps"`
	ExecutableSteps         int `json:"executable_steps"`
	SkippedSteps            int `json:"skipped_steps"`
	UnsafeSkippedSteps      int `json:"unsafe_skipped_steps"`
	UnsupportedSkippedSteps int `json:"unsupported_skipped_steps"`
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
	ReportIntelligence
}

type QARunRequest struct {
	Mode                        string   `json:"mode"`
	StartURL                    string   `json:"start_url,omitempty"`
	CredentialProfileID         string   `json:"credential_profile_id,omitempty"`
	MaxPages                    int      `json:"max_pages,omitempty"`
	MaxDepth                    int      `json:"max_depth,omitempty"`
	MaxScenarios                int      `json:"max_scenarios,omitempty"`
	Execute                     bool     `json:"execute,omitempty"`
	UseExistingDiscoveryRunID   string   `json:"use_existing_discovery_run_id,omitempty"`
	UseLatestDiscovery          bool     `json:"use_latest_discovery,omitempty"`
	ProviderID                  string   `json:"provider_id,omitempty"`
	ProductContext              string   `json:"product_context,omitempty"`
	FocusAreas                  []string `json:"focus_areas,omitempty"`
	IncludeQualityChecks        *bool    `json:"include_quality_checks,omitempty"`
	QualityMaxPages             int      `json:"quality_max_pages,omitempty"`
	QualityIncludeSecurity      *bool    `json:"quality_include_security,omitempty"`
	QualityIncludeAccessibility *bool    `json:"quality_include_accessibility,omitempty"`
	QualityIncludePerformance   *bool    `json:"quality_include_performance,omitempty"`
}

type QARun struct {
	ID                  string         `json:"id"`
	ProjectID           string         `json:"project_id"`
	Status              string         `json:"status"`
	Mode                string         `json:"mode"`
	DiscoveryRunID      string         `json:"discovery_run_id,omitempty"`
	QualityCheckRunID   string         `json:"quality_check_run_id,omitempty"`
	TestPlanID          string         `json:"test_plan_id,omitempty"`
	TestPlanExecutionID string         `json:"test_plan_execution_id,omitempty"`
	CredentialProfileID string         `json:"credential_profile_id,omitempty"`
	ErrorMessage        string         `json:"error_message,omitempty"`
	Summary             map[string]any `json:"summary"`
	StartedAt           *time.Time     `json:"started_at,omitempty"`
	CompletedAt         *time.Time     `json:"completed_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type QARunReport struct {
	Run              QARun                     `json:"run"`
	Project          Project                   `json:"project"`
	DiscoveryRun     *DiscoveryRun             `json:"discovery_run,omitempty"`
	DiscoverySummary *DiscoverySummary         `json:"discovery_summary,omitempty"`
	QualityCheckRun  *QualityCheckRun          `json:"quality_check_run,omitempty"`
	QualitySummary   *QualityCheckSummary      `json:"quality_summary,omitempty"`
	QualityResults   []QualityCheckResult      `json:"quality_results,omitempty"`
	TestPlan         *TestPlan                 `json:"test_plan,omitempty"`
	ExecutionPreview *TestPlanExecutionPreview `json:"execution_preview,omitempty"`
	ExecutionReport  *TestPlanExecutionReport  `json:"execution_report,omitempty"`
	Findings         []Finding                 `json:"findings"`
	Evidence         []Evidence                `json:"evidence"`
	SafetyNotes      []string                  `json:"safety_notes"`
	Limitations      []string                  `json:"limitations"`
	Baseline         *ReportBaseline           `json:"baseline,omitempty"`
	Comparison       *ReportComparison         `json:"comparison,omitempty"`
	QualityGate      *QualityGateResult        `json:"quality_gate,omitempty"`
	BaselineMessage  string                    `json:"baseline_message,omitempty"`
	GeneratedAt      time.Time                 `json:"generated_at"`
	ReportIntelligence
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

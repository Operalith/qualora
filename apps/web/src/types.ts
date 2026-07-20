export type Project = {
  id: string;
  name: string;
  frontend_url: string;
  api_base_url: string;
  openapi_url: string;
  allowed_hosts: string[];
  security_mode: string;
  destructive_actions: boolean;
  allow_private_targets: boolean;
  created_at: string;
  updated_at: string;
};

export type SetupStatus = {
  setup_required: boolean;
  version: string;
};

export type AuthUser = {
  id: string;
  email: string;
  display_name: string;
  role: "admin";
  last_login_at?: string;
  created_at: string;
  updated_at: string;
};

export type SetupAdminInput = {
  display_name: string;
  email: string;
  password: string;
  confirm_password: string;
};

export type LoginInput = {
  email: string;
  password: string;
};

export type AuthResponse = {
  user: AuthUser;
  expires_at: string;
  csrf_token?: string;
  setup_status?: string;
};

export type MeResponse = {
  authenticated: boolean;
  user?: AuthUser;
  expires_at?: string;
};

export type CreateProjectInput = {
  name: string;
  frontend_url: string;
  api_base_url: string;
  openapi_url: string;
  allowed_hosts: string[];
  security_mode: "passive";
  destructive_actions: false;
  allow_private_targets: boolean;
};

export type ProjectSetupInput = {
  project: CreateProjectInput;
  ai?: {
    mode?: "skip" | "existing" | "create" | "demo";
    provider_id?: string;
    provider?: AIProviderInput;
  };
  credential?: {
    mode?: "skip" | "create";
    profile?: CredentialProfileInput;
  };
  api_spec?: {
    mode?: "skip" | "import" | "demo";
    spec?: APISpecImportInput;
  };
  workflow?: {
    browser_smoke?: boolean;
    discovery?: boolean;
    quality_checks?: boolean;
    safe_qa_run?: boolean;
    execute_safe_qa?: boolean;
    api_smoke?: boolean;
    authenticated_smoke?: boolean;
    use_defaults?: boolean;
  };
};

export type ProjectSetupResponse = {
  project: Project;
  ai_provider?: AIProvider;
  credential_profile?: CredentialProfile;
  api_spec?: APISpec;
  started: {
    browser_smoke_run_id?: string;
    authenticated_smoke_run_id?: string;
    discovery_run_id?: string;
    quality_check_run_id?: string;
    safe_qa_run_id?: string;
    api_smoke_run_id?: string;
    ai_provider_id?: string;
    credential_profile_id?: string;
    api_spec_id?: string;
  };
  skipped: Array<{ action: string; reason: string }>;
  timeline: Array<{ step: string; status: string; resource?: string; reason?: string }>;
  next_links: Record<string, string>;
};

export type TestRun = {
  id: string;
  project_id: string;
  run_type: "full" | "browser_smoke" | "api_smoke" | "login_check" | "authenticated_browser_smoke" | string;
  api_spec_id?: string;
  credential_profile_id?: string;
  api_auth_profile_id?: string;
  target_path?: string;
  capture_screenshot?: boolean;
  max_duration_seconds?: number;
  status: "queued" | "pending" | "running" | "completed" | "failed" | "canceled" | "passed" | "error";
  error_message?: string;
  page_title?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

export type RunJob = {
  id: string;
  run_id: string;
  kind: string;
  status: string;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

export type Finding = {
  id: string;
  run_id?: string;
  test_plan_execution_id?: string;
  authorization_check_run_id?: string;
  discovery_run_id?: string;
  safe_explorer_run_id?: string;
  ai_browser_control_run_id?: string;
  scenario_execution_id?: string;
  step_execution_id?: string;
  title: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  category: string;
  confidence: string;
  description: string;
  recommendation: string;
  evidence_ids: string[];
  created_at?: string;
};

export type Evidence = {
  id: string;
  run_id?: string;
  test_plan_execution_id?: string;
  authorization_check_run_id?: string;
  discovery_run_id?: string;
  safe_explorer_run_id?: string;
  ai_browser_control_run_id?: string;
  type: string;
  uri: string;
  metadata: Record<string, unknown>;
  created_at?: string;
};

export type AIProvider = {
  id: string;
  name: string;
  preset: string;
  type: "openai-compatible";
  base_url: string;
  model: string;
  temperature: number;
  max_output_tokens: number;
  timeout_seconds: number;
  send_screenshots: boolean;
  send_html: boolean;
  send_network_bodies: boolean;
  redaction_enabled: boolean;
  is_default: boolean;
  api_key_configured: boolean;
  extra_headers_configured: boolean;
  created_at: string;
  updated_at: string;
};

export type AIProviderInput = {
  name: string;
  preset: string;
  type: "openai-compatible";
  base_url: string;
  model: string;
  api_key?: string;
  extra_headers?: Record<string, string>;
  temperature: number;
  max_output_tokens: number;
  timeout_seconds: number;
  send_screenshots: boolean;
  send_html: boolean;
  send_network_bodies: boolean;
  redaction_enabled: boolean;
  is_default: boolean;
};

export type AIProviderTestResult = {
  success: boolean;
  provider_name: string;
  model: string;
  latency_ms: number;
  error_message?: string;
};

export type CredentialProfile = {
  id: string;
  project_id: string;
  name: string;
  type: "username_password";
  role_name?: string;
  role_description?: string;
  subject_label?: string;
  username_configured: boolean;
  password_configured: boolean;
  username_display_hint?: string;
  login_url: string;
  username_selector: string;
  password_selector: string;
  submit_selector: string;
  success_url_contains?: string;
  success_text_contains?: string;
  failure_text_contains?: string;
  post_login_wait_ms: number;
  is_default: boolean;
  created_at: string;
  updated_at: string;
};

export type CredentialProfileInput = {
  name: string;
  type: "username_password";
  role_name?: string;
  role_description?: string;
  subject_label?: string;
  username?: string;
  password?: string;
  login_url: string;
  username_selector: string;
  password_selector: string;
  submit_selector: string;
  success_url_contains: string;
  success_text_contains: string;
  failure_text_contains: string;
  post_login_wait_ms: number;
  is_default: boolean;
};

export type AuthorizationCheck = {
  id: string;
  project_id: string;
  name: string;
  description?: string;
  type: "browser_url" | "api_get";
  resource_label?: string;
  owner_credential_profile_id?: string;
  actor_credential_profile_id: string;
  expected_outcome: "allowed" | "denied";
  target_url?: string;
  api_spec_id?: string;
  api_operation_id?: string;
  method?: string;
  path?: string;
  expected_statuses?: number[];
  success_text_contains?: string;
  denied_statuses?: number[];
  denied_text_contains?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

export type AuthorizationCheckInput = {
  name: string;
  description?: string;
  type: "browser_url";
  resource_label?: string;
  owner_credential_profile_id?: string;
  actor_credential_profile_id: string;
  expected_outcome: "allowed" | "denied";
  target_url: string;
  success_text_contains?: string;
  denied_text_contains?: string;
  enabled?: boolean;
};

export type AuthorizationCheckRun = {
  id: string;
  project_id: string;
  status: "queued" | "running" | "completed" | "failed" | "error" | string;
  check_ids?: string[];
  max_checks: number;
  total_checks: number;
  passed_checks: number;
  failed_checks: number;
  skipped_checks: number;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

export type AuthorizationCheckRunInput = {
  check_ids?: string[];
  max_checks?: number;
};

export type AuthorizationCheckResult = {
  id: string;
  run_id: string;
  check_id: string;
  status: "passed" | "failed" | "skipped" | "error" | string;
  expected_outcome: "allowed" | "denied";
  actual_outcome: "allowed" | "denied" | "unknown";
  actor_credential_profile_id: string;
  actor_role_name?: string;
  target_url?: string;
  final_url?: string;
  http_status?: number;
  page_title?: string;
  duration_ms?: number;
  evidence_id?: string;
  finding_id?: string;
  skip_reason?: string;
  error_message?: string;
  created_at: string;
};

export type AuthorizationCheckDetail = {
  run: AuthorizationCheckRun;
  checks?: AuthorizationCheck[];
  results: AuthorizationCheckResult[];
};

export type AuthorizationCheckReport = ReportIntelligenceFields & {
  run: AuthorizationCheckRun;
  project: Project;
  checks: AuthorizationCheck[];
  results: AuthorizationCheckResult[];
  summary: ReportSummary;
  findings: Finding[];
  evidence: Evidence[];
  metadata: Record<string, unknown>;
  generated_at: string;
};

export type DiscoveryRunInput = {
  start_url?: string;
  credential_profile_id?: string;
  max_pages?: number;
  max_depth?: number;
  same_origin_only?: boolean;
};

export type DiscoveryRun = {
  id: string;
  project_id: string;
  credential_profile_id?: string;
  status: "queued" | "running" | "completed" | "failed" | "error" | string;
  start_url: string;
  max_pages: number;
  max_depth: number;
  same_origin_only: boolean;
  started_at?: string;
  completed_at?: string;
  total_pages: number;
  total_links: number;
  total_forms: number;
  total_console_errors: number;
  total_failed_requests: number;
  total_findings: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type DiscoveredPage = {
  id: string;
  discovery_run_id: string;
  project_id: string;
  url: string;
  normalized_url: string;
  path: string;
  title?: string;
  http_status?: number;
  content_type?: string;
  body_text_length?: number;
  load_duration_ms?: number;
  depth: number;
  screenshot_evidence_id?: string;
  console_error_count: number;
  failed_request_count: number;
  discovered_at: string;
  created_at: string;
};

export type DiscoveredLink = {
  id: string;
  discovery_run_id: string;
  source_page_id: string;
  href: string;
  normalized_url?: string;
  link_text?: string;
  same_origin: boolean;
  skipped: boolean;
  skip_reason?: string;
  created_at: string;
};

export type DiscoveredFormField = {
  id: string;
  form_id: string;
  field_name?: string;
  field_type?: string;
  placeholder?: string;
  label?: string;
  required: boolean;
  created_at: string;
};

export type DiscoveredForm = {
  id: string;
  discovery_run_id: string;
  page_id: string;
  form_name?: string;
  form_action?: string;
  form_method?: string;
  field_count: number;
  password_field_count: number;
  submit_button_count: number;
  classification?: string;
  skipped_reason?: string;
  fields?: DiscoveredFormField[];
  created_at: string;
};

export type DiscoverySummary = {
  total_pages: number;
  total_links: number;
  total_forms: number;
  total_console_errors: number;
  total_failed_requests: number;
  total_findings: number;
  skipped_links: number;
  external_links_skipped: number;
  unsafe_links_skipped: number;
  pages_with_screenshots: number;
};

export type DiscoveryMap = {
  run: DiscoveryRun;
  project: Project;
  summary: DiscoverySummary;
  pages: DiscoveredPage[];
  links: DiscoveredLink[];
  forms: DiscoveredForm[];
  findings: Finding[];
  evidence: Evidence[];
};

export type DiscoveryReport = DiscoveryMap & ReportIntelligenceFields & {
  generated_at: string;
  settings: Record<string, unknown>;
  safety_notes: string[];
  limitations: string[];
  metadata: Record<string, unknown>;
};

export type SafeExplorerRunInput = {
  start_url?: string;
  credential_profile_id?: string;
  max_steps?: number;
  max_depth?: number;
  same_origin_only?: boolean;
  allow_get_forms?: boolean;
};

export type SafeExplorerRun = {
  id: string;
  project_id: string;
  credential_profile_id?: string;
  status: "queued" | "running" | "completed" | "failed" | "error" | string;
  start_url: string;
  max_steps: number;
  max_depth: number;
  same_origin_only: boolean;
  allow_get_forms: boolean;
  started_at?: string;
  completed_at?: string;
  total_steps: number;
  total_pages_observed: number;
  total_actions_detected: number;
  total_actions_executed: number;
  total_actions_skipped: number;
  total_findings: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type SafeExplorerStep = {
  id: string;
  run_id: string;
  project_id: string;
  step_index: number;
  page_url: string;
  normalized_url: string;
  page_title?: string;
  depth: number;
  action_id?: string;
  action_type?: string;
  action_label?: string;
  action_selector_hint?: string;
  action_target_url?: string;
  action_safety: "safe" | "unsafe" | "unsupported" | "unknown" | string;
  action_decision: "executed" | "skipped" | "observed" | string;
  skip_reason?: string;
  result_status: "ok" | "failed" | "skipped" | "error" | string;
  http_status?: number;
  final_url?: string;
  screenshot_evidence_id?: string;
  console_error_count: number;
  failed_request_count: number;
  duration_ms?: number;
  created_at: string;
};

export type SafeExplorerAction = {
  id: string;
  run_id: string;
  step_id: string;
  source_url: string;
  action_type: "link_navigation" | "button" | "form_get" | "form_post" | "input" | "unknown" | string;
  label?: string;
  text?: string;
  selector_hint?: string;
  href?: string;
  target_url?: string;
  method?: string;
  same_origin: boolean;
  safety: "safe" | "unsafe" | "unsupported" | "unknown" | string;
  decision: "execute" | "skip" | string;
  skip_reason?: string;
  created_at: string;
};

export type SafeExplorerSummary = {
  total_steps: number;
  total_pages_observed: number;
  total_actions_detected: number;
  total_actions_executed: number;
  total_actions_skipped: number;
  total_findings: number;
  safe_actions: number;
  unsafe_actions_skipped: number;
  external_actions_skipped: number;
  unsupported_actions: number;
  pages_with_screenshots: number;
};

export type SafeExplorerTrace = {
  run: SafeExplorerRun;
  project: Project;
  summary: SafeExplorerSummary;
  steps: SafeExplorerStep[];
  actions: SafeExplorerAction[];
  findings: Finding[];
  evidence: Evidence[];
};

export type SafeExplorerReport = SafeExplorerTrace & ReportIntelligenceFields & {
  generated_at: string;
  settings: Record<string, unknown>;
  safety_notes: string[];
  limitations: string[];
  metadata: Record<string, unknown>;
};

export type AIBrowserControlRunInput = {
  provider_id: string;
  goal: string;
  start_url?: string;
  credential_profile_id?: string;
  max_steps?: number;
  max_depth?: number;
  same_origin_only?: boolean;
};

export type AIBrowserControlRun = {
  id: string;
  project_id: string;
  provider_id: string;
  credential_profile_id?: string;
  status: "queued" | "running" | "completed" | "failed" | "error" | string;
  goal: string;
  start_url: string;
  max_steps: number;
  max_depth: number;
  same_origin_only: boolean;
  policy_version: string;
  execution_mode: string;
  started_at?: string;
  completed_at?: string;
  total_steps: number;
  total_ai_suggestions: number;
  total_actions_approved: number;
  total_actions_executed: number;
  total_actions_skipped: number;
  total_policy_blocks: number;
  total_findings: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type AIBrowserControlStep = {
  id: string;
  run_id: string;
  project_id: string;
  step_index: number;
  page_url: string;
  normalized_url: string;
  page_title?: string;
  depth: number;
  sanitized_observation: Record<string, unknown>;
  ai_suggestion: Record<string, unknown>;
  policy_decision: "approved" | "blocked" | "invalid" | "unsupported" | string;
  policy_reason?: string;
  action_type?: string;
  action_label?: string;
  action_target_url?: string;
  selector_hint?: string;
  execution_status: "executed" | "skipped" | "failed" | "error" | string;
  final_url?: string;
  http_status?: number;
  screenshot_evidence_id?: string;
  console_error_count: number;
  failed_request_count: number;
  duration_ms?: number;
  created_at: string;
};

export type AIBrowserControlSummary = {
  total_steps: number;
  total_ai_suggestions: number;
  actions_approved: number;
  actions_executed: number;
  actions_skipped: number;
  policy_blocks: number;
  findings: number;
  screenshots: number;
  console_errors: number;
  failed_requests: number;
};

export type AIBrowserControlTrace = {
  run: AIBrowserControlRun;
  project: Project;
  summary: AIBrowserControlSummary;
  steps: AIBrowserControlStep[];
  findings: Finding[];
  evidence: Evidence[];
};

export type AIBrowserControlReport = AIBrowserControlTrace & ReportIntelligenceFields & {
  generated_at: string;
  settings: Record<string, unknown>;
  safety_notes: string[];
  limitations: string[];
  metadata: Record<string, unknown>;
};

export type QualityCheckRunInput = {
  target_url?: string;
  credential_profile_id?: string;
  discovery_run_id?: string;
  use_latest_discovery?: boolean;
  max_pages?: number;
  include_security?: boolean;
  include_accessibility?: boolean;
  include_performance?: boolean;
};

export type QualityCheckRun = {
  id: string;
  project_id: string;
  discovery_run_id?: string;
  credential_profile_id?: string;
  status: "queued" | "running" | "completed" | "failed" | "error" | string;
  target_url: string;
  max_pages: number;
  include_security: boolean;
  include_accessibility: boolean;
  include_performance: boolean;
  started_at?: string;
  completed_at?: string;
  total_pages: number;
  total_findings: number;
  critical_findings: number;
  high_findings: number;
  medium_findings: number;
  low_findings: number;
  info_findings: number;
  error_message?: string;
  summary: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type QualityCheckResult = {
  id: string;
  run_id: string;
  project_id: string;
  category: "security" | "accessibility" | "performance" | string;
  rule_id: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  title: string;
  description: string;
  recommendation: string;
  url: string;
  evidence: Record<string, unknown>;
  created_at: string;
};

export type QualityCheckSummary = {
  total_findings: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  info: number;
  total_pages: number;
  security_findings: number;
  accessibility_findings: number;
  performance_findings: number;
};

export type QualityCheckReport = ReportIntelligenceFields & {
  generated_at: string;
  run: QualityCheckRun;
  project: Project;
  discovery_run?: DiscoveryRun;
  summary: QualityCheckSummary;
  results: QualityCheckResult[];
  findings: Finding[];
  safety_notes: string[];
  limitations: string[];
  metadata: Record<string, unknown>;
};

export type AuthenticatedBrowserSmokeInput = {
  credential_profile_id?: string;
  target_path?: string;
  capture_screenshot?: boolean;
  max_duration_seconds?: number;
};

export type AIAnalysis = {
  id: string;
  run_id: string;
  provider_id?: string;
  provider_name?: string;
  model: string;
  status: "queued" | "running" | "completed" | "failed";
  executive_summary: string;
  technical_summary: string;
  risk_level: "low" | "medium" | "high" | "critical" | "";
  analysis_json: Record<string, unknown>;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type TestPlanStep = {
  order: number;
  action: string;
  target: string;
  data: string;
  expected_result: string;
};

export type TestPlanScenario = {
  id: string;
  name: string;
  type: string;
  priority: "low" | "medium" | "high" | "critical";
  risk: "low" | "medium" | "high" | "critical";
  description: string;
  preconditions: string[];
  steps: TestPlanStep[];
  assertions: string[];
  test_data_needed: string[];
  automation_candidate: boolean;
  destructive: boolean;
  requires_authentication: boolean;
  related_findings: string[];
  tags: string[];
};

export type TestPlanPayload = {
  title: string;
  summary: string;
  assumptions: string[];
  coverage_goals: string[];
  scenarios: TestPlanScenario[];
  suggested_next_instrumentation: string[];
  limitations: string[];
};

export type TestPlan = {
  id: string;
  project_id: string;
  run_id?: string;
  discovery_run_id?: string;
  source_type?: string;
  provider_id?: string;
  provider_name?: string;
  model: string;
  status: "queued" | "running" | "completed" | "failed";
  title: string;
  summary: string;
  plan_json: TestPlanPayload;
  risk_level: "low" | "medium" | "high" | "critical" | "";
  total_scenarios: number;
  execution_coverage: TestPlanExecutableCoverage;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type TestPlanRef = {
  id: string;
  title: string;
  status: string;
  risk_level: string;
  total_scenarios: number;
  created_at: string;
};

export type AITestPlanInput = {
  provider_id?: string;
  run_id?: string;
  discovery_run_id?: string;
  use_latest_discovery?: boolean;
  include_discovery_map?: boolean;
  execution_mode?: "review_only" | "safe_executable";
  max_pages_from_discovery?: number;
  product_context?: string;
  focus_areas: string[];
  max_scenarios: number;
};

export type TestPlanExecutableCoverage = {
  total_scenarios: number;
  executable_scenarios: number;
  skipped_scenarios: number;
  total_steps: number;
  executable_steps: number;
  skipped_steps: number;
  unsafe_skipped_steps: number;
  unsupported_skipped_steps: number;
};

export type TestPlanExecutionRequest = {
  max_scenarios: number;
  max_steps_per_scenario: number;
  scenario_ids?: string[];
  dry_run: boolean;
};

export type MappedExecutionStep = {
  step_order: number;
  original_action: string;
  mapped_action: string;
  target: string;
  expected_result: string;
  status: string;
  skip_reason?: string;
};

export type MappedExecutionScenario = {
  scenario_id_from_plan: string;
  name: string;
  type: string;
  priority: string;
  status: string;
  skip_reason?: string;
  steps: MappedExecutionStep[];
};

export type TestPlanExecutionPreview = {
  dry_run: boolean;
  test_plan_id: string;
  project_id: string;
  max_scenarios: number;
  max_steps_per_scenario: number;
  total_scenarios: number;
  executable_scenarios: number;
  skipped_scenarios: number;
  total_steps: number;
  executable_steps: number;
  skipped_steps: number;
  scenarios: MappedExecutionScenario[];
  safety_summary: TestPlanExecutionSafetyReport;
};

export type TestPlanExecution = {
  id: string;
  test_plan_id: string;
  project_id: string;
  source_run_id?: string;
  status: string;
  total_scenarios: number;
  passed_scenarios: number;
  failed_scenarios: number;
  skipped_scenarios: number;
  total_steps: number;
  passed_steps: number;
  failed_steps: number;
  skipped_steps: number;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

export type TestPlanExecutionScenario = {
  id: string;
  execution_id: string;
  scenario_id_from_plan: string;
  name: string;
  type: string;
  priority: string;
  status: string;
  skip_reason?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  steps?: TestPlanExecutionStep[];
};

export type TestPlanExecutionStep = {
  id: string;
  execution_id: string;
  scenario_execution_id: string;
  step_order: number;
  original_action: string;
  mapped_action: string;
  target: string;
  expected_result: string;
  status: string;
  skip_reason?: string;
  actual_result?: string;
  error_message?: string;
  duration_ms?: number;
  evidence_id?: string;
  created_at: string;
  updated_at: string;
};

export type TestPlanExecutionDetail = {
  execution: TestPlanExecution;
  scenarios: TestPlanExecutionScenario[];
};

export type TestPlanExecutionSafetyReport = {
  executed_steps: number;
  skipped_unsafe_steps: number;
  skipped_unsupported_steps: number;
  skipped_scenarios: number;
};

export type TestPlanExecutionReport = ReportIntelligenceFields & {
  execution: TestPlanExecution;
  test_plan: TestPlan;
  project: Project;
  scenarios: TestPlanExecutionScenario[];
  findings: Finding[];
  evidence: Evidence[];
  safety_summary: TestPlanExecutionSafetyReport;
  generated_at: string;
};

export type QARunInput = {
  mode?: "safe";
  start_url?: string;
  credential_profile_id?: string;
  api_auth_profile_id?: string;
  max_pages?: number;
  max_depth?: number;
  max_scenarios?: number;
  execute?: boolean;
  use_existing_discovery_run_id?: string;
  use_latest_discovery?: boolean;
  provider_id?: string;
  product_context?: string;
  focus_areas?: string[];
  include_quality_checks?: boolean;
  quality_max_pages?: number;
  quality_include_security?: boolean;
  quality_include_accessibility?: boolean;
  quality_include_performance?: boolean;
  include_api_checks?: boolean;
  api_validate_contract?: boolean;
  api_validate_schema?: boolean;
  api_include_unauthenticated_comparison?: boolean;
};

export type QARun = {
  id: string;
  project_id: string;
  status: "queued" | "running_discovery" | "generating_plan" | "previewing_execution" | "executing_plan" | "completed" | "failed" | "error" | string;
  mode: "safe" | string;
  discovery_run_id?: string;
  quality_check_run_id?: string;
  api_smoke_run_id?: string;
  test_plan_id?: string;
  test_plan_execution_id?: string;
  credential_profile_id?: string;
  api_auth_profile_id?: string;
  error_message?: string;
  summary: Record<string, unknown>;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

export type QARunReport = ReportIntelligenceFields & {
  run: QARun;
  project: Project;
  discovery_run?: DiscoveryRun;
  discovery_summary?: DiscoverySummary;
  quality_check_run?: QualityCheckRun;
  quality_summary?: QualityCheckSummary;
  quality_results?: QualityCheckResult[];
  api_smoke_run?: TestRun;
  api_spec?: APISpec;
  api_auth?: APIAuthSummary;
  api_summary?: APISmokeSummary;
  api_results?: APICheckResult[];
  test_plan?: TestPlan;
  execution_preview?: TestPlanExecutionPreview;
  execution_report?: TestPlanExecutionReport;
  findings: Finding[];
  evidence: Evidence[];
  safety_notes: string[];
  limitations: string[];
  baseline?: ReportBaseline;
  comparison?: ReportComparison;
  quality_gate?: QualityGateResult;
  baseline_message?: string;
  generated_at: string;
};

export type APISpecImportInput = {
  name: string;
  source_type: "url" | "inline" | "demo";
  source_url?: string;
  raw_spec?: string;
};

export type APISpec = {
  id: string;
  project_id: string;
  name: string;
  source_type: "url" | "inline" | "demo" | string;
  source_url?: string;
  parsed_title?: string;
  parsed_version?: string;
  server_url?: string;
  operation_count: number;
  safe_operation_count: number;
  skipped_operation_count: number;
  status: "pending" | "parsed" | "invalid" | "error" | string;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type APIOperation = {
  id: string;
  api_spec_id: string;
  project_id: string;
  method: string;
  path: string;
  resolved_path?: string;
  query_string?: string;
  operation_id?: string;
  summary?: string;
  description?: string;
  tags: string[];
  expected_statuses: string[];
  expected_content_types: string[];
  response_schemas?: Record<string, unknown>;
  requires_authentication?: boolean;
  safe_to_execute: boolean;
  skip_reason?: string;
  created_at: string;
  updated_at: string;
};

export type APISpecDetail = {
  spec: APISpec;
  operations?: APIOperation[];
};

export type APISmokeRunInput = {
  api_auth_profile_id?: string;
  authenticated?: boolean;
  validate_contract?: boolean;
  validate_schema?: boolean;
  max_operations?: number;
  include_unauthenticated_comparison?: boolean;
};

export type APICheckResult = {
  id: string;
  run_id: string;
  api_spec_id: string;
  operation_id?: string;
  api_auth_profile_id?: string;
  auth_mode?: string;
  method: string;
  path: string;
  resolved_url?: string;
  status: "passed" | "failed" | "skipped" | "error" | string;
  http_status?: number;
  actual_status?: number;
  duration_ms?: number;
  response_time_ms?: number;
  response_content_type?: string;
  actual_content_type?: string;
  response_size_bytes?: number;
  expected_statuses?: string[];
  expected_content_types?: string[];
  contract_validation_status?: string;
  schema_validation_errors?: string[];
  unauthenticated_status?: number;
  error_message?: string;
  skipped_reason?: string;
  created_at: string;
};

export type APISmokeSummary = {
  total_operations: number;
  executed_operations: number;
  skipped_operations: number;
  passed_operations: number;
  failed_operations: number;
  errored_operations: number;
  authenticated_operations?: number;
  unauthenticated_comparisons?: number;
  contract_passed?: number;
  contract_failed?: number;
  contract_skipped?: number;
  contract_unknown?: number;
  schema_validation_error_count?: number;
};

export type APIAuthProfile = {
  id: string;
  project_id: string;
  name: string;
  type: "bearer_token" | "api_key_header" | "api_key_query" | "basic_auth" | "none" | string;
  header_name?: string;
  query_param_name?: string;
  username_display_hint?: string;
  token_display_hint?: string;
  api_key_display_hint?: string;
  username_configured: boolean;
  password_configured: boolean;
  token_configured: boolean;
  api_key_configured: boolean;
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

export type APIAuthProfileInput = {
  name: string;
  type: "bearer_token" | "api_key_header" | "api_key_query" | "basic_auth" | "none";
  header_name?: string;
  query_param_name?: string;
  username?: string;
  password?: string;
  token?: string;
  api_key?: string;
  enabled?: boolean;
};

export type APIAuthProfileTestInput = {
  test_path?: string;
  method?: "GET" | "HEAD";
};

export type APIAuthProfileTestResult = {
  success: boolean;
  profile_id: string;
  profile_name: string;
  auth_mode: string;
  method: string;
  url: string;
  http_status?: number;
  response_content_type?: string;
  duration_ms: number;
  redacted_headers?: Record<string, string>;
  error_message?: string;
};

export type APIAuthSummary = {
  profile_id?: string;
  profile_name?: string;
  type?: string;
  auth_mode: string;
  display_hint?: string;
  authenticated: boolean;
  secrets_stored: string;
  secrets_returned: boolean;
  secrets_sent_to_ai: boolean;
  auth_headers_stored: boolean;
};

export type ReportSummary = {
  total_findings: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  info: number;
};

export type ReportFindingOccurrenceRef = {
  source_type: string;
  source_run_id?: string;
  finding_id?: string;
  quality_result_id?: string;
  evidence_id?: string;
  category?: string;
  severity?: string;
  affected_url?: string;
  affected_path?: string;
};

export type GroupedFinding = {
  group_id: string;
  fingerprint: string;
  category: string;
  title: string;
  normalized_severity: string;
  summary?: string;
  recommendation?: string;
  occurrences_count: number;
  affected_urls?: string[];
  affected_paths?: string[];
  sources: string[];
  representative_evidence_id?: string;
  confidence?: string;
  noise_level: string;
  raw_occurrence_refs: ReportFindingOccurrenceRef[];
};

export type AffectedPageSummary = {
  url?: string;
  path?: string;
  findings_count: number;
  highest_severity: string;
};

export type NoiseSummary = {
  high_noise: number;
  medium_noise: number;
  low_noise: number;
  high_signal: number;
  needs_attention: number;
  informational: number;
  noisy_repeated: number;
};

export type DeduplicationSummary = {
  raw_findings_count: number;
  grouped_findings_count: number;
  duplicate_findings_reduced: number;
  grouped_repeated_findings: number;
  cross_source_grouped_findings: number;
};

export type ReportExecutiveSummary = {
  overall_status: string;
  headline: string;
  total_findings: number;
  grouped_findings: number;
  severity_counts: ReportSummary;
  checks_completed: string[];
  checks_skipped: string[];
  recommended_next_actions: string[];
  what_was_tested: string[];
  what_was_not_tested: string[];
  safety_limitations: string[];
};

export type ReportIntelligenceFields = {
  executive_summary?: ReportExecutiveSummary;
  severity_counts?: ReportSummary;
  grouped_findings?: GroupedFinding[];
  top_findings?: GroupedFinding[];
  top_affected_pages?: AffectedPageSummary[];
  noise_summary?: NoiseSummary;
  raw_findings_count?: number;
  deduplication_summary?: DeduplicationSummary;
  safety_limitations?: string[];
};

export type ReportBaseline = {
  id: string;
  project_id: string;
  name: string;
  description?: string;
  report_type: string;
  report_id: string;
  source_run_id?: string;
  fingerprint_set: GroupedFinding[];
  severity_counts: ReportSummary;
  grouped_findings_count: number;
  raw_findings_count: number;
  created_by_user_id?: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
};

export type ReportBaselineInput = {
  name: string;
  description?: string;
  report_type: string;
  report_id: string;
  is_default?: boolean;
};

export type SeverityChange = {
  fingerprint: string;
  title: string;
  previous_severity: string;
  current_severity: string;
};

export type AffectedScopeChange = {
  fingerprint: string;
  title: string;
  previous_affected_urls: number;
  current_affected_urls: number;
  previous_affected_paths: number;
  current_affected_paths: number;
};

export type ReportComparisonSummary = {
  new_findings_count: number;
  fixed_findings_count: number;
  unchanged_findings_count: number;
  severity_changes: SeverityChange[];
  new_critical: number;
  new_high: number;
  new_medium: number;
  fixed_critical: number;
  fixed_high: number;
  fixed_medium: number;
};

export type ReportComparison = {
  comparison_id?: string;
  project_id: string;
  report_type: string;
  baseline_id?: string;
  current_report_id: string;
  status: "improved" | "regressed" | "unchanged" | "mixed" | "unknown" | string;
  summary: ReportComparisonSummary;
  new_findings: GroupedFinding[];
  fixed_findings: GroupedFinding[];
  unchanged_findings: GroupedFinding[];
  severity_delta: ReportSummary;
  affected_pages_delta: AffectedScopeChange[];
  recommendation: string;
  generated_at: string;
};

export type ReportComparisonInput = {
  report_type: string;
  current_report_id: string;
  baseline_id?: string;
  use_default_baseline?: boolean;
};

export type QualityGateConfig = {
  fail_on_new_critical?: boolean;
  fail_on_new_high?: boolean;
  fail_on_new_medium?: boolean;
  max_new_high?: number;
  max_new_medium?: number;
  max_total_critical?: number;
  max_total_high?: number;
  fail_on_run_error?: boolean;
  fail_on_missing_report?: boolean;
  ignore_info?: boolean;
  ignore_noisy?: boolean;
};

export type QualityGateResult = {
  status: "passed" | "failed" | "warning" | string;
  failed_rules: string[];
  warnings: string[];
  comparison_summary: ReportComparisonSummary;
  severity_counts: ReportSummary;
  recommendation: string;
  ci_exit_code: number;
  generated_at: string;
};

export type QualityGateEvaluationInput = {
  report_type: string;
  current_report_id: string;
  baseline_id?: string;
  use_default_baseline?: boolean;
  gate_config?: QualityGateConfig;
  format?: string;
};

export type CIQualityGateResult = {
  status: string;
  exit_code: number;
  summary: string;
  report_url: string;
  comparison_url?: string;
  failed_rules: string[];
};

export type CIRunInput = {
  mode?: "safe_qa";
  use_latest_baseline?: boolean;
  baseline_id?: string;
  run_safe_qa?: boolean;
  use_latest_discovery?: boolean;
  start_url?: string;
  credential_profile_id?: string;
  provider_id?: string;
  max_pages?: number;
  max_depth?: number;
  max_scenarios?: number;
  include_quality_checks?: boolean;
  include_safe_explorer?: boolean;
  execute_safe_plan?: boolean;
  gate_config?: QualityGateConfig;
  wait?: boolean;
  timeout_seconds?: number;
  export_issues?: boolean;
  issue_export_config_id?: string;
};

export type CIRun = {
  id: string;
  project_id: string;
  qa_run_id?: string;
  baseline_id?: string;
  status: "passed" | "failed" | "warning" | "running" | "error" | string;
  exit_code: number;
  gate_status?: string;
  comparison_status?: string;
  report_url?: string;
  html_report_url?: string;
  issue_export_status?: string;
  summary_json?: Record<string, unknown>;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  error_message?: string;
};

export type CIRunResponse = {
  ci_run_id: string;
  project_id: string;
  status: string;
  qa_run_id?: string;
  report_url?: string;
  html_report_url?: string;
  baseline_id?: string;
  comparison_summary?: ReportComparisonSummary;
  quality_gate_result?: QualityGateResult;
  issue_export_summary?: IssueExportResult;
  exit_code: number;
  summary: string;
  created_at: string;
  completed_at?: string;
  error_message?: string;
};

export type IssueExportConfig = {
  id: string;
  project_id: string;
  provider: "github" | "gitlab" | string;
  name: string;
  base_url?: string;
  owner_or_namespace: string;
  repository_or_project: string;
  token_configured: boolean;
  default_labels?: string[];
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

export type IssueExportConfigInput = {
  provider: "github" | "gitlab";
  name: string;
  base_url?: string;
  owner_or_namespace: string;
  repository_or_project: string;
  token?: string;
  default_labels?: string[];
  enabled?: boolean;
};

export type IssueExportConfigTestResult = {
  success: boolean;
  provider: string;
  target: string;
  error_message?: string;
};

export type IssueExportInput = {
  issue_export_config_id?: string;
  severity_threshold?: "critical" | "high" | "medium";
  include_medium?: boolean;
  max_issues?: number;
  dry_run?: boolean;
  deduplicate_by_fingerprint?: boolean;
  labels?: string[];
  title_prefix?: string;
};

export type IssueExportPreview = {
  title: string;
  body: string;
  severity: string;
  category: string;
  affected_pages_count: number;
  representative_paths?: string[];
  labels?: string[];
  fingerprint: string;
};

export type IssueExportSkippedFinding = {
  fingerprint?: string;
  title: string;
  severity: string;
  reason: string;
};

export type IssueExportResult = {
  provider?: string;
  dry_run: boolean;
  status: string;
  created_count: number;
  skipped_count: number;
  issue_urls?: string[];
  errors?: string[];
  issues_to_create: IssueExportPreview[];
  skipped_findings: IssueExportSkippedFinding[];
  reasons?: string[];
  generated_at: string;
};

export type Report = ReportIntelligenceFields & {
  run_id: string;
  project_id: string;
  run_type: string;
  status: string;
  summary: ReportSummary;
  findings: Finding[];
  evidence: Evidence[];
  metadata: {
    page_title?: string;
    created_at?: string;
    jobs?: RunJob[];
    error_message?: string;
    [key: string]: unknown;
  };
  ai_analysis: AIAnalysis | null;
  test_plans: TestPlanRef[];
  api_spec?: APISpec;
  api_auth?: APIAuthSummary;
  api_summary?: APISmokeSummary;
  api_results?: APICheckResult[];
  login_summary?: {
    credential_profile_id?: string;
    credential_profile_name?: string;
    login_status?: string;
    login_url?: string;
    login_final_url?: string;
    page_title?: string;
    login_duration_ms?: number;
    authenticated_target_url?: string;
    failure_reason?: string;
  };
};

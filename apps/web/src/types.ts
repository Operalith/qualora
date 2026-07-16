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

export type TestRun = {
  id: string;
  project_id: string;
  run_type: "full" | "browser_smoke" | "api_smoke" | "login_check" | "authenticated_browser_smoke" | string;
  api_spec_id?: string;
  credential_profile_id?: string;
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

export type AuthorizationCheckReport = {
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

export type DiscoveryReport = DiscoveryMap & {
  generated_at: string;
  settings: Record<string, unknown>;
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

export type TestPlanExecutionReport = {
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
  max_pages?: number;
  max_depth?: number;
  max_scenarios?: number;
  execute?: boolean;
  use_existing_discovery_run_id?: string;
  use_latest_discovery?: boolean;
  provider_id?: string;
  product_context?: string;
  focus_areas?: string[];
};

export type QARun = {
  id: string;
  project_id: string;
  status: "queued" | "running_discovery" | "generating_plan" | "previewing_execution" | "executing_plan" | "completed" | "failed" | "error" | string;
  mode: "safe" | string;
  discovery_run_id?: string;
  test_plan_id?: string;
  test_plan_execution_id?: string;
  credential_profile_id?: string;
  error_message?: string;
  summary: Record<string, unknown>;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

export type QARunReport = {
  run: QARun;
  project: Project;
  discovery_run?: DiscoveryRun;
  discovery_summary?: DiscoverySummary;
  test_plan?: TestPlan;
  execution_preview?: TestPlanExecutionPreview;
  execution_report?: TestPlanExecutionReport;
  findings: Finding[];
  evidence: Evidence[];
  safety_notes: string[];
  limitations: string[];
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

export type APICheckResult = {
  id: string;
  run_id: string;
  api_spec_id: string;
  operation_id?: string;
  method: string;
  path: string;
  resolved_url?: string;
  status: "passed" | "failed" | "skipped" | "error" | string;
  http_status?: number;
  duration_ms?: number;
  response_content_type?: string;
  response_size_bytes?: number;
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
};

export type ReportSummary = {
  total_findings: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  info: number;
};

export type Report = {
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

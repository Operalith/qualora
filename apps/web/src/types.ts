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
  provider_id?: string;
  provider_name?: string;
  model: string;
  status: "queued" | "running" | "completed" | "failed";
  title: string;
  summary: string;
  plan_json: TestPlanPayload;
  risk_level: "low" | "medium" | "high" | "critical" | "";
  total_scenarios: number;
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
  product_context?: string;
  focus_areas: string[];
  max_scenarios: number;
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
};

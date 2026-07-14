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
};

import type {
  AIAnalysis,
  APICheckResult,
  APIOperation,
  APISpec,
  APISpecDetail,
  APISpecImportInput,
  AIProvider,
  AIProviderInput,
  AIProviderTestResult,
  AITestPlanInput,
  CreateProjectInput,
  Project,
  Report,
  TestPlan,
  TestPlanExecution,
  TestPlanExecutionDetail,
  TestPlanExecutionPreview,
  TestPlanExecutionReport,
  TestPlanExecutionRequest,
  TestRun
} from "./types";

declare global {
  interface Window {
    __QUALORA_CONFIG__?: {
      apiBaseUrl?: string;
    };
  }
}

export const API_BASE_URL = normalizeBaseURL(
  window.__QUALORA_CONFIG__?.apiBaseUrl || import.meta.env.VITE_QUALORA_API_BASE_URL || "http://localhost:8080"
);

export function htmlReportURL(runID: string): string {
  return `${API_BASE_URL}/api/v1/runs/${runID}/report.html`;
}

export function evidenceDownloadURL(evidenceID: string): string {
  return `${API_BASE_URL}/api/v1/evidence/${evidenceID}`;
}

export function testPlanExportURL(testPlanID: string): string {
  return `${API_BASE_URL}/api/v1/test-plans/${testPlanID}/export.json`;
}

export function testPlanExecutionHTMLReportURL(executionID: string): string {
  return `${API_BASE_URL}/api/v1/test-plan-executions/${executionID}/report.html`;
}

export async function listProjects(): Promise<Project[]> {
  const response = await request<{ projects: Project[] }>("/api/v1/projects");
  return response.projects;
}

export async function getProject(projectID: string): Promise<Project> {
  return request<Project>(`/api/v1/projects/${projectID}`);
}

export async function createProject(input: CreateProjectInput): Promise<Project> {
  return request<Project>("/api/v1/projects", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function listRuns(projectID?: string): Promise<TestRun[]> {
  const path = projectID ? `/api/v1/projects/${projectID}/runs` : "/api/v1/runs";
  const response = await request<{ runs: TestRun[] }>(path);
  return response.runs;
}

export async function startRun(projectID: string): Promise<TestRun> {
  return request<TestRun>(`/api/v1/projects/${projectID}/runs`, {
    method: "POST"
  });
}

export async function startBrowserSmokeRun(projectID: string): Promise<TestRun> {
  return request<TestRun>(`/api/v1/projects/${projectID}/browser-smoke-runs`, {
    method: "POST"
  });
}

export async function importAPISpec(projectID: string, input: APISpecImportInput): Promise<APISpecDetail> {
  return request<APISpecDetail>(`/api/v1/projects/${projectID}/api-specs`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function listAPISpecs(projectID: string): Promise<APISpec[]> {
  const response = await request<{ api_specs: APISpec[] }>(`/api/v1/projects/${projectID}/api-specs`);
  return response.api_specs;
}

export async function getAPISpec(apiSpecID: string): Promise<APISpecDetail> {
  return request<APISpecDetail>(`/api/v1/api-specs/${apiSpecID}`);
}

export async function listAPIOperations(apiSpecID: string): Promise<APIOperation[]> {
  const response = await request<{ operations: APIOperation[] }>(`/api/v1/api-specs/${apiSpecID}/operations`);
  return response.operations || [];
}

export async function deleteAPISpec(apiSpecID: string): Promise<void> {
  await request<{ deleted: boolean }>(`/api/v1/api-specs/${apiSpecID}`, {
    method: "DELETE"
  });
}

export async function startAPISmokeRun(apiSpecID: string): Promise<TestRun> {
  return request<TestRun>(`/api/v1/api-specs/${apiSpecID}/api-smoke-runs`, {
    method: "POST"
  });
}

export async function getRun(runID: string): Promise<TestRun> {
  return request<TestRun>(`/api/v1/runs/${runID}`);
}

export async function getReport(runID: string): Promise<Report> {
  return request<Report>(`/api/v1/runs/${runID}/report`);
}

export async function getAPIResults(runID: string): Promise<APICheckResult[]> {
  const response = await request<{ api_results: APICheckResult[] }>(`/api/v1/runs/${runID}/api-results`);
  return response.api_results;
}

export async function listAIProviders(): Promise<AIProvider[]> {
  const response = await request<{ providers: AIProvider[] }>("/api/v1/ai/providers");
  return response.providers;
}

export async function createAIProvider(input: AIProviderInput): Promise<AIProvider> {
  return request<AIProvider>("/api/v1/ai/providers", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function updateAIProvider(providerID: string, input: AIProviderInput): Promise<AIProvider> {
  return request<AIProvider>(`/api/v1/ai/providers/${providerID}`, {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export async function deleteAIProvider(providerID: string): Promise<void> {
  await request<{ deleted: boolean }>(`/api/v1/ai/providers/${providerID}`, {
    method: "DELETE"
  });
}

export async function testAIProvider(providerID: string): Promise<AIProviderTestResult> {
  return request<AIProviderTestResult>(`/api/v1/ai/providers/${providerID}/test`, {
    method: "POST"
  });
}

export async function runAIAnalysis(runID: string, providerID?: string): Promise<AIAnalysis> {
  return request<AIAnalysis>(`/api/v1/runs/${runID}/ai-analysis`, {
    method: "POST",
    body: JSON.stringify(providerID ? { provider_id: providerID } : {})
  });
}

export async function getAIAnalysis(runID: string): Promise<AIAnalysis | null> {
  const response = await request<{ ai_analysis: AIAnalysis | null }>(`/api/v1/runs/${runID}/ai-analysis`);
  return response.ai_analysis;
}

export async function generateAITestPlan(projectID: string, input: AITestPlanInput): Promise<TestPlan> {
  return request<TestPlan>(`/api/v1/projects/${projectID}/ai-test-plans`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function listTestPlans(projectID: string): Promise<TestPlan[]> {
  const response = await request<{ test_plans: TestPlan[] }>(`/api/v1/projects/${projectID}/test-plans`);
  return response.test_plans;
}

export async function getTestPlan(testPlanID: string): Promise<TestPlan> {
  return request<TestPlan>(`/api/v1/test-plans/${testPlanID}`);
}

export async function deleteTestPlan(testPlanID: string): Promise<void> {
  await request<{ deleted: boolean }>(`/api/v1/test-plans/${testPlanID}`, {
    method: "DELETE"
  });
}

export async function previewTestPlanExecution(testPlanID: string, input: TestPlanExecutionRequest): Promise<TestPlanExecutionPreview> {
  return request<TestPlanExecutionPreview>(`/api/v1/test-plans/${testPlanID}/executions`, {
    method: "POST",
    body: JSON.stringify({ ...input, dry_run: true })
  });
}

export async function executeTestPlan(testPlanID: string, input: TestPlanExecutionRequest): Promise<TestPlanExecutionDetail> {
  return request<TestPlanExecutionDetail>(`/api/v1/test-plans/${testPlanID}/executions`, {
    method: "POST",
    body: JSON.stringify({ ...input, dry_run: false })
  });
}

export async function listTestPlanExecutions(testPlanID: string): Promise<TestPlanExecution[]> {
  const response = await request<{ executions: TestPlanExecution[] }>(`/api/v1/test-plans/${testPlanID}/executions`);
  return response.executions;
}

export async function getTestPlanExecution(executionID: string): Promise<TestPlanExecutionDetail> {
  return request<TestPlanExecutionDetail>(`/api/v1/test-plan-executions/${executionID}`);
}

export async function getTestPlanExecutionReport(executionID: string): Promise<TestPlanExecutionReport> {
  return request<TestPlanExecutionReport>(`/api/v1/test-plan-executions/${executionID}/report`);
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...init.headers
    }
  });
  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok) {
    const message = payload?.error?.message || `${response.status} ${response.statusText}`;
    throw new Error(message);
  }
  return payload as T;
}

function normalizeBaseURL(value: string): string {
  return value.replace(/\/+$/, "");
}

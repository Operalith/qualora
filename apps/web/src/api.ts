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
  AuthorizationCheck,
  AuthorizationCheckDetail,
  AuthorizationCheckInput,
  AuthorizationCheckReport,
  AuthorizationCheckRun,
  AuthorizationCheckRunInput,
  AuthenticatedBrowserSmokeInput,
  AuthResponse,
  CreateProjectInput,
  CredentialProfile,
  CredentialProfileInput,
  DiscoveryMap,
  DiscoveryReport,
  DiscoveryRun,
  DiscoveryRunInput,
  LoginInput,
  MeResponse,
  Project,
  ProjectSetupInput,
  ProjectSetupResponse,
  QualityCheckReport,
  QualityCheckRun,
  QualityCheckRunInput,
  QARun,
  QARunInput,
  QARunReport,
  SafeExplorerReport,
  SafeExplorerRun,
  SafeExplorerRunInput,
  SafeExplorerTrace,
  Report,
  SetupAdminInput,
  SetupStatus,
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

export function authorizationCheckHTMLReportURL(runID: string): string {
  return `${API_BASE_URL}/api/v1/authorization-check-runs/${runID}/report.html`;
}

export function discoveryHTMLReportURL(runID: string): string {
  return `${API_BASE_URL}/api/v1/discovery-runs/${runID}/report.html`;
}

export function qualityCheckHTMLReportURL(runID: string): string {
  return `${API_BASE_URL}/api/v1/quality-check-runs/${runID}/report.html`;
}

export function qaRunHTMLReportURL(runID: string): string {
  return `${API_BASE_URL}/api/v1/qa-runs/${runID}/report.html`;
}

export function safeExplorerHTMLReportURL(runID: string): string {
  return `${API_BASE_URL}/api/v1/safe-explorer-runs/${runID}/report.html`;
}

export async function getSetupStatus(): Promise<SetupStatus> {
  return request<SetupStatus>("/api/v1/setup/status");
}

export async function setupAdmin(input: SetupAdminInput): Promise<AuthResponse> {
  return request<AuthResponse>("/api/v1/setup/admin", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function login(input: LoginInput): Promise<AuthResponse> {
  return request<AuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function logout(): Promise<void> {
  await request<{ logged_out: boolean }>("/api/v1/auth/logout", {
    method: "POST"
  });
}

export async function me(): Promise<MeResponse> {
  return request<MeResponse>("/api/v1/auth/me");
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

export async function runProjectSetup(input: ProjectSetupInput): Promise<ProjectSetupResponse> {
  return request<ProjectSetupResponse>("/api/v1/onboarding/project-setup", {
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

export async function startAuthenticatedBrowserSmokeRun(projectID: string, input: AuthenticatedBrowserSmokeInput): Promise<TestRun> {
  return request<TestRun>(`/api/v1/projects/${projectID}/authenticated-browser-smoke-runs`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function listCredentialProfiles(projectID: string): Promise<CredentialProfile[]> {
  const response = await request<{ credential_profiles: CredentialProfile[] }>(`/api/v1/projects/${projectID}/credential-profiles`);
  return response.credential_profiles;
}

export async function createCredentialProfile(projectID: string, input: CredentialProfileInput): Promise<CredentialProfile> {
  return request<CredentialProfile>(`/api/v1/projects/${projectID}/credential-profiles`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function getCredentialProfile(profileID: string): Promise<CredentialProfile> {
  return request<CredentialProfile>(`/api/v1/credential-profiles/${profileID}`);
}

export async function updateCredentialProfile(profileID: string, input: CredentialProfileInput): Promise<CredentialProfile> {
  return request<CredentialProfile>(`/api/v1/credential-profiles/${profileID}`, {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export async function deleteCredentialProfile(profileID: string): Promise<void> {
  await request<void>(`/api/v1/credential-profiles/${profileID}`, {
    method: "DELETE"
  });
}

export async function testCredentialProfileLogin(profileID: string): Promise<TestRun> {
  return request<TestRun>(`/api/v1/credential-profiles/${profileID}/test-login`, {
    method: "POST"
  });
}

export async function listAuthorizationChecks(projectID: string): Promise<AuthorizationCheck[]> {
  const response = await request<{ authorization_checks: AuthorizationCheck[] }>(`/api/v1/projects/${projectID}/authorization-checks`);
  return response.authorization_checks;
}

export async function createAuthorizationCheck(projectID: string, input: AuthorizationCheckInput): Promise<AuthorizationCheck> {
  return request<AuthorizationCheck>(`/api/v1/projects/${projectID}/authorization-checks`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function updateAuthorizationCheck(checkID: string, input: AuthorizationCheckInput): Promise<AuthorizationCheck> {
  return request<AuthorizationCheck>(`/api/v1/authorization-checks/${checkID}`, {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export async function deleteAuthorizationCheck(checkID: string): Promise<void> {
  await request<void>(`/api/v1/authorization-checks/${checkID}`, {
    method: "DELETE"
  });
}

export async function listAuthorizationCheckRuns(projectID: string): Promise<AuthorizationCheckRun[]> {
  const response = await request<{ authorization_check_runs: AuthorizationCheckRun[] }>(`/api/v1/projects/${projectID}/authorization-check-runs`);
  return response.authorization_check_runs;
}

export async function startAuthorizationCheckRun(projectID: string, input: AuthorizationCheckRunInput = {}): Promise<AuthorizationCheckRun> {
  return request<AuthorizationCheckRun>(`/api/v1/projects/${projectID}/authorization-check-runs`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function getAuthorizationCheckRun(runID: string): Promise<AuthorizationCheckDetail> {
  return request<AuthorizationCheckDetail>(`/api/v1/authorization-check-runs/${runID}`);
}

export async function getAuthorizationCheckReport(runID: string): Promise<AuthorizationCheckReport> {
  return request<AuthorizationCheckReport>(`/api/v1/authorization-check-runs/${runID}/report`);
}

export async function listDiscoveryRuns(projectID: string): Promise<DiscoveryRun[]> {
  const response = await request<{ discovery_runs: DiscoveryRun[] }>(`/api/v1/projects/${projectID}/discovery-runs`);
  return response.discovery_runs;
}

export async function startDiscoveryRun(projectID: string, input: DiscoveryRunInput): Promise<DiscoveryRun> {
  return request<DiscoveryRun>(`/api/v1/projects/${projectID}/discovery-runs`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function getDiscoveryRun(runID: string): Promise<DiscoveryRun> {
  return request<DiscoveryRun>(`/api/v1/discovery-runs/${runID}`);
}

export async function getDiscoveryMap(runID: string): Promise<DiscoveryMap> {
  return request<DiscoveryMap>(`/api/v1/discovery-runs/${runID}/map`);
}

export async function getDiscoveryReport(runID: string): Promise<DiscoveryReport> {
  return request<DiscoveryReport>(`/api/v1/discovery-runs/${runID}/report`);
}

export async function listSafeExplorerRuns(projectID: string): Promise<SafeExplorerRun[]> {
  const response = await request<{ safe_explorer_runs: SafeExplorerRun[] }>(`/api/v1/projects/${projectID}/safe-explorer-runs`);
  return response.safe_explorer_runs;
}

export async function startSafeExplorerRun(projectID: string, input: SafeExplorerRunInput): Promise<SafeExplorerRun> {
  return request<SafeExplorerRun>(`/api/v1/projects/${projectID}/safe-explorer-runs`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function getSafeExplorerRun(runID: string): Promise<SafeExplorerRun> {
  return request<SafeExplorerRun>(`/api/v1/safe-explorer-runs/${runID}`);
}

export async function getSafeExplorerTrace(runID: string): Promise<SafeExplorerTrace> {
  return request<SafeExplorerTrace>(`/api/v1/safe-explorer-runs/${runID}/trace`);
}

export async function getSafeExplorerReport(runID: string): Promise<SafeExplorerReport> {
  return request<SafeExplorerReport>(`/api/v1/safe-explorer-runs/${runID}/report`);
}

export async function listQualityCheckRuns(projectID: string): Promise<QualityCheckRun[]> {
  const response = await request<{ quality_check_runs: QualityCheckRun[] }>(`/api/v1/projects/${projectID}/quality-check-runs`);
  return response.quality_check_runs;
}

export async function startQualityCheckRun(projectID: string, input: QualityCheckRunInput): Promise<QualityCheckRun> {
  return request<QualityCheckRun>(`/api/v1/projects/${projectID}/quality-check-runs`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function getQualityCheckRun(runID: string): Promise<QualityCheckRun> {
  return request<QualityCheckRun>(`/api/v1/quality-check-runs/${runID}`);
}

export async function getQualityCheckReport(runID: string): Promise<QualityCheckReport> {
  return request<QualityCheckReport>(`/api/v1/quality-check-runs/${runID}/report`);
}

export async function listQARuns(projectID: string): Promise<QARun[]> {
  const response = await request<{ qa_runs: QARun[] }>(`/api/v1/projects/${projectID}/qa-runs`);
  return response.qa_runs;
}

export async function startQARun(projectID: string, input: QARunInput): Promise<QARun> {
  return request<QARun>(`/api/v1/projects/${projectID}/qa-runs`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function getQARun(runID: string): Promise<QARun> {
  return request<QARun>(`/api/v1/qa-runs/${runID}`);
}

export async function executeQARun(runID: string): Promise<QARun> {
  return request<QARun>(`/api/v1/qa-runs/${runID}/execute`, {
    method: "POST"
  });
}

export async function getQARunReport(runID: string): Promise<QARunReport> {
  return request<QARunReport>(`/api/v1/qa-runs/${runID}/report`);
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
  const method = (init.method || "GET").toUpperCase();
  const csrfToken = csrfTokenFromCookie();
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      Accept: "application/json",
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...(requiresCSRF(method) && csrfToken ? { "X-Qualora-CSRF": csrfToken } : {}),
      ...init.headers
    }
  });
  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok) {
    const message = payload?.error?.message || `${response.status} ${response.statusText}`;
    if (response.status === 401) {
      window.dispatchEvent(new CustomEvent("qualora:unauthorized"));
    }
    throw new Error(message);
  }
  return payload as T;
}

function requiresCSRF(method: string): boolean {
  return !["GET", "HEAD", "OPTIONS"].includes(method);
}

function csrfTokenFromCookie(): string {
  const entry = document.cookie
    .split(";")
    .map((part) => part.trim())
    .find((part) => part.startsWith("qualora_csrf="));
  return entry ? decodeURIComponent(entry.slice("qualora_csrf=".length)) : "";
}

function normalizeBaseURL(value: string): string {
  return value.replace(/\/+$/, "");
}

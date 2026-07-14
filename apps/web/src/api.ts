import type { CreateProjectInput, Project, Report, TestRun } from "./types";

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

export async function getRun(runID: string): Promise<TestRun> {
  return request<TestRun>(`/api/v1/runs/${runID}`);
}

export async function getReport(runID: string): Promise<Report> {
  return request<Report>(`/api/v1/runs/${runID}/report`);
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

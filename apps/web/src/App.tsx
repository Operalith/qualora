import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  API_BASE_URL,
  createProject,
  getProject,
  getReport,
  getRun,
  htmlReportURL,
  listProjects,
  listRuns,
  startRun
} from "./api";
import type { CreateProjectInput, Evidence, Project, Report, RunJob, TestRun } from "./types";

type Route =
  | { name: "dashboard" }
  | { name: "new-project" }
  | { name: "project"; id: string }
  | { name: "runs" }
  | { name: "run"; id: string };

type LoadState<T> = {
  data: T;
  loading: boolean;
  error: string;
};

const emptyProjects: LoadState<Project[]> = { data: [], loading: true, error: "" };
const emptyRuns: LoadState<TestRun[]> = { data: [], loading: true, error: "" };

export default function App() {
  const [route, setRoute] = useHashRoute();
  const [projects, setProjects] = useState<LoadState<Project[]>>(emptyProjects);
  const [runs, setRuns] = useState<LoadState<TestRun[]>>(emptyRuns);

  const refresh = useCallback(async () => {
    setProjects((current) => ({ ...current, loading: true, error: "" }));
    setRuns((current) => ({ ...current, loading: true, error: "" }));
    try {
      const [nextProjects, nextRuns] = await Promise.all([listProjects(), listRuns()]);
      setProjects({ data: nextProjects, loading: false, error: "" });
      setRuns({ data: nextRuns, loading: false, error: "" });
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      setProjects((current) => ({ ...current, loading: false, error: message }));
      setRuns((current) => ({ ...current, loading: false, error: message }));
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!runs.data.some((run) => run.status === "pending" || run.status === "running")) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, runs.data]);

  const projectByID = useMemo(() => new Map(projects.data.map((project) => [project.id, project])), [projects.data]);

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <a className="brand" href="#/">
          <span>Qualora</span>
          <small>v0.3.0-alpha</small>
        </a>
        <nav>
          <a className={route.name === "dashboard" ? "active" : ""} href="#/">
            Projects
          </a>
          <a className={route.name === "runs" ? "active" : ""} href="#/runs">
            Runs
          </a>
          <a className={route.name === "new-project" ? "active" : ""} href="#/projects/new">
            New Project
          </a>
        </nav>
        <div className="sidebar-note">
          <span>API</span>
          <code>{API_BASE_URL}</code>
        </div>
      </aside>

      <main className="content">
        <header className="topbar">
          <div>
            <p className="eyebrow">Self-hosted QA</p>
            <h1>{titleForRoute(route)}</h1>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </header>

        {(projects.error || runs.error) && <Notice tone="danger" message={projects.error || runs.error} />}

        {route.name === "dashboard" && (
          <Dashboard projects={projects} runs={runs} projectByID={projectByID} onStartRun={refreshAfterStart(refresh)} />
        )}
        {route.name === "new-project" && <ProjectForm onCreated={(project) => setRoute({ name: "project", id: project.id })} />}
        {route.name === "project" && (
          <ProjectPage projectID={route.id} cachedProject={projectByID.get(route.id)} onRunStarted={refreshAfterStart(refresh)} />
        )}
        {route.name === "runs" && <RunsPage runs={runs} projectByID={projectByID} />}
        {route.name === "run" && <RunReportPage runID={route.id} cachedRun={runs.data.find((run) => run.id === route.id)} projectByID={projectByID} />}
      </main>
    </div>
  );
}

function Dashboard({
  projects,
  runs,
  projectByID,
  onStartRun
}: {
  projects: LoadState<Project[]>;
  runs: LoadState<TestRun[]>;
  projectByID: Map<string, Project>;
  onStartRun: (projectID: string) => Promise<void>;
}) {
  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>Projects</h2>
            <p>Configured frontend and API targets.</p>
          </div>
          <a className="button" href="#/projects/new">
            Create Project
          </a>
        </div>
        {projects.loading ? <SkeletonRows /> : <ProjectTable projects={projects.data} onStartRun={onStartRun} />}
      </section>

      <section>
        <div className="section-heading">
          <div>
            <h2>Recent Runs</h2>
            <p>Latest browser and API QA activity.</p>
          </div>
          <a className="button secondary-link" href="#/runs">
            View All Runs
          </a>
        </div>
        {runs.loading ? <SkeletonRows /> : <RunTable runs={runs.data.slice(0, 8)} projectByID={projectByID} />}
      </section>
    </div>
  );
}

function ProjectTable({ projects, onStartRun }: { projects: Project[]; onStartRun: (projectID: string) => Promise<void> }) {
  if (projects.length === 0) {
    return <EmptyState title="No projects yet" body="Create a project with a frontend URL, API base URL, or OpenAPI document." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Targets</th>
            <th>Allowed Hosts</th>
            <th>Created</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {projects.map((project) => (
            <tr key={project.id}>
              <td>
                <a href={`#/projects/${project.id}`}>{project.name}</a>
              </td>
              <td>{targetSummary(project)}</td>
              <td>{project.allowed_hosts.join(", ")}</td>
              <td>{formatDate(project.created_at)}</td>
              <td className="actions">
                <button type="button" className="secondary" onClick={() => void onStartRun(project.id)}>
                  Start Run
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function RunsPage({ runs, projectByID }: { runs: LoadState<TestRun[]>; projectByID: Map<string, Project> }) {
  return <section>{runs.loading ? <SkeletonRows /> : <RunTable runs={runs.data} projectByID={projectByID} />}</section>;
}

function RunTable({ runs, projectByID }: { runs: TestRun[]; projectByID: Map<string, Project> }) {
  if (runs.length === 0) {
    return <EmptyState title="No runs yet" body="Start a project run to collect findings and evidence." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Project</th>
            <th>Run</th>
            <th>Created</th>
            <th>Completed</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id}>
              <td>
                <StatusBadge status={run.status} />
              </td>
              <td>{projectByID.get(run.project_id)?.name || run.project_id}</td>
              <td>
                <a href={`#/runs/${run.id}`}>{shortID(run.id)}</a>
              </td>
              <td>{formatDate(run.created_at)}</td>
              <td>{run.completed_at ? formatDate(run.completed_at) : "Not completed"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function ProjectPage({
  projectID,
  cachedProject,
  onRunStarted
}: {
  projectID: string;
  cachedProject?: Project;
  onRunStarted: (projectID: string) => Promise<void>;
}) {
  const [project, setProject] = useState<Project | undefined>(cachedProject);
  const [runs, setRuns] = useState<LoadState<TestRun[]>>({ data: [], loading: true, error: "" });
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setRuns((current) => ({ ...current, loading: true, error: "" }));
    setError("");
    try {
      const [nextProject, nextRuns] = await Promise.all([cachedProject ? Promise.resolve(cachedProject) : getProject(projectID), listRuns(projectID)]);
      setProject(nextProject);
      setRuns({ data: nextRuns, loading: false, error: "" });
    } catch (loadError) {
      const message = loadError instanceof Error ? loadError.message : String(loadError);
      setError(message);
      setRuns((current) => ({ ...current, loading: false, error: message }));
    }
  }, [cachedProject, projectID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (!project) {
    return <SkeletonRows />;
  }

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{project.name}</h2>
            <p>{targetSummary(project)}</p>
          </div>
          <button
            type="button"
            onClick={async () => {
              await onRunStarted(project.id);
              await refresh();
            }}
          >
            Start Run
          </button>
        </div>
        <div className="detail-grid">
          <Field label="Frontend URL" value={project.frontend_url || "Not configured"} />
          <Field label="API Base URL" value={project.api_base_url || "Not configured"} />
          <Field label="OpenAPI URL" value={project.openapi_url || "Not configured"} />
          <Field label="Allowed Hosts" value={project.allowed_hosts.join(", ")} />
          <Field label="Private Targets" value={project.allow_private_targets ? "Allowed" : "Blocked by default"} />
          <Field label="Created" value={formatDate(project.created_at)} />
        </div>
      </section>

      <section>
        <div className="section-heading">
          <div>
            <h2>Runs</h2>
            <p>Runs for this project.</p>
          </div>
        </div>
        {runs.loading ? <SkeletonRows /> : <RunTable runs={runs.data} projectByID={new Map([[project.id, project]])} />}
      </section>
    </div>
  );
}

function ProjectForm({ onCreated }: { onCreated: (project: Project) => void }) {
  const [form, setForm] = useState({
    name: "",
    frontend_url: "",
    api_base_url: "",
    openapi_url: "",
    allowed_hosts: "",
    allow_private_targets: false
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError("");
    const payload: CreateProjectInput = {
      name: form.name.trim(),
      frontend_url: form.frontend_url.trim(),
      api_base_url: form.api_base_url.trim(),
      openapi_url: form.openapi_url.trim(),
      allowed_hosts: splitHosts(form.allowed_hosts),
      security_mode: "passive",
      destructive_actions: false,
      allow_private_targets: form.allow_private_targets
    };
    try {
      const project = await createProject(payload);
      onCreated(project);
    } catch (createError) {
      setError(createError instanceof Error ? createError.message : String(createError));
    } finally {
      setSaving(false);
    }
  }

  return (
    <section>
      <form className="project-form" onSubmit={(event) => void submit(event)}>
        {error && <Notice tone="danger" message={error} />}
        <label>
          Project Name
          <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
        </label>
        <label>
          Frontend URL
          <input
            value={form.frontend_url}
            placeholder="https://example.com"
            onChange={(event) => setForm({ ...form, frontend_url: event.target.value })}
          />
        </label>
        <label>
          API Base URL
          <input
            value={form.api_base_url}
            placeholder="https://api.example.com"
            onChange={(event) => setForm({ ...form, api_base_url: event.target.value })}
          />
        </label>
        <label>
          OpenAPI URL
          <input
            value={form.openapi_url}
            placeholder="https://api.example.com/openapi.json"
            onChange={(event) => setForm({ ...form, openapi_url: event.target.value })}
          />
        </label>
        <label>
          Allowed Hosts
          <textarea
            value={form.allowed_hosts}
            placeholder="example.com, api.example.com"
            onChange={(event) => setForm({ ...form, allowed_hosts: event.target.value })}
            required
          />
        </label>
        <label className="check-row">
          <input
            type="checkbox"
            checked={form.allow_private_targets}
            onChange={(event) => setForm({ ...form, allow_private_targets: event.target.checked })}
          />
          Allow private or local targets for this project
        </label>
        <div className="form-actions">
          <button type="submit" disabled={saving}>
            {saving ? "Creating" : "Create Project"}
          </button>
          <a className="button secondary-link" href="#/">
            Cancel
          </a>
        </div>
      </form>
    </section>
  );
}

function RunReportPage({ runID, cachedRun, projectByID }: { runID: string; cachedRun?: TestRun; projectByID: Map<string, Project> }) {
  const [run, setRun] = useState<TestRun | undefined>(cachedRun);
  const [report, setReport] = useState<Report | undefined>();
  const [project, setProject] = useState<Project | undefined>(cachedRun ? projectByID.get(cachedRun.project_id) : undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextRun = await getRun(runID);
      const [nextReport, nextProject] = await Promise.all([getReport(runID), projectByID.get(nextRun.project_id) ?? getProject(nextRun.project_id)]);
      setRun(nextRun);
      setReport(nextReport);
      setProject(nextProject);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [projectByID, runID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!run || (run.status !== "pending" && run.status !== "running")) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, run]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading && !report) {
    return <SkeletonRows />;
  }
  if (!run || !report || !project) {
    return <Notice tone="danger" message="Run report could not be loaded." />;
  }

  const browserEvidence = report.evidence.filter((item) => item.type === "browser_observations");
  const apiEvidence = report.evidence.filter((item) => item.type === "api_observations" || item.type === "openapi_summary");

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{project.name}</h2>
            <p>
              <StatusBadge status={run.status} /> <span className="muted">Run {run.id}</span>
            </p>
          </div>
          <a className="button" href={htmlReportURL(run.id)} target="_blank" rel="noreferrer">
            HTML Report
          </a>
        </div>
        <div className="summary-grid">
          <Metric label="Total" value={report.summary.total_findings} />
          <Metric label="Critical" value={report.summary.critical} tone="critical" />
          <Metric label="High" value={report.summary.high} tone="high" />
          <Metric label="Medium" value={report.summary.medium} tone="medium" />
          <Metric label="Low" value={report.summary.low} tone="low" />
          <Metric label="Info" value={report.summary.info} tone="info" />
        </div>
        <div className="detail-grid compact">
          <Field label="Created" value={formatDate(run.created_at)} />
          <Field label="Started" value={run.started_at ? formatDate(run.started_at) : "Not started"} />
          <Field label="Completed" value={run.completed_at ? formatDate(run.completed_at) : "Not completed"} />
          <Field label="Page Title" value={run.page_title || "Not captured"} />
        </div>
      </section>

      <section>
        <h2>Findings</h2>
        <FindingsTable report={report} />
      </section>

      <section>
        <h2>Evidence</h2>
        <EvidenceTable evidence={report.evidence} />
      </section>

      <section>
        <h2>Browser Metadata</h2>
        <MetadataBlocks evidence={browserEvidence} empty="No browser metadata for this run." />
      </section>

      <section>
        <h2>API Metadata</h2>
        <MetadataBlocks evidence={apiEvidence} empty="No API metadata for this run." />
      </section>

      <section>
        <h2>Jobs Metadata</h2>
        <JobsTable jobs={report.metadata.jobs || []} />
      </section>
    </div>
  );
}

function FindingsTable({ report }: { report: Report }) {
  if (report.findings.length === 0) {
    return <EmptyState title="No findings" body="This run did not record any findings." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Severity</th>
            <th>Title</th>
            <th>Category</th>
            <th>Description</th>
            <th>Recommendation</th>
          </tr>
        </thead>
        <tbody>
          {report.findings.map((finding) => (
            <tr key={finding.id}>
              <td>
                <span className={`severity ${finding.severity}`}>{finding.severity}</span>
              </td>
              <td>{finding.title}</td>
              <td>{finding.category}</td>
              <td>{finding.description}</td>
              <td>{finding.recommendation}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function EvidenceTable({ evidence }: { evidence: Evidence[] }) {
  if (evidence.length === 0) {
    return <EmptyState title="No evidence" body="This run did not record evidence metadata." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Type</th>
            <th>URI</th>
            <th>Created</th>
            <th>Metadata</th>
          </tr>
        </thead>
        <tbody>
          {evidence.map((item) => (
            <tr key={item.id}>
              <td>{item.type}</td>
              <td>
                <code>{item.uri}</code>
              </td>
              <td>{item.created_at ? formatDate(item.created_at) : "Not set"}</td>
              <td>
                <pre>{JSON.stringify(item.metadata, null, 2)}</pre>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function MetadataBlocks({ evidence, empty }: { evidence: Evidence[]; empty: string }) {
  if (evidence.length === 0) {
    return <p className="muted">{empty}</p>;
  }
  return (
    <div className="metadata-stack">
      {evidence.map((item) => (
        <div key={item.id}>
          <h3>{item.type}</h3>
          <pre>{JSON.stringify(item.metadata, null, 2)}</pre>
        </div>
      ))}
    </div>
  );
}

function JobsTable({ jobs }: { jobs: RunJob[] }) {
  if (jobs.length === 0) {
    return <p className="muted">No job metadata is available for this run.</p>;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Kind</th>
            <th>Status</th>
            <th>Started</th>
            <th>Completed</th>
            <th>Error</th>
          </tr>
        </thead>
        <tbody>
          {jobs.map((job) => (
            <tr key={job.id}>
              <td>{job.kind}</td>
              <td>
                <StatusBadge status={job.status} />
              </td>
              <td>{job.started_at ? formatDate(job.started_at) : "Not started"}</td>
              <td>{job.completed_at ? formatDate(job.completed_at) : "Not completed"}</td>
              <td>{job.error_message || ""}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div className="field">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function Metric({ label, value, tone }: { label: string; value: number; tone?: string }) {
  return (
    <div className={`metric ${tone || ""}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  return <span className={`status ${status}`}>{status}</span>;
}

function Notice({ tone, message }: { tone: "danger" | "info"; message: string }) {
  return <div className={`notice ${tone}`}>{message}</div>;
}

function EmptyState({ title, body }: { title: string; body: string }) {
  return (
    <div className="empty-state">
      <h3>{title}</h3>
      <p>{body}</p>
    </div>
  );
}

function SkeletonRows() {
  return (
    <div className="skeleton-stack" aria-label="Loading">
      <span />
      <span />
      <span />
    </div>
  );
}

function useHashRoute(): [Route, (route: Route) => void] {
  const [route, setRouteState] = useState<Route>(() => parseHash(window.location.hash));

  useEffect(() => {
    const onHashChange = () => setRouteState(parseHash(window.location.hash));
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, []);

  const setRoute = (nextRoute: Route) => {
    window.location.hash = hashForRoute(nextRoute);
    setRouteState(nextRoute);
  };

  return [route, setRoute];
}

function parseHash(hash: string): Route {
  const parts = hash.replace(/^#\/?/, "").split("/").filter(Boolean);
  if (parts.length === 0) {
    return { name: "dashboard" };
  }
  if (parts[0] === "projects" && parts[1] === "new") {
    return { name: "new-project" };
  }
  if (parts[0] === "projects" && parts[1]) {
    return { name: "project", id: parts[1] };
  }
  if (parts[0] === "runs" && parts[1]) {
    return { name: "run", id: parts[1] };
  }
  if (parts[0] === "runs") {
    return { name: "runs" };
  }
  return { name: "dashboard" };
}

function hashForRoute(route: Route): string {
  switch (route.name) {
    case "dashboard":
      return "/";
    case "new-project":
      return "/projects/new";
    case "project":
      return `/projects/${route.id}`;
    case "runs":
      return "/runs";
    case "run":
      return `/runs/${route.id}`;
  }
}

function titleForRoute(route: Route): string {
  switch (route.name) {
    case "dashboard":
      return "Projects";
    case "new-project":
      return "Create Project";
    case "project":
      return "Project Details";
    case "runs":
      return "Runs";
    case "run":
      return "Run Report";
  }
}

function refreshAfterStart(refresh: () => Promise<void>) {
  return async (projectID: string) => {
    await startRun(projectID);
    await refresh();
  };
}

function splitHosts(input: string): string[] {
  return input
    .split(/[,\n]/)
    .map((host) => host.trim())
    .filter(Boolean);
}

function targetSummary(project: Project): string {
  const targets = [];
  if (project.frontend_url) {
    targets.push("browser");
  }
  if (project.api_base_url || project.openapi_url) {
    targets.push("api");
  }
  return targets.length > 0 ? targets.join(" + ") : "no runnable targets";
}

function shortID(id: string): string {
  return id.slice(0, 8);
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(new Date(value));
}

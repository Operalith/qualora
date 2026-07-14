import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  API_BASE_URL,
  createAIProvider,
  createProject,
  deleteAIProvider,
  deleteTestPlan,
  evidenceDownloadURL,
  generateAITestPlan,
  getProject,
  getReport,
  getRun,
  getTestPlan,
  htmlReportURL,
  listAIProviders,
  listProjects,
  listRuns,
  listTestPlans,
  runAIAnalysis,
  startBrowserSmokeRun,
  startRun,
  testPlanExportURL,
  testAIProvider,
  updateAIProvider
} from "./api";
import type {
  AIAnalysis,
  AIProvider,
  AIProviderInput,
  AIProviderTestResult,
  AITestPlanInput,
  CreateProjectInput,
  Evidence,
  Project,
  Report,
  RunJob,
  TestPlan,
  TestPlanPayload,
  TestPlanScenario,
  TestRun
} from "./types";

type Route =
  | { name: "dashboard" }
  | { name: "new-project" }
  | { name: "project"; id: string }
  | { name: "runs" }
  | { name: "ai-providers" }
  | { name: "test-plans" }
  | { name: "test-plan"; id: string }
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
    if (!runs.data.some((run) => isActiveRunStatus(run.status))) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, runs.data]);

  const projectByID = useMemo(() => new Map(projects.data.map((project) => [project.id, project])), [projects.data]);
  const startFullRun = async (projectID: string) => {
    const run = await startRun(projectID);
    await refresh();
    setRoute({ name: "run", id: run.id });
  };
  const startBrowserRun = async (projectID: string) => {
    const run = await startBrowserSmokeRun(projectID);
    await refresh();
    setRoute({ name: "run", id: run.id });
  };

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <a className="brand" href="#/">
          <span>Qualora</span>
          <small>v0.6.0-alpha</small>
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
          <a className={route.name === "ai-providers" ? "active" : ""} href="#/ai-providers">
            AI Providers
          </a>
          <a className={route.name === "test-plans" || route.name === "test-plan" ? "active" : ""} href="#/test-plans">
            Test Plans
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
          <Dashboard projects={projects} runs={runs} projectByID={projectByID} onStartRun={startFullRun} />
        )}
        {route.name === "new-project" && <ProjectForm onCreated={(project) => setRoute({ name: "project", id: project.id })} />}
        {route.name === "project" && (
          <ProjectPage
            projectID={route.id}
            cachedProject={projectByID.get(route.id)}
            onStartRun={startFullRun}
            onStartBrowserSmoke={startBrowserRun}
            onOpenTestPlan={(id) => setRoute({ name: "test-plan", id })}
          />
        )}
        {route.name === "runs" && <RunsPage runs={runs} projectByID={projectByID} />}
        {route.name === "ai-providers" && <AIProvidersPage />}
        {route.name === "test-plans" && <TestPlansPage projects={projects} />}
        {route.name === "test-plan" && <TestPlanDetailPage testPlanID={route.id} projectByID={projectByID} />}
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

type ProviderFormState = {
  id: string;
  name: string;
  preset: string;
  type: "openai-compatible";
  base_url: string;
  model: string;
  api_key: string;
  extra_headers: string;
  temperature: number;
  max_output_tokens: number;
  timeout_seconds: number;
  send_screenshots: boolean;
  send_html: boolean;
  send_network_bodies: boolean;
  redaction_enabled: boolean;
  is_default: boolean;
};

const providerPresets = [
  { value: "disabled", label: "Disabled" },
  { value: "openai", label: "OpenAI" },
  { value: "openrouter", label: "OpenRouter" },
  { value: "ollama", label: "Ollama" },
  { value: "custom", label: "Custom OpenAI-compatible" }
];

function AIProvidersPage() {
  const [providers, setProviders] = useState<LoadState<AIProvider[]>>({ data: [], loading: true, error: "" });
  const [form, setForm] = useState<ProviderFormState>(() => providerFormDefaults("openai"));
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [testResults, setTestResults] = useState<Record<string, AIProviderTestResult>>({});

  const refresh = useCallback(async () => {
    setProviders((current) => ({ ...current, loading: true, error: "" }));
    try {
      const nextProviders = await listAIProviders();
      setProviders({ data: nextProviders, loading: false, error: "" });
    } catch (loadError) {
      const nextError = loadError instanceof Error ? loadError.message : String(loadError);
      setProviders((current) => ({ ...current, loading: false, error: nextError }));
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  function selectPreset(preset: string) {
    const defaults = providerFormDefaults(preset);
    setForm({ ...defaults, id: form.id, is_default: form.is_default });
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (form.preset === "disabled") {
      setError("AI is disabled. Choose a provider preset to create or update a provider.");
      return;
    }
    setSaving(true);
    setError("");
    setMessage("");
    try {
      const input = inputForProviderForm(form, form.id !== "");
      if (form.id) {
        await updateAIProvider(form.id, input);
        setMessage("AI provider updated.");
      } else {
        await createAIProvider(input);
        setMessage("AI provider created.");
      }
      setForm(providerFormDefaults("openai"));
      await refresh();
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : String(saveError));
    } finally {
      setSaving(false);
    }
  }

  function editProvider(provider: AIProvider) {
    setError("");
    setMessage("");
    setForm({
      id: provider.id,
      name: provider.name,
      preset: provider.preset || "custom",
      type: provider.type,
      base_url: provider.base_url,
      model: provider.model,
      api_key: "",
      extra_headers: "",
      temperature: provider.temperature,
      max_output_tokens: provider.max_output_tokens,
      timeout_seconds: provider.timeout_seconds,
      send_screenshots: provider.send_screenshots,
      send_html: provider.send_html,
      send_network_bodies: provider.send_network_bodies,
      redaction_enabled: provider.redaction_enabled,
      is_default: provider.is_default
    });
  }

  async function testProvider(providerID: string) {
    setError("");
    try {
      const result = await testAIProvider(providerID);
      setTestResults((current) => ({ ...current, [providerID]: result }));
    } catch (testError) {
      setError(testError instanceof Error ? testError.message : String(testError));
    }
  }

  async function removeProvider(providerID: string) {
    if (!window.confirm("Delete this AI provider?")) {
      return;
    }
    setError("");
    setMessage("");
    try {
      await deleteAIProvider(providerID);
      if (form.id === providerID) {
        setForm(providerFormDefaults("openai"));
      }
      await refresh();
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : String(deleteError));
    }
  }

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>AI Providers</h2>
            <p>Optional OpenAI-compatible providers for human-friendly report analysis.</p>
          </div>
          {form.id && (
            <button type="button" className="secondary" onClick={() => setForm(providerFormDefaults("openai"))}>
              New Provider
            </button>
          )}
        </div>
        <Notice
          tone="info"
          message="This alpha build has no authentication. Only configure provider credentials in trusted local/self-hosted environments."
        />
        {providers.error && <Notice tone="danger" message={providers.error} />}
        {error && <Notice tone="danger" message={error} />}
        {message && <Notice tone="info" message={message} />}
        {providers.loading ? (
          <SkeletonRows />
        ) : providers.data.length === 0 ? (
          <EmptyState title="No AI providers" body="Qualora works without AI. Add a provider only when you want optional analysis." />
        ) : (
          <AIProviderTable providers={providers.data} testResults={testResults} onEdit={editProvider} onTest={testProvider} onDelete={removeProvider} />
        )}
      </section>

      <section>
        <h2>{form.id ? "Edit Provider" : "Create Provider"}</h2>
        <form className="project-form provider-form" onSubmit={(event) => void submit(event)}>
          <label>
            Preset
            <select value={form.preset} onChange={(event) => selectPreset(event.target.value)}>
              {providerPresets.map((preset) => (
                <option key={preset.value} value={preset.value}>
                  {preset.label}
                </option>
              ))}
            </select>
          </label>
          {form.preset === "disabled" ? (
            <p className="muted">AI analysis is disabled until you configure a provider. Existing deterministic QA reports still work.</p>
          ) : (
            <>
              <div className="form-grid two">
                <label>
                  Name
                  <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
                </label>
                <label>
                  Provider Type
                  <input value={form.type} readOnly />
                </label>
              </div>
              <label>
                Base URL
                <input value={form.base_url} onChange={(event) => setForm({ ...form, base_url: event.target.value })} required />
              </label>
              <div className="form-grid two">
                <label>
                  Model
                  <input value={form.model} onChange={(event) => setForm({ ...form, model: event.target.value })} required />
                </label>
                <label>
                  API Key
                  <input
                    type="password"
                    value={form.api_key}
                    placeholder={form.id ? "Leave blank to keep existing key" : "Optional for local providers"}
                    onChange={(event) => setForm({ ...form, api_key: event.target.value })}
                  />
                </label>
              </div>
              <label>
                Extra Headers
                <textarea
                  value={form.extra_headers}
                  placeholder='{"X-OpenRouter-Title":"Qualora"}'
                  onChange={(event) => setForm({ ...form, extra_headers: event.target.value })}
                />
              </label>
              <div className="form-grid three">
                <label>
                  Temperature
                  <input
                    type="number"
                    min="0"
                    max="2"
                    step="0.1"
                    value={form.temperature}
                    onChange={(event) => setForm({ ...form, temperature: Number(event.target.value) })}
                  />
                </label>
                <label>
                  Max Output Tokens
                  <input
                    type="number"
                    min="1"
                    max="10000"
                    value={form.max_output_tokens}
                    onChange={(event) => setForm({ ...form, max_output_tokens: Number(event.target.value) })}
                  />
                </label>
                <label>
                  Timeout Seconds
                  <input
                    type="number"
                    min="1"
                    max="180"
                    value={form.timeout_seconds}
                    onChange={(event) => setForm({ ...form, timeout_seconds: Number(event.target.value) })}
                  />
                </label>
              </div>
              <div className="toggle-grid">
                <label className="check-row">
                  <input
                    type="checkbox"
                    checked={form.redaction_enabled}
                    onChange={(event) => setForm({ ...form, redaction_enabled: event.target.checked })}
                  />
                  Redaction enabled
                </label>
                <label className="check-row">
                  <input
                    type="checkbox"
                    checked={form.send_screenshots}
                    onChange={(event) => setForm({ ...form, send_screenshots: event.target.checked })}
                  />
                  Send screenshots
                </label>
                <label className="check-row">
                  <input type="checkbox" checked={form.send_html} onChange={(event) => setForm({ ...form, send_html: event.target.checked })} />
                  Send HTML
                </label>
                <label className="check-row">
                  <input
                    type="checkbox"
                    checked={form.send_network_bodies}
                    onChange={(event) => setForm({ ...form, send_network_bodies: event.target.checked })}
                  />
                  Send network bodies
                </label>
                <label className="check-row">
                  <input type="checkbox" checked={form.is_default} onChange={(event) => setForm({ ...form, is_default: event.target.checked })} />
                  Default provider
                </label>
              </div>
            </>
          )}
          <div className="form-actions">
            <button type="submit" disabled={saving || form.preset === "disabled"}>
              {saving ? "Saving" : form.id ? "Update Provider" : "Create Provider"}
            </button>
            <button type="button" className="secondary" onClick={() => setForm(providerFormDefaults("openai"))}>
              Reset
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}

function AIProviderTable({
  providers,
  testResults,
  onEdit,
  onTest,
  onDelete
}: {
  providers: AIProvider[];
  testResults: Record<string, AIProviderTestResult>;
  onEdit: (provider: AIProvider) => void;
  onTest: (providerID: string) => void;
  onDelete: (providerID: string) => void;
}) {
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Preset</th>
            <th>Model</th>
            <th>Credentials</th>
            <th>Safe Defaults</th>
            <th>Test</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {providers.map((provider) => {
            const result = testResults[provider.id];
            return (
              <tr key={provider.id}>
                <td>
                  <strong>{provider.name}</strong>
                  {provider.is_default && <span className="pill">default</span>}
                </td>
                <td>{provider.preset || "custom"}</td>
                <td>
                  <code>{provider.model}</code>
                </td>
                <td>
                  API key {provider.api_key_configured ? "configured" : "not set"}
                  <br />
                  Headers {provider.extra_headers_configured ? "configured" : "not set"}
                </td>
                <td>
                  Redaction {provider.redaction_enabled ? "on" : "off"}
                  <br />
                  Screenshots {provider.send_screenshots ? "on" : "off"}
                </td>
                <td>
                  {result ? (
                    <span className={result.success ? "result-ok" : "result-failed"}>
                      {result.success ? `OK ${result.latency_ms}ms` : result.error_message || "Failed"}
                    </span>
                  ) : (
                    <span className="muted">Not tested</span>
                  )}
                </td>
                <td className="actions">
                  <div className="button-row compact">
                    <button type="button" className="secondary" onClick={() => onEdit(provider)}>
                      Edit
                    </button>
                    <button type="button" className="secondary" onClick={() => onTest(provider.id)}>
                      Test
                    </button>
                    <button type="button" className="secondary danger" onClick={() => onDelete(provider.id)}>
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
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
  onStartRun,
  onStartBrowserSmoke,
  onOpenTestPlan
}: {
  projectID: string;
  cachedProject?: Project;
  onStartRun: (projectID: string) => Promise<void>;
  onStartBrowserSmoke: (projectID: string) => Promise<void>;
  onOpenTestPlan: (testPlanID: string) => void;
}) {
  const [project, setProject] = useState<Project | undefined>(cachedProject);
  const [runs, setRuns] = useState<LoadState<TestRun[]>>({ data: [], loading: true, error: "" });
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [testPlans, setTestPlans] = useState<LoadState<TestPlan[]>>({ data: [], loading: true, error: "" });
  const [error, setError] = useState("");
  const [starting, setStarting] = useState("");

  const refresh = useCallback(async () => {
    setRuns((current) => ({ ...current, loading: true, error: "" }));
    setTestPlans((current) => ({ ...current, loading: true, error: "" }));
    setError("");
    try {
      const [nextProject, nextRuns, nextProviders, nextTestPlans] = await Promise.all([
        cachedProject ? Promise.resolve(cachedProject) : getProject(projectID),
        listRuns(projectID),
        listAIProviders(),
        listTestPlans(projectID)
      ]);
      setProject(nextProject);
      setRuns({ data: nextRuns, loading: false, error: "" });
      setProviders(nextProviders);
      setTestPlans({ data: nextTestPlans, loading: false, error: "" });
    } catch (loadError) {
      const message = loadError instanceof Error ? loadError.message : String(loadError);
      setError(message);
      setRuns((current) => ({ ...current, loading: false, error: message }));
      setTestPlans((current) => ({ ...current, loading: false, error: message }));
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

  async function startProjectRun(kind: "browser" | "full") {
    if (!project) {
      return;
    }
    setStarting(kind);
    setError("");
    try {
      if (kind === "browser") {
        await onStartBrowserSmoke(project.id);
      } else {
        await onStartRun(project.id);
      }
    } catch (startError) {
      setError(startError instanceof Error ? startError.message : String(startError));
    } finally {
      setStarting("");
    }
  }

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{project.name}</h2>
            <p>{targetSummary(project)}</p>
          </div>
          <div className="button-row">
            {project.frontend_url && (
              <button type="button" disabled={starting !== ""} onClick={() => void startProjectRun("browser")}>
                {starting === "browser" ? "Starting" : "Run browser smoke test"}
              </button>
            )}
            <button type="button" className="secondary" disabled={starting !== ""} onClick={() => void startProjectRun("full")}>
              {starting === "full" ? "Starting" : "Start full run"}
            </button>
          </div>
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

      <section>
        <div className="section-heading">
          <div>
            <h2>AI Test Plans</h2>
            <p>Reviewable test ideas generated from project and run metadata.</p>
          </div>
          <a className="button secondary-link" href="#/test-plans">
            View All Plans
          </a>
        </div>
        <Notice tone="info" message="AI test plans are suggestions only. Qualora does not execute generated steps in this alpha release." />
        <GenerateAITestPlanForm
          project={project}
          runs={runs.data}
          providers={providers}
          onGenerated={async (plan) => {
            await refresh();
            onOpenTestPlan(plan.id);
          }}
        />
        <div className="section-split">
          {testPlans.loading ? (
            <SkeletonRows />
          ) : (
            <TestPlanTable testPlans={testPlans.data} projectsByID={new Map([[project.id, project]])} onDeleted={() => void refresh()} />
          )}
        </div>
      </section>
    </div>
  );
}

const focusAreaOptions = [
  { value: "smoke", label: "Smoke" },
  { value: "functional", label: "Functional" },
  { value: "negative", label: "Negative" },
  { value: "accessibility", label: "Accessibility" },
  { value: "performance", label: "Performance" },
  { value: "security-passive", label: "Passive security" },
  { value: "authorization", label: "Authorization" },
  { value: "api", label: "API" },
  { value: "visual", label: "Visual" },
  { value: "regression", label: "Regression" }
];

function GenerateAITestPlanForm({
  project,
  runs,
  providers,
  onGenerated
}: {
  project: Project;
  runs: TestRun[];
  providers: AIProvider[];
  onGenerated: (plan: TestPlan) => Promise<void>;
}) {
  const [providerID, setProviderID] = useState("");
  const [runID, setRunID] = useState("");
  const [productContext, setProductContext] = useState("");
  const [focusAreas, setFocusAreas] = useState<string[]>(["smoke", "functional", "negative", "accessibility", "regression"]);
  const [maxScenarios, setMaxScenarios] = useState(10);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (providerID || providers.length === 0) {
      return;
    }
    setProviderID(providers.find((provider) => provider.is_default)?.id || providers[0].id);
  }, [providerID, providers]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setGenerating(true);
    setError("");
    const payload: AITestPlanInput = {
      provider_id: providerID || undefined,
      run_id: runID || undefined,
      product_context: productContext.trim() || undefined,
      focus_areas: focusAreas,
      max_scenarios: maxScenarios
    };
    try {
      const plan = await generateAITestPlan(project.id, payload);
      await onGenerated(plan);
    } catch (generateError) {
      setError(generateError instanceof Error ? generateError.message : String(generateError));
    } finally {
      setGenerating(false);
    }
  }

  function toggleFocusArea(value: string) {
    setFocusAreas((current) => {
      if (current.includes(value)) {
        const next = current.filter((item) => item !== value);
        return next.length > 0 ? next : current;
      }
      return [...current, value];
    });
  }

  if (providers.length === 0) {
    return <Notice tone="info" message="Configure an AI provider to generate test plans. Deterministic QA runs still work without AI." />;
  }

  return (
    <form className="project-form test-plan-form" onSubmit={(event) => void submit(event)}>
      {error && <Notice tone="danger" message={error} />}
      <div className="form-grid two">
        <label>
          Provider
          <select value={providerID} onChange={(event) => setProviderID(event.target.value)} required>
            {providers.map((provider) => (
              <option key={provider.id} value={provider.id}>
                {provider.name} ({provider.model})
              </option>
            ))}
          </select>
        </label>
        <label>
          Run Context
          <select value={runID} onChange={(event) => setRunID(event.target.value)}>
            <option value="">Latest run when available</option>
            {runs.map((run) => (
              <option key={run.id} value={run.id}>
                {shortID(run.id)} · {run.status} · {formatDate(run.created_at)}
              </option>
            ))}
          </select>
        </label>
      </div>
      <label>
        Product Context
        <textarea
          value={productContext}
          placeholder="Optional product behavior, user journeys, or areas to emphasize. Do not include secrets."
          onChange={(event) => setProductContext(event.target.value)}
        />
      </label>
      <div>
        <p className="field-label">Focus Areas</p>
        <div className="checkbox-grid">
          {focusAreaOptions.map((option) => (
            <label key={option.value} className="check-row">
              <input type="checkbox" checked={focusAreas.includes(option.value)} onChange={() => toggleFocusArea(option.value)} />
              {option.label}
            </label>
          ))}
        </div>
      </div>
      <div className="form-grid two">
        <label>
          Max Scenarios
          <input
            type="number"
            min="1"
            max="30"
            value={maxScenarios}
            onChange={(event) => setMaxScenarios(Number(event.target.value))}
          />
        </label>
        <Field label="Project Targets" value={targetSummary(project)} />
      </div>
      <div className="form-actions">
        <button type="submit" disabled={generating || focusAreas.length === 0}>
          {generating ? "Generating" : "Generate AI Test Plan"}
        </button>
      </div>
    </form>
  );
}

function TestPlansPage({ projects }: { projects: LoadState<Project[]> }) {
  const [testPlans, setTestPlans] = useState<LoadState<TestPlan[]>>({ data: [], loading: true, error: "" });
  const projectsByID = useMemo(() => new Map(projects.data.map((project) => [project.id, project])), [projects.data]);

  const refresh = useCallback(async () => {
    if (projects.loading) {
      return;
    }
    setTestPlans((current) => ({ ...current, loading: true, error: "" }));
    try {
      const plansByProject = await Promise.all(projects.data.map((project) => listTestPlans(project.id)));
      const plans = plansByProject.flat().sort((left, right) => Date.parse(right.created_at) - Date.parse(left.created_at));
      setTestPlans({ data: plans, loading: false, error: "" });
    } catch (loadError) {
      setTestPlans({ data: [], loading: false, error: loadError instanceof Error ? loadError.message : String(loadError) });
    }
  }, [projects.data, projects.loading]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  if (projects.error) {
    return <Notice tone="danger" message={projects.error} />;
  }

  return (
    <section>
      <div className="section-heading">
        <div>
          <h2>AI Test Plans</h2>
          <p>Review generated plans, scenarios, and exports across projects.</p>
        </div>
        <button type="button" className="secondary" onClick={() => void refresh()}>
          Refresh
        </button>
      </div>
      {testPlans.error && <Notice tone="danger" message={testPlans.error} />}
      {projects.loading || testPlans.loading ? (
        <SkeletonRows />
      ) : (
        <TestPlanTable testPlans={testPlans.data} projectsByID={projectsByID} onDeleted={() => void refresh()} />
      )}
    </section>
  );
}

function TestPlanTable({
  testPlans,
  projectsByID,
  onDeleted
}: {
  testPlans: TestPlan[];
  projectsByID: Map<string, Project>;
  onDeleted: () => void;
}) {
  if (testPlans.length === 0) {
    return <EmptyState title="No test plans" body="Generate a plan from a project once an AI provider is configured." />;
  }

  async function removePlan(testPlanID: string) {
    if (!window.confirm("Delete this AI test plan?")) {
      return;
    }
    await deleteTestPlan(testPlanID);
    onDeleted();
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Title</th>
            <th>Project</th>
            <th>Risk</th>
            <th>Scenarios</th>
            <th>Run</th>
            <th>Created</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {testPlans.map((plan) => (
            <tr key={plan.id}>
              <td>
                <StatusBadge status={plan.status} />
              </td>
              <td>
                <a href={`#/test-plans/${plan.id}`}>{plan.title || shortID(plan.id)}</a>
                {plan.error_message && <p className="muted">{plan.error_message}</p>}
              </td>
              <td>{projectsByID.get(plan.project_id)?.name || plan.project_id}</td>
              <td>{plan.risk_level ? <span className={`severity ${plan.risk_level}`}>{plan.risk_level}</span> : "Not set"}</td>
              <td>{plan.total_scenarios}</td>
              <td>{plan.run_id ? <a href={`#/runs/${plan.run_id}`}>{shortID(plan.run_id)}</a> : "Latest/project only"}</td>
              <td>{formatDate(plan.created_at)}</td>
              <td className="actions">
                <div className="button-row compact">
                  <a className="button secondary-link" href={testPlanExportURL(plan.id)} target="_blank" rel="noreferrer">
                    Export
                  </a>
                  <button type="button" className="secondary danger" onClick={() => void removePlan(plan.id)}>
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function TestPlanDetailPage({ testPlanID, projectByID }: { testPlanID: string; projectByID: Map<string, Project> }) {
  const [plan, setPlan] = useState<TestPlan | undefined>();
  const [project, setProject] = useState<Project | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextPlan = await getTestPlan(testPlanID);
      const nextProject = projectByID.get(nextPlan.project_id) ?? (await getProject(nextPlan.project_id));
      setPlan(nextPlan);
      setProject(nextProject);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [projectByID, testPlanID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading || !plan || !project) {
    return <SkeletonRows />;
  }

  const payload = normalizeTestPlanPayload(plan.plan_json);

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{plan.title || payload.title || "AI Test Plan"}</h2>
            <p>{plan.summary || payload.summary || "No summary provided."}</p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/projects/${plan.project_id}`}>
              Project
            </a>
            {plan.run_id && (
              <a className="button secondary-link" href={`#/runs/${plan.run_id}`}>
                Run
              </a>
            )}
            <a className="button" href={testPlanExportURL(plan.id)} target="_blank" rel="noreferrer">
              Export JSON
            </a>
          </div>
        </div>
        <div className="detail-grid compact">
          <Field label="Status" value={plan.status} />
          <Field label="Project" value={project.name} />
          <Field label="Risk Level" value={plan.risk_level || "Not set"} />
          <Field label="Scenarios" value={String(plan.total_scenarios || payload.scenarios.length)} />
          <Field label="Provider" value={plan.provider_name || plan.provider_id || "Not available"} />
          <Field label="Model" value={plan.model || "Not available"} />
          <Field label="Created" value={formatDate(plan.created_at)} />
          <Field label="Updated" value={formatDate(plan.updated_at)} />
        </div>
        {plan.error_message && <Notice tone="danger" message={plan.error_message} />}
      </section>

      <section>
        <h2>Coverage</h2>
        <div className="analysis-grid">
          <AnalysisList title="Assumptions" items={payload.assumptions} />
          <AnalysisList title="Coverage Goals" items={payload.coverage_goals} />
          <AnalysisList title="Next Instrumentation" items={payload.suggested_next_instrumentation} />
          <AnalysisList title="Limitations" items={payload.limitations} />
        </div>
      </section>

      <section>
        <h2>Scenarios</h2>
        {payload.scenarios.length === 0 ? <p className="muted">No scenarios were returned.</p> : <ScenarioList scenarios={payload.scenarios} />}
      </section>
    </div>
  );
}

function ScenarioList({ scenarios }: { scenarios: TestPlanScenario[] }) {
  return (
    <div className="scenario-stack">
      {scenarios.map((scenario) => (
        <div key={scenario.id} className="scenario-card">
          <div className="scenario-heading">
            <div>
              <h3>{scenario.name}</h3>
              <p>{scenario.description}</p>
            </div>
            <div className="scenario-badges">
              <span className="pill">{scenario.type}</span>
              <span className={`severity ${scenario.priority}`}>priority {scenario.priority}</span>
              <span className={`severity ${scenario.risk}`}>risk {scenario.risk}</span>
            </div>
          </div>
          <div className="detail-grid compact">
            <Field label="Automation Candidate" value={scenario.automation_candidate ? "Yes" : "No"} />
            <Field label="Destructive" value={scenario.destructive ? "Yes" : "No"} />
            <Field label="Requires Auth" value={scenario.requires_authentication ? "Yes" : "No"} />
            <Field label="Tags" value={scenario.tags.length ? scenario.tags.join(", ") : "None"} />
          </div>
          <div className="scenario-columns">
            <AnalysisList title="Preconditions" items={scenario.preconditions} />
            <AnalysisList title="Assertions" items={scenario.assertions} />
            <AnalysisList title="Test Data" items={scenario.test_data_needed} />
            <AnalysisList title="Related Findings" items={scenario.related_findings} />
          </div>
          <div>
            <h3>Steps</h3>
            <ol className="step-list">
              {scenario.steps.map((step) => (
                <li key={`${scenario.id}-${step.order}`}>
                  <strong>{step.action}</strong>
                  <span>{step.target || "Target not specified"}</span>
                  {step.data && <small>Data: {step.data}</small>}
                  <small>Expected: {step.expected_result}</small>
                </li>
              ))}
            </ol>
          </div>
        </div>
      ))}
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
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [selectedProviderID, setSelectedProviderID] = useState("");
  const [analyzing, setAnalyzing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [analysisError, setAnalysisError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextRun = await getRun(runID);
      const [nextReport, nextProject, nextProviders] = await Promise.all([
        getReport(runID),
        projectByID.get(nextRun.project_id) ?? getProject(nextRun.project_id),
        listAIProviders()
      ]);
      setRun(nextRun);
      setReport(nextReport);
      setProject(nextProject);
      setProviders(nextProviders);
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
    if (!run || !isActiveRunStatus(run.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, run]);

  useEffect(() => {
    if (selectedProviderID || providers.length === 0) {
      return;
    }
    setSelectedProviderID(providers.find((provider) => provider.is_default)?.id || providers[0].id);
  }, [providers, selectedProviderID]);

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
  const relatedTestPlans = report.test_plans || [];

  async function runAnalysis() {
    if (!run) {
      return;
    }
    setAnalyzing(true);
    setAnalysisError("");
    try {
      await runAIAnalysis(run.id, selectedProviderID || undefined);
      await refresh();
    } catch (analysisRunError) {
      setAnalysisError(analysisRunError instanceof Error ? analysisRunError.message : String(analysisRunError));
    } finally {
      setAnalyzing(false);
    }
  }

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
        <div className="section-heading">
          <div>
            <h2>AI Analysis</h2>
            <p>Optional analysis of the deterministic findings and evidence metadata.</p>
          </div>
          {providers.length > 0 && (
            <div className="button-row">
              {providers.length > 1 && (
                <select value={selectedProviderID} onChange={(event) => setSelectedProviderID(event.target.value)}>
                  {providers.map((provider) => (
                    <option key={provider.id} value={provider.id}>
                      {provider.name} ({provider.model})
                    </option>
                  ))}
                </select>
              )}
              <button type="button" disabled={analyzing || run.status !== "completed"} onClick={() => void runAnalysis()}>
                {analyzing ? "Analyzing" : "Run AI analysis"}
              </button>
            </div>
          )}
        </div>
        {providers.length === 0 && <Notice tone="info" message="AI analysis is optional. Configure an AI provider to enable it." />}
        {analysisError && <Notice tone="danger" message={analysisError} />}
        {providers.length > 0 && run.status !== "completed" && (
          <p className="muted">AI analysis can run after the QA run completes.</p>
        )}
        <AIAnalysisPanel analysis={report.ai_analysis} />
      </section>

      {relatedTestPlans.length > 0 && (
        <section>
          <h2>Related AI Test Plans</h2>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Title</th>
                  <th>Risk</th>
                  <th>Scenarios</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {relatedTestPlans.map((plan) => (
                  <tr key={plan.id}>
                    <td>
                      <StatusBadge status={plan.status} />
                    </td>
                    <td>
                      <a href={`#/test-plans/${plan.id}`}>{plan.title || shortID(plan.id)}</a>
                    </td>
                    <td>{plan.risk_level ? <span className={`severity ${plan.risk_level}`}>{plan.risk_level}</span> : "Not set"}</td>
                    <td>{plan.total_scenarios}</td>
                    <td>{formatDate(plan.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

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

function AIAnalysisPanel({ analysis }: { analysis: AIAnalysis | null }) {
  if (!analysis) {
    return <p className="muted">No AI analysis has been generated for this run.</p>;
  }

  const likelyCauses = analysisStringList(analysis.analysis_json.likely_causes);
  const recommendedActions = analysisStringList(analysis.analysis_json.recommended_actions);
  const suggestedNextTests = analysisStringList(analysis.analysis_json.suggested_next_tests);
  const limitations = analysisStringList(analysis.analysis_json.limitations);
  const confidence = typeof analysis.analysis_json.confidence === "number" ? analysis.analysis_json.confidence : undefined;

  return (
    <div className="analysis-panel">
      <div className="detail-grid compact">
        <Field label="Status" value={analysis.status} />
        <Field label="Risk Level" value={analysis.risk_level || "Not set"} />
        <Field label="Provider" value={analysis.provider_name || analysis.provider_id || "Not available"} />
        <Field label="Model" value={analysis.model || "Not available"} />
        <Field label="Tokens" value={String(analysis.total_tokens || 0)} />
        <Field label="Confidence" value={confidence === undefined ? "Not provided" : confidence.toFixed(2)} />
      </div>
      {analysis.error_message && <Notice tone="danger" message={analysis.error_message} />}
      {analysis.executive_summary && (
        <div>
          <h3>Executive Summary</h3>
          <p>{analysis.executive_summary}</p>
        </div>
      )}
      {analysis.technical_summary && (
        <div>
          <h3>Technical Summary</h3>
          <p>{analysis.technical_summary}</p>
        </div>
      )}
      <div className="analysis-grid">
        <AnalysisList title="Likely Causes" items={likelyCauses} />
        <AnalysisList title="Recommended Actions" items={recommendedActions} />
        <AnalysisList title="Suggested Next Tests" items={suggestedNextTests} />
        <AnalysisList title="Limitations" items={limitations} />
      </div>
    </div>
  );
}

function AnalysisList({ title, items }: { title: string; items: string[] }) {
  return (
    <div>
      <h3>{title}</h3>
      {items.length === 0 ? (
        <p className="muted">Not provided.</p>
      ) : (
        <ul className="plain-list">
          {items.map((item, index) => (
            <li key={`${title}-${index}`}>{item}</li>
          ))}
        </ul>
      )}
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
            <th>Object</th>
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
                {item.type === "screenshot" ? (
                  <div className="evidence-object">
                    <a className="button secondary-link" href={evidenceDownloadURL(item.id)} target="_blank" rel="noreferrer">
                      Download
                    </a>
                    <img className="evidence-preview" src={evidenceDownloadURL(item.id)} alt={`Screenshot evidence ${shortID(item.id)}`} />
                  </div>
                ) : (
                  <span className="muted">Inline metadata</span>
                )}
              </td>
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
  if (parts[0] === "ai-providers") {
    return { name: "ai-providers" };
  }
  if (parts[0] === "test-plans" && parts[1]) {
    return { name: "test-plan", id: parts[1] };
  }
  if (parts[0] === "test-plans") {
    return { name: "test-plans" };
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
    case "ai-providers":
      return "/ai-providers";
    case "test-plans":
      return "/test-plans";
    case "test-plan":
      return `/test-plans/${route.id}`;
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
    case "ai-providers":
      return "AI Providers";
    case "test-plans":
      return "AI Test Plans";
    case "test-plan":
      return "Test Plan";
    case "run":
      return "Run Report";
  }
}

function providerFormDefaults(preset: string): ProviderFormState {
  const base: ProviderFormState = {
    id: "",
    name: "OpenAI",
    preset,
    type: "openai-compatible",
    base_url: "https://api.openai.com/v1",
    model: "gpt-4o-mini",
    api_key: "",
    extra_headers: "",
    temperature: 0.2,
    max_output_tokens: 1200,
    timeout_seconds: 30,
    send_screenshots: false,
    send_html: false,
    send_network_bodies: false,
    redaction_enabled: true,
    is_default: false
  };
  if (preset === "openrouter") {
    return {
      ...base,
      name: "OpenRouter",
      base_url: "https://openrouter.ai/api/v1",
      model: "openai/gpt-4o-mini",
      extra_headers: JSON.stringify({ "X-OpenRouter-Title": "Qualora" }, null, 2)
    };
  }
  if (preset === "ollama") {
    return { ...base, name: "Ollama", base_url: "http://ollama:11434/v1", model: "qwen2.5-coder:7b", timeout_seconds: 60 };
  }
  if (preset === "custom") {
    return { ...base, name: "Custom OpenAI-compatible", base_url: "", model: "" };
  }
  if (preset === "disabled") {
    return { ...base, name: "Disabled", base_url: "", model: "" };
  }
  return base;
}

function inputForProviderForm(form: ProviderFormState, preserveBlankSecrets: boolean): AIProviderInput {
  const input: AIProviderInput = {
    name: form.name.trim(),
    preset: form.preset,
    type: "openai-compatible",
    base_url: form.base_url.trim(),
    model: form.model.trim(),
    temperature: form.temperature,
    max_output_tokens: form.max_output_tokens,
    timeout_seconds: form.timeout_seconds,
    send_screenshots: form.send_screenshots,
    send_html: form.send_html,
    send_network_bodies: form.send_network_bodies,
    redaction_enabled: form.redaction_enabled,
    is_default: form.is_default
  };
  const apiKey = form.api_key.trim();
  if (apiKey) {
    input.api_key = apiKey;
  }
  const headers = form.extra_headers.trim();
  if (headers) {
    const parsed = JSON.parse(headers);
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      throw new Error("extra headers must be a JSON object");
    }
    input.extra_headers = Object.fromEntries(Object.entries(parsed).map(([key, value]) => [key, String(value)]));
  } else if (!preserveBlankSecrets) {
    input.extra_headers = {};
  }
  return input;
}

function analysisStringList(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.map((item) => String(item)).filter(Boolean);
}

function normalizeTestPlanPayload(value: Partial<TestPlanPayload> | Record<string, unknown> | undefined): TestPlanPayload {
  const payload = value || {};
  return {
    title: typeof payload.title === "string" ? payload.title : "",
    summary: typeof payload.summary === "string" ? payload.summary : "",
    assumptions: analysisStringList(payload.assumptions),
    coverage_goals: analysisStringList(payload.coverage_goals),
    scenarios: normalizeScenarios(payload.scenarios),
    suggested_next_instrumentation: analysisStringList(payload.suggested_next_instrumentation),
    limitations: analysisStringList(payload.limitations)
  };
}

function normalizeScenarios(value: unknown): TestPlanScenario[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === "object")
    .map((item, index) => ({
      id: typeof item.id === "string" && item.id ? item.id : `scenario-${index + 1}`,
      name: typeof item.name === "string" ? item.name : `Scenario ${index + 1}`,
      type: typeof item.type === "string" ? item.type : "functional",
      priority: riskValue(item.priority),
      risk: riskValue(item.risk),
      description: typeof item.description === "string" ? item.description : "",
      preconditions: analysisStringList(item.preconditions),
      steps: normalizeSteps(item.steps),
      assertions: analysisStringList(item.assertions),
      test_data_needed: analysisStringList(item.test_data_needed),
      automation_candidate: Boolean(item.automation_candidate),
      destructive: Boolean(item.destructive),
      requires_authentication: Boolean(item.requires_authentication),
      related_findings: analysisStringList(item.related_findings),
      tags: analysisStringList(item.tags)
    }));
}

function normalizeSteps(value: unknown) {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === "object")
    .map((item, index) => ({
      order: typeof item.order === "number" ? item.order : index + 1,
      action: typeof item.action === "string" ? item.action : "",
      target: typeof item.target === "string" ? item.target : "",
      data: typeof item.data === "string" ? item.data : "",
      expected_result: typeof item.expected_result === "string" ? item.expected_result : ""
    }));
}

function riskValue(value: unknown): "low" | "medium" | "high" | "critical" {
  return value === "low" || value === "medium" || value === "high" || value === "critical" ? value : "medium";
}

function splitHosts(input: string): string[] {
  return input
    .split(/[,\n]/)
    .map((host) => host.trim())
    .filter(Boolean);
}

function isActiveRunStatus(status: string): boolean {
  return status === "queued" || status === "pending" || status === "running";
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

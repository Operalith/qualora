import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  API_BASE_URL,
  authorizationCheckHTMLReportURL,
  createAuthorizationCheck,
  createCredentialProfile,
  createAIProvider,
  createProject,
  deleteAIProvider,
  deleteAuthorizationCheck,
  deleteAPISpec,
  deleteCredentialProfile,
  deleteTestPlan,
  discoveryHTMLReportURL,
  evidenceDownloadURL,
  executeTestPlan,
  executeQARun,
  generateAITestPlan,
  getAuthorizationCheckReport,
  getAPISpec,
  getDiscoveryReport,
  getProject,
  getQARunReport,
  getQualityCheckReport,
  getReport,
  getRun,
  getSafeExplorerReport,
  getSetupStatus,
  getTestPlan,
  getTestPlanExecutionReport,
  htmlReportURL,
  importAPISpec,
  listAIProviders,
  listAuthorizationCheckRuns,
  listAuthorizationChecks,
  listAPISpecs,
  listCredentialProfiles,
  listDiscoveryRuns,
  listProjects,
  listQARuns,
  listQualityCheckRuns,
  listSafeExplorerRuns,
  listRuns,
  listTestPlanExecutions,
  listTestPlans,
  login as loginUser,
  logout,
  me,
  previewTestPlanExecution,
  qualityCheckHTMLReportURL,
  qaRunHTMLReportURL,
  runAIAnalysis,
  runProjectSetup,
  safeExplorerHTMLReportURL,
  setupAdmin,
  startAuthorizationCheckRun,
  startAPISmokeRun,
  startAuthenticatedBrowserSmokeRun,
  startBrowserSmokeRun,
  startDiscoveryRun,
  startQARun,
  startQualityCheckRun,
  startSafeExplorerRun,
  startRun,
  testCredentialProfileLogin,
  testPlanExportURL,
  testPlanExecutionHTMLReportURL,
  testAIProvider,
  updateAuthorizationCheck,
  updateCredentialProfile,
  updateAIProvider
} from "./api";
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
  AuthUser,
  AuthorizationCheck,
  AuthorizationCheckInput,
  AuthorizationCheckReport,
  AuthorizationCheckRun,
  CreateProjectInput,
  CredentialProfile,
  CredentialProfileInput,
  DiscoveryReport,
  DiscoveryRun,
  DiscoveryRunInput,
  Evidence,
  LoginInput,
  Project,
  ProjectSetupInput,
  ProjectSetupResponse,
  QARun,
  QARunInput,
  QARunReport,
  QualityCheckReport,
  QualityCheckRun,
  QualityCheckRunInput,
  QualityCheckResult,
  SafeExplorerAction,
  SafeExplorerReport,
  SafeExplorerRun,
  SafeExplorerRunInput,
  Report,
  RunJob,
  SetupAdminInput,
  TestPlan,
  TestPlanExecution,
  TestPlanExecutionPreview,
  TestPlanExecutionReport,
  TestPlanExecutionRequest,
  TestPlanPayload,
  TestPlanScenario,
  TestRun
} from "./types";

type Route =
  | { name: "dashboard" }
  | { name: "new-project" }
  | { name: "setup-project" }
  | { name: "reports" }
  | { name: "project"; id: string }
  | { name: "runs" }
  | { name: "ai-providers" }
  | { name: "api-spec"; id: string }
  | { name: "test-plans" }
  | { name: "test-plan"; id: string }
  | { name: "test-plan-execution"; id: string }
  | { name: "authorization-check-run"; id: string }
  | { name: "discovery-run"; id: string }
  | { name: "quality-check-run"; id: string }
  | { name: "safe-explorer-run"; id: string }
  | { name: "qa-run"; id: string }
  | { name: "run"; id: string };

type LoadState<T> = {
  data: T;
  loading: boolean;
  error: string;
};

const emptyProjects: LoadState<Project[]> = { data: [], loading: true, error: "" };
const emptyRuns: LoadState<TestRun[]> = { data: [], loading: true, error: "" };

export default function App() {
  const [auth, setAuth] = useState<{
    loading: boolean;
    setupRequired: boolean;
    user: AuthUser | null;
    version: string;
    error: string;
  }>({ loading: true, setupRequired: false, user: null, version: "0.16.0-alpha", error: "" });

  const loadAuthState = useCallback(async () => {
    setAuth((current) => ({ ...current, loading: true, error: "" }));
    try {
      const setup = await getSetupStatus();
      if (setup.setup_required) {
        setAuth({ loading: false, setupRequired: true, user: null, version: setup.version, error: "" });
        return;
      }
      const current = await me();
      setAuth({
        loading: false,
        setupRequired: false,
        user: current.authenticated && current.user ? current.user : null,
        version: setup.version,
        error: ""
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      setAuth((current) => ({ ...current, loading: false, user: null, error: message }));
    }
  }, []);

  useEffect(() => {
    void loadAuthState();
  }, [loadAuthState]);

  useEffect(() => {
    const listener = () => {
      setAuth((current) => ({ ...current, user: null, setupRequired: false, error: "Your session expired. Please log in again." }));
    };
    window.addEventListener("qualora:unauthorized", listener);
    return () => window.removeEventListener("qualora:unauthorized", listener);
  }, []);

  const completeSetup = async (input: SetupAdminInput) => {
    const response = await setupAdmin(input);
    setAuth({ loading: false, setupRequired: false, user: response.user, version: auth.version, error: "" });
  };

  const completeLogin = async (input: LoginInput) => {
    const response = await loginUser(input);
    setAuth({ loading: false, setupRequired: false, user: response.user, version: auth.version, error: "" });
  };

  const completeLogout = async () => {
    await logout();
    setAuth((current) => ({ ...current, user: null, setupRequired: false, error: "" }));
    window.location.hash = "#/";
  };

  if (auth.loading) {
    return <AuthFrame version={auth.version} title="Loading Qualora" subtitle="Checking local authentication state." />;
  }
  if (auth.setupRequired) {
    return <SetupPage version={auth.version} error={auth.error} onSubmit={completeSetup} />;
  }
  if (!auth.user) {
    return <LoginPage version={auth.version} message={auth.error} onSubmit={completeLogin} />;
  }

  return <AuthenticatedApp user={auth.user} version={auth.version} onLogout={completeLogout} />;
}

function AuthenticatedApp({ user, version, onLogout }: { user: AuthUser; version: string; onLogout: () => Promise<void> }) {
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
  const startDemoWorkflow = async () => {
    const response = await runProjectSetup(demoProjectSetupPayload());
    await refresh();
    if (response.started.safe_qa_run_id) {
      setRoute({ name: "qa-run", id: response.started.safe_qa_run_id });
    } else {
      setRoute({ name: "project", id: response.project.id });
    }
  };

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <a className="brand" href="#/">
          <span>Qualora</span>
          <small>{formatVersionBadge(version)}</small>
        </a>
        <nav>
          <a className={route.name === "dashboard" ? "active" : ""} href="#/">
            Dashboard
          </a>
          <a className={route.name === "setup-project" ? "active" : ""} href="#/setup-project">
            Guided Setup
          </a>
          <a className={route.name === "runs" ? "active" : ""} href="#/runs">
            Browser Testing
          </a>
          <a className={route.name === "reports" ? "active" : ""} href="#/reports">
            Reports
          </a>
          <a className={route.name === "new-project" ? "active" : ""} href="#/projects/new">
            Projects
          </a>
          <a className={route.name === "ai-providers" ? "active" : ""} href="#/ai-providers">
            AI Providers
          </a>
          <a
            className={route.name === "test-plans" || route.name === "test-plan" || route.name === "test-plan-execution" ? "active" : ""}
            href="#/test-plans"
          >
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
          <div className="topbar-actions">
            <div className="current-user">
              <span>{user.display_name}</span>
              <small>{user.email}</small>
            </div>
            <button type="button" className="secondary" onClick={() => void refresh()}>
              Refresh
            </button>
            <button type="button" className="secondary" onClick={() => void onLogout()}>
              Log out
            </button>
          </div>
        </header>

        {(projects.error || runs.error) && <Notice tone="danger" message={projects.error || runs.error} />}

        {route.name === "dashboard" && (
          <Dashboard
            version={version}
            projects={projects}
            runs={runs}
            projectByID={projectByID}
            onStartRun={startFullRun}
            onRunDemoWorkflow={startDemoWorkflow}
          />
        )}
        {route.name === "setup-project" && (
          <ProjectSetupWizard
            onCompleted={async (response) => {
              await refresh();
              setRoute({ name: "setup-project" });
            }}
          />
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
        {route.name === "api-spec" && (
          <APISpecPage apiSpecID={route.id} projectByID={projectByID} onOpenRun={(id) => setRoute({ name: "run", id })} />
        )}
        {route.name === "runs" && <RunsPage runs={runs} projectByID={projectByID} />}
        {route.name === "reports" && <ReportsPage projects={projects} runs={runs} projectByID={projectByID} />}
        {route.name === "ai-providers" && <AIProvidersPage />}
        {route.name === "test-plans" && <TestPlansPage projects={projects} />}
        {route.name === "test-plan" && (
          <TestPlanDetailPage
            testPlanID={route.id}
            projectByID={projectByID}
            onOpenExecution={(id) => setRoute({ name: "test-plan-execution", id })}
          />
        )}
        {route.name === "test-plan-execution" && <TestPlanExecutionPage executionID={route.id} />}
        {route.name === "authorization-check-run" && <AuthorizationCheckRunPage runID={route.id} />}
        {route.name === "discovery-run" && <DiscoveryRunPage runID={route.id} />}
        {route.name === "quality-check-run" && <QualityCheckRunPage runID={route.id} />}
        {route.name === "safe-explorer-run" && <SafeExplorerRunPage runID={route.id} />}
        {route.name === "qa-run" && <QARunPage runID={route.id} />}
        {route.name === "run" && <RunReportPage runID={route.id} cachedRun={runs.data.find((run) => run.id === route.id)} projectByID={projectByID} />}
      </main>
    </div>
  );
}

function AuthFrame({ version, title, subtitle, children }: { version: string; title: string; subtitle: string; children?: ReactNode }) {
  return (
    <main className="auth-shell">
      <section className="auth-panel">
        <a className="auth-brand" href="#/">
          <span>Qualora</span>
          <small>{formatVersionBadge(version)}</small>
        </a>
        <h1>{title}</h1>
        <p className="muted">{subtitle}</p>
        {children}
      </section>
    </main>
  );
}

function SetupPage({ version, error, onSubmit }: { version: string; error: string; onSubmit: (input: SetupAdminInput) => Promise<void> }) {
  const [form, setForm] = useState<SetupAdminInput>({
    display_name: "Qualora Admin",
    email: "admin@qualora.local",
    password: "",
    confirm_password: ""
  });
  const [message, setMessage] = useState(error);
  const [submitting, setSubmitting] = useState(false);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setMessage("");
    try {
      await onSubmit(form);
    } catch (caught) {
      setMessage(caught instanceof Error ? caught.message : String(caught));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AuthFrame version={version} title="Create Admin" subtitle="Complete first-run setup for this self-hosted Qualora instance.">
      {message && <Notice tone="danger" message={message} />}
      <form className="auth-form" onSubmit={(event) => void submit(event)}>
        <label>
          Display name
          <input value={form.display_name} onChange={(event) => setForm({ ...form, display_name: event.target.value })} required />
        </label>
        <label>
          Email
          <input type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} required />
        </label>
        <label>
          Password
          <input type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} minLength={12} required />
        </label>
        <label>
          Confirm password
          <input
            type="password"
            value={form.confirm_password}
            onChange={(event) => setForm({ ...form, confirm_password: event.target.value })}
            minLength={12}
            required
          />
        </label>
        <button type="submit" disabled={submitting}>
          {submitting ? "Creating" : "Create admin"}
        </button>
      </form>
    </AuthFrame>
  );
}

function LoginPage({ version, message, onSubmit }: { version: string; message: string; onSubmit: (input: LoginInput) => Promise<void> }) {
  const [form, setForm] = useState<LoginInput>({ email: "admin@qualora.local", password: "" });
  const [error, setError] = useState(message);
  const [submitting, setSubmitting] = useState(false);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await onSubmit(form);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : String(caught));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AuthFrame version={version} title="Log In" subtitle="Use the local admin account for this Qualora instance.">
      {error && <Notice tone="danger" message={error} />}
      <form className="auth-form" onSubmit={(event) => void submit(event)}>
        <label>
          Email
          <input type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} required />
        </label>
        <label>
          Password
          <input type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} required />
        </label>
        <button type="submit" disabled={submitting}>
          {submitting ? "Logging in" : "Log in"}
        </button>
      </form>
    </AuthFrame>
  );
}

function formatVersionBadge(version: string): string {
  const normalized = version.trim().replace(/^v/i, "");
  return `Qualora v${normalized}`;
}

function Dashboard({
  version,
  projects,
  runs,
  projectByID,
  onStartRun,
  onRunDemoWorkflow
}: {
  version: string;
  projects: LoadState<Project[]>;
  runs: LoadState<TestRun[]>;
  projectByID: Map<string, Project>;
  onStartRun: (projectID: string) => Promise<void>;
  onRunDemoWorkflow: () => Promise<void>;
}) {
  const [providers, setProviders] = useState<LoadState<AIProvider[]>>({ data: [], loading: true, error: "" });
  const [qaRuns, setQARuns] = useState<LoadState<QARun[]>>({ data: [], loading: true, error: "" });
  const [demoBusy, setDemoBusy] = useState(false);

  useEffect(() => {
    let canceled = false;
    async function loadDashboardDetails() {
      setProviders((current) => ({ ...current, loading: true, error: "" }));
      setQARuns((current) => ({ ...current, loading: true, error: "" }));
      try {
        const nextProviders = await listAIProviders();
        const nextQARuns = projects.data.length > 0 ? (await Promise.all(projects.data.map((project) => listQARuns(project.id)))).flat() : [];
        nextQARuns.sort((left, right) => Date.parse(right.created_at) - Date.parse(left.created_at));
        if (!canceled) {
          setProviders({ data: nextProviders, loading: false, error: "" });
          setQARuns({ data: nextQARuns, loading: false, error: "" });
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        if (!canceled) {
          setProviders((current) => ({ ...current, loading: false, error: message }));
          setQARuns((current) => ({ ...current, loading: false, error: message }));
        }
      }
    }
    if (!projects.loading) {
      void loadDashboardDetails();
    }
    return () => {
      canceled = true;
    };
  }, [projects.data, projects.loading]);

  const latestRun = runs.data[0];
  const totalFindings = runs.data.reduce((total, run) => total + (run.status === "failed" ? 1 : 0), 0);
  async function runDemo() {
    setDemoBusy(true);
    try {
      await onRunDemoWorkflow();
    } finally {
      setDemoBusy(false);
    }
  }

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <p className="eyebrow">Guided first run</p>
            <h2>Qualora version badge: {formatVersionBadge(version)}</h2>
            <p>Configure a real project, optional AI, optional login, optional OpenAPI, and start the first safe workflow.</p>
          </div>
          <span className="pill">v0.16 safe explorer</span>
        </div>
        <div className="quick-grid">
          <a className="quick-card" href="#/setup-project">
            <strong>Create project with guided setup</strong>
            <span>Use the wizard for target URLs, AI, credentials, OpenAPI, and first checks.</span>
          </a>
          <button type="button" className="quick-card button-reset" disabled={demoBusy} onClick={() => void runDemo()}>
            <strong>{demoBusy ? "Starting demo workflow" : "Run demo workflow"}</strong>
            <span>Use demo-web, demo-api, and fake-llm for deterministic local verification.</span>
          </button>
        </div>
      </section>

      <section>
        <div className="section-heading">
          <div>
            <h2>Status</h2>
            <p>Local self-hosted health and setup snapshot.</p>
          </div>
        </div>
        <div className="detail-grid compact">
          <Field label="API status" value="Reachable" />
          <Field label="Web status" value="Reachable" />
          <Field label="AI configured" value={providers.loading ? "Checking" : providers.data.length > 0 ? "Yes" : "No"} />
          <Field label="Projects" value={String(projects.data.length)} />
          <Field label="Latest run" value={latestRun ? `${formatRunType(latestRun.run_type)} ${latestRun.status}` : "No runs yet"} />
          <Field label="Recent failed runs" value={String(totalFindings)} />
        </div>
        {!projects.loading && projects.data.length === 0 && <EmptyState title="Start by creating your first project." body="The guided setup is the fastest way to get a useful first report." />}
      </section>

      <section>
        <div className="section-heading">
          <div>
            <h2>Recent Projects</h2>
            <p>Configured frontend and API targets.</p>
          </div>
          <a className="button" href="#/setup-project">
            Guided Setup
          </a>
        </div>
        {projects.loading ? <SkeletonRows /> : <ProjectTable projects={projects.data.slice(0, 6)} onStartRun={onStartRun} />}
      </section>

      <section>
        <div className="section-heading">
          <div>
            <h2>Recent Safe QA Runs</h2>
            <p>Latest guided and discovery-aware QA workflows.</p>
          </div>
          <a className="button secondary-link" href="#/reports">
            Reports
          </a>
        </div>
        {qaRuns.loading ? <SkeletonRows /> : <QARunDashboardTable runs={qaRuns.data.slice(0, 8)} projectByID={projectByID} />}
      </section>

      <section>
        <div className="section-heading">
          <div>
            <h2>Recent Runs</h2>
            <p>Browser, API, login, authenticated, and full run reports.</p>
          </div>
          <a className="button secondary-link" href="#/runs">
            Browser Testing
          </a>
        </div>
        {runs.loading ? <SkeletonRows /> : <RunTable runs={runs.data.slice(0, 8)} projectByID={projectByID} />}
      </section>
    </div>
  );
}

function QARunDashboardTable({ runs, projectByID }: { runs: QARun[]; projectByID: Map<string, Project> }) {
  if (runs.length === 0) {
    return <EmptyState title="No Safe QA runs yet" body="Use guided setup or a project detail page to start a Safe QA preview." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Project</th>
            <th>Run</th>
            <th>Discovery</th>
            <th>Quality</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id}>
              <td><StatusBadge status={run.status} /></td>
              <td>{projectByID.get(run.project_id)?.name || run.project_id}</td>
              <td><a href={`#/qa-runs/${run.id}`}>{shortID(run.id)}</a></td>
              <td>{run.discovery_run_id ? shortID(run.discovery_run_id) : "Not linked"}</td>
              <td>{run.quality_check_run_id ? shortID(run.quality_check_run_id) : "Not linked"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

type WizardWorkflowState = {
  browser_smoke: boolean;
  discovery: boolean;
  quality_checks: boolean;
  safe_qa_run: boolean;
  execute_safe_qa: boolean;
  api_smoke: boolean;
  authenticated_smoke: boolean;
};

function ProjectSetupWizard({ onCompleted }: { onCompleted: (response: ProjectSetupResponse) => Promise<void> }) {
  const [step, setStep] = useState(1);
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [projectForm, setProjectForm] = useState({
    name: "My Web App",
    frontend_url: "",
    api_base_url: "",
    allowed_hosts: "",
    allow_private_targets: false
  });
  const [aiMode, setAIMode] = useState<"skip" | "existing" | "create" | "demo">("skip");
  const [selectedProviderID, setSelectedProviderID] = useState("");
  const [providerForm, setProviderForm] = useState<ProviderFormState>(() => providerFormDefaults("openai"));
  const [credentialMode, setCredentialMode] = useState<"skip" | "create">("skip");
  const [credentialForm, setCredentialForm] = useState<CredentialProfileInput>(() => wizardCredentialDefaults());
  const [apiSpecMode, setAPISpecMode] = useState<"skip" | "url" | "inline" | "demo">("skip");
  const [apiSpecForm, setAPISpecForm] = useState({ name: "OpenAPI Spec", source_url: "", raw_spec: "" });
  const [workflow, setWorkflow] = useState<WizardWorkflowState>({
    browser_smoke: true,
    discovery: true,
    quality_checks: true,
    safe_qa_run: false,
    execute_safe_qa: false,
    api_smoke: false,
    authenticated_smoke: false
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<ProjectSetupResponse | undefined>();

  useEffect(() => {
    let canceled = false;
    async function loadProviders() {
      try {
        const nextProviders = await listAIProviders();
        if (!canceled) {
          setProviders(nextProviders);
          const defaultProvider = nextProviders.find((provider) => provider.is_default) || nextProviders[0];
          if (defaultProvider) {
            setSelectedProviderID(defaultProvider.id);
          }
        }
      } catch {
        if (!canceled) {
          setProviders([]);
        }
      }
    }
    void loadProviders();
    return () => {
      canceled = true;
    };
  }, []);

  useEffect(() => {
    setWorkflow((current) => ({
      ...current,
      safe_qa_run: aiMode !== "skip" && current.safe_qa_run,
      api_smoke: apiSpecMode !== "skip" && current.api_smoke,
      authenticated_smoke: credentialMode === "create" && current.authenticated_smoke
    }));
  }, [aiMode, apiSpecMode, credentialMode]);

  function loadDemoDefaults() {
    setProjectForm({
      name: "Qualora Demo Workflow",
      frontend_url: "http://demo-web:8080",
      api_base_url: "http://demo-api:8080",
      allowed_hosts: "demo-web, demo-api",
      allow_private_targets: true
    });
    setAIMode("demo");
    setCredentialMode("create");
    setCredentialForm({
      ...wizardCredentialDefaults(),
      name: "Qualora Demo Login",
      username: "demo@example.com",
      password: "qualora-demo-password",
      login_url: "http://demo-web:8080/login",
      success_url_contains: "/dashboard",
      success_text_contains: "Welcome to the Qualora demo dashboard",
      failure_text_contains: "Invalid credentials"
    });
    setAPISpecMode("demo");
    setWorkflow({
      browser_smoke: true,
      discovery: true,
      quality_checks: true,
      safe_qa_run: true,
      execute_safe_qa: false,
      api_smoke: true,
      authenticated_smoke: true
    });
    setStep(5);
    setResult(undefined);
    setError("");
  }

  async function submit() {
    setSubmitting(true);
    setError("");
    setResult(undefined);
    try {
      const payload = buildProjectSetupPayload(projectForm, aiMode, selectedProviderID, providerForm, credentialMode, credentialForm, apiSpecMode, apiSpecForm, workflow);
      const response = await runProjectSetup(payload);
      setResult(response);
      setStep(6);
      await onCompleted(response);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : String(caught));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <p className="eyebrow">Guided setup</p>
            <h2>Create a project and run first checks</h2>
            <p>Configure only what you need. AI, login, and OpenAPI are optional.</p>
          </div>
          <button type="button" className="secondary" onClick={loadDemoDefaults}>
            Load demo workflow
          </button>
        </div>
        <div className="wizard-steps">
          {["Basics", "AI", "Login", "OpenAPI", "Workflow", "Results"].map((label, index) => (
            <button key={label} type="button" className={step === index + 1 ? "" : "secondary"} onClick={() => setStep(index + 1)}>
              {index + 1}. {label}
            </button>
          ))}
        </div>
        {error && <Notice tone="danger" message={error} />}
      </section>

      {step === 1 && (
        <section>
          <h2>Project basics</h2>
          <form className="project-form">
            <label>
              Project name
              <input value={projectForm.name} onChange={(event) => setProjectForm({ ...projectForm, name: event.target.value })} required />
            </label>
            <label>
              Frontend/base URL
              <input value={projectForm.frontend_url} onChange={(event) => setProjectForm({ ...projectForm, frontend_url: event.target.value })} placeholder="https://app.example.com" />
            </label>
            <label>
              Optional API base URL
              <input value={projectForm.api_base_url} onChange={(event) => setProjectForm({ ...projectForm, api_base_url: event.target.value })} placeholder="https://api.example.com" />
            </label>
            <label>
              Allowed hosts
              <textarea value={projectForm.allowed_hosts} onChange={(event) => setProjectForm({ ...projectForm, allowed_hosts: event.target.value })} placeholder="app.example.com, api.example.com" />
            </label>
            <div className="detail-grid">
              <Field label="Safety mode" value="Safe default / passive" />
              <Field label="Destructive actions" value="Disabled" />
            </div>
            <label className="check-row">
              <input type="checkbox" checked={projectForm.allow_private_targets} onChange={(event) => setProjectForm({ ...projectForm, allow_private_targets: event.target.checked })} />
              Allow private/local targets for this self-hosted setup
            </label>
          </form>
        </section>
      )}

      {step === 2 && (
        <section>
          <h2>AI provider optional</h2>
          <Notice tone="info" message="AI is optional. Qualora can still run deterministic checks without AI." />
          <div className="toggle-grid">
            <label className="check-row"><input type="radio" checked={aiMode === "skip"} onChange={() => setAIMode("skip")} /> Skip AI</label>
            <label className="check-row"><input type="radio" checked={aiMode === "existing"} onChange={() => setAIMode("existing")} /> Use existing provider</label>
            <label className="check-row"><input type="radio" checked={aiMode === "demo"} onChange={() => setAIMode("demo")} /> Configure fake/demo provider</label>
            <label className="check-row"><input type="radio" checked={aiMode === "create"} onChange={() => setAIMode("create")} /> Configure provider</label>
          </div>
          {aiMode === "existing" && (
            <label>
              Existing provider
              <select value={selectedProviderID} onChange={(event) => setSelectedProviderID(event.target.value)}>
                {providers.map((provider) => <option key={provider.id} value={provider.id}>{provider.name} ({provider.model})</option>)}
              </select>
            </label>
          )}
          {aiMode === "create" && (
            <div className="project-form">
              <label>
                Preset
                <select value={providerForm.preset} onChange={(event) => setProviderForm(providerFormDefaults(event.target.value))}>
                  {providerPresets.filter((preset) => preset.value !== "disabled").map((preset) => <option key={preset.value} value={preset.value}>{preset.label}</option>)}
                </select>
              </label>
              <label>Name<input value={providerForm.name} onChange={(event) => setProviderForm({ ...providerForm, name: event.target.value })} /></label>
              <label>Base URL<input value={providerForm.base_url} onChange={(event) => setProviderForm({ ...providerForm, base_url: event.target.value })} /></label>
              <label>Model<input value={providerForm.model} onChange={(event) => setProviderForm({ ...providerForm, model: event.target.value })} /></label>
              <label>API Key<input type="password" value={providerForm.api_key} onChange={(event) => setProviderForm({ ...providerForm, api_key: event.target.value })} /></label>
            </div>
          )}
        </section>
      )}

      {step === 3 && (
        <section>
          <h2>Login optional</h2>
          <Notice tone="info" message="Credentials are encrypted and never sent to AI. Use dedicated test accounts." />
          <div className="toggle-grid">
            <label className="check-row"><input type="radio" checked={credentialMode === "skip"} onChange={() => setCredentialMode("skip")} /> Skip login</label>
            <label className="check-row"><input type="radio" checked={credentialMode === "create"} onChange={() => setCredentialMode("create")} /> Add credential profile</label>
          </div>
          {credentialMode === "create" && (
            <form className="project-form">
              <div className="form-grid two">
                <label>Profile name<input value={credentialForm.name} onChange={(event) => setCredentialForm({ ...credentialForm, name: event.target.value })} /></label>
                <label>Role name optional<input value={credentialForm.role_name || ""} onChange={(event) => setCredentialForm({ ...credentialForm, role_name: event.target.value })} /></label>
              </div>
              <div className="form-grid two">
                <label>Username<input value={credentialForm.username || ""} onChange={(event) => setCredentialForm({ ...credentialForm, username: event.target.value })} /></label>
                <label>Password<input type="password" value={credentialForm.password || ""} onChange={(event) => setCredentialForm({ ...credentialForm, password: event.target.value })} /></label>
              </div>
              <label>Login URL<input value={credentialForm.login_url} onChange={(event) => setCredentialForm({ ...credentialForm, login_url: event.target.value })} /></label>
              <div className="form-grid three">
                <label>Username selector<input value={credentialForm.username_selector} onChange={(event) => setCredentialForm({ ...credentialForm, username_selector: event.target.value })} /></label>
                <label>Password selector<input value={credentialForm.password_selector} onChange={(event) => setCredentialForm({ ...credentialForm, password_selector: event.target.value })} /></label>
                <label>Submit selector<input value={credentialForm.submit_selector} onChange={(event) => setCredentialForm({ ...credentialForm, submit_selector: event.target.value })} /></label>
              </div>
              <div className="form-grid three">
                <label>Success URL contains<input value={credentialForm.success_url_contains} onChange={(event) => setCredentialForm({ ...credentialForm, success_url_contains: event.target.value })} /></label>
                <label>Success text contains<input value={credentialForm.success_text_contains} onChange={(event) => setCredentialForm({ ...credentialForm, success_text_contains: event.target.value })} /></label>
                <label>Failure text contains<input value={credentialForm.failure_text_contains} onChange={(event) => setCredentialForm({ ...credentialForm, failure_text_contains: event.target.value })} /></label>
              </div>
            </form>
          )}
        </section>
      )}

      {step === 4 && (
        <section>
          <h2>OpenAPI optional</h2>
          <Notice tone="info" message="Only safe read-only operations are executed by default." />
          <div className="toggle-grid">
            <label className="check-row"><input type="radio" checked={apiSpecMode === "skip"} onChange={() => setAPISpecMode("skip")} /> Skip API testing</label>
            <label className="check-row"><input type="radio" checked={apiSpecMode === "url"} onChange={() => setAPISpecMode("url")} /> Import OpenAPI URL</label>
            <label className="check-row"><input type="radio" checked={apiSpecMode === "inline"} onChange={() => setAPISpecMode("inline")} /> Paste OpenAPI JSON/YAML</label>
            <label className="check-row"><input type="radio" checked={apiSpecMode === "demo"} onChange={() => setAPISpecMode("demo")} /> Use demo API spec</label>
          </div>
          {apiSpecMode !== "skip" && apiSpecMode !== "demo" && (
            <form className="project-form">
              <label>Spec name<input value={apiSpecForm.name} onChange={(event) => setAPISpecForm({ ...apiSpecForm, name: event.target.value })} /></label>
              {apiSpecMode === "url" ? (
                <label>Source URL<input value={apiSpecForm.source_url} onChange={(event) => setAPISpecForm({ ...apiSpecForm, source_url: event.target.value })} /></label>
              ) : (
                <label>Inline spec<textarea className="spec-textarea" value={apiSpecForm.raw_spec} onChange={(event) => setAPISpecForm({ ...apiSpecForm, raw_spec: event.target.value })} /></label>
              )}
            </form>
          )}
        </section>
      )}

      {step === 5 && (
        <section>
          <h2>Select first workflow</h2>
          <div className="checkbox-grid wizard-checks">
            <label className="check-row"><input type="checkbox" checked={workflow.browser_smoke} onChange={(event) => setWorkflow({ ...workflow, browser_smoke: event.target.checked })} /> Browser smoke</label>
            <label className="check-row"><input type="checkbox" checked={workflow.discovery} onChange={(event) => setWorkflow({ ...workflow, discovery: event.target.checked })} /> Discovery</label>
            <label className="check-row"><input type="checkbox" checked={workflow.quality_checks} onChange={(event) => setWorkflow({ ...workflow, quality_checks: event.target.checked })} /> Quality checks</label>
            <label className="check-row"><input type="checkbox" checked={workflow.safe_qa_run} disabled={aiMode === "skip"} onChange={(event) => setWorkflow({ ...workflow, safe_qa_run: event.target.checked })} /> Safe QA run</label>
            <label className="check-row"><input type="checkbox" checked={workflow.api_smoke} disabled={apiSpecMode === "skip"} onChange={(event) => setWorkflow({ ...workflow, api_smoke: event.target.checked })} /> API smoke</label>
            <label className="check-row"><input type="checkbox" checked={workflow.authenticated_smoke} disabled={credentialMode === "skip"} onChange={(event) => setWorkflow({ ...workflow, authenticated_smoke: event.target.checked })} /> Authenticated smoke</label>
            <label className="check-row"><input type="checkbox" checked={workflow.execute_safe_qa} disabled={!workflow.safe_qa_run} onChange={(event) => setWorkflow({ ...workflow, execute_safe_qa: event.target.checked })} /> Execute Safe QA after preview</label>
          </div>
          {aiMode === "skip" && <Notice tone="info" message="Safe QA Run is disabled because AI was skipped. Deterministic browser, discovery, quality, and API checks can still run." />}
          <div className="form-actions">
            <button type="button" disabled={submitting} onClick={() => void submit()}>
              {submitting ? "Creating" : "Create project and run checks"}
            </button>
          </div>
        </section>
      )}

      {step === 6 && (
        <section>
          <h2>Progress and results</h2>
          {!result ? (
            <EmptyState title="No setup run yet" body="Complete the workflow step to create a project and start checks." />
          ) : (
            <SetupResultPanel response={result} />
          )}
        </section>
      )}
    </div>
  );
}

function SetupResultPanel({ response }: { response: ProjectSetupResponse }) {
  return (
    <div className="metadata-stack">
      <div className="detail-grid compact">
        <Field label="Project" value={response.project.name} />
        <Field label="Project ID" value={shortID(response.project.id)} />
        <Field label="AI provider" value={response.ai_provider ? response.ai_provider.name : "Skipped"} />
        <Field label="OpenAPI" value={response.api_spec ? response.api_spec.status : "Skipped"} />
      </div>
      <div className="table-wrap">
        <table>
          <thead><tr><th>Step</th><th>Status</th><th>Resource</th><th>Reason</th></tr></thead>
          <tbody>
            {response.timeline.map((item, index) => (
              <tr key={`${item.step}-${index}`}>
                <td>{humanizeIdentifier(item.step)}</td>
                <td><StatusBadge status={item.status} /></td>
                <td>{item.resource ? shortID(item.resource) : "None"}</td>
                <td>{item.reason || "None"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {response.skipped.length > 0 && (
        <Notice tone="info" message={`Skipped: ${response.skipped.map((item) => `${humanizeIdentifier(item.action)} (${item.reason})`).join("; ")}`} />
      )}
      <div className="button-row">
        <a className="button" href={`#/projects/${response.project.id}`}>Project details</a>
        {response.started.safe_qa_run_id && <a className="button secondary-link" href={`#/qa-runs/${response.started.safe_qa_run_id}`}>Safe QA report</a>}
        {response.started.discovery_run_id && <a className="button secondary-link" href={`#/discovery-runs/${response.started.discovery_run_id}`}>Discovery report</a>}
        {response.started.quality_check_run_id && <a className="button secondary-link" href={`#/quality-check-runs/${response.started.quality_check_run_id}`}>Quality report</a>}
        {response.started.browser_smoke_run_id && <a className="button secondary-link" href={`#/runs/${response.started.browser_smoke_run_id}`}>Browser report</a>}
        {response.started.api_smoke_run_id && <a className="button secondary-link" href={`#/runs/${response.started.api_smoke_run_id}`}>API report</a>}
      </div>
    </div>
  );
}

function ProjectReadinessChecklist({
  project,
  providers,
  runs,
  credentialProfiles,
  apiSpecs,
  discoveryRuns,
  qualityRuns,
  safeExplorerRuns,
  qaRuns
}: {
  project: Project;
  providers: AIProvider[];
  runs: TestRun[];
  credentialProfiles: CredentialProfile[];
  apiSpecs: APISpec[];
  discoveryRuns: DiscoveryRun[];
  qualityRuns: QualityCheckRun[];
  safeExplorerRuns: SafeExplorerRun[];
  qaRuns: QARun[];
}) {
  const latestCompletedRun = runs.find((run) => run.status === "completed");
  const items = [
    { label: "Frontend URL configured", ready: Boolean(project.frontend_url), action: "Edit via guided setup", href: "#/setup-project" },
    { label: "AI provider configured", ready: providers.length > 0, action: "Configure AI", href: "#/ai-providers" },
    { label: "Discovery run exists", ready: discoveryRuns.length > 0, action: "Run discovery", href: `#/projects/${project.id}` },
    { label: "Quality check exists", ready: qualityRuns.length > 0, action: "Run quality checks", href: `#/projects/${project.id}` },
    { label: "Safe Explorer run exists", ready: safeExplorerRuns.length > 0, action: "Start Safe Explorer", href: `#/projects/${project.id}` },
    { label: "Credential profile exists", ready: credentialProfiles.length > 0, action: "Add credentials", href: `#/projects/${project.id}` },
    { label: "OpenAPI spec imported", ready: apiSpecs.length > 0, action: "Import OpenAPI", href: `#/projects/${project.id}` },
    { label: "Latest Safe QA run exists", ready: qaRuns.length > 0, action: "Start Safe QA", href: `#/projects/${project.id}` },
    { label: "Latest report available", ready: Boolean(latestCompletedRun || qaRuns.length || qualityRuns.length || discoveryRuns.length || safeExplorerRuns.length), action: "View reports", href: "#/reports" }
  ];
  return (
    <div className="readiness-grid">
      {items.map((item) => (
        <div className="readiness-item" key={item.label}>
          <span className={item.ready ? "result-ok" : "result-failed"}>{item.ready ? "Ready" : "Not ready"}</span>
          <strong>{item.label}</strong>
          <a href={item.href}>{item.ready ? "View" : item.action}</a>
        </div>
      ))}
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

type ReportIndexItem = {
  id: string;
  project_id: string;
  project_name: string;
  type: string;
  status: string;
  created_at: string;
  web_href: string;
  html_href?: string;
};

function ReportsPage({ projects, runs, projectByID }: { projects: LoadState<Project[]>; runs: LoadState<TestRun[]>; projectByID: Map<string, Project> }) {
  const [items, setItems] = useState<LoadState<ReportIndexItem[]>>({ data: [], loading: true, error: "" });

  useEffect(() => {
    let canceled = false;
    async function loadReports() {
      setItems((current) => ({ ...current, loading: true, error: "" }));
      try {
        const projectItems = await Promise.all(
          projects.data.map(async (project) => {
            const [discovery, quality, safeExplorer, qaRuns] = await Promise.all([
              listDiscoveryRuns(project.id),
              listQualityCheckRuns(project.id),
              listSafeExplorerRuns(project.id),
              listQARuns(project.id)
            ]);
            return [
              ...discovery.map((run): ReportIndexItem => ({
                id: run.id,
                project_id: project.id,
                project_name: project.name,
                type: "Discovery",
                status: run.status,
                created_at: run.created_at,
                web_href: `#/discovery-runs/${run.id}`,
                html_href: discoveryHTMLReportURL(run.id)
              })),
              ...quality.map((run): ReportIndexItem => ({
                id: run.id,
                project_id: project.id,
                project_name: project.name,
                type: "Quality",
                status: run.status,
                created_at: run.created_at,
                web_href: `#/quality-check-runs/${run.id}`,
                html_href: qualityCheckHTMLReportURL(run.id)
              })),
              ...safeExplorer.map((run): ReportIndexItem => ({
                id: run.id,
                project_id: project.id,
                project_name: project.name,
                type: "Safe Explorer",
                status: run.status,
                created_at: run.created_at,
                web_href: `#/safe-explorer-runs/${run.id}`,
                html_href: safeExplorerHTMLReportURL(run.id)
              })),
              ...qaRuns.map((run): ReportIndexItem => ({
                id: run.id,
                project_id: project.id,
                project_name: project.name,
                type: "Safe QA",
                status: run.status,
                created_at: run.created_at,
                web_href: `#/qa-runs/${run.id}`,
                html_href: qaRunHTMLReportURL(run.id)
              }))
            ];
          })
        );
        const runItems = runs.data.map((run): ReportIndexItem => ({
          id: run.id,
          project_id: run.project_id,
          project_name: projectByID.get(run.project_id)?.name || run.project_id,
          type: formatRunType(run.run_type),
          status: run.status,
          created_at: run.created_at,
          web_href: `#/runs/${run.id}`,
          html_href: htmlReportURL(run.id)
        }));
        const nextItems = [...runItems, ...projectItems.flat()].sort((left, right) => Date.parse(right.created_at) - Date.parse(left.created_at));
        if (!canceled) {
          setItems({ data: nextItems, loading: false, error: "" });
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        if (!canceled) {
          setItems({ data: [], loading: false, error: message });
        }
      }
    }
    if (!projects.loading && !runs.loading) {
      void loadReports();
    }
    return () => {
      canceled = true;
    };
  }, [projectByID, projects.data, projects.loading, runs.data, runs.loading]);

  return (
    <section>
      <div className="section-heading">
        <div>
          <h2>Reports</h2>
          <p>Recent browser, API, discovery, Safe Explorer, quality, Safe QA, authorization, and test-plan execution reports.</p>
        </div>
        <a className="button secondary-link" href="#/setup-project">Guided Setup</a>
      </div>
      {items.error && <Notice tone="danger" message={items.error} />}
      {items.loading ? <SkeletonRows /> : <ReportIndexTable items={items.data} />}
    </section>
  );
}

function ReportIndexTable({ items }: { items: ReportIndexItem[] }) {
  if (items.length === 0) {
    return <EmptyState title="No reports yet" body="Run browser, API, discovery, Safe Explorer, quality, or Safe QA checks to populate reports." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Type</th>
            <th>Project</th>
            <th>Report</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {items.slice(0, 30).map((item) => (
            <tr key={`${item.type}-${item.id}`}>
              <td><StatusBadge status={item.status} /></td>
              <td>{item.type}</td>
              <td><a href={`#/projects/${item.project_id}`}>{item.project_name}</a></td>
              <td>
                <a href={item.web_href}>{shortID(item.id)}</a>
                {item.html_href && <>{" "} <a href={item.html_href} target="_blank" rel="noreferrer">HTML</a></>}
              </td>
              <td>{formatDate(item.created_at)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
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
          message="AI provider credentials are protected by local admin authentication and encrypted at rest, but this alpha is still intended for trusted self-hosted environments."
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
            <th>Type</th>
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
              <td>{formatRunType(run.run_type)}</td>
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
  const [apiSpecs, setAPISpecs] = useState<LoadState<APISpec[]>>({ data: [], loading: true, error: "" });
  const [credentialProfiles, setCredentialProfiles] = useState<LoadState<CredentialProfile[]>>({ data: [], loading: true, error: "" });
  const [authorizationChecks, setAuthorizationChecks] = useState<LoadState<AuthorizationCheck[]>>({ data: [], loading: true, error: "" });
  const [authorizationRuns, setAuthorizationRuns] = useState<LoadState<AuthorizationCheckRun[]>>({ data: [], loading: true, error: "" });
  const [discoveryRuns, setDiscoveryRuns] = useState<LoadState<DiscoveryRun[]>>({ data: [], loading: true, error: "" });
  const [qualityRuns, setQualityRuns] = useState<LoadState<QualityCheckRun[]>>({ data: [], loading: true, error: "" });
  const [safeExplorerRuns, setSafeExplorerRuns] = useState<LoadState<SafeExplorerRun[]>>({ data: [], loading: true, error: "" });
  const [qaRuns, setQARuns] = useState<LoadState<QARun[]>>({ data: [], loading: true, error: "" });
  const [error, setError] = useState("");
  const [starting, setStarting] = useState("");

  const refresh = useCallback(async () => {
    setRuns((current) => ({ ...current, loading: true, error: "" }));
    setTestPlans((current) => ({ ...current, loading: true, error: "" }));
    setAPISpecs((current) => ({ ...current, loading: true, error: "" }));
    setCredentialProfiles((current) => ({ ...current, loading: true, error: "" }));
    setAuthorizationChecks((current) => ({ ...current, loading: true, error: "" }));
    setAuthorizationRuns((current) => ({ ...current, loading: true, error: "" }));
    setDiscoveryRuns((current) => ({ ...current, loading: true, error: "" }));
    setQualityRuns((current) => ({ ...current, loading: true, error: "" }));
    setSafeExplorerRuns((current) => ({ ...current, loading: true, error: "" }));
    setQARuns((current) => ({ ...current, loading: true, error: "" }));
    setError("");
    try {
      const [
        nextProject,
        nextRuns,
        nextProviders,
        nextTestPlans,
        nextAPISpecs,
        nextCredentialProfiles,
        nextAuthorizationChecks,
        nextAuthorizationRuns,
        nextDiscoveryRuns,
        nextQualityRuns,
        nextSafeExplorerRuns,
        nextQARuns
      ] = await Promise.all([
        cachedProject ? Promise.resolve(cachedProject) : getProject(projectID),
        listRuns(projectID),
        listAIProviders(),
        listTestPlans(projectID),
        listAPISpecs(projectID),
        listCredentialProfiles(projectID),
        listAuthorizationChecks(projectID),
        listAuthorizationCheckRuns(projectID),
        listDiscoveryRuns(projectID),
        listQualityCheckRuns(projectID),
        listSafeExplorerRuns(projectID),
        listQARuns(projectID)
      ]);
      setProject(nextProject);
      setRuns({ data: nextRuns, loading: false, error: "" });
      setProviders(nextProviders);
      setTestPlans({ data: nextTestPlans, loading: false, error: "" });
      setAPISpecs({ data: nextAPISpecs, loading: false, error: "" });
      setCredentialProfiles({ data: nextCredentialProfiles, loading: false, error: "" });
      setAuthorizationChecks({ data: nextAuthorizationChecks, loading: false, error: "" });
      setAuthorizationRuns({ data: nextAuthorizationRuns, loading: false, error: "" });
      setDiscoveryRuns({ data: nextDiscoveryRuns, loading: false, error: "" });
      setQualityRuns({ data: nextQualityRuns, loading: false, error: "" });
      setSafeExplorerRuns({ data: nextSafeExplorerRuns, loading: false, error: "" });
      setQARuns({ data: nextQARuns, loading: false, error: "" });
    } catch (loadError) {
      const message = loadError instanceof Error ? loadError.message : String(loadError);
      setError(message);
      setRuns((current) => ({ ...current, loading: false, error: message }));
      setTestPlans((current) => ({ ...current, loading: false, error: message }));
      setAPISpecs((current) => ({ ...current, loading: false, error: message }));
      setCredentialProfiles((current) => ({ ...current, loading: false, error: message }));
      setAuthorizationChecks((current) => ({ ...current, loading: false, error: message }));
      setAuthorizationRuns((current) => ({ ...current, loading: false, error: message }));
      setDiscoveryRuns((current) => ({ ...current, loading: false, error: message }));
      setQualityRuns((current) => ({ ...current, loading: false, error: message }));
      setSafeExplorerRuns((current) => ({ ...current, loading: false, error: message }));
      setQARuns((current) => ({ ...current, loading: false, error: message }));
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

  async function startAuthenticatedRun(profileID?: string) {
    if (!project) {
      return;
    }
    setStarting("authenticated");
    setError("");
    try {
      const run = await startAuthenticatedBrowserSmokeRun(project.id, {
        credential_profile_id: profileID,
        target_path: "/dashboard",
        capture_screenshot: true,
        max_duration_seconds: 30
      });
      window.location.hash = `#/runs/${run.id}`;
    } catch (startError) {
      setError(startError instanceof Error ? startError.message : String(startError));
    } finally {
      setStarting("");
    }
  }

  async function startAuthorizationRun() {
    if (!project) {
      return;
    }
    setStarting("authorization");
    setError("");
    try {
      const run = await startAuthorizationCheckRun(project.id, { max_checks: 10 });
      await refresh();
      window.location.hash = `#/authorization-check-runs/${run.id}`;
    } catch (startError) {
      setError(startError instanceof Error ? startError.message : String(startError));
    } finally {
      setStarting("");
    }
  }

  async function startProjectDiscovery(input: DiscoveryRunInput) {
    if (!project) {
      return;
    }
    setStarting("discovery");
    setError("");
    try {
      const run = await startDiscoveryRun(project.id, input);
      await refresh();
      window.location.hash = `#/discovery-runs/${run.id}`;
    } catch (startError) {
      setError(startError instanceof Error ? startError.message : String(startError));
    } finally {
      setStarting("");
    }
  }

  async function startProjectQualityCheck(input: QualityCheckRunInput) {
    if (!project) {
      return;
    }
    setStarting("quality");
    setError("");
    try {
      const run = await startQualityCheckRun(project.id, input);
      await refresh();
      window.location.hash = `#/quality-check-runs/${run.id}`;
    } catch (startError) {
      setError(startError instanceof Error ? startError.message : String(startError));
    } finally {
      setStarting("");
    }
  }

  async function startProjectSafeExplorer(input: SafeExplorerRunInput) {
    if (!project) {
      return;
    }
    setStarting("safe-explorer");
    setError("");
    try {
      const run = await startSafeExplorerRun(project.id, input);
      await refresh();
      window.location.hash = `#/safe-explorer-runs/${run.id}`;
    } catch (startError) {
      setError(startError instanceof Error ? startError.message : String(startError));
    } finally {
      setStarting("");
    }
  }

  async function startSafeQARun(input: QARunInput) {
    if (!project) {
      return;
    }
    setStarting("qa-run");
    setError("");
    try {
      const run = await startQARun(project.id, input);
      await refresh();
      window.location.hash = `#/qa-runs/${run.id}`;
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
            {project.frontend_url && credentialProfiles.data.length > 0 && (
              <button type="button" className="secondary" disabled={starting !== ""} onClick={() => void startAuthenticatedRun()}>
                {starting === "authenticated" ? "Starting" : "Run authenticated browser smoke test"}
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
            <h2>Project Readiness</h2>
            <p>Checklist for a useful first report and repeatable Safe QA workflow.</p>
          </div>
          <a className="button secondary-link" href="#/setup-project">
            Guided Setup
          </a>
        </div>
        <ProjectReadinessChecklist
          project={project}
          providers={providers}
          runs={runs.data}
          credentialProfiles={credentialProfiles.data}
          apiSpecs={apiSpecs.data}
          discoveryRuns={discoveryRuns.data}
          qualityRuns={qualityRuns.data}
          safeExplorerRuns={safeExplorerRuns.data}
          qaRuns={qaRuns.data}
        />
      </section>

      <section id="application-discovery">
        <div className="section-heading">
          <div>
            <h2>Application Discovery</h2>
            <p>Safely crawls same-origin pages, collects links/forms/browser signals, and builds an application map.</p>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
        <Notice
          tone="info"
          message="Discovery does not submit forms or perform destructive actions. Unsafe, external, unsupported, and non-HTML links are skipped."
        />
        <DiscoveryRunForm
          project={project}
          profiles={credentialProfiles.data}
          disabled={starting !== ""}
          onStart={(input) => void startProjectDiscovery(input)}
        />
        <div className="section-split">
          {discoveryRuns.error && <Notice tone="danger" message={discoveryRuns.error} />}
          {discoveryRuns.loading ? <SkeletonRows /> : <DiscoveryRunTable runs={discoveryRuns.data} />}
        </div>
      </section>

      <section id="interactive-safe-explorer">
        <div className="section-heading">
          <div>
            <h2>Interactive Safe Explorer</h2>
            <p>Deterministically explores safe same-origin navigation actions and records executed/skipped action timelines.</p>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
        <Notice
          tone="info"
          message="Safe Explorer does not use AI to choose actions. It skips destructive-looking links, POST forms, unsafe buttons, external hosts, sensitive queries, and unsupported interactions."
        />
        <SafeExplorerRunForm
          project={project}
          profiles={credentialProfiles.data}
          disabled={starting !== ""}
          onStart={(input) => void startProjectSafeExplorer(input)}
        />
        <div className="section-split">
          {safeExplorerRuns.error && <Notice tone="danger" message={safeExplorerRuns.error} />}
          {safeExplorerRuns.loading ? <SkeletonRows /> : <SafeExplorerRunTable runs={safeExplorerRuns.data} />}
        </div>
      </section>

      <section id="safe-qa-runs">
        <div className="section-heading">
          <div>
            <h2>Safe QA Runs</h2>
            <p>Generate discovery-aware AI plans, preview safe steps, and optionally execute the safe subset.</p>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
        <Notice
          tone="info"
          message="Safe QA runs send only sanitized discovery metadata to AI. They do not send credentials, cookies, storage, auth headers, screenshots, full HTML, or response bodies."
        />
        <QARunForm
          project={project}
          providers={providers}
          profiles={credentialProfiles.data}
          discoveryRuns={discoveryRuns.data}
          disabled={starting !== ""}
          onStart={(input) => void startSafeQARun(input)}
        />
        <div className="section-split">
          {qaRuns.error && <Notice tone="danger" message={qaRuns.error} />}
          {qaRuns.loading ? <SkeletonRows /> : <QARunTable runs={qaRuns.data} />}
        </div>
      </section>

      <section id="quality-checks">
        <div className="section-heading">
          <div>
            <h2>Quality Checks</h2>
            <p>Passive front-end security, accessibility, and performance heuristics for allowed pages.</p>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
        <Notice
          tone="info"
          message="Quality checks are read-only and passive. They do not submit forms, run payloads, fuzz inputs, or perform active security scanning."
        />
        <QualityCheckRunForm
          project={project}
          profiles={credentialProfiles.data}
          discoveryRuns={discoveryRuns.data}
          disabled={starting !== ""}
          onStart={(input) => void startProjectQualityCheck(input)}
        />
        <div className="section-split">
          {qualityRuns.error && <Notice tone="danger" message={qualityRuns.error} />}
          {qualityRuns.loading ? <SkeletonRows /> : <QualityCheckRunTable runs={qualityRuns.data} />}
        </div>
      </section>

      <section id="authorization-checks">
        <div className="section-heading">
          <div>
            <h2>Authorization Checks</h2>
            <p>Explicit, read-only role-aware browser URL checks using configured test credentials.</p>
          </div>
          <div className="button-row">
            <button
              type="button"
              className="secondary"
              disabled={starting !== "" || authorizationChecks.data.length === 0}
              onClick={() => void startAuthorizationRun()}
            >
              {starting === "authorization" ? "Starting" : "Run authorization checks"}
            </button>
            <button type="button" className="secondary" onClick={() => void refresh()}>
              Refresh
            </button>
          </div>
        </div>
        <Notice
          tone="info"
          message="Authorization checks are explicit, read-only, and use configured test credentials. No destructive actions are performed."
        />
        <AuthorizationCheckForm project={project} profiles={credentialProfiles.data} onSaved={() => void refresh()} />
        <div className="section-split">
          {authorizationChecks.error && <Notice tone="danger" message={authorizationChecks.error} />}
          {authorizationChecks.loading ? (
            <SkeletonRows />
          ) : (
            <AuthorizationCheckTable
              checks={authorizationChecks.data}
              profiles={credentialProfiles.data}
              onChanged={() => void refresh()}
            />
          )}
          <h3>Authorization Runs</h3>
          {authorizationRuns.error && <Notice tone="danger" message={authorizationRuns.error} />}
          {authorizationRuns.loading ? <SkeletonRows /> : <AuthorizationRunTable runs={authorizationRuns.data} />}
        </div>
      </section>

      <section id="api-specs">
        <div className="section-heading">
          <div>
            <h2>API Specs</h2>
            <p>Import OpenAPI 3.x specs and discover safe read-only operations.</p>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
        <Notice
          tone="info"
          message="Safe API smoke tests execute only GET, HEAD, and OPTIONS operations. Mutating, authenticated, ambiguous, or unsafe operations are skipped."
        />
        <APISpecImportForm project={project} onImported={() => void refresh()} />
        <div className="section-split">
          {apiSpecs.error && <Notice tone="danger" message={apiSpecs.error} />}
          {apiSpecs.loading ? <SkeletonRows /> : <APISpecTable specs={apiSpecs.data} onDeleted={() => void refresh()} />}
        </div>
      </section>

      <section id="credential-profiles">
        <div className="section-heading">
          <div>
            <h2>Credential Profiles</h2>
            <p>Deterministic test credentials for target application login flows.</p>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
        <Notice
          tone="info"
          message="This alpha build has no Qualora authentication. Store test credentials only in trusted local/self-hosted environments. Credentials are encrypted at rest and are never sent to AI."
        />
        <CredentialProfileForm project={project} onSaved={() => void refresh()} />
        <div className="section-split">
          {credentialProfiles.error && <Notice tone="danger" message={credentialProfiles.error} />}
          {credentialProfiles.loading ? (
            <SkeletonRows />
          ) : (
            <CredentialProfileTable
              profiles={credentialProfiles.data}
              onChanged={() => void refresh()}
              onTestRun={(run) => {
                window.location.hash = `#/runs/${run.id}`;
              }}
              onAuthenticatedRun={(profileID) => void startAuthenticatedRun(profileID)}
            />
          )}
          {!credentialProfiles.loading && credentialProfiles.data.length === 0 && (
            <Notice tone="info" message="Create a credential profile to enable authenticated browser smoke testing." />
          )}
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
        <Notice
          tone="info"
          message="AI test plans are suggestions. Qualora can execute only approved safe DSL steps, and skips unsupported, authenticated, destructive, or ambiguous steps."
        />
        <GenerateAITestPlanForm
          project={project}
          runs={runs.data}
          discoveryRuns={discoveryRuns.data}
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

function DiscoveryRunForm({
  project,
  profiles,
  disabled,
  onStart
}: {
  project: Project;
  profiles: CredentialProfile[];
  disabled: boolean;
  onStart: (input: DiscoveryRunInput) => void;
}) {
  const [form, setForm] = useState<Required<Pick<DiscoveryRunInput, "start_url" | "credential_profile_id" | "max_pages" | "max_depth" | "same_origin_only">>>({
    start_url: project.frontend_url || "",
    credential_profile_id: "",
    max_pages: 20,
    max_depth: 2,
    same_origin_only: true
  });

  useEffect(() => {
    setForm((current) => ({ ...current, start_url: project.frontend_url || current.start_url }));
  }, [project.frontend_url]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onStart({
      start_url: form.start_url.trim() || undefined,
      credential_profile_id: form.credential_profile_id || undefined,
      max_pages: Number(form.max_pages || 20),
      max_depth: Number(form.max_depth || 2),
      same_origin_only: form.same_origin_only
    });
  }

  return (
    <form className="project-form compact-form" onSubmit={submit}>
      <label>
        Start URL
        <input value={form.start_url} onChange={(event) => setForm({ ...form, start_url: event.target.value })} />
      </label>
      <label>
        Credential Profile
        <select
          value={form.credential_profile_id}
          onChange={(event) => setForm({ ...form, credential_profile_id: event.target.value })}
        >
          <option value="">Unauthenticated</option>
          {profiles.map((profile) => (
            <option key={profile.id} value={profile.id}>
              {credentialProfileLabel(profile)}
            </option>
          ))}
        </select>
      </label>
      <label>
        Max Pages
        <input
          type="number"
          min={1}
          max={100}
          value={form.max_pages}
          onChange={(event) => setForm({ ...form, max_pages: Number(event.target.value) })}
        />
      </label>
      <label>
        Max Depth
        <input
          type="number"
          min={0}
          max={5}
          value={form.max_depth}
          onChange={(event) => setForm({ ...form, max_depth: Number(event.target.value) })}
        />
      </label>
      <label className="check-row">
        <input
          type="checkbox"
          checked={form.same_origin_only}
          onChange={(event) => setForm({ ...form, same_origin_only: event.target.checked })}
        />
        Same origin only
      </label>
      <div className="form-actions">
        <button type="submit" disabled={disabled || !project.frontend_url}>
          {disabled ? "Starting" : "Start discovery"}
        </button>
      </div>
    </form>
  );
}

function DiscoveryRunTable({ runs }: { runs: DiscoveryRun[] }) {
  if (runs.length === 0) {
    return <EmptyState title="No discovery runs" body="Start discovery to build a persistent application map." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Start URL</th>
            <th>Pages</th>
            <th>Links</th>
            <th>Forms</th>
            <th>Findings</th>
            <th>Created</th>
            <th>Report</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id}>
              <td>
                <StatusBadge status={run.status} />
              </td>
              <td>
                <code>{run.start_url}</code>
              </td>
              <td>{run.total_pages}</td>
              <td>{run.total_links}</td>
              <td>{run.total_forms}</td>
              <td>{run.total_findings}</td>
              <td>{formatDate(run.created_at)}</td>
              <td>
                <a href={`#/discovery-runs/${run.id}`}>Open</a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function SafeExplorerRunForm({
  project,
  profiles,
  disabled,
  onStart
}: {
  project: Project;
  profiles: CredentialProfile[];
  disabled: boolean;
  onStart: (input: SafeExplorerRunInput) => void;
}) {
  const [form, setForm] = useState<Required<Pick<SafeExplorerRunInput, "start_url" | "credential_profile_id" | "max_steps" | "max_depth" | "same_origin_only" | "allow_get_forms">>>({
    start_url: project.frontend_url || "",
    credential_profile_id: "",
    max_steps: 10,
    max_depth: 2,
    same_origin_only: true,
    allow_get_forms: false
  });

  useEffect(() => {
    setForm((current) => ({ ...current, start_url: project.frontend_url || current.start_url }));
  }, [project.frontend_url]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onStart({
      start_url: form.start_url.trim() || undefined,
      credential_profile_id: form.credential_profile_id || undefined,
      max_steps: Number(form.max_steps || 10),
      max_depth: Number(form.max_depth || 2),
      same_origin_only: form.same_origin_only,
      allow_get_forms: form.allow_get_forms
    });
  }

  return (
    <form className="project-form compact-form" onSubmit={submit}>
      <label>
        Start URL
        <input value={form.start_url} onChange={(event) => setForm({ ...form, start_url: event.target.value })} />
      </label>
      <label>
        Credential Profile
        <select
          value={form.credential_profile_id}
          onChange={(event) => setForm({ ...form, credential_profile_id: event.target.value })}
        >
          <option value="">Unauthenticated</option>
          {profiles.map((profile) => (
            <option key={profile.id} value={profile.id}>
              {credentialProfileLabel(profile)}
            </option>
          ))}
        </select>
      </label>
      <label>
        Max Steps
        <input
          type="number"
          min={1}
          max={50}
          value={form.max_steps}
          onChange={(event) => setForm({ ...form, max_steps: Number(event.target.value) })}
        />
      </label>
      <label>
        Max Depth
        <input
          type="number"
          min={0}
          max={5}
          value={form.max_depth}
          onChange={(event) => setForm({ ...form, max_depth: Number(event.target.value) })}
        />
      </label>
      <label className="check-row">
        <input
          type="checkbox"
          checked={form.same_origin_only}
          onChange={(event) => setForm({ ...form, same_origin_only: event.target.checked })}
        />
        Same origin only
      </label>
      <label className="check-row">
        <input
          type="checkbox"
          checked={form.allow_get_forms}
          onChange={(event) => setForm({ ...form, allow_get_forms: event.target.checked })}
        />
        Allow safe GET forms
      </label>
      <div className="form-actions">
        <button type="submit" disabled={disabled || !project.frontend_url}>
          {disabled ? "Starting" : "Start Safe Explorer"}
        </button>
      </div>
    </form>
  );
}

function SafeExplorerRunTable({ runs }: { runs: SafeExplorerRun[] }) {
  if (runs.length === 0) {
    return <EmptyState title="No Safe Explorer runs" body="Start Safe Explorer to record a deterministic action timeline." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Start URL</th>
            <th>Pages</th>
            <th>Detected</th>
            <th>Executed</th>
            <th>Skipped</th>
            <th>Findings</th>
            <th>Created</th>
            <th>Report</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id}>
              <td>
                <StatusBadge status={run.status} />
              </td>
              <td>
                <code>{run.start_url}</code>
              </td>
              <td>{run.total_pages_observed}</td>
              <td>{run.total_actions_detected}</td>
              <td>{run.total_actions_executed}</td>
              <td>{run.total_actions_skipped}</td>
              <td>{run.total_findings}</td>
              <td>{formatDate(run.created_at)}</td>
              <td>
                <a href={`#/safe-explorer-runs/${run.id}`}>Open</a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function QualityCheckRunForm({
  project,
  profiles,
  discoveryRuns,
  disabled,
  onStart
}: {
  project: Project;
  profiles: CredentialProfile[];
  discoveryRuns: DiscoveryRun[];
  disabled: boolean;
  onStart: (input: QualityCheckRunInput) => void;
}) {
  const completedDiscoveryRuns = discoveryRuns.filter((run) => run.status === "completed");
  const [source, setSource] = useState<"latest" | "existing" | "target">("latest");
  const [targetURL, setTargetURL] = useState(project.frontend_url || "");
  const [discoveryRunID, setDiscoveryRunID] = useState("");
  const [credentialProfileID, setCredentialProfileID] = useState("");
  const [maxPages, setMaxPages] = useState(10);
  const [includeSecurity, setIncludeSecurity] = useState(true);
  const [includeAccessibility, setIncludeAccessibility] = useState(true);
  const [includePerformance, setIncludePerformance] = useState(true);

  useEffect(() => {
    setTargetURL((current) => current || project.frontend_url || "");
  }, [project.frontend_url]);

  useEffect(() => {
    if (discoveryRunID || completedDiscoveryRuns.length === 0) {
      return;
    }
    setDiscoveryRunID(completedDiscoveryRuns[0].id);
  }, [completedDiscoveryRuns, discoveryRunID]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const input: QualityCheckRunInput = {
      credential_profile_id: credentialProfileID || undefined,
      max_pages: maxPages,
      include_security: includeSecurity,
      include_accessibility: includeAccessibility,
      include_performance: includePerformance
    };
    if (source === "latest") {
      input.use_latest_discovery = true;
      input.target_url = targetURL.trim() || undefined;
    } else if (source === "existing") {
      input.discovery_run_id = discoveryRunID || undefined;
      input.target_url = targetURL.trim() || undefined;
    } else {
      input.target_url = targetURL.trim() || undefined;
    }
    onStart(input);
  }

  return (
    <form className="project-form compact-form" onSubmit={submit}>
      <label>
        Source
        <select value={source} onChange={(event) => setSource(event.target.value as "latest" | "existing" | "target")}>
          <option value="latest">Use latest completed discovery when available</option>
          <option value="existing">Use selected discovery run</option>
          <option value="target">Check target URL only</option>
        </select>
      </label>
      {source === "existing" && (
        <label>
          Completed Discovery Run
          <select value={discoveryRunID} onChange={(event) => setDiscoveryRunID(event.target.value)} required>
            {completedDiscoveryRuns.map((run) => (
              <option key={run.id} value={run.id}>
                {shortID(run.id)} · {run.total_pages} pages · {formatDate(run.created_at)}
              </option>
            ))}
          </select>
        </label>
      )}
      <label>
        Target URL
        <input value={targetURL} onChange={(event) => setTargetURL(event.target.value)} />
      </label>
      <label>
        Credential Profile
        <select value={credentialProfileID} onChange={(event) => setCredentialProfileID(event.target.value)}>
          <option value="">Unauthenticated</option>
          {profiles.map((profile) => (
            <option key={profile.id} value={profile.id}>
              {credentialProfileLabel(profile)}
            </option>
          ))}
        </select>
      </label>
      <label>
        Max Pages
        <input type="number" min={1} max={50} value={maxPages} onChange={(event) => setMaxPages(Number(event.target.value))} />
      </label>
      <div className="checkbox-grid">
        <label className="check-row">
          <input type="checkbox" checked={includeSecurity} onChange={(event) => setIncludeSecurity(event.target.checked)} />
          Passive security
        </label>
        <label className="check-row">
          <input type="checkbox" checked={includeAccessibility} onChange={(event) => setIncludeAccessibility(event.target.checked)} />
          Accessibility
        </label>
        <label className="check-row">
          <input type="checkbox" checked={includePerformance} onChange={(event) => setIncludePerformance(event.target.checked)} />
          Performance
        </label>
      </div>
      <div className="form-actions">
        <button type="submit" disabled={disabled || !project.frontend_url || (!includeSecurity && !includeAccessibility && !includePerformance)}>
          {disabled ? "Starting" : "Start quality checks"}
        </button>
      </div>
    </form>
  );
}

function QualityCheckRunTable({ runs }: { runs: QualityCheckRun[] }) {
  if (runs.length === 0) {
    return <EmptyState title="No quality checks" body="Start passive quality checks to collect security, accessibility, and performance findings." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Target</th>
            <th>Pages</th>
            <th>Findings</th>
            <th>Categories</th>
            <th>Created</th>
            <th>Report</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id}>
              <td>
                <StatusBadge status={run.status} />
                {run.error_message && <p className="muted">{run.error_message}</p>}
              </td>
              <td>
                <code>{run.target_url}</code>
                {run.discovery_run_id && <p className="muted">Discovery {shortID(run.discovery_run_id)}</p>}
              </td>
              <td>{run.total_pages}</td>
              <td>
                {run.total_findings} total
                <p className="muted">
                  {run.high_findings} high · {run.medium_findings} medium · {run.low_findings} low
                </p>
              </td>
              <td>
                {[
                  run.include_security ? "security" : "",
                  run.include_accessibility ? "a11y" : "",
                  run.include_performance ? "performance" : ""
                ]
                  .filter(Boolean)
                  .join(", ")}
              </td>
              <td>{formatDate(run.created_at)}</td>
              <td>
                <div className="button-row compact">
                  <a href={`#/quality-check-runs/${run.id}`}>Open</a>
                  <a href={qualityCheckHTMLReportURL(run.id)} target="_blank" rel="noreferrer">
                    HTML
                  </a>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function QARunForm({
  project,
  providers,
  profiles,
  discoveryRuns,
  disabled,
  onStart
}: {
  project: Project;
  providers: AIProvider[];
  profiles: CredentialProfile[];
  discoveryRuns: DiscoveryRun[];
  disabled: boolean;
  onStart: (input: QARunInput) => void;
}) {
  const completedDiscoveryRuns = discoveryRuns.filter((run) => run.status === "completed");
  const [providerID, setProviderID] = useState("");
  const [discoveryMode, setDiscoveryMode] = useState<"latest" | "existing" | "new">("latest");
  const [existingDiscoveryRunID, setExistingDiscoveryRunID] = useState("");
  const [credentialProfileID, setCredentialProfileID] = useState("");
  const [startURL, setStartURL] = useState(project.frontend_url || "");
  const [maxPages, setMaxPages] = useState(20);
  const [maxDepth, setMaxDepth] = useState(2);
  const [maxScenarios, setMaxScenarios] = useState(10);
  const [includeQualityChecks, setIncludeQualityChecks] = useState(true);
  const [qualityMaxPages, setQualityMaxPages] = useState(10);
  const [qualityIncludeSecurity, setQualityIncludeSecurity] = useState(true);
  const [qualityIncludeAccessibility, setQualityIncludeAccessibility] = useState(true);
  const [qualityIncludePerformance, setQualityIncludePerformance] = useState(true);
  const [execute, setExecute] = useState(false);
  const [productContext, setProductContext] = useState("");
  const [focusAreas, setFocusAreas] = useState<string[]>(["smoke", "functional", "regression"]);

  useEffect(() => {
    if (providerID || providers.length === 0) {
      return;
    }
    setProviderID(providers.find((provider) => provider.is_default)?.id || providers[0].id);
  }, [providerID, providers]);

  useEffect(() => {
    if (existingDiscoveryRunID || completedDiscoveryRuns.length === 0) {
      return;
    }
    setExistingDiscoveryRunID(completedDiscoveryRuns[0].id);
  }, [completedDiscoveryRuns, existingDiscoveryRunID]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const input: QARunInput = {
      mode: "safe",
      provider_id: providerID || undefined,
      credential_profile_id: credentialProfileID || undefined,
      max_pages: maxPages,
      max_depth: maxDepth,
      max_scenarios: maxScenarios,
      include_quality_checks: includeQualityChecks,
      quality_max_pages: qualityMaxPages,
      quality_include_security: qualityIncludeSecurity,
      quality_include_accessibility: qualityIncludeAccessibility,
      quality_include_performance: qualityIncludePerformance,
      execute,
      product_context: productContext.trim() || undefined,
      focus_areas: focusAreas
    };
    if (discoveryMode === "existing") {
      input.use_existing_discovery_run_id = existingDiscoveryRunID || undefined;
    } else if (discoveryMode === "latest") {
      input.use_latest_discovery = true;
    } else {
      input.start_url = startURL.trim() || undefined;
    }
    onStart(input);
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
    return <Notice tone="info" message="Configure an AI provider to start a discovery-aware safe QA run." />;
  }

  return (
    <form className="project-form test-plan-form" onSubmit={submit}>
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
          Discovery Source
          <select value={discoveryMode} onChange={(event) => setDiscoveryMode(event.target.value as "latest" | "existing" | "new")}>
            <option value="latest">Use latest completed, or run new</option>
            <option value="existing">Use selected completed discovery</option>
            <option value="new">Run new safe discovery</option>
          </select>
        </label>
        {discoveryMode === "existing" && (
          <label>
            Completed Discovery Run
            <select value={existingDiscoveryRunID} onChange={(event) => setExistingDiscoveryRunID(event.target.value)} required>
              {completedDiscoveryRuns.map((run) => (
                <option key={run.id} value={run.id}>
                  {shortID(run.id)} · {run.total_pages} pages · {formatDate(run.created_at)}
                </option>
              ))}
            </select>
          </label>
        )}
        {discoveryMode === "new" && (
          <label>
            Start URL
            <input value={startURL} onChange={(event) => setStartURL(event.target.value)} />
          </label>
        )}
        <label>
          Credential Profile
          <select value={credentialProfileID} onChange={(event) => setCredentialProfileID(event.target.value)}>
            <option value="">Unauthenticated discovery</option>
            {profiles.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {credentialProfileLabel(profile)}
              </option>
            ))}
          </select>
        </label>
        <label>
          Max Pages
          <input type="number" min="1" max="100" value={maxPages} onChange={(event) => setMaxPages(Number(event.target.value))} />
        </label>
        <label>
          Max Depth
          <input type="number" min="0" max="5" value={maxDepth} onChange={(event) => setMaxDepth(Number(event.target.value))} />
        </label>
        <label>
          Max Scenarios
          <input type="number" min="1" max="30" value={maxScenarios} onChange={(event) => setMaxScenarios(Number(event.target.value))} />
        </label>
      </div>
      <label>
        Product Context
        <textarea
          value={productContext}
          placeholder="Optional product behavior or workflows to emphasize. Do not include secrets."
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
      <label className="check-row">
        <input type="checkbox" checked={execute} onChange={(event) => setExecute(event.target.checked)} />
        Execute supported safe steps after preview
      </label>
      <label className="check-row">
        <input type="checkbox" checked={includeQualityChecks} onChange={(event) => setIncludeQualityChecks(event.target.checked)} />
        Include passive quality checks after discovery
      </label>
      {includeQualityChecks && (
        <div className="form-grid two">
          <label>
            Quality Max Pages
            <input type="number" min="1" max="50" value={qualityMaxPages} onChange={(event) => setQualityMaxPages(Number(event.target.value))} />
          </label>
          <div className="checkbox-grid">
            <label className="check-row">
              <input type="checkbox" checked={qualityIncludeSecurity} onChange={(event) => setQualityIncludeSecurity(event.target.checked)} />
              Passive security
            </label>
            <label className="check-row">
              <input type="checkbox" checked={qualityIncludeAccessibility} onChange={(event) => setQualityIncludeAccessibility(event.target.checked)} />
              Accessibility
            </label>
            <label className="check-row">
              <input type="checkbox" checked={qualityIncludePerformance} onChange={(event) => setQualityIncludePerformance(event.target.checked)} />
              Performance
            </label>
          </div>
        </div>
      )}
      <div className="form-actions">
        <button
          type="submit"
          disabled={
            disabled ||
            !project.frontend_url ||
            focusAreas.length === 0 ||
            (includeQualityChecks && !qualityIncludeSecurity && !qualityIncludeAccessibility && !qualityIncludePerformance)
          }
        >
          {disabled ? "Starting" : execute ? "Start safe QA run" : "Generate safe QA preview"}
        </button>
      </div>
    </form>
  );
}

function QARunTable({ runs }: { runs: QARun[] }) {
  if (runs.length === 0) {
    return <EmptyState title="No safe QA runs" body="Start a safe QA run to generate a discovery-aware plan and combined report." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>QA Run</th>
            <th>Discovery</th>
            <th>Quality</th>
            <th>Plan</th>
            <th>Execution</th>
            <th>Created</th>
            <th>Report</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id}>
              <td>
                <StatusBadge status={run.status} />
                {run.error_message && <p className="muted">{run.error_message}</p>}
              </td>
              <td>{shortID(run.id)}</td>
              <td>{run.discovery_run_id ? <a href={`#/discovery-runs/${run.discovery_run_id}`}>{shortID(run.discovery_run_id)}</a> : "Pending"}</td>
              <td>{run.quality_check_run_id ? <a href={`#/quality-check-runs/${run.quality_check_run_id}`}>{shortID(run.quality_check_run_id)}</a> : "Not included"}</td>
              <td>{run.test_plan_id ? <a href={`#/test-plans/${run.test_plan_id}`}>{shortID(run.test_plan_id)}</a> : "Pending"}</td>
              <td>
                {run.test_plan_execution_id ? (
                  <a href={`#/test-plan-executions/${run.test_plan_execution_id}`}>{shortID(run.test_plan_execution_id)}</a>
                ) : (
                  "Preview only"
                )}
              </td>
              <td>{formatDate(run.created_at)}</td>
              <td>
                <div className="button-row compact">
                  <a href={`#/qa-runs/${run.id}`}>Open</a>
                  <a href={qaRunHTMLReportURL(run.id)} target="_blank" rel="noreferrer">
                    HTML
                  </a>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function CredentialProfileForm({ project, onSaved }: { project: Project; onSaved: () => void }) {
  const [form, setForm] = useState<CredentialProfileInput>({
    name: "Demo Login",
    type: "username_password",
    role_name: "",
    role_description: "",
    subject_label: "",
    username: "",
    password: "",
    login_url: project.frontend_url ? new URL("/login", project.frontend_url).toString() : "",
    username_selector: "#username",
    password_selector: "#password",
    submit_selector: "button[type=submit]",
    success_url_contains: "/dashboard",
    success_text_contains: "Welcome to the Qualora demo dashboard",
    failure_text_contains: "Invalid credentials",
    post_login_wait_ms: 250,
    is_default: true
  });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setMessage("");
    setError("");
    try {
      const saved = await createCredentialProfile(project.id, {
        ...form,
        name: form.name.trim(),
        role_name: form.role_name?.trim(),
        role_description: form.role_description?.trim(),
        subject_label: form.subject_label?.trim(),
        username: form.username?.trim(),
        login_url: form.login_url.trim(),
        username_selector: form.username_selector.trim(),
        password_selector: form.password_selector.trim(),
        submit_selector: form.submit_selector.trim(),
        success_url_contains: form.success_url_contains.trim(),
        success_text_contains: form.success_text_contains.trim(),
        failure_text_contains: form.failure_text_contains.trim(),
        post_login_wait_ms: Number(form.post_login_wait_ms || 0)
      });
      setMessage(`Saved credential profile ${saved.name}.`);
      setForm({ ...form, username: "", password: "", is_default: false });
      onSaved();
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : String(saveError));
    } finally {
      setSaving(false);
    }
  }

  return (
    <form className="project-form credential-form" onSubmit={(event) => void submit(event)}>
      {error && <Notice tone="danger" message={error} />}
      {message && <Notice tone="info" message={message} />}
      <div className="form-grid two">
        <label>
          Name
          <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
        </label>
        <label>
          Login URL
          <input value={form.login_url} onChange={(event) => setForm({ ...form, login_url: event.target.value })} required />
        </label>
        <label>
          Role name
          <input value={form.role_name || ""} placeholder="admin, readonly, customer-a" onChange={(event) => setForm({ ...form, role_name: event.target.value })} />
        </label>
        <label>
          Subject label
          <input value={form.subject_label || ""} placeholder="Demo Admin" onChange={(event) => setForm({ ...form, subject_label: event.target.value })} />
        </label>
        <label>
          Username
          <input value={form.username || ""} onChange={(event) => setForm({ ...form, username: event.target.value })} required />
        </label>
        <label>
          Password
          <input type="password" value={form.password || ""} onChange={(event) => setForm({ ...form, password: event.target.value })} required />
        </label>
        <label>
          Username selector
          <input value={form.username_selector} onChange={(event) => setForm({ ...form, username_selector: event.target.value })} required />
        </label>
        <label>
          Password selector
          <input value={form.password_selector} onChange={(event) => setForm({ ...form, password_selector: event.target.value })} required />
        </label>
        <label>
          Submit selector
          <input value={form.submit_selector} onChange={(event) => setForm({ ...form, submit_selector: event.target.value })} required />
        </label>
        <label>
          Post-login wait ms
          <input
            type="number"
            min="0"
            max="30000"
            value={form.post_login_wait_ms}
            onChange={(event) => setForm({ ...form, post_login_wait_ms: Number(event.target.value) })}
          />
        </label>
      </div>
      <div className="form-grid three">
        <label>
          Role description
          <input value={form.role_description || ""} onChange={(event) => setForm({ ...form, role_description: event.target.value })} />
        </label>
        <label>
          Success URL contains
          <input value={form.success_url_contains} onChange={(event) => setForm({ ...form, success_url_contains: event.target.value })} />
        </label>
        <label>
          Success text contains
          <input value={form.success_text_contains} onChange={(event) => setForm({ ...form, success_text_contains: event.target.value })} />
        </label>
        <label>
          Failure text contains
          <input value={form.failure_text_contains} onChange={(event) => setForm({ ...form, failure_text_contains: event.target.value })} />
        </label>
      </div>
      <label className="checkbox-row">
        <input type="checkbox" checked={form.is_default} onChange={(event) => setForm({ ...form, is_default: event.target.checked })} />
        Default profile
      </label>
      <p className="muted">Credentials are encrypted at rest and are never sent to AI.</p>
      <div className="form-actions">
        <button type="submit" disabled={saving}>
          {saving ? "Saving" : "Add Credential Profile"}
        </button>
      </div>
    </form>
  );
}

function CredentialProfileTable({
  profiles,
  onChanged,
  onTestRun,
  onAuthenticatedRun
}: {
  profiles: CredentialProfile[];
  onChanged: () => void;
  onTestRun: (run: TestRun) => void;
  onAuthenticatedRun: (profileID: string) => void;
}) {
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");

  if (profiles.length === 0) {
    return <EmptyState title="No credential profiles" body="Create a credential profile to test a deterministic login flow." />;
  }

  async function runAction(profile: CredentialProfile, action: "test" | "auth" | "default" | "edit" | "delete") {
    setBusy(`${action}:${profile.id}`);
    setError("");
    try {
      if (action === "test") {
        const run = await testCredentialProfileLogin(profile.id);
        onTestRun(run);
      } else if (action === "auth") {
        onAuthenticatedRun(profile.id);
      } else if (action === "default") {
        await updateCredentialProfile(profile.id, profileInputFromProfile(profile, { is_default: true }));
        onChanged();
      } else if (action === "edit") {
        const next = promptProfileEdit(profile);
        if (next) {
          await updateCredentialProfile(profile.id, next);
          onChanged();
        }
      } else if (action === "delete") {
        if (window.confirm(`Delete credential profile ${profile.name}?`)) {
          await deleteCredentialProfile(profile.id);
          onChanged();
        }
      }
    } catch (actionError) {
      setError(actionError instanceof Error ? actionError.message : String(actionError));
    } finally {
      setBusy("");
    }
  }

  return (
    <div>
      {error && <Notice tone="danger" message={error} />}
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Role</th>
            <th>Type</th>
            <th>Configured</th>
            <th>Login URL</th>
            <th>Default</th>
            <th>Updated</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {profiles.map((profile) => (
            <tr key={profile.id}>
              <td>
                <strong>{profile.name}</strong>
                {profile.username_display_hint && <div className="muted">{profile.username_display_hint}</div>}
              </td>
              <td>
                {profile.role_name || "Not set"}
                {profile.subject_label && <div className="muted">{profile.subject_label}</div>}
              </td>
              <td>{profile.type}</td>
              <td>
                Username {profile.username_configured ? "yes" : "no"} · Password {profile.password_configured ? "yes" : "no"}
              </td>
              <td>
                <code>{profile.login_url}</code>
              </td>
              <td>{profile.is_default ? "Default" : "No"}</td>
              <td>{formatDate(profile.updated_at)}</td>
              <td>
                <div className="button-row compact">
                  <button type="button" className="secondary" disabled={busy !== ""} onClick={() => void runAction(profile, "test")}>
                    {busy === `test:${profile.id}` ? "Testing" : "Test login"}
                  </button>
                  <button type="button" className="secondary" disabled={busy !== ""} onClick={() => void runAction(profile, "auth")}>
                    Auth smoke
                  </button>
                  {!profile.is_default && (
                    <button type="button" className="secondary" disabled={busy !== ""} onClick={() => void runAction(profile, "default")}>
                      Set default
                    </button>
                  )}
                  <button type="button" className="secondary" disabled={busy !== ""} onClick={() => void runAction(profile, "edit")}>
                    Edit
                  </button>
                  <button type="button" className="secondary" disabled={busy !== ""} onClick={() => void runAction(profile, "delete")}>
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

function profileInputFromProfile(profile: CredentialProfile, overrides: Partial<CredentialProfileInput> = {}): CredentialProfileInput {
  return {
    name: profile.name,
    type: "username_password",
    role_name: profile.role_name || "",
    role_description: profile.role_description || "",
    subject_label: profile.subject_label || "",
    login_url: profile.login_url,
    username_selector: profile.username_selector,
    password_selector: profile.password_selector,
    submit_selector: profile.submit_selector,
    success_url_contains: profile.success_url_contains || "",
    success_text_contains: profile.success_text_contains || "",
    failure_text_contains: profile.failure_text_contains || "",
    post_login_wait_ms: profile.post_login_wait_ms || 0,
    is_default: profile.is_default,
    ...overrides
  };
}

function promptProfileEdit(profile: CredentialProfile): CredentialProfileInput | null {
  const name = window.prompt("Credential profile name", profile.name);
  if (name === null) {
    return null;
  }
  const loginURL = window.prompt("Login URL", profile.login_url);
  if (loginURL === null) {
    return null;
  }
  const usernameSelector = window.prompt("Username selector", profile.username_selector);
  if (usernameSelector === null) {
    return null;
  }
  const passwordSelector = window.prompt("Password selector", profile.password_selector);
  if (passwordSelector === null) {
    return null;
  }
  const submitSelector = window.prompt("Submit selector", profile.submit_selector);
  if (submitSelector === null) {
    return null;
  }
  const successURL = window.prompt("Success URL contains", profile.success_url_contains || "");
  if (successURL === null) {
    return null;
  }
  const successText = window.prompt("Success text contains", profile.success_text_contains || "");
  if (successText === null) {
    return null;
  }
  const failureText = window.prompt("Failure text contains", profile.failure_text_contains || "");
  if (failureText === null) {
    return null;
  }
  return profileInputFromProfile(profile, {
    name,
    login_url: loginURL,
    username_selector: usernameSelector,
    password_selector: passwordSelector,
    submit_selector: submitSelector,
    success_url_contains: successURL,
    success_text_contains: successText,
    failure_text_contains: failureText
  });
}

function AuthorizationCheckForm({
  project,
  profiles,
  onSaved
}: {
  project: Project;
  profiles: CredentialProfile[];
  onSaved: () => void;
}) {
  const [form, setForm] = useState<AuthorizationCheckInput>({
    name: "Admin route is denied for readonly",
    description: "",
    type: "browser_url",
    resource_label: "Admin area",
    owner_credential_profile_id: "",
    actor_credential_profile_id: profiles[0]?.id || "",
    expected_outcome: "denied",
    target_url: project.frontend_url ? new URL("/admin", project.frontend_url).toString() : "/admin",
    success_text_contains: "",
    denied_text_contains: "Access denied",
    enabled: true
  });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!form.actor_credential_profile_id && profiles[0]?.id) {
      setForm((current) => ({ ...current, actor_credential_profile_id: profiles[0].id }));
    }
  }, [form.actor_credential_profile_id, profiles]);

  if (profiles.length === 0) {
    return <Notice tone="info" message="Create credential profiles before adding role-aware authorization checks." />;
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setMessage("");
    setError("");
    try {
      const saved = await createAuthorizationCheck(project.id, {
        ...form,
        name: form.name.trim(),
        description: form.description?.trim(),
        resource_label: form.resource_label?.trim(),
        owner_credential_profile_id: form.owner_credential_profile_id || undefined,
        target_url: form.target_url.trim(),
        success_text_contains: form.success_text_contains?.trim(),
        denied_text_contains: form.denied_text_contains?.trim(),
        enabled: form.enabled !== false
      });
      setMessage(`Saved authorization check ${saved.name}.`);
      setForm({ ...form, name: "", description: "", resource_label: "", enabled: true });
      onSaved();
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : String(saveError));
    } finally {
      setSaving(false);
    }
  }

  return (
    <form className="project-form authorization-form" onSubmit={(event) => void submit(event)}>
      {error && <Notice tone="danger" message={error} />}
      {message && <Notice tone="info" message={message} />}
      <div className="form-grid two">
        <label>
          Name
          <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
        </label>
        <label>
          Actor credential profile
          <select
            value={form.actor_credential_profile_id}
            onChange={(event) => setForm({ ...form, actor_credential_profile_id: event.target.value })}
            required
          >
            {profiles.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {credentialProfileLabel(profile)}
              </option>
            ))}
          </select>
        </label>
        <label>
          Expected outcome
          <select
            value={form.expected_outcome}
            onChange={(event) => setForm({ ...form, expected_outcome: event.target.value as AuthorizationCheckInput["expected_outcome"] })}
          >
            <option value="allowed">Allowed</option>
            <option value="denied">Denied</option>
          </select>
        </label>
        <label>
          Target URL or path
          <input value={form.target_url} onChange={(event) => setForm({ ...form, target_url: event.target.value })} required />
        </label>
        <label>
          Resource label
          <input value={form.resource_label || ""} onChange={(event) => setForm({ ...form, resource_label: event.target.value })} />
        </label>
        <label>
          Owner credential profile
          <select
            value={form.owner_credential_profile_id || ""}
            onChange={(event) => setForm({ ...form, owner_credential_profile_id: event.target.value })}
          >
            <option value="">Not set</option>
            {profiles.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {credentialProfileLabel(profile)}
              </option>
            ))}
          </select>
        </label>
        <label>
          Success text contains
          <input value={form.success_text_contains || ""} onChange={(event) => setForm({ ...form, success_text_contains: event.target.value })} />
        </label>
        <label>
          Denied text contains
          <input value={form.denied_text_contains || ""} onChange={(event) => setForm({ ...form, denied_text_contains: event.target.value })} />
        </label>
      </div>
      <label>
        Description
        <textarea value={form.description || ""} onChange={(event) => setForm({ ...form, description: event.target.value })} />
      </label>
      <label className="checkbox-row">
        <input type="checkbox" checked={form.enabled !== false} onChange={(event) => setForm({ ...form, enabled: event.target.checked })} />
        Enabled
      </label>
      <p className="muted">Use dedicated test accounts and test data. Do not use real user credentials.</p>
      <div className="form-actions">
        <button type="submit" disabled={saving || profiles.length === 0}>
          {saving ? "Saving" : "Add Authorization Check"}
        </button>
      </div>
    </form>
  );
}

function AuthorizationCheckTable({
  checks,
  profiles,
  onChanged
}: {
  checks: AuthorizationCheck[];
  profiles: CredentialProfile[];
  onChanged: () => void;
}) {
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const profileByID = useMemo(() => new Map(profiles.map((profile) => [profile.id, profile])), [profiles]);

  if (checks.length === 0) {
    return <EmptyState title="No authorization checks" body="Add an explicit browser URL check to verify allowed or denied role access." />;
  }

  async function toggleCheck(check: AuthorizationCheck) {
    setBusy(`toggle:${check.id}`);
    setError("");
    try {
      await updateAuthorizationCheck(check.id, authorizationInputFromCheck(check, { enabled: !check.enabled }));
      onChanged();
    } catch (toggleError) {
      setError(toggleError instanceof Error ? toggleError.message : String(toggleError));
    } finally {
      setBusy("");
    }
  }

  async function removeCheck(check: AuthorizationCheck) {
    if (!window.confirm(`Delete authorization check ${check.name}?`)) {
      return;
    }
    setBusy(`delete:${check.id}`);
    setError("");
    try {
      await deleteAuthorizationCheck(check.id);
      onChanged();
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : String(deleteError));
    } finally {
      setBusy("");
    }
  }

  return (
    <div>
      {error && <Notice tone="danger" message={error} />}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Type</th>
              <th>Actor</th>
              <th>Expected</th>
              <th>Target</th>
              <th>Enabled</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {checks.map((check) => {
              const actor = profileByID.get(check.actor_credential_profile_id);
              return (
                <tr key={check.id}>
                  <td>
                    <strong>{check.name}</strong>
                    {check.resource_label && <div className="muted">{check.resource_label}</div>}
                  </td>
                  <td>{check.type}</td>
                  <td>{actor ? credentialProfileLabel(actor) : check.actor_credential_profile_id}</td>
                  <td>{check.expected_outcome}</td>
                  <td>
                    <code>{check.target_url || check.path}</code>
                  </td>
                  <td>{check.enabled ? "Yes" : "No"}</td>
                  <td className="actions">
                    <div className="button-row compact">
                      <button type="button" className="secondary" disabled={busy !== ""} onClick={() => void toggleCheck(check)}>
                        {check.enabled ? "Disable" : "Enable"}
                      </button>
                      <button type="button" className="secondary danger" disabled={busy !== ""} onClick={() => void removeCheck(check)}>
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
    </div>
  );
}

function AuthorizationRunTable({ runs }: { runs: AuthorizationCheckRun[] }) {
  if (runs.length === 0) {
    return <EmptyState title="No authorization runs" body="Run authorization checks to create a report." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Run</th>
            <th>Checks</th>
            <th>Created</th>
            <th>Completed</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id}>
              <td>
                <StatusBadge status={run.status} />
              </td>
              <td>
                <a href={`#/authorization-check-runs/${run.id}`}>{shortID(run.id)}</a>
                {run.error_message && <p className="muted">{run.error_message}</p>}
              </td>
              <td>
                {run.passed_checks} passed · {run.failed_checks} failed · {run.skipped_checks} skipped
              </td>
              <td>{formatDate(run.created_at)}</td>
              <td>{run.completed_at ? formatDate(run.completed_at) : "Not completed"}</td>
              <td className="actions">
                <a className="button secondary-link" href={authorizationCheckHTMLReportURL(run.id)} target="_blank" rel="noreferrer">
                  HTML
                </a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function authorizationInputFromCheck(check: AuthorizationCheck, overrides: Partial<AuthorizationCheckInput> = {}): AuthorizationCheckInput {
  return {
    name: check.name,
    description: check.description || "",
    type: "browser_url",
    resource_label: check.resource_label || "",
    owner_credential_profile_id: check.owner_credential_profile_id || undefined,
    actor_credential_profile_id: check.actor_credential_profile_id,
    expected_outcome: check.expected_outcome,
    target_url: check.target_url || check.path || "",
    success_text_contains: check.success_text_contains || "",
    denied_text_contains: check.denied_text_contains || "",
    enabled: check.enabled,
    ...overrides
  };
}

function credentialProfileLabel(profile: CredentialProfile): string {
  return [profile.name, profile.role_name ? `role ${profile.role_name}` : "", profile.username_display_hint || ""].filter(Boolean).join(" · ");
}

function APISpecImportForm({ project, onImported }: { project: Project; onImported: () => void }) {
  const [form, setForm] = useState({
    name: "Demo OpenAPI Spec",
    source_type: "url" as APISpecImportInput["source_type"],
    source_url: project.openapi_url || "",
    raw_spec: ""
  });
  const [importing, setImporting] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setImporting(true);
    setMessage("");
    setError("");
    const payload: APISpecImportInput = {
      name: form.name.trim(),
      source_type: form.source_type,
      source_url: form.source_type === "url" ? form.source_url.trim() : undefined,
      raw_spec: form.source_type === "url" ? undefined : form.raw_spec.trim()
    };
    try {
      const detail = await importAPISpec(project.id, payload);
      setMessage(`Imported ${detail.spec.operation_count} operations. ${detail.spec.safe_operation_count} are safe to execute.`);
      onImported();
    } catch (importError) {
      setError(importError instanceof Error ? importError.message : String(importError));
    } finally {
      setImporting(false);
    }
  }

  return (
    <form className="project-form api-spec-form" onSubmit={(event) => void submit(event)}>
      {error && <Notice tone="danger" message={error} />}
      {message && <Notice tone="info" message={message} />}
      <div className="form-grid two">
        <label>
          Name
          <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
        </label>
        <label>
          Source Type
          <select value={form.source_type} onChange={(event) => setForm({ ...form, source_type: event.target.value as APISpecImportInput["source_type"] })}>
            <option value="url">OpenAPI URL</option>
            <option value="inline">Inline JSON/YAML</option>
            <option value="demo">Demo inline spec</option>
          </select>
        </label>
      </div>
      {form.source_type === "url" ? (
        <label>
          OpenAPI URL
          <input
            value={form.source_url}
            placeholder={project.openapi_url || "http://demo-api:8080/openapi.yaml"}
            onChange={(event) => setForm({ ...form, source_url: event.target.value })}
            required
          />
        </label>
      ) : (
        <label>
          OpenAPI JSON/YAML
          <textarea
            className="spec-textarea"
            value={form.raw_spec}
            placeholder="openapi: 3.0.3&#10;info:&#10;  title: Demo API&#10;paths: {}"
            onChange={(event) => setForm({ ...form, raw_spec: event.target.value })}
            required
          />
        </label>
      )}
      <div className="form-actions">
        <button type="submit" disabled={importing}>
          {importing ? "Importing" : "Import OpenAPI Spec"}
        </button>
      </div>
    </form>
  );
}

function APISpecTable({ specs, onDeleted }: { specs: APISpec[]; onDeleted: () => void }) {
  if (specs.length === 0) {
    return <EmptyState title="No API specs" body="Import an OpenAPI document to inspect operations and run safe API smoke checks." />;
  }

  async function removeSpec(apiSpecID: string) {
    if (!window.confirm("Delete this API spec and its discovered operations?")) {
      return;
    }
    await deleteAPISpec(apiSpecID);
    onDeleted();
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Name</th>
            <th>Operations</th>
            <th>Source</th>
            <th>Imported</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {specs.map((spec) => (
            <tr key={spec.id}>
              <td>
                <StatusBadge status={spec.status} />
              </td>
              <td>
                <a href={`#/api-specs/${spec.id}`}>{spec.name}</a>
                {spec.error_message && <p className="muted">{spec.error_message}</p>}
              </td>
              <td>
                {spec.operation_count} total · {spec.safe_operation_count} safe · {spec.skipped_operation_count} skipped
              </td>
              <td>{spec.source_type}</td>
              <td>{formatDate(spec.created_at)}</td>
              <td className="actions">
                <div className="button-row compact">
                  <a className="button secondary-link" href={`#/api-specs/${spec.id}`}>
                    Open
                  </a>
                  <button type="button" className="secondary danger" onClick={() => void removeSpec(spec.id)}>
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

function APISpecPage({
  apiSpecID,
  projectByID,
  onOpenRun
}: {
  apiSpecID: string;
  projectByID: Map<string, Project>;
  onOpenRun: (runID: string) => void;
}) {
  const [detail, setDetail] = useState<APISpecDetail | undefined>();
  const [project, setProject] = useState<Project | undefined>();
  const [running, setRunning] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextDetail = await getAPISpec(apiSpecID);
      const nextProject = projectByID.get(nextDetail.spec.project_id) ?? (await getProject(nextDetail.spec.project_id));
      setDetail(nextDetail);
      setProject(nextProject);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [apiSpecID, projectByID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function runSmoke() {
    if (!detail) {
      return;
    }
    setRunning(true);
    setError("");
    try {
      const run = await startAPISmokeRun(detail.spec.id);
      onOpenRun(run.id);
    } catch (runError) {
      setError(runError instanceof Error ? runError.message : String(runError));
    } finally {
      setRunning(false);
    }
  }

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading || !detail || !project) {
    return <SkeletonRows />;
  }

  const operations = detail.operations || [];

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{detail.spec.name}</h2>
            <p>
              <StatusBadge status={detail.spec.status} /> <span className="muted">{detail.spec.parsed_title || "OpenAPI 3.x document"}</span>
            </p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/projects/${project.id}`}>
              Project
            </a>
            <button type="button" disabled={running || detail.spec.status !== "parsed"} onClick={() => void runSmoke()}>
              {running ? "Running" : "Run safe API smoke test"}
            </button>
          </div>
        </div>
        {detail.spec.error_message && <Notice tone="danger" message={detail.spec.error_message} />}
        <Notice
          tone="info"
          message="Only safe read-only operations are executed. Mutating, authenticated, ambiguous, or unsafe operations are skipped."
        />
        <div className="summary-grid">
          <Metric label="Operations" value={detail.spec.operation_count} />
          <Metric label="Safe" value={detail.spec.safe_operation_count} />
          <Metric label="Skipped" value={detail.spec.skipped_operation_count} tone="medium" />
        </div>
        <div className="detail-grid compact">
          <Field label="Project" value={project.name} />
          <Field label="Title" value={detail.spec.parsed_title || "Not provided"} />
          <Field label="Version" value={detail.spec.parsed_version || "Not provided"} />
          <Field label="Server URL" value={detail.spec.server_url || project.api_base_url || "Not provided"} />
        </div>
      </section>

      <section>
        <h2>Operations</h2>
        <APIOperationTable operations={operations} />
      </section>
    </div>
  );
}

function APIOperationTable({ operations }: { operations: APIOperation[] }) {
  if (operations.length === 0) {
    return <EmptyState title="No operations" body="No OpenAPI operations were discovered for this spec." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Safe</th>
            <th>Method</th>
            <th>Path</th>
            <th>Summary</th>
            <th>Tags</th>
            <th>Skip Reason</th>
          </tr>
        </thead>
        <tbody>
          {operations.map((operation) => (
            <tr key={operation.id}>
              <td>
                <StatusBadge status={operation.safe_to_execute ? "passed" : "skipped"} />
              </td>
              <td>
                <code>{operation.method}</code>
              </td>
              <td>
                <code>{operation.path}</code>
                {operation.query_string && <p className="muted">?{operation.query_string}</p>}
              </td>
              <td>{operation.summary || operation.operation_id || "Not provided"}</td>
              <td>{operation.tags.length ? operation.tags.join(", ") : "None"}</td>
              <td>{operation.skip_reason || ""}</td>
            </tr>
          ))}
        </tbody>
      </table>
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
  discoveryRuns = [],
  providers,
  onGenerated
}: {
  project: Project;
  runs: TestRun[];
  discoveryRuns?: DiscoveryRun[];
  providers: AIProvider[];
  onGenerated: (plan: TestPlan) => Promise<void>;
}) {
  const [providerID, setProviderID] = useState("");
  const [runID, setRunID] = useState("");
  const [discoveryRunID, setDiscoveryRunID] = useState("");
  const [useLatestDiscovery, setUseLatestDiscovery] = useState(false);
  const [executionMode, setExecutionMode] = useState<"review_only" | "safe_executable">("review_only");
  const [productContext, setProductContext] = useState("");
  const [focusAreas, setFocusAreas] = useState<string[]>(["smoke", "functional", "negative", "accessibility", "regression"]);
  const [maxScenarios, setMaxScenarios] = useState(10);
  const [maxPagesFromDiscovery, setMaxPagesFromDiscovery] = useState(20);
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
      discovery_run_id: discoveryRunID || undefined,
      use_latest_discovery: useLatestDiscovery || undefined,
      include_discovery_map: discoveryRunID || useLatestDiscovery ? true : undefined,
      execution_mode: executionMode,
      max_pages_from_discovery: maxPagesFromDiscovery,
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
        <label>
          Discovery Map
          <select
            value={discoveryRunID}
            onChange={(event) => {
              setDiscoveryRunID(event.target.value);
              if (event.target.value) {
                setUseLatestDiscovery(false);
                setExecutionMode("safe_executable");
              }
            }}
          >
            <option value="">No explicit discovery map</option>
            {discoveryRuns
              .filter((run) => run.status === "completed")
              .map((run) => (
                <option key={run.id} value={run.id}>
                  {shortID(run.id)} · {run.total_pages} pages · {formatDate(run.created_at)}
                </option>
              ))}
          </select>
        </label>
        <label>
          Execution Mode
          <select value={executionMode} onChange={(event) => setExecutionMode(event.target.value as "review_only" | "safe_executable")}>
            <option value="review_only">Review only</option>
            <option value="safe_executable">Safe executable candidates</option>
          </select>
        </label>
      </div>
      <div className="form-grid two">
        <label className="check-row">
          <input
            type="checkbox"
            checked={useLatestDiscovery}
            onChange={(event) => {
              setUseLatestDiscovery(event.target.checked);
              if (event.target.checked) {
                setDiscoveryRunID("");
                setExecutionMode("safe_executable");
              }
            }}
          />
          Use latest completed discovery map
        </label>
        <label>
          Max Discovery Pages
          <input
            type="number"
            min="1"
            max="100"
            value={maxPagesFromDiscovery}
            onChange={(event) => setMaxPagesFromDiscovery(Number(event.target.value))}
          />
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
            <th>Source</th>
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
              <td>{plan.source_type || "run_report"}</td>
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

function TestPlanDetailPage({
  testPlanID,
  projectByID,
  onOpenExecution
}: {
  testPlanID: string;
  projectByID: Map<string, Project>;
  onOpenExecution: (executionID: string) => void;
}) {
  const [plan, setPlan] = useState<TestPlan | undefined>();
  const [project, setProject] = useState<Project | undefined>();
  const [executions, setExecutions] = useState<LoadState<TestPlanExecution[]>>({ data: [], loading: true, error: "" });
  const [preview, setPreview] = useState<TestPlanExecutionPreview | undefined>();
  const [executionInput, setExecutionInput] = useState({ max_scenarios: 5, max_steps_per_scenario: 10 });
  const [executionBusy, setExecutionBusy] = useState("");
  const [executionError, setExecutionError] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setExecutions((current) => ({ ...current, loading: true, error: "" }));
    setError("");
    try {
      const nextPlan = await getTestPlan(testPlanID);
      const [nextProject, nextExecutions] = await Promise.all([
        projectByID.get(nextPlan.project_id) ?? getProject(nextPlan.project_id),
        listTestPlanExecutions(testPlanID)
      ]);
      setPlan(nextPlan);
      setProject(nextProject);
      setExecutions({ data: nextExecutions, loading: false, error: "" });
    } catch (loadError) {
      const message = loadError instanceof Error ? loadError.message : String(loadError);
      setError(message);
      setExecutions((current) => ({ ...current, loading: false, error: message }));
    } finally {
      setLoading(false);
    }
  }, [projectByID, testPlanID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!executions.data.some((execution) => isActiveRunStatus(execution.status))) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [executions.data, refresh]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading || !plan || !project) {
    return <SkeletonRows />;
  }

  const payload = normalizeTestPlanPayload(plan.plan_json);
  const currentPlan = plan;
  const canExecutePlan = plan.status === "completed" && Boolean(project.frontend_url);
  const requestPayload = (): TestPlanExecutionRequest => ({
    max_scenarios: executionInput.max_scenarios,
    max_steps_per_scenario: executionInput.max_steps_per_scenario,
    dry_run: false
  });

  async function previewSafeExecution() {
    setExecutionBusy("preview");
    setExecutionError("");
    try {
      const nextPreview = await previewTestPlanExecution(currentPlan.id, { ...requestPayload(), dry_run: true });
      setPreview(nextPreview);
    } catch (previewError) {
      setExecutionError(previewError instanceof Error ? previewError.message : String(previewError));
    } finally {
      setExecutionBusy("");
    }
  }

  async function executeSafePlan() {
    setExecutionBusy("execute");
    setExecutionError("");
    try {
      const detail = await executeTestPlan(currentPlan.id, requestPayload());
      setPreview(undefined);
      await refresh();
      onOpenExecution(detail.execution.id);
    } catch (executeError) {
      setExecutionError(executeError instanceof Error ? executeError.message : String(executeError));
    } finally {
      setExecutionBusy("");
    }
  }

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
          <Field label="Source" value={plan.source_type || "run_report"} />
          <Field label="Discovery" value={plan.discovery_run_id ? shortID(plan.discovery_run_id) : "Not linked"} />
          <Field label="Created" value={formatDate(plan.created_at)} />
          <Field label="Updated" value={formatDate(plan.updated_at)} />
        </div>
        {plan.error_message && <Notice tone="danger" message={plan.error_message} />}
      </section>

      <section>
        <div className="section-heading">
          <div>
            <h2>Approved Safe Execution</h2>
            <p>Preview or run supported safe browser checks from this plan.</p>
          </div>
          <button type="button" className="secondary" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
        <Notice
          tone="info"
          message="Execution is limited to the Qualora safe DSL. Login, form submit, mutation, upload, admin, destructive, and unsupported steps are skipped with reasons."
        />
        {!canExecutePlan && (
          <Notice tone="danger" message="This plan needs completed status and a project frontend URL before safe browser execution is available." />
        )}
        {executionError && <Notice tone="danger" message={executionError} />}
        <div className="form-grid two execution-controls">
          <label>
            Max Scenarios
            <input
              type="number"
              min="1"
              max="20"
              value={executionInput.max_scenarios}
              onChange={(event) => setExecutionInput({ ...executionInput, max_scenarios: Number(event.target.value) })}
            />
          </label>
          <label>
            Max Steps Per Scenario
            <input
              type="number"
              min="1"
              max="30"
              value={executionInput.max_steps_per_scenario}
              onChange={(event) => setExecutionInput({ ...executionInput, max_steps_per_scenario: Number(event.target.value) })}
            />
          </label>
        </div>
        <div className="form-actions">
          <button type="button" className="secondary" disabled={!canExecutePlan || executionBusy !== ""} onClick={() => void previewSafeExecution()}>
            {executionBusy === "preview" ? "Previewing" : "Preview safe execution"}
          </button>
          <button type="button" disabled={!canExecutePlan || executionBusy !== ""} onClick={() => void executeSafePlan()}>
            {executionBusy === "execute" ? "Starting" : "Execute safe plan"}
          </button>
        </div>
        {preview && <ExecutionPreview preview={preview} />}
        <div className="section-split">
          <h3>Executions</h3>
          {executions.error && <Notice tone="danger" message={executions.error} />}
          {executions.loading ? (
            <SkeletonRows />
          ) : (
            <TestPlanExecutionTable executions={executions.data} onOpen={onOpenExecution} />
          )}
        </div>
      </section>

      <section>
        <h2>Coverage</h2>
        <div className="summary-grid">
          <Metric label="Executable Scenarios" value={plan.execution_coverage?.executable_scenarios ?? 0} />
          <Metric label="Skipped Scenarios" value={plan.execution_coverage?.skipped_scenarios ?? 0} tone="medium" />
          <Metric label="Executable Steps" value={plan.execution_coverage?.executable_steps ?? 0} />
          <Metric label="Skipped Steps" value={plan.execution_coverage?.skipped_steps ?? 0} tone="medium" />
          <Metric label="Unsafe Skips" value={plan.execution_coverage?.unsafe_skipped_steps ?? 0} tone="high" />
          <Metric label="Unsupported Skips" value={plan.execution_coverage?.unsupported_skipped_steps ?? 0} tone="info" />
        </div>
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

function ExecutionPreview({ preview }: { preview: TestPlanExecutionPreview }) {
  return (
    <div className="execution-preview">
      <div className="summary-grid">
        <Metric label="Executable Scenarios" value={preview.executable_scenarios} />
        <Metric label="Skipped Scenarios" value={preview.skipped_scenarios} tone="medium" />
        <Metric label="Executable Steps" value={preview.executable_steps} />
        <Metric label="Skipped Steps" value={preview.skipped_steps} tone="medium" />
        <Metric label="Unsafe Skips" value={preview.safety_summary.skipped_unsafe_steps} tone="high" />
        <Metric label="Unsupported Skips" value={preview.safety_summary.skipped_unsupported_steps} tone="info" />
      </div>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Scenario</th>
              <th>Status</th>
              <th>Step</th>
              <th>Action</th>
              <th>Target</th>
              <th>Reason</th>
            </tr>
          </thead>
          <tbody>
            {preview.scenarios.flatMap((scenario) =>
              scenario.steps.map((step) => (
                <tr key={`${scenario.scenario_id_from_plan}-${step.step_order}`}>
                  <td>{scenario.name}</td>
                  <td>
                    <StatusBadge status={step.status} />
                  </td>
                  <td>{step.step_order}</td>
                  <td>
                    <code>{step.mapped_action || step.original_action}</code>
                  </td>
                  <td>
                    <code>{step.target || ""}</code>
                  </td>
                  <td>{step.skip_reason || scenario.skip_reason || ""}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function TestPlanExecutionTable({ executions, onOpen }: { executions: TestPlanExecution[]; onOpen: (executionID: string) => void }) {
  if (executions.length === 0) {
    return <EmptyState title="No executions" body="Preview or execute this plan to create an approved safe execution record." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Execution</th>
            <th>Scenarios</th>
            <th>Steps</th>
            <th>Created</th>
            <th>Completed</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {executions.map((execution) => (
            <tr key={execution.id}>
              <td>
                <StatusBadge status={execution.status} />
              </td>
              <td>
                <button type="button" className="link-button" onClick={() => onOpen(execution.id)}>
                  {shortID(execution.id)}
                </button>
                {execution.error_message && <p className="muted">{execution.error_message}</p>}
              </td>
              <td>
                {execution.passed_scenarios} passed · {execution.failed_scenarios} failed · {execution.skipped_scenarios} skipped
              </td>
              <td>
                {execution.passed_steps} passed · {execution.failed_steps} failed · {execution.skipped_steps} skipped
              </td>
              <td>{formatDate(execution.created_at)}</td>
              <td>{execution.completed_at ? formatDate(execution.completed_at) : "Not completed"}</td>
              <td className="actions">
                <div className="button-row compact">
                  <button type="button" className="secondary" onClick={() => onOpen(execution.id)}>
                    Open
                  </button>
                  <a className="button secondary-link" href={testPlanExecutionHTMLReportURL(execution.id)} target="_blank" rel="noreferrer">
                    HTML
                  </a>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function TestPlanExecutionPage({ executionID }: { executionID: string }) {
  const [report, setReport] = useState<TestPlanExecutionReport | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextReport = await getTestPlanExecutionReport(executionID);
      setReport(nextReport);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [executionID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!report || !isActiveRunStatus(report.execution.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, report]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading && !report) {
    return <SkeletonRows />;
  }
  if (!report) {
    return <Notice tone="danger" message="Test plan execution report could not be loaded." />;
  }

  const summary = summarizeFindingsForUI(report.findings);

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{report.test_plan.title}</h2>
            <p>
              <StatusBadge status={report.execution.status} /> <span className="muted">Execution {report.execution.id}</span>
            </p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/test-plans/${report.test_plan.id}`}>
              Test Plan
            </a>
            <a className="button secondary-link" href={`#/projects/${report.project.id}`}>
              Project
            </a>
            <a className="button" href={testPlanExecutionHTMLReportURL(report.execution.id)} target="_blank" rel="noreferrer">
              HTML Report
            </a>
          </div>
        </div>
        {report.execution.error_message && <Notice tone="danger" message={report.execution.error_message} />}
        <div className="summary-grid">
          <Metric label="Findings" value={summary.total_findings} />
          <Metric label="Passed Steps" value={report.execution.passed_steps} />
          <Metric label="Failed Steps" value={report.execution.failed_steps} tone="high" />
          <Metric label="Skipped Steps" value={report.execution.skipped_steps} tone="medium" />
          <Metric label="Unsafe Skips" value={report.safety_summary.skipped_unsafe_steps} tone="high" />
          <Metric label="Evidence" value={report.evidence.length} tone="info" />
        </div>
        <div className="detail-grid compact">
          <Field label="Project" value={report.project.name} />
          <Field label="Created" value={formatDate(report.execution.created_at)} />
          <Field label="Started" value={report.execution.started_at ? formatDate(report.execution.started_at) : "Not started"} />
          <Field label="Completed" value={report.execution.completed_at ? formatDate(report.execution.completed_at) : "Not completed"} />
        </div>
      </section>

      <section>
        <h2>Safety Scope</h2>
        <div className="detail-grid compact">
          <Field label="Executed Steps" value={String(report.safety_summary.executed_steps)} />
          <Field label="Skipped Unsafe" value={String(report.safety_summary.skipped_unsafe_steps)} />
          <Field label="Skipped Unsupported" value={String(report.safety_summary.skipped_unsupported_steps)} />
          <Field label="Skipped Scenarios" value={String(report.safety_summary.skipped_scenarios)} />
        </div>
      </section>

      <section>
        <h2>Scenarios and Steps</h2>
        <ExecutionScenarioList scenarios={report.scenarios} />
      </section>

      <section>
        <h2>Findings</h2>
        <ExecutionFindingsTable findings={report.findings} />
      </section>

      <section>
        <h2>Evidence</h2>
        <EvidenceTable evidence={report.evidence} />
      </section>
    </div>
  );
}

function ExecutionScenarioList({ scenarios }: { scenarios: TestPlanExecutionReport["scenarios"] }) {
  if (scenarios.length === 0) {
    return <EmptyState title="No scenarios" body="No execution scenarios were persisted for this run." />;
  }
  return (
    <div className="scenario-stack">
      {scenarios.map((scenario) => (
        <div key={scenario.id} className="scenario-card">
          <div className="scenario-heading">
            <div>
              <h3>{scenario.name}</h3>
              {scenario.skip_reason && <p>{scenario.skip_reason}</p>}
            </div>
            <div className="scenario-badges">
              <span className="pill">{scenario.type || "scenario"}</span>
              <StatusBadge status={scenario.status} />
            </div>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>#</th>
                  <th>Action</th>
                  <th>Target</th>
                  <th>Status</th>
                  <th>Duration</th>
                  <th>Result</th>
                </tr>
              </thead>
              <tbody>
                {(scenario.steps || []).map((step) => (
                  <tr key={step.id}>
                    <td>{step.step_order}</td>
                    <td>
                      <code>{step.mapped_action}</code>
                    </td>
                    <td>
                      <code>{step.target}</code>
                    </td>
                    <td>
                      <StatusBadge status={step.status} />
                    </td>
                    <td>{step.duration_ms === undefined ? "" : `${step.duration_ms}ms`}</td>
                    <td>{step.actual_result || step.error_message || step.skip_reason || ""}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ))}
    </div>
  );
}

function ExecutionFindingsTable({ findings }: { findings: TestPlanExecutionReport["findings"] }) {
  if (findings.length === 0) {
    return <EmptyState title="No findings" body="This execution did not record findings." />;
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
          {findings.map((finding) => (
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

function QualityCheckRunPage({ runID }: { runID: string }) {
  const [report, setReport] = useState<QualityCheckReport | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setReport(await getQualityCheckReport(runID));
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [runID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!report || !isActiveRunStatus(report.run.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, report]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading && !report) {
    return <SkeletonRows />;
  }
  if (!report) {
    return <Notice tone="danger" message="Quality report could not be loaded." />;
  }

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{report.project.name}</h2>
            <p>
              <StatusBadge status={report.run.status} /> <span className="muted">Quality run {report.run.id}</span>
            </p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/projects/${report.project.id}`}>
              Project
            </a>
            {report.run.discovery_run_id && (
              <a className="button secondary-link" href={`#/discovery-runs/${report.run.discovery_run_id}`}>
                Discovery
              </a>
            )}
            <a className="button secondary-link" href={`${API_BASE_URL}/api/v1/quality-check-runs/${report.run.id}/report`} target="_blank" rel="noreferrer">
              JSON Report
            </a>
            <a className="button" href={qualityCheckHTMLReportURL(report.run.id)} target="_blank" rel="noreferrer">
              HTML Report
            </a>
          </div>
        </div>
        {report.run.error_message && <Notice tone="danger" message={report.run.error_message} />}
        <div className="summary-grid">
          <Metric label="Pages" value={report.summary.total_pages} />
          <Metric label="Findings" value={report.summary.total_findings} tone="critical" />
          <Metric label="Critical" value={report.summary.critical} tone="critical" />
          <Metric label="High" value={report.summary.high} tone="high" />
          <Metric label="Medium" value={report.summary.medium} tone="medium" />
          <Metric label="Low" value={report.summary.low} tone="low" />
        </div>
        <div className="detail-grid compact">
          <Field label="Target URL" value={report.run.target_url} />
          <Field label="Max Pages" value={String(report.run.max_pages)} />
          <Field label="Security" value={report.run.include_security ? "Enabled" : "Disabled"} />
          <Field label="Accessibility" value={report.run.include_accessibility ? "Enabled" : "Disabled"} />
          <Field label="Performance" value={report.run.include_performance ? "Enabled" : "Disabled"} />
          <Field label="Generated" value={formatDate(report.generated_at)} />
        </div>
      </section>

      <section>
        <h2>Safety Scope</h2>
        <Notice
          tone="info"
          message="Quality checks are passive metadata checks only. They do not submit forms, fuzz inputs, run payloads, or perform active security scanning."
        />
        <div className="analysis-grid">
          <AnalysisList title="Safety Notes" items={report.safety_notes} />
          <AnalysisList title="Limitations" items={report.limitations} />
        </div>
      </section>

      <section>
        <h2>Quality Findings</h2>
        <QualityResultTable results={report.results} />
      </section>
    </div>
  );
}

function QualityResultTable({ results }: { results: QualityCheckResult[] }) {
  if (results.length === 0) {
    return <EmptyState title="No quality findings" body="The quality run did not record passive quality findings." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Severity</th>
            <th>Category</th>
            <th>Rule</th>
            <th>Finding</th>
            <th>URL</th>
            <th>Evidence</th>
          </tr>
        </thead>
        <tbody>
          {results.map((result) => (
            <tr key={result.id}>
              <td>
                <span className={`severity ${result.severity}`}>{result.severity}</span>
              </td>
              <td>{result.category}</td>
              <td>
                <code>{result.rule_id}</code>
              </td>
              <td>
                <strong>{result.title}</strong>
                <p className="muted">{result.recommendation}</p>
              </td>
              <td>
                <code>{result.url}</code>
              </td>
              <td>
                <details>
                  <summary>Metadata</summary>
                  <pre>{JSON.stringify(result.evidence, null, 2)}</pre>
                </details>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function QARunPage({ runID }: { runID: string }) {
  const [report, setReport] = useState<QARunReport | undefined>();
  const [loading, setLoading] = useState(true);
  const [executing, setExecuting] = useState(false);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setReport(await getQARunReport(runID));
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [runID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!report || !isActiveRunStatus(report.run.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, report]);

  async function executePreviewedRun() {
    setExecuting(true);
    setError("");
    try {
      await executeQARun(runID);
      await refresh();
    } catch (executeError) {
      setError(executeError instanceof Error ? executeError.message : String(executeError));
    } finally {
      setExecuting(false);
    }
  }

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading && !report) {
    return <SkeletonRows />;
  }
  if (!report) {
    return <Notice tone="danger" message="Safe QA report could not be loaded." />;
  }

  const summary = summarizeFindingsForUI(report.findings);
  const canExecutePreview = report.run.status === "completed" && Boolean(report.run.test_plan_id) && !report.run.test_plan_execution_id;

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{report.project.name}</h2>
            <p>
              <StatusBadge status={report.run.status} /> <span className="muted">Safe QA run {report.run.id}</span>
            </p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/projects/${report.project.id}`}>
              Project
            </a>
            {report.run.discovery_run_id && (
              <a className="button secondary-link" href={`#/discovery-runs/${report.run.discovery_run_id}`}>
                Discovery
              </a>
            )}
            {report.run.quality_check_run_id && (
              <a className="button secondary-link" href={`#/quality-check-runs/${report.run.quality_check_run_id}`}>
                Quality
              </a>
            )}
            {report.run.test_plan_id && (
              <a className="button secondary-link" href={`#/test-plans/${report.run.test_plan_id}`}>
                Plan
              </a>
            )}
            {report.run.test_plan_execution_id && (
              <a className="button secondary-link" href={`#/test-plan-executions/${report.run.test_plan_execution_id}`}>
                Execution
              </a>
            )}
            <a className="button secondary-link" href={`${API_BASE_URL}/api/v1/qa-runs/${report.run.id}/report`} target="_blank" rel="noreferrer">
              JSON Report
            </a>
            <a className="button" href={qaRunHTMLReportURL(report.run.id)} target="_blank" rel="noreferrer">
              HTML Report
            </a>
          </div>
        </div>
        {report.run.error_message && <Notice tone="danger" message={report.run.error_message} />}
        <div className="summary-grid">
          <Metric label="Findings" value={summary.total_findings} tone="critical" />
          <Metric label="Critical" value={summary.critical} tone="critical" />
          <Metric label="High" value={summary.high} tone="high" />
          <Metric label="Medium" value={summary.medium} tone="medium" />
          <Metric label="Low" value={summary.low} tone="low" />
          <Metric label="Evidence" value={report.evidence.length} tone="info" />
        </div>
        <div className="detail-grid compact">
          <Field label="Created" value={formatDate(report.run.created_at)} />
          <Field label="Started" value={report.run.started_at ? formatDate(report.run.started_at) : "Not started"} />
          <Field label="Completed" value={report.run.completed_at ? formatDate(report.run.completed_at) : "Not completed"} />
          <Field label="Generated" value={formatDate(report.generated_at)} />
          <Field label="Mode" value={report.run.mode} />
          <Field label="Execution" value={report.run.test_plan_execution_id ? shortID(report.run.test_plan_execution_id) : "Preview only"} />
        </div>
        {canExecutePreview && (
          <div className="form-actions">
            <button type="button" disabled={executing} onClick={() => void executePreviewedRun()}>
              {executing ? "Starting" : "Execute previewed safe steps"}
            </button>
          </div>
        )}
      </section>

      {report.discovery_summary && (
        <section>
          <h2>Discovery Summary</h2>
          <div className="summary-grid">
            <Metric label="Pages" value={report.discovery_summary.total_pages} />
            <Metric label="Links" value={report.discovery_summary.total_links} />
            <Metric label="Forms" value={report.discovery_summary.total_forms} />
            <Metric label="Skipped Links" value={report.discovery_summary.skipped_links} tone="medium" />
            <Metric label="Console Errors" value={report.discovery_summary.total_console_errors} tone="medium" />
            <Metric label="Failed Requests" value={report.discovery_summary.total_failed_requests} tone="medium" />
          </div>
        </section>
      )}

      {report.quality_summary && (
        <section>
          <h2>Quality Summary</h2>
          <div className="summary-grid">
            <Metric label="Pages" value={report.quality_summary.total_pages} />
            <Metric label="Findings" value={report.quality_summary.total_findings} tone="critical" />
            <Metric label="Security" value={report.quality_summary.security_findings} tone="medium" />
            <Metric label="Accessibility" value={report.quality_summary.accessibility_findings} tone="medium" />
            <Metric label="Performance" value={report.quality_summary.performance_findings} tone="medium" />
            <Metric label="High" value={report.quality_summary.high} tone="high" />
          </div>
          {report.quality_check_run && (
            <div className="detail-grid compact">
              <Field label="Quality Run" value={shortID(report.quality_check_run.id)} />
              <Field label="Target" value={report.quality_check_run.target_url} />
              <Field label="Status" value={report.quality_check_run.status} />
              <Field label="Max Pages" value={String(report.quality_check_run.max_pages)} />
            </div>
          )}
        </section>
      )}

      {report.quality_results && report.quality_results.length > 0 && (
        <section>
          <h2>Quality Findings</h2>
          <QualityResultTable results={report.quality_results} />
        </section>
      )}

      {report.test_plan && (
        <section>
          <h2>AI Test Plan</h2>
          <div className="detail-grid compact">
            <Field label="Title" value={report.test_plan.title || "Untitled"} />
            <Field label="Source" value={report.test_plan.source_type || "Not set"} />
            <Field label="Risk" value={report.test_plan.risk_level || "Not set"} />
            <Field label="Scenarios" value={String(report.test_plan.total_scenarios)} />
            <Field label="Executable Scenarios" value={String(report.test_plan.execution_coverage?.executable_scenarios ?? 0)} />
            <Field label="Skipped Steps" value={String(report.test_plan.execution_coverage?.skipped_steps ?? 0)} />
          </div>
        </section>
      )}

      {report.execution_preview && (
        <section>
          <h2>Safe Execution Preview</h2>
          <ExecutionPreview preview={report.execution_preview} />
        </section>
      )}

      {report.execution_report && (
        <section>
          <h2>Execution Result</h2>
          <div className="summary-grid">
            <Metric label="Scenarios" value={report.execution_report.execution.total_scenarios} />
            <Metric label="Passed" value={report.execution_report.execution.passed_scenarios} />
            <Metric label="Failed" value={report.execution_report.execution.failed_scenarios} tone="high" />
            <Metric label="Skipped" value={report.execution_report.execution.skipped_scenarios} tone="medium" />
            <Metric label="Steps" value={report.execution_report.execution.total_steps} />
            <Metric label="Failed Steps" value={report.execution_report.execution.failed_steps} tone="high" />
          </div>
        </section>
      )}

      <section>
        <h2>Safety Scope</h2>
        <Notice
          tone="info"
          message="Safe QA runs use sanitized discovery metadata for AI planning and only execute approved deterministic browser DSL steps after explicit user action."
        />
        <div className="analysis-grid">
          <AnalysisList title="Safety Notes" items={report.safety_notes} />
          <AnalysisList title="Limitations" items={report.limitations} />
        </div>
      </section>

      <section>
        <h2>Findings</h2>
        <FindingTable findings={report.findings} />
      </section>

      <section>
        <h2>Evidence</h2>
        <EvidenceTable evidence={report.evidence} />
      </section>
    </div>
  );
}

function SafeExplorerRunPage({ runID }: { runID: string }) {
  const [report, setReport] = useState<SafeExplorerReport | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextReport = await getSafeExplorerReport(runID);
      setReport(nextReport);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [runID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!report || !isActiveRunStatus(report.run.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, report]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading && !report) {
    return <SkeletonRows />;
  }
  if (!report) {
    return <Notice tone="danger" message="Safe Explorer report could not be loaded." />;
  }

  const executedActions = report.actions.filter((action) => action.decision === "execute");
  const skippedActions = report.actions.filter((action) => action.decision === "skip");

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{report.project.name}</h2>
            <p>
              <StatusBadge status={report.run.status} /> <span className="muted">Safe Explorer run {report.run.id}</span>
            </p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/projects/${report.project.id}`}>
              Project
            </a>
            <a className="button secondary-link" href={`${API_BASE_URL}/api/v1/safe-explorer-runs/${report.run.id}/report`} target="_blank" rel="noreferrer">
              JSON Report
            </a>
            <a className="button" href={safeExplorerHTMLReportURL(report.run.id)} target="_blank" rel="noreferrer">
              HTML Report
            </a>
          </div>
        </div>
        {report.run.error_message && <Notice tone="danger" message={report.run.error_message} />}
        <div className="summary-grid">
          <Metric label="Steps" value={report.summary.total_steps} />
          <Metric label="Pages" value={report.summary.total_pages_observed} />
          <Metric label="Detected" value={report.summary.total_actions_detected} />
          <Metric label="Executed" value={report.summary.total_actions_executed} />
          <Metric label="Skipped" value={report.summary.total_actions_skipped} tone="medium" />
          <Metric label="Findings" value={report.summary.total_findings} tone="critical" />
        </div>
        <div className="detail-grid compact">
          <Field label="Start URL" value={report.run.start_url} />
          <Field label="Max Steps" value={String(report.run.max_steps)} />
          <Field label="Max Depth" value={String(report.run.max_depth)} />
          <Field label="Same Origin Only" value={report.run.same_origin_only ? "Yes" : "No"} />
          <Field label="Allow GET Forms" value={report.run.allow_get_forms ? "Yes" : "No"} />
          <Field label="Generated" value={formatDate(report.generated_at)} />
        </div>
      </section>

      <section>
        <h2>Safety Scope</h2>
        <Notice
          tone="info"
          message="Safe Explorer is deterministic: it executes only safe classified navigation actions, skips unsafe or unsupported actions with reasons, and does not use AI browser control."
        />
      </section>

      <section>
        <h2>Timeline</h2>
        <SafeExplorerTimelineTable report={report} />
      </section>

      <section>
        <h2>Executed Actions</h2>
        <SafeExplorerActionsTable actions={executedActions} emptyTitle="No executed actions" emptyBody="Safe Explorer has not executed safe actions for this run yet." />
      </section>

      <section>
        <h2>Skipped Actions</h2>
        <SafeExplorerActionsTable actions={skippedActions} emptyTitle="No skipped actions" emptyBody="Safe Explorer has not recorded skipped actions for this run." />
      </section>

      <section>
        <h2>Findings</h2>
        <FindingTable findings={report.findings} />
      </section>

      <section>
        <h2>Evidence</h2>
        <EvidenceTable evidence={report.evidence} />
      </section>

      <section>
        <h2>Safety Notes</h2>
        <div className="grid two-column">
          <AnalysisList title="Safety Notes" items={report.safety_notes} />
          <AnalysisList title="Limitations" items={report.limitations} />
        </div>
      </section>
    </div>
  );
}

function SafeExplorerTimelineTable({ report }: { report: SafeExplorerReport }) {
  if (report.steps.length === 0) {
    return <EmptyState title="No timeline" body="The Safe Explorer worker has not recorded steps yet." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>#</th>
            <th>Decision</th>
            <th>Depth</th>
            <th>Status</th>
            <th>Title / Action</th>
            <th>URL</th>
            <th>Signals</th>
            <th>Screenshot</th>
          </tr>
        </thead>
        <tbody>
          {report.steps.map((step) => (
            <tr key={step.id}>
              <td>{step.step_index}</td>
              <td>{step.action_decision}</td>
              <td>{step.depth}</td>
              <td>
                {step.result_status} {step.http_status ?? ""}
              </td>
              <td>{step.action_label || step.page_title || "Untitled"}</td>
              <td>
                <code>{step.action_target_url || step.normalized_url}</code>
              </td>
              <td>
                {step.console_error_count} console, {step.failed_request_count} network
              </td>
              <td>{step.screenshot_evidence_id ? <a href={evidenceDownloadURL(step.screenshot_evidence_id)}>Open</a> : "None"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function SafeExplorerActionsTable({
  actions,
  emptyTitle,
  emptyBody
}: {
  actions: SafeExplorerAction[];
  emptyTitle: string;
  emptyBody: string;
}) {
  if (actions.length === 0) {
    return <EmptyState title={emptyTitle} body={emptyBody} />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Decision</th>
            <th>Safety</th>
            <th>Type</th>
            <th>Label</th>
            <th>Target</th>
            <th>Skip Reason</th>
          </tr>
        </thead>
        <tbody>
          {actions.map((action) => (
            <tr key={action.id}>
              <td>{action.decision}</td>
              <td>
                <StatusBadge status={action.safety} />
              </td>
              <td>{action.action_type}</td>
              <td>{action.label || action.text || ""}</td>
              <td>
                <code>{action.target_url || action.href || ""}</code>
              </td>
              <td>{action.skip_reason || "None"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DiscoveryRunPage({ runID }: { runID: string }) {
  const [report, setReport] = useState<DiscoveryReport | undefined>();
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [providerID, setProviderID] = useState("");
  const [actionBusy, setActionBusy] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [nextReport, nextProviders] = await Promise.all([getDiscoveryReport(runID), listAIProviders()]);
      setReport(nextReport);
      setProviders(nextProviders);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [runID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!report || !isActiveRunStatus(report.run.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, report]);

  useEffect(() => {
    if (providerID || providers.length === 0) {
      return;
    }
    setProviderID(providers.find((provider) => provider.is_default)?.id || providers[0].id);
  }, [providerID, providers]);

  async function generateTestsFromDiscovery() {
    if (!report) {
      return;
    }
    setActionBusy("generate");
    setError("");
    try {
      const plan = await generateAITestPlan(report.project.id, {
        provider_id: providerID || undefined,
        discovery_run_id: report.run.id,
        include_discovery_map: true,
        execution_mode: "safe_executable",
        max_pages_from_discovery: 20,
        focus_areas: ["smoke", "functional", "regression"],
        max_scenarios: 10
      });
      window.location.hash = `#/test-plans/${plan.id}`;
    } catch (actionError) {
      setError(actionError instanceof Error ? actionError.message : String(actionError));
    } finally {
      setActionBusy("");
    }
  }

  async function startQAPreviewFromDiscovery() {
    if (!report) {
      return;
    }
    setActionBusy("qa");
    setError("");
    try {
      const run = await startQARun(report.project.id, {
        provider_id: providerID || undefined,
        use_existing_discovery_run_id: report.run.id,
        execute: false,
        max_pages: 20,
        max_depth: 2,
        max_scenarios: 10,
        focus_areas: ["smoke", "functional", "regression"]
      });
      window.location.hash = `#/qa-runs/${run.id}`;
    } catch (actionError) {
      setError(actionError instanceof Error ? actionError.message : String(actionError));
    } finally {
      setActionBusy("");
    }
  }

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading && !report) {
    return <SkeletonRows />;
  }
  if (!report) {
    return <Notice tone="danger" message="Discovery report could not be loaded." />;
  }

  const skippedLinks = report.links.filter((link) => link.skipped);

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{report.project.name}</h2>
            <p>
              <StatusBadge status={report.run.status} /> <span className="muted">Discovery run {report.run.id}</span>
            </p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/projects/${report.project.id}`}>
              Project
            </a>
            <a className="button secondary-link" href={`${API_BASE_URL}/api/v1/discovery-runs/${report.run.id}/report`} target="_blank" rel="noreferrer">
              JSON Report
            </a>
            <a className="button" href={discoveryHTMLReportURL(report.run.id)} target="_blank" rel="noreferrer">
              HTML Report
            </a>
          </div>
        </div>
        {providers.length > 0 && report.run.status === "completed" && (
          <div className="form-grid two">
            <label>
              AI Provider
              <select value={providerID} onChange={(event) => setProviderID(event.target.value)}>
                {providers.map((provider) => (
                  <option key={provider.id} value={provider.id}>
                    {provider.name} ({provider.model})
                  </option>
                ))}
              </select>
            </label>
            <div className="button-row">
              <button type="button" className="secondary" disabled={actionBusy !== ""} onClick={() => void generateTestsFromDiscovery()}>
                {actionBusy === "generate" ? "Generating" : "Generate tests from discovery"}
              </button>
              <button type="button" disabled={actionBusy !== ""} onClick={() => void startQAPreviewFromDiscovery()}>
                {actionBusy === "qa" ? "Starting" : "Start Safe QA preview"}
              </button>
            </div>
          </div>
        )}
        {providers.length === 0 && report.run.status === "completed" && (
          <Notice tone="info" message="Configure an AI provider to generate discovery-aware plans or Safe QA previews from this map." />
        )}
        {report.run.error_message && <Notice tone="danger" message={report.run.error_message} />}
        <div className="summary-grid">
          <Metric label="Pages" value={report.summary.total_pages} />
          <Metric label="Links" value={report.summary.total_links} />
          <Metric label="Forms" value={report.summary.total_forms} />
          <Metric label="Findings" value={report.summary.total_findings} tone="critical" />
          <Metric label="Console Errors" value={report.summary.total_console_errors} tone="medium" />
          <Metric label="Failed Requests" value={report.summary.total_failed_requests} tone="medium" />
        </div>
        <div className="detail-grid compact">
          <Field label="Start URL" value={report.run.start_url} />
          <Field label="Max Pages" value={String(report.run.max_pages)} />
          <Field label="Max Depth" value={String(report.run.max_depth)} />
          <Field label="Same Origin Only" value={report.run.same_origin_only ? "Yes" : "No"} />
          <Field label="Created" value={formatDate(report.run.created_at)} />
          <Field label="Generated" value={formatDate(report.generated_at)} />
        </div>
      </section>

      <section>
        <h2>Safety Scope</h2>
        <Notice
          tone="info"
          message="Discovery is deterministic and safe by default: it follows safe links, skips unsafe/external links, and never submits forms or runs AI browser control."
        />
      </section>

      <section>
        <h2>Pages</h2>
        <DiscoveryPagesTable report={report} />
      </section>

      <section>
        <h2>Skipped Links</h2>
        <DiscoveryLinksTable links={skippedLinks} pages={report.pages} />
      </section>

      <section>
        <h2>Forms</h2>
        <DiscoveryFormsTable report={report} />
      </section>

      <section>
        <h2>Findings</h2>
        <FindingTable findings={report.findings} />
      </section>

      <section>
        <h2>Evidence</h2>
        <EvidenceTable evidence={report.evidence} />
      </section>
    </div>
  );
}

function DiscoveryPagesTable({ report }: { report: DiscoveryReport }) {
  if (report.pages.length === 0) {
    return <EmptyState title="No pages" body="The discovery worker has not recorded pages yet." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Depth</th>
            <th>Status</th>
            <th>Title</th>
            <th>URL</th>
            <th>Signals</th>
            <th>Screenshot</th>
          </tr>
        </thead>
        <tbody>
          {report.pages.map((page) => (
            <tr key={page.id}>
              <td>{page.depth}</td>
              <td>{page.http_status ?? "n/a"}</td>
              <td>{page.title || "Untitled"}</td>
              <td>
                <code>{page.normalized_url}</code>
              </td>
              <td>
                {page.console_error_count} console, {page.failed_request_count} network
              </td>
              <td>{page.screenshot_evidence_id ? <a href={evidenceDownloadURL(page.screenshot_evidence_id)}>Open</a> : "None"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DiscoveryLinksTable({ links, pages }: { links: DiscoveryReport["links"]; pages: DiscoveryReport["pages"] }) {
  if (links.length === 0) {
    return <EmptyState title="No skipped links" body="Discovery did not record skipped links for this run." />;
  }
  const pageByID = new Map(pages.map((page) => [page.id, page]));
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Reason</th>
            <th>Source</th>
            <th>Text</th>
            <th>Href</th>
          </tr>
        </thead>
        <tbody>
          {links.map((link) => (
            <tr key={link.id}>
              <td>{link.skip_reason}</td>
              <td>{pageByID.get(link.source_page_id)?.path || shortID(link.source_page_id)}</td>
              <td>{link.link_text || ""}</td>
              <td>
                <code>{link.normalized_url || link.href}</code>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DiscoveryFormsTable({ report }: { report: DiscoveryReport }) {
  if (report.forms.length === 0) {
    return <EmptyState title="No forms" body="Discovery did not find visible forms." />;
  }
  const pageByID = new Map(report.pages.map((page) => [page.id, page]));
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Page</th>
            <th>Method</th>
            <th>Action</th>
            <th>Fields</th>
            <th>Password Fields</th>
            <th>Classification</th>
          </tr>
        </thead>
        <tbody>
          {report.forms.map((form) => (
            <tr key={form.id}>
              <td>{pageByID.get(form.page_id)?.path || shortID(form.page_id)}</td>
              <td>{form.form_method || "get"}</td>
              <td>
                <code>{form.form_action || ""}</code>
              </td>
              <td>{form.field_count}</td>
              <td>{form.password_field_count}</td>
              <td>{form.classification || ""}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function AuthorizationCheckRunPage({ runID }: { runID: string }) {
  const [report, setReport] = useState<AuthorizationCheckReport | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const nextReport = await getAuthorizationCheckReport(runID);
      setReport(nextReport);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : String(loadError));
    } finally {
      setLoading(false);
    }
  }, [runID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!report || !isActiveRunStatus(report.run.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [refresh, report]);

  if (error) {
    return <Notice tone="danger" message={error} />;
  }
  if (loading && !report) {
    return <SkeletonRows />;
  }
  if (!report) {
    return <Notice tone="danger" message="Authorization report could not be loaded." />;
  }

  return (
    <div className="grid">
      <section>
        <div className="section-heading">
          <div>
            <h2>{report.project.name}</h2>
            <p>
              <StatusBadge status={report.run.status} /> <span className="muted">Authorization run {report.run.id}</span>
            </p>
          </div>
          <div className="button-row">
            <a className="button secondary-link" href={`#/projects/${report.project.id}`}>
              Project
            </a>
            <a className="button" href={authorizationCheckHTMLReportURL(report.run.id)} target="_blank" rel="noreferrer">
              HTML Report
            </a>
          </div>
        </div>
        {report.run.error_message && <Notice tone="danger" message={report.run.error_message} />}
        <div className="summary-grid">
          <Metric label="Checks" value={report.run.total_checks} />
          <Metric label="Passed" value={report.run.passed_checks} />
          <Metric label="Failed" value={report.run.failed_checks} tone="high" />
          <Metric label="Skipped" value={report.run.skipped_checks} tone="medium" />
          <Metric label="Findings" value={report.summary.total_findings} tone="critical" />
          <Metric label="Evidence" value={report.evidence.length} tone="info" />
        </div>
        <div className="detail-grid compact">
          <Field label="Created" value={formatDate(report.run.created_at)} />
          <Field label="Started" value={report.run.started_at ? formatDate(report.run.started_at) : "Not started"} />
          <Field label="Completed" value={report.run.completed_at ? formatDate(report.run.completed_at) : "Not completed"} />
          <Field label="Generated" value={formatDate(report.generated_at)} />
        </div>
      </section>

      <section>
        <h2>Safety Scope</h2>
        <Notice
          tone="info"
          message="Authorization checks are explicit, deterministic, read-only browser URL checks. Qualora does not crawl, fuzz, submit arbitrary forms, execute payloads, or expose credentials."
        />
      </section>

      <section>
        <h2>Check Results</h2>
        <AuthorizationResultTable report={report} />
      </section>

      <section>
        <h2>Findings</h2>
        <AuthorizationFindingsTable findings={report.findings} />
      </section>

      <section>
        <h2>Evidence</h2>
        <EvidenceTable evidence={report.evidence} />
      </section>

      <section>
        <h2>Metadata</h2>
        <pre>{JSON.stringify(report.metadata, null, 2)}</pre>
      </section>
    </div>
  );
}

function AuthorizationResultTable({ report }: { report: AuthorizationCheckReport }) {
  if (report.results.length === 0) {
    return <EmptyState title="No results" body="The authorization worker has not recorded results yet." />;
  }
  const checksByID = new Map(report.checks.map((check) => [check.id, check]));
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Check</th>
            <th>Actor Role</th>
            <th>Expected</th>
            <th>Actual</th>
            <th>Target</th>
            <th>HTTP</th>
            <th>Reason</th>
          </tr>
        </thead>
        <tbody>
          {report.results.map((result) => {
            const check = checksByID.get(result.check_id);
            return (
              <tr key={result.id}>
                <td>
                  <StatusBadge status={result.status} />
                </td>
                <td>{check?.name || shortID(result.check_id)}</td>
                <td>{result.actor_role_name || "Not set"}</td>
                <td>{result.expected_outcome}</td>
                <td>{result.actual_outcome}</td>
                <td>
                  <code>{result.target_url || check?.target_url || ""}</code>
                  {result.final_url && <p className="muted">Final: {result.final_url}</p>}
                </td>
                <td>{result.http_status ?? "n/a"}</td>
                <td>{result.skip_reason || result.error_message || ""}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function AuthorizationFindingsTable({ findings }: { findings: AuthorizationCheckReport["findings"] }) {
  if (findings.length === 0) {
    return <EmptyState title="No findings" body="This authorization run did not record findings." />;
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
          {findings.map((finding) => (
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
  const loginEvidence = report.evidence.filter((item) => item.type === "login_observations");
  const apiEvidence = report.evidence.filter((item) => item.type === "api_observations" || item.type === "openapi_summary");
  const relatedTestPlans = report.test_plans || [];
  const apiResults = report.api_results || [];

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
          <Field label="Run Type" value={formatRunType(run.run_type || report.run_type)} />
          <Field label="Created" value={formatDate(run.created_at)} />
          <Field label="Started" value={run.started_at ? formatDate(run.started_at) : "Not started"} />
          <Field label="Completed" value={run.completed_at ? formatDate(run.completed_at) : "Not completed"} />
          <Field label="Page Title" value={run.page_title || "Not captured"} />
        </div>
      </section>

      {report.login_summary && (
        <section>
          <div className="section-heading">
            <div>
              <h2>Login Summary</h2>
              <p>Deterministic selector-based login metadata. Credentials are not displayed or sent to AI.</p>
            </div>
          </div>
          <div className="detail-grid compact">
            <Field label="Login Status" value={report.login_summary.login_status || "Not captured"} />
            <Field label="Credential Profile" value={report.login_summary.credential_profile_name || "Not available"} />
            <Field label="Login URL" value={report.login_summary.login_url || "Not captured"} />
            <Field label="Final URL" value={report.login_summary.login_final_url || "Not captured"} />
            <Field label="Page Title" value={report.login_summary.page_title || "Not captured"} />
            <Field label="Duration" value={`${report.login_summary.login_duration_ms || 0}ms`} />
            <Field label="Authenticated Target" value={report.login_summary.authenticated_target_url || "Login check only"} />
            <Field label="Failure Reason" value={report.login_summary.failure_reason || "None"} />
          </div>
        </section>
      )}

      {apiResults.length > 0 && (
        <section>
          <div className="section-heading">
            <div>
              <h2>API Smoke Results</h2>
              <p>{report.api_spec ? `${report.api_spec.name} · ${report.api_spec.parsed_title || "OpenAPI spec"}` : "Safe API operation results."}</p>
            </div>
          </div>
          {report.api_summary && (
            <div className="summary-grid">
              <Metric label="Total" value={report.api_summary.total_operations} />
              <Metric label="Executed" value={report.api_summary.executed_operations} />
              <Metric label="Passed" value={report.api_summary.passed_operations} />
              <Metric label="Failed" value={report.api_summary.failed_operations} tone="high" />
              <Metric label="Errors" value={report.api_summary.errored_operations} tone="high" />
              <Metric label="Skipped" value={report.api_summary.skipped_operations} tone="medium" />
            </div>
          )}
          <Notice tone="info" message="Response bodies, request bodies, cookies, auth headers, and tokens are not stored or sent to AI." />
          <APIResultsTable results={apiResults} />
        </section>
      )}

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
        <h2>Login Metadata</h2>
        <MetadataBlocks evidence={loginEvidence} empty="No login metadata for this run." />
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

function FindingTable({ findings }: { findings: Report["findings"] }) {
  if (findings.length === 0) {
    return <EmptyState title="No findings" body="This report did not record findings." />;
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
          {findings.map((finding) => (
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

function APIResultsTable({ results }: { results: APICheckResult[] }) {
  if (results.length === 0) {
    return <EmptyState title="No API results" body="This run did not record API operation results." />;
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Status</th>
            <th>Method</th>
            <th>Path</th>
            <th>HTTP</th>
            <th>Duration</th>
            <th>Size</th>
            <th>Content Type</th>
            <th>Reason/Error</th>
          </tr>
        </thead>
        <tbody>
          {results.map((result) => (
            <tr key={result.id}>
              <td>
                <StatusBadge status={result.status} />
              </td>
              <td>
                <code>{result.method}</code>
              </td>
              <td>
                <code>{result.path}</code>
                {result.resolved_url && <p className="muted">{result.resolved_url}</p>}
              </td>
              <td>{result.http_status ?? "n/a"}</td>
              <td>{result.duration_ms === undefined ? "n/a" : `${result.duration_ms}ms`}</td>
              <td>{result.response_size_bytes === undefined ? "n/a" : `${result.response_size_bytes} bytes`}</td>
              <td>{result.response_content_type || "n/a"}</td>
              <td>{result.skipped_reason || result.error_message || ""}</td>
            </tr>
          ))}
        </tbody>
      </table>
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
  if (parts[0] === "setup-project") {
    return { name: "setup-project" };
  }
  if (parts[0] === "reports") {
    return { name: "reports" };
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
  if (parts[0] === "api-specs" && parts[1]) {
    return { name: "api-spec", id: parts[1] };
  }
  if (parts[0] === "test-plans" && parts[1]) {
    return { name: "test-plan", id: parts[1] };
  }
  if (parts[0] === "test-plans") {
    return { name: "test-plans" };
  }
  if (parts[0] === "test-plan-executions" && parts[1]) {
    return { name: "test-plan-execution", id: parts[1] };
  }
  if (parts[0] === "authorization-check-runs" && parts[1]) {
    return { name: "authorization-check-run", id: parts[1] };
  }
  if (parts[0] === "discovery-runs" && parts[1]) {
    return { name: "discovery-run", id: parts[1] };
  }
  if (parts[0] === "quality-check-runs" && parts[1]) {
    return { name: "quality-check-run", id: parts[1] };
  }
  if (parts[0] === "safe-explorer-runs" && parts[1]) {
    return { name: "safe-explorer-run", id: parts[1] };
  }
  if (parts[0] === "qa-runs" && parts[1]) {
    return { name: "qa-run", id: parts[1] };
  }
  return { name: "dashboard" };
}

function hashForRoute(route: Route): string {
  switch (route.name) {
    case "dashboard":
      return "/";
    case "new-project":
      return "/projects/new";
    case "setup-project":
      return "/setup-project";
    case "reports":
      return "/reports";
    case "project":
      return `/projects/${route.id}`;
    case "runs":
      return "/runs";
    case "ai-providers":
      return "/ai-providers";
    case "api-spec":
      return `/api-specs/${route.id}`;
    case "test-plans":
      return "/test-plans";
    case "test-plan":
      return `/test-plans/${route.id}`;
    case "test-plan-execution":
      return `/test-plan-executions/${route.id}`;
    case "authorization-check-run":
      return `/authorization-check-runs/${route.id}`;
    case "discovery-run":
      return `/discovery-runs/${route.id}`;
    case "quality-check-run":
      return `/quality-check-runs/${route.id}`;
    case "safe-explorer-run":
      return `/safe-explorer-runs/${route.id}`;
    case "qa-run":
      return `/qa-runs/${route.id}`;
    case "run":
      return `/runs/${route.id}`;
  }
}

function titleForRoute(route: Route): string {
  switch (route.name) {
    case "dashboard":
      return "Dashboard";
    case "new-project":
      return "Create Project";
    case "setup-project":
      return "Guided Project Setup";
    case "reports":
      return "Reports";
    case "project":
      return "Project Details";
    case "runs":
      return "Runs";
    case "ai-providers":
      return "AI Providers";
    case "api-spec":
      return "API Spec";
    case "test-plans":
      return "AI Test Plans";
    case "test-plan":
      return "Test Plan";
    case "test-plan-execution":
      return "Plan Execution";
    case "authorization-check-run":
      return "Authorization Report";
    case "discovery-run":
      return "Discovery Report";
    case "quality-check-run":
      return "Quality Report";
    case "safe-explorer-run":
      return "Safe Explorer Report";
    case "qa-run":
      return "Safe QA Report";
    case "run":
      return "Run Report";
  }
}

function buildProjectSetupPayload(
  projectForm: { name: string; frontend_url: string; api_base_url: string; allowed_hosts: string; allow_private_targets: boolean },
  aiMode: "skip" | "existing" | "create" | "demo",
  selectedProviderID: string,
  providerForm: ProviderFormState,
  credentialMode: "skip" | "create",
  credentialForm: CredentialProfileInput,
  apiSpecMode: "skip" | "url" | "inline" | "demo",
  apiSpecForm: { name: string; source_url: string; raw_spec: string },
  workflow: WizardWorkflowState
): ProjectSetupInput {
  const payload: ProjectSetupInput = {
    project: {
      name: projectForm.name.trim(),
      frontend_url: projectForm.frontend_url.trim(),
      api_base_url: projectForm.api_base_url.trim(),
      openapi_url: "",
      allowed_hosts: splitHosts(projectForm.allowed_hosts),
      security_mode: "passive",
      destructive_actions: false,
      allow_private_targets: projectForm.allow_private_targets
    },
    ai: { mode: aiMode },
    credential: { mode: credentialMode },
    api_spec: { mode: apiSpecMode === "skip" ? "skip" : apiSpecMode === "demo" ? "demo" : "import" },
    workflow
  };
  if (aiMode === "existing") {
    payload.ai = { mode: "existing", provider_id: selectedProviderID };
  } else if (aiMode === "create") {
    payload.ai = { mode: "create", provider: inputForProviderForm(providerForm, false) };
  }
  if (credentialMode === "create") {
    payload.credential = { mode: "create", profile: credentialForm };
  }
  if (apiSpecMode === "url") {
    payload.api_spec = {
      mode: "import",
      spec: { name: apiSpecForm.name.trim() || "OpenAPI Spec", source_type: "url", source_url: apiSpecForm.source_url.trim(), raw_spec: "" }
    };
  } else if (apiSpecMode === "inline") {
    payload.api_spec = {
      mode: "import",
      spec: { name: apiSpecForm.name.trim() || "OpenAPI Spec", source_type: "inline", source_url: "", raw_spec: apiSpecForm.raw_spec }
    };
  }
  return payload;
}

function demoProjectSetupPayload(): ProjectSetupInput {
  return {
    project: {
      name: "Qualora Demo Workflow",
      frontend_url: "http://demo-web:8080",
      api_base_url: "http://demo-api:8080",
      openapi_url: "",
      allowed_hosts: ["demo-web", "demo-api"],
      security_mode: "passive",
      destructive_actions: false,
      allow_private_targets: true
    },
    ai: { mode: "demo" },
    credential: {
      mode: "create",
      profile: {
        ...wizardCredentialDefaults(),
        name: "Qualora Demo Login",
        username: "demo@example.com",
        password: "qualora-demo-password",
        login_url: "http://demo-web:8080/login",
        success_url_contains: "/dashboard",
        success_text_contains: "Welcome to the Qualora demo dashboard",
        failure_text_contains: "Invalid credentials"
      }
    },
    api_spec: { mode: "demo" },
    workflow: {
      browser_smoke: true,
      discovery: true,
      quality_checks: true,
      safe_qa_run: true,
      execute_safe_qa: false,
      api_smoke: true,
      authenticated_smoke: true
    }
  };
}

function wizardCredentialDefaults(): CredentialProfileInput {
  return {
    name: "Test Login",
    type: "username_password",
    role_name: "",
    role_description: "",
    subject_label: "",
    username: "",
    password: "",
    login_url: "",
    username_selector: "#email",
    password_selector: "#password",
    submit_selector: "button[type=submit]",
    success_url_contains: "",
    success_text_contains: "",
    failure_text_contains: "",
    post_login_wait_ms: 500,
    is_default: true
  };
}

function humanizeIdentifier(value: string): string {
  return value
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
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

function summarizeFindingsForUI(findings: { severity: string }[]): Report["summary"] {
  const summary: Report["summary"] = {
    total_findings: findings.length,
    critical: 0,
    high: 0,
    medium: 0,
    low: 0,
    info: 0
  };
  for (const finding of findings) {
    if (finding.severity === "critical") {
      summary.critical += 1;
    } else if (finding.severity === "high") {
      summary.high += 1;
    } else if (finding.severity === "medium") {
      summary.medium += 1;
    } else if (finding.severity === "low") {
      summary.low += 1;
    } else if (finding.severity === "info") {
      summary.info += 1;
    }
  }
  return summary;
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
  return (
    status === "queued" ||
    status === "pending" ||
    status === "running" ||
    status === "running_discovery" ||
    status === "running_quality_checks" ||
    status === "generating_plan" ||
    status === "previewing_execution" ||
    status === "executing_plan"
  );
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

function formatRunType(value: string): string {
  if (value === "api_smoke") {
    return "API smoke";
  }
  if (value === "browser_smoke") {
    return "Browser smoke";
  }
  if (value === "login_check") {
    return "Login check";
  }
  if (value === "authenticated_browser_smoke") {
    return "Authenticated browser smoke";
  }
  if (value === "full") {
    return "Full";
  }
  return value || "Full";
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(new Date(value));
}

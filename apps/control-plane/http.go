package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type App struct {
	store         *Store
	queue         *Queue
	evidenceStore *EvidenceStore
	secretBox     *SecretBox
	aiClient      *OpenAICompatibleClient
	logger        *slog.Logger
	corsOrigins   []string
	auth          AuthConfig
}

func NewApp(store *Store, queue *Queue, evidenceStore *EvidenceStore, secretBox *SecretBox, aiClient *OpenAICompatibleClient, logger *slog.Logger, corsOrigins []string, auth AuthConfig) *App {
	return &App{store: store, queue: queue, evidenceStore: evidenceStore, secretBox: secretBox, aiClient: aiClient, logger: logger, corsOrigins: corsOrigins, auth: auth}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/api/v1/setup/status", a.handleSetupStatus)
	mux.HandleFunc("/api/v1/setup/admin", a.handleSetupAdmin)
	mux.HandleFunc("/api/v1/auth/login", a.handleLogin)
	mux.HandleFunc("/api/v1/auth/logout", a.handleLogout)
	mux.HandleFunc("/api/v1/auth/me", a.handleMe)
	mux.HandleFunc("/api/v1/projects", a.handleProjects)
	mux.HandleFunc("/api/v1/projects/", a.handleProjectSubroutes)
	mux.HandleFunc("/api/v1/runs", a.handleRuns)
	mux.HandleFunc("/api/v1/runs/", a.handleRunSubroutes)
	mux.HandleFunc("/api/v1/evidence/", a.handleEvidenceSubroutes)
	mux.HandleFunc("/api/v1/ai/providers", a.handleAIProviders)
	mux.HandleFunc("/api/v1/ai/providers/", a.handleAIProviderSubroutes)
	mux.HandleFunc("/api/v1/test-plans/", a.handleTestPlanSubroutes)
	mux.HandleFunc("/api/v1/test-plan-executions/", a.handleTestPlanExecutionSubroutes)
	mux.HandleFunc("/api/v1/api-specs/", a.handleAPISpecSubroutes)
	mux.HandleFunc("/api/v1/credential-profiles/", a.handleCredentialProfileSubroutes)
	mux.HandleFunc("/api/v1/authorization-checks/", a.handleAuthorizationCheckSubroutes)
	mux.HandleFunc("/api/v1/authorization-check-runs/", a.handleAuthorizationCheckRunSubroutes)
	mux.HandleFunc("/api/v1/discovery-runs/", a.handleDiscoveryRunSubroutes)
	mux.HandleFunc("/api/v1/quality-check-runs/", a.handleQualityCheckRunSubroutes)
	mux.HandleFunc("/api/v1/qa-runs/", a.handleQARunSubroutes)
	return withCORS(a.corsOrigins, withJSONContentType(withRequestLog(a.logger, a.withAuth(mux))))
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := a.store.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database_unavailable", "database is unavailable")
		return
	}
	if err := a.queue.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "redis_unavailable", "redis is unavailable")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (a *App) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.createProject(w, r)
	case http.MethodGet:
		a.listProjects(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (a *App) createProject(w http.ResponseWriter, r *http.Request) {
	var input CreateProjectRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid project JSON")
		return
	}

	normalized, err := NormalizeProjectRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_project", err.Error())
		return
	}

	project, err := a.store.CreateProject(r.Context(), normalized)
	if err != nil {
		a.logger.Error("create project failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_project_failed", "project could not be created")
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (a *App) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := a.store.ListProjects(r.Context())
	if err != nil {
		a.logger.Error("list projects failed", "error", err)
		writeError(w, http.StatusInternalServerError, "list_projects_failed", "projects could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (a *App) handleProjectSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" && r.Method == http.MethodGet {
		a.getProject(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "runs" {
		switch r.Method {
		case http.MethodPost:
			a.createRun(w, r, parts[0])
		case http.MethodGet:
			a.listRuns(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "browser-smoke-runs" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.createBrowserSmokeRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "authenticated-browser-smoke-runs" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.createAuthenticatedBrowserSmokeRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "ai-test-plans" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.createAITestPlan(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "test-plans" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.listTestPlans(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "api-specs" {
		switch r.Method {
		case http.MethodPost:
			a.createAPISpec(w, r, parts[0])
		case http.MethodGet:
			a.listAPISpecs(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "credential-profiles" {
		switch r.Method {
		case http.MethodPost:
			a.createCredentialProfile(w, r, parts[0])
		case http.MethodGet:
			a.listCredentialProfiles(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "authorization-checks" {
		switch r.Method {
		case http.MethodPost:
			a.createAuthorizationCheck(w, r, parts[0])
		case http.MethodGet:
			a.listAuthorizationChecks(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "authorization-check-runs" {
		switch r.Method {
		case http.MethodPost:
			a.createAuthorizationCheckRun(w, r, parts[0])
		case http.MethodGet:
			a.listAuthorizationCheckRuns(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "discovery-runs" {
		switch r.Method {
		case http.MethodPost:
			a.createDiscoveryRun(w, r, parts[0])
		case http.MethodGet:
			a.listDiscoveryRuns(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "quality-check-runs" {
		switch r.Method {
		case http.MethodPost:
			a.createQualityCheckRun(w, r, parts[0])
		case http.MethodGet:
			a.listQualityCheckRuns(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "qa-runs" {
		switch r.Method {
		case http.MethodPost:
			a.createQARun(w, r, parts[0])
		case http.MethodGet:
			a.listQARuns(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getProject(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (a *App) createRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}

	run, jobs, err := a.store.CreateRun(r.Context(), *project)
	if err != nil {
		a.logger.Error("create run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_run_failed", "run could not be created")
		return
	}

	if err := a.enqueueRunJobs(r.Context(), *project, run, jobs); err != nil {
		a.logger.Error("enqueue run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkRunFailed(r.Context(), run.ID, "run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (a *App) createBrowserSmokeRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for browser smoke run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.FrontendURL == "" {
		writeError(w, http.StatusBadRequest, "frontend_url_required", "project must have frontend_url to start a browser smoke run")
		return
	}

	run, jobs, err := a.store.CreateRunForKindsWithType(r.Context(), *project, []string{JobKindBrowser}, RunTypeBrowserSmoke, "")
	if err != nil {
		a.logger.Error("create browser smoke run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_run_failed", "browser smoke run could not be created")
		return
	}
	if err := a.enqueueRunJobs(r.Context(), *project, run, jobs); err != nil {
		a.logger.Error("enqueue browser smoke run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkRunFailed(r.Context(), run.ID, "run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "browser smoke run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (a *App) enqueueRunJobs(ctx context.Context, project Project, run *TestRun, jobs []RunJob) error {
	for _, job := range jobs {
		switch job.Kind {
		case JobKindBrowser:
			if err := a.queue.EnqueueBrowserRun(ctx, BrowserRunJob{JobID: job.ID, RunID: run.ID, ProjectID: project.ID}); err != nil {
				return err
			}
		case JobKindAPI:
			if err := a.queue.EnqueueAPIRun(ctx, APIRunJob{JobID: job.ID, RunID: run.ID, ProjectID: project.ID}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) handleRuns(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/runs" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	a.listRuns(w, r, "")
}

func (a *App) listRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	runs, err := a.store.ListRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_runs_failed", "runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func (a *App) handleRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/runs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" && r.Method == http.MethodGet {
		a.getRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" && r.Method == http.MethodGet {
		a.getReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" && r.Method == http.MethodGet {
		a.getHTMLReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "ai-analysis" {
		switch r.Method {
		case http.MethodGet:
			a.getAIAnalysis(w, r, parts[0])
		case http.MethodPost:
			a.createAIAnalysis(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "api-results" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAPIResults(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := a.store.GetRun(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "run_not_found", "run was not found")
			return
		}
		a.logger.Error("get run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_run_failed", "run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) getReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "run_not_found", "run was not found")
			return
		}
		a.logger.Error("get report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_report_failed", "report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getHTMLReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "run_not_found", "run was not found")
			return
		}
		a.logger.Error("get html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_report_failed", "report could not be loaded")
		return
	}

	run, err := a.store.GetRun(r.Context(), runID)
	if err != nil {
		a.logger.Error("get run for html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_run_failed", "run could not be loaded")
		return
	}

	project, err := a.store.GetProject(r.Context(), report.ProjectID)
	if err != nil {
		a.logger.Error("get project for html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderHTMLReport(w, project, run, report, time.Now().UTC()); err != nil {
		a.logger.Error("render html report failed", "error", err)
	}
}

func (a *App) handleEvidenceSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/evidence/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 1 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	a.getEvidenceObject(w, r, parts[0])
}

func (a *App) handleTestPlanSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/test-plans/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodGet:
			a.getTestPlan(w, r, parts[0])
		case http.MethodDelete:
			a.deleteTestPlan(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "export.json" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.exportTestPlanJSON(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "executions" {
		switch r.Method {
		case http.MethodPost:
			a.createTestPlanExecution(w, r, parts[0])
		case http.MethodGet:
			a.listTestPlanExecutions(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) handleTestPlanExecutionSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/test-plan-executions/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getTestPlanExecution(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getTestPlanExecutionReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getTestPlanExecutionHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) createTestPlanExecution(w http.ResponseWriter, r *http.Request, planID string) {
	input, err := decodeTestPlanExecutionRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_test_plan_execution", err.Error())
		return
	}

	plan, err := a.store.GetTestPlan(r.Context(), planID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_not_found", "test plan was not found")
			return
		}
		a.logger.Error("get test plan for execution failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_test_plan_failed", "test plan could not be loaded")
		return
	}
	if plan.Status != StatusCompleted {
		writeError(w, http.StatusBadRequest, "test_plan_not_executable", "only completed test plans can be executed")
		return
	}

	project, err := a.store.GetProject(r.Context(), plan.ProjectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for test plan execution failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}

	preview, err := BuildTestPlanExecutionPreview(*plan, *project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "test_plan_not_executable", err.Error())
		return
	}
	if preview.DryRun {
		writeJSON(w, http.StatusOK, preview)
		return
	}

	detail, err := a.store.CreateTestPlanExecution(r.Context(), *plan, *preview)
	if err != nil {
		a.logger.Error("create test plan execution failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_test_plan_execution_failed", "test plan execution could not be created")
		return
	}
	if preview.ExecutableSteps > 0 {
		if err := a.queue.EnqueueTestPlanExecution(r.Context(), TestPlanExecutionJob{ExecutionID: detail.Execution.ID}); err != nil {
			a.logger.Error("enqueue test plan execution failed", "execution_id", detail.Execution.ID, "error", err)
			_ = a.store.MarkTestPlanExecutionFailed(r.Context(), detail.Execution.ID, "test plan execution could not be queued")
			writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "test plan execution could not be queued")
			return
		}
	}
	writeJSON(w, http.StatusCreated, detail)
}

func (a *App) listTestPlanExecutions(w http.ResponseWriter, r *http.Request, planID string) {
	if _, err := a.store.GetTestPlan(r.Context(), planID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_not_found", "test plan was not found")
			return
		}
		a.logger.Error("get test plan for executions failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_test_plan_failed", "test plan could not be loaded")
		return
	}
	executions, err := a.store.ListTestPlanExecutions(r.Context(), planID)
	if err != nil {
		a.logger.Error("list test plan executions failed", "test_plan_id", planID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_test_plan_executions_failed", "test plan executions could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"executions": executions})
}

func (a *App) getTestPlanExecution(w http.ResponseWriter, r *http.Request, executionID string) {
	detail, err := a.store.GetTestPlanExecution(r.Context(), executionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_execution_not_found", "test plan execution was not found")
			return
		}
		a.logger.Error("get test plan execution failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_test_plan_execution_failed", "test plan execution could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (a *App) getTestPlanExecutionReport(w http.ResponseWriter, r *http.Request, executionID string) {
	report, err := a.store.GetTestPlanExecutionReport(r.Context(), executionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_execution_not_found", "test plan execution was not found")
			return
		}
		a.logger.Error("get test plan execution report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_test_plan_execution_report_failed", "test plan execution report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getTestPlanExecutionHTMLReport(w http.ResponseWriter, r *http.Request, executionID string) {
	report, err := a.store.GetTestPlanExecutionReport(r.Context(), executionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_execution_not_found", "test plan execution was not found")
			return
		}
		a.logger.Error("get test plan execution html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_test_plan_execution_report_failed", "test plan execution report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderTestPlanExecutionHTMLReport(w, report); err != nil {
		a.logger.Error("render test plan execution html report failed", "error", err)
	}
}

func (a *App) createAITestPlan(w http.ResponseWriter, r *http.Request, projectID string) {
	input, err := decodeAITestPlanRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_test_plan", err.Error())
		return
	}

	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for AI test plan failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}

	plan, err := a.generateAITestPlan(r.Context(), *project, input)
	if err != nil {
		if plan != nil {
			writeJSON(w, http.StatusCreated, plan)
			return
		}
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusBadRequest, "ai_provider_or_input_not_found", "AI provider, run, or discovery run could not be found")
			return
		}
		a.logger.Error("generate AI test plan failed", "error", err)
		writeError(w, http.StatusBadRequest, "generate_ai_test_plan_failed", RedactSecrets(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, plan)
}

func (a *App) listTestPlans(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for test plan list failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	plans, err := a.store.ListTestPlans(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list test plans failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_test_plans_failed", "test plans could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"test_plans": plans})
}

func (a *App) getTestPlan(w http.ResponseWriter, r *http.Request, planID string) {
	plan, err := a.store.GetTestPlan(r.Context(), planID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_not_found", "test plan was not found")
			return
		}
		a.logger.Error("get test plan failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_test_plan_failed", "test plan could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (a *App) deleteTestPlan(w http.ResponseWriter, r *http.Request, planID string) {
	if err := a.store.DeleteTestPlan(r.Context(), planID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_not_found", "test plan was not found")
			return
		}
		a.logger.Error("delete test plan failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_test_plan_failed", "test plan could not be deleted")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *App) exportTestPlanJSON(w http.ResponseWriter, r *http.Request, planID string) {
	plan, err := a.store.GetTestPlan(r.Context(), planID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "test_plan_not_found", "test plan was not found")
			return
		}
		a.logger.Error("export test plan failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_test_plan_failed", "test plan could not be loaded")
		return
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": "qualora-test-plan-" + plan.ID + ".json"}))
	writeJSON(w, http.StatusOK, plan.PlanJSON)
}

func (a *App) getEvidenceObject(w http.ResponseWriter, r *http.Request, evidenceID string) {
	record, err := a.store.GetEvidence(r.Context(), evidenceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "evidence_not_found", "evidence was not found")
			return
		}
		a.logger.Error("get evidence failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_evidence_failed", "evidence could not be loaded")
		return
	}

	object, err := a.evidenceStore.Open(r.Context(), *record)
	if err != nil {
		a.logger.Error("open evidence object failed", "evidence_id", evidenceID, "error", err)
		writeError(w, http.StatusNotFound, "evidence_object_unavailable", "evidence object could not be loaded")
		return
	}
	defer object.Body.Close()

	w.Header().Set("Content-Type", object.ContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": object.Filename}))
	if object.ContentLength >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(object.ContentLength, 10))
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, object.Body); err != nil {
		a.logger.Error("stream evidence object failed", "evidence_id", evidenceID, "error", err)
	}
}

func (a *App) handleAIProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		providers, err := a.store.ListAIProviders(r.Context())
		if err != nil {
			a.logger.Error("list AI providers failed", "error", err)
			writeError(w, http.StatusInternalServerError, "list_ai_providers_failed", "AI providers could not be listed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"providers": providers})
	case http.MethodPost:
		a.createAIProvider(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (a *App) handleAIProviderSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/providers/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodGet:
			a.getAIProvider(w, r, parts[0])
		case http.MethodPut:
			a.updateAIProvider(w, r, parts[0])
		case http.MethodDelete:
			a.deleteAIProvider(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "test" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.testAIProvider(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) createAIProvider(w http.ResponseWriter, r *http.Request) {
	input, err := decodeAIProviderRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_provider", err.Error())
		return
	}
	provider, err := a.providerFromInput(input, "", "")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_provider", err.Error())
		return
	}
	created, err := a.store.CreateAIProvider(r.Context(), provider)
	if err != nil {
		a.logger.Error("create AI provider failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_ai_provider_failed", "AI provider could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *App) getAIProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := a.store.GetAIProvider(r.Context(), providerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_provider_not_found", "AI provider was not found")
			return
		}
		a.logger.Error("get AI provider failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_provider_failed", "AI provider could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, provider)
}

func (a *App) updateAIProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	current, err := a.store.GetAIProvider(r.Context(), providerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_provider_not_found", "AI provider was not found")
			return
		}
		a.logger.Error("get AI provider for update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_provider_failed", "AI provider could not be loaded")
		return
	}
	input, err := decodeAIProviderRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_provider", err.Error())
		return
	}
	provider, err := a.providerFromInput(input, current.APIKeyEncrypted, current.ExtraHeadersEncrypted)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_provider", err.Error())
		return
	}
	updated, err := a.store.UpdateAIProvider(r.Context(), providerID, provider)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_provider_not_found", "AI provider was not found")
			return
		}
		a.logger.Error("update AI provider failed", "error", err)
		writeError(w, http.StatusInternalServerError, "update_ai_provider_failed", "AI provider could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *App) deleteAIProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	if err := a.store.DeleteAIProvider(r.Context(), providerID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_provider_not_found", "AI provider was not found")
			return
		}
		a.logger.Error("delete AI provider failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_ai_provider_failed", "AI provider could not be deleted")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *App) testAIProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := a.store.GetAIProvider(r.Context(), providerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_provider_not_found", "AI provider was not found")
			return
		}
		a.logger.Error("get AI provider for test failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_provider_failed", "AI provider could not be loaded")
		return
	}
	clientRequest, err := a.clientRequestForProvider(*provider, []AIChatMessage{
		{Role: "system", Content: "Return strict JSON only."},
		{Role: "user", Content: `Return {"ok":true,"message":"Qualora provider test"} as JSON.`},
	})
	if err != nil {
		writeJSON(w, http.StatusOK, AIProviderTestResult{Success: false, ProviderName: provider.Name, Model: provider.Model, ErrorMessage: RedactSecrets(err.Error())})
		return
	}
	started := time.Now()
	_, err = a.aiClient.ChatCompletion(r.Context(), clientRequest)
	result := AIProviderTestResult{
		Success:      err == nil,
		ProviderName: provider.Name,
		Model:        provider.Model,
		LatencyMS:    time.Since(started).Milliseconds(),
	}
	if err != nil {
		result.ErrorMessage = RedactSecrets(err.Error())
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) getAIAnalysis(w http.ResponseWriter, r *http.Request, runID string) {
	analysis, err := a.store.GetLatestAIAnalysis(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"ai_analysis": nil})
			return
		}
		a.logger.Error("get AI analysis failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_analysis_failed", "AI analysis could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ai_analysis": analysis})
}

func (a *App) createAIAnalysis(w http.ResponseWriter, r *http.Request, runID string) {
	var input AIAnalysisRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if r.Body != nil {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid AI analysis JSON")
			return
		}
	}

	report, err := a.store.GetReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "run_not_found", "run was not found")
			return
		}
		a.logger.Error("get report for AI analysis failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_report_failed", "report could not be loaded")
		return
	}

	provider, err := a.providerForAnalysis(r.Context(), input.ProviderID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusBadRequest, "ai_provider_required", "configure an AI provider before running AI analysis")
			return
		}
		a.logger.Error("load AI provider for analysis failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_provider_failed", "AI provider could not be loaded")
		return
	}

	analysis, err := a.store.CreateAIAnalysis(r.Context(), runID, provider.ID, provider.Model)
	if err != nil {
		a.logger.Error("create AI analysis failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_ai_analysis_failed", "AI analysis could not be created")
		return
	}

	userPrompt, err := BuildAIUserPrompt(report)
	if err != nil {
		analysis, _ = a.store.FailAIAnalysis(r.Context(), analysis.ID, RedactSecrets(err.Error()))
		writeJSON(w, http.StatusCreated, analysis)
		return
	}
	clientRequest, err := a.clientRequestForProvider(*provider, []AIChatMessage{
		{Role: "system", Content: AIAnalysisSystemPrompt()},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		analysis, _ = a.store.FailAIAnalysis(r.Context(), analysis.ID, RedactSecrets(err.Error()))
		writeJSON(w, http.StatusCreated, analysis)
		return
	}

	clientResponse, err := a.aiClient.ChatCompletion(r.Context(), clientRequest)
	if err != nil {
		analysis, _ = a.store.FailAIAnalysis(r.Context(), analysis.ID, RedactSecrets(err.Error()))
		writeJSON(w, http.StatusCreated, analysis)
		return
	}
	payload, analysisJSON, err := ParseAIAnalysisPayload(clientResponse.Content)
	if err != nil {
		analysis, _ = a.store.FailAIAnalysis(r.Context(), analysis.ID, RedactSecrets(err.Error()))
		writeJSON(w, http.StatusCreated, analysis)
		return
	}
	completed, err := a.store.CompleteAIAnalysis(r.Context(), analysis.ID, payload, analysisJSON, *clientResponse)
	if err != nil {
		a.logger.Error("complete AI analysis failed", "error", err)
		writeError(w, http.StatusInternalServerError, "complete_ai_analysis_failed", "AI analysis could not be saved")
		return
	}
	writeJSON(w, http.StatusCreated, completed)
}

func decodeAIProviderRequest(w http.ResponseWriter, r *http.Request) (AIProviderRequest, error) {
	var input AIProviderRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, fmt.Errorf("request body must be valid AI provider JSON")
	}
	return normalizeProviderRequest(input)
}

func decodeAITestPlanRequest(w http.ResponseWriter, r *http.Request) (AITestPlanRequest, error) {
	var input AITestPlanRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil && !errors.Is(err, io.EOF) {
		return input, fmt.Errorf("request body must be valid AI test plan JSON")
	}
	return NormalizeAITestPlanRequest(input)
}

func decodeTestPlanExecutionRequest(w http.ResponseWriter, r *http.Request) (TestPlanExecutionRequest, error) {
	var input TestPlanExecutionRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil && !errors.Is(err, io.EOF) {
		return input, fmt.Errorf("request body must be valid test plan execution JSON")
	}
	return NormalizeTestPlanExecutionRequest(input), nil
}

func (a *App) reportForTestPlan(ctx context.Context, projectID string, runID string) (*Report, error) {
	if runID != "" {
		report, err := a.store.GetReport(ctx, runID)
		if err != nil {
			return nil, err
		}
		if report.ProjectID != projectID {
			return nil, fmt.Errorf("selected run does not belong to the project")
		}
		return report, nil
	}
	latestRun, err := a.store.GetLatestRunForProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return a.store.GetReport(ctx, latestRun.ID)
}

func (a *App) providerFromInput(input AIProviderRequest, existingEncryptedAPIKey string, existingEncryptedExtraHeaders string) (AIProvider, error) {
	encryptedAPIKey := existingEncryptedAPIKey
	if input.APIKey != "" {
		value, err := a.secretBox.Encrypt(input.APIKey)
		if err != nil {
			return AIProvider{}, err
		}
		encryptedAPIKey = value
	}
	encryptedHeaders := existingEncryptedExtraHeaders
	if input.ExtraHeaders != nil {
		rawHeaders, err := encodeExtraHeaders(input.ExtraHeaders)
		if err != nil {
			return AIProvider{}, err
		}
		value, err := a.secretBox.Encrypt(rawHeaders)
		if err != nil {
			return AIProvider{}, err
		}
		encryptedHeaders = value
	}
	return aiProviderFromRequest(input, encryptedAPIKey, encryptedHeaders), nil
}

func (a *App) providerForAnalysis(ctx context.Context, providerID string) (*AIProvider, error) {
	if providerID != "" {
		return a.store.GetAIProvider(ctx, providerID)
	}
	return a.store.GetDefaultAIProvider(ctx)
}

func (a *App) clientRequestForProvider(provider AIProvider, messages []AIChatMessage) (AIClientRequest, error) {
	apiKey, err := a.secretBox.Decrypt(provider.APIKeyEncrypted)
	if err != nil {
		return AIClientRequest{}, err
	}
	headersRaw, err := a.secretBox.Decrypt(provider.ExtraHeadersEncrypted)
	if err != nil {
		return AIClientRequest{}, err
	}
	headers, err := decodeExtraHeaders(headersRaw)
	if err != nil {
		return AIClientRequest{}, err
	}
	return AIClientRequest{
		BaseURL:         provider.BaseURL,
		Model:           provider.Model,
		APIKey:          apiKey,
		ExtraHeaders:    headers,
		Temperature:     provider.Temperature,
		MaxOutputTokens: provider.MaxOutputTokens,
		TimeoutSeconds:  provider.TimeoutSeconds,
		Messages:        messages,
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func withJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func withRequestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request handled", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}

func withCORS(allowedOrigins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && originAllowed(origin, allowedOrigins) {
			if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-Qualora-CSRF")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "600")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func originAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

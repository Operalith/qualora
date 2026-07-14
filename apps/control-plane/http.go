package main

import (
	"context"
	"encoding/json"
	"errors"
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
	logger        *slog.Logger
	corsOrigins   []string
}

func NewApp(store *Store, queue *Queue, evidenceStore *EvidenceStore, logger *slog.Logger, corsOrigins []string) *App {
	return &App{store: store, queue: queue, evidenceStore: evidenceStore, logger: logger, corsOrigins: corsOrigins}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/api/v1/projects", a.handleProjects)
	mux.HandleFunc("/api/v1/projects/", a.handleProjectSubroutes)
	mux.HandleFunc("/api/v1/runs", a.handleRuns)
	mux.HandleFunc("/api/v1/runs/", a.handleRunSubroutes)
	mux.HandleFunc("/api/v1/evidence/", a.handleEvidenceSubroutes)
	return withCORS(a.corsOrigins, withJSONContentType(withRequestLog(a.logger, mux)))
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

	run, jobs, err := a.store.CreateRunForKinds(r.Context(), *project, []string{JobKindBrowser})
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
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
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

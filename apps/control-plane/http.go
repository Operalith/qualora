package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type App struct {
	store  *Store
	queue  *Queue
	logger *slog.Logger
}

func NewApp(store *Store, queue *Queue, logger *slog.Logger) *App {
	return &App{store: store, queue: queue, logger: logger}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/api/v1/projects", a.handleProjects)
	mux.HandleFunc("/api/v1/projects/", a.handleProjectSubroutes)
	mux.HandleFunc("/api/v1/runs/", a.handleRunSubroutes)
	return withJSONContentType(withRequestLog(a.logger, mux))
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
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
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
	if len(parts) == 2 && parts[0] != "" && parts[1] == "runs" && r.Method == http.MethodPost {
		a.createRun(w, r, parts[0])
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

	if _, err := ValidateTargetURL(project.FrontendURL, project.AllowedHosts, project.AllowPrivateTargets); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_project_scope", err.Error())
		return
	}

	run, err := a.store.CreateRun(r.Context(), project.ID)
	if err != nil {
		a.logger.Error("create run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_run_failed", "run could not be created")
		return
	}

	if err := a.queue.EnqueueBrowserRun(r.Context(), BrowserRunJob{RunID: run.ID, ProjectID: project.ID}); err != nil {
		a.logger.Error("enqueue run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkRunFailed(r.Context(), run.ID, "run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
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

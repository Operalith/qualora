package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *App) createAIBrowserControlRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for AI Browser Control run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.FrontendURL == "" {
		writeError(w, http.StatusBadRequest, "frontend_url_required", "project must have frontend_url to start AI Browser Control")
		return
	}

	input, err := decodeAIBrowserControlRunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	normalized, err := NormalizeAIBrowserControlRunRequest(*project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ai_browser_control_run", err.Error())
		return
	}
	provider, err := a.store.GetAIProvider(r.Context(), normalized.ProviderID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusBadRequest, "ai_provider_not_found", "AI provider was not found")
			return
		}
		a.logger.Error("get AI provider for AI Browser Control failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_provider_failed", "AI provider could not be loaded")
		return
	}
	if provider.Type != AIProviderOpenAICompatible {
		writeError(w, http.StatusBadRequest, "unsupported_ai_provider", "AI Browser Control supports OpenAI-compatible providers only")
		return
	}
	if normalized.CredentialProfileID != "" {
		profile, err := a.store.GetCredentialProfile(r.Context(), normalized.CredentialProfileID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusBadRequest, "credential_profile_not_found", "credential profile was not found")
				return
			}
			a.logger.Error("get credential profile for AI Browser Control failed", "error", err)
			writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
			return
		}
		if profile.ProjectID != project.ID {
			writeError(w, http.StatusBadRequest, "credential_profile_project_mismatch", "credential profile does not belong to the project")
			return
		}
	}

	run, err := a.store.CreateAIBrowserControlRun(r.Context(), *project, normalized)
	if err != nil {
		a.logger.Error("create AI Browser Control run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_ai_browser_control_run_failed", "AI Browser Control run could not be created")
		return
	}
	if err := a.queue.EnqueueAIBrowserControlRun(r.Context(), AIBrowserControlRunJob{AIBrowserControlRunID: run.ID, ProjectID: project.ID}); err != nil {
		a.logger.Error("enqueue AI Browser Control run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkAIBrowserControlRunFailed(r.Context(), run.ID, "AI Browser Control run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "AI Browser Control run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (a *App) listAIBrowserControlRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for AI Browser Control list failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListAIBrowserControlRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list AI Browser Control runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_ai_browser_control_runs_failed", "AI Browser Control runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ai_browser_control_runs": runs})
}

func (a *App) handleAIBrowserControlRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/ai-browser-control-runs/")
	parts := stringsSplitPath(path)
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAIBrowserControlRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "trace" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAIBrowserControlTrace(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAIBrowserControlReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAIBrowserControlHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getAIBrowserControlRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := a.store.GetAIBrowserControlRun(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_browser_control_run_not_found", "AI Browser Control run was not found")
			return
		}
		a.logger.Error("get AI Browser Control run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_browser_control_run_failed", "AI Browser Control run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) getAIBrowserControlTrace(w http.ResponseWriter, r *http.Request, runID string) {
	trace, err := a.store.GetAIBrowserControlTrace(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_browser_control_run_not_found", "AI Browser Control run was not found")
			return
		}
		a.logger.Error("get AI Browser Control trace failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_browser_control_trace_failed", "AI Browser Control trace could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, trace)
}

func (a *App) getAIBrowserControlReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetAIBrowserControlReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_browser_control_run_not_found", "AI Browser Control run was not found")
			return
		}
		a.logger.Error("get AI Browser Control report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_browser_control_report_failed", "AI Browser Control report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getAIBrowserControlHTMLReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetAIBrowserControlReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ai_browser_control_run_not_found", "AI Browser Control run was not found")
			return
		}
		a.logger.Error("get AI Browser Control html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_browser_control_report_failed", "AI Browser Control report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderAIBrowserControlHTMLReport(w, report, time.Now().UTC()); err != nil {
		a.logger.Error("render AI Browser Control html report failed", "error", err)
	}
}

func decodeAIBrowserControlRunRequest(w http.ResponseWriter, r *http.Request) (AIBrowserControlRunRequest, error) {
	var input AIBrowserControlRunRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if r.Body == http.NoBody {
		return input, nil
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		if errors.Is(err, io.EOF) {
			return input, nil
		}
		return input, fmt.Errorf("request body must be valid AI Browser Control run JSON")
	}
	return input, nil
}

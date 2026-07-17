package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *App) createSafeExplorerRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for safe explorer run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.FrontendURL == "" {
		writeError(w, http.StatusBadRequest, "frontend_url_required", "project must have frontend_url to start Safe Explorer")
		return
	}

	input, err := decodeSafeExplorerRunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	normalized, err := NormalizeSafeExplorerRunRequest(*project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_safe_explorer_run", err.Error())
		return
	}
	if normalized.CredentialProfileID != "" {
		profile, err := a.store.GetCredentialProfile(r.Context(), normalized.CredentialProfileID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusBadRequest, "credential_profile_not_found", "credential profile was not found")
				return
			}
			a.logger.Error("get credential profile for safe explorer run failed", "error", err)
			writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
			return
		}
		if profile.ProjectID != project.ID {
			writeError(w, http.StatusBadRequest, "credential_profile_project_mismatch", "credential profile does not belong to the project")
			return
		}
	}

	run, err := a.store.CreateSafeExplorerRun(r.Context(), *project, normalized)
	if err != nil {
		a.logger.Error("create safe explorer run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_safe_explorer_run_failed", "Safe Explorer run could not be created")
		return
	}
	if err := a.queue.EnqueueSafeExplorerRun(r.Context(), SafeExplorerRunJob{SafeExplorerRunID: run.ID, ProjectID: project.ID}); err != nil {
		a.logger.Error("enqueue safe explorer run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkSafeExplorerRunFailed(r.Context(), run.ID, "Safe Explorer run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "Safe Explorer run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (a *App) listSafeExplorerRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for safe explorer list failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListSafeExplorerRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list safe explorer runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_safe_explorer_runs_failed", "Safe Explorer runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"safe_explorer_runs": runs})
}

func (a *App) handleSafeExplorerRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/safe-explorer-runs/")
	parts := stringsSplitPath(path)
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getSafeExplorerRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "trace" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getSafeExplorerTrace(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getSafeExplorerReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getSafeExplorerHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getSafeExplorerRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := a.store.GetSafeExplorerRun(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "safe_explorer_run_not_found", "Safe Explorer run was not found")
			return
		}
		a.logger.Error("get safe explorer run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_safe_explorer_run_failed", "Safe Explorer run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) getSafeExplorerTrace(w http.ResponseWriter, r *http.Request, runID string) {
	trace, err := a.store.GetSafeExplorerTrace(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "safe_explorer_run_not_found", "Safe Explorer run was not found")
			return
		}
		a.logger.Error("get safe explorer trace failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_safe_explorer_trace_failed", "Safe Explorer trace could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, trace)
}

func (a *App) getSafeExplorerReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetSafeExplorerReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "safe_explorer_run_not_found", "Safe Explorer run was not found")
			return
		}
		a.logger.Error("get safe explorer report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_safe_explorer_report_failed", "Safe Explorer report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getSafeExplorerHTMLReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetSafeExplorerReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "safe_explorer_run_not_found", "Safe Explorer run was not found")
			return
		}
		a.logger.Error("get safe explorer html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_safe_explorer_report_failed", "Safe Explorer report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderSafeExplorerHTMLReport(w, report, time.Now().UTC()); err != nil {
		a.logger.Error("render safe explorer html report failed", "error", err)
	}
}

func decodeSafeExplorerRunRequest(w http.ResponseWriter, r *http.Request) (SafeExplorerRunRequest, error) {
	var input SafeExplorerRunRequest
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
		return input, fmt.Errorf("request body must be valid Safe Explorer run JSON")
	}
	return input, nil
}

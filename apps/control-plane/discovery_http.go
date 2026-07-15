package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (a *App) createDiscoveryRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for discovery run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.FrontendURL == "" {
		writeError(w, http.StatusBadRequest, "frontend_url_required", "project must have frontend_url to start discovery")
		return
	}

	input, err := decodeDiscoveryRunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	normalized, err := NormalizeDiscoveryRunRequest(*project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_discovery_run", err.Error())
		return
	}
	if normalized.CredentialProfileID != "" {
		profile, err := a.store.GetCredentialProfile(r.Context(), normalized.CredentialProfileID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusBadRequest, "credential_profile_not_found", "credential profile was not found")
				return
			}
			a.logger.Error("get credential profile for discovery run failed", "error", err)
			writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
			return
		}
		if profile.ProjectID != project.ID {
			writeError(w, http.StatusBadRequest, "credential_profile_project_mismatch", "credential profile does not belong to the project")
			return
		}
	}

	run, err := a.store.CreateDiscoveryRun(r.Context(), *project, normalized)
	if err != nil {
		a.logger.Error("create discovery run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_discovery_run_failed", "discovery run could not be created")
		return
	}
	if err := a.queue.EnqueueDiscoveryRun(r.Context(), DiscoveryRunJob{DiscoveryRunID: run.ID, ProjectID: project.ID}); err != nil {
		a.logger.Error("enqueue discovery run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkDiscoveryRunFailed(r.Context(), run.ID, "discovery run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "discovery run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (a *App) listDiscoveryRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for discovery list failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListDiscoveryRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list discovery runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_discovery_runs_failed", "discovery runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"discovery_runs": runs})
}

func (a *App) handleDiscoveryRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/discovery-runs/")
	parts := stringsSplitPath(path)
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getDiscoveryRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "map" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getDiscoveryMap(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getDiscoveryReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getDiscoveryHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getDiscoveryRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := a.store.GetDiscoveryRun(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "discovery_run_not_found", "discovery run was not found")
			return
		}
		a.logger.Error("get discovery run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_discovery_run_failed", "discovery run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) getDiscoveryMap(w http.ResponseWriter, r *http.Request, runID string) {
	discoveryMap, err := a.store.GetDiscoveryMap(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "discovery_run_not_found", "discovery run was not found")
			return
		}
		a.logger.Error("get discovery map failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_discovery_map_failed", "discovery map could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, discoveryMap)
}

func (a *App) getDiscoveryReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetDiscoveryReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "discovery_run_not_found", "discovery run was not found")
			return
		}
		a.logger.Error("get discovery report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_discovery_report_failed", "discovery report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getDiscoveryHTMLReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetDiscoveryReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "discovery_run_not_found", "discovery run was not found")
			return
		}
		a.logger.Error("get discovery html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_discovery_report_failed", "discovery report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderDiscoveryHTMLReport(w, report, time.Now().UTC()); err != nil {
		a.logger.Error("render discovery html report failed", "error", err)
	}
}

func decodeDiscoveryRunRequest(w http.ResponseWriter, r *http.Request) (DiscoveryRunRequest, error) {
	var input DiscoveryRunRequest
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
		return input, fmt.Errorf("request body must be valid discovery run JSON")
	}
	return input, nil
}

func stringsTrimPrefix(value string, prefix string) string {
	if len(value) >= len(prefix) && value[:len(prefix)] == prefix {
		return value[len(prefix):]
	}
	return value
}

func stringsSplitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

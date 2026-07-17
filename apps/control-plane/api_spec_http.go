package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func (a *App) createAPISpec(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for api spec failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}

	var input APISpecImportRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxStoredSpecBytes+65536)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid API spec import JSON")
		return
	}
	input, err = NormalizeAPISpecImportRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_api_spec_import", err.Error())
		return
	}

	detail, err := a.createAPISpecFromInput(r.Context(), *project, input)
	if err != nil {
		a.logger.Error("create api spec failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_api_spec_failed", "API spec could not be imported")
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

func (a *App) createAPISpecFromInput(ctx context.Context, project Project, input APISpecImportRequest) (*APISpecDetail, error) {
	rawSpec := input.RawSpec
	sourceStatus := "parsed"
	errorMessage := ""
	if input.SourceType == "url" {
		var normalizedURL string
		var err error
		rawSpec, normalizedURL, err = FetchOpenAPISource(ctx, project, input.SourceURL)
		if normalizedURL != "" {
			input.SourceURL = normalizedURL
		}
		if err != nil {
			sourceStatus = "error"
			errorMessage = RedactSecrets(err.Error())
		}
	}

	var parsed *parsedOpenAPISpec
	if sourceStatus == "parsed" {
		var err error
		parsed, err = ParseOpenAPISpec(rawSpec, input.SourceURL, project.APIBaseURL)
		if err != nil {
			sourceStatus = "invalid"
			errorMessage = RedactSecrets(err.Error())
		}
	}

	detail, err := a.store.CreateAPISpec(ctx, project.ID, input, rawSpec, parsed, sourceStatus, errorMessage)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (a *App) listAPISpecs(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for api specs failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	specs, err := a.store.ListAPISpecs(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list api specs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_api_specs_failed", "API specs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"api_specs": specs})
}

func (a *App) handleAPISpecSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/api-specs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodGet:
			a.getAPISpec(w, r, parts[0])
		case http.MethodDelete:
			a.deleteAPISpec(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "operations" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.listAPIOperations(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "api-smoke-runs" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.createAPISmokeRun(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getAPISpec(w http.ResponseWriter, r *http.Request, apiSpecID string) {
	detail, err := a.store.GetAPISpecDetail(r.Context(), apiSpecID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_spec_not_found", "API spec was not found")
			return
		}
		a.logger.Error("get api spec failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_api_spec_failed", "API spec could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (a *App) listAPIOperations(w http.ResponseWriter, r *http.Request, apiSpecID string) {
	if _, err := a.store.GetAPISpec(r.Context(), apiSpecID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_spec_not_found", "API spec was not found")
			return
		}
		a.logger.Error("get api spec for operations failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_api_spec_failed", "API spec could not be loaded")
		return
	}
	operations, err := a.store.ListAPIOperations(r.Context(), apiSpecID)
	if err != nil {
		a.logger.Error("list api operations failed", "api_spec_id", apiSpecID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_api_operations_failed", "API operations could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"operations": operations})
}

func (a *App) deleteAPISpec(w http.ResponseWriter, r *http.Request, apiSpecID string) {
	if err := a.store.DeleteAPISpec(r.Context(), apiSpecID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_spec_not_found", "API spec was not found")
			return
		}
		a.logger.Error("delete api spec failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_api_spec_failed", "API spec could not be deleted")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *App) createAPISmokeRun(w http.ResponseWriter, r *http.Request, apiSpecID string) {
	spec, err := a.store.GetAPISpec(r.Context(), apiSpecID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_spec_not_found", "API spec was not found")
			return
		}
		a.logger.Error("get api spec for smoke run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_api_spec_failed", "API spec could not be loaded")
		return
	}
	if spec.Status != "parsed" {
		writeError(w, http.StatusBadRequest, "api_spec_not_parsed", "only parsed API specs can be executed")
		return
	}
	project, err := a.store.GetProject(r.Context(), spec.ProjectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for api smoke run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.DestructiveActions {
		writeError(w, http.StatusBadRequest, "destructive_actions_not_supported", "safe API smoke runs require destructive_actions=false")
		return
	}
	operations, err := a.store.ListAPIOperations(r.Context(), apiSpecID)
	if err != nil {
		a.logger.Error("list operations for api smoke run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "list_api_operations_failed", "API operations could not be loaded")
		return
	}
	run, err := a.ExecuteAPISmokeRun(r.Context(), *project, *spec, operations)
	if err != nil {
		a.logger.Error("api smoke run failed", "api_spec_id", apiSpecID, "error", err)
		writeError(w, http.StatusInternalServerError, "api_smoke_run_failed", "API smoke run could not complete")
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

func (a *App) getAPIResults(w http.ResponseWriter, r *http.Request, runID string) {
	if _, err := a.store.GetRun(r.Context(), runID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "run_not_found", "run was not found")
			return
		}
		a.logger.Error("get run for api results failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_run_failed", "run could not be loaded")
		return
	}
	results, err := a.store.ListAPICheckResults(r.Context(), runID)
	if err != nil {
		a.logger.Error("list api results failed", "run_id", runID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_api_results_failed", "API results could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"api_results": results})
}

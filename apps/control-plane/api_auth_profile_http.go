package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (a *App) handleAPIAuthProfileSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/api-auth-profiles/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodGet:
			a.getAPIAuthProfile(w, r, parts[0])
		case http.MethodPut:
			a.updateAPIAuthProfile(w, r, parts[0])
		case http.MethodDelete:
			a.deleteAPIAuthProfile(w, r, parts[0])
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
		a.testAPIAuthProfileHandler(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) listAPIAuthProfiles(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for API auth profiles failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	profiles, err := a.store.ListAPIAuthProfiles(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list API auth profiles failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_api_auth_profiles_failed", "API auth profiles could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"api_auth_profiles": profiles})
}

func (a *App) createAPIAuthProfile(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for API auth profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeAPIAuthProfileRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeAPIAuthProfileRequest(input, true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_api_auth_profile", err.Error())
		return
	}
	profile, err := a.apiAuthProfileFromInput(input, nil)
	if err != nil {
		a.logger.Error("encrypt API auth profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "encrypt_api_auth_profile_failed", "API auth profile could not be encrypted")
		return
	}
	created, err := a.store.CreateAPIAuthProfile(r.Context(), projectID, profile)
	if err != nil {
		a.logger.Error("create API auth profile failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "create_api_auth_profile_failed", "API auth profile could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *App) getAPIAuthProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	profile, err := a.store.GetAPIAuthProfile(r.Context(), profileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_auth_profile_not_found", "API auth profile was not found")
			return
		}
		a.logger.Error("get API auth profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_api_auth_profile_failed", "API auth profile could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (a *App) updateAPIAuthProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	current, err := a.store.GetAPIAuthProfile(r.Context(), profileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_auth_profile_not_found", "API auth profile was not found")
			return
		}
		a.logger.Error("get API auth profile for update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_api_auth_profile_failed", "API auth profile could not be loaded")
		return
	}
	input, err := decodeAPIAuthProfileRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeAPIAuthProfileRequest(input, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_api_auth_profile", err.Error())
		return
	}
	profile, err := a.apiAuthProfileFromInput(input, current)
	if err != nil {
		a.logger.Error("encrypt API auth profile update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "encrypt_api_auth_profile_failed", "API auth profile could not be encrypted")
		return
	}
	updated, err := a.store.UpdateAPIAuthProfile(r.Context(), profileID, profile)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_auth_profile_not_found", "API auth profile was not found")
			return
		}
		a.logger.Error("update API auth profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "update_api_auth_profile_failed", "API auth profile could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *App) deleteAPIAuthProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	if err := a.store.DeleteAPIAuthProfile(r.Context(), profileID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_auth_profile_not_found", "API auth profile was not found")
			return
		}
		a.logger.Error("delete API auth profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_api_auth_profile_failed", "API auth profile could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) testAPIAuthProfileHandler(w http.ResponseWriter, r *http.Request, profileID string) {
	profile, err := a.store.GetAPIAuthProfile(r.Context(), profileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "api_auth_profile_not_found", "API auth profile was not found")
			return
		}
		a.logger.Error("get API auth profile for test failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_api_auth_profile_failed", "API auth profile could not be loaded")
		return
	}
	project, err := a.store.GetProject(r.Context(), profile.ProjectID)
	if err != nil {
		a.logger.Error("get project for API auth profile test failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeAPIAuthProfileTestRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeAPIAuthProfileTestRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_api_auth_profile_test", err.Error())
		return
	}
	result := a.testAPIAuthProfile(r.Context(), *project, *profile, input)
	status := http.StatusOK
	if !result.Success {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

func decodeAPIAuthProfileRequest(w http.ResponseWriter, r *http.Request) (APIAuthProfileRequest, error) {
	var input APIAuthProfileRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		if errors.Is(err, io.EOF) {
			return input, errors.New("request body is required")
		}
		return input, errors.New("request body must be valid API auth profile JSON")
	}
	return input, nil
}

func decodeAPIAuthProfileTestRequest(w http.ResponseWriter, r *http.Request) (APIAuthProfileTestRequest, error) {
	var input APIAuthProfileTestRequest
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil && !errors.Is(err, io.EOF) {
		return input, errors.New("request body must be valid API auth profile test JSON")
	}
	return input, nil
}

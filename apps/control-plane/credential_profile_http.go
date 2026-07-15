package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (a *App) handleCredentialProfileSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/credential-profiles/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodGet:
			a.getCredentialProfile(w, r, parts[0])
		case http.MethodPut:
			a.updateCredentialProfile(w, r, parts[0])
		case http.MethodDelete:
			a.deleteCredentialProfile(w, r, parts[0])
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "test-login" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.testCredentialProfileLogin(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) listCredentialProfiles(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for credential profiles failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	profiles, err := a.store.ListCredentialProfiles(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list credential profiles failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_credential_profiles_failed", "credential profiles could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"credential_profiles": profiles})
}

func (a *App) createCredentialProfile(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for credential profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}

	input, err := decodeCredentialProfileRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeCredentialProfileRequest(input, *project, true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_credential_profile", err.Error())
		return
	}
	profile, err := a.credentialProfileFromInput(input, "", "", "")
	if err != nil {
		a.logger.Error("encrypt credential profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "encrypt_credential_profile_failed", "credential profile could not be encrypted")
		return
	}
	created, err := a.store.CreateCredentialProfile(r.Context(), project.ID, profile)
	if err != nil {
		a.logger.Error("create credential profile failed", "project_id", project.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "create_credential_profile_failed", "credential profile could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *App) getCredentialProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	profile, err := a.store.GetCredentialProfile(r.Context(), profileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "credential_profile_not_found", "credential profile was not found")
			return
		}
		a.logger.Error("get credential profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (a *App) updateCredentialProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	current, err := a.store.GetCredentialProfile(r.Context(), profileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "credential_profile_not_found", "credential profile was not found")
			return
		}
		a.logger.Error("get credential profile for update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
		return
	}
	project, err := a.store.GetProject(r.Context(), current.ProjectID)
	if err != nil {
		a.logger.Error("get project for credential profile update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}

	input, err := decodeCredentialProfileRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeCredentialProfileRequest(input, *project, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_credential_profile", err.Error())
		return
	}
	profile, err := a.credentialProfileFromInput(input, current.UsernameEncrypted, current.PasswordEncrypted, current.UsernameDisplayHint)
	if err != nil {
		a.logger.Error("encrypt credential profile update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "encrypt_credential_profile_failed", "credential profile could not be encrypted")
		return
	}
	profile.ProjectID = current.ProjectID
	updated, err := a.store.UpdateCredentialProfile(r.Context(), profileID, profile)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "credential_profile_not_found", "credential profile was not found")
			return
		}
		a.logger.Error("update credential profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "update_credential_profile_failed", "credential profile could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *App) deleteCredentialProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	if err := a.store.DeleteCredentialProfile(r.Context(), profileID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "credential_profile_not_found", "credential profile was not found")
			return
		}
		a.logger.Error("delete credential profile failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_credential_profile_failed", "credential profile could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) testCredentialProfileLogin(w http.ResponseWriter, r *http.Request, profileID string) {
	profile, err := a.store.GetCredentialProfile(r.Context(), profileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "credential_profile_not_found", "credential profile was not found")
			return
		}
		a.logger.Error("get credential profile for test-login failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
		return
	}
	project, err := a.store.GetProject(r.Context(), profile.ProjectID)
	if err != nil {
		a.logger.Error("get project for test-login failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	run, jobs, err := a.store.CreateRunForKindsWithOptions(r.Context(), *project, []string{JobKindBrowser}, RunOptions{
		RunType:             RunTypeLoginCheck,
		CredentialProfileID: profile.ID,
		CaptureScreenshot:   true,
		MaxDurationSeconds:  30,
	})
	if err != nil {
		a.logger.Error("create login check run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_run_failed", "login check run could not be created")
		return
	}
	if err := a.enqueueRunJobs(r.Context(), *project, run, jobs); err != nil {
		a.logger.Error("enqueue login check run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkRunFailed(r.Context(), run.ID, "run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "login check run could not be queued")
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

func (a *App) createAuthenticatedBrowserSmokeRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for authenticated browser smoke run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.FrontendURL == "" {
		writeError(w, http.StatusBadRequest, "frontend_url_required", "project must have frontend_url to start an authenticated browser smoke run")
		return
	}

	input, err := decodeAuthenticatedBrowserSmokeRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeAuthenticatedBrowserSmokeRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_authenticated_browser_smoke", err.Error())
		return
	}

	var profile *CredentialProfile
	if input.CredentialProfileID != "" {
		profile, err = a.store.GetCredentialProfile(r.Context(), input.CredentialProfileID)
	} else {
		profile, err = a.store.GetDefaultCredentialProfile(r.Context(), project.ID)
	}
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusBadRequest, "credential_profile_required", "create a credential profile to enable authenticated browser smoke testing")
			return
		}
		a.logger.Error("get credential profile for authenticated browser smoke failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
		return
	}
	if profile.ProjectID != project.ID {
		writeError(w, http.StatusBadRequest, "credential_profile_project_mismatch", "credential profile does not belong to the project")
		return
	}

	captureScreenshot := true
	if input.CaptureScreenshot != nil {
		captureScreenshot = *input.CaptureScreenshot
	}
	run, jobs, err := a.store.CreateRunForKindsWithOptions(r.Context(), *project, []string{JobKindBrowser}, RunOptions{
		RunType:             RunTypeAuthenticatedBrowserSmoke,
		CredentialProfileID: profile.ID,
		TargetPath:          input.TargetPath,
		CaptureScreenshot:   captureScreenshot,
		MaxDurationSeconds:  input.MaxDurationSeconds,
	})
	if err != nil {
		a.logger.Error("create authenticated browser smoke run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_run_failed", "authenticated browser smoke run could not be created")
		return
	}
	if err := a.enqueueRunJobs(r.Context(), *project, run, jobs); err != nil {
		a.logger.Error("enqueue authenticated browser smoke run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkRunFailed(r.Context(), run.ID, "run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "authenticated browser smoke run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func decodeCredentialProfileRequest(w http.ResponseWriter, r *http.Request) (CredentialProfileRequest, error) {
	var input CredentialProfileRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, fmt.Errorf("request body must be valid credential profile JSON")
	}
	return input, nil
}

func decodeAuthenticatedBrowserSmokeRequest(w http.ResponseWriter, r *http.Request) (AuthenticatedBrowserSmokeRequest, error) {
	var input AuthenticatedBrowserSmokeRequest
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
		return input, fmt.Errorf("request body must be valid authenticated browser smoke JSON")
	}
	return input, nil
}

func normalizeAuthenticatedBrowserSmokeRequest(input AuthenticatedBrowserSmokeRequest) (AuthenticatedBrowserSmokeRequest, error) {
	input.CredentialProfileID = strings.TrimSpace(input.CredentialProfileID)
	input.TargetPath = strings.TrimSpace(input.TargetPath)
	if input.TargetPath == "" {
		input.TargetPath = "/"
	}
	parsed, err := url.Parse(input.TargetPath)
	if err != nil {
		return input, fmt.Errorf("target_path is invalid")
	}
	if parsed.IsAbs() || strings.HasPrefix(input.TargetPath, "//") {
		return input, fmt.Errorf("target_path must be a relative project path")
	}
	if !strings.HasPrefix(input.TargetPath, "/") {
		input.TargetPath = "/" + input.TargetPath
	}
	if hasSensitiveTargetPathQuery(input.TargetPath) {
		return input, fmt.Errorf("target_path query contains sensitive parameter names")
	}
	if input.MaxDurationSeconds == 0 {
		input.MaxDurationSeconds = 30
	}
	if input.MaxDurationSeconds < 5 || input.MaxDurationSeconds > 120 {
		return input, fmt.Errorf("max_duration_seconds must be between 5 and 120")
	}
	return input, nil
}

func hasSensitiveTargetPathQuery(targetPath string) bool {
	parsed, err := url.Parse(targetPath)
	if err != nil {
		return true
	}
	for name := range parsed.Query() {
		normalized := strings.ToLower(name)
		for _, marker := range []string{"authorization", "password", "passwd", "token", "secret", "api_key", "apikey", "cookie", "session"} {
			if strings.Contains(normalized, marker) {
				return true
			}
		}
	}
	return false
}

func (a *App) credentialProfileFromInput(input CredentialProfileRequest, existingEncryptedUsername string, existingEncryptedPassword string, existingUsernameHint string) (CredentialProfile, error) {
	encryptedUsername := existingEncryptedUsername
	usernameHint := existingUsernameHint
	if strings.TrimSpace(input.Username) != "" {
		value, err := a.secretBox.Encrypt(input.Username)
		if err != nil {
			return CredentialProfile{}, err
		}
		encryptedUsername = value
		usernameHint = usernameDisplayHint(input.Username)
	}
	encryptedPassword := existingEncryptedPassword
	if input.Password != "" {
		value, err := a.secretBox.Encrypt(input.Password)
		if err != nil {
			return CredentialProfile{}, err
		}
		encryptedPassword = value
	}
	return credentialProfileFromRequest(input, encryptedUsername, encryptedPassword, usernameHint), nil
}

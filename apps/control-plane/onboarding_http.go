package main

import (
	"encoding/json"
	"net/http"
)

func (a *App) handleOnboardingProjectSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/onboarding/project-setup" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}

	var input ProjectSetupRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxStoredSpecBytes+(1<<20))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid project setup JSON")
		return
	}

	response, err := a.RunProjectSetup(r.Context(), input)
	if err != nil {
		a.logger.Error("guided project setup failed", "error", RedactSecrets(err.Error()))
		writeError(w, http.StatusBadRequest, "project_setup_failed", RedactSecrets(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

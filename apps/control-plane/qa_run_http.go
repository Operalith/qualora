package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (a *App) createQARun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for QA run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeQARunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_qa_run", err.Error())
		return
	}
	normalized, err := NormalizeQARunRequest(*project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_qa_run", err.Error())
		return
	}
	if normalized.CredentialProfileID != "" {
		profile, err := a.store.GetCredentialProfile(r.Context(), normalized.CredentialProfileID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusBadRequest, "credential_profile_not_found", "credential profile was not found")
				return
			}
			a.logger.Error("get credential profile for QA run failed", "error", err)
			writeError(w, http.StatusInternalServerError, "get_credential_profile_failed", "credential profile could not be loaded")
			return
		}
		if profile.ProjectID != project.ID {
			writeError(w, http.StatusBadRequest, "credential_profile_project_mismatch", "credential profile does not belong to the project")
			return
		}
	}
	if normalized.APIAuthProfileID != "" {
		profile, err := a.store.GetAPIAuthProfile(r.Context(), normalized.APIAuthProfileID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusBadRequest, "api_auth_profile_not_found", "API auth profile was not found")
				return
			}
			a.logger.Error("get API auth profile for QA run failed", "error", err)
			writeError(w, http.StatusInternalServerError, "get_api_auth_profile_failed", "API auth profile could not be loaded")
			return
		}
		if profile.ProjectID != project.ID {
			writeError(w, http.StatusBadRequest, "api_auth_profile_project_mismatch", "API auth profile does not belong to the project")
			return
		}
	}
	if _, err := a.providerForAnalysis(r.Context(), normalized.ProviderID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusBadRequest, "ai_provider_required", "configure an AI provider before starting a safe QA run")
			return
		}
		a.logger.Error("load AI provider for QA run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ai_provider_failed", "AI provider could not be loaded")
		return
	}

	qaRun, err := a.store.CreateQARun(r.Context(), project.ID, normalized)
	if err != nil {
		a.logger.Error("create QA run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_qa_run_failed", "QA run could not be created")
		return
	}
	go a.runSafeQARun(qaRun.ID, *project, normalized)
	writeJSON(w, http.StatusCreated, qaRun)
}

func (a *App) listQARuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for QA runs failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListQARuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list QA runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_qa_runs_failed", "QA runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"qa_runs": runs})
}

func (a *App) handleQARunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/qa-runs/")
	parts := stringsSplitPath(path)
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getQARun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "execute" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.executeQARun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getQARunReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getQARunHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getQARun(w http.ResponseWriter, r *http.Request, qaRunID string) {
	run, err := a.store.GetQARun(r.Context(), qaRunID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "qa_run_not_found", "QA run was not found")
			return
		}
		a.logger.Error("get QA run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_qa_run_failed", "QA run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) executeQARun(w http.ResponseWriter, r *http.Request, qaRunID string) {
	run, err := a.store.GetQARun(r.Context(), qaRunID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "qa_run_not_found", "QA run was not found")
			return
		}
		a.logger.Error("get QA run for execute failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_qa_run_failed", "QA run could not be loaded")
		return
	}
	if run.Status != StatusCompleted || run.TestPlanID == "" || run.TestPlanExecutionID != "" {
		writeError(w, http.StatusBadRequest, "qa_run_not_executable", "QA run must be a completed preview with a test plan and no execution")
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), qaRunTimeout)
		defer cancel()
		if err := a.executeExistingQARun(ctx, qaRunID); err != nil {
			_, _ = a.store.FailQARun(ctx, qaRunID, err.Error(), map[string]any{"error": RedactSecrets(err.Error())})
			a.logger.Error("execute previewed QA run failed", "qa_run_id", qaRunID, "error", RedactSecrets(err.Error()))
		}
	}()
	writeJSON(w, http.StatusAccepted, run)
}

func (a *App) getQARunReport(w http.ResponseWriter, r *http.Request, qaRunID string) {
	report, err := a.store.GetQARunReport(r.Context(), qaRunID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "qa_run_not_found", "QA run was not found")
			return
		}
		a.logger.Error("get QA run report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_qa_run_report_failed", "QA run report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getQARunHTMLReport(w http.ResponseWriter, r *http.Request, qaRunID string) {
	report, err := a.store.GetQARunReport(r.Context(), qaRunID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "qa_run_not_found", "QA run was not found")
			return
		}
		a.logger.Error("get QA run html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_qa_run_report_failed", "QA run report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderQARunHTMLReport(w, report); err != nil {
		a.logger.Error("render QA run html report failed", "error", err)
	}
}

func decodeQARunRequest(w http.ResponseWriter, r *http.Request) (QARunRequest, error) {
	var input QARunRequest
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
		return input, fmt.Errorf("request body must be valid QA run JSON")
	}
	return input, nil
}

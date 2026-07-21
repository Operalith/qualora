package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *App) createFormTestRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for form test run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.FrontendURL == "" {
		writeError(w, http.StatusBadRequest, "frontend_url_required", "project must have frontend_url to start Safe Form Testing")
		return
	}

	input, err := decodeFormTestRunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := a.resolveFormTestReferences(r, *project, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_form_test_run", err.Error())
		return
	}
	normalized, err := NormalizeFormTestRunRequest(*project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_form_test_run", err.Error())
		return
	}

	run, err := a.store.CreateFormTestRun(r.Context(), *project, normalized)
	if err != nil {
		a.logger.Error("create form test run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_form_test_run_failed", "Safe Form Testing run could not be created")
		return
	}
	if err := a.queue.EnqueueFormTestRun(r.Context(), FormTestRunJob{FormTestRunID: run.ID, ProjectID: project.ID}); err != nil {
		a.logger.Error("enqueue form test run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkFormTestRunFailed(r.Context(), run.ID, "Safe Form Testing run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "Safe Form Testing run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (a *App) listFormTestRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for form test list failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListFormTestRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list form test runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_form_test_runs_failed", "Safe Form Testing runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"form_test_runs": runs})
}

func (a *App) handleFormTestRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/form-test-runs/")
	parts := stringsSplitPath(path)
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getFormTestRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getFormTestReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getFormTestHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getFormTestRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := a.store.GetFormTestRun(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "form_test_run_not_found", "Safe Form Testing run was not found")
			return
		}
		a.logger.Error("get form test run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_form_test_run_failed", "Safe Form Testing run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) getFormTestReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetFormTestReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "form_test_run_not_found", "Safe Form Testing run was not found")
			return
		}
		a.logger.Error("get form test report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_form_test_report_failed", "Safe Form Testing report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getFormTestHTMLReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetFormTestReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "form_test_run_not_found", "Safe Form Testing run was not found")
			return
		}
		a.logger.Error("get form test html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_form_test_report_failed", "Safe Form Testing report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderFormTestHTMLReport(w, report, time.Now().UTC()); err != nil {
		a.logger.Error("render form test html report failed", "error", err)
	}
}

func (a *App) resolveFormTestReferences(r *http.Request, project Project, input *FormTestRunRequest) error {
	if input.UseLatestDiscovery && input.DiscoveryRunID == "" {
		run, err := a.store.GetLatestCompletedDiscoveryRun(r.Context(), project.ID)
		if err == nil {
			input.DiscoveryRunID = run.ID
		} else if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
	}
	if input.DiscoveryRunID != "" {
		run, err := a.store.GetDiscoveryRun(r.Context(), input.DiscoveryRunID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("discovery run was not found")
			}
			return err
		}
		if run.ProjectID != project.ID {
			return fmt.Errorf("discovery run does not belong to the project")
		}
		if run.Status != StatusCompleted {
			return fmt.Errorf("discovery run must be completed before Safe Form Testing can reuse it")
		}
	}
	if input.CredentialProfileID != "" {
		profile, err := a.store.GetCredentialProfile(r.Context(), input.CredentialProfileID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("credential profile was not found")
			}
			return err
		}
		if profile.ProjectID != project.ID {
			return fmt.Errorf("credential profile does not belong to the project")
		}
	}
	return nil
}

func decodeFormTestRunRequest(w http.ResponseWriter, r *http.Request) (FormTestRunRequest, error) {
	var input FormTestRunRequest
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
		return input, fmt.Errorf("request body must be valid Safe Form Testing run JSON")
	}
	return input, nil
}

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (a *App) handleAuthorizationCheckSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/authorization-checks/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 1 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.getAuthorizationCheck(w, r, parts[0])
	case http.MethodPut:
		a.updateAuthorizationCheck(w, r, parts[0])
	case http.MethodDelete:
		a.deleteAuthorizationCheck(w, r, parts[0])
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (a *App) handleAuthorizationCheckRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/authorization-check-runs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAuthorizationCheckRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAuthorizationCheckReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getAuthorizationCheckHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) listAuthorizationChecks(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for authorization checks failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	checks, err := a.store.ListAuthorizationChecks(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list authorization checks failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_authorization_checks_failed", "authorization checks could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authorization_checks": checks})
}

func (a *App) createAuthorizationCheck(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for authorization check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeAuthorizationCheckRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeAuthorizationCheckRequest(input, *project)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_authorization_check", err.Error())
		return
	}
	if err := a.validateAuthorizationCheckProfiles(r.Context(), project.ID, input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_authorization_check", err.Error())
		return
	}
	check, err := a.store.CreateAuthorizationCheck(r.Context(), authorizationCheckFromRequest(project.ID, input))
	if err != nil {
		a.logger.Error("create authorization check failed", "project_id", project.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "create_authorization_check_failed", "authorization check could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, check)
}

func (a *App) getAuthorizationCheck(w http.ResponseWriter, r *http.Request, checkID string) {
	check, err := a.store.GetAuthorizationCheck(r.Context(), checkID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "authorization_check_not_found", "authorization check was not found")
			return
		}
		a.logger.Error("get authorization check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_authorization_check_failed", "authorization check could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, check)
}

func (a *App) updateAuthorizationCheck(w http.ResponseWriter, r *http.Request, checkID string) {
	current, err := a.store.GetAuthorizationCheck(r.Context(), checkID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "authorization_check_not_found", "authorization check was not found")
			return
		}
		a.logger.Error("get authorization check for update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_authorization_check_failed", "authorization check could not be loaded")
		return
	}
	project, err := a.store.GetProject(r.Context(), current.ProjectID)
	if err != nil {
		a.logger.Error("get project for authorization check update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeAuthorizationCheckRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeAuthorizationCheckRequest(input, *project)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_authorization_check", err.Error())
		return
	}
	if err := a.validateAuthorizationCheckProfiles(r.Context(), project.ID, input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_authorization_check", err.Error())
		return
	}
	check := authorizationCheckFromRequest(project.ID, input)
	updated, err := a.store.UpdateAuthorizationCheck(r.Context(), checkID, check)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "authorization_check_not_found", "authorization check was not found")
			return
		}
		a.logger.Error("update authorization check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "update_authorization_check_failed", "authorization check could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *App) deleteAuthorizationCheck(w http.ResponseWriter, r *http.Request, checkID string) {
	if err := a.store.DeleteAuthorizationCheck(r.Context(), checkID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "authorization_check_not_found", "authorization check was not found")
			return
		}
		a.logger.Error("delete authorization check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_authorization_check_failed", "authorization check could not be deleted")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) createAuthorizationCheckRun(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for authorization check run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeAuthorizationCheckRunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeAuthorizationRunRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_authorization_check_run", err.Error())
		return
	}
	run, err := a.store.CreateAuthorizationCheckRun(r.Context(), projectID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "create_authorization_check_run_failed", err.Error())
		return
	}
	if err := a.queue.EnqueueAuthorizationCheckRun(r.Context(), AuthorizationCheckRunJob{AuthorizationCheckRunID: run.ID, ProjectID: projectID}); err != nil {
		a.logger.Error("enqueue authorization check run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkAuthorizationCheckRunFailed(r.Context(), run.ID, "authorization check run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "authorization check run could not be queued")
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

func (a *App) listAuthorizationCheckRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for authorization check runs failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListAuthorizationCheckRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list authorization check runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_authorization_check_runs_failed", "authorization check runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authorization_check_runs": runs})
}

func (a *App) getAuthorizationCheckRun(w http.ResponseWriter, r *http.Request, runID string) {
	detail, err := a.store.GetAuthorizationCheckDetail(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "authorization_check_run_not_found", "authorization check run was not found")
			return
		}
		a.logger.Error("get authorization check run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_authorization_check_run_failed", "authorization check run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (a *App) getAuthorizationCheckReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetAuthorizationCheckReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "authorization_check_run_not_found", "authorization check run was not found")
			return
		}
		a.logger.Error("get authorization check report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_authorization_check_report_failed", "authorization check report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getAuthorizationCheckHTMLReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetAuthorizationCheckReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "authorization_check_run_not_found", "authorization check run was not found")
			return
		}
		a.logger.Error("get authorization check html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_authorization_check_report_failed", "authorization check report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderAuthorizationCheckHTMLReport(w, report); err != nil {
		a.logger.Error("render authorization check html report failed", "error", err)
	}
}

func decodeAuthorizationCheckRequest(w http.ResponseWriter, r *http.Request) (AuthorizationCheckRequest, error) {
	var input AuthorizationCheckRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, fmt.Errorf("request body must be valid authorization check JSON")
	}
	return input, nil
}

func decodeAuthorizationCheckRunRequest(w http.ResponseWriter, r *http.Request) (AuthorizationCheckRunRequest, error) {
	var input AuthorizationCheckRunRequest
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
		return input, fmt.Errorf("request body must be valid authorization check run JSON")
	}
	return input, nil
}

func (a *App) validateAuthorizationCheckProfiles(ctx context.Context, projectID string, input AuthorizationCheckRequest) error {
	profile, err := a.store.GetCredentialProfile(ctx, input.ActorCredentialProfileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("actor credential profile was not found")
		}
		return fmt.Errorf("actor credential profile could not be loaded")
	}
	if profile.ProjectID != projectID {
		return fmt.Errorf("actor credential profile does not belong to the project")
	}
	if input.OwnerCredentialProfileID != "" {
		owner, err := a.store.GetCredentialProfile(ctx, input.OwnerCredentialProfileID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("owner credential profile was not found")
			}
			return fmt.Errorf("owner credential profile could not be loaded")
		}
		if owner.ProjectID != projectID {
			return fmt.Errorf("owner credential profile does not belong to the project")
		}
	}
	return nil
}

func authorizationReportGeneratedAt(report *AuthorizationCheckReport) string {
	if report == nil || report.GeneratedAt.IsZero() {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return report.GeneratedAt.Format(time.RFC3339)
}

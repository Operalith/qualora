package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *App) createQualityCheckRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for quality check run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	if project.FrontendURL == "" {
		writeError(w, http.StatusBadRequest, "frontend_url_required", "project must have frontend_url to start quality checks")
		return
	}

	input, err := decodeQualityCheckRunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_quality_check_run", err.Error())
		return
	}
	if err := a.resolveQualityCheckReferences(r, *project, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_quality_check_run", err.Error())
		return
	}
	normalized, err := NormalizeQualityCheckRunRequest(*project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_quality_check_run", err.Error())
		return
	}

	run, err := a.store.CreateQualityCheckRun(r.Context(), project.ID, normalized)
	if err != nil {
		a.logger.Error("create quality check run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_quality_check_run_failed", "quality check run could not be created")
		return
	}
	if err := a.queue.EnqueueQualityCheckRun(r.Context(), QualityCheckRunJob{QualityCheckRunID: run.ID, ProjectID: project.ID}); err != nil {
		a.logger.Error("enqueue quality check run failed", "run_id", run.ID, "error", err)
		_ = a.store.MarkQualityCheckRunFailed(r.Context(), run.ID, "quality check run could not be queued")
		writeError(w, http.StatusServiceUnavailable, "queue_unavailable", "quality check run could not be queued")
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

func (a *App) listQualityCheckRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for quality check runs failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListQualityCheckRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list quality check runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_quality_check_runs_failed", "quality check runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"quality_check_runs": runs})
}

func (a *App) handleQualityCheckRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/quality-check-runs/")
	parts := stringsSplitPath(path)
	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getQualityCheckRun(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getQualityCheckReport(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "report.html" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		a.getQualityCheckHTMLReport(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getQualityCheckRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := a.store.GetQualityCheckRun(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "quality_check_run_not_found", "quality check run was not found")
			return
		}
		a.logger.Error("get quality check run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_quality_check_run_failed", "quality check run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) getQualityCheckReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetQualityCheckReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "quality_check_run_not_found", "quality check run was not found")
			return
		}
		a.logger.Error("get quality check report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_quality_check_report_failed", "quality check report could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) getQualityCheckHTMLReport(w http.ResponseWriter, r *http.Request, runID string) {
	report, err := a.store.GetQualityCheckReport(r.Context(), runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "quality_check_run_not_found", "quality check run was not found")
			return
		}
		a.logger.Error("get quality check html report failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_quality_check_report_failed", "quality check report could not be loaded")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := RenderQualityCheckHTMLReport(w, report, time.Now().UTC()); err != nil {
		a.logger.Error("render quality check html report failed", "error", err)
	}
}

func (a *App) resolveQualityCheckReferences(r *http.Request, project Project, input *QualityCheckRunRequest) error {
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
			return fmt.Errorf("discovery run must be completed before quality checks can reuse it")
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

func decodeQualityCheckRunRequest(w http.ResponseWriter, r *http.Request) (QualityCheckRunRequest, error) {
	var input QualityCheckRunRequest
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
		return input, fmt.Errorf("request body must be valid quality check run JSON")
	}
	return input, nil
}

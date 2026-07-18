package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *App) createReportBaseline(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for report baseline failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeReportBaselineRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report_baseline", err.Error())
		return
	}
	input, err = NormalizeReportBaselineRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report_baseline", err.Error())
		return
	}
	snapshot, err := a.store.GetReportSnapshot(r.Context(), input.ReportType, input.ReportID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "report_not_found", "report was not found")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_report", err.Error())
		return
	}
	if snapshot.ProjectID != projectID {
		writeError(w, http.StatusBadRequest, "report_project_mismatch", "report does not belong to the project")
		return
	}

	createdByUserID := ""
	if user, ok := authenticatedUser(r); ok {
		createdByUserID = user.ID
	}
	baseline := BuildReportBaselineFromSnapshot(input, snapshot, createdByUserID, time.Now().UTC())
	created, err := a.store.CreateReportBaseline(r.Context(), baseline)
	if err != nil {
		a.logger.Error("create report baseline failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_report_baseline_failed", "report baseline could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *App) listReportBaselines(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for report baselines failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	reportType := r.URL.Query().Get("report_type")
	if reportType != "" {
		normalized, err := NormalizeReportType(reportType)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_report_type", err.Error())
			return
		}
		reportType = normalized
	}
	baselines, err := a.store.ListReportBaselines(r.Context(), projectID, reportType)
	if err != nil {
		a.logger.Error("list report baselines failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_report_baselines_failed", "report baselines could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"report_baselines": baselines})
}

func (a *App) createReportComparison(w http.ResponseWriter, r *http.Request, projectID string) {
	comparison, err := a.buildReportComparisonFromRequest(w, r, projectID)
	if err != nil {
		return
	}
	writeJSON(w, http.StatusOK, comparison)
}

func (a *App) evaluateQualityGate(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for quality gate failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeQualityGateEvaluationRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_quality_gate", err.Error())
		return
	}
	if r.URL.Query().Get("format") != "" {
		input.Format = r.URL.Query().Get("format")
	}
	input, err = NormalizeQualityGateEvaluationRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_quality_gate", err.Error())
		return
	}
	snapshot, err := a.store.GetReportSnapshot(r.Context(), input.ReportType, input.CurrentReportID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			result := MissingBaselineQualityGateResult(input.GateConfig, time.Now().UTC())
			result.FailedRules = append(result.FailedRules, "missing_report")
			result.Recommendation = "The current report could not be loaded; rerun the Qualora check before evaluating the gate."
			a.writeQualityGateResponse(w, projectID, input, result, ReportSnapshot{})
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_report", err.Error())
		return
	}
	if snapshot.ProjectID != projectID {
		writeError(w, http.StatusBadRequest, "report_project_mismatch", "report does not belong to the project")
		return
	}

	baseline, err := a.baselineForComparison(r, projectID, input.ReportType, input.BaselineID, input.UseDefaultBaseline || input.BaselineID == "")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			result := MissingBaselineQualityGateResult(input.GateConfig, time.Now().UTC())
			a.writeQualityGateResponse(w, projectID, input, result, snapshot)
			return
		}
		a.logger.Error("load baseline for quality gate failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_report_baseline_failed", "report baseline could not be loaded")
		return
	}
	if baseline.ProjectID != projectID || baseline.ReportType != input.ReportType {
		writeError(w, http.StatusBadRequest, "baseline_mismatch", "baseline does not match the project and report type")
		return
	}
	comparison := CompareReportToBaseline(snapshot, *baseline, time.Now().UTC())
	result := EvaluateQualityGate(comparison, snapshot.Intelligence.SeverityCounts, snapshot.Status, input.GateConfig, time.Now().UTC())
	a.writeQualityGateResponse(w, projectID, input, result, snapshot)
}

func (a *App) writeQualityGateResponse(w http.ResponseWriter, projectID string, input QualityGateEvaluationRequest, result QualityGateResult, snapshot ReportSnapshot) {
	if input.Format == "ci" {
		response := CIQualityGateResponse(projectID, input, result)
		if snapshot.ReportID != "" {
			response.ReportURL = reportURLForSnapshot(snapshot)
		}
		writeJSON(w, http.StatusOK, response)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) buildReportComparisonFromRequest(w http.ResponseWriter, r *http.Request, projectID string) (*ReportComparison, error) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return nil, err
		}
		a.logger.Error("get project for report comparison failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return nil, err
	}
	input, err := decodeReportComparisonRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report_comparison", err.Error())
		return nil, err
	}
	input, err = NormalizeReportComparisonRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report_comparison", err.Error())
		return nil, err
	}
	snapshot, err := a.store.GetReportSnapshot(r.Context(), input.ReportType, input.CurrentReportID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "report_not_found", "report was not found")
			return nil, err
		}
		writeError(w, http.StatusBadRequest, "invalid_report", err.Error())
		return nil, err
	}
	if snapshot.ProjectID != projectID {
		writeError(w, http.StatusBadRequest, "report_project_mismatch", "report does not belong to the project")
		return nil, fmt.Errorf("report project mismatch")
	}
	baseline, err := a.baselineForComparison(r, projectID, input.ReportType, input.BaselineID, input.UseDefaultBaseline || input.BaselineID == "")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "baseline_not_found", "baseline was not found")
			return nil, err
		}
		a.logger.Error("load baseline for report comparison failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_report_baseline_failed", "report baseline could not be loaded")
		return nil, err
	}
	if baseline.ProjectID != projectID || baseline.ReportType != input.ReportType {
		writeError(w, http.StatusBadRequest, "baseline_mismatch", "baseline does not match the project and report type")
		return nil, fmt.Errorf("baseline mismatch")
	}
	comparison := CompareReportToBaseline(snapshot, *baseline, time.Now().UTC())
	return &comparison, nil
}

func (a *App) baselineForComparison(r *http.Request, projectID string, reportType string, baselineID string, useDefault bool) (*ReportBaseline, error) {
	if baselineID != "" {
		return a.store.GetReportBaseline(r.Context(), baselineID)
	}
	if useDefault {
		return a.store.GetDefaultReportBaseline(r.Context(), projectID, reportType)
	}
	return nil, ErrNotFound
}

func (a *App) handleReportBaselineSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/report-baselines/")
	parts := stringsSplitPath(path)
	if len(parts) != 1 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.getReportBaseline(w, r, parts[0])
	case http.MethodPut:
		a.updateReportBaseline(w, r, parts[0])
	case http.MethodDelete:
		a.deleteReportBaseline(w, r, parts[0])
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (a *App) getReportBaseline(w http.ResponseWriter, r *http.Request, baselineID string) {
	baseline, err := a.store.GetReportBaseline(r.Context(), baselineID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "baseline_not_found", "baseline was not found")
			return
		}
		a.logger.Error("get report baseline failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_report_baseline_failed", "report baseline could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, baseline)
}

func (a *App) updateReportBaseline(w http.ResponseWriter, r *http.Request, baselineID string) {
	input, err := decodeReportBaselineUpdateRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report_baseline", err.Error())
		return
	}
	input, err = NormalizeReportBaselineUpdateRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report_baseline", err.Error())
		return
	}
	baseline, err := a.store.UpdateReportBaseline(r.Context(), baselineID, input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "baseline_not_found", "baseline was not found")
			return
		}
		a.logger.Error("update report baseline failed", "error", err)
		writeError(w, http.StatusInternalServerError, "update_report_baseline_failed", "report baseline could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, baseline)
}

func (a *App) deleteReportBaseline(w http.ResponseWriter, r *http.Request, baselineID string) {
	if err := a.store.DeleteReportBaseline(r.Context(), baselineID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "baseline_not_found", "baseline was not found")
			return
		}
		a.logger.Error("delete report baseline failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_report_baseline_failed", "report baseline could not be deleted")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func decodeReportBaselineRequest(w http.ResponseWriter, r *http.Request) (ReportBaselineRequest, error) {
	var input ReportBaselineRequest
	if err := decodeJSONBody(w, r, &input, 1<<20); err != nil {
		return input, fmt.Errorf("request body must be valid report baseline JSON")
	}
	return input, nil
}

func decodeReportBaselineUpdateRequest(w http.ResponseWriter, r *http.Request) (ReportBaselineUpdateRequest, error) {
	var input ReportBaselineUpdateRequest
	if err := decodeJSONBody(w, r, &input, 1<<20); err != nil {
		return input, fmt.Errorf("request body must be valid report baseline update JSON")
	}
	return input, nil
}

func decodeReportComparisonRequest(w http.ResponseWriter, r *http.Request) (ReportComparisonRequest, error) {
	var input ReportComparisonRequest
	if err := decodeJSONBody(w, r, &input, 1<<20); err != nil {
		return input, fmt.Errorf("request body must be valid report comparison JSON")
	}
	return input, nil
}

func decodeQualityGateEvaluationRequest(w http.ResponseWriter, r *http.Request) (QualityGateEvaluationRequest, error) {
	var input QualityGateEvaluationRequest
	if err := decodeJSONBody(w, r, &input, 1<<20); err != nil {
		return input, fmt.Errorf("request body must be valid quality gate JSON")
	}
	return input, nil
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dest any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if r.Body == http.NoBody {
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

func reportURLForSnapshot(snapshot ReportSnapshot) string {
	switch snapshot.ReportType {
	case ReportTypeSafeQA:
		return "/api/v1/qa-runs/" + snapshot.ReportID + "/report"
	case ReportTypeQualityCheck:
		return "/api/v1/quality-check-runs/" + snapshot.ReportID + "/report"
	case ReportTypeDiscovery:
		return "/api/v1/discovery-runs/" + snapshot.ReportID + "/report"
	case ReportTypeSafeExplorer:
		return "/api/v1/safe-explorer-runs/" + snapshot.ReportID + "/report"
	case ReportTypeAuthorization:
		return "/api/v1/authorization-check-runs/" + snapshot.ReportID + "/report"
	default:
		return "/api/v1/runs/" + snapshot.ReportID + "/report"
	}
}

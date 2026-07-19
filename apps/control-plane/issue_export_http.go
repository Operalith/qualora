package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (a *App) createIssueExportConfig(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for issue export config failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeIssueExportConfigRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_issue_export_config", err.Error())
		return
	}
	input, err = NormalizeIssueExportConfigRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_issue_export_config", err.Error())
		return
	}
	config, err := a.issueExportConfigFromRequest(input, projectID, "")
	if err != nil {
		a.logger.Error("encrypt issue export token failed", "error", err)
		writeError(w, http.StatusInternalServerError, "encrypt_issue_token_failed", "issue export token could not be encrypted")
		return
	}
	created, err := a.store.CreateIssueExportConfig(r.Context(), config)
	if err != nil {
		a.logger.Error("create issue export config failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_issue_export_config_failed", "issue export config could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *App) listIssueExportConfigs(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for issue export configs failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	configs, err := a.store.ListIssueExportConfigs(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list issue export configs failed", "error", err)
		writeError(w, http.StatusInternalServerError, "list_issue_export_configs_failed", "issue export configs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"issue_export_configs": configs})
}

func (a *App) handleIssueExportConfigSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/issue-export-configs/")
	parts := stringsSplitPath(path)
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodGet:
			a.getIssueExportConfig(w, r, parts[0])
		case http.MethodPut:
			a.updateIssueExportConfig(w, r, parts[0])
		case http.MethodDelete:
			a.deleteIssueExportConfig(w, r, parts[0])
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
		a.testIssueExportConfig(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "route not found")
}

func (a *App) getIssueExportConfig(w http.ResponseWriter, r *http.Request, configID string) {
	config, err := a.store.GetIssueExportConfig(r.Context(), configID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "issue_export_config_not_found", "issue export config was not found")
			return
		}
		a.logger.Error("get issue export config failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_issue_export_config_failed", "issue export config could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (a *App) updateIssueExportConfig(w http.ResponseWriter, r *http.Request, configID string) {
	existing, err := a.store.GetIssueExportConfig(r.Context(), configID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "issue_export_config_not_found", "issue export config was not found")
			return
		}
		a.logger.Error("get issue export config for update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_issue_export_config_failed", "issue export config could not be loaded")
		return
	}
	input, err := decodeIssueExportConfigUpdateRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_issue_export_config", err.Error())
		return
	}
	input, err = NormalizeIssueExportConfigUpdateRequest(*existing, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_issue_export_config", err.Error())
		return
	}
	encrypted := existing.TokenEncrypted
	if input.Token != "" {
		encrypted, err = a.secretBox.Encrypt(input.Token)
		if err != nil {
			a.logger.Error("encrypt issue export token failed", "error", err)
			writeError(w, http.StatusInternalServerError, "encrypt_issue_token_failed", "issue export token could not be encrypted")
			return
		}
	}
	updatedInput := issueExportConfigFromUpdate(*existing, input, encrypted)
	updated, err := a.store.UpdateIssueExportConfig(r.Context(), configID, updatedInput)
	if err != nil {
		a.logger.Error("update issue export config failed", "error", err)
		writeError(w, http.StatusInternalServerError, "update_issue_export_config_failed", "issue export config could not be updated")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *App) deleteIssueExportConfig(w http.ResponseWriter, r *http.Request, configID string) {
	if err := a.store.DeleteIssueExportConfig(r.Context(), configID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "issue_export_config_not_found", "issue export config was not found")
			return
		}
		a.logger.Error("delete issue export config failed", "error", err)
		writeError(w, http.StatusInternalServerError, "delete_issue_export_config_failed", "issue export config could not be deleted")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *App) testIssueExportConfig(w http.ResponseWriter, r *http.Request, configID string) {
	config, err := a.store.GetIssueExportConfig(r.Context(), configID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "issue_export_config_not_found", "issue export config was not found")
			return
		}
		a.logger.Error("get issue export config for test failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_issue_export_config_failed", "issue export config could not be loaded")
		return
	}
	result := IssueExportConfigTestResult{
		Success:  config.Enabled && config.TokenConfigured,
		Provider: config.Provider,
		Target:   fmt.Sprintf("%s/%s", config.OwnerOrNamespace, config.RepositoryOrProject),
	}
	if !config.Enabled {
		result.ErrorMessage = "issue export config is disabled"
	} else if !config.TokenConfigured {
		result.ErrorMessage = "issue export config has no token"
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) handleReportIssueExportSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/reports/")
	parts := stringsSplitPath(path)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] != "export-issues" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	a.exportReportIssues(w, r, parts[0], parts[1])
}

func (a *App) exportReportIssues(w http.ResponseWriter, r *http.Request, reportType string, reportID string) {
	normalizedType, err := NormalizeReportType(reportType)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report_type", err.Error())
		return
	}
	snapshot, err := a.store.GetReportSnapshot(r.Context(), normalizedType, reportID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "report_not_found", "report was not found")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_report", err.Error())
		return
	}
	input, err := decodeIssueExportRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_issue_export", err.Error())
		return
	}
	result, err := a.exportIssuesForSnapshot(r.Context(), normalizedType, reportID, snapshot, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "issue_export_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func decodeIssueExportConfigRequest(w http.ResponseWriter, r *http.Request) (IssueExportConfigRequest, error) {
	var input IssueExportConfigRequest
	if err := decodeJSONBody(w, r, &input, 1<<20); err != nil {
		return input, fmt.Errorf("request body must be valid issue export config JSON")
	}
	return input, nil
}

func decodeIssueExportConfigUpdateRequest(w http.ResponseWriter, r *http.Request) (IssueExportConfigUpdateRequest, error) {
	var input IssueExportConfigUpdateRequest
	if err := decodeJSONBody(w, r, &input, 1<<20); err != nil {
		return input, fmt.Errorf("request body must be valid issue export config JSON")
	}
	return input, nil
}

func decodeIssueExportRequest(w http.ResponseWriter, r *http.Request) (IssueExportRequest, error) {
	var input IssueExportRequest
	if err := decodeJSONBody(w, r, &input, 1<<20); err != nil {
		return input, fmt.Errorf("request body must be valid issue export JSON")
	}
	return input, nil
}

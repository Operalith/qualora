package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(Bearer|Basic)\s+[A-Za-z0-9._~+/=-]+`),
	regexp.MustCompile(`(?i)(authorization|password|passwd|token|secret|api[_-]?key|access[_-]?token|refresh[_-]?token|session[_-]?id|cookie)=([^&\s",}]+)`),
	regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`),
}

var urlPattern = regexp.MustCompile(`https?://[^\s"'<>)]+`)

type AIAnalysisPayload struct {
	ExecutiveSummary   string   `json:"executive_summary"`
	TechnicalSummary   string   `json:"technical_summary"`
	RiskLevel          string   `json:"risk_level"`
	LikelyCauses       []string `json:"likely_causes"`
	RecommendedActions []string `json:"recommended_actions"`
	SuggestedNextTests []string `json:"suggested_next_tests"`
	Confidence         float64  `json:"confidence"`
	Limitations        []string `json:"limitations"`
}

func AIAnalysisSystemPrompt() string {
	return strings.Join([]string{
		"You are Qualora's AI QA analyst.",
		"Analyze only the provided browser/API run data.",
		"Do not invent evidence.",
		"Do not claim something was tested unless it appears in the input.",
		"Do not expose secrets.",
		"Return strict JSON only.",
		"If the data is insufficient, say so in limitations.",
		`Use this JSON shape exactly: {"executive_summary":"string","technical_summary":"string","risk_level":"low|medium|high|critical","likely_causes":["string"],"recommended_actions":["string"],"suggested_next_tests":["string"],"confidence":0.0,"limitations":["string"]}.`,
	}, " ")
}

func BuildAIUserPrompt(report *Report) (string, error) {
	input := BuildSafeAIInput(report)
	raw, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal safe AI input: %w", err)
	}
	return "Analyze this Qualora run report. Use only this sanitized JSON input:\n" + string(raw), nil
}

func BuildSafeAIInput(report *Report) map[string]any {
	input := map[string]any{
		"run_id":     report.RunID,
		"project_id": report.ProjectID,
		"run_type":   report.RunType,
		"status":     report.Status,
		"summary":    report.Summary,
		"metadata":   safeMetadata(report.Metadata),
		"findings":   safeFindings(report.Findings),
		"evidence":   safeEvidence(report.Evidence),
	}
	if report.APISummary != nil {
		input["api_summary"] = report.APISummary
	}
	if len(report.APIResults) > 0 {
		input["api_results"] = safeAPIResults(report.APIResults)
	}
	return sanitizeValue(input).(map[string]any)
}

func BuildSafeAuthorizationAIInput(report *AuthorizationCheckReport) map[string]any {
	input := map[string]any{
		"run_id":     report.Run.ID,
		"project_id": report.Project.ID,
		"status":     report.Run.Status,
		"summary":    report.Summary,
		"metadata": map[string]any{
			"authorization_checks":  len(report.Checks),
			"authorization_results": len(report.Results),
			"safe_methods_only":     true,
			"destructive_actions":   false,
		},
		"results":              safeAuthorizationResults(report.Results),
		"findings":             safeFindings(report.Findings),
		"evidence":             safeEvidence(report.Evidence),
		"authorization_checks": safeAuthorizationChecks(report.Checks),
	}
	return sanitizeValue(input).(map[string]any)
}

func BuildSafeDiscoveryAIInput(report *DiscoveryReport) map[string]any {
	input := map[string]any{
		"run_id":     report.Run.ID,
		"project_id": report.Project.ID,
		"status":     report.Run.Status,
		"settings": map[string]any{
			"max_pages":        report.Run.MaxPages,
			"max_depth":        report.Run.MaxDepth,
			"same_origin_only": report.Run.SameOriginOnly,
			"safe_links_only":  true,
			"forms_submitted":  false,
		},
		"summary":      report.Summary,
		"pages":        safeDiscoveryPages(report.Pages),
		"links":        safeDiscoveryLinks(report.Links),
		"forms":        safeDiscoveryForms(report.Forms),
		"findings":     safeFindings(report.Findings),
		"evidence":     safeEvidence(report.Evidence),
		"safety_notes": report.SafetyNotes,
		"limitations":  report.Limitations,
	}
	return sanitizeValue(input).(map[string]any)
}

func BuildSafeQualityAIInput(report *QualityCheckReport) map[string]any {
	input := map[string]any{
		"run_id":     report.Run.ID,
		"project_id": report.Project.ID,
		"status":     report.Run.Status,
		"settings": map[string]any{
			"target_url":              report.Run.TargetURL,
			"max_pages":               report.Run.MaxPages,
			"include_security":        report.Run.IncludeSecurity,
			"include_accessibility":   report.Run.IncludeAccessibility,
			"include_performance":     report.Run.IncludePerformance,
			"uses_discovery_run":      report.Run.DiscoveryRunID != "",
			"uses_credential_profile": report.Run.CredentialProfileID != "",
			"forms_submitted":         false,
			"destructive_actions":     false,
			"active_scanning":         false,
		},
		"summary":      report.Summary,
		"results":      safeQualityResults(report.Results),
		"safety_notes": report.SafetyNotes,
		"limitations":  report.Limitations,
		"metadata":     safeMetadata(report.Metadata),
	}
	return sanitizeValue(input).(map[string]any)
}

func ParseAIAnalysisPayload(raw string) (*AIAnalysisPayload, map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}
	var payload AIAnalysisPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, nil, fmt.Errorf("AI analysis was not valid JSON: %w", err)
	}
	payload.ExecutiveSummary = strings.TrimSpace(RedactSecrets(payload.ExecutiveSummary))
	payload.TechnicalSummary = strings.TrimSpace(RedactSecrets(payload.TechnicalSummary))
	payload.RiskLevel = strings.ToLower(strings.TrimSpace(payload.RiskLevel))
	if !validRiskLevel(payload.RiskLevel) {
		return nil, nil, fmt.Errorf("AI analysis risk_level must be low, medium, high, or critical")
	}
	payload.LikelyCauses = sanitizeStringSlice(payload.LikelyCauses, 12)
	payload.RecommendedActions = sanitizeStringSlice(payload.RecommendedActions, 12)
	payload.SuggestedNextTests = sanitizeStringSlice(payload.SuggestedNextTests, 12)
	payload.Limitations = sanitizeStringSlice(payload.Limitations, 12)
	if payload.Confidence < 0 {
		payload.Confidence = 0
	}
	if payload.Confidence > 1 {
		payload.Confidence = 1
	}
	rawJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal parsed AI analysis: %w", err)
	}
	var analysisJSON map[string]any
	if err := json.Unmarshal(rawJSON, &analysisJSON); err != nil {
		return nil, nil, fmt.Errorf("normalize parsed AI analysis: %w", err)
	}
	return &payload, analysisJSON, nil
}

func RedactSecrets(input string) string {
	output := input
	for _, pattern := range secretPatterns {
		output = pattern.ReplaceAllStringFunc(output, func(match string) string {
			if strings.Contains(match, "=") {
				parts := strings.SplitN(match, "=", 2)
				return parts[0] + "=[REDACTED]"
			}
			fields := strings.Fields(match)
			if len(fields) > 1 {
				return fields[0] + " [REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return output
}

func safeMetadata(metadata map[string]any) map[string]any {
	output := make(map[string]any)
	for _, key := range []string{"page_title", "created_at", "error_message"} {
		if value, ok := metadata[key]; ok {
			output[key] = sanitizeValue(value)
		}
	}
	if jobs, ok := metadata["jobs"]; ok {
		output["jobs"] = sanitizeValue(jobs)
	}
	if apiSummary, ok := metadata["api_summary"]; ok {
		output["api_summary"] = sanitizeValue(apiSummary)
	}
	if login, ok := metadata["login"]; ok {
		output["login"] = sanitizeValue(login)
	}
	if credentialProfile, ok := metadata["credential_profile"]; ok {
		output["credential_profile"] = sanitizeValue(credentialProfile)
	}
	return output
}

func safeFindings(findings []Finding) []map[string]any {
	output := make([]map[string]any, 0, len(findings))
	for _, finding := range findings {
		output = append(output, map[string]any{
			"title":      finding.Title,
			"severity":   finding.Severity,
			"category":   finding.Category,
			"confidence": finding.Confidence,
			"summary":    firstLine(finding.Description),
		})
	}
	return output
}

func safeAPIResults(results []APICheckResult) []map[string]any {
	output := make([]map[string]any, 0, min(len(results), 100))
	for i, result := range results {
		if i >= 100 {
			break
		}
		output = append(output, map[string]any{
			"method":                result.Method,
			"path":                  result.Path,
			"status":                result.Status,
			"http_status":           result.HTTPStatus,
			"duration_ms":           result.DurationMS,
			"response_content_type": result.ResponseContentType,
			"response_size_bytes":   result.ResponseSizeBytes,
			"error":                 firstLine(result.ErrorMessage),
			"skipped_reason":        result.SkippedReason,
		})
	}
	return output
}

func safeQualityResults(results []QualityCheckResult) []map[string]any {
	output := make([]map[string]any, 0, min(len(results), 200))
	for i, result := range results {
		if i >= 200 {
			break
		}
		output = append(output, map[string]any{
			"category":       result.Category,
			"rule_id":        result.RuleID,
			"severity":       result.Severity,
			"title":          result.Title,
			"summary":        firstLine(result.Description),
			"recommendation": result.Recommendation,
			"url":            result.URL,
			"evidence":       sanitizeValue(result.Evidence),
		})
	}
	return output
}

func safeAuthorizationChecks(checks []AuthorizationCheck) []map[string]any {
	output := make([]map[string]any, 0, min(len(checks), 100))
	for i, check := range checks {
		if i >= 100 {
			break
		}
		output = append(output, map[string]any{
			"name":             check.Name,
			"type":             check.Type,
			"resource_label":   check.ResourceLabel,
			"expected_outcome": check.ExpectedOutcome,
			"target_url":       check.TargetURL,
			"enabled":          check.Enabled,
		})
	}
	return output
}

func safeAuthorizationResults(results []AuthorizationCheckResult) []map[string]any {
	output := make([]map[string]any, 0, min(len(results), 100))
	for i, result := range results {
		if i >= 100 {
			break
		}
		output = append(output, map[string]any{
			"status":         result.Status,
			"expected":       result.ExpectedOutcome,
			"actual":         result.ActualOutcome,
			"actor_role":     result.ActorRoleName,
			"target_url":     result.TargetURL,
			"final_url":      result.FinalURL,
			"http_status":    result.HTTPStatus,
			"page_title":     result.PageTitle,
			"duration_ms":    result.DurationMS,
			"skip_reason":    result.SkipReason,
			"error_message":  firstLine(result.ErrorMessage),
			"credential_ref": result.ActorCredentialProfileID,
		})
	}
	return output
}

func safeDiscoveryPages(pages []DiscoveredPage) []map[string]any {
	output := make([]map[string]any, 0, min(len(pages), 100))
	for i, page := range pages {
		if i >= 100 {
			break
		}
		output = append(output, map[string]any{
			"path":                 page.Path,
			"normalized_url":       page.NormalizedURL,
			"title":                page.Title,
			"http_status":          page.HTTPStatus,
			"content_type":         page.ContentType,
			"body_text_length":     page.BodyTextLength,
			"depth":                page.Depth,
			"console_error_count":  page.ConsoleErrorCount,
			"failed_request_count": page.FailedRequestCount,
			"has_screenshot":       page.ScreenshotEvidenceID != "",
		})
	}
	return output
}

func safeDiscoveryLinks(links []DiscoveredLink) []map[string]any {
	output := make([]map[string]any, 0, min(len(links), 200))
	for i, link := range links {
		if i >= 200 {
			break
		}
		output = append(output, map[string]any{
			"normalized_url": link.NormalizedURL,
			"link_text":      link.LinkText,
			"same_origin":    link.SameOrigin,
			"skipped":        link.Skipped,
			"skip_reason":    link.SkipReason,
		})
	}
	return output
}

func safeDiscoveryForms(forms []DiscoveredForm) []map[string]any {
	output := make([]map[string]any, 0, min(len(forms), 100))
	for i, form := range forms {
		if i >= 100 {
			break
		}
		fields := make([]map[string]any, 0, min(len(form.Fields), 50))
		for index, field := range form.Fields {
			if index >= 50 {
				break
			}
			fields = append(fields, map[string]any{
				"field_name": field.FieldName,
				"field_type": field.FieldType,
				"label":      field.Label,
				"required":   field.Required,
			})
		}
		output = append(output, map[string]any{
			"form_method":          form.FormMethod,
			"form_action":          form.FormAction,
			"field_count":          form.FieldCount,
			"password_field_count": form.PasswordFieldCount,
			"submit_button_count":  form.SubmitButtonCount,
			"classification":       form.Classification,
			"skipped_reason":       form.SkippedReason,
			"fields":               fields,
		})
	}
	return output
}

func safeEvidence(records []Evidence) []map[string]any {
	output := make([]map[string]any, 0, len(records))
	for _, record := range records {
		output = append(output, map[string]any{
			"type":     record.Type,
			"metadata": safeEvidenceMetadata(record.Metadata),
		})
	}
	return output
}

func safeEvidenceMetadata(metadata map[string]any) map[string]any {
	output := make(map[string]any)
	for _, key := range []string{
		"target_url", "final_url", "page_title", "status_code", "body_text_length", "timed_out",
		"load_error", "content_type", "size_bytes", "filename", "key", "storage",
		"page_id", "normalized_url", "load_duration_ms", "console_error_count",
		"failed_request_count", "blocked_request_count",
		"checked_endpoints", "failed_endpoints", "safe_methods_only", "version", "paths",
		"operations", "safe_operations", "skipped_unsafe_operations", "skipped_endpoints",
		"api_spec_id", "api_spec_name", "title", "server_url", "authenticated_tests",
		"response_bodies", "request_response_bodies_saved",
		"credential_profile_name", "login_status", "login_url", "final_url", "duration_ms",
		"success", "failure_reason", "authenticated_target_url",
		"authorization_check_id", "authorization_check_run_id", "check_name", "check_type",
		"resource_label", "actor_credential_profile_name", "actor_role_name", "expected_outcome",
		"actual_outcome", "result_status", "login_final_url", "login_duration_ms",
		"success_text_configured", "denied_text_configured", "destructive_actions",
		"autonomous_ai_browser_control",
	} {
		if value, ok := metadata[key]; ok {
			output[key] = sanitizeValue(value)
		}
	}
	if value, ok := metadata["console_errors"]; ok {
		output["console_errors"] = sanitizeValue(value)
	}
	if value, ok := metadata["failed_requests"]; ok {
		output["failed_requests"] = sanitizeValue(value)
	}
	if value, ok := metadata["blocked_requests"]; ok {
		output["blocked_requests"] = sanitizeValue(value)
	}
	if value, ok := metadata["observations"]; ok {
		output["observations"] = sanitizeValue(value)
	}
	return output
}

func sanitizeValue(value any) any {
	switch typed := value.(type) {
	case string:
		return sanitizeText(RedactSecrets(limitString(typed, 1000)))
	case []string:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, sanitizeValue(item))
		}
		return out
	case []any:
		out := make([]any, 0, min(len(typed), 50))
		for i, item := range typed {
			if i >= 50 {
				break
			}
			out = append(out, sanitizeValue(item))
		}
		return out
	case []map[string]any:
		out := make([]any, 0, min(len(typed), 50))
		for i, item := range typed {
			if i >= 50 {
				break
			}
			out = append(out, sanitizeValue(item))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			if sensitiveKey(key) {
				out[key] = "[REDACTED]"
				continue
			}
			out[key] = sanitizeValue(item)
		}
		return out
	default:
		return typed
	}
}

func sanitizeText(value string) string {
	value = sanitizePotentialURL(value)
	return urlPattern.ReplaceAllStringFunc(value, sanitizePotentialURL)
}

func sanitizePotentialURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return value
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func sensitiveKey(key string) bool {
	key = strings.ToLower(key)
	for _, part := range []string{"authorization", "password", "passwd", "username", "token", "secret", "api_key", "apikey", "cookie", "session", "local_storage", "session_storage", "browser_storage", "body", "html"} {
		if strings.Contains(key, part) {
			return true
		}
	}
	return false
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	line := strings.Split(value, "\n")[0]
	return limitString(line, 500)
}

func limitString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "...[truncated]"
}

func sanitizeStringSlice(values []string, limit int) []string {
	output := make([]string, 0, min(len(values), limit))
	for i, value := range values {
		if i >= limit {
			break
		}
		value = strings.TrimSpace(sanitizeText(RedactSecrets(value)))
		if value != "" {
			output = append(output, limitString(value, 1000))
		}
	}
	return output
}

func validRiskLevel(value string) bool {
	switch value {
	case "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}

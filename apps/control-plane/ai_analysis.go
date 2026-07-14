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
		"status":     report.Status,
		"summary":    report.Summary,
		"metadata":   safeMetadata(report.Metadata),
		"findings":   safeFindings(report.Findings),
		"evidence":   safeEvidence(report.Evidence),
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
		"checked_endpoints", "failed_endpoints", "safe_methods_only", "version", "paths",
		"operations", "safe_operations", "skipped_unsafe_operations",
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
	for _, part := range []string{"authorization", "password", "passwd", "token", "secret", "api_key", "apikey", "cookie", "session", "body", "html"} {
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
		value = strings.TrimSpace(RedactSecrets(value))
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

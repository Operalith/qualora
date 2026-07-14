package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListAIProviders(ctx context.Context) ([]AIProvider, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, name, preset, type, base_url, model, api_key_encrypted, extra_headers_encrypted,
	temperature, max_output_tokens, timeout_seconds, send_screenshots, send_html,
	send_network_bodies, redaction_enabled, is_default, created_at, updated_at
FROM ai_providers
ORDER BY is_default DESC, created_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("query AI providers: %w", err)
	}
	defer rows.Close()

	providers := make([]AIProvider, 0)
	for rows.Next() {
		provider, err := scanAIProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI providers: %w", err)
	}
	return providers, nil
}

func (s *Store) GetAIProvider(ctx context.Context, id string) (*AIProvider, error) {
	provider, err := scanAIProvider(s.db.QueryRow(ctx, `
SELECT id, name, preset, type, base_url, model, api_key_encrypted, extra_headers_encrypted,
	temperature, max_output_tokens, timeout_seconds, send_screenshots, send_html,
	send_network_bodies, redaction_enabled, is_default, created_at, updated_at
FROM ai_providers
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &provider, nil
}

func (s *Store) GetDefaultAIProvider(ctx context.Context) (*AIProvider, error) {
	provider, err := scanAIProvider(s.db.QueryRow(ctx, `
SELECT id, name, preset, type, base_url, model, api_key_encrypted, extra_headers_encrypted,
	temperature, max_output_tokens, timeout_seconds, send_screenshots, send_html,
	send_network_bodies, redaction_enabled, is_default, created_at, updated_at
FROM ai_providers
ORDER BY is_default DESC, created_at DESC
LIMIT 1
`))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &provider, nil
}

func (s *Store) CreateAIProvider(ctx context.Context, provider AIProvider) (*AIProvider, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create AI provider: %w", err)
	}
	defer tx.Rollback(ctx)

	if provider.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE ai_providers SET is_default = false, updated_at = now() WHERE is_default = true`); err != nil {
			return nil, fmt.Errorf("clear default AI provider: %w", err)
		}
	}

	created, err := scanAIProvider(tx.QueryRow(ctx, `
INSERT INTO ai_providers (
	id, name, preset, type, base_url, model, api_key_encrypted, extra_headers_encrypted,
	temperature, max_output_tokens, timeout_seconds, send_screenshots, send_html,
	send_network_bodies, redaction_enabled, is_default
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING id, name, preset, type, base_url, model, api_key_encrypted, extra_headers_encrypted,
	temperature, max_output_tokens, timeout_seconds, send_screenshots, send_html,
	send_network_bodies, redaction_enabled, is_default, created_at, updated_at
`,
		uuid.NewString(),
		provider.Name,
		provider.Preset,
		provider.Type,
		provider.BaseURL,
		provider.Model,
		provider.APIKeyEncrypted,
		provider.ExtraHeadersEncrypted,
		provider.Temperature,
		provider.MaxOutputTokens,
		provider.TimeoutSeconds,
		provider.SendScreenshots,
		provider.SendHTML,
		provider.SendNetworkBodies,
		provider.RedactionEnabled,
		provider.IsDefault,
	))
	if err != nil {
		return nil, fmt.Errorf("insert AI provider: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create AI provider: %w", err)
	}
	return &created, nil
}

func (s *Store) UpdateAIProvider(ctx context.Context, id string, provider AIProvider) (*AIProvider, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update AI provider: %w", err)
	}
	defer tx.Rollback(ctx)

	if provider.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE ai_providers SET is_default = false, updated_at = now() WHERE id <> $1 AND is_default = true`, id); err != nil {
			return nil, fmt.Errorf("clear default AI provider: %w", err)
		}
	}

	updated, err := scanAIProvider(tx.QueryRow(ctx, `
UPDATE ai_providers
SET name = $2,
	preset = $3,
	type = $4,
	base_url = $5,
	model = $6,
	api_key_encrypted = $7,
	extra_headers_encrypted = $8,
	temperature = $9,
	max_output_tokens = $10,
	timeout_seconds = $11,
	send_screenshots = $12,
	send_html = $13,
	send_network_bodies = $14,
	redaction_enabled = $15,
	is_default = $16,
	updated_at = now()
WHERE id = $1
RETURNING id, name, preset, type, base_url, model, api_key_encrypted, extra_headers_encrypted,
	temperature, max_output_tokens, timeout_seconds, send_screenshots, send_html,
	send_network_bodies, redaction_enabled, is_default, created_at, updated_at
`,
		id,
		provider.Name,
		provider.Preset,
		provider.Type,
		provider.BaseURL,
		provider.Model,
		provider.APIKeyEncrypted,
		provider.ExtraHeadersEncrypted,
		provider.Temperature,
		provider.MaxOutputTokens,
		provider.TimeoutSeconds,
		provider.SendScreenshots,
		provider.SendHTML,
		provider.SendNetworkBodies,
		provider.RedactionEnabled,
		provider.IsDefault,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update AI provider: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update AI provider: %w", err)
	}
	return &updated, nil
}

func (s *Store) DeleteAIProvider(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM ai_providers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete AI provider: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateAIAnalysis(ctx context.Context, runID string, providerID string, model string) (*AIAnalysis, error) {
	analysis, err := scanAIAnalysis(s.db.QueryRow(ctx, `
INSERT INTO ai_analyses (id, run_id, provider_id, model, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, run_id, provider_id::text, '', model, status, executive_summary, technical_summary,
	risk_level, analysis_json, prompt_tokens, completion_tokens, total_tokens, error_message, created_at, updated_at
`, uuid.NewString(), runID, providerID, model, StatusRunning))
	if err != nil {
		return nil, fmt.Errorf("insert AI analysis: %w", err)
	}
	return &analysis, nil
}

func (s *Store) CompleteAIAnalysis(ctx context.Context, id string, payload *AIAnalysisPayload, analysisJSON map[string]any, usage AIClientResponse) (*AIAnalysis, error) {
	rawJSON, err := json.Marshal(analysisJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal AI analysis json: %w", err)
	}
	analysis, err := scanAIAnalysis(s.db.QueryRow(ctx, `
UPDATE ai_analyses
SET status = $2,
	executive_summary = $3,
	technical_summary = $4,
	risk_level = $5,
	analysis_json = $6,
	prompt_tokens = $7,
	completion_tokens = $8,
	total_tokens = $9,
	error_message = '',
	updated_at = now()
WHERE id = $1
RETURNING id, run_id, provider_id::text, '', model, status, executive_summary, technical_summary,
	risk_level, analysis_json, prompt_tokens, completion_tokens, total_tokens, error_message, created_at, updated_at
`, id, StatusCompleted, payload.ExecutiveSummary, payload.TechnicalSummary, payload.RiskLevel, rawJSON, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens))
	if err != nil {
		return nil, fmt.Errorf("complete AI analysis: %w", err)
	}
	return &analysis, nil
}

func (s *Store) FailAIAnalysis(ctx context.Context, id string, message string) (*AIAnalysis, error) {
	analysis, err := scanAIAnalysis(s.db.QueryRow(ctx, `
UPDATE ai_analyses
SET status = $2, error_message = $3, updated_at = now()
WHERE id = $1
RETURNING id, run_id, provider_id::text, '', model, status, executive_summary, technical_summary,
	risk_level, analysis_json, prompt_tokens, completion_tokens, total_tokens, error_message, created_at, updated_at
`, id, StatusFailed, message))
	if err != nil {
		return nil, fmt.Errorf("fail AI analysis: %w", err)
	}
	return &analysis, nil
}

func (s *Store) GetLatestAIAnalysis(ctx context.Context, runID string) (*AIAnalysis, error) {
	analysis, err := scanAIAnalysis(s.db.QueryRow(ctx, `
SELECT a.id, a.run_id, a.provider_id::text, COALESCE(p.name, ''), a.model, a.status,
	a.executive_summary, a.technical_summary, a.risk_level, a.analysis_json,
	a.prompt_tokens, a.completion_tokens, a.total_tokens, a.error_message, a.created_at, a.updated_at
FROM ai_analyses a
LEFT JOIN ai_providers p ON p.id = a.provider_id
WHERE a.run_id = $1
ORDER BY a.created_at DESC
LIMIT 1
`, runID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &analysis, nil
}

func scanAIProvider(row scanRow) (AIProvider, error) {
	var provider AIProvider
	if err := row.Scan(
		&provider.ID,
		&provider.Name,
		&provider.Preset,
		&provider.Type,
		&provider.BaseURL,
		&provider.Model,
		&provider.APIKeyEncrypted,
		&provider.ExtraHeadersEncrypted,
		&provider.Temperature,
		&provider.MaxOutputTokens,
		&provider.TimeoutSeconds,
		&provider.SendScreenshots,
		&provider.SendHTML,
		&provider.SendNetworkBodies,
		&provider.RedactionEnabled,
		&provider.IsDefault,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	); err != nil {
		return AIProvider{}, fmt.Errorf("scan AI provider: %w", err)
	}
	provider.APIKeyConfigured = provider.APIKeyEncrypted != ""
	provider.ExtraHeadersConfigured = provider.ExtraHeadersEncrypted != ""
	return provider, nil
}

func scanAIAnalysis(row scanRow) (AIAnalysis, error) {
	var analysis AIAnalysis
	var providerID sql.NullString
	var analysisRaw []byte
	if err := row.Scan(
		&analysis.ID,
		&analysis.RunID,
		&providerID,
		&analysis.ProviderName,
		&analysis.Model,
		&analysis.Status,
		&analysis.ExecutiveSummary,
		&analysis.TechnicalSummary,
		&analysis.RiskLevel,
		&analysisRaw,
		&analysis.PromptTokens,
		&analysis.CompletionTokens,
		&analysis.TotalTokens,
		&analysis.ErrorMessage,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	); err != nil {
		return AIAnalysis{}, fmt.Errorf("scan AI analysis: %w", err)
	}
	if providerID.Valid {
		analysis.ProviderID = providerID.String
	}
	if len(analysisRaw) == 0 {
		analysis.AnalysisJSON = map[string]any{}
	} else if err := json.Unmarshal(analysisRaw, &analysis.AnalysisJSON); err != nil {
		return AIAnalysis{}, fmt.Errorf("unmarshal AI analysis json: %w", err)
	}
	return analysis, nil
}

func decodeExtraHeaders(raw string) (map[string]string, error) {
	if raw == "" {
		return map[string]string{}, nil
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(raw), &headers); err != nil {
		return nil, fmt.Errorf("parse extra headers: %w", err)
	}
	return headers, nil
}

func encodeExtraHeaders(headers map[string]string) (string, error) {
	if len(headers) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(headers)
	if err != nil {
		return "", fmt.Errorf("marshal extra headers: %w", err)
	}
	return string(raw), nil
}

func providerPresetDefaults(preset string) AIProviderRequest {
	switch preset {
	case "openai":
		return AIProviderRequest{Preset: "openai", Type: AIProviderOpenAICompatible, BaseURL: "https://api.openai.com/v1", Model: "gpt-4o-mini", Temperature: 0.2, MaxOutputTokens: 1200, TimeoutSeconds: 30}
	case "openrouter":
		return AIProviderRequest{Preset: "openrouter", Type: AIProviderOpenAICompatible, BaseURL: "https://openrouter.ai/api/v1", Model: "openai/gpt-4o-mini", ExtraHeaders: map[string]string{"X-OpenRouter-Title": "Qualora"}, Temperature: 0.2, MaxOutputTokens: 1200, TimeoutSeconds: 30}
	case "ollama":
		return AIProviderRequest{Preset: "ollama", Type: AIProviderOpenAICompatible, BaseURL: "http://ollama:11434/v1", Model: "qwen2.5-coder:7b", Temperature: 0.2, MaxOutputTokens: 1200, TimeoutSeconds: 60}
	default:
		return AIProviderRequest{Preset: "custom", Type: AIProviderOpenAICompatible, Temperature: 0.2, MaxOutputTokens: 1200, TimeoutSeconds: 30}
	}
}

func normalizeProviderRequest(input AIProviderRequest) (AIProviderRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Preset = strings.TrimSpace(input.Preset)
	input.Type = strings.TrimSpace(input.Type)
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.Model = strings.TrimSpace(input.Model)
	if input.Preset != "" && !validProviderPreset(input.Preset) {
		return input, fmt.Errorf("preset must be openai, openrouter, ollama, or custom")
	}
	defaults := providerPresetDefaults(input.Preset)
	if input.Preset == "" {
		input.Preset = defaults.Preset
	}
	if input.Type == "" {
		input.Type = defaults.Type
	}
	if input.BaseURL == "" {
		input.BaseURL = defaults.BaseURL
	}
	if input.Model == "" {
		input.Model = defaults.Model
	}
	if input.Temperature == 0 {
		input.Temperature = defaults.Temperature
	}
	if input.MaxOutputTokens == 0 {
		input.MaxOutputTokens = defaults.MaxOutputTokens
	}
	if input.TimeoutSeconds == 0 {
		input.TimeoutSeconds = defaults.TimeoutSeconds
	}
	if input.ExtraHeaders == nil && len(defaults.ExtraHeaders) > 0 {
		input.ExtraHeaders = defaults.ExtraHeaders
	}
	if input.RedactionEnabled == nil {
		enabled := true
		input.RedactionEnabled = &enabled
	}
	if input.SendScreenshots == nil {
		disabled := false
		input.SendScreenshots = &disabled
	}
	if input.SendHTML == nil {
		disabled := false
		input.SendHTML = &disabled
	}
	if input.SendNetworkBodies == nil {
		disabled := false
		input.SendNetworkBodies = &disabled
	}
	if input.Name == "" {
		input.Name = defaults.Preset
	}
	if input.Type != AIProviderOpenAICompatible {
		return input, fmt.Errorf("type must be openai-compatible")
	}
	if input.BaseURL == "" {
		return input, fmt.Errorf("base_url is required")
	}
	if _, err := chatCompletionsURL(input.BaseURL); err != nil {
		return input, err
	}
	if input.Model == "" {
		return input, fmt.Errorf("model is required")
	}
	if input.Temperature < 0 || input.Temperature > 2 {
		return input, fmt.Errorf("temperature must be between 0 and 2")
	}
	if input.MaxOutputTokens < 1 || input.MaxOutputTokens > 10000 {
		return input, fmt.Errorf("max_output_tokens must be between 1 and 10000")
	}
	if input.TimeoutSeconds < 1 || input.TimeoutSeconds > 180 {
		return input, fmt.Errorf("timeout_seconds must be between 1 and 180")
	}
	return input, nil
}

func validProviderPreset(preset string) bool {
	switch preset {
	case "openai", "openrouter", "ollama", "custom":
		return true
	default:
		return false
	}
}

func aiProviderFromRequest(input AIProviderRequest, encryptedAPIKey string, encryptedExtraHeaders string) AIProvider {
	return AIProvider{
		Name:                  input.Name,
		Preset:                input.Preset,
		Type:                  input.Type,
		BaseURL:               input.BaseURL,
		Model:                 input.Model,
		APIKeyEncrypted:       encryptedAPIKey,
		ExtraHeadersEncrypted: encryptedExtraHeaders,
		Temperature:           input.Temperature,
		MaxOutputTokens:       input.MaxOutputTokens,
		TimeoutSeconds:        input.TimeoutSeconds,
		SendScreenshots:       *input.SendScreenshots,
		SendHTML:              *input.SendHTML,
		SendNetworkBodies:     *input.SendNetworkBodies,
		RedactionEnabled:      *input.RedactionEnabled,
		IsDefault:             input.IsDefault,
	}
}

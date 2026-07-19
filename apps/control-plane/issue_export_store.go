package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateIssueExportConfig(ctx context.Context, config IssueExportConfig) (*IssueExportConfig, error) {
	labels, err := json.Marshal(config.DefaultLabels)
	if err != nil {
		return nil, fmt.Errorf("marshal issue export labels: %w", err)
	}
	created, err := scanIssueExportConfig(s.db.QueryRow(ctx, `
INSERT INTO issue_export_configs (
	id, project_id, provider, name, base_url, owner_or_namespace,
	repository_or_project, token_encrypted, default_labels_json, enabled
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, project_id, provider, name, base_url, owner_or_namespace,
	repository_or_project, token_encrypted, default_labels_json, enabled, created_at, updated_at
`, uuid.NewString(), config.ProjectID, config.Provider, config.Name, config.BaseURL, config.OwnerOrNamespace, config.RepositoryOrProject, config.TokenEncrypted, labels, config.Enabled))
	if err != nil {
		return nil, fmt.Errorf("insert issue export config: %w", err)
	}
	return &created, nil
}

func (s *Store) ListIssueExportConfigs(ctx context.Context, projectID string) ([]IssueExportConfig, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, provider, name, base_url, owner_or_namespace,
	repository_or_project, token_encrypted, default_labels_json, enabled, created_at, updated_at
FROM issue_export_configs
WHERE project_id = $1
ORDER BY enabled DESC, created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query issue export configs: %w", err)
	}
	defer rows.Close()
	configs := []IssueExportConfig{}
	for rows.Next() {
		config, err := scanIssueExportConfig(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue export configs: %w", err)
	}
	return configs, nil
}

func (s *Store) GetIssueExportConfig(ctx context.Context, id string) (*IssueExportConfig, error) {
	config, err := scanIssueExportConfig(s.db.QueryRow(ctx, `
SELECT id, project_id, provider, name, base_url, owner_or_namespace,
	repository_or_project, token_encrypted, default_labels_json, enabled, created_at, updated_at
FROM issue_export_configs
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &config, nil
}

func (s *Store) UpdateIssueExportConfig(ctx context.Context, id string, config IssueExportConfig) (*IssueExportConfig, error) {
	labels, err := json.Marshal(config.DefaultLabels)
	if err != nil {
		return nil, fmt.Errorf("marshal issue export labels: %w", err)
	}
	updated, err := scanIssueExportConfig(s.db.QueryRow(ctx, `
UPDATE issue_export_configs
SET provider = $2,
	name = $3,
	base_url = $4,
	owner_or_namespace = $5,
	repository_or_project = $6,
	token_encrypted = $7,
	default_labels_json = $8,
	enabled = $9,
	updated_at = now()
WHERE id = $1
RETURNING id, project_id, provider, name, base_url, owner_or_namespace,
	repository_or_project, token_encrypted, default_labels_json, enabled, created_at, updated_at
`, id, config.Provider, config.Name, config.BaseURL, config.OwnerOrNamespace, config.RepositoryOrProject, config.TokenEncrypted, labels, config.Enabled))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update issue export config: %w", err)
	}
	return &updated, nil
}

func (s *Store) DeleteIssueExportConfig(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM issue_export_configs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete issue export config: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanIssueExportConfig(row scanRow) (IssueExportConfig, error) {
	var config IssueExportConfig
	var labelsRaw []byte
	if err := row.Scan(
		&config.ID,
		&config.ProjectID,
		&config.Provider,
		&config.Name,
		&config.BaseURL,
		&config.OwnerOrNamespace,
		&config.RepositoryOrProject,
		&config.TokenEncrypted,
		&labelsRaw,
		&config.Enabled,
		&config.CreatedAt,
		&config.UpdatedAt,
	); err != nil {
		return IssueExportConfig{}, fmt.Errorf("scan issue export config: %w", err)
	}
	config.DefaultLabels = []string{}
	if len(labelsRaw) > 0 {
		if err := json.Unmarshal(labelsRaw, &config.DefaultLabels); err != nil {
			return IssueExportConfig{}, fmt.Errorf("unmarshal issue export labels: %w", err)
		}
	}
	config.TokenConfigured = config.TokenEncrypted != ""
	return config, nil
}

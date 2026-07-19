package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListAPIAuthProfiles(ctx context.Context, projectID string) ([]APIAuthProfile, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, name, type, header_name, query_param_name,
	username_encrypted, password_encrypted, token_encrypted, api_key_encrypted,
	username_display_hint, token_display_hint, api_key_display_hint, enabled, created_at, updated_at
FROM api_auth_profiles
WHERE project_id = $1
ORDER BY enabled DESC, created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query API auth profiles: %w", err)
	}
	defer rows.Close()

	profiles := make([]APIAuthProfile, 0)
	for rows.Next() {
		profile, err := scanAPIAuthProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate API auth profiles: %w", err)
	}
	return profiles, nil
}

func (s *Store) GetAPIAuthProfile(ctx context.Context, id string) (*APIAuthProfile, error) {
	profile, err := scanAPIAuthProfile(s.db.QueryRow(ctx, `
SELECT id, project_id, name, type, header_name, query_param_name,
	username_encrypted, password_encrypted, token_encrypted, api_key_encrypted,
	username_display_hint, token_display_hint, api_key_display_hint, enabled, created_at, updated_at
FROM api_auth_profiles
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (s *Store) CreateAPIAuthProfile(ctx context.Context, projectID string, profile APIAuthProfile) (*APIAuthProfile, error) {
	created, err := scanAPIAuthProfile(s.db.QueryRow(ctx, `
INSERT INTO api_auth_profiles (
	id, project_id, name, type, header_name, query_param_name,
	username_encrypted, password_encrypted, token_encrypted, api_key_encrypted,
	username_display_hint, token_display_hint, api_key_display_hint, enabled
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, project_id, name, type, header_name, query_param_name,
	username_encrypted, password_encrypted, token_encrypted, api_key_encrypted,
	username_display_hint, token_display_hint, api_key_display_hint, enabled, created_at, updated_at
`, uuid.NewString(), projectID, profile.Name, profile.Type, profile.HeaderName, profile.QueryParamName,
		profile.UsernameEncrypted, profile.PasswordEncrypted, profile.TokenEncrypted, profile.APIKeyEncrypted,
		profile.UsernameDisplayHint, profile.TokenDisplayHint, profile.APIKeyDisplayHint, profile.Enabled))
	if err != nil {
		return nil, fmt.Errorf("insert API auth profile: %w", err)
	}
	return &created, nil
}

func (s *Store) UpdateAPIAuthProfile(ctx context.Context, id string, profile APIAuthProfile) (*APIAuthProfile, error) {
	updated, err := scanAPIAuthProfile(s.db.QueryRow(ctx, `
UPDATE api_auth_profiles
SET name = $2,
	type = $3,
	header_name = $4,
	query_param_name = $5,
	username_encrypted = $6,
	password_encrypted = $7,
	token_encrypted = $8,
	api_key_encrypted = $9,
	username_display_hint = $10,
	token_display_hint = $11,
	api_key_display_hint = $12,
	enabled = $13,
	updated_at = now()
WHERE id = $1
RETURNING id, project_id, name, type, header_name, query_param_name,
	username_encrypted, password_encrypted, token_encrypted, api_key_encrypted,
	username_display_hint, token_display_hint, api_key_display_hint, enabled, created_at, updated_at
`, id, profile.Name, profile.Type, profile.HeaderName, profile.QueryParamName,
		profile.UsernameEncrypted, profile.PasswordEncrypted, profile.TokenEncrypted, profile.APIKeyEncrypted,
		profile.UsernameDisplayHint, profile.TokenDisplayHint, profile.APIKeyDisplayHint, profile.Enabled))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update API auth profile: %w", err)
	}
	return &updated, nil
}

func (s *Store) DeleteAPIAuthProfile(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM api_auth_profiles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete API auth profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanAPIAuthProfile(row scanRow) (APIAuthProfile, error) {
	var profile APIAuthProfile
	if err := row.Scan(
		&profile.ID,
		&profile.ProjectID,
		&profile.Name,
		&profile.Type,
		&profile.HeaderName,
		&profile.QueryParamName,
		&profile.UsernameEncrypted,
		&profile.PasswordEncrypted,
		&profile.TokenEncrypted,
		&profile.APIKeyEncrypted,
		&profile.UsernameDisplayHint,
		&profile.TokenDisplayHint,
		&profile.APIKeyDisplayHint,
		&profile.Enabled,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return APIAuthProfile{}, fmt.Errorf("scan API auth profile: %w", err)
	}
	profile.setConfiguredFlags()
	return profile, nil
}

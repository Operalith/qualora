package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const CredentialProfileTypeUsernamePassword = "username_password"

func normalizeCredentialProfileRequest(input CredentialProfileRequest, project Project, create bool) (CredentialProfileRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return input, fmt.Errorf("name is required")
	}
	if len(input.Name) > 120 {
		return input, fmt.Errorf("name must be 120 characters or fewer")
	}

	input.Type = strings.TrimSpace(input.Type)
	if input.Type == "" {
		input.Type = CredentialProfileTypeUsernamePassword
	}
	if input.Type != CredentialProfileTypeUsernamePassword {
		return input, fmt.Errorf("type must be username_password")
	}

	input.Username = strings.TrimSpace(input.Username)
	if create && input.Username == "" {
		return input, fmt.Errorf("username is required")
	}
	if create && input.Password == "" {
		return input, fmt.Errorf("password is required")
	}

	input.LoginURL = strings.TrimSpace(input.LoginURL)
	loginURL, err := ValidateTargetURL(input.LoginURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return input, fmt.Errorf("login_url: %w", err)
	}
	if project.FrontendURL != "" {
		frontendURL, err := url.Parse(project.FrontendURL)
		if err != nil {
			return input, fmt.Errorf("project frontend_url is invalid")
		}
		if loginURL.Scheme != frontendURL.Scheme || loginURL.Host != frontendURL.Host {
			return input, fmt.Errorf("login_url must use the same origin as frontend_url")
		}
	}
	input.LoginURL = loginURL.String()

	var errSelector error
	input.UsernameSelector, errSelector = normalizeLoginSelector(input.UsernameSelector, "username_selector")
	if errSelector != nil {
		return input, errSelector
	}
	input.PasswordSelector, errSelector = normalizeLoginSelector(input.PasswordSelector, "password_selector")
	if errSelector != nil {
		return input, errSelector
	}
	input.SubmitSelector, errSelector = normalizeLoginSelector(input.SubmitSelector, "submit_selector")
	if errSelector != nil {
		return input, errSelector
	}

	input.SuccessURLContains = strings.TrimSpace(input.SuccessURLContains)
	input.SuccessTextContains = strings.TrimSpace(input.SuccessTextContains)
	input.FailureTextContains = strings.TrimSpace(input.FailureTextContains)
	if input.SuccessURLContains == "" && input.SuccessTextContains == "" {
		return input, fmt.Errorf("success_url_contains or success_text_contains is required")
	}
	if input.PostLoginWaitMS < 0 || input.PostLoginWaitMS > 30000 {
		return input, fmt.Errorf("post_login_wait_ms must be between 0 and 30000")
	}
	return input, nil
}

func normalizeLoginSelector(selector string, field string) (string, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if len(selector) > 500 {
		return "", fmt.Errorf("%s must be 500 characters or fewer", field)
	}
	if strings.ContainsRune(selector, '\x00') {
		return "", fmt.Errorf("%s is invalid", field)
	}
	return selector, nil
}

func credentialProfileFromRequest(input CredentialProfileRequest, encryptedUsername string, encryptedPassword string, usernameHint string) CredentialProfile {
	return CredentialProfile{
		Name:                input.Name,
		Type:                input.Type,
		UsernameEncrypted:   encryptedUsername,
		PasswordEncrypted:   encryptedPassword,
		UsernameDisplayHint: usernameHint,
		LoginURL:            input.LoginURL,
		UsernameSelector:    input.UsernameSelector,
		PasswordSelector:    input.PasswordSelector,
		SubmitSelector:      input.SubmitSelector,
		SuccessURLContains:  input.SuccessURLContains,
		SuccessTextContains: input.SuccessTextContains,
		FailureTextContains: input.FailureTextContains,
		PostLoginWaitMS:     input.PostLoginWaitMS,
		IsDefault:           input.IsDefault,
	}
}

func usernameDisplayHint(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}
	if at := strings.Index(username, "@"); at > 0 {
		return username[:1] + "***" + username[at:]
	}
	if len(username) <= 2 {
		return username[:1] + "***"
	}
	return username[:1] + "***" + username[len(username)-1:]
}

func (s *Store) ListCredentialProfiles(ctx context.Context, projectID string) ([]CredentialProfile, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, name, type, username_encrypted, password_encrypted, username_display_hint,
	login_url, username_selector, password_selector, submit_selector, success_url_contains,
	success_text_contains, failure_text_contains, post_login_wait_ms, is_default, created_at, updated_at
FROM credential_profiles
WHERE project_id = $1
ORDER BY is_default DESC, created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query credential profiles: %w", err)
	}
	defer rows.Close()

	profiles := make([]CredentialProfile, 0)
	for rows.Next() {
		profile, err := scanCredentialProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credential profiles: %w", err)
	}
	return profiles, nil
}

func (s *Store) GetCredentialProfile(ctx context.Context, id string) (*CredentialProfile, error) {
	profile, err := scanCredentialProfile(s.db.QueryRow(ctx, `
SELECT id, project_id, name, type, username_encrypted, password_encrypted, username_display_hint,
	login_url, username_selector, password_selector, submit_selector, success_url_contains,
	success_text_contains, failure_text_contains, post_login_wait_ms, is_default, created_at, updated_at
FROM credential_profiles
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

func (s *Store) GetDefaultCredentialProfile(ctx context.Context, projectID string) (*CredentialProfile, error) {
	profile, err := scanCredentialProfile(s.db.QueryRow(ctx, `
SELECT id, project_id, name, type, username_encrypted, password_encrypted, username_display_hint,
	login_url, username_selector, password_selector, submit_selector, success_url_contains,
	success_text_contains, failure_text_contains, post_login_wait_ms, is_default, created_at, updated_at
FROM credential_profiles
WHERE project_id = $1
ORDER BY is_default DESC, created_at DESC
LIMIT 1
`, projectID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (s *Store) CreateCredentialProfile(ctx context.Context, projectID string, profile CredentialProfile) (*CredentialProfile, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create credential profile: %w", err)
	}
	defer tx.Rollback(ctx)

	if !profile.IsDefault {
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM credential_profiles WHERE project_id = $1`, projectID).Scan(&count); err != nil {
			return nil, fmt.Errorf("count credential profiles: %w", err)
		}
		profile.IsDefault = count == 0
	}
	if profile.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE credential_profiles SET is_default = false, updated_at = now() WHERE project_id = $1`, projectID); err != nil {
			return nil, fmt.Errorf("clear default credential profile: %w", err)
		}
	}

	created, err := scanCredentialProfile(tx.QueryRow(ctx, `
INSERT INTO credential_profiles (
	id, project_id, name, type, username_encrypted, password_encrypted, username_display_hint,
	login_url, username_selector, password_selector, submit_selector, success_url_contains,
	success_text_contains, failure_text_contains, post_login_wait_ms, is_default
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING id, project_id, name, type, username_encrypted, password_encrypted, username_display_hint,
	login_url, username_selector, password_selector, submit_selector, success_url_contains,
	success_text_contains, failure_text_contains, post_login_wait_ms, is_default, created_at, updated_at
`, uuid.NewString(), projectID, profile.Name, profile.Type, profile.UsernameEncrypted, profile.PasswordEncrypted, profile.UsernameDisplayHint,
		profile.LoginURL, profile.UsernameSelector, profile.PasswordSelector, profile.SubmitSelector, profile.SuccessURLContains,
		profile.SuccessTextContains, profile.FailureTextContains, profile.PostLoginWaitMS, profile.IsDefault))
	if err != nil {
		return nil, fmt.Errorf("insert credential profile: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create credential profile: %w", err)
	}
	return &created, nil
}

func (s *Store) UpdateCredentialProfile(ctx context.Context, id string, profile CredentialProfile) (*CredentialProfile, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update credential profile: %w", err)
	}
	defer tx.Rollback(ctx)

	if profile.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE credential_profiles SET is_default = false, updated_at = now() WHERE project_id = $1 AND id <> $2`, profile.ProjectID, id); err != nil {
			return nil, fmt.Errorf("clear default credential profile: %w", err)
		}
	}

	updated, err := scanCredentialProfile(tx.QueryRow(ctx, `
UPDATE credential_profiles
SET name = $2,
	type = $3,
	username_encrypted = $4,
	password_encrypted = $5,
	username_display_hint = $6,
	login_url = $7,
	username_selector = $8,
	password_selector = $9,
	submit_selector = $10,
	success_url_contains = $11,
	success_text_contains = $12,
	failure_text_contains = $13,
	post_login_wait_ms = $14,
	is_default = $15,
	updated_at = now()
WHERE id = $1
RETURNING id, project_id, name, type, username_encrypted, password_encrypted, username_display_hint,
	login_url, username_selector, password_selector, submit_selector, success_url_contains,
	success_text_contains, failure_text_contains, post_login_wait_ms, is_default, created_at, updated_at
`, id, profile.Name, profile.Type, profile.UsernameEncrypted, profile.PasswordEncrypted, profile.UsernameDisplayHint,
		profile.LoginURL, profile.UsernameSelector, profile.PasswordSelector, profile.SubmitSelector, profile.SuccessURLContains,
		profile.SuccessTextContains, profile.FailureTextContains, profile.PostLoginWaitMS, profile.IsDefault))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update credential profile: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update credential profile: %w", err)
	}
	return &updated, nil
}

func (s *Store) DeleteCredentialProfile(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM credential_profiles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete credential profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanCredentialProfile(row scanRow) (CredentialProfile, error) {
	var profile CredentialProfile
	if err := row.Scan(
		&profile.ID,
		&profile.ProjectID,
		&profile.Name,
		&profile.Type,
		&profile.UsernameEncrypted,
		&profile.PasswordEncrypted,
		&profile.UsernameDisplayHint,
		&profile.LoginURL,
		&profile.UsernameSelector,
		&profile.PasswordSelector,
		&profile.SubmitSelector,
		&profile.SuccessURLContains,
		&profile.SuccessTextContains,
		&profile.FailureTextContains,
		&profile.PostLoginWaitMS,
		&profile.IsDefault,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return CredentialProfile{}, fmt.Errorf("scan credential profile: %w", err)
	}
	profile.UsernameConfigured = profile.UsernameEncrypted != ""
	profile.PasswordConfigured = profile.PasswordEncrypted != ""
	return profile, nil
}

func summarizeLoginEvidence(run *TestRun, evidence []Evidence) *LoginSummary {
	for _, record := range evidence {
		if record.Type != "login_observations" {
			continue
		}
		summary := &LoginSummary{
			CredentialProfileID:    run.CredentialProfileID,
			CredentialProfileName:  metadataStringValue(record.Metadata, "credential_profile_name"),
			LoginStatus:            metadataStringValue(record.Metadata, "login_status"),
			LoginURL:               metadataStringValue(record.Metadata, "login_url"),
			LoginFinalURL:          metadataStringValue(record.Metadata, "final_url"),
			PageTitle:              metadataStringValue(record.Metadata, "page_title"),
			AuthenticatedTargetURL: metadataStringValue(record.Metadata, "authenticated_target_url"),
			FailureReason:          metadataStringValue(record.Metadata, "failure_reason"),
		}
		if value := metadataIntValue(record.Metadata, "duration_ms"); value > 0 {
			summary.LoginDurationMS = value
		}
		return summary
	}
	return nil
}

func metadataStringValue(metadata map[string]any, key string) string {
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		raw, _ := json.Marshal(typed)
		return string(raw)
	}
}

func metadataIntValue(metadata map[string]any, key string) int {
	value, ok := metadata[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		n, _ := typed.Int64()
		return int(n)
	default:
		return 0
	}
}

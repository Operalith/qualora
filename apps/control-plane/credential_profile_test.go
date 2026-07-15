package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeCredentialProfileRequestEnforcesLoginSafety(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}

	input, err := normalizeCredentialProfileRequest(CredentialProfileRequest{
		Name:                "Demo Login",
		Username:            "demo@example.com",
		Password:            "demo-password",
		LoginURL:            "http://demo-web:8080/login",
		UsernameSelector:    "#username",
		PasswordSelector:    "#password",
		SubmitSelector:      "button[type=submit]",
		SuccessURLContains:  "/dashboard",
		SuccessTextContains: "Welcome",
	}, project, true)
	if err != nil {
		t.Fatalf("normalize credential profile: %v", err)
	}
	if input.Type != CredentialProfileTypeUsernamePassword {
		t.Fatalf("expected default username_password type, got %q", input.Type)
	}

	_, err = normalizeCredentialProfileRequest(CredentialProfileRequest{
		Name:               "Bad",
		Username:           "demo@example.com",
		Password:           "demo-password",
		LoginURL:           "http://evil.example/login",
		UsernameSelector:   "#username",
		PasswordSelector:   "#password",
		SubmitSelector:     "button",
		SuccessURLContains: "/dashboard",
	}, project, true)
	if err == nil || !strings.Contains(err.Error(), "login_url") {
		t.Fatalf("expected external login_url to be rejected, got %v", err)
	}
}

func TestCredentialProfileSecretsAreEncryptedAndNotSerialized(t *testing.T) {
	box, err := NewSecretBox("test-key")
	if err != nil {
		t.Fatalf("new secret box: %v", err)
	}
	app := &App{secretBox: box}
	input := CredentialProfileRequest{
		Name:               "Demo Login",
		Type:               CredentialProfileTypeUsernamePassword,
		Username:           "demo@example.com",
		Password:           "demo-password",
		LoginURL:           "http://demo-web:8080/login",
		UsernameSelector:   "#username",
		PasswordSelector:   "#password",
		SubmitSelector:     "button[type=submit]",
		SuccessURLContains: "/dashboard",
	}
	profile, err := app.credentialProfileFromInput(input, "", "", "")
	if err != nil {
		t.Fatalf("profile from input: %v", err)
	}
	if profile.UsernameEncrypted == input.Username || profile.PasswordEncrypted == input.Password {
		t.Fatal("expected profile secrets to be encrypted")
	}
	profile.UsernameConfigured = true
	profile.PasswordConfigured = true

	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("marshal profile: %v", err)
	}
	body := string(raw)
	for _, leaked := range []string{"demo@example.com", "demo-password", profile.UsernameEncrypted, profile.PasswordEncrypted} {
		if strings.Contains(body, leaked) {
			t.Fatalf("credential profile JSON leaked secret %q in %s", leaked, body)
		}
	}
	if !strings.Contains(body, `"username_configured":true`) || !strings.Contains(body, `"password_configured":true`) {
		t.Fatalf("credential profile JSON should expose configured flags, got %s", body)
	}
}

func TestCredentialProfileUpdatePreservesExistingSecretsWhenOmitted(t *testing.T) {
	box, err := NewSecretBox("test-key")
	if err != nil {
		t.Fatalf("new secret box: %v", err)
	}
	app := &App{secretBox: box}
	currentUsername, err := box.Encrypt("demo@example.com")
	if err != nil {
		t.Fatalf("encrypt username: %v", err)
	}
	currentPassword, err := box.Encrypt("demo-password")
	if err != nil {
		t.Fatalf("encrypt password: %v", err)
	}
	profile, err := app.credentialProfileFromInput(CredentialProfileRequest{
		Name:               "Edited",
		Type:               CredentialProfileTypeUsernamePassword,
		LoginURL:           "http://demo-web:8080/login",
		UsernameSelector:   "#username",
		PasswordSelector:   "#password",
		SubmitSelector:     "button[type=submit]",
		SuccessURLContains: "/dashboard",
	}, currentUsername, currentPassword, "d***@example.com")
	if err != nil {
		t.Fatalf("profile from update input: %v", err)
	}
	if profile.UsernameEncrypted != currentUsername || profile.PasswordEncrypted != currentPassword {
		t.Fatal("expected omitted update credentials to preserve existing encrypted values")
	}
}

func TestNormalizeAuthenticatedBrowserSmokeRequest(t *testing.T) {
	input, err := normalizeAuthenticatedBrowserSmokeRequest(AuthenticatedBrowserSmokeRequest{
		TargetPath: "dashboard",
	})
	if err != nil {
		t.Fatalf("normalize authenticated smoke request: %v", err)
	}
	if input.TargetPath != "/dashboard" {
		t.Fatalf("expected relative path to be normalized, got %q", input.TargetPath)
	}
	if input.MaxDurationSeconds != 30 {
		t.Fatalf("expected default max duration, got %d", input.MaxDurationSeconds)
	}

	for _, targetPath := range []string{"https://example.com/dashboard", "//example.com/dashboard", "/dashboard?token=secret"} {
		_, err := normalizeAuthenticatedBrowserSmokeRequest(AuthenticatedBrowserSmokeRequest{
			TargetPath: targetPath,
		})
		if err == nil {
			t.Fatalf("expected target path %q to be rejected", targetPath)
		}
	}
}

func TestAuthenticatedRunSafeAIInputDoesNotIncludeCredentials(t *testing.T) {
	report := &Report{
		RunID:   "run-1",
		RunType: RunTypeAuthenticatedBrowserSmoke,
		Status:  StatusCompleted,
		Metadata: map[string]any{
			"login": map[string]any{
				"credential_profile_name": "Demo Login",
				"login_status":            "passed",
				"username":                "demo@example.com",
				"password":                "demo-password",
				"token":                   "secret-token",
			},
		},
		Evidence: []Evidence{{
			Type: "login_observations",
			Metadata: map[string]any{
				"credential_profile_name": "Demo Login",
				"login_status":            "passed",
				"final_url":               "http://demo-web:8080/dashboard?session=secret",
				"password":                "demo-password",
			},
		}},
	}
	raw, err := json.Marshal(BuildSafeAIInput(report))
	if err != nil {
		t.Fatalf("marshal safe input: %v", err)
	}
	body := string(raw)
	for _, leaked := range []string{"demo@example.com", "demo-password", "secret-token", "session=secret"} {
		if strings.Contains(body, leaked) {
			t.Fatalf("safe AI input leaked %q in %s", leaked, body)
		}
	}
	if !strings.Contains(body, "Demo Login") || !strings.Contains(body, "passed") {
		t.Fatalf("safe AI input should keep non-secret login metadata, got %s", body)
	}
}

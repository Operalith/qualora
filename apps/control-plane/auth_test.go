package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	hash, err := hashPassword("qualora-admin-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if strings.Contains(hash, "qualora-admin-password") {
		t.Fatalf("password hash included plaintext password")
	}
	if !verifyPassword("qualora-admin-password", hash) {
		t.Fatalf("expected password verification to succeed")
	}
	if verifyPassword("wrong-password", hash) {
		t.Fatalf("expected wrong password verification to fail")
	}
}

func TestSetupAdminValidation(t *testing.T) {
	input, err := normalizeSetupAdminRequest(SetupAdminRequest{
		Email:           " Admin@Qualora.Local ",
		DisplayName:     " Qualora Admin ",
		Password:        "qualora-admin-password",
		ConfirmPassword: "qualora-admin-password",
	})
	if err != nil {
		t.Fatalf("normalize setup admin: %v", err)
	}
	if input.Email != "admin@qualora.local" || input.DisplayName != "Qualora Admin" {
		t.Fatalf("unexpected normalized setup input: %#v", input)
	}

	if _, err := normalizeSetupAdminRequest(SetupAdminRequest{
		Email:           "admin@qualora.local",
		DisplayName:     "Admin",
		Password:        "short",
		ConfirmPassword: "short",
	}); err == nil {
		t.Fatalf("expected weak password to be rejected")
	}
	if _, err := normalizeSetupAdminRequest(SetupAdminRequest{
		Email:           "admin@qualora.local",
		DisplayName:     "Admin",
		Password:        "qualora-admin-password",
		ConfirmPassword: "different-password",
	}); err == nil {
		t.Fatalf("expected mismatched password confirmation to be rejected")
	}
}

func TestLocalUserJSONDoesNotExposePasswordHash(t *testing.T) {
	user := LocalUser{
		ID:           "user-id",
		Email:        "admin@qualora.local",
		DisplayName:  "Admin",
		PasswordHash: "$argon2id$secret",
		Role:         "admin",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}
	if strings.Contains(string(body), "password") || strings.Contains(string(body), "$argon2id") {
		t.Fatalf("user JSON leaked password material: %s", body)
	}
}

func TestPublicAPIPathClassification(t *testing.T) {
	for _, path := range []string{"/healthz", "/api/v1/setup/status", "/api/v1/setup/admin", "/api/v1/auth/login", "/api/v1/auth/logout", "/api/v1/auth/me"} {
		if !isPublicAPIPath(path) {
			t.Fatalf("expected %s to be public", path)
		}
	}
	for _, path := range []string{"/api/v1/projects", "/api/v1/runs/run-id/report", "/api/v1/evidence/evidence-id"} {
		if isPublicAPIPath(path) {
			t.Fatalf("expected %s to be protected", path)
		}
	}
}

func TestCSRFValidationUsesCookieHeaderAndSessionHash(t *testing.T) {
	app := &App{}
	token := "csrf-token"
	request := httptest.NewRequest(http.MethodPost, "/api/v1/projects", nil)
	request.Header.Set(csrfHeaderName, token)
	request.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
	session := &UserSession{CSRFTokenHash: hashToken(token)}
	if !app.validCSRF(request, session) {
		t.Fatalf("expected CSRF validation to pass")
	}

	request.Header.Set(csrfHeaderName, "different")
	if app.validCSRF(request, session) {
		t.Fatalf("expected mismatched CSRF header to fail")
	}
}

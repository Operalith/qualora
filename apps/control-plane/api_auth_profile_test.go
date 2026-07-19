package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeAPIAuthProfileRequestRequiresSecrets(t *testing.T) {
	if _, err := normalizeAPIAuthProfileRequest(APIAuthProfileRequest{Name: "Bearer", Type: APIAuthProfileTypeBearerToken}, true); err == nil {
		t.Fatal("expected bearer token profile to require token on create")
	}
	if _, err := normalizeAPIAuthProfileRequest(APIAuthProfileRequest{Name: "Key", Type: APIAuthProfileTypeAPIKeyHeader, HeaderName: "X-API-Key"}, true); err == nil {
		t.Fatal("expected api key header profile to require api_key on create")
	}
	if _, err := normalizeAPIAuthProfileRequest(APIAuthProfileRequest{Name: "Basic", Type: APIAuthProfileTypeBasicAuth, Username: "demo"}, true); err == nil {
		t.Fatal("expected basic auth profile to require password on create")
	}
}

func TestAPIAuthProfileEncryptsAndRedactsSecrets(t *testing.T) {
	box, err := NewSecretBox("test-key")
	if err != nil {
		t.Fatal(err)
	}
	app := &App{secretBox: box}
	input, err := normalizeAPIAuthProfileRequest(APIAuthProfileRequest{
		Name:  "Demo bearer",
		Type:  APIAuthProfileTypeBearerToken,
		Token: "demo-api-token",
	}, true)
	if err != nil {
		t.Fatalf("normalize profile: %v", err)
	}
	profile, err := app.apiAuthProfileFromInput(input, nil)
	if err != nil {
		t.Fatalf("profile from input: %v", err)
	}
	if profile.TokenEncrypted == "" || strings.Contains(profile.TokenEncrypted, "demo-api-token") {
		t.Fatalf("expected encrypted token, got %q", profile.TokenEncrypted)
	}
	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "demo-api-token") || strings.Contains(string(raw), profile.TokenEncrypted) {
		t.Fatalf("profile JSON leaked token material: %s", string(raw))
	}
	if !strings.Contains(string(raw), `"token_configured":true`) {
		t.Fatalf("profile JSON should expose only configured hint, got %s", string(raw))
	}
}

func TestSanitizeURLForStorageWithSensitiveRedactsCustomQueryKey(t *testing.T) {
	got := sanitizeURLForStorageWithSensitive("https://api.example.test/private?x-key=secret-value&safe=true", []string{"x-key"})
	if strings.Contains(got, "secret-value") || !strings.Contains(got, "x-key=%5BREDACTED%5D") {
		t.Fatalf("expected custom query key redaction, got %q", got)
	}
}

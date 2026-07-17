package main

import "testing"

func TestNormalizeSafeExplorerRunRequestDefaultsAndRedactsSensitiveQuery(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	normalized, err := NormalizeSafeExplorerRunRequest(project, SafeExplorerRunRequest{
		StartURL: "http://demo-web:8080/dashboard?token=secret&page=1#section",
	})
	if err != nil {
		t.Fatalf("NormalizeSafeExplorerRunRequest() error = %v", err)
	}
	if normalized.StartURL != "http://demo-web:8080/dashboard?page=1&token=%5BREDACTED%5D" {
		t.Fatalf("expected redacted URL, got %q", normalized.StartURL)
	}
	if normalized.MaxSteps != defaultSafeExplorerMaxSteps {
		t.Fatalf("expected default max steps, got %d", normalized.MaxSteps)
	}
	if normalized.MaxDepth != defaultSafeExplorerMaxDepth {
		t.Fatalf("expected default max depth, got %d", normalized.MaxDepth)
	}
	if normalized.SameOriginOnly == nil || !*normalized.SameOriginOnly {
		t.Fatalf("expected same_origin_only default true")
	}
}

func TestNormalizeSafeExplorerRunRequestRejectsOutOfScopeStartURL(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	_, err := NormalizeSafeExplorerRunRequest(project, SafeExplorerRunRequest{
		StartURL: "http://other.local/",
	})
	if err == nil {
		t.Fatalf("expected out-of-scope URL to be rejected")
	}
}

func TestNormalizeSafeExplorerRunRequestAllowsAllowedHostWhenSameOriginDisabled(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web", "docs.local"},
		AllowPrivateTargets: true,
	}
	sameOriginOnly := false
	normalized, err := NormalizeSafeExplorerRunRequest(project, SafeExplorerRunRequest{
		StartURL:       "http://docs.local/start",
		SameOriginOnly: &sameOriginOnly,
	})
	if err != nil {
		t.Fatalf("NormalizeSafeExplorerRunRequest() error = %v", err)
	}
	if normalized.StartURL != "http://docs.local/start" {
		t.Fatalf("expected allowed cross-origin URL, got %q", normalized.StartURL)
	}
}

func TestNormalizeSafeExplorerRunRequestRejectsUnsafeLimits(t *testing.T) {
	project := Project{FrontendURL: "http://demo-web:8080/", AllowedHosts: []string{"demo-web"}, AllowPrivateTargets: true}
	if _, err := NormalizeSafeExplorerRunRequest(project, SafeExplorerRunRequest{MaxSteps: 51}); err == nil {
		t.Fatalf("expected max_steps over limit to fail")
	}
	if _, err := NormalizeSafeExplorerRunRequest(project, SafeExplorerRunRequest{MaxDepth: 6}); err == nil {
		t.Fatalf("expected max_depth over limit to fail")
	}
}

package main

import (
	"strings"
	"testing"
)

func TestNormalizeDiscoveryRunRequestDefaultsAndRedactsSensitiveQuery(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}

	normalized, err := NormalizeDiscoveryRunRequest(project, DiscoveryRunRequest{
		StartURL: "http://demo-web:8080/dashboard?token=secret-value&page=1#section",
	})
	if err != nil {
		t.Fatalf("normalize discovery request: %v", err)
	}
	if normalized.MaxPages != defaultDiscoveryMaxPages {
		t.Fatalf("expected default max pages, got %d", normalized.MaxPages)
	}
	if normalized.MaxDepth != defaultDiscoveryMaxDepth {
		t.Fatalf("expected default max depth, got %d", normalized.MaxDepth)
	}
	if normalized.SameOriginOnly == nil || !*normalized.SameOriginOnly {
		t.Fatalf("expected same_origin_only default true")
	}
	if strings.Contains(normalized.StartURL, "secret-value") || strings.Contains(normalized.StartURL, "#section") {
		t.Fatalf("expected sensitive query value and fragment to be removed, got %q", normalized.StartURL)
	}
	if !strings.Contains(normalized.StartURL, "token=%5BREDACTED%5D") || !strings.Contains(normalized.StartURL, "page=1") {
		t.Fatalf("expected redacted sensitive query and preserved safe query, got %q", normalized.StartURL)
	}
}

func TestNormalizeDiscoveryRunRequestRejectsOutOfScopeStartURL(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web", "other-web"},
		AllowPrivateTargets: true,
	}

	_, err := NormalizeDiscoveryRunRequest(project, DiscoveryRunRequest{
		StartURL: "http://other-web:8080/",
	})
	if err == nil || !strings.Contains(err.Error(), "same_origin_only") {
		t.Fatalf("expected same-origin rejection, got %v", err)
	}
}

func TestNormalizeDiscoveryRunRequestAllowsAllowedHostWhenSameOriginDisabled(t *testing.T) {
	sameOriginOnly := false
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web", "other-web"},
		AllowPrivateTargets: true,
	}

	normalized, err := NormalizeDiscoveryRunRequest(project, DiscoveryRunRequest{
		StartURL:       "http://other-web:8080/",
		MaxPages:       5,
		MaxDepth:       1,
		SameOriginOnly: &sameOriginOnly,
	})
	if err != nil {
		t.Fatalf("normalize discovery request with same-origin disabled: %v", err)
	}
	if normalized.StartURL != "http://other-web:8080/" {
		t.Fatalf("unexpected normalized start URL: %q", normalized.StartURL)
	}
}

func TestNormalizeDiscoveryRunRequestRejectsUnsafeLimits(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}

	if _, err := NormalizeDiscoveryRunRequest(project, DiscoveryRunRequest{MaxPages: 101}); err == nil {
		t.Fatalf("expected max_pages rejection")
	}
	if _, err := NormalizeDiscoveryRunRequest(project, DiscoveryRunRequest{MaxDepth: 6}); err == nil {
		t.Fatalf("expected max_depth rejection")
	}
}

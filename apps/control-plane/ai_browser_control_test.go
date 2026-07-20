package main

import "testing"

func TestNormalizeAIBrowserControlRunRequestDefaultsAndRedactsSensitiveQuery(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	normalized, err := NormalizeAIBrowserControlRunRequest(project, AIBrowserControlRunRequest{
		ProviderID: "provider-1",
		Goal:       "Explore safe public pages.",
		StartURL:   "http://demo-web:8080/dashboard?token=secret&page=1#section",
	})
	if err != nil {
		t.Fatalf("NormalizeAIBrowserControlRunRequest() error = %v", err)
	}
	if normalized.StartURL != "http://demo-web:8080/dashboard?page=1&token=%5BREDACTED%5D" {
		t.Fatalf("expected redacted URL, got %q", normalized.StartURL)
	}
	if normalized.MaxSteps != defaultAIBrowserMaxSteps {
		t.Fatalf("expected default max steps, got %d", normalized.MaxSteps)
	}
	if normalized.MaxDepth != defaultAIBrowserMaxDepth {
		t.Fatalf("expected default max depth, got %d", normalized.MaxDepth)
	}
	if normalized.SameOriginOnly == nil || !*normalized.SameOriginOnly {
		t.Fatalf("expected same_origin_only default true")
	}
}

func TestNormalizeAIBrowserControlRunRequestRequiresProvider(t *testing.T) {
	project := Project{FrontendURL: "http://demo-web:8080/", AllowedHosts: []string{"demo-web"}, AllowPrivateTargets: true}
	_, err := NormalizeAIBrowserControlRunRequest(project, AIBrowserControlRunRequest{Goal: "Explore safely."})
	if err == nil {
		t.Fatalf("expected missing provider_id to fail")
	}
}

func TestNormalizeAIBrowserControlRunRequestRejectsOutOfScopeStartURL(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	_, err := NormalizeAIBrowserControlRunRequest(project, AIBrowserControlRunRequest{
		ProviderID: "provider-1",
		Goal:       "Explore safely.",
		StartURL:   "http://other.local/",
	})
	if err == nil {
		t.Fatalf("expected out-of-scope URL to be rejected")
	}
}

func TestNormalizeAIBrowserControlRunRequestAllowsAllowedHostWhenSameOriginDisabled(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web", "docs.local"},
		AllowPrivateTargets: true,
	}
	sameOriginOnly := false
	normalized, err := NormalizeAIBrowserControlRunRequest(project, AIBrowserControlRunRequest{
		ProviderID:     "provider-1",
		Goal:           "Explore safe docs pages.",
		StartURL:       "http://docs.local/start",
		SameOriginOnly: &sameOriginOnly,
	})
	if err != nil {
		t.Fatalf("NormalizeAIBrowserControlRunRequest() error = %v", err)
	}
	if normalized.StartURL != "http://docs.local/start" {
		t.Fatalf("expected allowed cross-origin URL, got %q", normalized.StartURL)
	}
}

func TestNormalizeAIBrowserControlRunRequestRejectsUnsafeLimits(t *testing.T) {
	project := Project{FrontendURL: "http://demo-web:8080/", AllowedHosts: []string{"demo-web"}, AllowPrivateTargets: true}
	if _, err := NormalizeAIBrowserControlRunRequest(project, AIBrowserControlRunRequest{ProviderID: "provider-1", Goal: "Explore safely.", MaxSteps: maxAIBrowserMaxSteps + 1}); err == nil {
		t.Fatalf("expected max_steps over limit to fail")
	}
	if _, err := NormalizeAIBrowserControlRunRequest(project, AIBrowserControlRunRequest{ProviderID: "provider-1", Goal: "Explore safely.", MaxDepth: maxAIBrowserMaxDepth + 1}); err == nil {
		t.Fatalf("expected max_depth over limit to fail")
	}
}

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNormalizeIssueExportConfigRequestRequiresEncryptedTokenInput(t *testing.T) {
	enabled := true
	req, err := NormalizeIssueExportConfigRequest(IssueExportConfigRequest{
		Provider:            "GitHub",
		Name:                " GitHub Issues ",
		OwnerOrNamespace:    "Operalith",
		RepositoryOrProject: "qualora",
		Token:               "ghp_secret",
		DefaultLabels:       []string{"qa", "qa", "password=secret"},
		Enabled:             &enabled,
	})
	if err != nil {
		t.Fatalf("NormalizeIssueExportConfigRequest() error = %v", err)
	}
	if req.Provider != IssueProviderGitHub {
		t.Fatalf("provider = %q, want github", req.Provider)
	}
	if len(req.DefaultLabels) != 2 {
		t.Fatalf("expected labels to be deduplicated/redacted, got %#v", req.DefaultLabels)
	}
	if strings.Contains(strings.Join(req.DefaultLabels, ","), "secret") {
		t.Fatalf("label leaked secret: %#v", req.DefaultLabels)
	}
	if _, err := NormalizeIssueExportConfigRequest(IssueExportConfigRequest{Provider: "github", Name: "Missing token", OwnerOrNamespace: "o", RepositoryOrProject: "r"}); err == nil {
		t.Fatal("expected missing token to be rejected")
	}
}

func TestBuildIssueExportPreviewUsesGroupedSanitizedHighSignalFindings(t *testing.T) {
	dryRun := true
	deduplicate := true
	input, err := NormalizeIssueExportRequest(IssueExportRequest{
		SeverityThreshold:        "high",
		MaxIssues:                2,
		DryRun:                   &dryRun,
		DeduplicateByFingerprint: &deduplicate,
		Labels:                   []string{"regression"},
	})
	if err != nil {
		t.Fatalf("NormalizeIssueExportRequest() error = %v", err)
	}
	high := testGroupedFinding("fp-high", "high")
	high.Title = "Authorization token=abc leaked"
	high.Summary = "Bearer abc.def.ghi appeared in a safe summary"
	high.Recommendation = "Rotate password=secret and fix the page"
	high.AffectedURLs = []string{"https://app.example.com/admin?token=secret"}
	medium := testGroupedFinding("fp-medium", "medium")
	snapshot := testSnapshot([]GroupedFinding{medium, high})
	result := BuildIssueExportPreview(snapshot, input, IssueExportConfig{DefaultLabels: []string{"qualora"}}, ReportTypeSafeQA, "qa-run-1", time.Now().UTC())

	if !result.DryRun || result.Status != "dry_run" {
		t.Fatalf("unexpected dry-run status: %#v", result)
	}
	if len(result.IssuesToCreate) != 1 {
		t.Fatalf("expected one high issue preview, got %#v", result.IssuesToCreate)
	}
	rendered := result.IssuesToCreate[0].Title + "\n" + result.IssuesToCreate[0].Body
	if strings.Contains(rendered, "abc.def.ghi") || strings.Contains(rendered, "password=secret") || strings.Contains(rendered, "token=secret") {
		t.Fatalf("issue preview leaked secret material:\n%s", rendered)
	}
	if !strings.Contains(rendered, "screenshots") || !strings.Contains(rendered, "Fingerprint") {
		t.Fatalf("issue preview missed safety/fingerprint content:\n%s", rendered)
	}
	if len(result.SkippedFindings) != 1 || result.SkippedFindings[0].Reason != "below_severity_threshold" {
		t.Fatalf("expected medium finding to be skipped by threshold: %#v", result.SkippedFindings)
	}
}

func TestGitHubIssueClientUsesFakeHTTPServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/Operalith/qualora/issues" {
			t.Fatalf("unexpected GitHub path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer gh-token" {
			t.Fatalf("unexpected auth header %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["title"] == "" || payload["body"] == "" {
			t.Fatalf("payload missed title/body: %#v", payload)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"html_url":"https://github.example/issues/1"}`))
	}))
	defer server.Close()

	client := httpIssueTrackerClient{httpClient: server.Client()}
	issueURL, err := client.CreateIssue(t.Context(), IssueExportConfig{
		Provider:            IssueProviderGitHub,
		BaseURL:             server.URL,
		OwnerOrNamespace:    "Operalith",
		RepositoryOrProject: "qualora",
	}, "gh-token", issueCreateRequest{Title: "t", Body: "b", Labels: []string{"qa"}})
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}
	if issueURL != "https://github.example/issues/1" {
		t.Fatalf("issueURL = %q", issueURL)
	}
}

func TestGitLabIssueClientUsesFakeHTTPServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.EscapedPath(), "/api/v4/projects/Operalith%2Fqualora/issues") {
			t.Fatalf("unexpected GitLab path %s", r.URL.EscapedPath())
		}
		if got := r.Header.Get("PRIVATE-TOKEN"); got != "gl-token" {
			t.Fatalf("unexpected private token header %q", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"web_url":"https://gitlab.example/issues/1"}`))
	}))
	defer server.Close()

	client := httpIssueTrackerClient{httpClient: server.Client()}
	issueURL, err := client.CreateIssue(t.Context(), IssueExportConfig{
		Provider:            IssueProviderGitLab,
		BaseURL:             server.URL,
		OwnerOrNamespace:    "Operalith",
		RepositoryOrProject: "qualora",
	}, "gl-token", issueCreateRequest{Title: "t", Body: "b", Labels: []string{"qa"}})
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}
	if issueURL != "https://gitlab.example/issues/1" {
		t.Fatalf("issueURL = %q", issueURL)
	}
}

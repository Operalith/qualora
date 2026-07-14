package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRequestSafeAPIOperationBlocksExternalRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://example.com/outside", http.StatusFound)
	}))
	defer server.Close()

	target, err := url.Parse(server.URL + "/redirect")
	if err != nil {
		t.Fatal(err)
	}
	status, _, _, redirectBlocked, err := requestSafeAPIOperation(context.Background(), http.MethodGet, target)
	if err != nil {
		t.Fatalf("requestSafeAPIOperation returned error: %v", err)
	}
	if status == nil || *status != http.StatusFound {
		t.Fatalf("expected 302 status, got %#v", status)
	}
	if !redirectBlocked {
		t.Fatal("expected external redirect to be blocked")
	}
}

func TestBuildAPIOperationFindingsFor5xxAndUnexpectedStatus(t *testing.T) {
	status := 500
	duration := 12
	size := 20
	result := APICheckResult{
		Method:              "GET",
		Path:                "/broken",
		ResolvedURL:         "http://api.example.test/broken",
		Status:              StatusPassed,
		HTTPStatus:          &status,
		DurationMS:          &duration,
		ResponseContentType: "application/json",
		ResponseSizeBytes:   &size,
	}
	operation := APIOperation{
		Method:               "GET",
		Path:                 "/broken",
		ExpectedStatuses:     []string{"200"},
		ExpectedContentTypes: []string{"application/json"},
	}

	findings := buildAPIOperationFindings(operation, result, []byte(`{"error":"boom"}`), false)
	if len(findings) < 2 {
		t.Fatalf("expected 5xx and unexpected status findings, got %#v", findings)
	}
	if findings[0].Severity != "high" {
		t.Fatalf("expected high severity 5xx finding, got %#v", findings[0])
	}
}

func TestBuildAPIOperationFindingsForInvalidJSON(t *testing.T) {
	status := 200
	result := APICheckResult{
		Method:              "GET",
		Path:                "/bad-json",
		ResolvedURL:         "http://api.example.test/bad-json",
		Status:              StatusPassed,
		HTTPStatus:          &status,
		ResponseContentType: "application/json; charset=utf-8",
	}
	operation := APIOperation{
		Method:               "GET",
		Path:                 "/bad-json",
		ExpectedStatuses:     []string{"200"},
		ExpectedContentTypes: []string{"application/json"},
	}

	findings := buildAPIOperationFindings(operation, result, []byte(`{"broken":`), false)
	if len(findings) != 1 {
		t.Fatalf("expected one invalid JSON finding, got %#v", findings)
	}
	if findings[0].Title != "API endpoint returned invalid JSON" || findings[0].Severity != "medium" {
		t.Fatalf("unexpected invalid JSON finding: %#v", findings[0])
	}
}

func TestResolveAPIOperationURLPreservesBasePathAndQuery(t *testing.T) {
	base, err := url.Parse("http://api.example.test/v1")
	if err != nil {
		t.Fatal(err)
	}
	target, err := resolveAPIOperationURL(base, APIOperation{ResolvedPath: "/users/1", QueryString: "include=roles"})
	if err != nil {
		t.Fatalf("resolveAPIOperationURL returned error: %v", err)
	}
	if target.String() != "http://api.example.test/v1/users/1?include=roles" {
		t.Fatalf("unexpected target URL %q", target.String())
	}
}

package main

import (
	"context"
	"net"
	"testing"
)

type fakeResolver struct {
	records map[string][]net.IPAddr
	err     error
}

func (r fakeResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.records[host], nil
}

var publicResolver = fakeResolver{
	records: map[string][]net.IPAddr{
		"api.example.com": {{IP: net.ParseIP("93.184.216.34")}},
		"example.com":     {{IP: net.ParseIP("93.184.216.34")}},
		"example.net":     {{IP: net.ParseIP("93.184.216.34")}},
	},
}

func TestNormalizeProjectRequestAcceptsExampleDotCom(t *testing.T) {
	req, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "Example App",
		FrontendURL:  "https://example.com",
		AllowedHosts: []string{"example.com"},
	}, publicResolver)
	if err != nil {
		t.Fatalf("expected valid project request: %v", err)
	}
	if req.SecurityMode != "passive" {
		t.Fatalf("expected passive security mode, got %q", req.SecurityMode)
	}
	if len(req.AllowedHosts) != 1 || req.AllowedHosts[0] != "example.com" {
		t.Fatalf("unexpected allowed hosts: %#v", req.AllowedHosts)
	}
}

func TestNormalizeProjectRequestRejectsHostOutsideAllowlist(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "Wrong Host",
		FrontendURL:  "https://example.net",
		AllowedHosts: []string{"example.com"},
	}, publicResolver)
	if err == nil {
		t.Fatal("expected host outside allowlist to be rejected")
	}
}

func TestNormalizeProjectRequestAcceptsAPIOnlyProject(t *testing.T) {
	req, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "API App",
		APIBaseURL:   "https://api.example.com",
		OpenAPIURL:   "https://api.example.com/openapi.json",
		AllowedHosts: []string{"api.example.com"},
	}, publicResolver)
	if err != nil {
		t.Fatalf("expected valid API-only project request: %v", err)
	}
	if req.FrontendURL != "" {
		t.Fatalf("expected empty frontend_url, got %q", req.FrontendURL)
	}
	if req.APIBaseURL != "https://api.example.com" {
		t.Fatalf("unexpected api_base_url: %q", req.APIBaseURL)
	}
}

func TestNormalizeProjectRequestRejectsProjectWithoutTargets(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "No Targets",
		AllowedHosts: []string{"example.com"},
	}, publicResolver)
	if err == nil {
		t.Fatal("expected project with no targets to be rejected")
	}
}

func TestNormalizeProjectRequestRejectsOpenAPIURLOutsideAllowlist(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "Wrong OpenAPI Host",
		APIBaseURL:   "https://api.example.com",
		OpenAPIURL:   "https://example.net/openapi.json",
		AllowedHosts: []string{"api.example.com"},
	}, publicResolver)
	if err == nil {
		t.Fatal("expected OpenAPI URL outside allowlist to be rejected")
	}
}

func TestNormalizeProjectRequestRejectsLocalhostByDefault(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "Local App",
		FrontendURL:  "http://localhost:3000",
		AllowedHosts: []string{"localhost"},
	}, publicResolver)
	if err == nil {
		t.Fatal("expected localhost to be rejected by default")
	}
}

func TestNormalizeProjectRequestAllowsLocalhostWhenExplicitlyEnabled(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:                "Local App",
		FrontendURL:         "http://localhost:3000",
		AllowedHosts:        []string{"localhost"},
		AllowPrivateTargets: true,
	}, publicResolver)
	if err != nil {
		t.Fatalf("expected localhost to be allowed with allow_private_targets: %v", err)
	}
}

func TestNormalizeProjectRequestRejectsMetadataAddressByDefault(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "Metadata",
		FrontendURL:  "http://169.254.169.254/latest/meta-data",
		AllowedHosts: []string{"169.254.169.254"},
	}, publicResolver)
	if err == nil {
		t.Fatal("expected metadata address to be rejected by default")
	}
}

func TestNormalizeProjectRequestRejectsNonHTTPURL(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "FTP",
		FrontendURL:  "ftp://example.com",
		AllowedHosts: []string{"example.com"},
	}, publicResolver)
	if err == nil {
		t.Fatal("expected non-http URL to be rejected")
	}
}

func TestNormalizeProjectRequestRejectsAllowedHostWithPath(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "Path Host",
		FrontendURL:  "https://example.com",
		AllowedHosts: []string{"https://example.com/path"},
	}, publicResolver)
	if err == nil {
		t.Fatal("expected allowed host with path to be rejected")
	}
}

func TestNormalizeProjectRequestRejectsPublicHostResolvingToPrivateIP(t *testing.T) {
	_, err := NormalizeProjectRequestWithResolver(CreateProjectRequest{
		Name:         "Rebinding App",
		FrontendURL:  "https://example.com",
		AllowedHosts: []string{"example.com"},
	}, fakeResolver{
		records: map[string][]net.IPAddr{
			"example.com": {{IP: net.ParseIP("10.0.0.10")}},
		},
	})
	if err == nil {
		t.Fatal("expected private DNS resolution to be rejected")
	}
}

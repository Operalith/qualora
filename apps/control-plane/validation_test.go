package main

import "testing"

func TestNormalizeProjectRequestAcceptsExampleDotCom(t *testing.T) {
	req, err := NormalizeProjectRequest(CreateProjectRequest{
		Name:         "Example App",
		FrontendURL:  "https://example.com",
		AllowedHosts: []string{"example.com"},
	})
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
	_, err := NormalizeProjectRequest(CreateProjectRequest{
		Name:         "Wrong Host",
		FrontendURL:  "https://example.net",
		AllowedHosts: []string{"example.com"},
	})
	if err == nil {
		t.Fatal("expected host outside allowlist to be rejected")
	}
}

func TestNormalizeProjectRequestRejectsLocalhostByDefault(t *testing.T) {
	_, err := NormalizeProjectRequest(CreateProjectRequest{
		Name:         "Local App",
		FrontendURL:  "http://localhost:3000",
		AllowedHosts: []string{"localhost"},
	})
	if err == nil {
		t.Fatal("expected localhost to be rejected by default")
	}
}

func TestNormalizeProjectRequestAllowsLocalhostWhenExplicitlyEnabled(t *testing.T) {
	_, err := NormalizeProjectRequest(CreateProjectRequest{
		Name:                "Local App",
		FrontendURL:         "http://localhost:3000",
		AllowedHosts:        []string{"localhost"},
		AllowPrivateTargets: true,
	})
	if err != nil {
		t.Fatalf("expected localhost to be allowed with allow_private_targets: %v", err)
	}
}

func TestNormalizeProjectRequestRejectsMetadataAddressByDefault(t *testing.T) {
	_, err := NormalizeProjectRequest(CreateProjectRequest{
		Name:         "Metadata",
		FrontendURL:  "http://169.254.169.254/latest/meta-data",
		AllowedHosts: []string{"169.254.169.254"},
	})
	if err == nil {
		t.Fatal("expected metadata address to be rejected by default")
	}
}

func TestNormalizeProjectRequestRejectsNonHTTPURL(t *testing.T) {
	_, err := NormalizeProjectRequest(CreateProjectRequest{
		Name:         "FTP",
		FrontendURL:  "ftp://example.com",
		AllowedHosts: []string{"example.com"},
	})
	if err == nil {
		t.Fatal("expected non-http URL to be rejected")
	}
}

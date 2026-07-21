package main

import "testing"

func TestNormalizeFormTestRunRequestDefaultsToSafeGET(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	input, err := NormalizeFormTestRunRequest(project, FormTestRunRequest{})
	if err != nil {
		t.Fatalf("normalize form test request: %v", err)
	}
	if input.TargetURL != "http://demo-web:8080/" {
		t.Fatalf("unexpected target url: %q", input.TargetURL)
	}
	if input.SafeGetOnly == nil || !*input.SafeGetOnly {
		t.Fatal("safe_get_only should default to true")
	}
	if input.MaxForms != defaultFormTestMaxForms {
		t.Fatalf("unexpected max forms: %d", input.MaxForms)
	}
	if input.MaxTestsPerForm != defaultFormTestMaxTestsPerForm {
		t.Fatalf("unexpected max tests per form: %d", input.MaxTestsPerForm)
	}
}

func TestNormalizeFormTestRunRequestRejectsExternalTargets(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	_, err := NormalizeFormTestRunRequest(project, FormTestRunRequest{TargetURL: "https://example.com/search"})
	if err == nil {
		t.Fatal("expected external target to be rejected")
	}
}

func TestNormalizeFormTestRunRequestRejectsSensitiveTargetQuery(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	_, err := NormalizeFormTestRunRequest(project, FormTestRunRequest{TargetURL: "http://demo-web:8080/search?token=secret"})
	if err == nil {
		t.Fatal("expected sensitive target query to be rejected")
	}
}

func TestNormalizeReportTypeSupportsFormTest(t *testing.T) {
	got, err := NormalizeReportType("safe_forms")
	if err != nil {
		t.Fatalf("normalize report type: %v", err)
	}
	if got != ReportTypeFormTest {
		t.Fatalf("expected %q, got %q", ReportTypeFormTest, got)
	}
}

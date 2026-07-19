package main

import "testing"

func TestNormalizeCIRunRequestDefaultsToSafeWorkflow(t *testing.T) {
	project := Project{ID: "project-1", FrontendURL: "https://app.example.com", AllowedHosts: []string{"app.example.com"}}
	req, err := NormalizeCIRunRequest(project, CIRunRequest{})
	if err != nil {
		t.Fatalf("NormalizeCIRunRequest() error = %v", err)
	}
	if req.Mode != ReportTypeSafeQA {
		t.Fatalf("mode = %q, want safe_qa", req.Mode)
	}
	if req.UseLatestBaseline == nil || !*req.UseLatestBaseline {
		t.Fatal("expected use_latest_baseline default true")
	}
	if req.RunSafeQA == nil || !*req.RunSafeQA {
		t.Fatal("expected run_safe_qa default true")
	}
	if req.IncludeQualityChecks == nil || !*req.IncludeQualityChecks {
		t.Fatal("expected include_quality_checks default true")
	}
	if req.ExecuteSafePlan == nil || !*req.ExecuteSafePlan {
		t.Fatal("expected execute_safe_plan default true")
	}
	if req.IssueExportDryRun == nil || !*req.IssueExportDryRun {
		t.Fatal("expected issue_export_dry_run default true")
	}
	if req.TimeoutSeconds != defaultCIRunTimeoutSeconds {
		t.Fatalf("timeout = %d, want %d", req.TimeoutSeconds, defaultCIRunTimeoutSeconds)
	}
}

func TestNormalizeCIRunRequestAllowsAIFreeLatestReportGate(t *testing.T) {
	runSafeQA := false
	project := Project{ID: "project-1", FrontendURL: "", AllowedHosts: []string{"app.example.com"}}
	req, err := NormalizeCIRunRequest(project, CIRunRequest{RunSafeQA: &runSafeQA})
	if err != nil {
		t.Fatalf("NormalizeCIRunRequest() error = %v", err)
	}
	if req.RunSafeQA == nil || *req.RunSafeQA {
		t.Fatal("expected run_safe_qa=false to be preserved")
	}
}

func TestNormalizeCIRunRequestRejectsUnsafeLimits(t *testing.T) {
	project := Project{ID: "project-1", FrontendURL: "https://app.example.com"}
	if _, err := NormalizeCIRunRequest(project, CIRunRequest{Mode: "autonomous"}); err == nil {
		t.Fatal("expected unsupported mode to be rejected")
	}
	if _, err := NormalizeCIRunRequest(project, CIRunRequest{MaxPages: maxDiscoveryMaxPages + 1}); err == nil {
		t.Fatal("expected max_pages to be capped")
	}
	if _, err := NormalizeCIRunRequest(project, CIRunRequest{TimeoutSeconds: 5}); err == nil {
		t.Fatal("expected timeout_seconds to be capped")
	}
}

func TestCIStatusFromGate(t *testing.T) {
	for gateStatus, want := range map[string]string{
		QualityGateStatusPassed:  CIRunStatusPassed,
		QualityGateStatusWarning: CIRunStatusWarning,
		QualityGateStatusFailed:  CIRunStatusFailed,
		"mystery":                CIRunStatusError,
	} {
		if got := ciStatusFromGate(gateStatus); got != want {
			t.Fatalf("ciStatusFromGate(%q) = %q, want %q", gateStatus, got, want)
		}
	}
}

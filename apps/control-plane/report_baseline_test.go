package main

import (
	"testing"
	"time"
)

func TestBuildReportBaselineFromSnapshotStoresGroupedFingerprints(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	snapshot := ReportSnapshot{
		ProjectID:  "project-1",
		ReportType: ReportTypeSafeQA,
		ReportID:   "qa-run-1",
		Status:     StatusCompleted,
		Intelligence: ReportIntelligence{
			GroupedFindings:  []GroupedFinding{testGroupedFinding("fp-b", "medium"), testGroupedFinding("fp-a", "high")},
			SeverityCounts:   ReportSummary{TotalFindings: 2, High: 1, Medium: 1},
			RawFindingsCount: 4,
		},
	}

	baseline := BuildReportBaselineFromSnapshot(ReportBaselineRequest{Name: "Release baseline", ReportType: ReportTypeSafeQA, ReportID: "qa-run-1", IsDefault: true}, snapshot, "user-1", now)
	if baseline.ProjectID != "project-1" || baseline.ReportType != ReportTypeSafeQA || baseline.ReportID != "qa-run-1" {
		t.Fatalf("baseline did not preserve identity: %#v", baseline)
	}
	if !baseline.IsDefault || baseline.CreatedByUserID != "user-1" {
		t.Fatalf("baseline did not preserve default/user metadata: %#v", baseline)
	}
	if baseline.GroupedFindingsCount != 2 || baseline.RawFindingsCount != 4 {
		t.Fatalf("unexpected baseline counts: %#v", baseline)
	}
	if baseline.FingerprintSet[0].Fingerprint != "fp-a" || baseline.FingerprintSet[1].Fingerprint != "fp-b" {
		t.Fatalf("fingerprint set was not sorted deterministically: %#v", baseline.FingerprintSet)
	}
}

func TestCompareReportToBaselineClassifiesUnchanged(t *testing.T) {
	now := time.Now().UTC()
	baseline := testBaseline([]GroupedFinding{testGroupedFinding("fp-a", "high"), testGroupedFinding("fp-b", "medium")})
	snapshot := testSnapshot([]GroupedFinding{testGroupedFinding("fp-a", "high"), testGroupedFinding("fp-b", "medium")})

	comparison := CompareReportToBaseline(snapshot, baseline, now)
	if comparison.Status != ComparisonStatusUnchanged {
		t.Fatalf("status = %q, want unchanged: %#v", comparison.Status, comparison)
	}
	if comparison.Summary.NewFindingsCount != 0 || comparison.Summary.FixedFindingsCount != 0 || comparison.Summary.UnchangedFindingsCount != 2 {
		t.Fatalf("unexpected summary: %#v", comparison.Summary)
	}
}

func TestCompareReportToBaselineClassifiesRegressed(t *testing.T) {
	baseline := testBaseline([]GroupedFinding{testGroupedFinding("fp-a", "medium")})
	snapshot := testSnapshot([]GroupedFinding{testGroupedFinding("fp-a", "medium"), testGroupedFinding("fp-b", "high")})

	comparison := CompareReportToBaseline(snapshot, baseline, time.Now().UTC())
	if comparison.Status != ComparisonStatusRegressed {
		t.Fatalf("status = %q, want regressed", comparison.Status)
	}
	if comparison.Summary.NewHigh != 1 || comparison.Summary.NewFindingsCount != 1 {
		t.Fatalf("unexpected new finding summary: %#v", comparison.Summary)
	}
	if comparison.SeverityDelta.High != 1 || comparison.SeverityDelta.TotalFindings != 1 {
		t.Fatalf("unexpected severity delta: %#v", comparison.SeverityDelta)
	}
}

func TestCompareReportToBaselineClassifiesImproved(t *testing.T) {
	baseline := testBaseline([]GroupedFinding{testGroupedFinding("fp-a", "high"), testGroupedFinding("fp-b", "medium")})
	snapshot := testSnapshot([]GroupedFinding{testGroupedFinding("fp-b", "medium")})

	comparison := CompareReportToBaseline(snapshot, baseline, time.Now().UTC())
	if comparison.Status != ComparisonStatusImproved {
		t.Fatalf("status = %q, want improved", comparison.Status)
	}
	if comparison.Summary.FixedHigh != 1 || comparison.Summary.FixedFindingsCount != 1 {
		t.Fatalf("unexpected fixed finding summary: %#v", comparison.Summary)
	}
}

func TestCompareReportToBaselineClassifiesMixed(t *testing.T) {
	baseline := testBaseline([]GroupedFinding{testGroupedFinding("fp-a", "medium")})
	snapshot := testSnapshot([]GroupedFinding{testGroupedFinding("fp-b", "medium")})

	comparison := CompareReportToBaseline(snapshot, baseline, time.Now().UTC())
	if comparison.Status != ComparisonStatusMixed {
		t.Fatalf("status = %q, want mixed", comparison.Status)
	}
	if comparison.Summary.NewMedium != 1 || comparison.Summary.FixedMedium != 1 {
		t.Fatalf("unexpected mixed summary: %#v", comparison.Summary)
	}
}

func TestCompareReportToBaselineRecordsSeverityAndScopeChanges(t *testing.T) {
	baselineGroup := testGroupedFinding("fp-a", "low")
	baselineGroup.AffectedURLs = []string{"https://example.com/a"}
	currentGroup := testGroupedFinding("fp-a", "medium")
	currentGroup.AffectedURLs = []string{"https://example.com/a", "https://example.com/b"}
	baseline := testBaseline([]GroupedFinding{baselineGroup})
	snapshot := testSnapshot([]GroupedFinding{currentGroup})

	comparison := CompareReportToBaseline(snapshot, baseline, time.Now().UTC())
	if len(comparison.Summary.SeverityChanges) != 1 {
		t.Fatalf("expected severity change: %#v", comparison.Summary.SeverityChanges)
	}
	if len(comparison.AffectedPagesDelta) != 1 {
		t.Fatalf("expected affected scope change: %#v", comparison.AffectedPagesDelta)
	}
}

func TestEvaluateQualityGatePassesUnchangedComparison(t *testing.T) {
	comparison := ReportComparison{
		Status:  ComparisonStatusUnchanged,
		Summary: ReportComparisonSummary{UnchangedFindingsCount: 1},
	}
	result := EvaluateQualityGate(comparison, ReportSummary{TotalFindings: 1, Medium: 1}, StatusCompleted, QualityGateConfig{}, time.Now().UTC())
	if result.Status != QualityGateStatusPassed || result.CIExitCode != 0 {
		t.Fatalf("expected passed quality gate: %#v", result)
	}
}

func TestEvaluateQualityGateFailsOnNewHigh(t *testing.T) {
	comparison := ReportComparison{
		Status:      ComparisonStatusRegressed,
		NewFindings: []GroupedFinding{testGroupedFinding("fp-new", "high")},
		Summary:     ReportComparisonSummary{NewFindingsCount: 1, NewHigh: 1},
	}
	result := EvaluateQualityGate(comparison, ReportSummary{TotalFindings: 1, High: 1}, StatusCompleted, QualityGateConfig{}, time.Now().UTC())
	if result.Status != QualityGateStatusFailed || result.CIExitCode != 1 {
		t.Fatalf("expected failed quality gate: %#v", result)
	}
	if len(result.FailedRules) == 0 {
		t.Fatal("expected failed rules")
	}
}

func TestMissingBaselineQualityGateResultCanWarnWhenConfigured(t *testing.T) {
	failOnMissing := false
	result := MissingBaselineQualityGateResult(QualityGateConfig{FailOnMissingReport: &failOnMissing}, time.Now().UTC())
	if result.Status != QualityGateStatusWarning || result.CIExitCode != 0 {
		t.Fatalf("expected warning missing baseline result: %#v", result)
	}
}

func testGroupedFinding(fingerprint string, severity string) GroupedFinding {
	return GroupedFinding{
		GroupID:            fingerprint,
		Fingerprint:        fingerprint,
		Category:           "security",
		Title:              "Example finding " + fingerprint,
		NormalizedSeverity: severity,
		OccurrencesCount:   1,
		AffectedPaths:      []string{"/"},
		Sources:            []string{"safe_qa_run"},
		NoiseLevel:         reportNoiseLow,
	}
}

func testBaseline(groups []GroupedFinding) ReportBaseline {
	return ReportBaseline{
		ID:                   "baseline-1",
		ProjectID:            "project-1",
		ReportType:           ReportTypeSafeQA,
		ReportID:             "baseline-run",
		FingerprintSet:       stableGroupedFindingSet(groups),
		SeverityCounts:       summarizeGroupedFindings(groups),
		GroupedFindingsCount: len(groups),
		RawFindingsCount:     len(groups),
		IsDefault:            true,
	}
}

func testSnapshot(groups []GroupedFinding) ReportSnapshot {
	return ReportSnapshot{
		ProjectID:  "project-1",
		ReportType: ReportTypeSafeQA,
		ReportID:   "current-run",
		Status:     StatusCompleted,
		Intelligence: ReportIntelligence{
			GroupedFindings:  stableGroupedFindingSet(groups),
			SeverityCounts:   summarizeGroupedFindings(groups),
			RawFindingsCount: len(groups),
		},
	}
}

func summarizeGroupedFindings(groups []GroupedFinding) ReportSummary {
	summary := ReportSummary{TotalFindings: len(groups)}
	for _, group := range groups {
		switch group.NormalizedSeverity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		case "info":
			summary.Info++
		}
	}
	return summary
}

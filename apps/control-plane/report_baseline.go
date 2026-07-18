package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	ComparisonStatusImproved  = "improved"
	ComparisonStatusRegressed = "regressed"
	ComparisonStatusUnchanged = "unchanged"
	ComparisonStatusMixed     = "mixed"
	ComparisonStatusUnknown   = "unknown"

	QualityGateStatusPassed  = "passed"
	QualityGateStatusFailed  = "failed"
	QualityGateStatusWarning = "warning"
)

type ReportSnapshot struct {
	ProjectID    string
	ReportType   string
	ReportID     string
	SourceRunID  string
	Status       string
	Intelligence ReportIntelligence
}

func NormalizeReportType(value string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ReportTypeSafeQA, "safe_qa_run", "qa_run":
		return ReportTypeSafeQA, nil
	case ReportTypeQualityCheck, "quality":
		return ReportTypeQualityCheck, nil
	case ReportTypeDiscovery, "app_discovery":
		return ReportTypeDiscovery, nil
	case ReportTypeSafeExplorer:
		return ReportTypeSafeExplorer, nil
	case ReportTypeAPISmoke:
		return ReportTypeAPISmoke, nil
	case ReportTypeBrowserSmoke:
		return ReportTypeBrowserSmoke, nil
	case ReportTypeAuthorization, "authorization_check":
		return ReportTypeAuthorization, nil
	default:
		return "", fmt.Errorf("unsupported report_type %q", value)
	}
}

func NormalizeReportBaselineRequest(input ReportBaselineRequest) (ReportBaselineRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	reportType, err := NormalizeReportType(input.ReportType)
	if err != nil {
		return input, err
	}
	input.ReportType = reportType
	input.ReportID = strings.TrimSpace(input.ReportID)
	if len(input.Name) < 1 || len(input.Name) > 160 {
		return input, fmt.Errorf("name must be between 1 and 160 characters")
	}
	if len(input.Description) > 1000 {
		return input, fmt.Errorf("description must be at most 1000 characters")
	}
	if input.ReportID == "" {
		return input, fmt.Errorf("report_id is required")
	}
	return input, nil
}

func NormalizeReportBaselineUpdateRequest(input ReportBaselineUpdateRequest) (ReportBaselineUpdateRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name != "" && (len(input.Name) < 1 || len(input.Name) > 160) {
		return input, fmt.Errorf("name must be between 1 and 160 characters")
	}
	if len(input.Description) > 1000 {
		return input, fmt.Errorf("description must be at most 1000 characters")
	}
	return input, nil
}

func NormalizeReportComparisonRequest(input ReportComparisonRequest) (ReportComparisonRequest, error) {
	reportType, err := NormalizeReportType(input.ReportType)
	if err != nil {
		return input, err
	}
	input.ReportType = reportType
	input.CurrentReportID = strings.TrimSpace(input.CurrentReportID)
	input.BaselineID = strings.TrimSpace(input.BaselineID)
	if input.CurrentReportID == "" {
		return input, fmt.Errorf("current_report_id is required")
	}
	return input, nil
}

func NormalizeQualityGateEvaluationRequest(input QualityGateEvaluationRequest) (QualityGateEvaluationRequest, error) {
	reportType, err := NormalizeReportType(input.ReportType)
	if err != nil {
		return input, err
	}
	input.ReportType = reportType
	input.CurrentReportID = strings.TrimSpace(input.CurrentReportID)
	input.BaselineID = strings.TrimSpace(input.BaselineID)
	input.Format = strings.TrimSpace(strings.ToLower(input.Format))
	if input.CurrentReportID == "" {
		return input, fmt.Errorf("current_report_id is required")
	}
	if input.Format != "" && input.Format != "ci" {
		return input, fmt.Errorf("format must be ci when provided")
	}
	return input, nil
}

func BuildReportBaselineFromSnapshot(input ReportBaselineRequest, snapshot ReportSnapshot, createdByUserID string, now time.Time) ReportBaseline {
	groups := stableGroupedFindingSet(snapshot.Intelligence.GroupedFindings)
	return ReportBaseline{
		Name:                 input.Name,
		Description:          input.Description,
		ProjectID:            snapshot.ProjectID,
		ReportType:           snapshot.ReportType,
		ReportID:             snapshot.ReportID,
		SourceRunID:          snapshot.SourceRunID,
		FingerprintSet:       groups,
		SeverityCounts:       snapshot.Intelligence.SeverityCounts,
		GroupedFindingsCount: len(groups),
		RawFindingsCount:     snapshot.Intelligence.RawFindingsCount,
		CreatedByUserID:      createdByUserID,
		IsDefault:            input.IsDefault,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func CompareReportToBaseline(snapshot ReportSnapshot, baseline ReportBaseline, now time.Time) ReportComparison {
	current := stableGroupedFindingSet(snapshot.Intelligence.GroupedFindings)
	previous := stableGroupedFindingSet(baseline.FingerprintSet)
	currentByFingerprint := groupMapByFingerprint(current)
	previousByFingerprint := groupMapByFingerprint(previous)

	newFindings := []GroupedFinding{}
	fixedFindings := []GroupedFinding{}
	unchangedFindings := []GroupedFinding{}
	severityChanges := []SeverityChange{}
	scopeChanges := []AffectedScopeChange{}

	for fingerprint, currentGroup := range currentByFingerprint {
		previousGroup, ok := previousByFingerprint[fingerprint]
		if !ok {
			newFindings = append(newFindings, currentGroup)
			continue
		}
		unchangedFindings = append(unchangedFindings, currentGroup)
		if currentGroup.NormalizedSeverity != previousGroup.NormalizedSeverity {
			severityChanges = append(severityChanges, SeverityChange{
				Fingerprint:      fingerprint,
				Title:            currentGroup.Title,
				PreviousSeverity: previousGroup.NormalizedSeverity,
				CurrentSeverity:  currentGroup.NormalizedSeverity,
			})
		}
		if len(currentGroup.AffectedURLs) != len(previousGroup.AffectedURLs) || len(currentGroup.AffectedPaths) != len(previousGroup.AffectedPaths) {
			scopeChanges = append(scopeChanges, AffectedScopeChange{
				Fingerprint:           fingerprint,
				Title:                 currentGroup.Title,
				PreviousAffectedURLs:  len(previousGroup.AffectedURLs),
				CurrentAffectedURLs:   len(currentGroup.AffectedURLs),
				PreviousAffectedPaths: len(previousGroup.AffectedPaths),
				CurrentAffectedPaths:  len(currentGroup.AffectedPaths),
			})
		}
	}
	for fingerprint, previousGroup := range previousByFingerprint {
		if _, ok := currentByFingerprint[fingerprint]; !ok {
			fixedFindings = append(fixedFindings, previousGroup)
		}
	}

	sortGroupedFindings(newFindings)
	sortGroupedFindings(fixedFindings)
	sortGroupedFindings(unchangedFindings)
	sortSeverityChanges(severityChanges)
	sortScopeChanges(scopeChanges)

	summary := ReportComparisonSummary{
		NewFindingsCount:       len(newFindings),
		FixedFindingsCount:     len(fixedFindings),
		UnchangedFindingsCount: len(unchangedFindings),
		SeverityChanges:        severityChanges,
		NewCritical:            countGroupedBySeverity(newFindings, "critical"),
		NewHigh:                countGroupedBySeverity(newFindings, "high"),
		NewMedium:              countGroupedBySeverity(newFindings, "medium"),
		FixedCritical:          countGroupedBySeverity(fixedFindings, "critical"),
		FixedHigh:              countGroupedBySeverity(fixedFindings, "high"),
		FixedMedium:            countGroupedBySeverity(fixedFindings, "medium"),
	}
	status := classifyComparisonStatus(summary, scopeChanges)
	return ReportComparison{
		ProjectID:          snapshot.ProjectID,
		ReportType:         snapshot.ReportType,
		BaselineID:         baseline.ID,
		CurrentReportID:    snapshot.ReportID,
		Status:             status,
		Summary:            summary,
		NewFindings:        newFindings,
		FixedFindings:      fixedFindings,
		UnchangedFindings:  unchangedFindings,
		SeverityDelta:      subtractReportSummary(snapshot.Intelligence.SeverityCounts, baseline.SeverityCounts),
		AffectedPagesDelta: scopeChanges,
		Recommendation:     comparisonRecommendation(status, summary),
		GeneratedAt:        now,
	}
}

func EvaluateQualityGate(comparison ReportComparison, severityCounts ReportSummary, runStatus string, config QualityGateConfig, now time.Time) QualityGateResult {
	failedRules := []string{}
	warnings := []string{}
	newCounts := countGateFindingsBySeverity(comparison.NewFindings, config)

	if boolConfig(config.FailOnRunError, true) && isReportErrorStatus(runStatus) {
		failedRules = append(failedRules, "run_status_error")
	}
	if boolConfig(config.FailOnNewCritical, true) && newCounts.Critical > 0 {
		failedRules = append(failedRules, "new_critical_findings")
	}
	if boolConfig(config.FailOnNewHigh, true) && newCounts.High > 0 {
		failedRules = append(failedRules, "new_high_findings")
	}
	if boolConfig(config.FailOnNewMedium, false) && newCounts.Medium > 0 {
		failedRules = append(failedRules, "new_medium_findings")
	}
	maxNewHigh := intConfig(config.MaxNewHigh, 0)
	if newCounts.High > maxNewHigh {
		failedRules = append(failedRules, fmt.Sprintf("max_new_high_exceeded:%d>%d", newCounts.High, maxNewHigh))
	}
	if maxNewMedium, ok := optionalIntConfig(config.MaxNewMedium); ok && newCounts.Medium > maxNewMedium {
		failedRules = append(failedRules, fmt.Sprintf("max_new_medium_exceeded:%d>%d", newCounts.Medium, maxNewMedium))
	}
	maxTotalCritical := intConfig(config.MaxTotalCritical, 0)
	if severityCounts.Critical > maxTotalCritical {
		failedRules = append(failedRules, fmt.Sprintf("max_total_critical_exceeded:%d>%d", severityCounts.Critical, maxTotalCritical))
	}
	if maxTotalHigh, ok := optionalIntConfig(config.MaxTotalHigh); ok && severityCounts.High > maxTotalHigh {
		failedRules = append(failedRules, fmt.Sprintf("max_total_high_exceeded:%d>%d", severityCounts.High, maxTotalHigh))
	}

	if comparison.Status == ComparisonStatusUnknown {
		warnings = append(warnings, "comparison_status_unknown")
	}
	if len(failedRules) == 0 && (countGroupedBySeverity(comparison.NewFindings, "low") > 0 || (!boolConfig(config.IgnoreInfo, true) && countGroupedBySeverity(comparison.NewFindings, "info") > 0)) {
		warnings = append(warnings, "new_low_or_info_findings")
	}

	status := QualityGateStatusPassed
	exitCode := 0
	if len(failedRules) > 0 {
		status = QualityGateStatusFailed
		exitCode = 1
	} else if len(warnings) > 0 {
		status = QualityGateStatusWarning
	}
	return QualityGateResult{
		Status:            status,
		FailedRules:       stableStrings(failedRules),
		Warnings:          stableStrings(warnings),
		ComparisonSummary: comparison.Summary,
		SeverityCounts:    severityCounts,
		Recommendation:    qualityGateRecommendation(status),
		CIExitCode:        exitCode,
		GeneratedAt:       now,
	}
}

func MissingBaselineQualityGateResult(config QualityGateConfig, now time.Time) QualityGateResult {
	failedRules := []string{}
	warnings := []string{"missing_baseline"}
	status := QualityGateStatusWarning
	exitCode := 0
	if boolConfig(config.FailOnMissingReport, true) {
		failedRules = append(failedRules, "missing_baseline")
		warnings = nil
		status = QualityGateStatusFailed
		exitCode = 1
	}
	return QualityGateResult{
		Status:         status,
		FailedRules:    failedRules,
		Warnings:       warnings,
		Recommendation: "Create a default baseline before enforcing regression quality gates.",
		CIExitCode:     exitCode,
		GeneratedAt:    now,
	}
}

func CIQualityGateResponse(projectID string, request QualityGateEvaluationRequest, result QualityGateResult) CIQualityGateResult {
	return CIQualityGateResult{
		Status:      result.Status,
		ExitCode:    result.CIExitCode,
		Summary:     fmt.Sprintf("%s: %d new, %d fixed, %d unchanged", result.Status, result.ComparisonSummary.NewFindingsCount, result.ComparisonSummary.FixedFindingsCount, result.ComparisonSummary.UnchangedFindingsCount),
		ReportURL:   fmt.Sprintf("/api/v1/projects/%s/report-comparisons", projectID),
		FailedRules: result.FailedRules,
	}
}

func stableGroupedFindingSet(groups []GroupedFinding) []GroupedFinding {
	out := make([]GroupedFinding, 0, len(groups))
	for _, group := range groups {
		if strings.TrimSpace(group.Fingerprint) == "" {
			continue
		}
		copied := group
		copied.AffectedURLs = stableStrings(group.AffectedURLs)
		copied.AffectedPaths = stableStrings(group.AffectedPaths)
		copied.Sources = stableStrings(group.Sources)
		copied.RawOccurrenceRefs = append([]ReportFindingOccurrenceRef{}, group.RawOccurrenceRefs...)
		out = append(out, copied)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Fingerprint < out[j].Fingerprint
	})
	return out
}

func groupMapByFingerprint(groups []GroupedFinding) map[string]GroupedFinding {
	out := make(map[string]GroupedFinding, len(groups))
	for _, group := range groups {
		if group.Fingerprint == "" {
			continue
		}
		out[group.Fingerprint] = group
	}
	return out
}

func countGroupedBySeverity(groups []GroupedFinding, severity string) int {
	count := 0
	for _, group := range groups {
		if group.NormalizedSeverity == severity {
			count++
		}
	}
	return count
}

func countGateFindingsBySeverity(groups []GroupedFinding, config QualityGateConfig) ReportSummary {
	summary := ReportSummary{}
	for _, group := range groups {
		if boolConfig(config.IgnoreNoisy, true) && group.NoiseLevel == reportNoiseHigh {
			continue
		}
		if boolConfig(config.IgnoreInfo, true) && group.NormalizedSeverity == "info" {
			continue
		}
		summary.TotalFindings++
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

func classifyComparisonStatus(summary ReportComparisonSummary, scopeChanges []AffectedScopeChange) string {
	newMediumPlus := summary.NewCritical + summary.NewHigh + summary.NewMedium
	fixedMediumPlus := summary.FixedCritical + summary.FixedHigh + summary.FixedMedium
	if summary.NewCritical+summary.NewHigh > 0 {
		return ComparisonStatusRegressed
	}
	if newMediumPlus > 0 && fixedMediumPlus > 0 {
		return ComparisonStatusMixed
	}
	if newMediumPlus > 0 {
		return ComparisonStatusRegressed
	}
	if fixedMediumPlus > 0 && summary.NewFindingsCount == 0 {
		return ComparisonStatusImproved
	}
	if summary.NewFindingsCount == 0 && summary.FixedFindingsCount == 0 && len(summary.SeverityChanges) == 0 && len(scopeChanges) == 0 {
		return ComparisonStatusUnchanged
	}
	return ComparisonStatusMixed
}

func subtractReportSummary(current ReportSummary, baseline ReportSummary) ReportSummary {
	return ReportSummary{
		TotalFindings: current.TotalFindings - baseline.TotalFindings,
		Critical:      current.Critical - baseline.Critical,
		High:          current.High - baseline.High,
		Medium:        current.Medium - baseline.Medium,
		Low:           current.Low - baseline.Low,
		Info:          current.Info - baseline.Info,
	}
}

func sortSeverityChanges(changes []SeverityChange) {
	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].Fingerprint == changes[j].Fingerprint {
			return changes[i].Title < changes[j].Title
		}
		return changes[i].Fingerprint < changes[j].Fingerprint
	})
}

func sortScopeChanges(changes []AffectedScopeChange) {
	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].Fingerprint == changes[j].Fingerprint {
			return changes[i].Title < changes[j].Title
		}
		return changes[i].Fingerprint < changes[j].Fingerprint
	})
}

func comparisonRecommendation(status string, summary ReportComparisonSummary) string {
	switch status {
	case ComparisonStatusRegressed:
		return "Review new critical, high, and medium grouped findings before release."
	case ComparisonStatusMixed:
		return "Review both new and fixed grouped findings; decide whether the remaining regression risk is acceptable."
	case ComparisonStatusImproved:
		return "No new medium-or-higher findings were detected; review fixed findings and keep monitoring."
	case ComparisonStatusUnchanged:
		return "No deterministic grouped-finding changes were detected against the selected baseline."
	default:
		if summary.NewFindingsCount+summary.FixedFindingsCount == 0 {
			return "Configure a baseline with grouped findings to enable regression tracking."
		}
		return "Review comparison output manually."
	}
}

func qualityGateRecommendation(status string) string {
	switch status {
	case QualityGateStatusFailed:
		return "Treat this report as a failed CI quality gate and inspect failed_rules before release."
	case QualityGateStatusWarning:
		return "The gate did not fail, but warnings should be reviewed before relying on the result."
	default:
		return "The report satisfies the configured alpha quality gate."
	}
}

func boolConfig(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func intConfig(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func optionalIntConfig(value *int) (int, bool) {
	if value == nil {
		return 0, false
	}
	return *value, true
}

func isReportErrorStatus(status string) bool {
	switch status {
	case StatusFailed, StatusError, StatusCanceled:
		return true
	default:
		return false
	}
}

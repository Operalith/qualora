package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	reportNoiseHigh   = "high"
	reportNoiseMedium = "medium"
	reportNoiseLow    = "low"
)

type ReportFindingOccurrenceRef struct {
	SourceType      string `json:"source_type"`
	SourceRunID     string `json:"source_run_id,omitempty"`
	FindingID       string `json:"finding_id,omitempty"`
	QualityResultID string `json:"quality_result_id,omitempty"`
	EvidenceID      string `json:"evidence_id,omitempty"`
	Category        string `json:"category,omitempty"`
	Severity        string `json:"severity,omitempty"`
	AffectedURL     string `json:"affected_url,omitempty"`
	AffectedPath    string `json:"affected_path,omitempty"`
}

type NormalizedFinding struct {
	SourceType         string                     `json:"source_type"`
	SourceRunID        string                     `json:"source_run_id,omitempty"`
	Category           string                     `json:"category"`
	Title              string                     `json:"title"`
	Severity           string                     `json:"severity"`
	NormalizedSeverity string                     `json:"normalized_severity"`
	AffectedURL        string                     `json:"affected_url,omitempty"`
	AffectedPath       string                     `json:"affected_path,omitempty"`
	AffectedComponent  string                     `json:"affected_component,omitempty"`
	RuleID             string                     `json:"rule_id,omitempty"`
	Summary            string                     `json:"summary,omitempty"`
	Recommendation     string                     `json:"recommendation,omitempty"`
	EvidenceID         string                     `json:"evidence_id,omitempty"`
	Confidence         string                     `json:"confidence,omitempty"`
	Fingerprint        string                     `json:"fingerprint"`
	DedupGroupKey      string                     `json:"dedup_group_key"`
	FirstSeenAt        *time.Time                 `json:"first_seen_at,omitempty"`
	CreatedAt          *time.Time                 `json:"created_at,omitempty"`
	RawRef             ReportFindingOccurrenceRef `json:"raw_ref"`
}

type GroupedFinding struct {
	GroupID                  string                       `json:"group_id"`
	Fingerprint              string                       `json:"fingerprint"`
	Category                 string                       `json:"category"`
	Title                    string                       `json:"title"`
	NormalizedSeverity       string                       `json:"normalized_severity"`
	Summary                  string                       `json:"summary,omitempty"`
	Recommendation           string                       `json:"recommendation,omitempty"`
	OccurrencesCount         int                          `json:"occurrences_count"`
	AffectedURLs             []string                     `json:"affected_urls,omitempty"`
	AffectedPaths            []string                     `json:"affected_paths,omitempty"`
	Sources                  []string                     `json:"sources"`
	RepresentativeEvidenceID string                       `json:"representative_evidence_id,omitempty"`
	Confidence               string                       `json:"confidence,omitempty"`
	NoiseLevel               string                       `json:"noise_level"`
	RawOccurrenceRefs        []ReportFindingOccurrenceRef `json:"raw_occurrence_refs"`
}

type AffectedPageSummary struct {
	URL             string `json:"url,omitempty"`
	Path            string `json:"path,omitempty"`
	FindingsCount   int    `json:"findings_count"`
	HighestSeverity string `json:"highest_severity"`
}

type NoiseSummary struct {
	HighNoise      int `json:"high_noise"`
	MediumNoise    int `json:"medium_noise"`
	LowNoise       int `json:"low_noise"`
	HighSignal     int `json:"high_signal"`
	NeedsAttention int `json:"needs_attention"`
	Informational  int `json:"informational"`
	NoisyRepeated  int `json:"noisy_repeated"`
}

type DeduplicationSummary struct {
	RawFindingsCount           int `json:"raw_findings_count"`
	GroupedFindingsCount       int `json:"grouped_findings_count"`
	DuplicateFindingsReduced   int `json:"duplicate_findings_reduced"`
	GroupedRepeatedFindings    int `json:"grouped_repeated_findings"`
	CrossSourceGroupedFindings int `json:"cross_source_grouped_findings"`
}

type ReportExecutiveSummary struct {
	OverallStatus          string        `json:"overall_status"`
	Headline               string        `json:"headline"`
	TotalFindings          int           `json:"total_findings"`
	GroupedFindings        int           `json:"grouped_findings"`
	SeverityCounts         ReportSummary `json:"severity_counts"`
	ChecksCompleted        []string      `json:"checks_completed"`
	ChecksSkipped          []string      `json:"checks_skipped"`
	RecommendedNextActions []string      `json:"recommended_next_actions"`
	WhatWasTested          []string      `json:"what_was_tested"`
	WhatWasNotTested       []string      `json:"what_was_not_tested"`
	SafetyLimitations      []string      `json:"safety_limitations"`
}

type ReportIntelligence struct {
	ExecutiveSummary     ReportExecutiveSummary `json:"executive_summary"`
	SeverityCounts       ReportSummary          `json:"severity_counts"`
	GroupedFindings      []GroupedFinding       `json:"grouped_findings"`
	TopFindings          []GroupedFinding       `json:"top_findings"`
	TopAffectedPages     []AffectedPageSummary  `json:"top_affected_pages"`
	NoiseSummary         NoiseSummary           `json:"noise_summary"`
	RawFindingsCount     int                    `json:"raw_findings_count"`
	DeduplicationSummary DeduplicationSummary   `json:"deduplication_summary"`
	SafetyLimitations    []string               `json:"safety_limitations"`
}

type ReportIntelligenceInput struct {
	ReportType        string
	ReportID          string
	Status            string
	Project           *Project
	Findings          []Finding
	QualityResults    []QualityCheckResult
	Evidence          []Evidence
	ChecksCompleted   []string
	ChecksSkipped     []string
	WhatWasTested     []string
	WhatWasNotTested  []string
	SafetyLimitations []string
}

func QualityResultsToFindings(results []QualityCheckResult) []Finding {
	findings := make([]Finding, 0, len(results))
	for _, result := range results {
		findings = append(findings, Finding{
			ID:             result.ID,
			Title:          result.Title,
			Severity:       result.Severity,
			Category:       result.Category,
			Confidence:     "medium",
			Description:    result.Description,
			Recommendation: result.Recommendation,
			CreatedAt:      result.CreatedAt,
		})
	}
	return findings
}

func BuildReportIntelligence(input ReportIntelligenceInput) ReportIntelligence {
	normalized := normalizeReportFindings(input)
	groups := groupNormalizedFindings(normalized)
	sortGroupedFindings(groups)

	severityCounts := summarizeNormalizedFindings(normalized)
	topFindings := limitGroupedFindings(groups, 5)
	topAffectedPages := topAffectedPages(groups, 5)
	noiseSummary := summarizeNoise(groups)
	dedupSummary := summarizeDeduplication(len(normalized), groups)
	limitations := stableStrings(input.SafetyLimitations)
	executive := buildExecutiveSummary(input, severityCounts, len(normalized), len(groups), topFindings, limitations)

	return ReportIntelligence{
		ExecutiveSummary:     executive,
		SeverityCounts:       severityCounts,
		GroupedFindings:      groups,
		TopFindings:          topFindings,
		TopAffectedPages:     topAffectedPages,
		NoiseSummary:         noiseSummary,
		RawFindingsCount:     len(normalized),
		DeduplicationSummary: dedupSummary,
		SafetyLimitations:    limitations,
	}
}

func normalizeReportFindings(input ReportIntelligenceInput) []NormalizedFinding {
	evidenceByID := map[string]Evidence{}
	for _, record := range input.Evidence {
		if record.ID != "" {
			evidenceByID[record.ID] = record
		}
	}

	normalized := make([]NormalizedFinding, 0, len(input.Findings)+len(input.QualityResults))
	for _, finding := range input.Findings {
		sourceType := inferFindingSourceType(finding, input.ReportType)
		sourceRunID := inferFindingSourceRunID(finding, input.ReportID)
		affectedURL, affectedPath, evidenceID := affectedLocationForFinding(finding, evidenceByID)
		createdAt := finding.CreatedAt
		item := NormalizedFinding{
			SourceType:         sourceType,
			SourceRunID:        sourceRunID,
			Category:           normalizeToken(finding.Category),
			Title:              normalizeDisplayText(finding.Title, "Untitled finding"),
			Severity:           normalizeToken(finding.Severity),
			NormalizedSeverity: NormalizeFindingSeverity(finding.Severity, finding.Category, finding.Title, ""),
			AffectedURL:        affectedURL,
			AffectedPath:       affectedPath,
			Summary:            strings.TrimSpace(finding.Description),
			Recommendation:     strings.TrimSpace(finding.Recommendation),
			EvidenceID:         evidenceID,
			Confidence:         normalizeToken(finding.Confidence),
			CreatedAt:          &createdAt,
			RawRef: ReportFindingOccurrenceRef{
				SourceType:   sourceType,
				SourceRunID:  sourceRunID,
				FindingID:    finding.ID,
				EvidenceID:   evidenceID,
				Category:     normalizeToken(finding.Category),
				Severity:     normalizeToken(finding.Severity),
				AffectedURL:  affectedURL,
				AffectedPath: affectedPath,
			},
		}
		item.Fingerprint = FindingFingerprint(item)
		item.DedupGroupKey = FindingDedupGroupKey(item)
		normalized = append(normalized, item)
	}

	for _, result := range input.QualityResults {
		affectedURL, affectedPath := NormalizeFindingURL(result.URL)
		createdAt := result.CreatedAt
		item := NormalizedFinding{
			SourceType:         RunTypeQualityCheck,
			SourceRunID:        result.RunID,
			Category:           normalizeToken(result.Category),
			Title:              normalizeDisplayText(result.Title, "Untitled quality finding"),
			Severity:           normalizeToken(result.Severity),
			NormalizedSeverity: NormalizeFindingSeverity(result.Severity, result.Category, result.Title, result.RuleID),
			AffectedURL:        affectedURL,
			AffectedPath:       affectedPath,
			RuleID:             normalizeToken(result.RuleID),
			Summary:            strings.TrimSpace(result.Description),
			Recommendation:     strings.TrimSpace(result.Recommendation),
			Confidence:         "medium",
			CreatedAt:          &createdAt,
			RawRef: ReportFindingOccurrenceRef{
				SourceType:      RunTypeQualityCheck,
				SourceRunID:     result.RunID,
				QualityResultID: result.ID,
				Category:        normalizeToken(result.Category),
				Severity:        normalizeToken(result.Severity),
				AffectedURL:     affectedURL,
				AffectedPath:    affectedPath,
			},
		}
		item.Fingerprint = FindingFingerprint(item)
		item.DedupGroupKey = FindingDedupGroupKey(item)
		normalized = append(normalized, item)
	}
	return normalized
}

func NormalizeFindingSeverity(rawSeverity string, category string, title string, ruleID string) string {
	severity := normalizeToken(rawSeverity)
	category = normalizeToken(category)
	title = strings.ToLower(strings.TrimSpace(title))
	ruleID = normalizeToken(ruleID)

	switch severity {
	case "critical", "high", "medium", "low", "info":
		if category == "explorer_external_action_skipped" || strings.Contains(category, "external") && strings.Contains(category, "skipped") {
			return "info"
		}
		if category == "explorer_unsafe_action_skipped" && severity == "medium" {
			return "low"
		}
		return severity
	}

	switch {
	case category == "authorization_bypass":
		return "critical"
	case strings.Contains(category, "authorization") && (strings.Contains(category, "login_failure") || strings.Contains(category, "timeout") || strings.Contains(category, "access_denied")):
		return "high"
	case strings.Contains(category, "server_error") || strings.Contains(category, "5xx") || strings.Contains(title, "server error"):
		return "high"
	case ruleID == "missing_csp" || category == "missing_csp" || strings.Contains(title, "content-security-policy"):
		return "medium"
	case ruleID == "missing_hsts" || category == "missing_hsts" || strings.Contains(title, "strict-transport-security"):
		return "medium"
	case strings.Contains(category, "not_found") || strings.Contains(category, "broken") || strings.Contains(title, "404"):
		return "medium"
	case strings.Contains(category, "console") || strings.Contains(category, "network"):
		return "medium"
	case strings.Contains(category, "performance") || strings.Contains(title, "slow"):
		return "medium"
	case strings.Contains(category, "accessibility") || strings.Contains(ruleID, "missing_alt") || strings.Contains(ruleID, "missing_label"):
		return "low"
	case strings.Contains(category, "external") && strings.Contains(category, "skipped"):
		return "info"
	case strings.Contains(category, "unsupported") || strings.Contains(category, "duplicate") || strings.Contains(category, "policy_blocked"):
		return "info"
	case category == "":
		return "info"
	default:
		return "low"
	}
}

func NormalizeFindingURL(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		if strings.HasPrefix(raw, "/") {
			return "", raw
		}
		return "", ""
	}
	parsed.Fragment = ""
	parsed = RedactSensitiveURLQuery(parsed)
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	if parsed.RawQuery != "" {
		path = path + "?" + parsed.RawQuery
	}
	return parsed.String(), path
}

func FindingFingerprint(finding NormalizedFinding) string {
	parts := []string{
		normalizeToken(finding.Category),
		normalizeToken(finding.NormalizedSeverity),
		normalizeToken(finding.RuleID),
		normalizeFingerprintTitle(finding.Title),
		fingerprintLocation(finding),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])[:16]
}

func FindingDedupGroupKey(finding NormalizedFinding) string {
	category := normalizeToken(finding.Category)
	ruleID := normalizeToken(finding.RuleID)
	title := normalizeFingerprintTitle(finding.Title)
	location := fingerprintLocation(finding)
	if shouldGroupAcrossPages(category, ruleID, title) {
		location = "all-pages"
	}
	parts := []string{category, normalizeToken(finding.NormalizedSeverity), ruleID, title, location}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])[:16]
}

func groupNormalizedFindings(normalized []NormalizedFinding) []GroupedFinding {
	byKey := map[string][]NormalizedFinding{}
	for _, item := range normalized {
		byKey[item.DedupGroupKey] = append(byKey[item.DedupGroupKey], item)
	}

	groups := make([]GroupedFinding, 0, len(byKey))
	for key, items := range byKey {
		sort.SliceStable(items, func(i, j int) bool {
			if severityRank(items[i].NormalizedSeverity) != severityRank(items[j].NormalizedSeverity) {
				return severityRank(items[i].NormalizedSeverity) > severityRank(items[j].NormalizedSeverity)
			}
			return items[i].Title < items[j].Title
		})
		rep := items[0]
		paths := map[string]bool{}
		urls := map[string]bool{}
		sources := map[string]bool{}
		refs := make([]ReportFindingOccurrenceRef, 0, len(items))
		for _, item := range items {
			if item.AffectedPath != "" {
				paths[item.AffectedPath] = true
			}
			if item.AffectedURL != "" {
				urls[item.AffectedURL] = true
			}
			if item.SourceType != "" {
				sources[item.SourceType] = true
			}
			refs = append(refs, item.RawRef)
		}
		group := GroupedFinding{
			GroupID:                  key,
			Fingerprint:              rep.Fingerprint,
			Category:                 rep.Category,
			Title:                    rep.Title,
			NormalizedSeverity:       rep.NormalizedSeverity,
			Summary:                  rep.Summary,
			Recommendation:           rep.Recommendation,
			OccurrencesCount:         len(items),
			AffectedURLs:             sortedMapKeys(urls, 12),
			AffectedPaths:            sortedMapKeys(paths, 12),
			Sources:                  sortedMapKeys(sources, 12),
			RepresentativeEvidenceID: rep.EvidenceID,
			Confidence:               representativeConfidence(items),
			NoiseLevel:               classifyNoise(rep, len(items)),
			RawOccurrenceRefs:        refs,
		}
		groups = append(groups, group)
	}
	return groups
}

func sortGroupedFindings(groups []GroupedFinding) {
	sort.SliceStable(groups, func(i, j int) bool {
		leftSignal := noiseSortRank(groups[i].NoiseLevel)
		rightSignal := noiseSortRank(groups[j].NoiseLevel)
		if severityRank(groups[i].NormalizedSeverity) != severityRank(groups[j].NormalizedSeverity) {
			return severityRank(groups[i].NormalizedSeverity) > severityRank(groups[j].NormalizedSeverity)
		}
		if leftSignal != rightSignal {
			return leftSignal < rightSignal
		}
		if groups[i].OccurrencesCount != groups[j].OccurrencesCount {
			return groups[i].OccurrencesCount > groups[j].OccurrencesCount
		}
		if groups[i].Category != groups[j].Category {
			return groups[i].Category < groups[j].Category
		}
		return groups[i].Title < groups[j].Title
	})
}

func classifyNoise(finding NormalizedFinding, occurrences int) string {
	category := normalizeToken(finding.Category)
	severity := normalizeToken(finding.NormalizedSeverity)
	switch {
	case severity == "critical" || severity == "high":
		return reportNoiseLow
	case strings.Contains(category, "missing_csp") || strings.Contains(category, "missing_hsts") || strings.Contains(category, "security"):
		return reportNoiseLow
	case strings.Contains(category, "authorization_bypass"):
		return reportNoiseLow
	case strings.Contains(category, "external") && strings.Contains(category, "skipped"):
		return reportNoiseHigh
	case strings.Contains(category, "unsupported") || strings.Contains(category, "duplicate"):
		if occurrences > 1 {
			return reportNoiseHigh
		}
		return reportNoiseMedium
	case strings.Contains(category, "policy_blocked") || strings.Contains(category, "unsafe_action"):
		if occurrences > 2 {
			return reportNoiseHigh
		}
		return reportNoiseMedium
	case strings.Contains(category, "console") || strings.Contains(category, "network"):
		if occurrences > 1 {
			return reportNoiseMedium
		}
		return reportNoiseLow
	case severity == "info":
		return reportNoiseMedium
	default:
		return reportNoiseLow
	}
}

func summarizeNormalizedFindings(findings []NormalizedFinding) ReportSummary {
	summary := ReportSummary{TotalFindings: len(findings)}
	for _, finding := range findings {
		switch finding.NormalizedSeverity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		default:
			summary.Info++
		}
	}
	return summary
}

func summarizeNoise(groups []GroupedFinding) NoiseSummary {
	summary := NoiseSummary{}
	for _, group := range groups {
		switch group.NoiseLevel {
		case reportNoiseHigh:
			summary.HighNoise++
			summary.NoisyRepeated++
		case reportNoiseMedium:
			summary.MediumNoise++
			if group.OccurrencesCount > 1 {
				summary.NoisyRepeated++
			}
		default:
			summary.LowNoise++
		}
		switch group.NormalizedSeverity {
		case "critical", "high":
			summary.HighSignal++
		case "medium":
			summary.NeedsAttention++
		default:
			summary.Informational++
		}
	}
	return summary
}

func summarizeDeduplication(rawCount int, groups []GroupedFinding) DeduplicationSummary {
	summary := DeduplicationSummary{
		RawFindingsCount:         rawCount,
		GroupedFindingsCount:     len(groups),
		DuplicateFindingsReduced: rawCount - len(groups),
	}
	if summary.DuplicateFindingsReduced < 0 {
		summary.DuplicateFindingsReduced = 0
	}
	for _, group := range groups {
		if group.OccurrencesCount > 1 {
			summary.GroupedRepeatedFindings++
		}
		if len(group.Sources) > 1 {
			summary.CrossSourceGroupedFindings++
		}
	}
	return summary
}

func buildExecutiveSummary(input ReportIntelligenceInput, severity ReportSummary, rawCount int, groupedCount int, top []GroupedFinding, limitations []string) ReportExecutiveSummary {
	status := reportOverallStatus(input.Status, severity)
	checksCompleted := stableStrings(input.ChecksCompleted)
	checksSkipped := stableStrings(input.ChecksSkipped)
	whatWasTested := stableStrings(input.WhatWasTested)
	whatWasNotTested := stableStrings(input.WhatWasNotTested)
	if len(whatWasTested) == 0 {
		whatWasTested = defaultWhatWasTested(input.ReportType, input.Project)
	}
	if len(whatWasNotTested) == 0 {
		whatWasNotTested = defaultWhatWasNotTested(input.ReportType)
	}
	if len(limitations) == 0 {
		limitations = defaultReportSafetyLimitations(input.ReportType)
	}
	return ReportExecutiveSummary{
		OverallStatus:          status,
		Headline:               reportHeadline(input.ReportType, status, severity, groupedCount),
		TotalFindings:          rawCount,
		GroupedFindings:        groupedCount,
		SeverityCounts:         severity,
		ChecksCompleted:        checksCompleted,
		ChecksSkipped:          checksSkipped,
		RecommendedNextActions: recommendedNextActions(status, top),
		WhatWasTested:          whatWasTested,
		WhatWasNotTested:       whatWasNotTested,
		SafetyLimitations:      limitations,
	}
}

func reportOverallStatus(runStatus string, severity ReportSummary) string {
	switch normalizeToken(runStatus) {
	case StatusCompleted, StatusPassed:
	case "":
		return "unknown"
	default:
		return "unknown"
	}
	if severity.Critical > 0 || severity.High > 0 {
		return "fail"
	}
	if severity.Medium > 0 {
		return "warning"
	}
	return "pass"
}

func reportHeadline(reportType string, status string, severity ReportSummary, groupedCount int) string {
	label := friendlyReportType(reportType)
	switch status {
	case "fail":
		return fmt.Sprintf("%s found %d high-priority grouped issue(s) that should be reviewed first.", label, severity.Critical+severity.High)
	case "warning":
		return fmt.Sprintf("%s completed with medium-priority issues and no high-priority grouped findings.", label)
	case "pass":
		return fmt.Sprintf("%s completed without medium, high, or critical findings.", label)
	default:
		if groupedCount > 0 {
			return fmt.Sprintf("%s has %d grouped finding(s), but the run status is not completed.", label, groupedCount)
		}
		return fmt.Sprintf("%s has not completed enough checks for a final report status.", label)
	}
}

func recommendedNextActions(status string, top []GroupedFinding) []string {
	actions := []string{}
	switch status {
	case "fail":
		actions = append(actions, "Review critical and high grouped findings first, then inspect representative evidence and affected paths.")
	case "warning":
		actions = append(actions, "Review medium grouped findings and decide which should become regression checks.")
	case "pass":
		actions = append(actions, "Keep the report as a baseline and expand safe coverage where the limitations list calls out gaps.")
	default:
		actions = append(actions, "Wait for the run to complete or rerun the check before using this report for release decisions.")
	}
	if len(top) > 0 {
		actions = append(actions, fmt.Sprintf("Start with: %s.", top[0].Title))
	}
	actions = append(actions, "Use raw findings for evidence-level details; grouped findings are a deterministic summary, not a replacement.")
	return actions
}

func topAffectedPages(groups []GroupedFinding, limit int) []AffectedPageSummary {
	type pageAccumulator struct {
		url      string
		path     string
		count    int
		severity string
	}
	pages := map[string]*pageAccumulator{}
	for _, group := range groups {
		paths := group.AffectedPaths
		urls := group.AffectedURLs
		if len(paths) == 0 && len(urls) == 0 {
			continue
		}
		if len(paths) == 0 {
			paths = urls
		}
		for index, path := range paths {
			urlValue := ""
			if index < len(urls) {
				urlValue = urls[index]
			}
			key := path
			if key == "" {
				key = urlValue
			}
			if key == "" {
				continue
			}
			if pages[key] == nil {
				pages[key] = &pageAccumulator{url: urlValue, path: path, severity: group.NormalizedSeverity}
			}
			pages[key].count += group.OccurrencesCount
			if severityRank(group.NormalizedSeverity) > severityRank(pages[key].severity) {
				pages[key].severity = group.NormalizedSeverity
			}
		}
	}
	out := make([]AffectedPageSummary, 0, len(pages))
	for _, page := range pages {
		out = append(out, AffectedPageSummary{
			URL:             page.url,
			Path:            page.path,
			FindingsCount:   page.count,
			HighestSeverity: page.severity,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if severityRank(out[i].HighestSeverity) != severityRank(out[j].HighestSeverity) {
			return severityRank(out[i].HighestSeverity) > severityRank(out[j].HighestSeverity)
		}
		if out[i].FindingsCount != out[j].FindingsCount {
			return out[i].FindingsCount > out[j].FindingsCount
		}
		return out[i].Path < out[j].Path
	})
	if len(out) > limit {
		return out[:limit]
	}
	return out
}

func affectedLocationForFinding(finding Finding, evidenceByID map[string]Evidence) (string, string, string) {
	for _, evidenceID := range finding.EvidenceIDs {
		record, ok := evidenceByID[evidenceID]
		if !ok {
			continue
		}
		if raw := firstMetadataString(record.Metadata, "target_url", "final_url", "page_url", "url", "source_url", "normalized_url", "request_url", "api_base_url", "openapi_url"); raw != "" {
			affectedURL, affectedPath := NormalizeFindingURL(raw)
			if affectedURL != "" || affectedPath != "" {
				return affectedURL, affectedPath, evidenceID
			}
		}
		if path := firstMetadataString(record.Metadata, "path", "route"); path != "" {
			return "", path, evidenceID
		}
		return "", "", evidenceID
	}
	return "", "", ""
}

func inferFindingSourceType(finding Finding, fallback string) string {
	switch {
	case finding.DiscoveryRunID != "":
		return RunTypeAppDiscovery
	case finding.SafeExplorerRunID != "":
		return RunTypeSafeExplorer
	case finding.AIBrowserControlRunID != "":
		return RunTypeAIBrowserControl
	case finding.AuthorizationRunID != "":
		return "authorization_check"
	case finding.TestPlanExecutionID != "":
		return "test_plan_execution"
	case finding.RunID != "":
		if fallback != "" {
			return fallback
		}
		return "run"
	default:
		if fallback != "" {
			return fallback
		}
		return "finding"
	}
}

func inferFindingSourceRunID(finding Finding, fallback string) string {
	switch {
	case finding.DiscoveryRunID != "":
		return finding.DiscoveryRunID
	case finding.SafeExplorerRunID != "":
		return finding.SafeExplorerRunID
	case finding.AuthorizationRunID != "":
		return finding.AuthorizationRunID
	case finding.TestPlanExecutionID != "":
		return finding.TestPlanExecutionID
	case finding.RunID != "":
		return finding.RunID
	default:
		return fallback
	}
}

func firstMetadataString(metadata map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := metadata[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case fmt.Stringer:
			if strings.TrimSpace(typed.String()) != "" {
				return strings.TrimSpace(typed.String())
			}
		}
	}
	return ""
}

func representativeConfidence(items []NormalizedFinding) string {
	best := "low"
	for _, item := range items {
		confidence := normalizeToken(item.Confidence)
		switch confidence {
		case "high":
			return "high"
		case "medium":
			best = "medium"
		}
	}
	return best
}

func limitGroupedFindings(groups []GroupedFinding, limit int) []GroupedFinding {
	if len(groups) <= limit {
		return append([]GroupedFinding{}, groups...)
	}
	return append([]GroupedFinding{}, groups[:limit]...)
}

func sortedMapKeys(values map[string]bool, limit int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if limit > 0 && len(keys) > limit {
		return keys[:limit]
	}
	return keys
}

func stableStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func friendlyReportType(reportType string) string {
	switch reportType {
	case RunTypeBrowserSmoke:
		return "Browser smoke"
	case RunTypeAuthenticatedBrowserSmoke:
		return "Authenticated browser smoke"
	case RunTypeAPISmoke:
		return "API smoke"
	case RunTypeLoginCheck:
		return "Login check"
	case RunTypeAppDiscovery:
		return "Discovery"
	case RunTypeQualityCheck:
		return "Quality check"
	case RunTypeSafeExplorer:
		return "Safe Explorer"
	case RunTypeAIBrowserControl:
		return "AI Browser Control"
	case "authorization_check":
		return "Authorization check"
	case "test_plan_execution":
		return "Safe test plan execution"
	case "safe_qa_run":
		return "Safe QA"
	default:
		if reportType == "" {
			return "Report"
		}
		return normalizeDisplayText(strings.ReplaceAll(reportType, "_", " "), "Report")
	}
}

func defaultWhatWasTested(reportType string, project *Project) []string {
	switch reportType {
	case RunTypeAppDiscovery:
		return []string{"Same-origin application discovery metadata", "Visible links, forms, page load status, console errors, and failed network requests"}
	case RunTypeQualityCheck:
		return []string{"Passive security header heuristics", "Basic accessibility heuristics", "Front-end and performance metadata"}
	case RunTypeSafeExplorer:
		return []string{"Classified safe navigation actions", "Observed page metadata, skip reasons, screenshots, console errors, and failed network requests"}
	case RunTypeAIBrowserControl:
		return []string{"Sanitized page observations", "AI-proposed typed actions", "Deterministic policy decisions", "Policy-approved safe browser actions", "Screenshots, console error counts, and failed request counts"}
	case "safe_qa_run":
		return []string{"Discovery, optional quality checks, AI-assisted test planning, and approved safe browser DSL execution"}
	case "authorization_check":
		return []string{"Configured browser URL authorization targets with deterministic role-aware checks"}
	case "test_plan_execution":
		return []string{"Approved deterministic safe browser DSL steps from a reviewed test plan"}
	case RunTypeAPISmoke:
		return []string{"Safe read-only API smoke operations and OpenAPI-derived GET, HEAD, and OPTIONS checks"}
	case RunTypeBrowserSmoke, RunTypeAuthenticatedBrowserSmoke:
		return []string{"Browser page load smoke checks, screenshots, console errors, and failed network requests"}
	default:
		out := []string{"Stored run findings and evidence metadata"}
		if project != nil && project.FrontendURL != "" {
			out = append(out, "Project frontend target metadata")
		}
		if project != nil && (project.APIBaseURL != "" || project.OpenAPIURL != "") {
			out = append(out, "Project API target metadata")
		}
		return out
	}
}

func whatWasTestedForRun(run *TestRun) []string {
	if run == nil {
		return nil
	}
	return defaultWhatWasTested(run.RunType, nil)
}

func defaultWhatWasNotTested(reportType string) []string {
	common := []string{"Destructive actions", "Active exploitation or fuzzing", "Autonomous AI browser control"}
	switch reportType {
	case RunTypeAPISmoke:
		return append(common, "Authenticated API requests", "Request body mutation and schema fuzzing")
	case RunTypeQualityCheck:
		return append(common, "Full WCAG certification", "Lighthouse scoring", "Penetration testing")
	case RunTypeAppDiscovery, RunTypeSafeExplorer:
		return append(common, "Arbitrary form submission", "Unsafe clicks and mutating actions", "External-domain crawling by default")
	case RunTypeAIBrowserControl:
		return []string{"Destructive actions", "Active exploitation or fuzzing", "Direct AI browser control", "Arbitrary form submission", "Unsafe clicks and mutating actions", "External-domain crawling by default", "Payload attacks"}
	case "authorization_check":
		return append(common, "API authorization checks with credentials", "Broad access-control crawling")
	case "safe_qa_run":
		return append(common, "Unsupported generated steps", "Free-form model-controlled browser actions")
	default:
		return common
	}
}

func whatWasNotTestedForRun(run *TestRun) []string {
	if run == nil {
		return nil
	}
	return defaultWhatWasNotTested(run.RunType)
}

func completedRunJobKinds(jobs []RunJob) []string {
	out := []string{}
	for _, job := range jobs {
		if job.Status == StatusCompleted || job.Status == StatusPassed {
			out = append(out, friendlyReportType(job.Kind))
		}
	}
	return out
}

func skippedRunJobKinds(jobs []RunJob) []string {
	out := []string{}
	for _, job := range jobs {
		if job.Status == StatusSkipped || job.Status == StatusCanceled {
			out = append(out, friendlyReportType(job.Kind))
		}
	}
	return out
}

func qualityCheckCompletedChecks(run QualityCheckRun) []string {
	out := []string{}
	if run.IncludeSecurity {
		out = append(out, "Passive security heuristics")
	}
	if run.IncludeAccessibility {
		out = append(out, "Basic accessibility heuristics")
	}
	if run.IncludePerformance {
		out = append(out, "Front-end performance metadata")
	}
	return out
}

func qualityCheckSkippedChecks(run QualityCheckRun) []string {
	out := []string{}
	if !run.IncludeSecurity {
		out = append(out, "Passive security heuristics")
	}
	if !run.IncludeAccessibility {
		out = append(out, "Basic accessibility heuristics")
	}
	if !run.IncludePerformance {
		out = append(out, "Front-end performance metadata")
	}
	out = append(out, "Active security scanning", "Form submission", "Payload execution")
	return out
}

func testPlanExecutionSkippedChecks(summary TestPlanExecutionSafetyReport) []string {
	out := []string{"Destructive actions", "Unsupported browser DSL steps", "Free-form AI browser control"}
	if summary.SkippedUnsafeSteps > 0 {
		out = append(out, fmt.Sprintf("%d unsafe step(s) skipped", summary.SkippedUnsafeSteps))
	}
	if summary.SkippedUnsupportedSteps > 0 {
		out = append(out, fmt.Sprintf("%d unsupported step(s) skipped", summary.SkippedUnsupportedSteps))
	}
	if summary.SkippedScenarios > 0 {
		out = append(out, fmt.Sprintf("%d scenario(s) skipped", summary.SkippedScenarios))
	}
	return out
}

func qaRunCompletedChecks(report *QARunReport) []string {
	if report == nil {
		return nil
	}
	out := []string{}
	if report.DiscoveryRun != nil {
		out = append(out, "Application discovery")
	}
	if report.QualityCheckRun != nil {
		out = append(out, "Passive quality checks")
	}
	if report.APISmokeRun != nil {
		out = append(out, "Safe API smoke and contract validation")
	}
	if report.TestPlan != nil {
		out = append(out, "AI-assisted test planning")
	}
	if report.ExecutionReport != nil {
		out = append(out, "Approved safe test plan execution")
	}
	return out
}

func qaRunSkippedChecks(report *QARunReport) []string {
	if report == nil {
		return nil
	}
	out := []string{}
	if report.QualityCheckRun == nil {
		out = append(out, "Passive quality checks")
	}
	if report.APISmokeRun == nil {
		out = append(out, "Safe API smoke and contract validation")
	}
	if report.TestPlan == nil {
		out = append(out, "AI-assisted test planning")
	}
	if report.ExecutionReport == nil {
		out = append(out, "Safe test plan execution")
	}
	out = append(out, "Active security scanning", "Destructive actions", "Autonomous browser control")
	return out
}

func defaultReportSafetyLimitations(reportType string) []string {
	return []string{
		"Reports are deterministic summaries of alpha checks and are not exhaustive QA certification.",
		"Grouped findings reduce repeated signals but raw findings remain available for evidence-level review.",
		"Credentials, cookies, local storage, session storage, auth headers, screenshots, full HTML, request bodies, and response bodies are not sent to AI by this report intelligence layer.",
	}
}

func shouldGroupAcrossPages(category string, ruleID string, title string) bool {
	for _, marker := range []string{
		"missing_csp",
		"missing_hsts",
		"missing_security_header",
		"security_header",
		"console",
		"network",
		"external_action_skipped",
		"external_link_skipped",
		"unsafe_action_skipped",
		"unsupported_action",
		"duplicate_page",
		"policy_blocked",
		"missing_alt",
		"missing_label",
		"missing_html_lang",
		"missing_main_landmark",
		"buttons_missing_names",
		"links_missing_names",
		"inputs_missing_labels",
		"images_missing_alt",
		"password_form_detected",
	} {
		if strings.Contains(category, marker) || strings.Contains(ruleID, marker) || strings.Contains(title, strings.ReplaceAll(marker, "_", " ")) {
			return true
		}
	}
	return false
}

func fingerprintLocation(finding NormalizedFinding) string {
	if finding.AffectedPath != "" {
		return finding.AffectedPath
	}
	if finding.AffectedURL != "" {
		return finding.AffectedURL
	}
	return "global"
}

func normalizeToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func normalizeDisplayText(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return strings.Join(strings.Fields(value), " ")
}

func normalizeFingerprintTitle(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func severityRank(severity string) int {
	switch normalizeToken(severity) {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	default:
		return 1
	}
}

func noiseSortRank(noise string) int {
	switch normalizeToken(noise) {
	case reportNoiseLow:
		return 0
	case reportNoiseMedium:
		return 1
	default:
		return 2
	}
}

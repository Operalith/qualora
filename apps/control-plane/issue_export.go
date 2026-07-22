package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	IssueProviderGitHub = "github"
	IssueProviderGitLab = "gitlab"

	defaultIssueTitlePrefix = "[Qualora]"
	defaultIssueMaxIssues   = 10
	maxIssueExportIssues    = 50
)

type issueCreateRequest struct {
	Title  string
	Body   string
	Labels []string
}

type issueTrackerClient interface {
	CreateIssue(ctx context.Context, config IssueExportConfig, token string, issue issueCreateRequest) (string, error)
}

type httpIssueTrackerClient struct {
	httpClient *http.Client
}

func NormalizeIssueExportConfigRequest(input IssueExportConfigRequest) (IssueExportConfigRequest, error) {
	input.Provider = normalizeToken(input.Provider)
	input.Name = strings.TrimSpace(RedactSecrets(input.Name))
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.OwnerOrNamespace = strings.Trim(strings.TrimSpace(RedactSecrets(input.OwnerOrNamespace)), "/")
	input.RepositoryOrProject = strings.Trim(strings.TrimSpace(RedactSecrets(input.RepositoryOrProject)), "/")
	input.DefaultLabels = normalizeIssueLabels(input.DefaultLabels)
	if input.Enabled == nil {
		value := true
		input.Enabled = &value
	}
	if input.Provider != IssueProviderGitHub && input.Provider != IssueProviderGitLab {
		return input, fmt.Errorf("provider must be github or gitlab")
	}
	if input.Name == "" || len(input.Name) > 160 {
		return input, fmt.Errorf("name must be between 1 and 160 characters")
	}
	if input.OwnerOrNamespace == "" || len(input.OwnerOrNamespace) > 200 {
		return input, fmt.Errorf("owner_or_namespace is required and must be at most 200 characters")
	}
	if input.RepositoryOrProject == "" || len(input.RepositoryOrProject) > 200 {
		return input, fmt.Errorf("repository_or_project is required and must be at most 200 characters")
	}
	if strings.TrimSpace(input.Token) == "" {
		return input, fmt.Errorf("token is required")
	}
	if input.BaseURL != "" {
		if err := validateIssueBaseURL(input.BaseURL); err != nil {
			return input, err
		}
	}
	return input, nil
}

func NormalizeIssueExportConfigUpdateRequest(existing IssueExportConfig, input IssueExportConfigUpdateRequest) (IssueExportConfigUpdateRequest, error) {
	input.Provider = normalizeToken(input.Provider)
	input.Name = strings.TrimSpace(RedactSecrets(input.Name))
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.OwnerOrNamespace = strings.Trim(strings.TrimSpace(RedactSecrets(input.OwnerOrNamespace)), "/")
	input.RepositoryOrProject = strings.Trim(strings.TrimSpace(RedactSecrets(input.RepositoryOrProject)), "/")
	input.DefaultLabels = normalizeIssueLabels(input.DefaultLabels)

	provider := existing.Provider
	if input.Provider != "" {
		provider = input.Provider
	}
	if provider != IssueProviderGitHub && provider != IssueProviderGitLab {
		return input, fmt.Errorf("provider must be github or gitlab")
	}
	name := existing.Name
	if input.Name != "" {
		name = input.Name
	}
	if name == "" || len(name) > 160 {
		return input, fmt.Errorf("name must be between 1 and 160 characters")
	}
	owner := existing.OwnerOrNamespace
	if input.OwnerOrNamespace != "" {
		owner = input.OwnerOrNamespace
	}
	repo := existing.RepositoryOrProject
	if input.RepositoryOrProject != "" {
		repo = input.RepositoryOrProject
	}
	if owner == "" || len(owner) > 200 {
		return input, fmt.Errorf("owner_or_namespace is required and must be at most 200 characters")
	}
	if repo == "" || len(repo) > 200 {
		return input, fmt.Errorf("repository_or_project is required and must be at most 200 characters")
	}
	if input.BaseURL != "" {
		if err := validateIssueBaseURL(input.BaseURL); err != nil {
			return input, err
		}
	}
	return input, nil
}

func NormalizeIssueExportRequest(input IssueExportRequest) (IssueExportRequest, error) {
	input.IssueExportConfigID = strings.TrimSpace(input.IssueExportConfigID)
	input.SeverityThreshold = normalizeToken(input.SeverityThreshold)
	if input.SeverityThreshold == "" {
		input.SeverityThreshold = "high"
	}
	if input.SeverityThreshold != "critical" && input.SeverityThreshold != "high" && input.SeverityThreshold != "medium" {
		return input, fmt.Errorf("severity_threshold must be critical, high, or medium")
	}
	if input.IncludeMedium && input.SeverityThreshold == "high" {
		input.SeverityThreshold = "medium"
	}
	if input.MaxIssues == 0 {
		input.MaxIssues = defaultIssueMaxIssues
	}
	if input.MaxIssues < 1 || input.MaxIssues > maxIssueExportIssues {
		return input, fmt.Errorf("max_issues must be between 1 and %d", maxIssueExportIssues)
	}
	if input.DryRun == nil {
		value := true
		input.DryRun = &value
	}
	if input.DeduplicateByFingerprint == nil {
		value := true
		input.DeduplicateByFingerprint = &value
	}
	input.Labels = normalizeIssueLabels(input.Labels)
	input.TitlePrefix = strings.TrimSpace(RedactSecrets(input.TitlePrefix))
	if input.TitlePrefix == "" {
		input.TitlePrefix = defaultIssueTitlePrefix
	}
	if !*input.DryRun && input.IssueExportConfigID == "" {
		return input, fmt.Errorf("issue_export_config_id is required when dry_run is false")
	}
	return input, nil
}

func (a *App) issueExportConfigFromRequest(input IssueExportConfigRequest, projectID string, existingEncryptedToken string) (IssueExportConfig, error) {
	encrypted := existingEncryptedToken
	if input.Token != "" {
		value, err := a.secretBox.Encrypt(input.Token)
		if err != nil {
			return IssueExportConfig{}, err
		}
		encrypted = value
	}
	return IssueExportConfig{
		ProjectID:           projectID,
		Provider:            input.Provider,
		Name:                input.Name,
		BaseURL:             input.BaseURL,
		OwnerOrNamespace:    input.OwnerOrNamespace,
		RepositoryOrProject: input.RepositoryOrProject,
		TokenEncrypted:      encrypted,
		TokenConfigured:     encrypted != "",
		DefaultLabels:       input.DefaultLabels,
		Enabled:             input.Enabled == nil || *input.Enabled,
	}, nil
}

func issueExportConfigFromUpdate(existing IssueExportConfig, input IssueExportConfigUpdateRequest, encryptedToken string) IssueExportConfig {
	updated := existing
	if input.Provider != "" {
		updated.Provider = input.Provider
	}
	if input.Name != "" {
		updated.Name = input.Name
	}
	if input.BaseURL != "" {
		updated.BaseURL = input.BaseURL
	}
	if input.OwnerOrNamespace != "" {
		updated.OwnerOrNamespace = input.OwnerOrNamespace
	}
	if input.RepositoryOrProject != "" {
		updated.RepositoryOrProject = input.RepositoryOrProject
	}
	if encryptedToken != "" {
		updated.TokenEncrypted = encryptedToken
	}
	if input.DefaultLabels != nil {
		updated.DefaultLabels = input.DefaultLabels
	}
	if input.Enabled != nil {
		updated.Enabled = *input.Enabled
	}
	updated.TokenConfigured = updated.TokenEncrypted != ""
	return updated
}

func (a *App) exportIssuesForSnapshot(ctx context.Context, reportType string, reportID string, snapshot ReportSnapshot, input IssueExportRequest) (*IssueExportResult, error) {
	normalized, err := NormalizeIssueExportRequest(input)
	if err != nil {
		return nil, err
	}
	config := IssueExportConfig{}
	token := ""
	if normalized.IssueExportConfigID != "" {
		loaded, err := a.store.GetIssueExportConfig(ctx, normalized.IssueExportConfigID)
		if err != nil {
			return nil, err
		}
		if loaded.ProjectID != snapshot.ProjectID {
			return nil, fmt.Errorf("issue export config does not belong to the report project")
		}
		if !loaded.Enabled {
			return nil, fmt.Errorf("issue export config is disabled")
		}
		config = *loaded
		token, err = a.secretBox.Decrypt(config.TokenEncrypted)
		if err != nil {
			return nil, fmt.Errorf("decrypt issue export token: %w", err)
		}
	}
	result := BuildIssueExportPreview(snapshot, normalized, config, reportType, reportID, time.Now().UTC())
	if *normalized.DryRun {
		return &result, nil
	}
	client := httpIssueTrackerClient{httpClient: &http.Client{Timeout: 15 * time.Second}}
	for _, preview := range result.IssuesToCreate {
		issueURL, err := client.CreateIssue(ctx, config, token, issueCreateRequest{Title: preview.Title, Body: preview.Body, Labels: preview.Labels})
		if err != nil {
			result.Errors = append(result.Errors, RedactSecrets(err.Error()))
			continue
		}
		result.IssueURLs = append(result.IssueURLs, issueURL)
		result.CreatedCount++
	}
	result.Status = "created"
	if len(result.Errors) > 0 {
		result.Status = "partial"
	}
	if result.CreatedCount == 0 && len(result.Errors) > 0 {
		result.Status = "failed"
	}
	return &result, nil
}

func BuildIssueExportPreview(snapshot ReportSnapshot, input IssueExportRequest, config IssueExportConfig, reportType string, reportID string, now time.Time) IssueExportResult {
	groups := stableGroupedFindingSet(snapshot.Intelligence.GroupedFindings)
	sortGroupedFindings(groups)
	seen := map[string]bool{}
	labels := stableStrings(append(config.DefaultLabels, input.Labels...))
	result := IssueExportResult{
		Provider:        config.Provider,
		DryRun:          input.DryRun == nil || *input.DryRun,
		Status:          "dry_run",
		IssuesToCreate:  []IssueExportPreview{},
		SkippedFindings: []IssueExportSkippedFinding{},
		Reasons:         []string{"grouped_findings_only", "sanitized_issue_content", "screenshots_and_raw_bodies_not_exported"},
		GeneratedAt:     now,
	}
	for _, group := range groups {
		if input.DeduplicateByFingerprint == nil || *input.DeduplicateByFingerprint {
			if seen[group.Fingerprint] {
				result.SkippedFindings = append(result.SkippedFindings, skippedIssueFinding(group, "duplicate_fingerprint"))
				continue
			}
			seen[group.Fingerprint] = true
		}
		if !severityMeetsThreshold(group.NormalizedSeverity, input.SeverityThreshold) {
			result.SkippedFindings = append(result.SkippedFindings, skippedIssueFinding(group, "below_severity_threshold"))
			continue
		}
		if len(result.IssuesToCreate) >= input.MaxIssues {
			result.SkippedFindings = append(result.SkippedFindings, skippedIssueFinding(group, "max_issues_reached"))
			continue
		}
		result.IssuesToCreate = append(result.IssuesToCreate, buildIssuePreview(group, input.TitlePrefix, labels, reportType, reportID))
	}
	result.SkippedCount = len(result.SkippedFindings)
	return result
}

func (c httpIssueTrackerClient) CreateIssue(ctx context.Context, config IssueExportConfig, token string, issue issueCreateRequest) (string, error) {
	switch config.Provider {
	case IssueProviderGitHub:
		return c.createGitHubIssue(ctx, config, token, issue)
	case IssueProviderGitLab:
		return c.createGitLabIssue(ctx, config, token, issue)
	default:
		return "", fmt.Errorf("unsupported issue export provider %q", config.Provider)
	}
}

func (c httpIssueTrackerClient) createGitHubIssue(ctx context.Context, config IssueExportConfig, token string, issue issueCreateRequest) (string, error) {
	base := strings.TrimRight(config.BaseURL, "/")
	if base == "" {
		base = "https://api.github.com"
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues", base, url.PathEscape(config.OwnerOrNamespace), url.PathEscape(config.RepositoryOrProject))
	payload := map[string]any{"title": issue.Title, "body": issue.Body, "labels": issue.Labels}
	req, err := jsonIssueRequest(ctx, endpoint, payload)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	return c.doIssueRequest(req, "html_url")
}

func (c httpIssueTrackerClient) createGitLabIssue(ctx context.Context, config IssueExportConfig, token string, issue issueCreateRequest) (string, error) {
	base := strings.TrimRight(config.BaseURL, "/")
	if base == "" {
		base = "https://gitlab.com"
	}
	projectPath := config.OwnerOrNamespace + "/" + config.RepositoryOrProject
	endpoint := fmt.Sprintf("%s/api/v4/projects/%s/issues", base, url.PathEscape(projectPath))
	payload := map[string]any{
		"title":       issue.Title,
		"description": issue.Body,
		"labels":      strings.Join(issue.Labels, ","),
	}
	req, err := jsonIssueRequest(ctx, endpoint, payload)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Authorization", "Bearer "+token)
	return c.doIssueRequest(req, "web_url")
}

func (c httpIssueTrackerClient) doIssueRequest(req *http.Request, urlField string) (string, error) {
	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("issue tracker returned HTTP %d: %s", resp.StatusCode, RedactSecrets(string(body)))
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("issue tracker response was not JSON: %w", err)
	}
	if value, ok := payload[urlField].(string); ok && value != "" {
		return value, nil
	}
	if value, ok := payload["url"].(string); ok && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("issue tracker response did not include an issue URL")
}

func jsonIssueRequest(ctx context.Context, endpoint string, payload map[string]any) (*http.Request, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Qualora/0.23.0-alpha")
	return req, nil
}

func buildIssuePreview(group GroupedFinding, prefix string, labels []string, reportType string, reportID string) IssueExportPreview {
	title := RedactSecrets(fmt.Sprintf("%s %s: %s", strings.TrimSpace(prefix), strings.ToUpper(group.NormalizedSeverity), group.Title))
	paths := stableStrings(append(group.AffectedPaths, pathsFromURLs(group.AffectedURLs)...))
	if len(paths) > 5 {
		paths = paths[:5]
	}
	body := strings.Join([]string{
		"## Summary",
		RedactSecrets(firstIssueNonEmpty(group.Summary, group.Title)),
		"",
		"## Severity",
		group.NormalizedSeverity,
		"",
		"## Category",
		firstIssueNonEmpty(group.Category, "quality"),
		"",
		"## Affected Pages",
		fmt.Sprintf("%d affected page(s)", maxInt(len(group.AffectedURLs), len(group.AffectedPaths))),
		"",
		"## Representative Affected Paths",
		formatIssueBulletList(paths),
		"",
		"## Recommendation",
		RedactSecrets(firstIssueNonEmpty(group.Recommendation, "Review the grouped finding in Qualora and verify the fix in a follow-up run.")),
		"",
		"## Source Report",
		fmt.Sprintf("/api/v1/reports/%s/%s/export-issues", reportType, reportID),
		"",
		"## Safety Note",
		"Generated from sanitized grouped findings only. Qualora did not export credentials, cookies, tokens, browser storage, authorization headers, screenshots, full HTML, request bodies, response bodies, or raw logs.",
		"",
		"## Fingerprint",
		group.Fingerprint,
		"",
		"Generated by Qualora 0.23.0-alpha.",
	}, "\n")
	body = RedactSecrets(body)
	return IssueExportPreview{
		Title:               title,
		Body:                body,
		Severity:            group.NormalizedSeverity,
		Category:            group.Category,
		AffectedPagesCount:  maxInt(len(group.AffectedURLs), len(group.AffectedPaths)),
		RepresentativePaths: paths,
		Labels:              labels,
		Fingerprint:         group.Fingerprint,
	}
}

func skippedIssueFinding(group GroupedFinding, reason string) IssueExportSkippedFinding {
	return IssueExportSkippedFinding{
		Fingerprint: group.Fingerprint,
		Title:       group.Title,
		Severity:    group.NormalizedSeverity,
		Reason:      reason,
	}
}

func severityMeetsThreshold(severity string, threshold string) bool {
	rank := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1, "info": 0}
	return rank[normalizeToken(severity)] >= rank[normalizeToken(threshold)]
}

func normalizeIssueLabels(labels []string) []string {
	out := make([]string, 0, len(labels))
	seen := map[string]bool{}
	for _, label := range labels {
		label = strings.TrimSpace(RedactSecrets(label))
		if label == "" || len(label) > 80 || seen[label] {
			continue
		}
		seen[label] = true
		out = append(out, label)
		if len(out) >= 20 {
			break
		}
	}
	sort.Strings(out)
	return out
}

func validateIssueBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("base_url must be a valid HTTP(S) URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("base_url must use http or https")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("base_url must not include query or fragment")
	}
	return nil
}

func pathsFromURLs(urls []string) []string {
	out := []string{}
	for _, raw := range urls {
		_, path := NormalizeFindingURL(raw)
		if path != "" {
			out = append(out, path)
		}
	}
	return out
}

func formatIssueBulletList(items []string) string {
	if len(items) == 0 {
		return "- Not recorded"
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "- "+RedactSecrets(item))
	}
	return strings.Join(lines, "\n")
}

func firstIssueNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

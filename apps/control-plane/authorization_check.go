package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func normalizeAuthorizationCheckRequest(input AuthorizationCheckRequest, project Project) (AuthorizationCheckRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return input, fmt.Errorf("name is required")
	}
	if len(input.Name) > 160 {
		return input, fmt.Errorf("name must be 160 characters or fewer")
	}
	input.Description = strings.TrimSpace(input.Description)
	if len(input.Description) > 1000 {
		return input, fmt.Errorf("description must be 1000 characters or fewer")
	}
	input.ResourceLabel = strings.TrimSpace(input.ResourceLabel)
	if len(input.ResourceLabel) > 160 {
		return input, fmt.Errorf("resource_label must be 160 characters or fewer")
	}

	input.Type = strings.TrimSpace(input.Type)
	if input.Type == "" {
		input.Type = AuthorizationCheckTypeBrowserURL
	}
	if input.Type != AuthorizationCheckTypeBrowserURL && input.Type != AuthorizationCheckTypeAPIGet {
		return input, fmt.Errorf("type must be browser_url or api_get")
	}

	input.ActorCredentialProfileID = strings.TrimSpace(input.ActorCredentialProfileID)
	if input.ActorCredentialProfileID == "" {
		return input, fmt.Errorf("actor_credential_profile_id is required")
	}
	if _, err := uuid.Parse(input.ActorCredentialProfileID); err != nil {
		return input, fmt.Errorf("actor_credential_profile_id is invalid")
	}
	input.OwnerCredentialProfileID = strings.TrimSpace(input.OwnerCredentialProfileID)
	if input.OwnerCredentialProfileID != "" {
		if _, err := uuid.Parse(input.OwnerCredentialProfileID); err != nil {
			return input, fmt.Errorf("owner_credential_profile_id is invalid")
		}
	}

	input.ExpectedOutcome = strings.TrimSpace(strings.ToLower(input.ExpectedOutcome))
	if input.ExpectedOutcome != AuthorizationExpectedAllowed && input.ExpectedOutcome != AuthorizationExpectedDenied {
		return input, fmt.Errorf("expected_outcome must be allowed or denied")
	}

	input.SuccessTextContains = strings.TrimSpace(input.SuccessTextContains)
	if len(input.SuccessTextContains) > 500 {
		return input, fmt.Errorf("success_text_contains must be 500 characters or fewer")
	}
	input.DeniedTextContains = strings.TrimSpace(input.DeniedTextContains)
	if len(input.DeniedTextContains) > 500 {
		return input, fmt.Errorf("denied_text_contains must be 500 characters or fewer")
	}
	input.ExpectedStatuses = normalizeStatusList(input.ExpectedStatuses)
	input.DeniedStatuses = normalizeStatusList(input.DeniedStatuses)

	if input.Type == AuthorizationCheckTypeBrowserURL {
		targetURL, err := normalizeAuthorizationBrowserTarget(input.TargetURL, project)
		if err != nil {
			return input, fmt.Errorf("target_url: %w", err)
		}
		input.TargetURL = targetURL
		input.Method = ""
		input.Path = ""
		input.APISpecID = ""
		input.APIOperationID = ""
		return input, nil
	}

	input.Method = strings.ToUpper(strings.TrimSpace(input.Method))
	if input.Method == "" {
		input.Method = "GET"
	}
	if input.Method != "GET" {
		return input, fmt.Errorf("api_get authorization checks support GET only")
	}
	input.Path = strings.TrimSpace(input.Path)
	if input.Path == "" {
		input.Path = strings.TrimSpace(input.TargetURL)
	}
	if input.Path == "" {
		return input, fmt.Errorf("path or target_url is required for api_get")
	}
	if hasSensitiveTargetPathQuery(input.Path) {
		return input, fmt.Errorf("path query contains sensitive parameter names")
	}
	return input, nil
}

func normalizeAuthorizationBrowserTarget(raw string, project Project) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("target URL or path is required")
	}
	if project.FrontendURL == "" {
		return "", fmt.Errorf("project frontend_url is required for browser_url checks")
	}
	if strings.HasPrefix(value, "//") {
		return "", fmt.Errorf("target must stay on the project frontend origin")
	}

	root, err := url.Parse(project.FrontendURL)
	if err != nil || root.Scheme == "" || root.Host == "" {
		return "", fmt.Errorf("project frontend_url is invalid")
	}
	target, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("target is invalid")
	}
	if !target.IsAbs() {
		target = root.ResolveReference(target)
	}
	target.Fragment = ""
	if target.Scheme != "http" && target.Scheme != "https" {
		return "", fmt.Errorf("target must use http or https")
	}
	if target.Scheme != root.Scheme || target.Host != root.Host {
		return "", fmt.Errorf("target must stay on the project frontend origin")
	}
	if hasSensitiveTargetQuery(target) {
		return "", fmt.Errorf("target query contains sensitive parameter names")
	}
	checked, err := ValidateTargetURL(target.String(), project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return "", err
	}
	checked.Fragment = ""
	return checked.String(), nil
}

func normalizeAuthorizationRunRequest(input AuthorizationCheckRunRequest) (AuthorizationCheckRunRequest, error) {
	if input.MaxChecks == 0 {
		input.MaxChecks = 10
	}
	if input.MaxChecks < 1 || input.MaxChecks > 50 {
		return input, fmt.Errorf("max_checks must be between 1 and 50")
	}
	seen := map[string]struct{}{}
	checkIDs := make([]string, 0, len(input.CheckIDs))
	for _, id := range input.CheckIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, err := uuid.Parse(id); err != nil {
			return input, fmt.Errorf("check_ids contains an invalid id")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		checkIDs = append(checkIDs, id)
	}
	if len(checkIDs) > input.MaxChecks {
		checkIDs = checkIDs[:input.MaxChecks]
	}
	input.CheckIDs = checkIDs
	return input, nil
}

func authorizationCheckFromRequest(projectID string, input AuthorizationCheckRequest) AuthorizationCheck {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	return AuthorizationCheck{
		ProjectID:                projectID,
		Name:                     input.Name,
		Description:              input.Description,
		Type:                     input.Type,
		ResourceLabel:            input.ResourceLabel,
		OwnerCredentialProfileID: input.OwnerCredentialProfileID,
		ActorCredentialProfileID: input.ActorCredentialProfileID,
		ExpectedOutcome:          input.ExpectedOutcome,
		TargetURL:                input.TargetURL,
		APISpecID:                input.APISpecID,
		APIOperationID:           input.APIOperationID,
		Method:                   input.Method,
		Path:                     input.Path,
		ExpectedStatuses:         input.ExpectedStatuses,
		SuccessTextContains:      input.SuccessTextContains,
		DeniedStatuses:           input.DeniedStatuses,
		DeniedTextContains:       input.DeniedTextContains,
		Enabled:                  enabled,
	}
}

func normalizeStatusList(values []int) []int {
	seen := map[int]struct{}{}
	statuses := make([]int, 0, len(values))
	for _, value := range values {
		if value < 100 || value > 599 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		statuses = append(statuses, value)
	}
	return statuses
}

func mustMarshalJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte("[]")
	}
	return raw
}

func hasSensitiveTargetQuery(target *url.URL) bool {
	for name := range target.Query() {
		normalized := strings.ToLower(name)
		for _, marker := range []string{"authorization", "password", "passwd", "token", "secret", "api_key", "apikey", "cookie", "session", "jwt"} {
			if strings.Contains(normalized, marker) {
				return true
			}
		}
	}
	return false
}

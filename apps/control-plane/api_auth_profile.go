package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	APIAuthProfileTypeBearerToken  = "bearer_token"
	APIAuthProfileTypeAPIKeyHeader = "api_key_header"
	APIAuthProfileTypeAPIKeyQuery  = "api_key_query"
	APIAuthProfileTypeBasicAuth    = "basic_auth"
	APIAuthProfileTypeNone         = "none"
)

var apiHeaderNamePattern = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+.^_` + "`" + `|~-]+$`)

type apiAuthMaterial struct {
	ProfileID      string
	ProfileName    string
	Type           string
	HeaderName     string
	QueryParamName string
	Username       string
	Password       string
	Token          string
	APIKey         string
	DisplayHint    string
}

func normalizeAPIAuthProfileRequest(input APIAuthProfileRequest, create bool) (APIAuthProfileRequest, error) {
	input.Name = strings.TrimSpace(RedactSecrets(input.Name))
	if input.Name == "" {
		return input, fmt.Errorf("name is required")
	}
	if len(input.Name) > 120 {
		return input, fmt.Errorf("name must be 120 characters or fewer")
	}
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	if input.Type == "" {
		input.Type = APIAuthProfileTypeNone
	}
	switch input.Type {
	case APIAuthProfileTypeBearerToken, APIAuthProfileTypeAPIKeyHeader, APIAuthProfileTypeAPIKeyQuery, APIAuthProfileTypeBasicAuth, APIAuthProfileTypeNone:
	default:
		return input, fmt.Errorf("type must be bearer_token, api_key_header, api_key_query, basic_auth, or none")
	}
	input.HeaderName = strings.TrimSpace(input.HeaderName)
	input.QueryParamName = strings.TrimSpace(input.QueryParamName)
	input.Username = strings.TrimSpace(input.Username)

	if input.Enabled == nil {
		enabled := true
		input.Enabled = &enabled
	}
	switch input.Type {
	case APIAuthProfileTypeBearerToken:
		input.HeaderName = "Authorization"
		if create && strings.TrimSpace(input.Token) == "" {
			return input, fmt.Errorf("token is required for bearer_token profiles")
		}
	case APIAuthProfileTypeAPIKeyHeader:
		if input.HeaderName == "" {
			return input, fmt.Errorf("header_name is required for api_key_header profiles")
		}
		if err := validateAPIHeaderName(input.HeaderName); err != nil {
			return input, err
		}
		if create && strings.TrimSpace(input.APIKey) == "" {
			return input, fmt.Errorf("api_key is required for api_key_header profiles")
		}
	case APIAuthProfileTypeAPIKeyQuery:
		if input.QueryParamName == "" {
			return input, fmt.Errorf("query_param_name is required for api_key_query profiles")
		}
		if err := validateAPIQueryParamName(input.QueryParamName); err != nil {
			return input, err
		}
		if create && strings.TrimSpace(input.APIKey) == "" {
			return input, fmt.Errorf("api_key is required for api_key_query profiles")
		}
	case APIAuthProfileTypeBasicAuth:
		if create && input.Username == "" {
			return input, fmt.Errorf("username is required for basic_auth profiles")
		}
		if create && input.Password == "" {
			return input, fmt.Errorf("password is required for basic_auth profiles")
		}
	case APIAuthProfileTypeNone:
		input.HeaderName = ""
		input.QueryParamName = ""
		input.Username = ""
		input.Password = ""
		input.Token = ""
		input.APIKey = ""
	}
	return input, nil
}

func validateAPIHeaderName(name string) error {
	if name == "" || len(name) > 120 || !apiHeaderNamePattern.MatchString(name) {
		return fmt.Errorf("header_name must be a valid HTTP header name")
	}
	if strings.EqualFold(name, "Cookie") || strings.EqualFold(name, "Set-Cookie") {
		return fmt.Errorf("cookie headers are not supported for API auth profiles")
	}
	return nil
}

func validateAPIQueryParamName(name string) error {
	if name == "" || len(name) > 120 || strings.ContainsAny(name, "=&?#/\\\r\n") {
		return fmt.Errorf("query_param_name must be a simple query parameter name")
	}
	return nil
}

func (a *App) apiAuthProfileFromInput(input APIAuthProfileRequest, current *APIAuthProfile) (APIAuthProfile, error) {
	profile := APIAuthProfile{
		Name:           input.Name,
		Type:           input.Type,
		HeaderName:     input.HeaderName,
		QueryParamName: input.QueryParamName,
		Enabled:        input.Enabled == nil || *input.Enabled,
	}
	if current != nil {
		profile.UsernameEncrypted = current.UsernameEncrypted
		profile.PasswordEncrypted = current.PasswordEncrypted
		profile.TokenEncrypted = current.TokenEncrypted
		profile.APIKeyEncrypted = current.APIKeyEncrypted
		profile.UsernameDisplayHint = current.UsernameDisplayHint
		profile.TokenDisplayHint = current.TokenDisplayHint
		profile.APIKeyDisplayHint = current.APIKeyDisplayHint
	}
	if input.Username != "" {
		encrypted, err := a.secretBox.Encrypt(input.Username)
		if err != nil {
			return APIAuthProfile{}, err
		}
		profile.UsernameEncrypted = encrypted
		profile.UsernameDisplayHint = usernameDisplayHint(input.Username)
	}
	if input.Password != "" {
		encrypted, err := a.secretBox.Encrypt(input.Password)
		if err != nil {
			return APIAuthProfile{}, err
		}
		profile.PasswordEncrypted = encrypted
	}
	if strings.TrimSpace(input.Token) != "" {
		encrypted, err := a.secretBox.Encrypt(strings.TrimSpace(input.Token))
		if err != nil {
			return APIAuthProfile{}, err
		}
		profile.TokenEncrypted = encrypted
		profile.TokenDisplayHint = secretDisplayHint(input.Token)
	}
	if strings.TrimSpace(input.APIKey) != "" {
		encrypted, err := a.secretBox.Encrypt(strings.TrimSpace(input.APIKey))
		if err != nil {
			return APIAuthProfile{}, err
		}
		profile.APIKeyEncrypted = encrypted
		profile.APIKeyDisplayHint = secretDisplayHint(input.APIKey)
	}
	if input.Type == APIAuthProfileTypeNone {
		profile.UsernameEncrypted = ""
		profile.PasswordEncrypted = ""
		profile.TokenEncrypted = ""
		profile.APIKeyEncrypted = ""
		profile.UsernameDisplayHint = ""
		profile.TokenDisplayHint = ""
		profile.APIKeyDisplayHint = ""
	}
	profile.setConfiguredFlags()
	return profile, nil
}

func secretDisplayHint(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	if len(secret) <= 4 {
		return "configured"
	}
	return "..." + secret[len(secret)-4:]
}

func (profile *APIAuthProfile) setConfiguredFlags() {
	profile.UsernameConfigured = profile.UsernameEncrypted != ""
	profile.PasswordConfigured = profile.PasswordEncrypted != ""
	profile.TokenConfigured = profile.TokenEncrypted != ""
	profile.APIKeyConfigured = profile.APIKeyEncrypted != ""
}

func (a *App) apiAuthMaterial(ctx context.Context, profile *APIAuthProfile) (*apiAuthMaterial, error) {
	if profile == nil || profile.Type == APIAuthProfileTypeNone {
		return &apiAuthMaterial{Type: APIAuthProfileTypeNone, DisplayHint: "none"}, nil
	}
	if !profile.Enabled {
		return nil, fmt.Errorf("API auth profile is disabled")
	}
	material := &apiAuthMaterial{
		ProfileID:      profile.ID,
		ProfileName:    profile.Name,
		Type:           profile.Type,
		HeaderName:     profile.HeaderName,
		QueryParamName: profile.QueryParamName,
		DisplayHint:    firstNonEmpty(profile.TokenDisplayHint, profile.APIKeyDisplayHint, profile.UsernameDisplayHint, "configured"),
	}
	var err error
	switch profile.Type {
	case APIAuthProfileTypeBearerToken:
		material.Token, err = a.secretBox.Decrypt(profile.TokenEncrypted)
	case APIAuthProfileTypeAPIKeyHeader:
		material.APIKey, err = a.secretBox.Decrypt(profile.APIKeyEncrypted)
	case APIAuthProfileTypeAPIKeyQuery:
		material.APIKey, err = a.secretBox.Decrypt(profile.APIKeyEncrypted)
	case APIAuthProfileTypeBasicAuth:
		material.Username, err = a.secretBox.Decrypt(profile.UsernameEncrypted)
		if err == nil {
			material.Password, err = a.secretBox.Decrypt(profile.PasswordEncrypted)
		}
	default:
		return nil, fmt.Errorf("unsupported API auth profile type")
	}
	if err != nil {
		return nil, fmt.Errorf("decrypt API auth profile: %w", err)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return material, nil
}

func (material *apiAuthMaterial) authMode() string {
	if material == nil || material.Type == "" {
		return APIAuthProfileTypeNone
	}
	return material.Type
}

func (material *apiAuthMaterial) apply(req *http.Request) {
	if material == nil {
		return
	}
	switch material.Type {
	case APIAuthProfileTypeBearerToken:
		req.Header.Set("Authorization", "Bearer "+material.Token)
	case APIAuthProfileTypeAPIKeyHeader:
		req.Header.Set(material.HeaderName, material.APIKey)
	case APIAuthProfileTypeAPIKeyQuery:
		query := req.URL.Query()
		query.Set(material.QueryParamName, material.APIKey)
		req.URL.RawQuery = query.Encode()
	case APIAuthProfileTypeBasicAuth:
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(material.Username+":"+material.Password)))
	}
}

func (material *apiAuthMaterial) redactedHeaders() map[string]string {
	if material == nil {
		return nil
	}
	switch material.Type {
	case APIAuthProfileTypeBearerToken, APIAuthProfileTypeBasicAuth:
		return map[string]string{"Authorization": "[REDACTED]"}
	case APIAuthProfileTypeAPIKeyHeader:
		return map[string]string{material.HeaderName: "[REDACTED]"}
	default:
		return nil
	}
}

func (material *apiAuthMaterial) summary() APIAuthSummary {
	if material == nil || material.Type == APIAuthProfileTypeNone {
		return APIAuthSummary{
			AuthMode:          APIAuthProfileTypeNone,
			Authenticated:     false,
			SecretsStored:     "not_applicable",
			SecretsReturned:   false,
			SecretsSentToAI:   false,
			AuthHeadersStored: false,
		}
	}
	return APIAuthSummary{
		ProfileID:         material.ProfileID,
		ProfileName:       material.ProfileName,
		Type:              material.Type,
		AuthMode:          material.Type,
		DisplayHint:       material.DisplayHint,
		Authenticated:     true,
		SecretsStored:     "encrypted",
		SecretsReturned:   false,
		SecretsSentToAI:   false,
		AuthHeadersStored: false,
	}
}

func normalizeAPIAuthProfileTestRequest(input APIAuthProfileTestRequest) (APIAuthProfileTestRequest, error) {
	input.Method = strings.ToUpper(strings.TrimSpace(input.Method))
	if input.Method == "" {
		input.Method = http.MethodGet
	}
	if input.Method != http.MethodGet && input.Method != http.MethodHead {
		return input, fmt.Errorf("method must be GET or HEAD")
	}
	input.TestPath = strings.TrimSpace(input.TestPath)
	if input.TestPath != "" {
		if strings.ContainsAny(input.TestPath, "\r\n") {
			return input, fmt.Errorf("test_path is invalid")
		}
	}
	return input, nil
}

func apiAuthTestURL(project Project, input APIAuthProfileTestRequest) (string, error) {
	base := project.APIBaseURL
	if base == "" {
		if project.OpenAPIURL == "" {
			return "", fmt.Errorf("project API base URL or OpenAPI URL is required")
		}
		parsed, err := url.Parse(project.OpenAPIURL)
		if err != nil {
			return "", fmt.Errorf("project OpenAPI URL is invalid")
		}
		parsed.Path = ""
		parsed.RawQuery = ""
		parsed.Fragment = ""
		base = parsed.String()
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("project API base URL is invalid")
	}
	if input.TestPath == "" {
		return baseURL.String(), nil
	}
	pathURL, err := url.Parse(input.TestPath)
	if err != nil {
		return "", fmt.Errorf("test_path is invalid")
	}
	if pathURL.IsAbs() {
		if !sameAPIOrigin(baseURL, pathURL) {
			return "", fmt.Errorf("test_path must stay on the API base origin")
		}
		return pathURL.String(), nil
	}
	next := *baseURL
	next.Path = strings.TrimRight(next.Path, "/") + "/" + strings.TrimLeft(pathURL.Path, "/")
	next.RawQuery = pathURL.RawQuery
	next.Fragment = ""
	return next.String(), nil
}

func (a *App) testAPIAuthProfile(ctx context.Context, project Project, profile APIAuthProfile, input APIAuthProfileTestRequest) APIAuthProfileTestResult {
	started := time.Now()
	result := APIAuthProfileTestResult{
		ProfileID:   profile.ID,
		ProfileName: profile.Name,
		AuthMode:    profile.Type,
		Method:      input.Method,
	}
	target, err := apiAuthTestURL(project, input)
	if err != nil {
		result.ErrorMessage = RedactSecrets(err.Error())
		return result
	}
	parsed, err := ValidateTargetURL(target, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		result.ErrorMessage = RedactSecrets(err.Error())
		return result
	}
	material, err := a.apiAuthMaterial(ctx, &profile)
	if err != nil {
		result.ErrorMessage = RedactSecrets(err.Error())
		return result
	}
	status, contentType, _, _, err := requestSafeAPIOperationWithAuth(ctx, input.Method, parsed, material)
	result.DurationMS = int(time.Since(started).Milliseconds())
	result.URL = sanitizeURLForStorageWithSensitive(parsed.String(), []string{profile.QueryParamName})
	result.RedactedHeaders = material.redactedHeaders()
	if status != nil {
		result.HTTPStatus = status
	}
	result.ResponseContentType = contentType
	if err != nil {
		result.ErrorMessage = RedactSecrets(err.Error())
		return result
	}
	result.Success = status != nil && *status < 500
	return result
}

func sanitizeURLForStorageWithSensitive(raw string, sensitiveNames []string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return RedactSecrets(raw)
	}
	parsed.User = nil
	if parsed.RawQuery != "" {
		values := parsed.Query()
		for key := range values {
			if sensitiveAPIParameterName(key) || containsFold(sensitiveNames, key) {
				values.Set(key, "[REDACTED]")
			}
		}
		parsed.RawQuery = values.Encode()
	}
	return parsed.String()
}

func containsFold(items []string, value string) bool {
	for _, item := range items {
		if item != "" && strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

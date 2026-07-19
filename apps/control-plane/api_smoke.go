package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiSmokeRequestTimeout = 5 * time.Second
const maxAPIResponseReadBytes = 1024 * 1024
const defaultAPISmokeMaxOperations = 50

type APISmokeExecutionOptions struct {
	APIAuthProfile                   *APIAuthProfile
	AuthMaterial                     *apiAuthMaterial
	Authenticated                    bool
	ValidateContract                 bool
	ValidateSchema                   bool
	MaxOperations                    int
	IncludeUnauthenticatedComparison bool
}

func NormalizeAPISpecImportRequest(input APISpecImportRequest) (APISpecImportRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.SourceType = strings.ToLower(strings.TrimSpace(input.SourceType))
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	input.RawSpec = strings.TrimSpace(input.RawSpec)

	if input.Name == "" {
		input.Name = "OpenAPI Spec"
	}
	switch input.SourceType {
	case "url":
		if input.SourceURL == "" {
			return input, fmt.Errorf("source_url is required for URL imports")
		}
		input.RawSpec = ""
	case "inline", "demo":
		if input.RawSpec == "" {
			return input, fmt.Errorf("raw_spec is required for inline imports")
		}
	case "":
		return input, fmt.Errorf("source_type is required")
	default:
		return input, fmt.Errorf("source_type must be url, inline, or demo")
	}
	if len([]byte(input.RawSpec)) > maxStoredSpecBytes {
		return input, fmt.Errorf("raw_spec is too large for the v0.20 alpha import limit")
	}
	return input, nil
}

func FetchOpenAPISource(ctx context.Context, project Project, sourceURL string) (string, string, error) {
	parsed, err := ValidateTargetURL(sourceURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		return "", "", err
	}
	client := &http.Client{
		Timeout: apiSmokeRequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", parsed.String(), fmt.Errorf("build OpenAPI request: %w", err)
	}
	req.Header.Set("User-Agent", "Qualora OpenAPI Import v0.20.0-alpha")
	req.Header.Set("Accept", "application/json, application/yaml, text/yaml, application/x-yaml, text/plain, */*")

	resp, err := client.Do(req)
	if err != nil {
		return "", parsed.String(), fmt.Errorf("fetch OpenAPI document: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
		return "", parsed.String(), fmt.Errorf("OpenAPI URL returned a redirect that was not followed")
	}
	if resp.StatusCode >= 400 {
		return "", parsed.String(), fmt.Errorf("OpenAPI URL returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxStoredSpecBytes+1))
	if err != nil {
		return "", parsed.String(), fmt.Errorf("read OpenAPI document: %w", err)
	}
	if len(body) > maxStoredSpecBytes {
		return "", parsed.String(), fmt.Errorf("OpenAPI document is too large for the v0.20 alpha import limit")
	}
	return string(body), parsed.String(), nil
}

func NormalizeAPISmokeRunRequest(input APISmokeRunRequest) (APISmokeRunRequest, error) {
	input.APIAuthProfileID = strings.TrimSpace(input.APIAuthProfileID)
	if input.ValidateContract == nil {
		value := true
		input.ValidateContract = &value
	}
	if input.ValidateSchema == nil {
		value := true
		input.ValidateSchema = &value
	}
	if input.IncludeUnauthenticatedComparison == nil {
		value := false
		input.IncludeUnauthenticatedComparison = &value
	}
	if input.Authenticated == nil {
		value := input.APIAuthProfileID != ""
		input.Authenticated = &value
	}
	if *input.Authenticated && input.APIAuthProfileID == "" {
		return input, fmt.Errorf("api_auth_profile_id is required when authenticated=true")
	}
	if input.MaxOperations == 0 {
		input.MaxOperations = defaultAPISmokeMaxOperations
	}
	if input.MaxOperations < 1 || input.MaxOperations > 200 {
		return input, fmt.Errorf("max_operations must be between 1 and 200")
	}
	return input, nil
}

func (a *App) ExecuteAPISmokeRun(ctx context.Context, project Project, spec APISpec, operations []APIOperation, options APISmokeExecutionOptions) (*TestRun, error) {
	options = normalizeAPISmokeExecutionOptions(options)
	apiAuthProfileID := ""
	if options.APIAuthProfile != nil {
		apiAuthProfileID = options.APIAuthProfile.ID
	}
	run, err := a.store.CreateAPISmokeRunRecord(ctx, project.ID, spec.ID, apiAuthProfileID)
	if err != nil {
		return nil, err
	}

	baseURL, err := apiSmokeBaseURL(project, spec)
	if err != nil {
		_ = a.store.CompleteAPISmokeRun(ctx, run.ID, StatusFailed, err.Error())
		return run, err
	}
	baseParsed, err := ValidateTargetURL(baseURL, project.AllowedHosts, project.AllowPrivateTargets)
	if err != nil {
		_ = a.store.CompleteAPISmokeRun(ctx, run.ID, StatusFailed, err.Error())
		return run, err
	}

	results := make([]APICheckResult, 0, len(operations))
	for index, operation := range operations {
		if index >= options.MaxOperations {
			result, findings := skippedAPIOperationResult(run.ID, spec.ID, operation, "max_operations limit reached")
			storedResult, err := a.store.InsertAPICheckResult(ctx, result)
			if err != nil {
				_ = a.store.CompleteAPISmokeRun(ctx, run.ID, StatusFailed, "API result could not be recorded")
				return run, err
			}
			results = append(results, *storedResult)
			for _, finding := range findings {
				if err := a.store.InsertRunFinding(ctx, run.ID, finding); err != nil {
					_ = a.store.CompleteAPISmokeRun(ctx, run.ID, StatusFailed, "API finding could not be recorded")
					return run, err
				}
			}
			continue
		}
		result, findings, evidence := a.executeAPIOperation(ctx, project, spec, run.ID, baseParsed, operation, options)
		if evidence.Type != "" {
			evidenceID, err := a.store.InsertRunEvidence(ctx, run.ID, evidence)
			if err != nil {
				_ = a.store.CompleteAPISmokeRun(ctx, run.ID, StatusFailed, "API evidence could not be recorded")
				return run, err
			}
			for index := range findings {
				findings[index].EvidenceIDs = append(findings[index].EvidenceIDs, evidenceID)
			}
		}
		storedResult, err := a.store.InsertAPICheckResult(ctx, result)
		if err != nil {
			_ = a.store.CompleteAPISmokeRun(ctx, run.ID, StatusFailed, "API result could not be recorded")
			return run, err
		}
		results = append(results, *storedResult)
		for _, finding := range findings {
			if err := a.store.InsertRunFinding(ctx, run.ID, finding); err != nil {
				_ = a.store.CompleteAPISmokeRun(ctx, run.ID, StatusFailed, "API finding could not be recorded")
				return run, err
			}
		}
	}

	summary := summarizeAPICheckResults(results)
	_, _ = a.store.InsertRunEvidence(ctx, run.ID, Evidence{
		Type: "api_observations",
		URI:  "inline://api-observations",
		Metadata: map[string]any{
			"api_base_url":        baseParsed.String(),
			"api_spec_id":         spec.ID,
			"api_spec_name":       spec.Name,
			"checked_endpoints":   summary.ExecutedOperations,
			"failed_endpoints":    summary.FailedOperations + summary.ErroredOperations,
			"skipped_endpoints":   summary.SkippedOperations,
			"safe_methods_only":   true,
			"response_bodies":     "not_stored",
			"authenticated_tests": options.Authenticated,
			"auth_mode":           options.AuthMaterial.authMode(),
			"auth_profile_id":     apiAuthProfileID,
			"auth_profile_name":   apiAuthProfileName(options.APIAuthProfile),
			"contract_validation": map[string]any{
				"enabled":                  options.ValidateContract,
				"passed":                   summary.ContractPassed,
				"failed":                   summary.ContractFailed,
				"skipped":                  summary.ContractSkipped,
				"unknown":                  summary.ContractUnknown,
				"schema_validation_errors": summary.SchemaValidationErrorCount,
			},
			"unauthenticated_comparisons": summary.UnauthenticatedComparisons,
		},
	})
	_, _ = a.store.InsertRunEvidence(ctx, run.ID, Evidence{
		Type: "openapi_summary",
		URI:  "inline://openapi-summary",
		Metadata: map[string]any{
			"api_spec_id":                   spec.ID,
			"name":                          spec.Name,
			"title":                         spec.ParsedTitle,
			"version":                       spec.ParsedVersion,
			"server_url":                    spec.ServerURL,
			"operations":                    spec.OperationCount,
			"safe_operations":               spec.SafeOperationCount,
			"skipped_unsafe_operations":     spec.SkippedOperationCount,
			"safe_methods_only":             true,
			"authenticated_safe_methods":    options.Authenticated,
			"contract_validation":           options.ValidateContract,
			"schema_validation":             options.ValidateSchema,
			"request_response_bodies_saved": false,
		},
	})

	if err := a.store.CompleteAPISmokeRun(ctx, run.ID, StatusCompleted, ""); err != nil {
		return run, err
	}
	return a.store.GetRun(ctx, run.ID)
}

func normalizeAPISmokeExecutionOptions(options APISmokeExecutionOptions) APISmokeExecutionOptions {
	if options.MaxOperations <= 0 {
		options.MaxOperations = defaultAPISmokeMaxOperations
	}
	if options.AuthMaterial == nil {
		options.AuthMaterial = &apiAuthMaterial{Type: APIAuthProfileTypeNone, DisplayHint: "none"}
	}
	if options.APIAuthProfile != nil {
		options.Authenticated = true
	}
	if !options.ValidateContract && !options.ValidateSchema {
		options.ValidateContract = true
	}
	return options
}

func (a *App) executeAPIOperation(ctx context.Context, project Project, spec APISpec, runID string, baseURL *url.URL, operation APIOperation, options APISmokeExecutionOptions) (APICheckResult, []Finding, Evidence) {
	result := APICheckResult{
		RunID:                    runID,
		APISpecID:                spec.ID,
		OperationID:              operation.ID,
		APIAuthProfileID:         apiAuthProfileID(options.APIAuthProfile),
		AuthMode:                 options.AuthMaterial.authMode(),
		Method:                   operation.Method,
		Path:                     operation.Path,
		ExpectedStatuses:         operation.ExpectedStatuses,
		ExpectedContentTypes:     operation.ExpectedContentTypes,
		ContractValidationStatus: "unknown",
	}
	executable, skipReason := apiOperationExecutable(operation, options.Authenticated)
	if !executable {
		result.Status = StatusSkipped
		result.SkippedReason = firstNonEmpty(skipReason, operation.SkipReason)
		result.ContractValidationStatus = "skipped"
		findings := apiOperationSkippedFindings(operation, result)
		return result, findings, Evidence{}
	}

	targetURL, err := resolveAPIOperationURL(baseURL, operation)
	if err != nil {
		result.Status = StatusSkipped
		result.SkippedReason = err.Error()
		result.ContractValidationStatus = "skipped"
		return result, nil, Evidence{}
	}
	result.ResolvedURL = sanitizeURLForStorageWithSensitive(targetURL.String(), []string{apiAuthQueryName(options.AuthMaterial)})

	if !sameAPIOrigin(baseURL, targetURL) {
		result.Status = StatusSkipped
		result.SkippedReason = "resolved operation URL is outside the API base origin"
		result.ContractValidationStatus = "skipped"
		return result, nil, Evidence{}
	}
	if _, err := ValidateTargetURL(targetURL.String(), project.AllowedHosts, project.AllowPrivateTargets); err != nil {
		result.Status = StatusSkipped
		result.SkippedReason = err.Error()
		result.ContractValidationStatus = "skipped"
		return result, nil, Evidence{}
	}

	started := time.Now()
	var unauthenticatedStatus *int
	if options.IncludeUnauthenticatedComparison && options.Authenticated {
		unauthenticatedStatus, _, _, _, _ = requestSafeAPIOperationWithAuth(ctx, operation.Method, targetURL, nil)
	}
	statusCode, contentType, body, redirectBlocked, requestErr := requestSafeAPIOperationWithAuth(ctx, operation.Method, targetURL, options.AuthMaterial)
	duration := int(time.Since(started).Milliseconds())
	result.DurationMS = &duration
	result.ResponseTimeMS = &duration
	if statusCode != nil {
		result.HTTPStatus = statusCode
		result.ActualStatus = statusCode
	}
	result.ResponseContentType = contentType
	result.ActualContentType = contentType
	result.UnauthenticatedStatus = unauthenticatedStatus
	size := len(body)
	result.ResponseSizeBytes = &size
	if requestErr != nil {
		result.Status = StatusError
		result.ErrorMessage = RedactSecrets(requestErr.Error())
	} else {
		result.Status = StatusPassed
	}

	evidence := Evidence{
		Type: "api_request",
		URI:  result.ResolvedURL,
		Metadata: map[string]any{
			"method":                   operation.Method,
			"path":                     operation.Path,
			"resolved_url":             result.ResolvedURL,
			"status":                   result.Status,
			"http_status":              statusCode,
			"duration_ms":              duration,
			"response_content_type":    contentType,
			"response_size_bytes":      size,
			"expected_statuses":        operation.ExpectedStatuses,
			"expected_content_types":   operation.ExpectedContentTypes,
			"safe_methods_only":        true,
			"authenticated":            options.Authenticated,
			"auth_mode":                options.AuthMaterial.authMode(),
			"auth_profile_id":          apiAuthProfileID(options.APIAuthProfile),
			"auth_profile_name":        apiAuthProfileName(options.APIAuthProfile),
			"contract_validation":      options.ValidateContract,
			"schema_validation":        options.ValidateSchema,
			"contract_status":          result.ContractValidationStatus,
			"schema_validation_errors": result.SchemaValidationErrors,
			"unauthenticated_status":   unauthenticatedStatus,
			"response_body_stored":     false,
		},
	}
	if result.ErrorMessage != "" {
		evidence.Metadata["error"] = result.ErrorMessage
	}

	findings := buildAPIOperationFindingsWithOptions(operation, result, body, redirectBlocked, options)
	result.ContractValidationStatus = apiContractStatus(result, findings, options)
	result.SchemaValidationErrors = schemaValidationErrors(findings)
	evidence.Metadata["contract_status"] = result.ContractValidationStatus
	evidence.Metadata["schema_validation_errors"] = result.SchemaValidationErrors
	if len(findings) > 0 && result.Status == StatusPassed {
		result.Status = StatusFailed
		evidence.Metadata["status"] = result.Status
	}
	return result, findings, evidence
}

func requestSafeAPIOperation(ctx context.Context, method string, targetURL *url.URL) (*int, string, []byte, bool, error) {
	return requestSafeAPIOperationWithAuth(ctx, method, targetURL, nil)
}

func requestSafeAPIOperationWithAuth(ctx context.Context, method string, targetURL *url.URL, material *apiAuthMaterial) (*int, string, []byte, bool, error) {
	requestCtx, cancel := context.WithTimeout(ctx, apiSmokeRequestTimeout)
	defer cancel()

	client := &http.Client{
		Timeout: apiSmokeRequestTimeout + time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	requestURL := *targetURL
	req, err := http.NewRequestWithContext(requestCtx, method, requestURL.String(), nil)
	if err != nil {
		return nil, "", nil, false, err
	}
	req.Header.Set("User-Agent", "Qualora API Smoke v0.20.0-alpha")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	if material != nil {
		material.apply(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", nil, false, err
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	redirectBlocked := false
	if status >= 300 && status <= 399 {
		if location := resp.Header.Get("Location"); location != "" {
			nextURL, err := url.Parse(location)
			if err == nil {
				nextURL = targetURL.ResolveReference(nextURL)
				redirectBlocked = !sameAPIOrigin(targetURL, nextURL)
			}
		}
	}
	if method == http.MethodHead {
		return &status, contentType, nil, redirectBlocked, nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseReadBytes))
	if err != nil {
		return &status, contentType, nil, redirectBlocked, err
	}
	return &status, contentType, body, redirectBlocked, nil
}

func apiOperationExecutable(operation APIOperation, authenticated bool) (bool, string) {
	if operation.SafeToExecute {
		return true, ""
	}
	if authenticated && operation.RequiresAuthentication != nil && *operation.RequiresAuthentication && operation.SkipReason == "operation declares authentication requirements" && operation.ResolvedPath != "" {
		return true, ""
	}
	return false, operation.SkipReason
}

func skippedAPIOperationResult(runID string, apiSpecID string, operation APIOperation, reason string) (APICheckResult, []Finding) {
	result := APICheckResult{
		RunID:                    runID,
		APISpecID:                apiSpecID,
		OperationID:              operation.ID,
		Method:                   operation.Method,
		Path:                     operation.Path,
		Status:                   StatusSkipped,
		SkippedReason:            reason,
		AuthMode:                 APIAuthProfileTypeNone,
		ExpectedStatuses:         operation.ExpectedStatuses,
		ExpectedContentTypes:     operation.ExpectedContentTypes,
		ContractValidationStatus: "skipped",
	}
	return result, apiOperationSkippedFindings(operation, result)
}

func apiOperationSkippedFindings(operation APIOperation, result APICheckResult) []Finding {
	if result.SkippedReason == "" {
		return nil
	}
	return []Finding{
		{
			Title:          "API operation skipped",
			Severity:       "info",
			Category:       "api_operation_skipped",
			Confidence:     "high",
			Description:    fmt.Sprintf("%s %s was skipped: %s.", operation.Method, operation.Path, result.SkippedReason),
			Recommendation: "Review the OpenAPI operation metadata and Qualora's safe read-only policy before enabling broader API coverage.",
		},
	}
}

func apiAuthProfileID(profile *APIAuthProfile) string {
	if profile == nil {
		return ""
	}
	return profile.ID
}

func apiAuthProfileName(profile *APIAuthProfile) string {
	if profile == nil {
		return ""
	}
	return profile.Name
}

func apiAuthQueryName(material *apiAuthMaterial) string {
	if material == nil {
		return ""
	}
	return material.QueryParamName
}

func buildAPIOperationFindings(operation APIOperation, result APICheckResult, body []byte, redirectBlocked bool) []Finding {
	return buildAPIOperationFindingsWithOptions(operation, result, body, redirectBlocked, APISmokeExecutionOptions{
		ValidateContract: true,
		ValidateSchema:   true,
		AuthMaterial:     &apiAuthMaterial{Type: APIAuthProfileTypeNone},
	})
}

func buildAPIOperationFindingsWithOptions(operation APIOperation, result APICheckResult, body []byte, redirectBlocked bool, options APISmokeExecutionOptions) []Finding {
	findings := make([]Finding, 0)
	target := result.Method + " " + result.ResolvedURL
	if result.ErrorMessage != "" {
		findings = append(findings, Finding{
			Title:          "API endpoint unreachable",
			Severity:       "high",
			Category:       "api",
			Confidence:     "high",
			Description:    fmt.Sprintf("%s could not be reached: %s", target, result.ErrorMessage),
			Recommendation: "Verify DNS, TLS, networking, service availability, and the configured API base URL.",
		})
		return findings
	}
	if options.Authenticated && operation.RequiresAuthentication != nil && *operation.RequiresAuthentication && result.HTTPStatus != nil && (*result.HTTPStatus == http.StatusUnauthorized || *result.HTTPStatus == http.StatusForbidden) {
		findings = append(findings, Finding{
			Title:          "API authentication failed",
			Severity:       "high",
			Category:       "api_auth_failure",
			Confidence:     "high",
			Description:    fmt.Sprintf("%s returned HTTP %d with the selected API auth profile.", target, *result.HTTPStatus),
			Recommendation: "Verify the API auth profile secret, scheme, and OpenAPI security requirements. Auth headers and tokens are not stored in Qualora reports.",
		})
	}
	if redirectBlocked {
		findings = append(findings, Finding{
			Title:          "API endpoint redirected outside the allowed origin",
			Severity:       "medium",
			Category:       "api",
			Confidence:     "high",
			Description:    fmt.Sprintf("%s returned an external redirect that Qualora did not follow.", target),
			Recommendation: "Verify the endpoint redirect behavior and keep API smoke targets inside the allowed API origin.",
		})
	}
	if result.HTTPStatus != nil && *result.HTTPStatus >= 500 {
		findings = append(findings, Finding{
			Title:          "API contract unexpected error",
			Severity:       "high",
			Category:       "api_contract_unexpected_error",
			Confidence:     "high",
			Description:    fmt.Sprintf("%s returned HTTP %d.", target, *result.HTTPStatus),
			Recommendation: "Inspect the API service logs and upstream dependencies for server-side failures.",
		})
	}
	if options.ValidateContract && result.HTTPStatus != nil && *result.HTTPStatus >= 400 && *result.HTTPStatus < 500 && !statusMatchesExpected(*result.HTTPStatus, operation.ExpectedStatuses) {
		findings = append(findings, Finding{
			Title:          "API contract status mismatch",
			Severity:       "medium",
			Category:       "api_contract_status_mismatch",
			Confidence:     "medium",
			Description:    fmt.Sprintf("%s returned HTTP %d, which is not declared for this public safe operation.", target, *result.HTTPStatus),
			Recommendation: "Confirm the endpoint is intentionally public, or update the OpenAPI responses to describe expected 4xx behavior.",
		})
	}
	if options.ValidateContract && result.HTTPStatus != nil && !statusMatchesExpected(*result.HTTPStatus, operation.ExpectedStatuses) {
		findings = append(findings, Finding{
			Title:          "API contract status mismatch",
			Severity:       statusSeverity(*result.HTTPStatus),
			Category:       "api_contract_status_mismatch",
			Confidence:     "medium",
			Description:    fmt.Sprintf("%s returned HTTP %d, which is not declared in the OpenAPI responses.", target, *result.HTTPStatus),
			Recommendation: "Update the OpenAPI document or adjust the endpoint behavior to match the documented responses.",
		})
	}
	if options.ValidateContract && result.ResponseContentType != "" && !contentTypeMatches(result.ResponseContentType, operation.ExpectedContentTypes) {
		findings = append(findings, Finding{
			Title:          "API contract content type mismatch",
			Severity:       "low",
			Category:       "api_contract_content_type_mismatch",
			Confidence:     "medium",
			Description:    fmt.Sprintf("%s returned content type %s, which does not obviously match the OpenAPI response content.", target, result.ResponseContentType),
			Recommendation: "Verify the endpoint content type or update the OpenAPI response content definitions.",
		})
	}
	if options.ValidateContract && len(body) > 0 && isJSONContentType(result.ResponseContentType) && !json.Valid(body) {
		findings = append(findings, Finding{
			Title:          "API contract JSON parse failure",
			Severity:       "medium",
			Category:       "api_contract_json_parse_failure",
			Confidence:     "high",
			Description:    fmt.Sprintf("%s returned a JSON content type but the response body was not valid JSON.", target),
			Recommendation: "Return syntactically valid JSON for responses that declare a JSON content type.",
		})
	}
	if options.ValidateSchema && len(body) > 0 && isJSONContentType(result.ResponseContentType) && json.Valid(body) {
		schemaFindings := validateAPISchemaFindings(operation, result, body)
		findings = append(findings, schemaFindings...)
	}
	if options.IncludeUnauthenticatedComparison && options.Authenticated && operation.RequiresAuthentication != nil && *operation.RequiresAuthentication && result.HTTPStatus != nil && result.UnauthenticatedStatus != nil {
		authSucceeded := *result.HTTPStatus >= 200 && *result.HTTPStatus < 300
		unauthSucceeded := *result.UnauthenticatedStatus >= 200 && *result.UnauthenticatedStatus < 300
		if authSucceeded && unauthSucceeded {
			findings = append(findings, Finding{
				Title:          "Authenticated API operation also allowed unauthenticated access",
				Severity:       "low",
				Category:       "api_auth_comparison",
				Confidence:     "medium",
				Description:    fmt.Sprintf("%s succeeded with authentication and also returned HTTP %d without authentication.", target, *result.UnauthenticatedStatus),
				Recommendation: "Confirm whether this operation is intended to be public despite declaring authentication requirements in OpenAPI.",
			})
		}
	}
	return dedupeAPIFindings(findings)
}

func apiContractStatus(result APICheckResult, findings []Finding, options APISmokeExecutionOptions) string {
	if result.Status == StatusSkipped || !options.ValidateContract {
		return "skipped"
	}
	if result.ErrorMessage != "" {
		return "failed"
	}
	for _, finding := range findings {
		if strings.HasPrefix(finding.Category, "api_contract_") || finding.Category == "api_auth_failure" {
			return "failed"
		}
	}
	if result.HTTPStatus == nil {
		return "unknown"
	}
	return "passed"
}

func schemaValidationErrors(findings []Finding) []string {
	errors := make([]string, 0)
	for _, finding := range findings {
		if finding.Category == "api_contract_schema_mismatch" || finding.Category == "api_contract_required_field_missing" {
			errors = append(errors, finding.Description)
		}
	}
	return errors
}

func apiSmokeBaseURL(project Project, spec APISpec) (string, error) {
	if project.APIBaseURL != "" {
		return project.APIBaseURL, nil
	}
	if spec.ServerURL != "" {
		return spec.ServerURL, nil
	}
	if project.OpenAPIURL != "" {
		parsed, err := url.Parse(project.OpenAPIURL)
		if err == nil && parsed.Scheme != "" && parsed.Host != "" {
			parsed.Path = ""
			parsed.RawQuery = ""
			parsed.Fragment = ""
			return parsed.String(), nil
		}
	}
	return "", fmt.Errorf("project API base URL or OpenAPI server URL is required for API smoke execution")
}

func resolveAPIOperationURL(baseURL *url.URL, operation APIOperation) (*url.URL, error) {
	path := operation.ResolvedPath
	if path == "" {
		path = operation.Path
	}
	if strings.Contains(path, "{") || strings.Contains(path, "}") {
		return nil, fmt.Errorf("operation path still contains unresolved parameters")
	}
	next := *baseURL
	basePath := strings.TrimRight(next.Path, "/")
	next.Path = basePath + "/" + strings.TrimLeft(path, "/")
	next.RawQuery = operation.QueryString
	next.Fragment = ""
	return &next, nil
}

func sameAPIOrigin(left *url.URL, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) && strings.EqualFold(left.Host, right.Host)
}

func statusSeverity(status int) string {
	if status >= 500 {
		return "high"
	}
	return "medium"
}

func isJSONContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "+json")
}

func dedupeAPIFindings(findings []Finding) []Finding {
	seen := make(map[string]struct{}, len(findings))
	out := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		key := finding.Title + "|" + finding.Description
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, finding)
	}
	return out
}

func sanitizeURLForStorage(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return RedactSecrets(raw)
	}
	parsed.User = nil
	if parsed.RawQuery != "" {
		values := parsed.Query()
		for key := range values {
			if sensitiveAPIParameterName(key) {
				values.Set(key, "[REDACTED]")
			}
		}
		parsed.RawQuery = values.Encode()
	}
	return parsed.String()
}

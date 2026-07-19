package main

import "testing"

func TestValidateAPISchemaFindingsDetectsRequiredFieldMissing(t *testing.T) {
	status := 200
	operation := APIOperation{
		Method: "GET",
		Path:   "/private/profile",
		ResponseSchemas: map[string]any{
			"200": map[string]any{
				"application/json": map[string]any{
					"type":     "object",
					"required": []any{"id", "email"},
					"properties": map[string]any{
						"id":    map[string]any{"type": "string"},
						"email": map[string]any{"type": "string"},
					},
				},
			},
		},
	}
	result := APICheckResult{
		Method:              "GET",
		Path:                "/private/profile",
		ResponseContentType: "application/json",
		HTTPStatus:          &status,
	}
	findings := validateAPISchemaFindings(operation, result, []byte(`{"id":"demo"}`))
	if len(findings) != 1 {
		t.Fatalf("expected one schema finding, got %#v", findings)
	}
	if findings[0].Category != "api_contract_required_field_missing" {
		t.Fatalf("expected required field category, got %#v", findings[0])
	}
}

func TestBuildAPIOperationFindingsWithUnauthenticatedComparison(t *testing.T) {
	authStatus := 200
	unauthStatus := 200
	requiresAuth := true
	findings := buildAPIOperationFindingsWithOptions(
		APIOperation{Method: "GET", Path: "/private/profile", RequiresAuthentication: &requiresAuth, ExpectedStatuses: []string{"200"}},
		APICheckResult{Method: "GET", Path: "/private/profile", ResolvedURL: "https://api.example.test/private/profile", HTTPStatus: &authStatus, UnauthenticatedStatus: &unauthStatus},
		[]byte(`{"id":"demo"}`),
		false,
		APISmokeExecutionOptions{Authenticated: true, IncludeUnauthenticatedComparison: true, ValidateContract: true, ValidateSchema: true, AuthMaterial: &apiAuthMaterial{Type: APIAuthProfileTypeBearerToken}},
	)
	found := false
	for _, finding := range findings {
		if finding.Category == "api_auth_comparison" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected unauthenticated comparison finding, got %#v", findings)
	}
}

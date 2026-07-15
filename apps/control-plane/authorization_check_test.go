package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeCredentialProfileRequestRoleMetadata(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	input, err := normalizeCredentialProfileRequest(CredentialProfileRequest{
		Name:                "Admin",
		RoleName:            " admin ",
		RoleDescription:     " Demo administrator ",
		SubjectLabel:        " Admin User ",
		Username:            "admin@example.com",
		Password:            "admin-password",
		LoginURL:            "http://demo-web:8080/login",
		UsernameSelector:    "#username",
		PasswordSelector:    "#password",
		SubmitSelector:      "#login-submit",
		SuccessURLContains:  "/dashboard",
		SuccessTextContains: "Authenticated area",
	}, project, true)
	if err != nil {
		t.Fatalf("normalize credential profile: %v", err)
	}
	if input.RoleName != "admin" || input.RoleDescription != "Demo administrator" || input.SubjectLabel != "Admin User" {
		t.Fatalf("role metadata was not normalized: %#v", input)
	}
}

func TestNormalizeAuthorizationCheckRequestBrowserTargetSafety(t *testing.T) {
	project := Project{
		FrontendURL:         "http://demo-web:8080/",
		AllowedHosts:        []string{"demo-web"},
		AllowPrivateTargets: true,
	}
	input, err := normalizeAuthorizationCheckRequest(AuthorizationCheckRequest{
		Name:                     "Readonly denied admin",
		Type:                     AuthorizationCheckTypeBrowserURL,
		ActorCredentialProfileID: "11111111-1111-1111-1111-111111111111",
		ExpectedOutcome:          AuthorizationExpectedDenied,
		TargetURL:                "/admin",
		DeniedTextContains:       "Access denied",
		Enabled:                  boolPtr(true),
	}, project)
	if err != nil {
		t.Fatalf("normalize authorization check: %v", err)
	}
	if input.TargetURL != "http://demo-web:8080/admin" {
		t.Fatalf("relative target was not normalized to frontend origin: %q", input.TargetURL)
	}

	_, err = normalizeAuthorizationCheckRequest(AuthorizationCheckRequest{
		Name:                     "External",
		Type:                     AuthorizationCheckTypeBrowserURL,
		ActorCredentialProfileID: "11111111-1111-1111-1111-111111111111",
		ExpectedOutcome:          AuthorizationExpectedDenied,
		TargetURL:                "http://evil.example/admin",
		Enabled:                  boolPtr(true),
	}, project)
	if err == nil || !strings.Contains(err.Error(), "target_url") {
		t.Fatalf("expected external target rejection, got %v", err)
	}

	_, err = normalizeAuthorizationCheckRequest(AuthorizationCheckRequest{
		Name:                     "Sensitive query",
		Type:                     AuthorizationCheckTypeBrowserURL,
		ActorCredentialProfileID: "11111111-1111-1111-1111-111111111111",
		ExpectedOutcome:          AuthorizationExpectedDenied,
		TargetURL:                "/admin?token=secret",
		Enabled:                  boolPtr(true),
	}, project)
	if err == nil || !strings.Contains(err.Error(), "sensitive") {
		t.Fatalf("expected sensitive query rejection, got %v", err)
	}
}

func TestNormalizeAuthorizationRunRequest(t *testing.T) {
	input, err := normalizeAuthorizationRunRequest(AuthorizationCheckRunRequest{
		CheckIDs: []string{
			"11111111-1111-1111-1111-111111111111",
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
		},
	})
	if err != nil {
		t.Fatalf("normalize authorization run: %v", err)
	}
	if input.MaxChecks != 10 {
		t.Fatalf("expected default max checks, got %d", input.MaxChecks)
	}
	if len(input.CheckIDs) != 2 {
		t.Fatalf("expected duplicate check IDs to be removed, got %#v", input.CheckIDs)
	}
}

func TestAuthorizationSafeAIInputDoesNotIncludeCredentialValues(t *testing.T) {
	status := 200
	duration := 123
	report := &AuthorizationCheckReport{
		Run: AuthorizationCheckRun{ID: "run-1", ProjectID: "project-1", Status: StatusCompleted},
		Project: Project{
			ID:   "project-1",
			Name: "Demo",
		},
		Checks: []AuthorizationCheck{{
			ID:                       "check-1",
			Name:                     "Admin route",
			Type:                     AuthorizationCheckTypeBrowserURL,
			ActorCredentialProfileID: "profile-1",
			ExpectedOutcome:          AuthorizationExpectedAllowed,
			TargetURL:                "http://demo-web:8080/admin?session=secret",
			Enabled:                  true,
		}},
		Results: []AuthorizationCheckResult{{
			ID:                       "result-1",
			RunID:                    "run-1",
			CheckID:                  "check-1",
			Status:                   StatusPassed,
			ExpectedOutcome:          AuthorizationExpectedAllowed,
			ActualOutcome:            AuthorizationActualAllowed,
			ActorCredentialProfileID: "profile-1",
			ActorRoleName:            "admin",
			TargetURL:                "http://demo-web:8080/admin?token=secret",
			FinalURL:                 "http://demo-web:8080/admin?session=secret",
			HTTPStatus:               &status,
			DurationMS:               &duration,
			ErrorMessage:             "authorization=Bearer abc",
		}},
		Evidence: []Evidence{{
			Type: "authorization_observations",
			Metadata: map[string]any{
				"actor_credential_profile_name": "Admin profile",
				"actor_role_name":               "admin",
				"username":                      "admin@example.com",
				"password":                      "admin-password",
				"cookie":                        "session=secret",
			},
		}},
	}
	raw, err := json.Marshal(BuildSafeAuthorizationAIInput(report))
	if err != nil {
		t.Fatalf("marshal safe authorization input: %v", err)
	}
	body := string(raw)
	for _, leaked := range []string{"admin@example.com", "admin-password", "session=secret", "Bearer abc", "token=secret"} {
		if strings.Contains(body, leaked) {
			t.Fatalf("safe authorization AI input leaked %q in %s", leaked, body)
		}
	}
	if !strings.Contains(body, "admin") || !strings.Contains(body, "allowed") {
		t.Fatalf("safe authorization AI input should preserve role and outcome context: %s", body)
	}
}

func boolPtr(value bool) *bool {
	return &value
}

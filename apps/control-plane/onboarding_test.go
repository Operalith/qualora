package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeProjectSetupRequestAppliesSafeDefaults(t *testing.T) {
	input, err := NormalizeProjectSetupRequest(ProjectSetupRequest{
		Project: CreateProjectRequest{
			Name:                "Demo",
			FrontendURL:         "http://demo-web:8080",
			APIBaseURL:          "http://demo-api:8080",
			AllowedHosts:        []string{"demo-web", "demo-api"},
			SecurityMode:        "passive",
			AllowPrivateTargets: true,
		},
		AI:       ProjectSetupAIConfig{Mode: "demo"},
		APISpec:  ProjectSetupAPISpecConfig{Mode: "demo"},
		Workflow: ProjectSetupWorkflowSelection{UseDefaults: true},
	})
	if err != nil {
		t.Fatalf("normalize setup request: %v", err)
	}
	if input.Project.DestructiveActions {
		t.Fatal("guided setup should keep destructive actions disabled")
	}
	if input.AI.Provider == nil || input.AI.Provider.BaseURL != demoFakeLLMBaseURL {
		t.Fatalf("expected demo fake LLM provider defaults, got %#v", input.AI.Provider)
	}
	if input.APISpec.Spec == nil || input.APISpec.Spec.SourceURL != "http://demo-api:8080/openapi.yaml" {
		t.Fatalf("expected demo OpenAPI defaults, got %#v", input.APISpec.Spec)
	}
	if !input.Workflow.BrowserSmoke || !input.Workflow.Discovery || !input.Workflow.QualityChecks || !input.Workflow.SafeQARun || !input.Workflow.APISmoke {
		t.Fatalf("expected guided default workflows to be enabled, got %#v", input.Workflow)
	}
	if input.Workflow.ExecuteSafeQA {
		t.Fatal("safe QA should default to preview-only unless explicitly approved")
	}
}

func TestNormalizeProjectSetupRequestKeepsExplicitEmptyWorkflow(t *testing.T) {
	input, err := NormalizeProjectSetupRequest(ProjectSetupRequest{
		Project: CreateProjectRequest{
			Name:                "Config Only",
			FrontendURL:         "http://demo-web:8080",
			AllowedHosts:        []string{"demo-web"},
			SecurityMode:        "passive",
			AllowPrivateTargets: true,
		},
		AI:       ProjectSetupAIConfig{Mode: "skip"},
		Workflow: ProjectSetupWorkflowSelection{},
	})
	if err != nil {
		t.Fatalf("normalize setup request: %v", err)
	}
	if input.Workflow.BrowserSmoke || input.Workflow.Discovery || input.Workflow.QualityChecks || input.Workflow.SafeQARun || input.Workflow.APISmoke || input.Workflow.AuthenticatedSmoke {
		t.Fatalf("expected explicit empty workflow to remain empty, got %#v", input.Workflow)
	}
}

func TestNormalizeProjectSetupRequestRejectsDestructiveActions(t *testing.T) {
	_, err := NormalizeProjectSetupRequest(ProjectSetupRequest{
		Project: CreateProjectRequest{
			Name:                "Unsafe",
			FrontendURL:         "http://demo-web:8080",
			AllowedHosts:        []string{"demo-web"},
			SecurityMode:        "passive",
			DestructiveActions:  true,
			AllowPrivateTargets: true,
		},
	})
	if err == nil {
		t.Fatal("expected destructive setup to be rejected")
	}
}

func TestNormalizeProjectSetupRequestValidatesOptionalModes(t *testing.T) {
	base := ProjectSetupRequest{
		Project: CreateProjectRequest{
			Name:                "Demo",
			FrontendURL:         "http://demo-web:8080",
			AllowedHosts:        []string{"demo-web"},
			SecurityMode:        "passive",
			AllowPrivateTargets: true,
		},
	}
	for name, mutate := range map[string]func(ProjectSetupRequest) ProjectSetupRequest{
		"missing existing provider": func(input ProjectSetupRequest) ProjectSetupRequest {
			input.AI.Mode = "existing"
			return input
		},
		"missing provider config": func(input ProjectSetupRequest) ProjectSetupRequest {
			input.AI.Mode = "create"
			return input
		},
		"missing credential profile": func(input ProjectSetupRequest) ProjectSetupRequest {
			input.Credential.Mode = "create"
			return input
		},
		"missing API spec": func(input ProjectSetupRequest) ProjectSetupRequest {
			input.APISpec.Mode = "import"
			return input
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NormalizeProjectSetupRequest(mutate(base)); err == nil {
				t.Fatal("expected setup request to be rejected")
			}
		})
	}
}

func TestProjectSetupResponseDoesNotExposeSecrets(t *testing.T) {
	response := ProjectSetupResponse{
		Project: Project{ID: "project-1", Name: "Demo"},
		AIProvider: &AIProvider{
			ID:              "provider-1",
			Name:            "Fake",
			APIKeyEncrypted: "encrypted-api-key",
		},
		CredentialProfile: &CredentialProfile{
			ID:                "credential-1",
			Name:              "Login",
			UsernameEncrypted: "encrypted-username",
			PasswordEncrypted: "encrypted-password",
		},
	}
	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal setup response: %v", err)
	}
	serialized := string(body)
	for _, secret := range []string{"encrypted-api-key", "encrypted-username", "encrypted-password", "raw-password", "fake-key"} {
		if strings.Contains(serialized, secret) {
			t.Fatalf("setup response leaked secret marker %q in %s", secret, serialized)
		}
	}
}

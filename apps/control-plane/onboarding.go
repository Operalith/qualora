package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const demoFakeLLMBaseURL = "http://fake-llm:8080/v1"

func NormalizeProjectSetupRequest(input ProjectSetupRequest) (ProjectSetupRequest, error) {
	normalizedProject, err := NormalizeProjectRequest(input.Project)
	if err != nil {
		return input, err
	}
	if normalizedProject.DestructiveActions {
		return input, fmt.Errorf("guided setup does not support destructive_actions=true")
	}
	input.Project = normalizedProject

	input.AI.Mode = strings.ToLower(strings.TrimSpace(input.AI.Mode))
	input.AI.ProviderID = strings.TrimSpace(input.AI.ProviderID)
	if input.AI.Mode == "" {
		if input.AI.ProviderID != "" {
			input.AI.Mode = "existing"
		} else if input.AI.Provider != nil {
			input.AI.Mode = "create"
		} else {
			input.AI.Mode = "skip"
		}
	}
	switch input.AI.Mode {
	case "skip":
		input.AI.ProviderID = ""
		input.AI.Provider = nil
	case "existing":
		if input.AI.ProviderID == "" {
			return input, fmt.Errorf("ai.provider_id is required when ai.mode is existing")
		}
		input.AI.Provider = nil
	case "create":
		if input.AI.Provider == nil {
			return input, fmt.Errorf("ai.provider is required when ai.mode is create")
		}
		provider, err := normalizeProviderRequest(*input.AI.Provider)
		if err != nil {
			return input, fmt.Errorf("ai.provider: %w", err)
		}
		input.AI.Provider = &provider
	case "demo":
		provider, err := normalizeProviderRequest(demoAIProviderRequest())
		if err != nil {
			return input, fmt.Errorf("demo AI provider defaults are invalid: %w", err)
		}
		input.AI.Provider = &provider
		input.AI.ProviderID = ""
	default:
		return input, fmt.Errorf("ai.mode must be skip, existing, create, or demo")
	}

	input.Credential.Mode = strings.ToLower(strings.TrimSpace(input.Credential.Mode))
	if input.Credential.Mode == "" {
		if input.Credential.Profile != nil {
			input.Credential.Mode = "create"
		} else {
			input.Credential.Mode = "skip"
		}
	}
	switch input.Credential.Mode {
	case "skip":
		input.Credential.Profile = nil
	case "create":
		if input.Credential.Profile == nil {
			return input, fmt.Errorf("credential.profile is required when credential.mode is create")
		}
	default:
		return input, fmt.Errorf("credential.mode must be skip or create")
	}

	input.APISpec.Mode = strings.ToLower(strings.TrimSpace(input.APISpec.Mode))
	if input.APISpec.Mode == "" {
		if input.APISpec.Spec != nil {
			input.APISpec.Mode = "import"
		} else {
			input.APISpec.Mode = "skip"
		}
	}
	switch input.APISpec.Mode {
	case "skip":
		input.APISpec.Spec = nil
	case "import":
		if input.APISpec.Spec == nil {
			return input, fmt.Errorf("api_spec.spec is required when api_spec.mode is import")
		}
		spec, err := NormalizeAPISpecImportRequest(*input.APISpec.Spec)
		if err != nil {
			return input, fmt.Errorf("api_spec.spec: %w", err)
		}
		input.APISpec.Spec = &spec
	case "demo":
		spec, err := NormalizeAPISpecImportRequest(APISpecImportRequest{
			Name:       "Qualora Demo API",
			SourceType: "url",
			SourceURL:  "http://demo-api:8080/openapi.yaml",
		})
		if err != nil {
			return input, fmt.Errorf("demo API spec defaults are invalid: %w", err)
		}
		input.APISpec.Spec = &spec
	default:
		return input, fmt.Errorf("api_spec.mode must be skip, import, or demo")
	}

	if input.Workflow.UseDefaults {
		input.Workflow = defaultSetupWorkflow(input)
	}
	return input, nil
}

func (a *App) RunProjectSetup(ctx context.Context, input ProjectSetupRequest) (*ProjectSetupResponse, error) {
	normalized, err := NormalizeProjectSetupRequest(input)
	if err != nil {
		return nil, err
	}

	project, err := a.store.CreateProject(ctx, normalized.Project)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	response := &ProjectSetupResponse{
		Project:   *project,
		Skipped:   []ProjectSetupSkipped{},
		Timeline:  []ProjectSetupTimelineItem{},
		NextLinks: map[string]string{"project": "/api/v1/projects/" + project.ID},
	}
	response.addTimeline("project_created", StatusCompleted, project.ID, "")

	var provider *AIProvider
	switch normalized.AI.Mode {
	case "skip":
		response.addSkipped("ai_provider", "AI is optional and was skipped")
	case "existing":
		provider, err = a.store.GetAIProvider(ctx, normalized.AI.ProviderID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				response.addSkipped("ai_provider", "selected AI provider was not found")
				break
			}
			return response, fmt.Errorf("load AI provider: %w", err)
		}
		response.AIProvider = provider
		response.Started.AIProviderID = provider.ID
		response.addTimeline("ai_provider_selected", StatusCompleted, provider.ID, "")
	case "create", "demo":
		createdProvider, err := a.createAIProviderFromSetup(ctx, *normalized.AI.Provider)
		if err != nil {
			return response, fmt.Errorf("create AI provider: %w", err)
		}
		provider = createdProvider
		response.AIProvider = createdProvider
		response.Started.AIProviderID = createdProvider.ID
		response.addTimeline("ai_provider_configured", StatusCompleted, createdProvider.ID, "")
	}

	var credential *CredentialProfile
	if normalized.Credential.Mode == "create" && normalized.Credential.Profile != nil {
		profileInput, err := normalizeCredentialProfileRequest(*normalized.Credential.Profile, *project, true)
		if err != nil {
			return response, fmt.Errorf("credential profile: %w", err)
		}
		profile, err := a.credentialProfileFromInput(profileInput, "", "", "")
		if err != nil {
			return response, fmt.Errorf("encrypt credential profile: %w", err)
		}
		createdProfile, err := a.store.CreateCredentialProfile(ctx, project.ID, profile)
		if err != nil {
			return response, fmt.Errorf("create credential profile: %w", err)
		}
		credential = createdProfile
		response.CredentialProfile = createdProfile
		response.Started.CredentialProfileID = createdProfile.ID
		response.addTimeline("credential_configured", StatusCompleted, createdProfile.ID, "")
	} else {
		response.addSkipped("credential_profile", "login credentials were skipped")
	}

	var apiSpec *APISpec
	if normalized.APISpec.Mode != "skip" && normalized.APISpec.Spec != nil {
		detail, err := a.createAPISpecFromInput(ctx, *project, *normalized.APISpec.Spec)
		if err != nil {
			return response, fmt.Errorf("import API spec: %w", err)
		}
		apiSpec = &detail.Spec
		response.APISpec = &detail.Spec
		response.Started.APISpecID = detail.Spec.ID
		response.NextLinks["api_spec"] = "/api/v1/api-specs/" + detail.Spec.ID
		response.addTimeline("openapi_imported", detail.Spec.Status, detail.Spec.ID, detail.Spec.ErrorMessage)
	} else {
		response.addSkipped("openapi_import", "OpenAPI import was skipped")
	}

	a.startSelectedSetupWorkflows(ctx, *project, provider, credential, apiSpec, normalized.Workflow, response)
	return response, nil
}

func (a *App) createAIProviderFromSetup(ctx context.Context, input AIProviderRequest) (*AIProvider, error) {
	provider, err := a.providerFromInput(input, "", "")
	if err != nil {
		return nil, err
	}
	return a.store.CreateAIProvider(ctx, provider)
}

func (a *App) startSelectedSetupWorkflows(ctx context.Context, project Project, provider *AIProvider, credential *CredentialProfile, apiSpec *APISpec, workflow ProjectSetupWorkflowSelection, response *ProjectSetupResponse) {
	if workflow.BrowserSmoke {
		if project.FrontendURL == "" {
			response.addSkipped("browser_smoke", "project has no frontend URL")
		} else if run, err := a.createAndEnqueueRun(ctx, project, []string{JobKindBrowser}, RunOptions{RunType: RunTypeBrowserSmoke, CaptureScreenshot: true, MaxDurationSeconds: 30}); err != nil {
			response.addSkipped("browser_smoke", RedactSecrets(err.Error()))
		} else {
			response.Started.BrowserSmokeRunID = run.ID
			response.NextLinks["browser_report"] = "/api/v1/runs/" + run.ID + "/report"
			response.addTimeline("browser_smoke_started", StatusQueued, run.ID, "")
		}
	}

	if workflow.AuthenticatedSmoke {
		if project.FrontendURL == "" {
			response.addSkipped("authenticated_smoke", "project has no frontend URL")
		} else if credential == nil {
			response.addSkipped("authenticated_smoke", "credential profile was not configured")
		} else if run, err := a.createAndEnqueueRun(ctx, project, []string{JobKindBrowser}, RunOptions{
			RunType:             RunTypeAuthenticatedBrowserSmoke,
			CredentialProfileID: credential.ID,
			TargetPath:          "/dashboard",
			CaptureScreenshot:   true,
			MaxDurationSeconds:  30,
		}); err != nil {
			response.addSkipped("authenticated_smoke", RedactSecrets(err.Error()))
		} else {
			response.Started.AuthenticatedSmokeRunID = run.ID
			response.NextLinks["authenticated_browser_report"] = "/api/v1/runs/" + run.ID + "/report"
			response.addTimeline("authenticated_smoke_started", StatusQueued, run.ID, "")
		}
	}

	var discoveryRun *DiscoveryRun
	if workflow.Discovery {
		if project.FrontendURL == "" {
			response.addSkipped("discovery", "project has no frontend URL")
		} else {
			input := DiscoveryRunRequest{MaxPages: defaultDiscoveryMaxPages, MaxDepth: defaultDiscoveryMaxDepth}
			if credential != nil {
				input.CredentialProfileID = credential.ID
			}
			run, err := a.createAndEnqueueDiscoveryRun(ctx, project, input)
			if err != nil {
				response.addSkipped("discovery", RedactSecrets(err.Error()))
			} else {
				discoveryRun = run
				response.Started.DiscoveryRunID = run.ID
				response.NextLinks["discovery_report"] = "/api/v1/discovery-runs/" + run.ID + "/report"
				response.addTimeline("discovery_started", StatusQueued, run.ID, "")
			}
		}
	}

	if workflow.QualityChecks {
		if project.FrontendURL == "" {
			response.addSkipped("quality_checks", "project has no frontend URL")
		} else {
			qualityInput := QualityCheckRunRequest{MaxPages: defaultQualityMaxPages}
			if credential != nil {
				qualityInput.CredentialProfileID = credential.ID
			}
			if discoveryRun != nil && discoveryRun.Status == StatusCompleted {
				qualityInput.DiscoveryRunID = discoveryRun.ID
			}
			run, err := a.createAndEnqueueQualityRun(ctx, project, qualityInput)
			if err != nil {
				response.addSkipped("quality_checks", RedactSecrets(err.Error()))
			} else {
				response.Started.QualityCheckRunID = run.ID
				response.NextLinks["quality_report"] = "/api/v1/quality-check-runs/" + run.ID + "/report"
				response.addTimeline("quality_check_started", StatusQueued, run.ID, "")
			}
		}
	}

	if workflow.SafeQARun {
		if project.FrontendURL == "" {
			response.addSkipped("safe_qa_run", "project has no frontend URL")
		} else if provider == nil {
			response.addSkipped("safe_qa_run", "AI provider is required for Safe QA planning")
		} else {
			includeQuality := workflow.QualityChecks
			input := QARunRequest{
				Mode:                 "safe",
				ProviderID:           provider.ID,
				Execute:              workflow.ExecuteSafeQA,
				MaxPages:             defaultDiscoveryMaxPages,
				MaxDepth:             defaultDiscoveryMaxDepth,
				MaxScenarios:         defaultMaxTestPlanScenarios,
				IncludeQualityChecks: &includeQuality,
				QualityMaxPages:      defaultQualityMaxPages,
				ProductContext:       "Guided setup first Safe QA workflow.",
				FocusAreas:           []string{"smoke", "functional", "regression"},
			}
			if credential != nil {
				input.CredentialProfileID = credential.ID
			}
			run, err := a.createSafeQARunFromSetup(ctx, project, input)
			if err != nil {
				response.addSkipped("safe_qa_run", RedactSecrets(err.Error()))
			} else {
				response.Started.SafeQARunID = run.ID
				response.NextLinks["safe_qa_report"] = "/api/v1/qa-runs/" + run.ID + "/report"
				response.addTimeline("safe_qa_run_started", StatusRunning, run.ID, "")
			}
		}
	}

	if workflow.APISmoke {
		if apiSpec == nil {
			response.addSkipped("api_smoke", "OpenAPI spec was not imported")
		} else if apiSpec.Status != "parsed" {
			response.addSkipped("api_smoke", "OpenAPI spec is not parsed")
		} else {
			operations, err := a.store.ListAPIOperations(ctx, apiSpec.ID)
			if err != nil {
				response.addSkipped("api_smoke", RedactSecrets(err.Error()))
			} else if run, err := a.ExecuteAPISmokeRun(ctx, project, *apiSpec, operations); err != nil {
				response.addSkipped("api_smoke", RedactSecrets(err.Error()))
				if run != nil {
					response.Started.APISmokeRunID = run.ID
					response.NextLinks["api_report"] = "/api/v1/runs/" + run.ID + "/report"
				}
			} else {
				response.Started.APISmokeRunID = run.ID
				response.NextLinks["api_report"] = "/api/v1/runs/" + run.ID + "/report"
				response.addTimeline("api_smoke_completed", run.Status, run.ID, run.ErrorMessage)
			}
		}
	}
}

func (a *App) createAndEnqueueRun(ctx context.Context, project Project, kinds []string, options RunOptions) (*TestRun, error) {
	run, jobs, err := a.store.CreateRunForKindsWithOptions(ctx, project, kinds, options)
	if err != nil {
		return nil, err
	}
	if err := a.enqueueRunJobs(ctx, project, run, jobs); err != nil {
		_ = a.store.MarkRunFailed(ctx, run.ID, "run could not be queued")
		return nil, err
	}
	return run, nil
}

func (a *App) createAndEnqueueDiscoveryRun(ctx context.Context, project Project, input DiscoveryRunRequest) (*DiscoveryRun, error) {
	normalized, err := NormalizeDiscoveryRunRequest(project, input)
	if err != nil {
		return nil, err
	}
	run, err := a.store.CreateDiscoveryRun(ctx, project, normalized)
	if err != nil {
		return nil, err
	}
	if err := a.queue.EnqueueDiscoveryRun(ctx, DiscoveryRunJob{DiscoveryRunID: run.ID, ProjectID: project.ID}); err != nil {
		_ = a.store.MarkDiscoveryRunFailed(ctx, run.ID, "discovery run could not be queued")
		return nil, err
	}
	return run, nil
}

func (a *App) createAndEnqueueQualityRun(ctx context.Context, project Project, input QualityCheckRunRequest) (*QualityCheckRun, error) {
	normalized, err := NormalizeQualityCheckRunRequest(project, input)
	if err != nil {
		return nil, err
	}
	run, err := a.store.CreateQualityCheckRun(ctx, project.ID, normalized)
	if err != nil {
		return nil, err
	}
	if err := a.queue.EnqueueQualityCheckRun(ctx, QualityCheckRunJob{QualityCheckRunID: run.ID, ProjectID: project.ID}); err != nil {
		_ = a.store.MarkQualityCheckRunFailed(ctx, run.ID, "quality check run could not be queued")
		return nil, err
	}
	return run, nil
}

func (a *App) createSafeQARunFromSetup(ctx context.Context, project Project, input QARunRequest) (*QARun, error) {
	normalized, err := NormalizeQARunRequest(project, input)
	if err != nil {
		return nil, err
	}
	run, err := a.store.CreateQARun(ctx, project.ID, normalized)
	if err != nil {
		return nil, err
	}
	go a.runSafeQARun(run.ID, project, normalized)
	return run, nil
}

func demoAIProviderRequest() AIProviderRequest {
	redaction := true
	sendArtifacts := false
	return AIProviderRequest{
		Name:              "Qualora Demo Fake LLM",
		Preset:            "custom",
		Type:              AIProviderOpenAICompatible,
		BaseURL:           demoFakeLLMBaseURL,
		Model:             "qualora-fake-analyst",
		APIKey:            "fake-key",
		Temperature:       0.1,
		MaxOutputTokens:   1200,
		TimeoutSeconds:    30,
		SendScreenshots:   &sendArtifacts,
		SendHTML:          &sendArtifacts,
		SendNetworkBodies: &sendArtifacts,
		RedactionEnabled:  &redaction,
		IsDefault:         true,
	}
}

func defaultSetupWorkflow(input ProjectSetupRequest) ProjectSetupWorkflowSelection {
	hasFrontend := input.Project.FrontendURL != ""
	hasAI := input.AI.Mode != "skip"
	hasCredential := input.Credential.Mode == "create"
	hasAPISpec := input.APISpec.Mode != "skip"
	return ProjectSetupWorkflowSelection{
		BrowserSmoke:       hasFrontend,
		Discovery:          hasFrontend,
		QualityChecks:      hasFrontend,
		SafeQARun:          hasFrontend && hasAI,
		ExecuteSafeQA:      false,
		APISmoke:           hasAPISpec,
		AuthenticatedSmoke: hasFrontend && hasCredential,
	}
}

func (r *ProjectSetupResponse) addSkipped(action string, reason string) {
	reason = RedactSecrets(reason)
	r.Skipped = append(r.Skipped, ProjectSetupSkipped{Action: action, Reason: reason})
	r.addTimeline(action, StatusSkipped, "", reason)
}

func (r *ProjectSetupResponse) addTimeline(step string, status string, resource string, reason string) {
	r.Timeline = append(r.Timeline, ProjectSetupTimelineItem{
		Step:     step,
		Status:   status,
		Resource: resource,
		Reason:   RedactSecrets(reason),
	})
}

package main

import "testing"

func TestNormalizeProviderRequestAppliesPresetsAndSafeDefaults(t *testing.T) {
	enabled := true
	input, err := normalizeProviderRequest(AIProviderRequest{
		Name:             "OpenRouter",
		Preset:           "openrouter",
		RedactionEnabled: &enabled,
	})
	if err != nil {
		t.Fatalf("normalize provider: %v", err)
	}
	if input.Type != AIProviderOpenAICompatible {
		t.Fatalf("unexpected provider type %q", input.Type)
	}
	if input.BaseURL != "https://openrouter.ai/api/v1" {
		t.Fatalf("unexpected base URL %q", input.BaseURL)
	}
	if input.Model != "openai/gpt-4o-mini" {
		t.Fatalf("unexpected model %q", input.Model)
	}
	if input.ExtraHeaders["X-OpenRouter-Title"] != "Qualora" {
		t.Fatalf("expected OpenRouter title header, got %#v", input.ExtraHeaders)
	}
	if input.SendScreenshots == nil || *input.SendScreenshots {
		t.Fatal("expected screenshots to be disabled by default")
	}
	if input.SendHTML == nil || *input.SendHTML {
		t.Fatal("expected HTML to be disabled by default")
	}
	if input.SendNetworkBodies == nil || *input.SendNetworkBodies {
		t.Fatal("expected network bodies to be disabled by default")
	}
}

func TestNormalizeProviderRequestRejectsInvalidValues(t *testing.T) {
	for name, request := range map[string]AIProviderRequest{
		"unsupported type": {
			Name:    "Bad",
			Type:    "native",
			BaseURL: "https://api.example.com/v1",
			Model:   "model",
		},
		"unsupported preset": {
			Name:    "Bad",
			Preset:  "mystery",
			Type:    AIProviderOpenAICompatible,
			BaseURL: "https://api.example.com/v1",
			Model:   "model",
		},
		"bad scheme": {
			Name:    "Bad",
			Type:    AIProviderOpenAICompatible,
			BaseURL: "file:///tmp/provider",
			Model:   "model",
		},
		"missing model": {
			Name:    "Bad",
			Type:    AIProviderOpenAICompatible,
			BaseURL: "https://api.example.com/v1",
		},
		"temperature": {
			Name:        "Bad",
			Type:        AIProviderOpenAICompatible,
			BaseURL:     "https://api.example.com/v1",
			Model:       "model",
			Temperature: 3,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := normalizeProviderRequest(request); err == nil {
				t.Fatal("expected request to be rejected")
			}
		})
	}
}

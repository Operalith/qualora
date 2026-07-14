package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleClientSendsExpectedRequestAndParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		if got := r.Header.Get("X-Test-Header"); got != "Qualora" {
			t.Fatalf("unexpected extra header %q", got)
		}

		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request["model"] != "qualora-test-model" {
			t.Fatalf("unexpected model %v", request["model"])
		}
		if request["max_tokens"] != float64(321) {
			t.Fatalf("unexpected max_tokens %v", request["max_tokens"])
		}
		if format, ok := request["response_format"].(map[string]any); !ok || format["type"] != "json_object" {
			t.Fatalf("expected json_object response format, got %#v", request["response_format"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"content":"{\"executive_summary\":\"ok\",\"technical_summary\":\"ok\",\"risk_level\":\"low\",\"likely_causes\":[],\"recommended_actions\":[],\"suggested_next_tests\":[],\"confidence\":0.7,\"limitations\":[]}"}}],
			"usage":{"prompt_tokens":11,"completion_tokens":22,"total_tokens":33}
		}`))
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient()
	response, err := client.ChatCompletion(context.Background(), AIClientRequest{
		BaseURL:         server.URL + "/v1",
		Model:           "qualora-test-model",
		APIKey:          "test-key",
		ExtraHeaders:    map[string]string{"X-Test-Header": "Qualora"},
		Temperature:     0.2,
		MaxOutputTokens: 321,
		TimeoutSeconds:  5,
		Messages:        []AIChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("chat completion: %v", err)
	}
	if !strings.Contains(response.Content, `"risk_level":"low"`) {
		t.Fatalf("unexpected response content %q", response.Content)
	}
	if response.TotalTokens != 33 {
		t.Fatalf("unexpected total tokens %d", response.TotalTokens)
	}
}

func TestOpenAICompatibleClientRedactsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"Bearer should-not-leak api_key=secret"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient()
	_, err := client.ChatCompletion(context.Background(), AIClientRequest{
		BaseURL:         server.URL + "/v1",
		Model:           "qualora-test-model",
		APIKey:          "test-key",
		MaxOutputTokens: 100,
		TimeoutSeconds:  5,
		Messages:        []AIChatMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected provider error")
	}
	if strings.Contains(err.Error(), "should-not-leak") || strings.Contains(err.Error(), "api_key=secret") {
		t.Fatalf("expected provider error to be redacted, got %q", err.Error())
	}
}

func TestChatCompletionsURLNormalizesBaseURL(t *testing.T) {
	endpoint, err := chatCompletionsURL("https://api.example.com/v1/?token=secret#frag")
	if err != nil {
		t.Fatalf("normalize URL: %v", err)
	}
	if endpoint != "https://api.example.com/v1/chat/completions" {
		t.Fatalf("unexpected endpoint %q", endpoint)
	}
}

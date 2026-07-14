package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OpenAICompatibleClient struct {
	httpClient *http.Client
}

type AIClientRequest struct {
	BaseURL         string
	Model           string
	APIKey          string
	ExtraHeaders    map[string]string
	Temperature     float64
	MaxOutputTokens int
	TimeoutSeconds  int
	Messages        []AIChatMessage
}

type AIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIClientResponse struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func NewOpenAICompatibleClient() *OpenAICompatibleClient {
	return &OpenAICompatibleClient{httpClient: &http.Client{}}
}

func (c *OpenAICompatibleClient) ChatCompletion(ctx context.Context, request AIClientRequest) (*AIClientResponse, error) {
	endpoint, err := chatCompletionsURL(request.BaseURL)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(request.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body := map[string]any{
		"model":       request.Model,
		"messages":    request.Messages,
		"temperature": request.Temperature,
		"max_tokens":  request.MaxOutputTokens,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal AI request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("create AI request: %w", err)
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Content-Type", "application/json")
	if request.APIKey != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+request.APIKey)
	}
	for key, value := range request.ExtraHeaders {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		httpRequest.Header.Set(key, value)
	}

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("AI provider request failed: %s", RedactSecrets(err.Error()))
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read AI response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("AI provider returned HTTP %d: %s", response.StatusCode, RedactSecrets(string(responseBody)))
	}

	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, fmt.Errorf("parse AI response: %w", err)
	}
	if len(payload.Choices) == 0 || strings.TrimSpace(payload.Choices[0].Message.Content) == "" {
		return nil, fmt.Errorf("AI provider returned no message content")
	}
	return &AIClientResponse{
		Content:          payload.Choices[0].Message.Content,
		PromptTokens:     payload.Usage.PromptTokens,
		CompletionTokens: payload.Usage.CompletionTokens,
		TotalTokens:      payload.Usage.TotalTokens,
	}, nil
}

func chatCompletionsURL(baseURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("base_url must be a valid http or https URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("base_url must use http or https")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/chat/completions"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

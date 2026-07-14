package main

import (
	"strings"
	"testing"
)

func TestParseOpenAPISpecDiscoversSafeAndSkippedOperations(t *testing.T) {
	raw := `{
		"openapi": "3.0.3",
		"info": {"title": "Demo API", "version": "1.2.3"},
		"servers": [{"url": "http://api.example.test"}],
		"paths": {
			"/health": {
				"get": {"responses": {"200": {"description": "ok", "content": {"application/json": {}}}}}
			},
			"/users/{id}": {
				"parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "example": "1"}}],
				"get": {"responses": {"200": {"description": "ok"}}},
				"delete": {"responses": {"204": {"description": "deleted"}}}
			},
			"/profile": {
				"get": {"security": [{"ApiKeyAuth": []}], "responses": {"200": {"description": "ok"}}}
			},
			"/items": {
				"post": {"requestBody": {"required": true}, "responses": {"201": {"description": "created"}}}
			}
		}
	}`

	parsed, err := ParseOpenAPISpec(raw, "", "")
	if err != nil {
		t.Fatalf("ParseOpenAPISpec returned error: %v", err)
	}
	if parsed.Title != "Demo API" || parsed.Version != "1.2.3" {
		t.Fatalf("unexpected info: %#v", parsed)
	}
	if parsed.ServerURL != "http://api.example.test" {
		t.Fatalf("unexpected server URL %q", parsed.ServerURL)
	}
	if len(parsed.Operations) != 5 {
		t.Fatalf("expected 5 operations, got %d", len(parsed.Operations))
	}

	byMethodPath := map[string]APIOperation{}
	for _, operation := range parsed.Operations {
		byMethodPath[operation.Method+" "+operation.Path] = operation
	}
	if !byMethodPath["GET /health"].SafeToExecute {
		t.Fatalf("GET /health should be safe")
	}
	userGet := byMethodPath["GET /users/{id}"]
	if !userGet.SafeToExecute || userGet.ResolvedPath != "/users/1" {
		t.Fatalf("GET /users/{id} should resolve safe path parameter, got %#v", userGet)
	}
	if byMethodPath["DELETE /users/{id}"].SafeToExecute || !strings.Contains(byMethodPath["DELETE /users/{id}"].SkipReason, "read-only") {
		t.Fatalf("DELETE should be skipped as unsafe: %#v", byMethodPath["DELETE /users/{id}"])
	}
	if byMethodPath["GET /profile"].SafeToExecute || !strings.Contains(byMethodPath["GET /profile"].SkipReason, "authentication") {
		t.Fatalf("auth-required GET should be skipped: %#v", byMethodPath["GET /profile"])
	}
	if byMethodPath["POST /items"].SafeToExecute || !strings.Contains(byMethodPath["POST /items"].SkipReason, "read-only") {
		t.Fatalf("POST should be skipped as unsafe: %#v", byMethodPath["POST /items"])
	}
}

func TestParseOpenAPISpecYAMLAndRequiredQuerySample(t *testing.T) {
	raw := `
openapi: 3.0.3
info:
  title: YAML API
  version: 0.1.0
paths:
  /search:
    get:
      parameters:
        - name: q
          in: query
          required: true
          schema:
            type: string
            default: smoke
      responses:
        "200":
          description: ok
`
	parsed, err := ParseOpenAPISpec(raw, "http://api.example.test/openapi.yaml", "")
	if err != nil {
		t.Fatalf("ParseOpenAPISpec returned error: %v", err)
	}
	if len(parsed.Operations) != 1 {
		t.Fatalf("expected one operation, got %d", len(parsed.Operations))
	}
	operation := parsed.Operations[0]
	if !operation.SafeToExecute || operation.QueryString != "q=smoke" {
		t.Fatalf("expected safe required query sample, got %#v", operation)
	}
}

func TestParseOpenAPISpecRejectsInvalidDocuments(t *testing.T) {
	if _, err := ParseOpenAPISpec(`{"swagger":"2.0","paths":{}}`, "", ""); err == nil {
		t.Fatal("expected Swagger 2.0 document to be rejected")
	}
	if _, err := ParseOpenAPISpec(`openapi: 3.0.3`, "", ""); err == nil {
		t.Fatal("expected missing paths to be rejected")
	}
}

func TestOpenAPISafeClassificationSkipsAmbiguousInputs(t *testing.T) {
	raw := `
openapi: 3.0.3
info:
  title: Unsafe Shapes
  version: 1.0.0
paths:
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok
  /download:
    get:
      parameters:
        - name: token
          in: query
          required: true
          schema:
            type: string
            example: secret-token
      responses:
        "200":
          description: ok
  /echo:
    get:
      requestBody:
        required: true
      responses:
        "200":
          description: ok
`
	parsed, err := ParseOpenAPISpec(raw, "", "http://api.example.test")
	if err != nil {
		t.Fatalf("ParseOpenAPISpec returned error: %v", err)
	}
	for _, operation := range parsed.Operations {
		if operation.SafeToExecute {
			t.Fatalf("operation should have been skipped: %#v", operation)
		}
		if operation.SkipReason == "" {
			t.Fatalf("skipped operation should include reason: %#v", operation)
		}
	}
}

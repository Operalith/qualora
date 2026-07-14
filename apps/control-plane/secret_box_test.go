package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSecretBoxEncryptsAndDecryptsValues(t *testing.T) {
	box, err := NewSecretBox("test-key-material")
	if err != nil {
		t.Fatalf("create secret box: %v", err)
	}

	encrypted, err := box.Encrypt("sk-test-secret")
	if err != nil {
		t.Fatalf("encrypt secret: %v", err)
	}
	if encrypted == "sk-test-secret" || strings.Contains(encrypted, "sk-test-secret") {
		t.Fatalf("expected encrypted value not to contain plaintext, got %q", encrypted)
	}
	if !strings.HasPrefix(encrypted, secretBoxPrefix) {
		t.Fatalf("expected encrypted value to use %q prefix, got %q", secretBoxPrefix, encrypted)
	}

	decrypted, err := box.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt secret: %v", err)
	}
	if decrypted != "sk-test-secret" {
		t.Fatalf("unexpected decrypted value %q", decrypted)
	}
}

func TestSecretBoxRejectsUnsupportedCiphertext(t *testing.T) {
	box, err := NewSecretBox("test-key-material")
	if err != nil {
		t.Fatalf("create secret box: %v", err)
	}
	if _, err := box.Decrypt("plaintext-secret"); err == nil {
		t.Fatal("expected unsupported ciphertext to be rejected")
	}
}

func TestAIProviderJSONDoesNotExposeEncryptedSecrets(t *testing.T) {
	raw, err := json.Marshal(AIProvider{
		Name:                   "OpenAI",
		Type:                   AIProviderOpenAICompatible,
		BaseURL:                "https://api.openai.com/v1",
		Model:                  "gpt-4o-mini",
		APIKeyEncrypted:        "v1:encrypted-key",
		ExtraHeadersEncrypted:  "v1:encrypted-headers",
		APIKeyConfigured:       true,
		ExtraHeadersConfigured: true,
		RedactionEnabled:       true,
		MaxOutputTokens:        1200,
		TimeoutSeconds:         30,
	})
	if err != nil {
		t.Fatalf("marshal provider: %v", err)
	}
	body := string(raw)
	if strings.Contains(body, "encrypted-key") || strings.Contains(body, "encrypted-headers") {
		t.Fatalf("provider JSON leaked encrypted secret fields: %s", body)
	}
	if !strings.Contains(body, `"api_key_configured":true`) {
		t.Fatalf("provider JSON should expose configured flag, got %s", body)
	}
}

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const secretBoxPrefix = "v1:"

type SecretBox struct {
	gcm cipher.AEAD
}

func NewSecretBox(keyMaterial string) (*SecretBox, error) {
	if strings.TrimSpace(keyMaterial) == "" {
		return nil, fmt.Errorf("encryption key is required")
	}
	key := sha256.Sum256([]byte(keyMaterial))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &SecretBox{gcm: gcm}, nil
}

func (b *SecretBox) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, b.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("create nonce: %w", err)
	}
	ciphertext := b.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return secretBoxPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (b *SecretBox) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if !strings.HasPrefix(ciphertext, secretBoxPrefix) {
		return "", fmt.Errorf("encrypted value has unsupported format")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, secretBoxPrefix))
	if err != nil {
		return "", fmt.Errorf("decode encrypted value: %w", err)
	}
	if len(raw) < b.gcm.NonceSize() {
		return "", fmt.Errorf("encrypted value is too short")
	}
	nonce := raw[:b.gcm.NonceSize()]
	payload := raw[b.gcm.NonceSize():]
	plaintext, err := b.gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt value: %w", err)
	}
	return string(plaintext), nil
}

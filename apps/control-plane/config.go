package main

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port               string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	BrowserQueue       string
	APIQueue           string
	PlanExecutionQueue string
	EvidenceDir        string
	S3Endpoint         string
	S3Region           string
	S3Bucket           string
	S3AccessKeyID      string
	S3SecretKey        string
	S3ForcePath        bool
	EncryptionKey      string
	CORSOrigins        []string
	SessionTTL         time.Duration
	CookieSecure       bool
	AuthDisabled       bool
	ShutdownTimeout    time.Duration
}

func LoadConfig() Config {
	return Config{
		Port:               env("PORT", "8080"),
		DatabaseURL:        env("DATABASE_URL", "postgres://qualora:qualora@localhost:5432/qualora?sslmode=disable"),
		RedisAddr:          env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		BrowserQueue:       env("BROWSER_RUN_QUEUE", env("RUN_QUEUE", "qualora:browser-runs")),
		APIQueue:           env("API_RUN_QUEUE", "qualora:api-runs"),
		PlanExecutionQueue: env("TEST_PLAN_EXECUTION_QUEUE", "qualora:test-plan-executions"),
		EvidenceDir:        env("EVIDENCE_DIR", "/tmp/qualora-evidence"),
		S3Endpoint:         env("S3_ENDPOINT", "http://localhost:9000"),
		S3Region:           env("S3_REGION", "us-east-1"),
		S3Bucket:           env("S3_BUCKET", "qualora-evidence"),
		S3AccessKeyID:      env("S3_ACCESS_KEY_ID", "qualora"),
		S3SecretKey:        env("S3_SECRET_ACCESS_KEY", "qualora-dev-secret"),
		S3ForcePath:        boolEnv("S3_FORCE_PATH_STYLE", true),
		EncryptionKey:      env("QUALORA_ENCRYPTION_KEY", "qualora-insecure-dev-key-change-me"),
		CORSOrigins:        csvEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		SessionTTL:         time.Duration(intEnv("QUALORA_SESSION_TTL_HOURS", 12)) * time.Hour,
		CookieSecure:       boolEnv("QUALORA_COOKIE_SECURE", false),
		AuthDisabled:       boolEnv("QUALORA_AUTH_DISABLED", false),
		ShutdownTimeout:    10 * time.Second,
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func csvEnv(key, fallback string) []string {
	value := env(key, fallback)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func boolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes"
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	var parsed int
	for _, r := range value {
		if r < '0' || r > '9' {
			return fallback
		}
		parsed = parsed*10 + int(r-'0')
	}
	if parsed <= 0 {
		return fallback
	}
	return parsed
}

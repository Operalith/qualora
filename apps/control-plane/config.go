package main

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	BrowserQueue    string
	APIQueue        string
	CORSOrigins     []string
	ShutdownTimeout time.Duration
}

func LoadConfig() Config {
	return Config{
		Port:            env("PORT", "8080"),
		DatabaseURL:     env("DATABASE_URL", "postgres://qualora:qualora@localhost:5432/qualora?sslmode=disable"),
		RedisAddr:       env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		BrowserQueue:    env("BROWSER_RUN_QUEUE", env("RUN_QUEUE", "qualora:browser-runs")),
		APIQueue:        env("API_RUN_QUEUE", "qualora:api-runs"),
		CORSOrigins:     csvEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		ShutdownTimeout: 10 * time.Second,
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

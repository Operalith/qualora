package main

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	BrowserQueue    string
	APIQueue        string
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

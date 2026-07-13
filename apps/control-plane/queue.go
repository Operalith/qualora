package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client *redis.Client
	name   string
}

type BrowserRunJob struct {
	RunID     string `json:"run_id"`
	ProjectID string `json:"project_id"`
}

func NewQueue(cfg Config) *Queue {
	return &Queue{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
		}),
		name: cfg.QueueName,
	}
}

func (q *Queue) Close() error {
	return q.client.Close()
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

func (q *Queue) EnqueueBrowserRun(ctx context.Context, job BrowserRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal browser run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.name, payload).Err(); err != nil {
		return fmt.Errorf("enqueue browser run: %w", err)
	}
	return nil
}

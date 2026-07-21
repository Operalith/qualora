package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client             *redis.Client
	browserQueue       string
	apiQueue           string
	planExecutionQueue string
}

type BrowserRunJob struct {
	JobID     string `json:"job_id"`
	RunID     string `json:"run_id"`
	ProjectID string `json:"project_id"`
}

type AuthorizationCheckRunJob struct {
	AuthorizationCheckRunID string `json:"authorization_check_run_id"`
	ProjectID               string `json:"project_id"`
}

type DiscoveryRunJob struct {
	DiscoveryRunID string `json:"discovery_run_id"`
	ProjectID      string `json:"project_id"`
}

type QualityCheckRunJob struct {
	QualityCheckRunID string `json:"quality_check_run_id"`
	ProjectID         string `json:"project_id"`
}

type SafeExplorerRunJob struct {
	SafeExplorerRunID string `json:"safe_explorer_run_id"`
	ProjectID         string `json:"project_id"`
}

type AIBrowserControlRunJob struct {
	AIBrowserControlRunID string `json:"ai_browser_control_run_id"`
	ProjectID             string `json:"project_id"`
}

type FormTestRunJob struct {
	FormTestRunID string `json:"form_test_run_id"`
	ProjectID     string `json:"project_id"`
}

type APIRunJob struct {
	JobID     string `json:"job_id"`
	RunID     string `json:"run_id"`
	ProjectID string `json:"project_id"`
}

type TestPlanExecutionJob struct {
	ExecutionID string `json:"execution_id"`
}

func NewQueue(cfg Config) *Queue {
	return &Queue{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
		}),
		browserQueue:       cfg.BrowserQueue,
		apiQueue:           cfg.APIQueue,
		planExecutionQueue: cfg.PlanExecutionQueue,
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
	if err := q.client.RPush(ctx, q.browserQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue browser run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueAuthorizationCheckRun(ctx context.Context, job AuthorizationCheckRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal authorization check run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.browserQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue authorization check run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueDiscoveryRun(ctx context.Context, job DiscoveryRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal discovery run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.browserQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue discovery run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueQualityCheckRun(ctx context.Context, job QualityCheckRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal quality check run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.browserQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue quality check run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueSafeExplorerRun(ctx context.Context, job SafeExplorerRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal safe explorer run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.browserQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue safe explorer run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueAIBrowserControlRun(ctx context.Context, job AIBrowserControlRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal AI Browser Control run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.browserQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue AI Browser Control run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueFormTestRun(ctx context.Context, job FormTestRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal form test run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.browserQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue form test run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueAPIRun(ctx context.Context, job APIRunJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal api run job: %w", err)
	}
	if err := q.client.RPush(ctx, q.apiQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue api run: %w", err)
	}
	return nil
}

func (q *Queue) EnqueueTestPlanExecution(ctx context.Context, job TestPlanExecutionJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal test plan execution job: %w", err)
	}
	if err := q.client.RPush(ctx, q.planExecutionQueue, payload).Err(); err != nil {
		return fmt.Errorf("enqueue test plan execution: %w", err)
	}
	return nil
}

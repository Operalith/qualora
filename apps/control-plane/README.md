# Control Plane

Go service for Qualora's API and orchestration layer.

Responsibilities:

- Project configuration.
- Test run lifecycle.
- Policy validation.
- Worker job scheduling.
- Metadata persistence.
- JSON and HTML report access.
- Stored evidence object access by evidence ID.
- Optional AI provider management.
- Optional AI report analysis.
- AI-assisted test plan storage.
- Approved safe test plan execution orchestration.

The MVP delegates browser checks, API checks, and safe test plan execution to workers through Redis queues.

The API performs project target validation, including `allowed_hosts` enforcement, blocked private/metadata targets by default, and DNS resolution checks for hostnames.

## Local Development

```bash
go test ./...
go run .
```

The service expects PostgreSQL and Redis. Defaults are suitable for local development:

- `DATABASE_URL=postgres://qualora:qualora@localhost:5432/qualora?sslmode=disable`
- `REDIS_ADDR=localhost:6379`
- `BROWSER_RUN_QUEUE=qualora:browser-runs`
- `API_RUN_QUEUE=qualora:api-runs`
- `TEST_PLAN_EXECUTION_QUEUE=qualora:test-plan-executions`
- `S3_ENDPOINT=http://localhost:9000`
- `S3_BUCKET=qualora-evidence`
- `S3_ACCESS_KEY_ID=qualora`
- `S3_SECRET_ACCESS_KEY=qualora-dev-secret`
- `EVIDENCE_DIR=/tmp/qualora-evidence`
- `CORS_ALLOWED_ORIGINS=http://localhost:3000`
- `QUALORA_ENCRYPTION_KEY=qualora-insecure-dev-key-change-me`

Current report endpoints:

- `GET /api/v1/runs/{run_id}/report`
- `GET /api/v1/runs/{run_id}/report.html`
- `GET /api/v1/evidence/{evidence_id}`
- `GET /api/v1/ai/providers`
- `POST /api/v1/ai/providers`
- `POST /api/v1/ai/providers/{provider_id}/test`
- `GET /api/v1/runs/{run_id}/ai-analysis`
- `POST /api/v1/runs/{run_id}/ai-analysis`
- `POST /api/v1/test-plans/{test_plan_id}/executions`
- `GET /api/v1/test-plans/{test_plan_id}/executions`
- `GET /api/v1/test-plan-executions/{execution_id}`
- `GET /api/v1/test-plan-executions/{execution_id}/report`
- `GET /api/v1/test-plan-executions/{execution_id}/report.html`

The default encryption key is for local development only. Set a strong `QUALORA_ENCRYPTION_KEY` before storing real AI provider credentials.

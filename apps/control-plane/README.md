# Control Plane

Go service for Qualora's API and orchestration layer.

Responsibilities:

- Project configuration.
- Test run lifecycle.
- Policy validation.
- Worker job scheduling.
- Metadata persistence.
- Local first-run admin setup and session authentication.
- JSON and HTML report access.
- Stored evidence object access by evidence ID.
- Optional AI provider management.
- Optional AI report analysis.
- AI-assisted test plan storage.
- Approved safe test plan execution orchestration.
- OpenAPI import, operation discovery, safe API smoke execution, and API result reporting.
- Passive quality check run orchestration and quality report rendering.
- Credential profile CRUD, deterministic login checks, authenticated browser smoke orchestration, and role-aware authorization check orchestration.

The MVP delegates browser checks, credential-profile login checks, authenticated browser smoke checks, application discovery, passive quality checks, role-aware authorization checks, legacy project API checks, and safe test plan execution to workers through Redis queues. Imported-spec API smoke execution runs in the control plane so API operations and result rows are first-class API/UI resources.

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
- `QUALORA_SESSION_TTL_HOURS=12`
- `QUALORA_COOKIE_SECURE=false`
- `QUALORA_AUTH_DISABLED=false`

Public endpoints:

- `GET /healthz`
- `GET /api/v1/setup/status`
- `POST /api/v1/setup/admin`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

Current protected report and workflow endpoints include:

- `GET /api/v1/runs/{run_id}/report`
- `GET /api/v1/runs/{run_id}/report.html`
- `GET /api/v1/evidence/{evidence_id}`
- `GET /api/v1/projects/{project_id}/credential-profiles`
- `POST /api/v1/projects/{project_id}/credential-profiles`
- `GET /api/v1/credential-profiles/{credential_profile_id}`
- `PUT /api/v1/credential-profiles/{credential_profile_id}`
- `DELETE /api/v1/credential-profiles/{credential_profile_id}`
- `POST /api/v1/credential-profiles/{credential_profile_id}/test-login`
- `POST /api/v1/projects/{project_id}/authenticated-browser-smoke-runs`
- `GET /api/v1/projects/{project_id}/authorization-checks`
- `POST /api/v1/projects/{project_id}/authorization-checks`
- `GET /api/v1/authorization-checks/{authorization_check_id}`
- `PUT /api/v1/authorization-checks/{authorization_check_id}`
- `DELETE /api/v1/authorization-checks/{authorization_check_id}`
- `GET /api/v1/projects/{project_id}/authorization-check-runs`
- `POST /api/v1/projects/{project_id}/authorization-check-runs`
- `GET /api/v1/authorization-check-runs/{authorization_check_run_id}`
- `GET /api/v1/authorization-check-runs/{authorization_check_run_id}/report`
- `GET /api/v1/authorization-check-runs/{authorization_check_run_id}/report.html`
- `GET /api/v1/projects/{project_id}/quality-check-runs`
- `POST /api/v1/projects/{project_id}/quality-check-runs`
- `GET /api/v1/quality-check-runs/{quality_check_run_id}`
- `GET /api/v1/quality-check-runs/{quality_check_run_id}/report`
- `GET /api/v1/quality-check-runs/{quality_check_run_id}/report.html`
- `GET /api/v1/ai/providers`
- `POST /api/v1/ai/providers`
- `POST /api/v1/ai/providers/{provider_id}/test`
- `POST /api/v1/projects/{project_id}/api-specs`
- `GET /api/v1/projects/{project_id}/api-specs`
- `GET /api/v1/api-specs/{api_spec_id}`
- `DELETE /api/v1/api-specs/{api_spec_id}`
- `GET /api/v1/api-specs/{api_spec_id}/operations`
- `POST /api/v1/api-specs/{api_spec_id}/api-smoke-runs`
- `GET /api/v1/runs/{run_id}/api-results`
- `GET /api/v1/runs/{run_id}/ai-analysis`
- `POST /api/v1/runs/{run_id}/ai-analysis`
- `POST /api/v1/test-plans/{test_plan_id}/executions`
- `GET /api/v1/test-plans/{test_plan_id}/executions`
- `GET /api/v1/test-plan-executions/{execution_id}`
- `GET /api/v1/test-plan-executions/{execution_id}/report`
- `GET /api/v1/test-plan-executions/{execution_id}/report.html`

After setup, project, credential, AI provider, evidence, report, API spec, authorization, and test-plan endpoints require a local admin session. Mutating protected requests must send `X-Qualora-CSRF` with the value from the `qualora_csrf` cookie.

The default encryption key is for local development only. Set a strong `QUALORA_ENCRYPTION_KEY` before storing real AI provider credentials or credential profiles. `QUALORA_AUTH_DISABLED=true` is a local development/debugging escape hatch only.

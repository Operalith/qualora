# Architecture

Qualora v0.3.0-alpha is a small Docker Compose MVP for browser and API QA smoke runs with a minimal web UI and human-friendly HTML reports.

## Runtime Components

```text
API client / smoke script / qualora-web
        |
        v
qualora-api
        |
        +--> PostgreSQL: projects, test_runs, run_jobs, findings, evidence
        +--> Redis: browser and API run queues
        |
        +--> qualora-worker-browser
        |       +--> Playwright Chromium smoke test
        |       +--> MinIO/S3 screenshot evidence
        |
        +--> qualora-worker-api
                +--> API base URL checks
                +--> OpenAPI 3.x safe method checks
                +--> PostgreSQL evidence and findings
```

## Services

### `qualora-api`

The Go control plane exposes the HTTP API, validates project scope, persists metadata, creates per-run jobs, queues worker jobs, and renders JSON/HTML reports.

Current endpoints:

- `GET /healthz`
- `POST /api/v1/projects`
- `GET /api/v1/projects`
- `GET /api/v1/projects/{project_id}`
- `POST /api/v1/projects/{project_id}/runs`
- `GET /api/v1/projects/{project_id}/runs`
- `GET /api/v1/runs`
- `GET /api/v1/runs/{run_id}`
- `GET /api/v1/runs/{run_id}/report`
- `GET /api/v1/runs/{run_id}/report.html`

### `qualora-web`

The React/Vite web UI is intentionally small. It calls the control-plane API from the browser and displays:

- Projects and project details.
- Run lists and run details.
- Structured JSON report data as readable tables.
- Findings, evidence metadata, browser metadata, API metadata, and job metadata.
- Links to the self-contained HTML report export.

It has no authentication in this alpha and should be exposed only in trusted local/self-hosted environments.

### `qualora-worker-browser`

The Node.js browser worker consumes Redis browser jobs and runs a Playwright smoke check against `frontend_url`.

It currently captures:

- Page title.
- Screenshot.
- Console errors.
- Failed network requests.
- Blocked out-of-scope browser requests.
- Basic findings for obvious load, console, request, and scope issues.

### `qualora-worker-api`

The Node.js API worker consumes Redis API jobs and performs safe API checks against `api_base_url` and `openapi_url`.

It currently captures:

- API base URL status code.
- Response content type.
- Response time.
- Connection, TLS, DNS, and fetch errors.
- OpenAPI 3.x document summary.
- Safe OpenAPI operation checks for `GET`, `HEAD`, and `OPTIONS`.
- Findings for unreachable APIs, invalid OpenAPI documents, 5xx responses, unexpected status codes, obvious content type mismatches, and visible stack traces.

It does not perform authenticated API checks, request body generation, schema fuzzing, or destructive methods.

### PostgreSQL

PostgreSQL stores durable metadata:

- `projects`
- `test_runs`
- `run_jobs`
- `findings`
- `evidence`

Reports are generated dynamically from run, job, finding, and evidence rows.

### Redis

Redis is the MVP queue for browser and API jobs. PostgreSQL remains the source of durable run state.

### MinIO

MinIO stores screenshot evidence through the S3-compatible API. API evidence is stored as metadata rows in PostgreSQL. The v0.3 UI displays evidence metadata and URIs, but it does not proxy or preview screenshot objects yet.

### `mock-api`

The `mock-api` Compose service is profile-gated for smoke tests. It is not part of the production runtime path.

## Run Lifecycle

1. A client or web UI creates a project with one or more targets: `frontend_url`, `api_base_url`, or `openapi_url`.
2. The API validates URL scope and target safety.
3. A client starts a run.
4. The API creates a `pending` run in PostgreSQL.
5. The API creates `run_jobs` for browser and/or API work.
6. The API pushes worker jobs to Redis.
7. Workers mark their jobs `running`.
8. Workers collect evidence and findings.
9. Workers mark jobs `completed` or `failed`.
10. PostgreSQL refreshes the parent run status from job statuses.
11. The API serves the structured JSON report and the self-contained HTML report.

## Intentional Alpha Constraints

- No user accounts or authentication.
- Web UI is alpha and suitable only for trusted self-hosted/local environments.
- No evidence object proxy or screenshot preview in the UI yet.
- No authenticated API testing.
- No login automation.
- No active security scanning.
- No destructive API testing by default.
- No full OpenAPI schema validation or fuzzing.
- No Kubernetes or Helm support yet.

The older MVP notes remain in [architecture/mvp.md](architecture/mvp.md) for implementation context, but this document is the release-facing architecture reference.

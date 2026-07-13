# Architecture

Qualora v0.1.0-alpha is a small Docker Compose MVP for browser-based QA smoke runs.

## Runtime Components

```text
API client / smoke script
        |
        v
qualora-api
        |
        +--> PostgreSQL: projects, test_runs, findings, evidence
        +--> Redis: browser run queue
        |
        v
qualora-worker-browser
        |
        +--> Playwright Chromium smoke test
        +--> MinIO/S3 screenshot evidence
        +--> PostgreSQL findings and evidence metadata
```

## Services

### `qualora-api`

The Go control plane exposes the HTTP API, validates project scope, persists metadata, and queues browser runs.

Current endpoints:

- `GET /healthz`
- `POST /api/v1/projects`
- `GET /api/v1/projects`
- `GET /api/v1/projects/{project_id}`
- `POST /api/v1/projects/{project_id}/runs`
- `GET /api/v1/runs/{run_id}`
- `GET /api/v1/runs/{run_id}/report`

### `qualora-worker-browser`

The Node.js worker consumes Redis jobs and runs a Playwright smoke check against the configured `frontend_url`.

It currently captures:

- Page title.
- Screenshot.
- Console errors.
- Failed network requests.
- Blocked out-of-scope browser requests.
- Basic findings for obvious load, console, request, and scope issues.

### PostgreSQL

PostgreSQL stores durable metadata:

- `projects`
- `test_runs`
- `findings`
- `evidence`

Reports are generated dynamically from run, finding, and evidence rows.

### Redis

Redis is the MVP queue for browser run jobs. PostgreSQL remains the source of durable run state.

### MinIO

MinIO stores screenshot evidence through the S3-compatible API. If MinIO writes fail, the browser worker falls back to local filesystem storage and records a `file://` evidence URI.

## Run Lifecycle

1. A client creates a project with `frontend_url` and `allowed_hosts`.
2. The API validates URL scope and target safety.
3. A client starts a run.
4. The API creates a `pending` run in PostgreSQL.
5. The API pushes a browser job to Redis.
6. The browser worker marks the run `running`.
7. Playwright opens the target page and enforces allowed hosts on browser requests.
8. The worker stores screenshot evidence and browser observations.
9. The worker creates findings for obvious issues.
10. The worker marks the run `completed` or `failed`.
11. The API serves the structured report.

## Intentional Alpha Constraints

- No web UI.
- No user accounts or authentication.
- No API worker.
- No OpenAPI contract testing.
- No login automation.
- No active security scanning.
- No Kubernetes or Helm support yet.

The older MVP notes remain in [architecture/mvp.md](architecture/mvp.md) for implementation context, but this document is the release-facing architecture reference.

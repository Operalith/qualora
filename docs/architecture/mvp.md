# MVP Architecture Notes

The release-facing architecture reference is [../architecture.md](../architecture.md).

This file records the implementation intent behind the early alpha MVP.

## Implemented In The Early Alpha

- Docker Compose stack.
- Go control plane API.
- Minimal React web UI.
- PostgreSQL metadata storage.
- Redis browser run queue.
- TypeScript/Node.js Playwright browser worker.
- TypeScript/Node.js API worker.
- MinIO/S3-compatible screenshot storage.
- Structured report endpoint.
- Self-contained HTML report endpoint.

## Current Run Lifecycle

1. User creates a project with the web UI or `POST /api/v1/projects`.
2. Control plane validates `frontend_url`, `allowed_hosts`, `security_mode`, and `destructive_actions`.
3. User starts a run with `POST /api/v1/projects/{project_id}/runs`.
4. Control plane creates a `pending` run and browser/API `run_jobs` based on configured targets.
5. Control plane pushes worker jobs to Redis.
6. Workers mark jobs `running`.
7. Workers collect evidence and findings.
8. Workers mark jobs `completed` or `failed`.
9. PostgreSQL refreshes the parent run status.
10. Control plane serves the JSON report with `GET /api/v1/runs/{run_id}/report`.
11. Control plane serves the HTML report with `GET /api/v1/runs/{run_id}/report.html`.

## Current Data Model

- `projects`
- `test_runs`
- `run_jobs`
- `findings`
- `evidence`

Reports are generated dynamically from these tables.

## Future Boundaries

These are planned boundaries, not implemented release features:

- Passive security worker.
- Analyzer worker.
- Report engine package.
- Helm chart.
- Evidence object proxy or signed URL preview.
- Login automation and credential storage.

Keep future work behind explicit docs and roadmap updates so the alpha remains honest about what it can do today.

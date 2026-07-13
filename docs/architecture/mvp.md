# MVP Architecture Notes

The release-facing architecture reference is [../architecture.md](../architecture.md).

This file records the implementation intent behind the v0.1.0-alpha MVP.

## Implemented In v0.1.0-alpha

- Docker Compose stack.
- Go control plane API.
- PostgreSQL metadata storage.
- Redis browser run queue.
- TypeScript/Node.js Playwright browser worker.
- MinIO/S3-compatible screenshot storage.
- Structured report endpoint.

## Current Run Lifecycle

1. User creates a project with `POST /api/v1/projects`.
2. Control plane validates `frontend_url`, `allowed_hosts`, `security_mode`, and `destructive_actions`.
3. User starts a run with `POST /api/v1/projects/{project_id}/runs`.
4. Control plane creates a `pending` run and pushes a browser job to Redis.
5. Browser worker marks the run `running`.
6. Browser worker opens the target page with Playwright.
7. Browser worker enforces `allowed_hosts` on browser requests.
8. Browser worker captures screenshot and browser observations.
9. Browser worker writes evidence metadata and findings to PostgreSQL.
10. Browser worker marks the run `completed` or `failed`.
11. Control plane serves the report with `GET /api/v1/runs/{run_id}/report`.

## Current Data Model

- `projects`
- `test_runs`
- `findings`
- `evidence`

Reports are generated dynamically from these tables.

## Future Boundaries

These are planned boundaries, not implemented release features:

- API worker.
- Passive security worker.
- Analyzer worker.
- Report engine package.
- Web UI.
- Helm chart.
- Login automation and credential storage.

Keep future work behind explicit docs and roadmap updates so the alpha remains honest about what it can do today.

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
- Evidence object download endpoint for stored artifacts.
- Project-scoped credential profiles.
- Deterministic selector-based login checks.
- Authenticated browser smoke runs.
- Explicit role-aware authorization checks.
- OpenAPI import, operation discovery, and safe API smoke execution.
- Optional AI provider management, AI report analysis, AI-assisted test planning, and approved safe plan execution.

## Current Run Lifecycle

1. User creates a project with the web UI or `POST /api/v1/projects`.
2. Control plane validates `frontend_url`, `allowed_hosts`, `security_mode`, and `destructive_actions`.
3. User starts a run with `POST /api/v1/projects/{project_id}/runs`, a credential-profile login check, an authenticated smoke run, an authorization check run, a safe API smoke run, or an approved safe test plan execution.
4. Control plane creates a `queued` run and browser/API `run_jobs` based on configured targets.
5. Control plane pushes worker jobs to Redis.
6. Workers mark jobs `running`.
7. Workers collect evidence and findings.
8. Workers mark jobs `completed` or `failed`.
9. PostgreSQL refreshes the parent run status.
10. Control plane serves the JSON report with `GET /api/v1/runs/{run_id}/report`.
11. Control plane serves the HTML report with `GET /api/v1/runs/{run_id}/report.html`.
12. Control plane serves stored evidence objects with `GET /api/v1/evidence/{evidence_id}`.

## Current Data Model

- `projects`
- `credential_profiles`
- `authorization_checks`
- `authorization_check_runs`
- `authorization_check_results`
- `api_specs`
- `api_operations`
- `api_check_results`
- `test_runs`
- `run_jobs`
- `findings`
- `evidence`
- `ai_providers`
- `ai_analyses`
- `test_plans`
- `test_plan_executions`

Reports are generated dynamically from these tables.

## Future Boundaries

These are planned boundaries, not implemented release features:

- Passive security worker.
- Analyzer worker.
- Report engine package.
- Helm chart.
- Signed URL support or stronger evidence access controls.
- Multi-step authenticated journeys, MFA, and session export.
- Authenticated API authorization checks.

Keep future work behind explicit docs and roadmap updates so the alpha remains honest about what it can do today.

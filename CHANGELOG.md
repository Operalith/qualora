# Changelog

All notable changes to Qualora will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses semantic versioning once stable releases begin.

## [v0.3.0-alpha] - 2026-07-14

### Added

- Minimal React/Vite web UI under `apps/web`.
- `qualora-web` Docker Compose service on `http://localhost:3000`.
- Project list, project creation, project details, and start-run workflows in the UI.
- Run list and run report pages in the UI.
- Findings, evidence metadata, browser metadata, API metadata, and job metadata display.
- Self-contained HTML report export at `GET /api/v1/runs/{run_id}/report.html`.
- Run listing endpoints for all runs and project-scoped runs.
- Web build/type-check coverage in Makefile and CI.
- Smoke script output for JSON report URLs, HTML report URLs, and web UI URL.

### Changed

- Control plane now allows the local web UI origin through a narrow CORS configuration.
- Documentation now describes the web UI as alpha and trusted-environment only.

### Known Limitations

- No authentication or authorization.
- Web UI is intentionally minimal.
- Evidence object proxy, screenshot preview, and signed artifact URLs are not implemented yet.
- No project editing, pagination, or advanced report filtering.

## [v0.2.0-alpha] - 2026-07-13

### Added

- API worker for safe API checks.
- Redis API run queue and per-run `run_jobs` tracking.
- API baseline checks for `api_base_url`.
- OpenAPI 3.x JSON/YAML fetch and parse for `openapi_url`.
- Safe OpenAPI operation checks for `GET`, `HEAD`, and `OPTIONS`.
- API evidence types: `api_observations`, `openapi_summary`, and `api_request`.
- API findings for unreachable APIs, invalid OpenAPI documents, 5xx responses, unexpected status codes, obvious content type mismatches, and visible stack traces.
- Deterministic local mock API smoke service.
- API worker CI build/test coverage.

### Changed

- Runs can now enqueue browser and/or API worker jobs.
- Reports include per-job status metadata.
- Project validation now accepts API-only projects when `api_base_url` or `openapi_url` is provided.

### Security

- API worker enforces `allowed_hosts` and default private/metadata target blocking.
- Unsafe OpenAPI methods are skipped by default.
- Destructive API testing remains unsupported in this alpha.

### Known Limitations

- No authenticated API testing.
- No request body generation.
- No full OpenAPI schema validation or fuzzing.
- Workers still write results directly to PostgreSQL.

## [v0.1.0-alpha] - 2026-07-13

### Added

- Go control plane API with health, project, run, and report endpoints.
- PostgreSQL migrations for projects, test runs, findings, and evidence.
- Redis-backed browser run queue.
- TypeScript/Node.js Playwright browser worker.
- Browser smoke checks for page title, screenshot capture, console errors, failed requests, and blocked out-of-scope requests.
- MinIO/S3 screenshot evidence storage with local filesystem fallback.
- Structured JSON report output.
- Docker Compose stack for local self-hosted use.
- Makefile with development, test, lint, Compose, logs, and smoke commands.
- Smoke script for `https://example.com`.
- OpenAPI contract for the current API.
- GitHub Actions CI workflow.
- Release, architecture, development, roadmap, and security model docs.

### Security

- Allowed-host validation for project targets.
- Default blocking for localhost, loopback, link-local, private IP literal targets, cloud metadata addresses, and public hostnames resolving to blocked IP ranges.
- Browser worker request blocking outside project `allowed_hosts`.
- Secret-like value redaction in worker logs.

### Known Limitations

- No web UI.
- No API worker or OpenAPI contract checks.
- No authentication.
- No login automation or credential storage.
- No active security scanning.
- No Helm/Kubernetes deployment.

# Changelog

All notable changes to Qualora will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses semantic versioning once stable releases begin.

## [v0.7.0-alpha] - 2026-07-14

### Added

- Approved safe test plan execution endpoint at `POST /api/v1/test-plans/{test_plan_id}/executions`.
- Dry-run preview mode for safe execution mapping with executable/skipped counts and skip reasons.
- Execution list, detail, JSON report, and HTML report endpoints for test plan executions.
- `test_plan_executions`, `test_plan_execution_scenarios`, and `test_plan_execution_steps` persistence.
- Nullable execution ownership on findings and evidence so execution reports can include findings, screenshots, browser observations, and link-check metadata.
- Browser worker support for the safe execution queue.
- Supported safe browser DSL actions: `goto`, title/URL/text/element/link assertions, same-origin link status checks, screenshots, browser signal collection, load-state waits, and no-error/no-failed-request assertions.
- Web UI safe execution preview, execute, history, and execution report pages.
- Deterministic `demo-web` routes for `/`, `/status`, and `/about`.
- Fake LLM test plan output using executable safe DSL steps.
- Smoke validation for safe execution preview, execution completion, JSON/HTML execution reports, and execution screenshot evidence download.
- Unit tests for safe execution request normalization, allowed-host/same-origin URL mapping, unsupported actions, sensitive query rejection, and unsafe scenario skips.

### Changed

- Browser worker now consumes both browser smoke jobs and safe test plan execution jobs.
- `make smoke` now validates approved safe plan execution in addition to browser/API runs, AI analysis, and test plan generation.
- OpenAPI contract now documents safe test plan execution endpoints and schemas.
- Package metadata has been updated to `0.7.0-alpha`.

### Security

- AI-generated plans remain suggestions and are never executed automatically.
- Safe execution runs only persisted deterministic DSL actions after explicit user action.
- Unsupported, ambiguous, authenticated, destructive, mutating, admin, upload, exploit, SQLi, XSS, SSRF, brute-force, out-of-scope, and sensitive-query steps are skipped with reasons.
- Browser worker revalidates same-origin frontend targets and project `allowed_hosts` before executing navigation or link checks.

### Known Limitations

- Safe test plan execution is browser-only and alpha.
- No login automation, form submission, authenticated flows, file uploads, POST/PUT/PATCH/DELETE plan actions, active security scanning, or autonomous AI browser control.
- No retries or robust interrupted-worker recovery for plan executions yet.

## [v0.6.0-alpha] - 2026-07-14

### Added

- AI-assisted test plan generation endpoint at `POST /api/v1/projects/{project_id}/ai-test-plans`.
- Project-scoped test plan listing at `GET /api/v1/projects/{project_id}/test-plans`.
- Test plan detail, delete, and JSON export endpoints at `GET/DELETE /api/v1/test-plans/{test_plan_id}` and `GET /api/v1/test-plans/{test_plan_id}/export.json`.
- `test_plans` persistence table with project, optional run, provider/model, status, normalized plan JSON, risk level, scenario count, and error metadata.
- Strict AI test plan response parser and normalizer for assumptions, coverage goals, scenarios, steps, assertions, test data, instrumentation suggestions, and limitations.
- Safe AI test planning prompt builder that uses sanitized project, run report, AI analysis, finding, and evidence metadata.
- Web UI support for generating, listing, viewing, deleting, and exporting AI test plans.
- Run report references for AI test plans generated from a run in both JSON and HTML reports.
- Fake LLM deterministic AI test plan response for local smoke tests.
- Smoke coverage for AI test plan generation, project listing, detail retrieval, JSON export, and run report cross-links.
- Unit tests for AI test plan request normalization, safe input redaction, parsing, scenario validation, and scenario limits.

### Changed

- `make smoke` now validates browser and API AI test planning in addition to AI analysis.
- OpenAPI contract now documents AI test plan endpoints and schemas.
- Package metadata has been updated to `0.6.0-alpha`.

### Security

- Generated test plans are reviewable suggestions only and are not executed by Qualora.
- AI test planning uses the same conservative redaction path as AI analysis and strips URL query strings/fragments from text sent to or stored from model output.
- Screenshots, full HTML, cookies, credentials, authorization headers, raw traces, and full network bodies are not sent to AI by default.

### Known Limitations

- No authentication or authorization.
- AI test planning is alpha and depends on the configured OpenAI-compatible provider.
- Generated steps are not executed automatically.
- No authenticated journey planning beyond high-level suggestions.
- No autonomous AI browser control, login automation, active security scanning, schema fuzzing, or destructive test execution.

## [v0.5.0-alpha] - 2026-07-14

### Added

- Optional OpenAI-compatible AI provider management in the control-plane API.
- AI Providers page in the web UI with presets for OpenAI, OpenRouter, Ollama, and custom OpenAI-compatible providers.
- Provider test endpoint with sanitized success/failure responses.
- Encrypted-at-rest storage for AI provider API keys and extra headers using `QUALORA_ENCRYPTION_KEY`.
- Synchronous AI analysis endpoint for completed runs at `POST /api/v1/runs/{run_id}/ai-analysis`.
- Latest AI analysis endpoint at `GET /api/v1/runs/{run_id}/ai-analysis`.
- Safe AI input builder that redacts secrets and strips URL query strings/fragments.
- AI analysis display in the web run report page.
- AI analysis inclusion in JSON reports and self-contained HTML reports.
- Deterministic `fake-llm` OpenAI-compatible smoke service.
- Smoke coverage for provider creation, provider test, AI analysis, JSON report inclusion, and HTML report inclusion.
- Unit tests for redaction, safe AI input generation, AI response parsing, provider validation, OpenAI-compatible client behavior, and secret encryption.

### Changed

- `make smoke` now starts `fake-llm` in addition to `demo-web` and `mock-api`.
- OpenAPI contract now documents AI provider and AI analysis endpoints.
- Package metadata has been updated to `0.5.0-alpha`.

### Security

- AI is disabled until a provider is configured.
- Redaction is enabled by default.
- Screenshots, full HTML, cookies, credentials, authorization headers, and full network bodies are not sent to AI by default.
- AI provider API keys and extra headers are never returned by the API.

### Known Limitations

- No authentication or authorization.
- AI provider management is intended only for trusted local/self-hosted alpha environments.
- Only OpenAI-compatible chat completion APIs are supported.
- AI analysis runs synchronously in the control plane for this alpha.
- No autonomous AI browser control, login automation, authenticated API testing, or active security scanning.

## [v0.4.0-alpha] - 2026-07-14

### Added

- Browser-only smoke run endpoint at `POST /api/v1/projects/{project_id}/browser-smoke-runs`.
- Web UI action to run a browser smoke test from project details and navigate to the created run.
- Stored evidence download endpoint at `GET /api/v1/evidence/{evidence_id}`.
- Screenshot preview/download in the run report UI.
- Screenshot evidence metadata for filename, object key, content type, size, storage backend, and created timestamp.
- Browser observation metadata for target URL, final URL, body text length, and timeout state.
- Deterministic local `demo-web` Compose smoke target.
- Browser finding tests for timeout, status, console error, failed request, empty page, and scope signals.
- Evidence store tests for local evidence download safety.

### Changed

- New runs and run jobs now start as `queued` before workers mark them `running`.
- Browser smoke findings now classify console errors as medium severity and include concise reproduction steps in finding descriptions.
- `make smoke` now uses local smoke targets and validates screenshot evidence download.
- Control-plane Docker/CI now use Go 1.24 to match current dependency requirements.

### Security

- Evidence downloads are served only by evidence ID for records already known to Qualora.
- MinIO credentials remain server-side; the web UI fetches evidence through the control plane.

### Known Limitations

- No authentication or authorization.
- Browser smoke remains a basic alpha smoke check, not full browser test coverage.
- No login automation, authenticated browser flows, Playwright trace export, or active security scanning.

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

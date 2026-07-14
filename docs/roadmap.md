# Roadmap

This roadmap is intentionally practical. Qualora should remain self-hosted and useful before it grows broader platform features.

## v0.1.0-alpha

Delivered:

- Docker Compose stack.
- Go control plane API.
- Project creation.
- Browser smoke run creation.
- Redis browser run queue.
- Playwright browser worker.
- Allowed-host enforcement.
- Screenshot evidence in MinIO.
- Browser observation evidence.
- Structured JSON reports.

## v0.2.0-alpha

Delivered:

- API worker.
- `api_base_url` baseline checks.
- OpenAPI 3.x JSON/YAML fetch and parse.
- Safe OpenAPI method checks for `GET`, `HEAD`, and `OPTIONS`.
- API findings and evidence in structured reports.
- Per-run worker job tracking.
- Deterministic local mock API smoke test.

## v0.3.0-alpha

Delivered:

- Minimal self-hosted web UI.
- Project creation, project lists, and project details in the UI.
- Run lists and run detail/report pages in the UI.
- Findings, evidence metadata, browser metadata, API metadata, and job metadata display.
- Self-contained HTML report export.
- Web service in Docker Compose on `http://localhost:3000`.
- CI build/type-check for the web app.

## v0.4.0-alpha

Delivered:

- Browser-only smoke run endpoint and web UI action.
- Richer browser observations: target URL, final URL, status code, body text length, timeout state, console errors, failed requests, and blocked requests.
- Screenshot evidence metadata with filename, object key, content type, size, storage backend, and created timestamp.
- Screenshot preview/download through the control-plane evidence endpoint.
- Deterministic local `demo-web` smoke target.
- Clearer queued/running/completed/failed run status display.

## v0.5.0-alpha

Delivered:

- Optional OpenAI-compatible AI provider management.
- Provider presets for OpenAI, OpenRouter, Ollama, and custom OpenAI-compatible endpoints.
- Encrypted-at-rest provider API keys and extra headers.
- Safe provider test endpoint.
- Synchronous AI analysis for completed runs.
- Sanitized AI input builder with redaction enabled by default.
- AI analysis display in the web UI, JSON report, and HTML report.
- Deterministic local `fake-llm` smoke target.

## v0.6.0-alpha

Delivered:

- AI-assisted test plan generation from sanitized project/run/report metadata.
- Strict normalized test plan JSON with assumptions, coverage goals, scenarios, steps, assertions, test data needs, instrumentation suggestions, and limitations.
- Project-scoped test plan list, detail, delete, and JSON export endpoints.
- Web UI generation form, plan list, detail page, scenario/step display, deletion, and export links.
- JSON/HTML run report references for plans generated from a run.
- Deterministic local fake LLM plan response and smoke validation.

## v0.7.0-alpha

Current alpha scope:

- Approved safe execution for AI test plans.
- Deterministic test plan safety mapper with explicit skip reasons.
- Supported browser DSL actions only: navigation, visibility assertions, link checks, screenshots, browser signal collection, and no-error assertions.
- Persisted execution, scenario, and step status.
- Execution findings and evidence linked to test plan executions.
- JSON and self-contained HTML execution reports.
- Web UI preview, execute, history, and detail pages.
- Deterministic demo-web routes and fake LLM executable plan output.
- Smoke validation for preview, execution, reports, and evidence download.

## Phase 8: Run And Worker Hardening

- Worker result API so workers do not write directly to PostgreSQL.
- Run retries and clearer failure states.
- Safe test plan execution retries and clearer interrupted-worker recovery.
- Better per-job error reporting in the public API.
- Signed URL support or stronger evidence access controls.
- Additional safety tests for DNS and worker request blocking.
- Better operational logs and container health checks.
- Move AI analysis to an async analyzer worker.
- Move AI-assisted test planning to an async analyzer worker.

## Phase 9: Deeper API Checks

- More OpenAPI validation.
- Response body/schema checks for safe methods.
- Configurable endpoint limits and path filters.
- Conservative authenticated API testing design.

## Phase 10: Passive Security Checks

- Passive security headers.
- Cookie flag checks.
- Mixed-content observations.
- TLS metadata where practical.

## Later

- Login automation with safe credential storage.
- Playwright trace capture and download.
- Project editing and report filtering in the web UI.
- Native provider integrations beyond OpenAI-compatible APIs.
- Helm chart after Docker Compose is stable.
- OWASP ZAP integration with explicit safe policies.

## Non-Goals For Now

- Hosted SaaS assumptions.
- Billing.
- Organizations and teams.
- Multi-tenancy.
- Enterprise RBAC.
- Active exploitation or destructive scanning.

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

Delivered:

- Approved safe execution for AI test plans.
- Deterministic test plan safety mapper with explicit skip reasons.
- Supported browser DSL actions only: navigation, visibility assertions, link checks, screenshots, browser signal collection, and no-error assertions.
- Persisted execution, scenario, and step status.
- Execution findings and evidence linked to test plan executions.
- JSON and self-contained HTML execution reports.
- Web UI preview, execute, history, and detail pages.
- Deterministic demo-web routes and fake LLM executable plan output.
- Smoke validation for preview, execution, reports, and evidence download.

## v0.8.0-alpha

Delivered:

- OpenAPI 3.x import from URL or inline JSON/YAML.
- Imported API spec metadata and operation discovery.
- Safe operation classification with persisted skip reasons.
- Safe API smoke runs for `GET`, `HEAD`, and `OPTIONS` only.
- Skips for mutating methods, auth-required operations, required request bodies, unresolved parameters, and sensitive paths/parameters.
- API check result rows with method, path, HTTP status, duration, content type, response size, error, and skip reason.
- JSON and HTML run reports with API summary and operation result tables.
- Web UI import, spec detail, operation list, safe API smoke run, and API result display.
- Deterministic local `demo-api` OpenAPI smoke target with a `/broken` 500 finding.
- Smoke validation for OpenAPI import, operation discovery, skipped unsafe operations, API results, API reports, browser smoke, AI analysis, AI test planning, and safe plan execution.

## v0.9.0-alpha

Delivered:

- Project-scoped encrypted credential profiles.
- Credential profile CRUD API and web UI workflow.
- Raw usernames/passwords are never returned by API responses.
- Deterministic selector-based login check endpoint.
- Authenticated browser smoke endpoint.
- Demo-web `/login` and protected `/dashboard` support.
- Login summary, login observations, screenshots, browser observations, findings, JSON reports, and HTML reports for login/authenticated runs.
- Findings for login failure, missing selectors, timeouts, console errors, failed requests, and authenticated navigation failures.
- Safe AI input support for authenticated browser runs without credentials, cookies, storage, auth headers, or tokens.
- Smoke validation for credential creation, login checks, authenticated smoke reports, password redaction, existing browser smoke, API smoke, AI analysis, AI test planning, and safe test plan execution.

## v0.10.0-alpha

Delivered:

- Credential profile role metadata for explicit test-account roles.
- Authorization check CRUD API for `browser_url` checks.
- Authorization check run API with JSON and self-contained HTML reports.
- Browser worker execution for deterministic login plus one configured same-origin authorization target.
- Expected `allowed` and `denied` outcome comparison.
- Findings for authorization bypass, unexpected denial, login failure, unknown outcome, timeout, blocked target, console errors, and network failures.
- Screenshot and `authorization_observations` evidence.
- Web UI section for authorization checks and authorization run reports.
- Demo-web admin, readonly, and customer role routes.
- Smoke validation for role credential creation, login checks, authorization checks, reports, screenshot evidence, and password redaction.

## v0.11.0-alpha

Delivered:

- First-run local admin setup.
- Single local admin role.
- Argon2id password hashing.
- HTTP-only session cookie plus CSRF token for mutating protected API requests.
- Protected API routes for project data, credential profiles, AI provider configuration, reports, evidence, runs, API specs, test plans, and authorization reports.
- Protected web UI with setup, login, session refresh, and logout states.
- Smoke validation for setup, login, logout, protected endpoints, and existing browser/API/AI/test-plan/authorization flows.
- Honest documentation for local-only alpha authentication limitations.

## v0.12.0-alpha

Delivered:

- Safe deterministic application discovery run model and API.
- Persistent application map tables for pages, links, forms, and fields.
- Browser-worker discovery execution with same-origin defaults, allowed-host enforcement, sensitive query redaction, duplicate avoidance, and crawl limits.
- Screenshot and browser observation evidence for discovered pages.
- Deterministic discovery findings for page load failures, 404/5xx responses, console errors, failed requests, empty pages, broken internal links, skipped unsafe/external links, forms without labels, and password forms.
- Web UI discovery section on project details plus discovery report/map page.
- JSON and self-contained HTML discovery reports.
- Demo-web routes, safe/unsafe/external links, and forms for deterministic smoke validation.
- Smoke validation for discovery completion, pages, links, forms, skipped links, screenshots, JSON report, and HTML report.

## v0.13.0-alpha

Delivered:

- Discovery-aware AI test plan generation from sanitized application maps.
- Safe executable test plan mode with optional deterministic DSL candidates.
- Test plan source metadata, discovery run links, and persisted executable coverage.
- Safe QA Run API that can reuse/latest/create discovery, generate an AI plan, preview safe execution, optionally execute the approved safe DSL path, and serve JSON/HTML reports.
- Web UI project and discovery report actions for generating discovery-aware plans and starting Safe QA Runs.
- Safe QA Run report page with discovery, plan, preview, optional execution, findings, evidence, and safety metadata.
- Smoke validation for discovery-aware planning, Safe QA Run preview, Safe QA Run execution, reports, password redaction, and existing browser/API/AI/test-plan flows.

## v0.14.0-alpha

Delivered:

- Standalone passive quality check runs for project frontends.
- Optional reuse of latest or selected application discovery runs as page lists.
- Optional deterministic selector-based login with credential profiles before quality checks.
- Passive security checks for missing security headers, cookie flags, mixed content, source maps, sensitive query names, and obvious form issues.
- Basic accessibility heuristics for page metadata, images, inputs, buttons, links, and landmarks.
- Basic performance/front-end observations for slow loads, failed resources, console errors, request counts, large JavaScript, and image dimensions.
- Quality check JSON reports and self-contained HTML reports.
- Web UI Quality Checks section, quality run list, and quality report page.
- Optional Safe QA Run quality-check integration and combined report fields.
- Smoke validation for standalone quality checks, Safe QA quality summaries, report redaction, and existing browser/API/AI/test-plan/discovery flows.

## v0.15.0-alpha

Delivered:

- Guided project setup wizard for project basics, optional AI provider, optional credential profile, optional OpenAPI import, workflow selection, and result links.
- `POST /api/v1/onboarding/project-setup` orchestration for project creation and selected safe first checks.
- Local demo workflow against `demo-web`, `demo-api`, and `fake-llm`.
- Dashboard quick-start actions, status indicators, recent projects, and recent Safe QA runs.
- Project readiness checklist for frontend URL, AI provider, discovery, quality checks, credentials, OpenAPI, Safe QA, and reports.
- Reports landing page for recent browser, API, discovery, quality, and Safe QA reports.
- Smoke validation for guided setup, dashboard/readiness/report UI discoverability, reports, redaction, and existing browser/API/AI/test-plan/discovery/quality/Safe QA flows.

## v0.16.0-alpha

Delivered:

- Interactive Safe Explorer run API and browser-worker queue job.
- Project-scoped Safe Explorer settings: start URL, optional credential profile, max steps, max depth, same-origin-only, and optional GET-form policy.
- Deterministic action extraction for visible links, forms, submit controls, buttons, and inputs without storing full HTML.
- Safety classifier for same-origin navigation, allowed hosts, dangerous labels, sensitive query values, unsupported schemes, mutating forms, duplicate URLs, and policy limits.
- Safe execution of classified navigation actions only; unsafe/external/unsupported/duplicate/policy-blocked actions are skipped with reasons.
- Persistent Safe Explorer runs, steps, actions, findings, screenshot evidence, trace endpoint, JSON report, and self-contained HTML report.
- Web UI project card/form, run list, report page, and reports landing-page integration.
- Demo-web fixtures for safe links, unsafe links, external links, GET forms, POST forms, unsupported buttons, and dangerous buttons.
- Smoke validation for Safe Explorer completion, executed/skipped actions, skip reasons, screenshot evidence, JSON/HTML reports, UI text, and secret redaction.

## v0.17.0-alpha

Delivered:

- Deterministic report intelligence for generic run, discovery, quality, Safe Explorer, authorization, safe test plan execution, and Safe QA reports.
- Severity normalization for findings and quality results.
- Stable finding fingerprints and grouped findings with raw details preserved.
- Duplicate reduction metadata, top findings, top affected pages, and noise/repeated-finding summaries.
- Executive summaries with pass/warning/fail/unknown status, checks completed/skipped, recommended next actions, what was tested, what was not tested, and safety limitations.
- Web UI report intelligence panels and recent report index severity/grouped counts.
- Smoke validation for report intelligence JSON fields, HTML sections, grouped findings, raw findings/results, no-secret output, and existing flows.

## v0.18.0-alpha

Current alpha scope:

- Project-scoped report baselines for Safe QA reports.
- Default baseline management with one default baseline per project/report type.
- Deterministic fingerprint-based comparison for new, fixed, unchanged, severity-changed, and affected-scope-changed findings.
- Safe QA report JSON/HTML integration for baseline status and comparison/gate summaries.
- Quality gate evaluation for CI/release checks without requiring AI.
- Compact CI quality gate response and HTTP helper script.
- Web UI actions to set a Safe QA report as baseline, compare with baseline, and evaluate quality gates.
- Project and reports UI baseline/regression indicators.
- Smoke validation for unchanged baseline comparison, passing quality gate, compact CI response, UI text, and existing flows.

## Phase 19: Run And Worker Hardening

- Worker result API so workers do not write directly to PostgreSQL.
- Run retries and clearer failure states.
- Safe test plan execution retries and clearer interrupted-worker recovery.
- Better per-job error reporting in the public API.
- Signed URL support or stronger evidence access controls.
- Login rate limiting, audit logging, password reset, and local auth hardening.
- Additional safety tests for DNS and worker request blocking.
- Better operational logs and container health checks.
- Move AI analysis to an async analyzer worker.
- Move AI-assisted test planning to an async analyzer worker.
- Safe QA Run integration for Safe Explorer summaries if it can remain explicit and non-autonomous.

## Phase 20: Deeper API Checks

- More OpenAPI validation.
- Response body/schema checks for safe methods.
- Configurable endpoint limits and path filters.
- Conservative authenticated API testing design.

## Phase 21: Quality Check Deepening

- Optional axe-core integration or richer accessibility summaries.
- Lighthouse/Core Web Vitals-style performance collection if it can stay safe and lightweight.
- More passive security metadata, including TLS details where practical.
- Better trend comparison across quality and API runs.

## Later

- Harden deterministic login support with richer validation and clearer troubleshooting.
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

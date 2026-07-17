# Changelog

All notable changes to Qualora will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses semantic versioning once stable releases begin.

## [v0.15.0-alpha] - 2026-07-17

### Added

- Guided project setup API at `POST /api/v1/onboarding/project-setup`.
- Web UI setup wizard for project basics, optional AI provider, optional credential profile, optional OpenAPI import, workflow selection, and result links.
- Dashboard quick-start cards, version badge, status indicators, recent projects, and recent Safe QA runs.
- Local demo workflow shortcut for `demo-web`, `demo-api`, and `fake-llm`.
- Project readiness checklist on project detail pages.
- Reports landing page for recent browser, API, discovery, quality, and Safe QA reports.
- Smoke coverage for guided project setup, dashboard/readiness/report UI discoverability, report links, and secret redaction.

### Changed

- Package metadata has been updated to `0.15.0-alpha`.
- The web UI navigation now surfaces Guided Setup, Browser Testing, Reports, Projects, AI Providers, and Test Plans more clearly.
- The smoke script now validates the guided setup flow in addition to existing browser, authenticated browser, API, AI, discovery, quality, Safe QA, authorization, and safe plan execution paths.

### Security

- Guided setup rejects destructive project setup and starts only existing safe workflows.
- Guided setup responses return IDs, safe metadata, skipped reasons, and report links only.
- Raw passwords, provider secrets, cookies, browser storage, authorization headers, tokens, and encrypted secret payloads are not returned by onboarding responses or sent to AI.

### Known Limitations

- Guided setup is alpha orchestration over existing Qualora capabilities.
- Demo workflow is intended for the local Docker Compose demo services.
- Reports index currently focuses on recent browser/API smoke, discovery, quality, and Safe QA reports.
- No autonomous AI browser control, arbitrary form submission, active security scanning, fuzzing, destructive testing, SSO/OIDC/SAML, multi-tenancy, or enterprise RBAC was added.

## [v0.14.0-alpha] - 2026-07-16

### Added

- Passive quality check run API for project frontends.
- Browser-worker quality checks for safe passive security headers/cookies/forms, basic accessibility heuristics, and simple performance/resource observations.
- Quality check JSON report and self-contained HTML report.
- Web UI Quality Checks section with standalone run form, run list, and report page.
- Optional Safe QA Run quality-check integration and combined report fields.
- Demo-web deterministic quality signals for smoke/demo coverage.
- Tests for quality request normalization, quality summary counts, and browser-worker quality rule generation.

### Changed

- Package metadata has been updated to `0.14.0-alpha`.
- Safe QA reports can include linked passive quality summaries and quality result rows when requested.
- The OpenAPI contract documents quality check endpoints and Safe QA quality options.

### Security

- Quality checks are passive and read-only by default.
- No active scanning, exploit payloads, fuzzing, arbitrary form submission, destructive actions, broad external crawling, or autonomous AI browser control were added.
- Quality evidence stores metadata only and excludes cookies values, browser storage, auth headers, secrets, request bodies, response bodies, and full HTML.

### Known Limitations

- Quality checks are alpha heuristics and are not full security, accessibility, performance, Lighthouse, Core Web Vitals, or WCAG audits.
- Checks currently focus on browser-observable metadata for configured frontend pages.
- Authenticated quality checks use only deterministic selector-based credential profiles.

## [v0.13.0-alpha] - 2026-07-16

### Added

- Discovery-aware AI test plan generation using sanitized application map summaries.
- Safe executable AI plan mode with optional deterministic DSL candidates.
- Stored test plan source metadata, discovery run links, and safe execution coverage counts.
- Safe QA Run API for reuse/latest/new discovery, AI plan generation, safe execution preview, optional explicit execution, and JSON/HTML reports.
- Web UI Safe QA Run controls on project pages and discovery report pages.
- Web UI Safe QA Run report page with discovery, generated plan, preview coverage, optional execution report, findings, evidence, and safety metadata.
- Fake OpenAI-compatible provider output for deterministic discovery-aware test plan smoke coverage.
- Smoke coverage for discovery-aware planning, Safe QA preview, Safe QA execution, Safe QA reports, and existing browser/API/AI/test-plan flows.

### Changed

- Package metadata has been updated to `0.13.0-alpha`.
- AI test plan requests can opt into `include_discovery_map`, `execution_mode`, and `max_pages_from_discovery`.
- The web UI now shows test plan source, discovery run links, and executable coverage when available.

### Security

- Discovery-aware AI input is sanitized and capped; it excludes credentials, cookies, browser storage, auth headers, tokens, full HTML, screenshots, request bodies, and response bodies.
- Safe QA Runs stop after preview by default and execute only persisted, mapped, non-destructive browser DSL steps after explicit user action.
- No autonomous AI browser control, arbitrary clicking, arbitrary form submission, active scanning, fuzzing, payload execution, or destructive actions were added.

### Known Limitations

- Safe QA Runs are alpha orchestration and depend on the quality of the sanitized discovery map and configured AI provider.
- Safe execution remains browser-only and limited to the supported deterministic DSL.
- Authenticated discovery and richer multi-step logged-in journeys are not part of this release.
- AI planning is optional and unavailable until an OpenAI-compatible provider is configured.

## [v0.12.0-alpha] - 2026-07-15

### Added

- Safe deterministic application discovery run API.
- Persistent application map storage for pages, links, forms, and form fields.
- Browser-worker discovery execution with same-origin defaults, allowed-host enforcement, sensitive query redaction, duplicate avoidance, crawl limits, and screenshot evidence.
- Discovery findings for page load failures, 404/5xx pages, console errors, network failures, empty pages, broken internal links, skipped unsafe/external links, forms without labels, and password forms.
- Discovery JSON report, application map endpoint, and self-contained HTML report.
- Web UI project discovery form, discovery run list, and discovery report/map page.
- Demo-web routes, safe links, unsafe/external links, and forms for deterministic discovery smoke coverage.
- Smoke coverage for discovery completion, pages, links, forms, skipped links, screenshots, JSON report, and HTML report.

### Changed

- Package metadata has been updated to `0.12.0-alpha`.
- Demo-web now includes stable `/pricing`, safe navigation links, skipped-link fixtures, and a safe newsletter form.

### Security

- Discovery does not submit forms, click arbitrary buttons, run payloads, perform destructive actions, crawl external domains by default, or use autonomous AI browser control.
- Discovery records metadata and screenshots only; it does not store full HTML, cookies, local/session storage, auth headers, tokens, credentials, request bodies, or response bodies.

### Known Limitations

- Discovery is bounded alpha coverage, not exhaustive crawling.
- Client-side routes that require arbitrary button clicks or form submission are not explored.
- Discovery reports are not automatically used as AI analysis input in this release.

## [v0.11.0-alpha] - 2026-07-15

### Added

- First-run local admin setup with `GET /api/v1/setup/status` and `POST /api/v1/setup/admin`.
- Local admin login, logout, and current-user endpoints.
- Argon2id password hashing and database-backed session records with hashed session/CSRF tokens.
- HTTP-only session cookie plus CSRF cookie/header protection for mutating authenticated API requests.
- Authentication middleware protecting projects, runs, reports, evidence downloads, AI providers, credential profiles, API specs, test plans, and authorization checks.
- Web UI setup, login, current-user, and logout states.
- Smoke coverage for setup, login, logout, protected endpoint rejection before login, authenticated smoke flow, and protected authorization report/evidence access.

### Changed

- The API and web UI are no longer openly usable after first-run setup; health and setup/auth endpoints remain public.
- Package metadata has been updated to `0.11.0-alpha`.

### Security

- Password hashes and session tokens are never returned by API responses.
- Session token hashes, not raw tokens, are stored in PostgreSQL.
- Evidence and HTML/JSON reports require authentication.

### Known Limitations

- One local `admin` role only.
- No multi-tenancy, SSO/OIDC/SAML, user-management UI, or advanced RBAC yet.
- Rate limiting and audit logging are not implemented in v0.11.

## [v0.10.0-alpha] - 2026-07-15

### Added

- Role metadata on project-scoped credential profiles.
- Authorization check CRUD API for explicit `browser_url` checks.
- Authorization check run API and browser-worker queue payload.
- Deterministic role-aware browser authorization checks using configured credential profiles.
- JSON and self-contained HTML authorization reports.
- `authorization_observations` evidence and authorization screenshot evidence.
- Web UI authorization check form, list, run action, run history, and report page.
- Demo-web admin, readonly, customer-a, and customer-b role accounts and protected routes.
- Smoke coverage for role credential profiles, login checks, authorization checks, reports, evidence download, and password redaction.
- Tests for authorization target validation, run request normalization, authorization finding logic, and safe AI input redaction.

### Changed

- Credential profile API/UI responses include optional role metadata but still never return raw credentials.
- Browser worker can now execute authorization check runs in addition to browser smoke, login, authenticated smoke, and safe test plan jobs.
- Package metadata has been updated to `0.10.0-alpha`.

### Security

- Authorization checks are explicit, read-only, same-origin, and allowed-host enforced.
- Authorization execution uses deterministic selector-based login and one configured target navigation.
- No crawling, fuzzing, payload attacks, arbitrary form submission, destructive actions, or autonomous AI browser control are performed.
- Credentials, cookies, storage, auth headers, and tokens are not sent to AI or included in authorization reports.

### Known Limitations

- No Qualora user authentication or authorization.
- Browser URL authorization checks are supported first.
- Authenticated API authorization testing is not fully supported yet.
- No active security scanning, payload-based exploitation, destructive testing, or arbitrary crawling.

## [v0.9.0-alpha] - 2026-07-15

### Added

- Project-scoped encrypted credential profiles for deterministic target-application logins.
- Credential profile CRUD API and web UI workflows.
- Deterministic selector-based login check endpoint at `POST /api/v1/credential-profiles/{credential_profile_id}/test-login`.
- Authenticated browser smoke endpoint at `POST /api/v1/projects/{project_id}/authenticated-browser-smoke-runs`.
- Browser worker support for decrypting credential profiles, filling configured selectors, clicking the configured submit selector, and checking configured success/failure criteria.
- Demo-web `/login` and protected `/dashboard` routes for deterministic authenticated smoke validation.
- Login summary support in JSON and HTML run reports.
- `login_observations` evidence metadata for login checks and authenticated browser smoke runs.
- Web UI credential profile section with create, edit, delete, default, login test, and authenticated smoke actions.
- Smoke validation for credential creation, login check, authenticated browser smoke, JSON/HTML reports, and password redaction.
- Go tests for credential profile validation, encryption, update preservation, and safe AI input redaction.

### Changed

- Browser worker now handles `login_check` and `authenticated_browser_smoke` run types in addition to browser smoke and safe test plan execution.
- AI-safe report input includes sanitized login metadata while excluding credentials, cookies, browser storage, auth headers, and tokens.
- Package metadata has been updated to `0.9.0-alpha`.

### Security

- Raw usernames and passwords are never returned by credential profile API responses.
- Credential profile username/password values are encrypted at rest with `QUALORA_ENCRYPTION_KEY`.
- Login automation is deterministic and selector-based; it does not use AI browser control or arbitrary form submission.
- Authenticated browser evidence does not expose cookies, local storage, session storage, authorization headers, tokens, or raw credentials.

### Known Limitations

- No Qualora user authentication or authorization.
- Authenticated browser smoke supports one configured login form and one same-origin target path per run.
- No MFA, role switching, session export, arbitrary form submission, broad crawling, or multi-step authenticated journeys.
- No authenticated API testing.
- No active security scanning or destructive testing.

## [v0.8.0-alpha] - 2026-07-14

### Added

- OpenAPI 3.x import from URL or inline JSON/YAML.
- API spec metadata and operation discovery tables.
- Safe API operation classifier with persisted skip reasons.
- Safe API smoke run endpoint at `POST /api/v1/api-specs/{api_spec_id}/api-smoke-runs`.
- API results endpoint at `GET /api/v1/runs/{run_id}/api-results`.
- JSON and HTML run report sections for API smoke results.
- Web UI API spec import, detail, operation list, and safe API smoke run workflows.
- Deterministic local `demo-api` service with `openapi.yaml`, safe GET endpoints, skipped unsafe operations, and `/broken` 500 finding.
- Go tests for OpenAPI parsing, safe operation classification, redirect blocking, 5xx findings, invalid JSON findings, and URL resolution.

### Changed

- Report and AI-safe input metadata now include API smoke summaries and sanitized API result summaries.
- Smoke flow now imports the demo OpenAPI spec and validates operation discovery, skipped unsafe operations, API results, API reports, and deterministic `/broken` findings.
- Package metadata has been updated to `0.8.0-alpha`.

### Security

- API smoke execution remains read-only and conservative.
- Mutating, authenticated, request-body, unresolved-parameter, and sensitive operations are skipped.
- Request bodies and response bodies are not stored or sent to AI.
- External redirects are not followed.

### Known Limitations

- No authenticated API testing.
- No schema fuzzing.
- No destructive API testing.
- OpenAPI import supports OpenAPI 3.x only.
- The web UI remains alpha and intended for trusted local/self-hosted environments.

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

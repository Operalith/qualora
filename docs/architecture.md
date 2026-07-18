# Architecture

Qualora v0.18.0-alpha is a small Docker Compose MVP for browser and safe API QA smoke runs with local first-run admin authentication, a minimal web UI, guided project onboarding, deterministic report intelligence, Safe QA report baselines, regression comparison, CI-friendly quality gates, safe deterministic application discovery, Interactive Safe Explorer, passive front-end quality checks, project-scoped credential profiles, deterministic selector-based login checks, authenticated browser smoke runs, explicit role-aware authorization checks, OpenAPI import and operation discovery, control-plane evidence download for stored artifacts, optional AI analysis of completed reports, discovery-aware AI test plan suggestions, Safe QA Runs, and approved safe execution of supported test plan steps.

## Runtime Components

```text
API client / smoke script / qualora-web
        |
        v
qualora-api
        |
        +--> PostgreSQL: local_users, user_sessions, projects, credential_profiles, discovery_runs, discovered_pages, discovered_links, discovered_forms, safe_explorer_runs, safe_explorer_steps, safe_explorer_actions, quality_check_runs, quality_check_results, authorization_checks, authorization_check_runs, authorization_check_results, test_runs, run_jobs, findings, evidence, api_specs, api_operations, api_check_results, ai_providers, ai_analyses, test_plans, test_plan_executions, qa_runs, report_baselines
        +--> Redis: browser, API, and test plan execution queues
        +--> MinIO/S3 evidence objects by evidence ID
        +--> Optional OpenAI-compatible AI provider
        |       +--> Report analysis
        |       +--> Test plan suggestions
        |
        +--> Deterministic report intelligence, baselines, comparisons, and quality gates
        |
        +--> qualora-worker-browser
        |       +--> Playwright Chromium smoke test
        |       +--> Deterministic selector-based login checks
        |       +--> Authenticated browser smoke test
        |       +--> Safe deterministic application discovery
        |       +--> Interactive Safe Explorer
        |       +--> Passive quality checks
        |       +--> Explicit role-aware authorization checks
        |       +--> Approved safe test plan execution
        |       +--> MinIO/S3 screenshot evidence
        |
        +--> qualora-worker-api
        |       +--> Legacy project API base URL/OpenAPI checks
        |
        +--> Control-plane API smoke executor
                +--> Imported OpenAPI 3.x operation discovery
                +--> Safe GET/HEAD/OPTIONS checks only
                +--> API check results, evidence, findings, and reports
```

## Services

### `qualora-api`

The Go control plane exposes the HTTP API, validates project scope, persists metadata, creates per-run jobs, queues worker jobs, renders JSON/HTML reports, and enforces local session authentication for protected API routes.

Report generation includes a deterministic intelligence layer. It normalizes finding severities, computes stable fingerprints, groups repeated findings, classifies noisy/repeated signals, summarizes affected pages, and produces executive summaries at report read/render time without changing stored finding schemas. The raw findings, evidence metadata, quality result rows, discovery maps, Safe Explorer traces, authorization results, and API result rows remain available.

The v0.18 baseline layer stores grouped finding fingerprints and summary metadata from known reports in `report_baselines`. Comparisons are computed deterministically from baseline fingerprints and the current report intelligence. Quality gate evaluation uses comparison summaries and current severity counts; it does not require AI and does not execute any new testing engine.

Current endpoints:

- `GET /healthz`
- `GET /api/v1/setup/status`
- `POST /api/v1/setup/admin`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `POST /api/v1/onboarding/project-setup`
- `POST /api/v1/projects`
- `GET /api/v1/projects`
- `GET /api/v1/projects/{project_id}`
- `POST /api/v1/projects/{project_id}/runs`
- `GET /api/v1/projects/{project_id}/runs`
- `POST /api/v1/projects/{project_id}/browser-smoke-runs`
- `POST /api/v1/projects/{project_id}/authenticated-browser-smoke-runs`
- `GET /api/v1/projects/{project_id}/credential-profiles`
- `POST /api/v1/projects/{project_id}/credential-profiles`
- `GET /api/v1/credential-profiles/{credential_profile_id}`
- `PUT /api/v1/credential-profiles/{credential_profile_id}`
- `DELETE /api/v1/credential-profiles/{credential_profile_id}`
- `POST /api/v1/credential-profiles/{credential_profile_id}/test-login`
- `GET /api/v1/projects/{project_id}/authorization-checks`
- `POST /api/v1/projects/{project_id}/authorization-checks`
- `GET /api/v1/authorization-checks/{authorization_check_id}`
- `PUT /api/v1/authorization-checks/{authorization_check_id}`
- `DELETE /api/v1/authorization-checks/{authorization_check_id}`
- `GET /api/v1/projects/{project_id}/authorization-check-runs`
- `POST /api/v1/projects/{project_id}/authorization-check-runs`
- `GET /api/v1/projects/{project_id}/discovery-runs`
- `POST /api/v1/projects/{project_id}/discovery-runs`
- `GET /api/v1/projects/{project_id}/quality-check-runs`
- `POST /api/v1/projects/{project_id}/quality-check-runs`
- `GET /api/v1/quality-check-runs/{quality_check_run_id}`
- `GET /api/v1/quality-check-runs/{quality_check_run_id}/report`
- `GET /api/v1/quality-check-runs/{quality_check_run_id}/report.html`
- `GET /api/v1/authorization-check-runs/{authorization_check_run_id}`
- `GET /api/v1/authorization-check-runs/{authorization_check_run_id}/report`
- `GET /api/v1/authorization-check-runs/{authorization_check_run_id}/report.html`
- `GET /api/v1/discovery-runs/{discovery_run_id}`
- `GET /api/v1/discovery-runs/{discovery_run_id}/map`
- `GET /api/v1/discovery-runs/{discovery_run_id}/report`
- `GET /api/v1/discovery-runs/{discovery_run_id}/report.html`
- `GET /api/v1/projects/{project_id}/safe-explorer-runs`
- `POST /api/v1/projects/{project_id}/safe-explorer-runs`
- `GET /api/v1/safe-explorer-runs/{safe_explorer_run_id}`
- `GET /api/v1/safe-explorer-runs/{safe_explorer_run_id}/trace`
- `GET /api/v1/safe-explorer-runs/{safe_explorer_run_id}/report`
- `GET /api/v1/safe-explorer-runs/{safe_explorer_run_id}/report.html`
- `GET /api/v1/projects/{project_id}/qa-runs`
- `POST /api/v1/projects/{project_id}/qa-runs`
- `GET /api/v1/qa-runs/{qa_run_id}`
- `POST /api/v1/qa-runs/{qa_run_id}/execute`
- `GET /api/v1/qa-runs/{qa_run_id}/report`
- `GET /api/v1/qa-runs/{qa_run_id}/report.html`
- `GET /api/v1/projects/{project_id}/report-baselines`
- `POST /api/v1/projects/{project_id}/report-baselines`
- `GET /api/v1/report-baselines/{baseline_id}`
- `PUT /api/v1/report-baselines/{baseline_id}`
- `DELETE /api/v1/report-baselines/{baseline_id}`
- `POST /api/v1/projects/{project_id}/report-comparisons`
- `POST /api/v1/projects/{project_id}/quality-gates/evaluate`
- `POST /api/v1/projects/{project_id}/ai-test-plans`
- `GET /api/v1/projects/{project_id}/test-plans`
- `POST /api/v1/projects/{project_id}/api-specs`
- `GET /api/v1/projects/{project_id}/api-specs`
- `GET /api/v1/api-specs/{api_spec_id}`
- `DELETE /api/v1/api-specs/{api_spec_id}`
- `GET /api/v1/api-specs/{api_spec_id}/operations`
- `POST /api/v1/api-specs/{api_spec_id}/api-smoke-runs`
- `GET /api/v1/runs`
- `GET /api/v1/runs/{run_id}`
- `GET /api/v1/runs/{run_id}/report`
- `GET /api/v1/runs/{run_id}/report.html`
- `GET /api/v1/runs/{run_id}/api-results`
- `GET /api/v1/evidence/{evidence_id}`
- `GET /api/v1/ai/providers`
- `POST /api/v1/ai/providers`
- `GET /api/v1/ai/providers/{provider_id}`
- `PUT /api/v1/ai/providers/{provider_id}`
- `DELETE /api/v1/ai/providers/{provider_id}`
- `POST /api/v1/ai/providers/{provider_id}/test`
- `GET /api/v1/runs/{run_id}/ai-analysis`
- `POST /api/v1/runs/{run_id}/ai-analysis`
- `GET /api/v1/test-plans/{test_plan_id}`
- `DELETE /api/v1/test-plans/{test_plan_id}`
- `GET /api/v1/test-plans/{test_plan_id}/export.json`
- `GET /api/v1/test-plans/{test_plan_id}/executions`
- `POST /api/v1/test-plans/{test_plan_id}/executions`
- `GET /api/v1/test-plan-executions/{execution_id}`
- `GET /api/v1/test-plan-executions/{execution_id}/report`
- `GET /api/v1/test-plan-executions/{execution_id}/report.html`

### `qualora-web`

The React/Vite web UI is intentionally small. It calls the control-plane API from the browser and displays:

- Dashboard quick-start actions, status indicators, recent projects, and recent Safe QA runs.
- A guided project setup wizard for project basics, optional AI, optional credentials, optional OpenAPI import, workflow selection, and result links.
- Project readiness checklist items that point users to AI, discovery, quality checks, credentials, OpenAPI, Safe QA, and reports.
- A reports landing page for recent browser, API, discovery, quality, and Safe QA reports.
- Projects and project details.
- Run lists and run details.
- Structured JSON report data as readable tables.
- Findings, evidence metadata, browser metadata, API metadata, and job metadata.
- Links to the self-contained HTML report export.
- Screenshot previews and downloads through the control-plane evidence endpoint.
- AI provider configuration for OpenAI-compatible endpoints.
- AI analysis status, summaries, risk level, recommendations, suggested next tests, and limitations.
- AI test plan generation from project/run context.
- AI test plan lists, detail pages, scenario/step display, deletion, and JSON export links.
- Safe execution preview controls for supported test plan steps.
- Safe execution history, detail pages, findings, evidence, and HTML report links.
- OpenAPI spec import, operation discovery, skip reasons, and safe API smoke run controls.
- API smoke result tables in run reports.
- Credential profile creation, listing, editing, deletion, default selection, login testing, and authenticated browser smoke actions.
- Login summary and login observation metadata in run reports.
- Authorization check creation, listing, enable/disable, deletion, run history, JSON report display, HTML report links, findings, and evidence display.
- Application discovery start form, run list, application map page, JSON/HTML report links, pages, skipped links, forms, findings, and evidence display.
- Quality Checks form, run list, JSON/HTML report links, category/severity summaries, findings, safety notes, and limitations.
- Discovery-aware AI test plan generation controls and safe executable coverage display.
- Safe QA Run preview/execution controls and Safe QA Run JSON/HTML report pages, including quality summaries when requested.
- Safe QA report baseline actions, comparison summaries, quality gate status, and failed rule display.
- Project-level Baselines & Regression card and reports index baseline indicators.

On a fresh database it shows a first-run local admin setup screen. After setup, project data, credential profiles, AI providers, runs, reports, evidence, API specs, test plans, and authorization reports require the local admin session. The UI is still alpha and should be exposed only in trusted local/self-hosted environments.

### Guided Project Onboarding

`POST /api/v1/onboarding/project-setup` is a thin orchestration endpoint used by the web wizard and smoke script. It validates the project request, rejects destructive actions, creates the project, optionally creates or reuses an OpenAI-compatible provider, optionally creates an encrypted credential profile, optionally imports an OpenAPI spec, and starts selected safe workflows.

The endpoint returns only resource IDs, safe provider metadata, skipped-action reasons, timeline entries, and report links. It must not return raw passwords, encrypted secret payloads, provider API keys, cookies, browser storage, authorization headers, or tokens. Guided onboarding does not add a new testing engine; it coordinates existing browser smoke, authenticated smoke, discovery, quality check, Safe QA, and imported-spec API smoke paths.

### Baselines, Comparisons, And Quality Gates

`report_baselines` is the only new durable model in v0.18. It stores project/report ownership, baseline metadata, default-baseline status, normalized grouped finding fingerprints, severity counts, grouped finding count, and raw finding count. A partial unique index keeps one default baseline per project and report type.

Report comparison is computed on demand. The control plane loads the baseline snapshot and the current report snapshot, extracts grouped finding fingerprints from existing report intelligence, then classifies findings as new, fixed, unchanged, severity-changed, or affected-scope-changed. No worker job, AI call, browser action, API request, or security scan is started by comparison.

Quality gates are also synchronous control-plane evaluation. Defaults fail on new critical/high findings and total critical findings, warn when no baseline exists, and return `ci_exit_code` for automation. The `format=ci` response shape is intentionally compact for curl-based CI usage. This is an alpha HTTP/script integration, not a full CLI.

### `qualora-worker-browser`

The Node.js browser worker consumes Redis browser jobs and runs a Playwright smoke check against `frontend_url`.

For credential-profile runs, the worker decrypts the project-scoped username/password with `QUALORA_ENCRYPTION_KEY`, opens only the configured login URL, fills only the configured username/password selectors, clicks only the configured submit selector, and evaluates configured success/failure criteria. Authenticated smoke then visits one configured same-origin target path.

For role-aware authorization runs, the worker logs in with the configured actor credential profile, navigates only to the configured same-origin authorization target, classifies the outcome as allowed, denied, or unknown, and records screenshot/observation evidence plus deterministic findings when the observed outcome does not match the expected outcome. It does not use AI, does not crawl, does not submit arbitrary forms, and does not expose cookies, storage, auth headers, tokens, usernames, or passwords in evidence.

For application discovery runs, the worker performs bounded deterministic navigation from the project frontend or requested start URL. It defaults to same-origin only, enforces `allowed_hosts`, strips fragments, redacts sensitive query values, avoids duplicate visits, records pages/links/forms/fields, stores screenshot evidence, and creates findings for obvious load, console, network, skipped-link, and form issues. It never submits forms, clicks arbitrary buttons, runs payloads, performs destructive actions, or uses AI browser control.

For Interactive Safe Explorer runs, the worker logs in only through an optional configured credential profile, starts from the project frontend or requested start URL, observes visible links/buttons/forms/inputs, classifies actions with a deterministic safety policy, and executes only safe same-origin navigation actions. It records a step timeline, action metadata, screenshots, findings, and skip reasons for unsafe, external, unsupported, duplicate, sensitive-query, and policy-blocked actions. It does not let AI choose actions, does not submit POST forms, does not click arbitrary buttons, does not expose cookies/storage/auth headers/tokens, and does not store full HTML or response bodies.

For quality check runs, the worker visits only the project frontend origin or pages from a completed discovery run. It can optionally perform the same deterministic selector-based credential-profile login before checking pages. It collects safe metadata for passive security checks, accessibility heuristics, and performance/front-end observations. Quality evidence stores metadata only; it must not contain cookie values, browser storage, auth headers, tokens, credentials, request bodies, response bodies, or full HTML. Quality checks never submit forms, click arbitrary buttons, guess sensitive paths, run payloads, fuzz inputs, perform active scans, perform destructive actions, or use autonomous AI browser control.

The same worker also consumes safe test plan execution jobs. It executes only persisted mapped actions from the supported DSL:

- `goto`
- `assert_title_contains`
- `assert_url_contains`
- `assert_text_visible`
- `assert_element_visible`
- `assert_link_exists`
- `check_link_status`
- `capture_screenshot`
- `collect_browser_signals`
- `wait_for_load_state`
- `assert_no_console_errors`
- `assert_no_failed_requests`

It does not execute free-form model text. It revalidates same-origin frontend targets, enforces `allowed_hosts`, records step pass/fail state, captures screenshot evidence, records browser observations, and creates findings for failed safe execution steps.

It currently captures:

- Page title.
- Final URL.
- Initial document status code.
- Body text length.
- Screenshot.
- Console errors.
- Failed network requests.
- Blocked out-of-scope browser requests.
- Safe Explorer observed pages, executed actions, skipped actions, selector hints, skip reasons, action safety decisions, and screenshots.
- Basic findings for obvious load, timeout, non-success status, console, request, empty page, and scope issues.
- Login findings for failed login, missing selectors, timeouts, console errors, failed requests, and authenticated navigation failures.
- Quality findings for missing security headers, cookie flag observations, mixed content, sensitive query names, source maps, password/form issues, basic accessibility issues, slow loads, console errors, failed resources, request-count issues, large JavaScript resources, and image dimension issues.

Login evidence is stored as `login_observations` metadata plus screenshots when configured. JSON and HTML run reports include a `login_summary` for login checks and authenticated browser smoke runs.

### Safe API Smoke Execution

The imported-spec API smoke flow runs in the Go control plane. It imports OpenAPI 3.x specs from project-scoped URLs or inline JSON/YAML, stores discovered operations, classifies safe operations, and persists skip reasons before any API test execution.

Safe by default means:

- Only `GET`, `HEAD`, and `OPTIONS` operations are eligible.
- `POST`, `PUT`, `PATCH`, `DELETE`, and `TRACE` are skipped.
- Authenticated operations are skipped.
- Operations with required request bodies are skipped.
- Operations with unresolved path parameters are skipped unless a safe `example`, `default`, or `enum` value exists.
- Required query parameters are sent only when a safe sample value exists.
- Sensitive paths or query parameter names are skipped.
- Redirects to external origins are not followed.
- Response bodies and request bodies are not stored.

API smoke reports include `api_results`, `api_summary`, API findings, `api_observations`, `openapi_summary`, and per-request `api_request` evidence metadata.

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

This worker remains available for legacy project-level API jobs. New imported-spec API smoke runs use the control-plane executor so operation discovery and result rows are first-class API/UI concepts.

### Local Authentication

Qualora stores one local admin user and session records in PostgreSQL for this alpha. Passwords are hashed with Argon2id. Sessions use an HTTP-only `qualora_session` cookie plus a separate `qualora_csrf` cookie that must match the `X-Qualora-CSRF` header for mutating protected API requests.

Public endpoints are limited to health, setup status, first-run admin setup, login, logout, and session introspection. All project, credential, AI, evidence, report, API spec, authorization, and test plan endpoints are protected after setup.

This is intentionally not full identity management: there is no user management UI, password reset flow, SSO/OIDC/SAML, multi-role RBAC, teams, or multi-tenancy in `v0.18.0-alpha`.

### PostgreSQL

PostgreSQL stores durable metadata:

- `local_users`
- `user_sessions`
- `projects`
- `credential_profiles`
- `discovery_runs`
- `discovered_pages`
- `discovered_links`
- `discovered_forms`
- `quality_check_runs`
- `quality_check_results`
- `authorization_checks`
- `authorization_check_runs`
- `authorization_check_results`
- `test_runs`
- `run_jobs`
- `findings`
- `evidence`
- `api_specs`
- `api_operations`
- `api_check_results`
- `ai_providers`
- `ai_analyses`
- `test_plans`
- `test_plan_executions`
- `test_plan_execution_scenarios`
- `test_plan_execution_steps`
- `qa_runs`

Reports are generated dynamically from run, job, finding, and evidence rows.

### Redis

Redis is the MVP queue for browser jobs, API jobs, discovery jobs, quality check jobs, authorization jobs, and safe test plan execution jobs. PostgreSQL remains the source of durable run and execution state.

### MinIO

MinIO stores screenshot evidence through the S3-compatible API. API evidence is stored as metadata rows in PostgreSQL. The control plane can stream stored evidence objects by evidence ID so the UI can preview/download screenshots without exposing MinIO credentials.

### `demo-api`

The `demo-api` Compose service is profile-gated for smoke tests. It serves deterministic API endpoints and an OpenAPI 3.x document at `/openapi.yaml` for OpenAPI import and safe API smoke validation.

### `demo-web`

The `demo-web` Compose service is profile-gated for smoke tests. It serves public pages for browser and safe test plan execution validation, deterministic passive quality issues, plus a deterministic `/login` and protected `/dashboard` route for credential profile and authenticated browser smoke validation.

### `mock-api`

The `mock-api` Compose service is profile-gated for smoke tests. It is not part of the production runtime path.

### `fake-llm`

The `fake-llm` Compose service is profile-gated for smoke tests. It implements the OpenAI-compatible `/v1/chat/completions` shape and returns deterministic JSON analysis or deterministic JSON test plans so tests never call an external LLM provider.

## Optional AI Features

AI is an enhancement layer, not a dependency for QA execution. Browser/API workers produce deterministic findings and evidence first. A user can then run AI analysis for an existing run or generate an AI-assisted test plan for a project.

Provider records store:

- OpenAI-compatible base URL.
- Model name.
- API key encrypted at rest.
- Optional extra headers encrypted at rest.
- Safe-send toggles for screenshots, HTML, and network bodies.
- Redaction setting, enabled by default.

For this alpha, AI analysis and AI test planning run synchronously in the control plane. The database models are separated so both paths can move to a dedicated analyzer worker later.

The safe AI input builder includes only sanitized report data such as run status, summary counts, finding titles/summaries, safe evidence metadata, browser/API/login metadata, API smoke result summaries, and job metadata. It strips or redacts URL queries, cookies, authorization values, usernames, tokens, passwords, API keys, session IDs, JWT-looking strings, full response bodies, full HTML, and secret-looking fields. Screenshots, HTML, request bodies, response bodies, browser storage, auth headers, cookies, and network bodies are not sent by default.

AI-assisted test plans use sanitized project configuration, optional product context, selected focus areas, optional latest/run-specific report metadata, optional discovery map summaries, and optional AI analysis summaries. The strict plan parser accepts only a reviewable JSON structure with assumptions, coverage goals, scenarios, steps, assertions, test data needs, instrumentation suggestions, limitations, and optional deterministic safe DSL candidates.

Generated plans are not executed automatically. A user can explicitly preview and start safe execution. The deterministic mapper only queues scenarios marked `automation_candidate=true`, `destructive=false`, and `requires_authentication=false`; skips unsafe terms such as login, payment, submit, upload, mutation, admin, SQLi, XSS, SSRF, brute force, and destructive actions; and maps only the supported browser DSL. Unsupported or ambiguous steps are persisted as skipped with reasons.

### Safe QA Runs

Safe QA Runs combine the existing safe pieces into a single alpha workflow:

1. Reuse a completed discovery run, use the latest completed discovery run, or create a new bounded discovery run.
2. Generate a discovery-aware AI test plan from sanitized project/report/discovery metadata.
3. Build and persist the deterministic safe execution preview.
4. Stop at review by default, or execute the approved safe DSL path when the user explicitly requests it.
5. Serve a Safe QA Run JSON/HTML report linking the discovery run, generated plan, preview coverage, optional execution report, and safety notes.

The workflow is orchestration, not free-form autonomy. It does not give an LLM browser control, does not execute model text, and does not allow arbitrary clicking, form submission, crawling, payloads, or destructive actions.

## Run Lifecycle

1. A client or web UI creates a project with one or more targets: `frontend_url`, `api_base_url`, or `openapi_url`.
2. The API validates URL scope and target safety.
3. A client starts a run.
4. The API creates a `queued` run in PostgreSQL.
5. The API creates `run_jobs` for browser and/or API work.
6. The API pushes worker jobs to Redis.
7. Workers mark their jobs `running`.
8. Workers collect evidence and findings.
9. Workers mark jobs `completed` or `failed`.
10. PostgreSQL refreshes the parent run status from job statuses.
11. The API serves the structured JSON report and the self-contained HTML report.
12. The API streams screenshot evidence by evidence ID when requested.
13. Optionally, a user runs AI analysis for the completed run.
14. The API stores the AI analysis and includes it in JSON/HTML reports.
15. Optionally, a user generates an AI-assisted test plan for a project using sanitized project/run context.
16. The API stores the test plan and includes a lightweight reference in JSON/HTML reports when it was generated from a run.
17. Optionally, a user previews safe execution for a test plan.
18. The API maps only supported safe steps and returns executable/skipped counts and reasons.
19. Optionally, a user starts the approved safe execution.
20. The API stores execution/scenario/step rows and queues the browser worker.
21. The browser worker executes persisted safe actions, writes evidence/findings, and updates step/scenario/execution status.
22. The API serves JSON and HTML test plan execution reports.
23. Optionally, a user imports an OpenAPI spec for a project.
24. The API stores spec metadata and discovered operations without executing them.
25. Optionally, a user starts a safe API smoke run for the imported spec.
26. The control plane executes only safe read-only operations, records API results/evidence/findings, and serves JSON/HTML run reports with API result tables.
27. Optionally, a user creates a credential profile for a project with encrypted username/password values and deterministic login selectors.
28. Optionally, a user starts a login check run for that profile.
29. The browser worker executes the configured selector-based login flow, records login evidence/findings, and serves JSON/HTML reports with a login summary.
30. Optionally, a user starts an authenticated browser smoke run.
31. The browser worker logs in through the credential profile, visits one configured same-origin target path, records browser/login evidence and findings, and serves JSON/HTML reports.
32. Optionally, a user starts a Safe QA Run.
33. The API reuses or creates a discovery run, generates a discovery-aware test plan from sanitized metadata, stores safe execution coverage, and either stops for review or explicitly starts the approved safe execution.
34. The API serves JSON and HTML Safe QA Run reports with discovery, plan, preview, execution, and safety metadata.

## Intentional Alpha Constraints

- Local authentication is intentionally minimal: one admin role, no user management UI, no password reset, no SSO/OIDC/SAML, no enterprise RBAC, no teams, and no multi-tenancy.
- Web UI is alpha and suitable only for trusted self-hosted/local environments.
- AI provider management is alpha and should be used only in trusted local/self-hosted environments.
- AI-assisted test planning is alpha and should be treated as human-reviewable suggestions.
- Safe test plan execution is alpha and limited to approved non-destructive browser DSL steps.
- API smoke execution is alpha and read-only by default.
- Credential profiles and authenticated browser smoke are alpha and intended for deterministic test accounts only.
- Authenticated API testing is not supported.
- Request bodies and response bodies are not stored.
- Only OpenAI-compatible chat completion providers are supported.
- Evidence download is limited to stored evidence records and is not a signed URL system.
- No authenticated API testing.
- No arbitrary form submission, MFA, session export, multi-step authenticated journeys, or AI-controlled login.
- No autonomous AI browser control.
- No automatic execution of generated AI test plan steps and no free-form model-controlled browser actions. Safe QA execution requires an explicit user request and still runs only persisted safe DSL steps.
- No active security scanning.
- No destructive API testing by default.
- No full OpenAPI schema validation or fuzzing.
- No Playwright trace export yet.
- No Kubernetes or Helm support yet.

The older MVP notes remain in [architecture/mvp.md](architecture/mvp.md) for implementation context, but this document is the release-facing architecture reference.

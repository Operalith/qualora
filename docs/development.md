# Development

This document covers local development for Qualora v0.24.0-alpha.

## Requirements

- Docker with Docker Compose.
- Go 1.24 or newer for control plane development.
- Node.js 22 or newer for web, browser worker, and API worker development.
- Python 3 for the smoke script.

## Common Commands

```bash
make dev
make test
make lint
make compose-up
make compose-down
make logs
make smoke
make showcase-smoke
make demo-lab
make demo-lab-real-llm
```

Command behavior:

- `make dev`: starts the Docker Compose stack.
- `make test`: runs Go tests, web build/type-check, and browser/API worker tests.
- `make lint`: runs the same checks plus `docker compose config`.
- `make compose-up`: runs `docker compose up -d --build`.
- `make compose-down`: runs `docker compose down`.
- `make logs`: tails API, web, browser worker, and API worker logs.
- `make smoke`: starts the local demo web, demo API, and fake LLM profile services; performs first-run local admin setup/login/logout checks; creates an AI provider, browser project, API project, credential profiles, API auth profiles, and role-aware authorization checks; imports the demo OpenAPI spec; exercises guided project setup; starts browser, login check, authenticated browser smoke, application discovery, Interactive Safe Explorer, Safe Form Testing, passive quality, authorization, Safe QA, unauthenticated API smoke, and authenticated API smoke plus contract validation runs; polls to completion; runs AI analysis; generates run-based and discovery-aware AI test plans; previews and executes safe browser test plans; creates a Safe QA baseline; compares a second Safe QA report; evaluates a quality gate and CI compact response; runs the native CI run endpoint; runs `scripts/qualora-ci-gate.sh` and `scripts/qualora-ci-run.sh`; configures fake issue export metadata; dry-runs grouped finding issue export; exercises policy-approved and policy-blocked AI Browser Control form suggestions; prints JSON/HTML report, discovery map, Safe Explorer trace, Safe Form Testing report, API spec, credential profile, API auth profile, quality, authorization, test-plan, execution, Safe QA Run, CI run, project setup, and report index URLs; validates HTML report export; validates protected report/evidence access; validates API result rows; validates skipped unsafe API operations; validates skipped discovery links; validates Safe Explorer executed/skipped action reasons; validates Safe Form Testing tested/skipped form reasons; validates quality finding counts; validates credential and API token redaction; validates test-plan export; validates issue preview redaction; and validates screenshot evidence download.
- `make showcase-smoke`: starts the `demo-lab` targets and fake LLM, then runs the same mature end-to-end assertions against the comprehensive Demo Lab fixtures.
- `make demo-lab`: runs the deterministic Fake LLM showcase and prints project and report links.
- `make demo-lab-real-llm`: validates the required real-provider environment, starts Demo Lab, and runs discovery, AI Browser Control, test planning, and Safe QA with that provider. It is opt-in and is never part of smoke or CI.
- `scripts/run-demo-lab.sh`: starts the entire core plus `demo-lab` profile, runs showcase smoke, and prints target plus run-specific report links.

## Start The Stack

```bash
docker compose up -d --build
```

If port `8080` is already in use:

```bash
QUALORA_API_PORT=18080 docker compose up -d --build
QUALORA_API_URL=http://localhost:18080 QUALORA_API_BASE_URL=http://localhost:18080 make smoke
```

The web UI is served on:

```text
http://localhost:3000
```

The dashboard's `Run Demo Lab Showcase` action creates a deterministic local setup with Fake LLM. Open the resulting Project Cockpit to start common QA workflows. AI Browser Control and Safe Explorer rows link to the Run Viewer, which polls existing report endpoints while active and replays the same step/evidence data after completion.

For a real OpenAI-compatible provider, see [real-llm-demo.md](real-llm-demo.md). Real requests may incur provider cost. Never add real-provider execution to deterministic tests.

## Run Tests

```bash
make test
make lint
```

Backend only:

```bash
cd apps/control-plane
go test ./...
go run .
```

Web UI only:

```bash
cd apps/web
npm ci
npm run dev
npm run build
```

Browser worker only:

```bash
cd workers/browser
npm ci
npm run build
npm test
```

API worker only:

```bash
cd workers/api
npm ci
npm run build
npm test
```

## Smoke Test

With the Compose stack running:

```bash
make smoke
```

The smoke script runs:

- First-run local admin setup, login, logout, `/auth/me`, CSRF, and protected endpoint checks.
- AI provider creation and provider-test against local `fake-llm`.
- Credential profile creation against local `demo-web` login selectors.
- Deterministic login check against local `demo-web`.
- Authenticated browser smoke against local `demo-web` `/dashboard`.
- Application discovery against local `demo-web`, including pages, links, forms, skipped unsafe/external links, screenshots, JSON report, and HTML report.
- Safe Form Testing against local `demo-web`, including safe same-origin GET form execution, skipped unsafe/external/mutating forms, screenshot evidence, JSON report, HTML report, report intelligence, and no raw test-value storage.
- AI Browser Control safe-form and unsafe-form policy fixtures through local `fake-llm`.
- Passive quality checks against local `demo-web`, including deterministic security, accessibility, and performance/front-end findings plus JSON/HTML reports.
- Role credential profile creation and explicit authorization checks against local `demo-web` `/admin` and customer invoice routes.
- Password and raw username redaction checks for credential, report, and AI paths.
- Browser smoke against the local `demo-web` Compose service.
- AI analysis for the completed browser smoke run.
- AI test plan generation/export validation for the browser smoke run.
- Discovery-aware AI test plan generation/export validation from the completed application map.
- Safe test plan execution preview and run validation for the browser smoke project.
- Safe QA Run preview and explicit execution validation.
- Safe QA Run quality summary validation when quality checks are included.
- Safe QA baseline creation from the first report.
- Second Safe QA report comparison against the baseline.
- Quality gate evaluation with human JSON and compact CI JSON.
- Guided project setup validation with demo AI provider, demo OpenAPI import, demo credential profile, browser smoke, authenticated smoke, discovery, quality checks, Safe QA, and API smoke.
- Dashboard, guided wizard, reports, and project readiness UI bundle text validation.
- OpenAPI import and safe API smoke against the local `demo-api` Compose service.
- Operation discovery validation, including skipped POST/DELETE/auth-required operations.
- Deterministic `/broken` API finding validation.
- AI analysis and AI test plan generation/export validation for the API smoke run.
- HTML report export validation for each completed run.
- Screenshot evidence metadata and download validation for the browser run.

Override browser target:

```bash
QUALORA_TARGET_URL=http://demo-web:8080 \
QUALORA_ALLOWED_HOST=demo-web \
make smoke
```

Override API target:

```bash
QUALORA_API_SMOKE_URL=http://demo-api:8080 \
QUALORA_API_SMOKE_OPENAPI_URL=http://demo-api:8080/openapi.yaml \
QUALORA_API_SMOKE_ALLOWED_HOST=demo-api \
make smoke
```

For private or local targets, create projects manually with `allow_private_targets: true` only when testing systems you control.

## Guided Onboarding Development

The guided setup flow is intentionally an orchestration layer over existing APIs. The backend endpoint is `POST /api/v1/onboarding/project-setup`; the web route is `#/setup-project`.

When changing guided onboarding:

- Keep `destructive_actions=false`.
- Do not add autonomous browser control or new test engines.
- Keep AI optional.
- Keep Safe QA disabled or skipped when no provider is available.
- Return only safe metadata, IDs, skipped reasons, timeline entries, and report links.
- Never return raw credential values, encrypted secret payloads, API keys, cookies, browser storage, authorization headers, or tokens.
- Update `scripts/smoke.py` so the deterministic demo setup still covers the full first-run path.

## AI Provider Development

The v0.21 AI path uses OpenAI-compatible chat completions only. AI analysis and AI-assisted test planning are optional and run synchronously in the control plane for this alpha.

Useful local values:

```text
QUALORA_ENCRYPTION_KEY=qualora-insecure-dev-key-change-me
QUALORA_SESSION_TTL_HOURS=12
QUALORA_COOKIE_SECURE=false
QUALORA_AUTH_DISABLED=false
QUALORA_FAKE_LLM_URL=http://fake-llm:8080/v1
QUALORA_ADMIN_EMAIL=admin@qualora.local
QUALORA_ADMIN_PASSWORD=qualora-admin-password
QUALORA_DEMO_USERNAME=demo@example.com
QUALORA_DEMO_PASSWORD=demo-password
FAKE_LLM_HEALTH_URL=http://localhost:18083/health
```

The default Compose encryption key is intentionally insecure and only for local development. Set a strong `QUALORA_ENCRYPTION_KEY` before storing real provider credentials or credential profiles.

`QUALORA_AUTH_DISABLED=true` is available only as a local escape hatch for development/debugging. Do not use it for shared self-hosted environments. The default is authenticated.

AI-assisted test plans are reviewable suggestions and are never executed automatically. Qualora can include a sanitized discovery map when the user requests discovery-aware planning, but it still executes only explicitly approved, deterministic safe DSL steps after a preview. It does not send screenshots/full HTML/raw traces/full network bodies/cookies/browser storage/auth headers/credentials to AI by default, and it redacts secret-looking values before prompt construction and storage.

Safe test plan execution currently supports only browser actions that stay on the project frontend origin: `goto`, `assert_title_contains`, `assert_url_contains`, `assert_text_visible`, `assert_element_visible`, `assert_link_exists`, `check_link_status`, `capture_screenshot`, `collect_browser_signals`, `wait_for_load_state`, `assert_no_console_errors`, and `assert_no_failed_requests`.

## Credential Profile Development

Credential profiles are project-scoped and store username/password values encrypted at rest with `QUALORA_ENCRYPTION_KEY`. API responses return configured flags and a masked username display hint only; raw usernames and passwords must never be logged, returned, included in reports, or sent to AI.

The login check path is deterministic:

- The login URL must satisfy project `allowed_hosts` and match the project `frontend_url` origin.
- The browser worker fills only the configured username and password selectors.
- The worker clicks only the configured submit selector.
- Success is determined by configured URL/text criteria, with optional failure text detection.
- Authenticated browser smoke visits one relative same-origin target path after login.
- Cookies, local storage, session storage, auth headers, tokens, and browser storage are not exposed in evidence.

## Application Discovery Development

Application discovery runs are queued on the browser worker. Keep the crawl deterministic and bounded:

- Defaults are `max_pages=20`, `max_depth=2`, and `same_origin_only=true`.
- The API caps discovery at `max_pages<=100` and `max_depth<=5`.
- Navigation must pass `allowed_hosts`; same-origin discovery must stay on the project frontend origin.
- Discovery follows safe links only and records skip reasons for external, unsafe-looking, unsupported-scheme, and non-HTML links.
- Discovery records forms and fields but must never submit forms or click arbitrary buttons.
- Discovery evidence may include screenshots and browser observations, but must not store full HTML, cookies, local/session storage, auth headers, tokens, credentials, request bodies, or response bodies.

## Quality Check Development

Quality checks run on the browser worker and stay passive:

- Defaults are `max_pages=10`; the API caps quality checks at `max_pages<=50`.
- A quality run may check only the project frontend URL, reuse the latest completed discovery run, reuse a selected completed discovery run, or log in through a deterministic credential profile before checking pages.
- All page visits must stay on the project frontend origin and pass `allowed_hosts`.
- Security checks are passive observations from loaded pages, response headers, cookie metadata, forms, and resource metadata.
- Accessibility checks are lightweight heuristics for page title/lang/main, images, labels, button/link names, and similar obvious issues.
- Performance checks are lightweight page-load, console, failed-resource, request-count, large-JS, and image-dimension observations.
- Evidence stores safe metadata only. Do not store cookie values, browser storage, auth headers, secrets, request bodies, response bodies, full HTML, or credentials.
- Do not add active scanning, payloads, fuzzing, guessed-path probing, form submission, arbitrary clicks, destructive actions, or AI browser control to this path.

## Safe QA Run Development

Safe QA Runs are an orchestration layer over discovery, AI test planning, and approved safe test plan execution. Keep the workflow reviewable and deterministic:

- The API may reuse a completed discovery run, use the latest completed discovery run, or start a bounded discovery run.
- AI planning input must use sanitized project/report/discovery metadata only.
- Discovery-aware plans should be tagged with `source_type=discovery` and persist `discovery_run_id`.
- Safe executable coverage should be computed and stored from the deterministic preview.
- `execute=false` must stop after preview so a user can review the plan and skipped reasons.
- `execute=true` or `POST /api/v1/qa-runs/{qa_run_id}/execute` may start only the persisted safe DSL execution path.
- Reports should link discovery, test plan, preview, optional execution report, and safety notes without exposing secrets.
- When `include_quality_checks=true`, reports should also link the quality run and include quality summaries/results without exposing secrets.
- Do not add autonomous AI browser control, arbitrary clicking, form submission, broad crawling, payloads, active scans, or destructive actions to this workflow.

## Safe API Smoke Development

Imported OpenAPI specs are parsed without executing API requests. Safe API smoke execution starts only after a user calls `POST /api/v1/api-specs/{api_spec_id}/api-smoke-runs`.

The v0.21 API executor:

- Supports OpenAPI 3.x JSON/YAML.
- Executes only `GET`, `HEAD`, and `OPTIONS`.
- Skips mutating methods, required request bodies, unresolved path parameters, sensitive paths, and sensitive required query parameters.
- Skips auth-required operations unless the caller explicitly selects an enabled project API auth profile.
- Supports bearer token, API key header, API key query, basic auth, and explicit no-auth profiles.
- Stores API auth secrets encrypted at rest using `QUALORA_ENCRYPTION_KEY`.
- Sends auth material only to the configured allowed-host API target during safe read-only checks.
- Sends no cookies, request bodies, or unauthenticated secrets.
- Does not store response bodies.
- Does not store request bodies, response bodies, raw Authorization headers, token values, API keys, or basic auth values.
- Blocks redirects to external origins.
- Persists `api_check_results` plus metadata-only API evidence, including auth mode, expected/actual status, expected/actual content type, contract validation status, sanitized schema validation messages, and unauthenticated comparison status.
- Performs lightweight contract validation for status codes, obvious content types, JSON parseability, simple required fields, scalar/object/array types, enum, nullable, and simple array item types.

Relevant files:

- `apps/control-plane/api_auth_profile.go`
- `apps/control-plane/api_auth_profile_http.go`
- `apps/control-plane/api_contract.go`
- `apps/control-plane/api_smoke.go`
- `apps/control-plane/migrations/017_api_auth_contract.sql`
- `scripts/demo-api/openapi.yaml`
- `scripts/smoke.py`

## AI Browser Control Development

`v0.24.0-alpha` keeps policy-gated AI Browser Control in the browser worker and adds a policy-approved safe GET form action. The control plane creates `ai_browser_control_runs`, queues a Redis browser job, and serves run detail, trace, JSON report, and self-contained HTML report endpoints. The browser worker captures sanitized observations, asks the selected OpenAI-compatible provider for one strict JSON action, validates it through deterministic policy, and executes only approved safe Playwright actions.

Important files:

- `apps/control-plane/ai_browser_control.go`
- `apps/control-plane/ai_browser_control_http.go`
- `apps/control-plane/ai_browser_control_store.go`
- `apps/control-plane/ai_browser_control_report_html.go`
- `apps/control-plane/migrations/018_ai_browser_control.sql`
- `workers/browser/src/ai_browser_control.ts`
- `workers/browser/src/ai_browser_control.test.ts`
- `scripts/fake-llm/server.js`
- `scripts/smoke.py`

Use the fake LLM for deterministic local testing. Its default AI Browser Control path navigates `/`, `/about`, `/status`, `/pricing`, captures a screenshot once, then stops. A goal containing `force_unsafe_ai_browser_action` makes the fake provider propose `/delete-account`, which the policy engine should block. A goal containing `force_safe_ai_browser_form_action` makes it propose `submit_safe_get_form` for the demo search form, while `force_unsafe_ai_browser_form_action` proposes an unsafe form action that policy must block.

AI Browser Control must not send credentials, cookies, browser storage, auth headers, screenshots, full HTML, request bodies, response bodies, or raw form values to AI. Do not add arbitrary selectors, arbitrary form submission, POST/mutating form submission, active scanning, fuzzing, payload execution, external crawling, destructive actions, or direct model control of Playwright.

## Safe Form Testing Development

`v0.24.0-alpha` adds standalone Safe Form Testing through the browser worker. The control plane creates `form_test_runs`, queues a Redis browser job, persists `form_test_results`, and serves run detail plus JSON/HTML report endpoints.

Important files:

- `apps/control-plane/form_test.go`
- `apps/control-plane/form_test_http.go`
- `apps/control-plane/form_test_store.go`
- `apps/control-plane/form_test_report_html.go`
- `apps/control-plane/migrations/019_form_tests.sql`
- `workers/browser/src/form_testing.ts`
- `workers/browser/src/form_testing.test.ts`
- `workers/browser/src/index.ts`
- `scripts/demo-web/server.js`
- `scripts/fake-llm/server.js`
- `scripts/smoke.py`

Safe Form Testing executes only same-origin GET forms classified as safe search/filter/sort/navigation-style forms. It uses bounded deterministic values and records classifications, tested/skipped decisions, skip reasons, redacted submitted URLs, screenshots, findings, and safe value summaries.

Do not add POST/PUT/PATCH/DELETE submission, arbitrary form filling, fuzzing, payload generation, login/password/payment/upload/admin/account mutation flows, external form execution, or AI-controlled form execution. Reports, evidence, AI inputs, CI output, and issue export previews must never contain raw form values, credentials, cookies, browser storage, auth headers, tokens, request bodies, response bodies, or full HTML.

## Report Intelligence Development

`v0.24.0-alpha` computes report intelligence when JSON or HTML reports are read. The helper lives in `apps/control-plane/report_intelligence.go` and is intentionally storage-neutral: it maps existing findings and quality result rows into a normalized internal model, computes deterministic fingerprints, groups repeated findings, normalizes severity, classifies noisy/repeated signals, and builds executive summaries.

When adding new finding sources, provide stable categories, titles, recommendations, evidence IDs, and safe URL metadata where practical. Do not add secrets, cookies, local/session storage, auth headers, request bodies, response bodies, full HTML, or screenshot bytes to report intelligence inputs.

## Baseline And Quality Gate Development

The v0.22 baseline, CI, and quality gate paths are control-plane only:

- `POST /api/v1/projects/{project_id}/report-baselines` stores normalized grouped finding fingerprints and summary metadata from a known report.
- `GET /api/v1/projects/{project_id}/report-baselines?report_type=safe_qa` lists baselines for the project.
- `POST /api/v1/projects/{project_id}/report-comparisons` computes new, fixed, unchanged, severity-changed, and affected-scope-changed findings on demand.
- `POST /api/v1/projects/{project_id}/quality-gates/evaluate` evaluates deterministic gates and supports `?format=ci`.
- `POST /api/v1/projects/{project_id}/ci-runs` starts or reuses Safe QA, compares with a baseline, evaluates gates, persists `ci_runs`, and returns a deterministic exit code.
- `POST /api/v1/reports/{report_type}/{report_id}/export-issues` dry-runs or exports grouped sanitized findings to an optional configured tracker.

Keep comparisons deterministic and based on report intelligence fingerprints. Do not add AI calls, browser actions, API requests, active scans, payloads, fuzzing, destructive checks, or new workers to this path.

The existing-report HTTP CI helper is:

```bash
QUALORA_API_URL=http://localhost:8080 \
QUALORA_PROJECT_ID="$PROJECT_ID" \
QUALORA_REPORT_ID="$QA_RUN_ID" \
QUALORA_EMAIL=admin@qualora.local \
QUALORA_PASSWORD="$QUALORA_PASSWORD" \
scripts/qualora-ci-gate.sh
```

The run-and-gate HTTP CI helper is:

```bash
QUALORA_URL=http://localhost:8080 \
QUALORA_EMAIL=admin@qualora.local \
QUALORA_PASSWORD="$QUALORA_PASSWORD" \
QUALORA_PROJECT_ID="$PROJECT_ID" \
QUALORA_RUN_SAFE_QA=false \
QUALORA_ISSUE_EXPORT_DRY_RUN=true \
scripts/qualora-ci-run.sh
```

Both scripts exit with the API exit code. They are intentionally small and not a full CLI.

Issue export config tokens must be encrypted with `QUALORA_ENCRYPTION_KEY` and must never be returned raw. Dry-run issue export should be covered before any real tracker integration changes. Tests should use fake HTTP servers for GitHub/GitLab clients rather than calling real trackers.

OpenRouter example headers:

```json
{
  "HTTP-Referer": "http://localhost:3000",
  "X-OpenRouter-Title": "Qualora"
}
```

## Clean Up

Stop containers:

```bash
docker compose down
```

Stop containers and delete local volumes:

```bash
docker compose down -v
```

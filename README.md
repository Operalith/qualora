# Qualora

**Open-source, self-hosted autonomous QA for web applications and APIs.**

Qualora is an open-source, self-hosted autonomous QA platform that runs browser-based and API smoke tests, collects evidence, and generates structured reports for web applications and APIs.

`v0.17.0-alpha` adds deterministic report intelligence: executive summaries, severity normalization, grouped findings, duplicate reduction, affected-page summaries, and noise/repeated-finding indicators across JSON reports, HTML exports, the reports landing page, and report detail pages. Qualora remains useful without AI: discovery, Safe Explorer, quality checks, browser checks, login checks, authenticated smoke checks, authorization checks, OpenAPI operation discovery, safe API smoke execution, evidence collection, report intelligence, HTML reports, and approved safe test plan execution do not depend on an LLM.

## Current Alpha Capabilities

- Run locally with Docker Compose.
- Complete first-run local admin setup.
- Protect project data, credential profiles, AI configuration, reports, evidence, runs, API specs, test plans, and authorization reports behind local authentication.
- Use HTTP-only session cookies with CSRF protection for mutating API requests.
- Create QA projects through an API.
- Create QA projects through a minimal web UI.
- Create projects through a guided setup wizard that can optionally configure AI, credentials, OpenAPI import, and selected first checks.
- Run a local demo workflow against `demo-web`, `demo-api`, and `fake-llm`.
- View dashboard quick-start cards, recent Safe QA runs, recent projects, status indicators, and a project readiness checklist.
- View a reports landing page for recent browser, API, discovery, Safe Explorer, quality, and Safe QA reports with recent severity and grouped-finding counts.
- Start runs that can include browser and API jobs.
- Start a browser-only smoke run for a project with `frontend_url`.
- Store project-scoped credential profiles encrypted at rest for deterministic test-account login.
- Add optional role metadata to credential profiles, such as `admin`, `readonly`, or customer roles.
- Test a credential profile login flow with configured selectors and success/failure criteria.
- Start an authenticated browser smoke run that logs in and visits one configured same-origin target path.
- Define explicit role-aware authorization checks for browser URL targets.
- Run deterministic authorization checks that log in with an actor credential profile, navigate only the configured target, and compare expected `allowed` or `denied` outcomes.
- View authorization run JSON/HTML reports, findings, screenshots, and `authorization_observations` evidence.
- Start safe application discovery runs for projects with `frontend_url`.
- Persist discovered pages, links, forms, fields, screenshots, browser observations, findings, and skip reasons.
- View discovery runs and application maps in the web UI.
- Export discovery JSON reports and self-contained HTML reports.
- Start Interactive Safe Explorer runs for projects with `frontend_url`.
- Observe visible links, buttons, forms, submit buttons, and inputs without storing full HTML.
- Execute only safe same-origin navigation actions by default.
- Record skipped unsafe, external, policy-blocked, duplicate, and unsupported actions with deterministic skip reasons.
- View Safe Explorer timelines, actions, findings, screenshot evidence, JSON reports, and self-contained HTML reports in the web UI.
- Start passive quality check runs for project frontends.
- Reuse latest or selected discovery runs as quality-check page lists.
- Run safe passive security header/cookie/form checks, basic accessibility heuristics, and simple performance/resource observations.
- View quality check runs and JSON/HTML quality reports in the web UI.
- View projects, runs, findings, evidence metadata, and reports in the web UI.
- Execute Playwright Chromium checks against a configured frontend URL.
- Execute safe API checks against `api_base_url`.
- Fetch and parse OpenAPI 3.x JSON/YAML from `openapi_url`.
- Import OpenAPI 3.x specs from URL or pasted JSON/YAML.
- Discover API operations, classify safe operations, and persist skip reasons.
- Run safe API smoke tests from imported OpenAPI specs.
- Test only safe OpenAPI methods by default: `GET`, `HEAD`, and `OPTIONS`.
- Skip mutating, authenticated, ambiguous, request-body, unresolved-parameter, and sensitive API operations.
- Enforce project `allowed_hosts` for browser and API requests.
- Collect page title, final URL, status code, screenshot evidence, browser observations, login observations, API observations, OpenAPI summaries, and API request evidence.
- Persist API smoke result rows with method, path, status, HTTP status, duration, content type, response size, error, and skip reason.
- Store metadata in PostgreSQL.
- Queue worker jobs with Redis.
- Store screenshots in MinIO/S3, with a local filesystem fallback.
- Generate structured JSON reports with executive summaries, normalized severity counts, grouped findings, top findings, affected-page summaries, noise summaries, raw finding counts, deduplication metadata, and safety limitations.
- Generate self-contained HTML reports at `GET /api/v1/runs/{run_id}/report.html` with grouped findings first and raw details still available.
- Download stored evidence objects at `GET /api/v1/evidence/{evidence_id}`.
- Configure optional OpenAI-compatible AI providers from the web UI or API.
- Test AI provider connectivity with a safe prompt.
- Run AI analysis for completed runs using sanitized report data.
- Show AI analysis in the web UI, JSON report, and HTML report when available.
- Generate AI-assisted test plans from sanitized project/run/report metadata.
- Generate discovery-aware AI-assisted test plans from sanitized application maps.
- View, delete, and export AI test plans in the web UI.
- Link AI test plans back into JSON and HTML run reports when they were generated from a run.
- Preview executable coverage for generated test plans.
- Preview which AI test plan steps are safely executable.
- Execute only approved, supported, same-origin, non-destructive browser DSL steps from a test plan.
- Persist test plan execution scenarios, steps, skip reasons, findings, evidence, JSON reports, and self-contained HTML reports.
- Start Safe QA Runs that reuse or create discovery, generate a discovery-aware plan, preview safe execution, and optionally run the approved safe DSL path.
- Optionally include passive quality checks in Safe QA Runs.
- View Safe QA Run JSON/HTML reports with discovery, quality checks, plan, preview, execution, deterministic report intelligence, and safety metadata.

## Architecture

```text
API client / smoke script / web UI
        |
        v
qualora-api
        |
        +--> PostgreSQL: local_users, user_sessions, projects, credential_profiles, discovery_runs, discovered_pages, discovered_links, discovered_forms, safe_explorer_runs, safe_explorer_steps, safe_explorer_actions, quality_check_runs, quality_check_results, authorization_checks, authorization_check_runs, authorization_check_results, test_runs, run_jobs, findings, evidence, api_specs, api_operations, api_check_results, ai_providers, ai_analyses, test_plans, test_plan_executions, qa_runs
        +--> Redis: browser, API, and test plan execution queues
        +--> MinIO/S3 evidence download proxy
        +--> Optional OpenAI-compatible AI provider for analysis and test planning
        |
        +--> qualora-worker-browser
        |       +--> Playwright browser smoke test
        |       +--> Deterministic selector-based login checks
        |       +--> Authenticated browser smoke test
        |       +--> Safe deterministic application discovery
        |       +--> Interactive Safe Explorer
        |       +--> Passive quality checks
        |       +--> Explicit role-aware authorization checks
        |       +--> Approved safe test plan execution steps
        |       +--> MinIO/S3 screenshot evidence
        |
        +--> qualora-worker-api
        |       +--> Legacy project API base URL/OpenAPI checks
        |
        +--> Safe OpenAPI import and API smoke execution in control plane
                +--> OpenAPI 3.x operation discovery
                +--> GET/HEAD/OPTIONS-only API smoke checks
                +--> API result rows, evidence, findings, reports
```

The web UI is served separately as `qualora-web` on `http://localhost:3000` and calls the API on `http://localhost:8080`.

See [docs/architecture.md](docs/architecture.md) for details.

## Quick Start

Requirements:

- Docker with Docker Compose.
- Python 3 for the smoke script.

Start Qualora:

```bash
docker compose up -d --build
```

Check health:

```bash
curl http://localhost:8080/healthz
```

Open the web UI:

```text
http://localhost:3000
```

On a fresh database, the web UI opens a first-run setup screen for the local admin account before showing project data. After login, use `#/setup-project` for guided setup or `Run demo workflow` on the dashboard for the deterministic local demo. The smoke script performs the same setup automatically for demo environments.

Run the smoke tests:

```bash
make smoke
```

The smoke target includes:

- Browser smoke against the local `demo-web` Compose service.
- Credential profile creation, deterministic login check, and authenticated browser smoke against `demo-web`.
- Role credential profile creation plus explicit authorization checks against demo `/admin` and customer invoice routes.
- Application discovery against `demo-web`, including discovered pages/forms, skipped unsafe/external links, screenshots, JSON report, and HTML report.
- Interactive Safe Explorer against `demo-web`, including observed pages/actions, executed safe navigation, skipped unsafe/external/POST/unsupported actions, screenshots, JSON report, and HTML report.
- Passive quality checks against `demo-web`, including security, accessibility, and performance findings plus JSON/HTML quality reports.
- OpenAPI import and safe API smoke against a local `demo-api` service started by the Makefile.
- AI provider smoke against a local fake OpenAI-compatible provider.
- Discovery-aware AI test plan generation from the application map.
- Safe QA Run preview and explicit execution against the approved safe browser DSL.
- Guided project setup through the onboarding API, including demo AI provider setup, demo OpenAPI import, credential profile creation, browser smoke, authenticated smoke, discovery, quality checks, Safe QA, and API smoke.
- Safe test plan execution smoke against the local `demo-web` service.

Stop the stack:

```bash
docker compose down
```

If local port `8080` is already in use:

```bash
QUALORA_API_PORT=18080 docker compose up -d --build
QUALORA_API_URL=http://localhost:18080 QUALORA_API_BASE_URL=http://localhost:18080 make smoke
```

## API Examples

The API is protected after first-run setup. For curl-based local testing, create or log in to the local admin account and reuse the cookie jar. Mutating requests must include the CSRF token.

```bash
COOKIE_JAR=/tmp/qualora.cookies

curl -s -c "$COOKIE_JAR" http://localhost:8080/api/v1/setup/status | python3 -m json.tool

curl -s -c "$COOKIE_JAR" http://localhost:8080/api/v1/setup/admin \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "admin@qualora.local",
    "display_name": "Qualora Admin",
    "password": "change-me-to-a-long-local-password",
    "confirm_password": "change-me-to-a-long-local-password"
  }' | python3 -m json.tool

CSRF=$(python3 - "$COOKIE_JAR" <<'PY'
import http.cookiejar, sys
jar = http.cookiejar.MozillaCookieJar(sys.argv[1])
jar.load(ignore_discard=True, ignore_expires=True)
for cookie in jar:
    if cookie.name == "qualora_csrf":
        print(cookie.value)
        break
PY
)
```

If setup is already complete, call `POST /api/v1/auth/login` with the same cookie jar and then refresh `CSRF` from the cookie jar. Use `-b "$COOKIE_JAR"` for protected `GET` requests, and use `-b "$COOKIE_JAR" -c "$COOKIE_JAR" -H "X-Qualora-CSRF: ${CSRF}"` for protected `POST`, `PUT`, and `DELETE` requests.

Create a project through guided onboarding and start safe first checks:

```bash
curl -s http://localhost:8080/api/v1/onboarding/project-setup \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "project": {
      "name": "Guided Example",
      "frontend_url": "https://example.com",
      "allowed_hosts": ["example.com"],
      "security_mode": "passive",
      "destructive_actions": false
    },
    "ai": {"mode": "skip"},
    "credential": {"mode": "skip"},
    "api_spec": {"mode": "skip"},
    "workflow": {
      "browser_smoke": true,
      "discovery": true,
      "quality_checks": true,
      "safe_qa": false
    }
  }' | python3 -m json.tool
```

Guided setup is orchestration over the existing safe features. It does not add autonomous browser control, active scanning, fuzzing, arbitrary form submission, or destructive testing.

Create a browser project:

```bash
curl -s http://localhost:8080/api/v1/projects \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Example Web App",
    "frontend_url": "https://example.com",
    "api_base_url": "",
    "openapi_url": "",
    "allowed_hosts": ["example.com"],
    "security_mode": "passive",
    "destructive_actions": false
  }'
```

Create an API/OpenAPI project:

```bash
curl -s http://localhost:8080/api/v1/projects \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Example API",
    "frontend_url": "",
    "api_base_url": "https://api.example.com",
    "openapi_url": "https://api.example.com/openapi.json",
    "allowed_hosts": ["api.example.com"],
    "security_mode": "passive",
    "destructive_actions": false
  }'
```

Import an OpenAPI spec for a project:

```bash
API_SPEC_ID=$(curl -s "http://localhost:8080/api/v1/projects/${PROJECT_ID}/api-specs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Demo API",
    "source_type": "url",
    "source_url": "http://demo-api:8080/openapi.yaml"
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["spec"]["id"])')
```

List discovered operations and skip reasons:

```bash
curl -s "http://localhost:8080/api/v1/api-specs/${API_SPEC_ID}/operations" \
  -b "$COOKIE_JAR" | python3 -m json.tool
```

Run a safe API smoke test from the imported spec:

```bash
API_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/api-specs/${API_SPEC_ID}/api-smoke-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Fetch API smoke results:

```bash
curl -s "http://localhost:8080/api/v1/runs/${API_RUN_ID}/api-results" \
  -b "$COOKIE_JAR" | python3 -m json.tool
curl -s "http://localhost:8080/api/v1/runs/${API_RUN_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
open "http://localhost:8080/api/v1/runs/${API_RUN_ID}/report.html"
```

Create a project and save its ID:

```bash
PROJECT_ID=$(curl -s http://localhost:8080/api/v1/projects \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Example App",
    "frontend_url": "https://example.com",
    "allowed_hosts": ["example.com"],
    "security_mode": "passive",
    "destructive_actions": false
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Start a run:

```bash
RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Start only a browser smoke run:

```bash
RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/browser-smoke-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Start a safe application discovery run:

```bash
DISCOVERY_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/discovery-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "max_pages": 20,
    "max_depth": 2,
    "same_origin_only": true
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')

curl -s "http://localhost:8080/api/v1/discovery-runs/${DISCOVERY_RUN_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
```

Discovery follows safe links only, skips external/unsafe/non-HTML links with recorded reasons, captures screenshots and browser observations, and never submits forms or clicks arbitrary buttons.

Start an Interactive Safe Explorer run:

```bash
SAFE_EXPLORER_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/safe-explorer-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "max_steps": 10,
    "max_depth": 2,
    "same_origin_only": true,
    "allow_get_forms": false
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')

curl -s "http://localhost:8080/api/v1/safe-explorer-runs/${SAFE_EXPLORER_RUN_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
open "http://localhost:8080/api/v1/safe-explorer-runs/${SAFE_EXPLORER_RUN_ID}/report.html"
```

Safe Explorer executes only safe classified navigation actions by default. It skips POST forms, unsafe labels, sensitive query values, external hosts, unsupported controls, duplicates, and policy-blocked actions with recorded reasons.

Start a passive quality check run from the latest completed discovery:

```bash
QUALITY_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/quality-check-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "use_latest_discovery": true,
    "max_pages": 10,
    "include_security": true,
    "include_accessibility": true,
    "include_performance": true
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')

curl -s "http://localhost:8080/api/v1/quality-check-runs/${QUALITY_RUN_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
open "http://localhost:8080/api/v1/quality-check-runs/${QUALITY_RUN_ID}/report.html"
```

Quality checks are passive metadata checks. They do not submit forms, click arbitrary buttons, run payloads, fuzz inputs, perform active security scanning, or use autonomous AI browser control.

Create a credential profile for deterministic login:

```bash
CREDENTIAL_PROFILE_ID=$(curl -s "http://localhost:8080/api/v1/projects/${PROJECT_ID}/credential-profiles" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Demo Login",
    "type": "username_password",
    "username": "demo@example.com",
    "password": "demo-password",
    "login_url": "http://demo-web:8080/login",
    "username_selector": "#username",
    "password_selector": "#password",
    "submit_selector": "#login-submit",
    "success_url_contains": "/dashboard",
    "success_text_contains": "Authenticated area",
    "failure_text_contains": "Invalid credentials",
    "post_login_wait_ms": 100,
    "is_default": true
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

The credential profile response includes only safe metadata such as configured flags and a masked username display hint. It never returns the raw username or password.

Test the configured login flow:

```bash
LOGIN_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/credential-profiles/${CREDENTIAL_PROFILE_ID}/test-login" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{}' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Run authenticated browser smoke after login:

```bash
AUTH_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/authenticated-browser-smoke-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "credential_profile_id": "'"${CREDENTIAL_PROFILE_ID}"'",
    "target_path": "/dashboard",
    "capture_screenshot": true,
    "max_duration_seconds": 30
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Create and run an explicit role-aware authorization check:

```bash
AUTHZ_CHECK_ID=$(curl -s "http://localhost:8080/api/v1/projects/${PROJECT_ID}/authorization-checks" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Readonly denied admin route",
    "type": "browser_url",
    "actor_credential_profile_id": "'"${READONLY_PROFILE_ID}"'",
    "expected_outcome": "denied",
    "target_url": "/admin",
    "denied_text_contains": "Access denied",
    "enabled": true
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')

AUTHZ_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/authorization-check-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{"check_ids":["'"${AUTHZ_CHECK_ID}"'"],"max_checks":10}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')

curl -s "http://localhost:8080/api/v1/authorization-check-runs/${AUTHZ_RUN_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
open "http://localhost:8080/api/v1/authorization-check-runs/${AUTHZ_RUN_ID}/report.html"
```

Authorization checks are explicit and conservative. They log in with the configured actor credential profile, navigate only the configured same-origin/allowed-host target, and do not crawl, fuzz, submit arbitrary forms, execute payloads, or use autonomous AI browser control.

Fetch the report:

```bash
curl -s "http://localhost:8080/api/v1/runs/${RUN_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
```

Open the HTML report:

```bash
open "http://localhost:8080/api/v1/runs/${RUN_ID}/report.html"
```

Download screenshot evidence by ID:

```bash
curl -L "http://localhost:8080/api/v1/evidence/${EVIDENCE_ID}" \
  -b "$COOKIE_JAR" -o screenshot.png
```

Configure a fake/local OpenAI-compatible provider:

```bash
curl -s http://localhost:8080/api/v1/ai/providers \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Local Fake LLM",
    "preset": "custom",
    "type": "openai-compatible",
    "base_url": "http://fake-llm:8080/v1",
    "model": "qualora-fake-analyst",
    "api_key": "fake-key",
    "temperature": 0.2,
    "max_output_tokens": 1200,
    "timeout_seconds": 10,
    "send_screenshots": false,
    "send_html": false,
    "send_network_bodies": false,
    "redaction_enabled": true,
    "is_default": true
  }'
```

Run AI analysis for an existing completed run:

```bash
curl -s -X POST "http://localhost:8080/api/v1/runs/${RUN_ID}/ai-analysis" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{}' | python3 -m json.tool
```

Generate an AI-assisted test plan for a project:

```bash
TEST_PLAN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/ai-test-plans" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "run_id": "'"${RUN_ID}"'",
    "discovery_run_id": "'"${DISCOVERY_RUN_ID}"'",
    "include_discovery_map": true,
    "execution_mode": "safe_executable",
    "max_pages_from_discovery": 20,
    "product_context": "Public checkout and account settings are high-priority flows. Do not include secrets here.",
    "focus_areas": ["smoke", "functional", "api", "regression"],
    "max_scenarios": 10
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

List and export test plans:

```bash
curl -s "http://localhost:8080/api/v1/projects/${PROJECT_ID}/test-plans" \
  -b "$COOKIE_JAR" | python3 -m json.tool
curl -s "http://localhost:8080/api/v1/test-plans/${TEST_PLAN_ID}/export.json" \
  -b "$COOKIE_JAR" | python3 -m json.tool
```

Preview the safe execution mapping for a test plan:

```bash
curl -s -X POST "http://localhost:8080/api/v1/test-plans/${TEST_PLAN_ID}/executions" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "max_scenarios": 5,
    "max_steps_per_scenario": 10,
    "dry_run": true
  }' | python3 -m json.tool
```

Start an approved safe execution:

```bash
EXECUTION_ID=$(curl -s -X POST "http://localhost:8080/api/v1/test-plans/${TEST_PLAN_ID}/executions" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "max_scenarios": 5,
    "max_steps_per_scenario": 10,
    "dry_run": false
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["execution"]["id"])')
```

Fetch the safe execution report:

```bash
curl -s "http://localhost:8080/api/v1/test-plan-executions/${EXECUTION_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
open "http://localhost:8080/api/v1/test-plan-executions/${EXECUTION_ID}/report.html"
```

Start a Safe QA Run preview from discovery:

```bash
QA_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/qa-runs" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{
    "discovery_run_id": "'"${DISCOVERY_RUN_ID}"'",
    "execution_mode": "preview",
    "use_latest_discovery": false,
    "execute": false,
    "max_pages": 20,
    "max_scenarios": 5,
    "max_steps_per_scenario": 10,
    "focus_areas": ["smoke", "navigation", "regression"]
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')

curl -s "http://localhost:8080/api/v1/qa-runs/${QA_RUN_ID}/report" \
  -b "$COOKIE_JAR" | python3 -m json.tool
open "http://localhost:8080/api/v1/qa-runs/${QA_RUN_ID}/report.html"
```

Execute a previewed Safe QA Run only after reviewing the generated plan and safe execution preview:

```bash
curl -s -X POST "http://localhost:8080/api/v1/qa-runs/${QA_RUN_ID}/execute" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "X-Qualora-CSRF: ${CSRF}" \
  -H 'Content-Type: application/json' \
  -d '{}' | python3 -m json.tool
```

## AI Providers

AI is optional. Configure a provider only when you want model-generated report analysis or test-plan suggestions.

Supported provider type in `v0.17.0-alpha`:

- `openai-compatible`

Preset values:

| Preset | Base URL | Model example | Notes |
| --- | --- | --- | --- |
| OpenAI | `https://api.openai.com/v1` | `gpt-4o-mini` | Requires an API key. |
| OpenRouter | `https://openrouter.ai/api/v1` | `openai/gpt-4o-mini` | Optional headers: `HTTP-Referer`, `X-OpenRouter-Title`. |
| Ollama | `http://ollama:11434/v1` | `qwen2.5-coder:7b` | API key can usually be blank or a dummy value. Ollama is not started by default. |
| Custom OpenAI-compatible | user-provided | user-provided | Works with vLLM, LM Studio, LiteLLM, LocalAI, or internal gateways that expose chat completions. |

AI prompt safety defaults:

- Redaction enabled.
- Screenshots disabled.
- Full HTML disabled.
- Network bodies disabled.

AI-assisted test plans are reviewable suggestions. In `v0.17.0-alpha`, AI planning can include a sanitized discovery map and can ask the model for safe executable DSL candidates, but a user may explicitly preview and execute only the supported safe browser DSL subset: `goto`, `assert_title_contains`, `assert_url_contains`, `assert_text_visible`, `assert_element_visible`, `assert_link_exists`, `check_link_status`, `capture_screenshot`, `collect_browser_signals`, `wait_for_load_state`, `assert_no_console_errors`, and `assert_no_failed_requests`. Unsupported, ambiguous, authenticated, destructive, mutating, upload, admin, exploit, and out-of-scope steps are skipped with reasons. Credential-profile login checks, role-aware authorization checks, application discovery, Interactive Safe Explorer, guided onboarding, and report intelligence are deterministic paths and are not AI-controlled.

## Report Intelligence

Every primary JSON and HTML report includes deterministic report intelligence in `v0.17.0-alpha`:

- `executive_summary` with pass/warning/fail/unknown status, what was tested, what was not tested, recommended next actions, and safety limitations.
- `severity_counts` normalized to `critical`, `high`, `medium`, `low`, and `info`.
- `grouped_findings` and `top_findings` built from deterministic fingerprints. Raw findings and quality result rows remain available.
- `top_affected_pages`, `noise_summary`, `raw_findings_count`, and `deduplication_summary`.

This is not AI summarization. It does not send credentials, cookies, local storage, session storage, auth headers, tokens, screenshots, full HTML, request bodies, or response bodies to any model. Optional AI analysis remains a separate user-triggered feature that uses sanitized report metadata only.

## Report Example

A browser smoke run includes screenshot and browser observation evidence:

```json
{
  "run_id": "0037c342-0394-4ef2-a87f-ebf568c3b713",
  "project_id": "9d3ed104-3b54-49d6-a307-0102c2d3fd3f",
  "status": "completed",
  "summary": {
    "total_findings": 0,
    "critical": 0,
    "high": 0,
    "medium": 0,
    "low": 0,
    "info": 0
  },
  "findings": [],
  "evidence": [
    {
      "id": "90d77c2a-7599-4e6f-8d66-d7e8fd0b7c1f",
      "type": "screenshot",
      "uri": "s3://qualora-evidence/runs/0037c342-0394-4ef2-a87f-ebf568c3b713/screenshots/1720944000-screen.png",
      "metadata": {
        "filename": "1720944000-screen.png",
        "content_type": "image/png",
        "size_bytes": 30421,
        "target_url": "http://demo-web:8080/",
        "final_url": "http://demo-web:8080/",
        "page_title": "Qualora Demo Web",
        "status_code": 200
      }
    },
    {
      "type": "browser_observations",
      "uri": "inline://browser-observations",
      "metadata": {
        "target_url": "http://demo-web:8080/",
        "final_url": "http://demo-web:8080/",
        "page_title": "Qualora Demo Web",
        "status_code": 200,
        "console_errors": [],
        "failed_requests": []
      }
    }
  ],
  "metadata": {
    "jobs": [
      {
        "kind": "browser",
        "status": "completed"
      }
    ]
  },
  "ai_analysis": null,
  "test_plans": []
}
```

When AI analysis has been generated, `ai_analysis` contains the provider/model metadata, status, summaries, risk level, token counts, and the parsed JSON analysis. When an AI test plan is generated from a run, `test_plans` contains lightweight references to related plans. Safe test plan execution reports are available separately at `/api/v1/test-plan-executions/{execution_id}/report`.

## Development Commands

```bash
make dev
make test
make lint
make compose-up
make compose-down
make logs
make smoke
```

See [docs/development.md](docs/development.md) for local development notes.

## Safety And Allowed Hosts

Only run Qualora against systems you own or are explicitly authorized to test.

The alpha is safe by default:

- Every project must define `allowed_hosts`.
- Browser navigation, browser network requests, API base URL checks, and OpenAPI checks are constrained by `allowed_hosts`.
- API worker tests only `GET`, `HEAD`, and `OPTIONS` by default.
- Quality checks are passive only and are not penetration tests, WCAG audits, Lighthouse audits, fuzzers, or active scanners.
- `security_mode` is currently limited to `passive`.
- `destructive_actions` must be `false`.
- `localhost`, `.local`, loopback, link-local, private IP literal targets, common cloud metadata targets, and public hostnames resolving to blocked IP ranges are blocked by default.
- `allow_private_targets: true` may be used for local/private systems you control.
- Authenticated browser smoke is limited to configured credential profiles and deterministic selectors.
- Authenticated API testing is not implemented in this release.
- Login automation is not autonomous and never uses AI browser control.
- Secrets, credentials, cookies, and authorization headers must not be logged.
- Screenshots and reports should be treated as sensitive evidence artifacts.
- The web UI and API require local admin authentication after first-run setup, but this alpha is still intended for trusted local/self-hosted environments only.
- AI is disabled until a provider is configured.
- AI prompts are built from sanitized report data only.
- Redaction is enabled by default.
- Screenshots, full HTML, cookies, credentials, authorization headers, and full network bodies are not sent to AI by default.
- AI provider API keys and extra headers are encrypted at rest using `QUALORA_ENCRYPTION_KEY`; the Compose fallback key is for local demo use only.
- Credential profile usernames and passwords are encrypted at rest using `QUALORA_ENCRYPTION_KEY`; raw credential values are never returned in API responses or sent to AI.
- AI-assisted test plans are stored as suggestions and are not executed automatically.
- Test plan execution is never autonomous: users must explicitly preview/start it, and only the supported safe browser DSL is executed.
- Test plan execution enforces same-origin frontend targets and project `allowed_hosts`.

See [docs/security-model.md](docs/security-model.md) and [SECURITY.md](SECURITY.md).

## Current Limitations

- Local authentication is limited to one admin role and is not production-hardened identity management.
- No user management UI, password reset flow, SSO/OIDC/SAML, multi-user RBAC, teams, or multi-tenancy.
- Web UI is alpha and intentionally minimal.
- AI provider management, AI analysis, AI-assisted test planning, and safe test plan execution are alpha and optional.
- Only OpenAI-compatible chat completion providers are supported.
- No native Anthropic, Gemini, or provider-specific SDK integrations yet.
- Generated test plans are not executed automatically or as free-form instructions.
- Safe test plan execution is limited to the supported non-destructive browser DSL and same-origin link checks.
- Screenshot preview/download is available only for evidence records known to Qualora.
- Quality checks are alpha heuristics, not full security, accessibility, performance, Lighthouse, Core Web Vitals, or WCAG coverage.
- No authenticated API testing.
- Authenticated browser smoke supports one configured login form and one same-origin target path per run.
- No arbitrary form submission, multi-step authenticated journeys, MFA, role switching, or session export.
- No active security scanning.
- No destructive API testing by default.
- No full OpenAPI schema validation or schema fuzzing.
- No request body generation.
- No Playwright trace download/export yet.
- No autonomous AI browser control.
- No Helm/Kubernetes deployment.
- Workers write results directly to PostgreSQL in this alpha.
- MinIO uses local development credentials in Docker Compose.

## Documentation

- [Architecture](docs/architecture.md)
- [Security model](docs/security-model.md)
- [Development](docs/development.md)
- [Release process](docs/release.md)
- [Roadmap](docs/roadmap.md)
- [OpenAPI contract](api/openapi/qualora.v1.yaml)
- [Changelog](CHANGELOG.md)

## Roadmap

Near-term work:

- Harden the worker result path so workers submit results through the control plane.
- Add run retries and clearer failure states.
- Add signed URL support or stronger evidence access controls.
- Move AI analysis to an async worker path.
- Move AI test planning to an async analyzer worker path.
- Harden safe test plan execution status/retry handling.
- Add audit logging, login rate limiting, and local auth hardening.
- Expand OpenAPI validation.
- Deepen quality checks with richer accessibility/performance signals once the passive alpha path is stable.

See [docs/roadmap.md](docs/roadmap.md).

## Contributing

Contributions are welcome. Start with:

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [AGENTS.md](AGENTS.md)

Please keep early contributions focused on the self-hosted MVP and avoid adding SaaS, billing, multi-tenancy, active scanning, or frontend UI assumptions unless they are explicitly part of the current roadmap.

## License

Qualora is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

# Qualora

**Open-source, self-hosted autonomous QA for web applications and APIs.**

Qualora is an open-source, self-hosted autonomous QA platform that runs browser-based and API smoke tests, collects evidence, and generates structured reports for web applications and APIs.

`v0.11.0-alpha` adds local first-run admin setup and session-based authentication for the self-hosted API and web UI. Qualora remains deterministic and useful without AI: browser checks, login checks, authenticated smoke checks, authorization checks, OpenAPI operation discovery, safe API smoke execution, evidence collection, JSON reports, HTML reports, and approved safe test plan execution do not depend on an LLM.

## Current Alpha Capabilities

- Run locally with Docker Compose.
- Complete first-run local admin setup.
- Protect project data, credential profiles, AI configuration, reports, evidence, runs, API specs, test plans, and authorization reports behind local authentication.
- Use HTTP-only session cookies with CSRF protection for mutating API requests.
- Create QA projects through an API.
- Create QA projects through a minimal web UI.
- Start runs that can include browser and API jobs.
- Start a browser-only smoke run for a project with `frontend_url`.
- Store project-scoped credential profiles encrypted at rest for deterministic test-account login.
- Add optional role metadata to credential profiles, such as `admin`, `readonly`, or customer roles.
- Test a credential profile login flow with configured selectors and success/failure criteria.
- Start an authenticated browser smoke run that logs in and visits one configured same-origin target path.
- Define explicit role-aware authorization checks for browser URL targets.
- Run deterministic authorization checks that log in with an actor credential profile, navigate only the configured target, and compare expected `allowed` or `denied` outcomes.
- View authorization run JSON/HTML reports, findings, screenshots, and `authorization_observations` evidence.
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
- Generate structured JSON reports.
- Generate self-contained HTML reports at `GET /api/v1/runs/{run_id}/report.html`.
- Download stored evidence objects at `GET /api/v1/evidence/{evidence_id}`.
- Configure optional OpenAI-compatible AI providers from the web UI or API.
- Test AI provider connectivity with a safe prompt.
- Run AI analysis for completed runs using sanitized report data.
- Show AI analysis in the web UI, JSON report, and HTML report when available.
- Generate AI-assisted test plans from sanitized project/run/report metadata.
- View, delete, and export AI test plans in the web UI.
- Link AI test plans back into JSON and HTML run reports when they were generated from a run.
- Preview which AI test plan steps are safely executable.
- Execute only approved, supported, same-origin, non-destructive browser DSL steps from a test plan.
- Persist test plan execution scenarios, steps, skip reasons, findings, evidence, JSON reports, and self-contained HTML reports.

## Architecture

```text
API client / smoke script / web UI
        |
        v
qualora-api
        |
        +--> PostgreSQL: local_users, user_sessions, projects, credential_profiles, authorization_checks, authorization_check_runs, authorization_check_results, test_runs, run_jobs, findings, evidence, api_specs, api_operations, api_check_results, ai_providers, ai_analyses, test_plans, test_plan_executions
        +--> Redis: browser, API, and test plan execution queues
        +--> MinIO/S3 evidence download proxy
        +--> Optional OpenAI-compatible AI provider for analysis and test planning
        |
        +--> qualora-worker-browser
        |       +--> Playwright browser smoke test
        |       +--> Deterministic selector-based login checks
        |       +--> Authenticated browser smoke test
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

On a fresh database, the web UI opens a first-run setup screen for the local admin account before showing project data. The smoke script performs the same setup automatically for demo environments.

Run the smoke tests:

```bash
make smoke
```

The smoke target includes:

- Browser smoke against the local `demo-web` Compose service.
- Credential profile creation, deterministic login check, and authenticated browser smoke against `demo-web`.
- Role credential profile creation plus explicit authorization checks against demo `/admin` and customer invoice routes.
- OpenAPI import and safe API smoke against a local `demo-api` service started by the Makefile.
- AI provider smoke against a local fake OpenAI-compatible provider.
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
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Demo API",
    "source_type": "url",
    "source_url": "http://demo-api:8080/openapi.yaml"
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["spec"]["id"])')
```

List discovered operations and skip reasons:

```bash
curl -s "http://localhost:8080/api/v1/api-specs/${API_SPEC_ID}/operations" | python3 -m json.tool
```

Run a safe API smoke test from the imported spec:

```bash
API_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/api-specs/${API_SPEC_ID}/api-smoke-runs" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Fetch API smoke results:

```bash
curl -s "http://localhost:8080/api/v1/runs/${API_RUN_ID}/api-results" | python3 -m json.tool
curl -s "http://localhost:8080/api/v1/runs/${API_RUN_ID}/report" | python3 -m json.tool
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
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Start only a browser smoke run:

```bash
RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/browser-smoke-runs" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Create a credential profile for deterministic login:

```bash
CREDENTIAL_PROFILE_ID=$(curl -s "http://localhost:8080/api/v1/projects/${PROJECT_ID}/credential-profiles" \
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
  -H 'Content-Type: application/json' \
  -d '{}' | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
```

Run authenticated browser smoke after login:

```bash
AUTH_RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/authenticated-browser-smoke-runs" \
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
  -H 'Content-Type: application/json' \
  -d '{"check_ids":["'"${AUTHZ_CHECK_ID}"'"],"max_checks":10}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')

curl -s "http://localhost:8080/api/v1/authorization-check-runs/${AUTHZ_RUN_ID}/report" | python3 -m json.tool
open "http://localhost:8080/api/v1/authorization-check-runs/${AUTHZ_RUN_ID}/report.html"
```

Authorization checks are explicit and conservative. They log in with the configured actor credential profile, navigate only the configured same-origin/allowed-host target, and do not crawl, fuzz, submit arbitrary forms, execute payloads, or use autonomous AI browser control.

Fetch the report:

```bash
curl -s "http://localhost:8080/api/v1/runs/${RUN_ID}/report" | python3 -m json.tool
```

Open the HTML report:

```bash
open "http://localhost:8080/api/v1/runs/${RUN_ID}/report.html"
```

Download screenshot evidence by ID:

```bash
curl -L "http://localhost:8080/api/v1/evidence/${EVIDENCE_ID}" -o screenshot.png
```

Configure a fake/local OpenAI-compatible provider:

```bash
curl -s http://localhost:8080/api/v1/ai/providers \
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
  -H 'Content-Type: application/json' \
  -d '{}' | python3 -m json.tool
```

Generate an AI-assisted test plan for a project:

```bash
curl -s -X POST "http://localhost:8080/api/v1/projects/${PROJECT_ID}/ai-test-plans" \
  -H 'Content-Type: application/json' \
  -d '{
    "run_id": "'"${RUN_ID}"'",
    "product_context": "Public checkout and account settings are high-priority flows. Do not include secrets here.",
    "focus_areas": ["smoke", "functional", "api", "regression"],
    "max_scenarios": 10
  }' | python3 -m json.tool
```

List and export test plans:

```bash
curl -s "http://localhost:8080/api/v1/projects/${PROJECT_ID}/test-plans" | python3 -m json.tool
curl -s "http://localhost:8080/api/v1/test-plans/${TEST_PLAN_ID}/export.json" | python3 -m json.tool
```

Preview the safe execution mapping for a test plan:

```bash
curl -s -X POST "http://localhost:8080/api/v1/test-plans/${TEST_PLAN_ID}/executions" \
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
  -H 'Content-Type: application/json' \
  -d '{
    "max_scenarios": 5,
    "max_steps_per_scenario": 10,
    "dry_run": false
  }' | python3 -c 'import json,sys; print(json.load(sys.stdin)["execution"]["id"])')
```

Fetch the safe execution report:

```bash
curl -s "http://localhost:8080/api/v1/test-plan-executions/${EXECUTION_ID}/report" | python3 -m json.tool
open "http://localhost:8080/api/v1/test-plan-executions/${EXECUTION_ID}/report.html"
```

## AI Providers

AI is optional. Configure a provider only when you want model-generated report analysis or test-plan suggestions.

Supported provider type in `v0.11.0-alpha`:

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

AI-assisted test plans are reviewable suggestions. In `v0.11.0-alpha`, a user may explicitly preview and execute only the supported safe browser DSL subset: `goto`, `assert_title_contains`, `assert_url_contains`, `assert_text_visible`, `assert_element_visible`, `assert_link_exists`, `check_link_status`, `capture_screenshot`, `collect_browser_signals`, `wait_for_load_state`, `assert_no_console_errors`, and `assert_no_failed_requests`. Unsupported, ambiguous, authenticated, destructive, mutating, upload, admin, exploit, and out-of-scope steps are skipped with reasons. Credential-profile login checks and role-aware authorization checks are separate deterministic browser-worker paths and are not AI-controlled.

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
- Add passive security checks.

See [docs/roadmap.md](docs/roadmap.md).

## Contributing

Contributions are welcome. Start with:

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [AGENTS.md](AGENTS.md)

Please keep early contributions focused on the self-hosted MVP and avoid adding SaaS, billing, multi-tenancy, active scanning, or frontend UI assumptions unless they are explicitly part of the current roadmap.

## License

Qualora is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

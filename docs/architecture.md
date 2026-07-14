# Architecture

Qualora v0.6.0-alpha is a small Docker Compose MVP for browser and API QA smoke runs with a minimal web UI, human-friendly reports, control-plane evidence download for stored artifacts, optional AI analysis of completed reports, and AI-assisted test plan suggestions.

## Runtime Components

```text
API client / smoke script / qualora-web
        |
        v
qualora-api
        |
        +--> PostgreSQL: projects, test_runs, run_jobs, findings, evidence, ai_providers, ai_analyses, test_plans
        +--> Redis: browser and API run queues
        +--> MinIO/S3 evidence objects by evidence ID
        +--> Optional OpenAI-compatible AI provider
        |       +--> Report analysis
        |       +--> Test plan suggestions
        |
        +--> qualora-worker-browser
        |       +--> Playwright Chromium smoke test
        |       +--> MinIO/S3 screenshot evidence
        |
        +--> qualora-worker-api
                +--> API base URL checks
                +--> OpenAPI 3.x safe method checks
                +--> PostgreSQL evidence and findings
```

## Services

### `qualora-api`

The Go control plane exposes the HTTP API, validates project scope, persists metadata, creates per-run jobs, queues worker jobs, and renders JSON/HTML reports.

Current endpoints:

- `GET /healthz`
- `POST /api/v1/projects`
- `GET /api/v1/projects`
- `GET /api/v1/projects/{project_id}`
- `POST /api/v1/projects/{project_id}/runs`
- `GET /api/v1/projects/{project_id}/runs`
- `POST /api/v1/projects/{project_id}/browser-smoke-runs`
- `POST /api/v1/projects/{project_id}/ai-test-plans`
- `GET /api/v1/projects/{project_id}/test-plans`
- `GET /api/v1/runs`
- `GET /api/v1/runs/{run_id}`
- `GET /api/v1/runs/{run_id}/report`
- `GET /api/v1/runs/{run_id}/report.html`
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

### `qualora-web`

The React/Vite web UI is intentionally small. It calls the control-plane API from the browser and displays:

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

It has no authentication in this alpha and should be exposed only in trusted local/self-hosted environments.

### `qualora-worker-browser`

The Node.js browser worker consumes Redis browser jobs and runs a Playwright smoke check against `frontend_url`.

It currently captures:

- Page title.
- Final URL.
- Initial document status code.
- Body text length.
- Screenshot.
- Console errors.
- Failed network requests.
- Blocked out-of-scope browser requests.
- Basic findings for obvious load, timeout, non-success status, console, request, empty page, and scope issues.

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

### PostgreSQL

PostgreSQL stores durable metadata:

- `projects`
- `test_runs`
- `run_jobs`
- `findings`
- `evidence`
- `ai_providers`
- `ai_analyses`
- `test_plans`

Reports are generated dynamically from run, job, finding, and evidence rows.

### Redis

Redis is the MVP queue for browser and API jobs. PostgreSQL remains the source of durable run state.

### MinIO

MinIO stores screenshot evidence through the S3-compatible API. API evidence is stored as metadata rows in PostgreSQL. The control plane can stream stored evidence objects by evidence ID so the UI can preview/download screenshots without exposing MinIO credentials.

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

For v0.6, AI analysis and AI test planning run synchronously in the control plane. The database models are separated so both paths can move to a dedicated analyzer worker later.

The safe AI input builder includes only sanitized report data such as run status, summary counts, finding titles/summaries, safe evidence metadata, browser/API metadata, and job metadata. It strips or redacts URL queries, cookies, authorization values, tokens, passwords, API keys, session IDs, JWT-looking strings, full response bodies, full HTML, and secret-looking fields. Screenshots, HTML, and network bodies are not sent by default.

AI-assisted test plans use sanitized project configuration, optional product context, selected focus areas, optional latest/run-specific report metadata, and optional AI analysis summaries. The strict plan parser accepts only a reviewable JSON structure with assumptions, coverage goals, scenarios, steps, assertions, test data needs, instrumentation suggestions, and limitations. Generated steps are never executed by Qualora in this release.

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

## Intentional Alpha Constraints

- No user accounts or authentication.
- Web UI is alpha and suitable only for trusted self-hosted/local environments.
- AI provider management is alpha and should be used only in trusted local/self-hosted environments.
- AI-assisted test planning is alpha and should be treated as human-reviewable suggestions only.
- Only OpenAI-compatible chat completion providers are supported.
- Evidence download is limited to stored evidence records and is not a signed URL system.
- No authenticated API testing.
- No login automation.
- No autonomous AI browser control.
- No automatic execution of generated AI test plan steps.
- No active security scanning.
- No destructive API testing by default.
- No full OpenAPI schema validation or fuzzing.
- No Playwright trace export yet.
- No Kubernetes or Helm support yet.

The older MVP notes remain in [architecture/mvp.md](architecture/mvp.md) for implementation context, but this document is the release-facing architecture reference.

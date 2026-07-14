# Qualora

**Open-source, self-hosted autonomous QA for web applications and APIs.**

Qualora is an open-source, self-hosted autonomous QA platform that runs browser-based and API smoke tests, collects evidence, and generates structured reports for web applications and APIs.

`v0.5.0-alpha` adds optional AI provider management and AI report analysis. Qualora remains fully useful without AI: browser/API checks, evidence collection, JSON reports, and HTML reports do not depend on an LLM. The AI layer analyzes existing deterministic run data through an OpenAI-compatible provider configured by the user.

## Current Alpha Capabilities

- Run locally with Docker Compose.
- Create QA projects through an API.
- Create QA projects through a minimal web UI.
- Start runs that can include browser and API jobs.
- Start a browser-only smoke run for a project with `frontend_url`.
- View projects, runs, findings, evidence metadata, and reports in the web UI.
- Execute Playwright Chromium checks against a configured frontend URL.
- Execute safe API checks against `api_base_url`.
- Fetch and parse OpenAPI 3.x JSON/YAML from `openapi_url`.
- Test only safe OpenAPI methods by default: `GET`, `HEAD`, and `OPTIONS`.
- Enforce project `allowed_hosts` for browser and API requests.
- Collect page title, final URL, status code, screenshot evidence, browser observations, API observations, OpenAPI summaries, and API request evidence.
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

## Architecture

```text
API client / smoke script / web UI
        |
        v
qualora-api
        |
        +--> PostgreSQL: projects, test_runs, run_jobs, findings, evidence, ai_providers, ai_analyses
        +--> Redis: browser and API run queues
        +--> MinIO/S3 evidence download proxy
        +--> Optional OpenAI-compatible AI provider
        |
        +--> qualora-worker-browser
        |       +--> Playwright browser smoke test
        |       +--> MinIO/S3 screenshot evidence
        |
        +--> qualora-worker-api
                +--> API base URL checks
                +--> OpenAPI 3.x safe method checks
                +--> PostgreSQL evidence and findings
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

Run the smoke tests:

```bash
make smoke
```

The smoke target includes:

- Browser smoke against the local `demo-web` Compose service.
- API/OpenAPI smoke against a local mock API service started by the Makefile.
- AI provider smoke against a local fake OpenAI-compatible provider.

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

Create a browser project:

```bash
curl -s http://localhost:8080/api/v1/projects \
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

Create a project and save its ID:

```bash
PROJECT_ID=$(curl -s http://localhost:8080/api/v1/projects \
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

## AI Providers

AI is optional. Configure a provider only when you want model-generated report analysis.

Supported provider type in `v0.5.0-alpha`:

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
  "ai_analysis": null
}
```

When AI analysis has been generated, `ai_analysis` contains the provider/model metadata, status, summaries, risk level, token counts, and the parsed JSON analysis.

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
- Authenticated API testing and login automation are not implemented in this release.
- Secrets, credentials, cookies, and authorization headers must not be logged.
- Screenshots and reports should be treated as sensitive evidence artifacts.
- The web UI has no authentication yet and is intended for trusted local/self-hosted alpha environments only.
- AI is disabled until a provider is configured.
- AI prompts are built from sanitized report data only.
- Redaction is enabled by default.
- Screenshots, full HTML, cookies, credentials, authorization headers, and full network bodies are not sent to AI by default.
- AI provider API keys and extra headers are encrypted at rest using `QUALORA_ENCRYPTION_KEY`; the Compose fallback key is for local demo use only.

See [docs/security-model.md](docs/security-model.md) and [SECURITY.md](SECURITY.md).

## Current Limitations

- No authentication.
- Web UI is alpha and intentionally minimal.
- AI provider management and AI analysis are alpha and optional.
- Only OpenAI-compatible chat completion providers are supported.
- No native Anthropic, Gemini, or provider-specific SDK integrations yet.
- Screenshot preview/download is available only for evidence records known to Qualora.
- No authenticated API testing.
- No login automation or storage for application test-account credentials.
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

# Development

This document covers local development for Qualora v0.7.0-alpha.

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
```

Command behavior:

- `make dev`: starts the Docker Compose stack.
- `make test`: runs Go tests, web build/type-check, and browser/API worker tests.
- `make lint`: runs the same checks plus `docker compose config`.
- `make compose-up`: runs `docker compose up -d --build`.
- `make compose-down`: runs `docker compose down`.
- `make logs`: tails API, web, browser worker, and API worker logs.
- `make smoke`: starts the local demo web, mock API, and fake LLM profile services; creates an AI provider, browser project, and API project; starts runs; polls to completion; runs AI analysis; generates AI test plans; previews and executes a safe browser test plan; prints JSON/HTML report, test-plan, and execution URLs; validates HTML report export; validates test-plan export; and validates screenshot evidence download.

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

- AI provider creation and provider-test against local `fake-llm`.
- Browser smoke against the local `demo-web` Compose service.
- AI analysis for the completed browser smoke run.
- AI test plan generation/export validation for the browser smoke run.
- Safe test plan execution preview and run validation for the browser smoke project.
- API/OpenAPI smoke against the local `mock-api` Compose service.
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
QUALORA_API_SMOKE_URL=http://mock-api:8080 \
QUALORA_API_SMOKE_OPENAPI_URL=http://mock-api:8080/openapi.json \
QUALORA_API_SMOKE_ALLOWED_HOST=mock-api \
make smoke
```

For private or local targets, create projects manually with `allow_private_targets: true` only when testing systems you control.

## AI Provider Development

The v0.7 AI path uses OpenAI-compatible chat completions only. AI analysis and AI-assisted test planning are optional and run synchronously in the control plane for this alpha.

Useful local values:

```text
QUALORA_ENCRYPTION_KEY=qualora-insecure-dev-key-change-me
QUALORA_FAKE_LLM_URL=http://fake-llm:8080/v1
FAKE_LLM_HEALTH_URL=http://localhost:18083/health
```

The default Compose encryption key is intentionally insecure and only for local development. Set a strong `QUALORA_ENCRYPTION_KEY` before storing real provider credentials.

AI-assisted test plans are reviewable suggestions and are never executed automatically. Qualora can execute only explicitly approved, deterministic safe DSL steps after a preview. It does not send screenshots/full HTML/raw traces/full network bodies to AI by default, and it redacts secret-looking values before prompt construction and storage.

Safe test plan execution currently supports only browser actions that stay on the project frontend origin: `goto`, `assert_title_contains`, `assert_url_contains`, `assert_text_visible`, `assert_element_visible`, `assert_link_exists`, `check_link_status`, `capture_screenshot`, `collect_browser_signals`, `wait_for_load_state`, `assert_no_console_errors`, and `assert_no_failed_requests`.

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

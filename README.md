# Qualora

Qualora is an open-source, self-hosted autonomous QA platform for web applications and APIs.

It is being built as part of Operalith's mission: **open-source AI-powered engineering tools for modern software operations**. The first alpha focuses on a small, useful loop: create a project, start a browser smoke test, collect evidence, and read a structured JSON report.

## Current MVP

The current Docker Compose stack includes:

- `qualora-api`: Go control plane API.
- `qualora-worker-browser`: TypeScript/Node.js Playwright browser worker.
- `postgres`: durable project, run, finding, and evidence metadata.
- `redis`: run queue.
- `minio`: S3-compatible screenshot storage.

Implemented behavior:

- Create projects through the API.
- Start a test run for a project.
- Queue the run in Redis.
- Execute a Playwright smoke check in the browser worker.
- Enforce project `allowed_hosts` before navigation and for browser network requests.
- Collect page title, screenshot, console errors, failed requests, and blocked out-of-scope requests.
- Persist evidence metadata and findings.
- Store screenshots in MinIO with a local filesystem fallback.
- Return a structured JSON report.

## Quick Start

Requirements:

- Docker with Docker Compose.
- Python 3 for the smoke script.

Start the stack from the repository root:

```bash
docker compose up -d --build
```

If port `8080` is already in use locally:

```bash
QUALORA_API_PORT=18080 docker compose up -d --build
QUALORA_API_URL=http://localhost:18080 make smoke
```

Check health:

```bash
curl http://localhost:8080/healthz
```

Run the smoke test against `https://example.com`:

```bash
make smoke
```

Stop the stack:

```bash
docker compose down
```

## API Examples

Create a project:

```bash
curl -s http://localhost:8080/api/v1/projects \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Example App",
    "frontend_url": "https://example.com",
    "api_base_url": "",
    "openapi_url": "",
    "allowed_hosts": ["example.com"],
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

Check run status:

```bash
curl -s "http://localhost:8080/api/v1/runs/${RUN_ID}" | python3 -m json.tool
```

Fetch the report:

```bash
curl -s "http://localhost:8080/api/v1/runs/${RUN_ID}/report" | python3 -m json.tool
```

List projects:

```bash
curl -s http://localhost:8080/api/v1/projects | python3 -m json.tool
```

## Development

Common commands:

```bash
make dev
make test
make lint
make compose-up
make compose-down
make logs
make smoke
```

Backend only:

```bash
cd apps/control-plane
go test ./...
go run .
```

Browser worker only:

```bash
cd workers/browser
npm install
npm run build
npm run dev
```

## Architecture Overview

```text
API client
   |
   v
qualora-api
   |
   +--> PostgreSQL: projects, test_runs, findings, evidence
   +--> Redis: browser run queue
   |
   v
qualora-worker-browser
   |
   +--> Playwright browser smoke test
   +--> MinIO/S3 screenshot storage
   +--> PostgreSQL evidence and findings
```

Detailed architecture notes live in [docs/architecture/mvp.md](docs/architecture/mvp.md).

The implemented API contract lives in [api/openapi/qualora.v1.yaml](api/openapi/qualora.v1.yaml).

## Repository Layout

```text
api/
  openapi/              Internal API contract definitions.
apps/
  control-plane/        Go API and orchestration service.
  web/                  Optional future web UI.
deploy/
  docker-compose/       Docker Compose notes.
  helm/                 Future Kubernetes packaging.
docs/
  architecture/         Architecture notes and implementation boundaries.
packages/
  report-engine/        Future report generation module.
  shared/               Shared schemas and utilities when needed.
scripts/                Developer automation and smoke checks.
workers/
  analyzer/             Future analyzer worker.
  api/                  Future API and OpenAPI checks.
  browser/              Playwright browser checks.
  security/             Future passive security checks.
```

## Security Warning And Scope Policy

Only run Qualora against systems you own or are explicitly authorized to test.

The MVP is safe by default:

- Browser navigation and network requests are constrained by project `allowed_hosts`.
- `localhost`, link-local, private IP ranges, and common cloud metadata endpoints are blocked by default.
- `security_mode` is currently limited to `passive`.
- `destructive_actions` must be `false`.
- Login automation and credential storage are intentionally not implemented in this phase.
- Secrets, credentials, cookies, and authorization headers must not be logged.
- Screenshots and reports should be treated as sensitive evidence artifacts.

Projects can set `allow_private_targets: true` for local/private test environments, but this should only be used for systems you control.

Report security issues using the process in [SECURITY.md](SECURITY.md).

## Current Limitations

- No web UI yet.
- No API worker yet.
- No OpenAPI contract checks yet.
- No passive security worker yet.
- No analyzer worker separate from the browser worker yet.
- No login automation or credential storage yet.
- The browser worker writes results directly to PostgreSQL in this MVP. A narrower worker result API can replace this later.
- Host safety checks block obvious unsafe literal hosts, but they do not yet perform DNS resolution checks against private addresses.
- MinIO uses local development credentials in Docker Compose.

## Roadmap

- Phase 1: Project foundation, contribution docs, security policy, and architecture boundaries.
- Phase 2: Docker Compose MVP with Go control plane, browser worker, screenshots, findings, and JSON reports.
- Phase 3: Harden run lifecycle, retries, worker result API, and artifact access URLs.
- Phase 4: API worker with basic checks and optional OpenAPI contract validation.
- Phase 5: Passive security checks constrained by allowed hosts and testing policy.
- Phase 6: Analyzer and report engine modules for richer structured findings.
- Phase 7: Optional web UI for project setup, run status, and reports.
- Phase 8: Helm chart and Kubernetes deployment hardening.

## License

Qualora is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

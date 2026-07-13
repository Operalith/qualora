# Qualora

**Open-source, self-hosted autonomous QA for web applications and APIs.**

Qualora is an open-source, self-hosted autonomous QA platform that runs browser-based and API smoke tests, collects evidence, and generates structured reports for web applications and APIs.

`v0.2.0-alpha` adds an alpha API worker and OpenAPI checks to the existing browser QA MVP. It remains intentionally small: Docker Compose, a Go control plane API, Playwright browser worker, API worker, PostgreSQL metadata, Redis queueing, MinIO evidence storage, and JSON reports.

## Current Alpha Capabilities

- Run locally with Docker Compose.
- Create QA projects through an API.
- Start runs that can include browser and API jobs.
- Execute Playwright Chromium checks against a configured frontend URL.
- Execute safe API checks against `api_base_url`.
- Fetch and parse OpenAPI 3.x JSON/YAML from `openapi_url`.
- Test only safe OpenAPI methods by default: `GET`, `HEAD`, and `OPTIONS`.
- Enforce project `allowed_hosts` for browser and API requests.
- Collect page title, screenshot evidence, browser observations, API observations, OpenAPI summaries, and API request evidence.
- Store metadata in PostgreSQL.
- Queue worker jobs with Redis.
- Store screenshots in MinIO/S3, with a local filesystem fallback.
- Generate structured JSON reports.

## Architecture

```text
API client / smoke script
        |
        v
qualora-api
        |
        +--> PostgreSQL: projects, test_runs, run_jobs, findings, evidence
        +--> Redis: browser and API run queues
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

Run the smoke tests:

```bash
make smoke
```

The smoke target includes:

- Browser smoke against `https://example.com`.
- API/OpenAPI smoke against a local mock API service started by the Makefile.

Stop the stack:

```bash
docker compose down
```

If local port `8080` is already in use:

```bash
QUALORA_API_PORT=18080 docker compose up -d --build
QUALORA_API_URL=http://localhost:18080 make smoke
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

Fetch the report:

```bash
curl -s "http://localhost:8080/api/v1/runs/${RUN_ID}/report" | python3 -m json.tool
```

## Report Example

An API/OpenAPI run includes API evidence alongside findings:

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
      "type": "api_observations",
      "uri": "inline://api-observations",
      "metadata": {
        "api_base_url": "http://mock-api:8080/",
        "openapi_url": "http://mock-api:8080/openapi.json",
        "checked_endpoints": 3,
        "failed_endpoints": 0,
        "safe_methods_only": true
      }
    },
    {
      "type": "openapi_summary",
      "uri": "inline://openapi-summary",
      "metadata": {
        "version": "3.0.3",
        "paths": 3,
        "safe_operations": 2,
        "skipped_unsafe_operations": 1
      }
    }
  ],
  "metadata": {
    "jobs": [
      {
        "kind": "api",
        "status": "completed"
      }
    ]
  }
}
```

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
- Authenticated API testing, login automation, and credential storage are not implemented in this release.
- Secrets, credentials, cookies, and authorization headers must not be logged.
- Screenshots and reports should be treated as sensitive evidence artifacts.

See [docs/security-model.md](docs/security-model.md) and [SECURITY.md](SECURITY.md).

## Current Limitations

- No web UI.
- No authentication.
- No authenticated API testing.
- No login automation or credential storage.
- No active security scanning.
- No destructive API testing by default.
- No full OpenAPI schema validation or schema fuzzing.
- No request body generation.
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
- Add artifact download or signed URL support.
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

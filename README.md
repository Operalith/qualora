# Qualora

**Open-source, self-hosted autonomous QA for web applications and APIs.**

Qualora is an open-source, self-hosted autonomous QA platform that runs browser-based smoke tests, collects evidence, and generates structured reports for web applications.

`v0.1.0-alpha` is an early MVP. It is intentionally small: a Docker Compose stack, a Go control plane API, a Playwright browser worker, PostgreSQL metadata, Redis queueing, MinIO evidence storage, and JSON reports.

## Current Alpha Capabilities

- Run locally with Docker Compose.
- Create QA projects through an API.
- Start browser smoke test runs.
- Execute Playwright Chromium checks against a configured frontend URL.
- Enforce project `allowed_hosts` for target URLs and browser requests.
- Collect page title, screenshot evidence, console errors, failed network requests, and blocked out-of-scope requests.
- Store metadata in PostgreSQL.
- Queue browser runs with Redis.
- Store screenshots in MinIO/S3, with a local filesystem fallback.
- Generate structured JSON reports.

## Architecture

```text
API client / smoke script
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
        +--> MinIO/S3 screenshot evidence
        +--> PostgreSQL findings and evidence metadata
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

Run the built-in smoke test against `https://example.com`:

```bash
make smoke
```

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

## Report Example

A clean run against `https://example.com` returns a report shaped like this:

```json
{
  "run_id": "d76f71c6-60d0-42fd-9bce-fae54ab6bef1",
  "project_id": "85db0da7-fd0d-44bc-ae5a-c41cd1f03d47",
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
      "id": "afcf6fd9-2512-45d0-a9a5-5d9073c55981",
      "type": "screenshot",
      "uri": "s3://qualora-evidence/runs/d76f71c6-60d0-42fd-9bce-fae54ab6bef1/screenshots/1783947796267.png",
      "metadata": {
        "page_title": "Example Domain",
        "status_code": 200
      }
    },
    {
      "id": "1c56226f-98cd-4953-afd1-086d8f37e056",
      "type": "browser_observations",
      "uri": "inline://browser-observations",
      "metadata": {
        "blocked_requests": [],
        "console_errors": [],
        "failed_requests": [],
        "load_error": "",
        "page_title": "Example Domain",
        "status_code": 200
      }
    }
  ],
  "metadata": {
    "page_title": "Example Domain"
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
- Browser navigation and network requests are constrained by `allowed_hosts`.
- `security_mode` is currently limited to `passive`.
- `destructive_actions` must be `false`.
- `localhost`, `.local`, loopback, link-local, private IP literal targets, common cloud metadata targets, and public hostnames resolving to blocked IP ranges are blocked by default.
- `allow_private_targets: true` may be used for local/private systems you control.
- Login automation and credential storage are not implemented in this release.
- Secrets, credentials, cookies, and authorization headers must not be logged.
- Screenshots and reports should be treated as sensitive evidence artifacts.

See [docs/security-model.md](docs/security-model.md) and [SECURITY.md](SECURITY.md).

## Current Limitations

- No web UI.
- No API worker.
- No OpenAPI contract checks.
- No passive security worker.
- No analyzer worker separate from the browser worker.
- No authentication.
- No login automation or credential storage.
- No active security scanning.
- No Helm/Kubernetes deployment.
- The browser worker writes results directly to PostgreSQL in this MVP.
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
- Add API checks and optional OpenAPI contract validation.
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

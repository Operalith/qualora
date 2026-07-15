# API Worker

Alpha worker for safe API and OpenAPI checks.

Note: imported OpenAPI spec management and safe API smoke result rows are implemented in the Go control plane. This worker remains the legacy project-level API job runner for `api_base_url` and `openapi_url`.

Responsibilities:

- Basic API reachability checks.
- Safe endpoint checks.
- Optional OpenAPI 3.x parsing and safe method checks.
- API error collection.

All outbound requests must respect project allowed hosts and redact credentials from logs and reports.

## Current Alpha Behavior

- Consumes jobs from Redis queue `API_RUN_QUEUE`.
- Checks `api_base_url` with a safe `GET` request.
- Fetches and parses `openapi_url` when provided.
- Tests only `GET`, `HEAD`, and `OPTIONS` operations by default.
- Skips unsafe methods such as `POST`, `PUT`, `PATCH`, and `DELETE`.
- Writes `api_observations`, `openapi_summary`, and `api_request` evidence rows.
- Creates findings for unreachable APIs, invalid OpenAPI documents, 5xx responses, unexpected status codes, obvious content type mismatches, and visible stack traces.

## Local Development

```bash
npm ci
npm run build
npm test
npm run dev
```

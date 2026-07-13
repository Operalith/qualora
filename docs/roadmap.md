# Roadmap

This roadmap is intentionally practical. Qualora should remain self-hosted and useful before it grows broader platform features.

## v0.1.0-alpha

Delivered:

- Docker Compose stack.
- Go control plane API.
- Project creation.
- Browser smoke run creation.
- Redis browser run queue.
- Playwright browser worker.
- Allowed-host enforcement.
- Screenshot evidence in MinIO.
- Browser observation evidence.
- Structured JSON reports.

## v0.2.0-alpha

Current alpha scope:

- API worker.
- `api_base_url` baseline checks.
- OpenAPI 3.x JSON/YAML fetch and parse.
- Safe OpenAPI method checks for `GET`, `HEAD`, and `OPTIONS`.
- API findings and evidence in structured reports.
- Per-run worker job tracking.
- Deterministic local mock API smoke test.

## Phase 3: Run And Worker Hardening

- Worker result API so workers do not write directly to PostgreSQL.
- Run retries and clearer failure states.
- Better per-job error reporting in the public API.
- Artifact download or signed URL support.
- Additional safety tests for DNS and worker request blocking.
- Better operational logs and container health checks.

## Phase 4: Deeper API Checks

- More OpenAPI validation.
- Response body/schema checks for safe methods.
- Configurable endpoint limits and path filters.
- Conservative authenticated API testing design.

## Phase 5: Passive Security Checks

- Passive security headers.
- Cookie flag checks.
- Mixed-content observations.
- TLS metadata where practical.

## Later

- Optional web UI.
- Login automation with safe credential storage.
- Helm chart after Docker Compose is stable.
- OWASP ZAP integration with explicit safe policies.

## Non-Goals For Now

- Hosted SaaS assumptions.
- Billing.
- Organizations and teams.
- Multi-tenancy.
- Enterprise RBAC.
- Active exploitation or destructive scanning.

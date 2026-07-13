# Roadmap

This roadmap is intentionally practical. Qualora should remain self-hosted and useful before it grows broader platform features.

## v0.1.0-alpha

Current alpha scope:

- Docker Compose stack.
- Go control plane API.
- Project creation.
- Browser smoke run creation.
- Redis run queue.
- Playwright browser worker.
- Allowed-host enforcement.
- Screenshot evidence in MinIO.
- Browser observation evidence.
- Structured JSON reports.

## Phase 3: MVP Hardening

- Worker result API so workers do not write directly to PostgreSQL.
- Run retries and clearer failure states.
- Artifact download or signed URL support.
- Better report metadata and report schema tests.
- Additional safety tests for DNS and worker request blocking.
- Better operational logs and container health checks.

## Phase 4: API Checks

- API worker for basic HTTP checks.
- Optional OpenAPI contract fetch and validation.
- API error evidence and findings.

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

# AGENTS.md

Project-specific guidance for future Codex work in Qualora.

## Product Direction

Qualora is an open-source, self-hosted autonomous QA platform for web applications and APIs. Keep the MVP small, useful, and runnable on-prem with Docker Compose before adding Kubernetes, SaaS, enterprise, or marketplace features.

Default positioning: **Open-source AI-powered engineering tools for modern software operations.**

## Current Priorities

- Keep the `v0.1.0-alpha` Docker Compose MVP working.
- Backend/control plane first, with browser worker support.
- Docker Compose as the first deployment target.
- PostgreSQL for durable metadata.
- Redis for queues and short-lived run state.
- MinIO/S3-compatible storage for evidence artifacts.
- Playwright for real browser automation.
- Modular worker boundaries, but avoid premature distributed-system complexity.

## Repository Boundaries

- `apps/control-plane`: Go API and orchestration service.
- `apps/web`: Optional web UI, added only when the backend loop is useful.
- `workers/browser`: Playwright browser checks.
- `workers/api`: API and OpenAPI checks.
- `workers/security`: Passive, safe security checks.
- `workers/analyzer`: Evidence normalization and finding generation.
- `packages/report-engine`: Structured report generation.
- `packages/shared`: Shared schemas/utilities only when duplication becomes real.
- `api/openapi`: Internal API contracts.
- `deploy/docker-compose`: First supported deployment path.
- `deploy/helm`: Future Kubernetes packaging.

## Implementation Rules

- Prefer simple, explicit code over framework-heavy abstractions.
- Do not introduce paid SaaS assumptions.
- Do not add Kubernetes-only concepts before the Docker Compose path works.
- Do not add a full UI until the API and worker loop are more stable.
- Do not introduce Temporal, OWASP ZAP, login automation, or active security scanning in the MVP without an explicit request.
- Keep worker contracts narrow and serializable.
- Prefer OpenAPI-first internal API design where practical.
- Keep report schemas structured enough for future UI/API consumers.
- Add tests around orchestration, host allowlisting, secret redaction, and report generation when those areas are implemented.
- Do not claim unsupported features in README, release notes, OpenAPI, or docs.

## Security And Safety Rules

- Never log test credentials, tokens, cookies, authorization headers, or secret values.
- Redact secrets in errors, traces, reports, and debug output.
- All browser, API, and security checks must enforce project allowed hosts.
- Security checks are passive and non-destructive by default.
- Do not add active exploitation, destructive payloads, brute force behavior, or broad crawling unless the user explicitly asks and a safe policy model exists.
- Treat screenshots, traces, and reports as sensitive artifacts.
- Store credentials through a dedicated abstraction so local MVP storage can later move to Vault or Kubernetes Secrets.

## Contribution Style

- Update `README.md` or `docs/architecture/mvp.md` when changing product architecture.
- Keep PRs focused and explain user-facing behavior changes.
- When adding a service, include local run instructions and Docker Compose integration.
- When adding a worker, document its inputs, outputs, safety checks, and artifact behavior.

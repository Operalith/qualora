# Qualora

Qualora is an open-source, self-hosted autonomous QA platform for web applications and APIs.

It is being built as part of Operalith's mission: **open-source AI-powered engineering tools for modern software operations**. The first public release should feel like a practical autonomous QA engineer that teams can run on-prem with Docker Compose.

## MVP Scope

The first release focuses on a small, useful loop:

1. Create a project with a frontend URL, API base URL, optional OpenAPI URL, test credentials, allowed hosts, and a testing policy.
2. Start a test run.
3. Execute safe checks:
   - Browser smoke tests with Playwright.
   - Basic API checks.
   - Optional OpenAPI contract checks.
   - Passive, non-destructive security checks.
4. Collect evidence:
   - Screenshots.
   - Playwright traces when available.
   - Console errors.
   - Failed network requests.
   - API errors.
5. Generate a structured report with findings, severity, reproduction steps, evidence, and recommendations.

Qualora is not intended to be a broad scanner or destructive testing system. Safety, host allowlisting, and secret handling are core product constraints.

## Architecture Overview

The MVP is organized around a Go control plane, focused workers, and S3-compatible evidence storage.

```text
User / UI / API client
        |
        v
Go control plane API
        |
        +--> PostgreSQL: projects, runs, findings, metadata
        +--> Redis: queue and short-lived run state
        +--> MinIO/S3: screenshots, traces, logs, reports
        |
        +--> browser worker: Playwright smoke checks
        +--> API worker: HTTP checks and OpenAPI contract checks
        +--> security worker: passive safe checks within allowed hosts
        +--> analyzer worker: normalize evidence into findings
        +--> report engine: structured reports
```

Detailed MVP architecture notes live in [docs/architecture/mvp.md](docs/architecture/mvp.md).

## Repository Layout

```text
api/
  openapi/              Internal API contract definitions.
apps/
  control-plane/        Go API and orchestration service.
  web/                  Optional web UI, deferred until the backend loop is useful.
deploy/
  docker-compose/       Local self-hosted deployment target for the first release.
  helm/                 Future Kubernetes packaging.
docs/
  architecture/         Architecture notes and implementation boundaries.
packages/
  report-engine/        Report generation module.
  shared/               Shared schemas and utilities when needed.
scripts/                Developer automation.
workers/
  analyzer/             Evidence normalization and finding generation.
  api/                  API and OpenAPI checks.
  browser/              Playwright browser checks.
  security/             Passive, safe security checks.
```

## Quick Start

The runnable stack is not implemented yet. The intended local workflow for the first release is:

```bash
git clone https://github.com/Operalith/qualora.git
cd qualora
docker compose -f deploy/docker-compose/docker-compose.yml up --build
```

Until the Docker Compose stack exists, use this repository as the project foundation and architecture reference.

## Roadmap

- Phase 0: Project foundation, contribution docs, security policy, and architecture boundaries.
- Phase 1: Go control plane with project and test run APIs.
- Phase 2: Docker Compose stack with PostgreSQL, Redis, MinIO, and the control plane.
- Phase 3: Browser worker with Playwright smoke tests and evidence capture.
- Phase 4: API worker with basic checks and optional OpenAPI contract validation.
- Phase 5: Passive security checks constrained by allowed hosts and testing policy.
- Phase 6: Analyzer and report engine for structured findings.
- Phase 7: Optional web UI for project setup, run status, and reports.
- Phase 8: Helm chart and Kubernetes deployment hardening.

## Security Warning And Scope Policy

Only run Qualora against systems you own or are explicitly authorized to test.

The MVP must be safe by default:

- Respect project-level allowed hosts for every browser, API, and security check.
- Avoid destructive actions unless a future policy explicitly enables them.
- Do not perform aggressive scanning in the first release.
- Do not log secrets, credentials, tokens, cookies, or authorization headers.
- Store credentials behind an abstraction so local MVP storage can later move to Vault, Kubernetes Secrets, or another secret manager.
- Treat screenshots, traces, and reports as sensitive evidence artifacts.

Report security issues using the process in [SECURITY.md](SECURITY.md).

## License

Qualora is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

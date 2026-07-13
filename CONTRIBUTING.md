# Contributing To Qualora

Thanks for helping build Qualora.

Qualora is early-stage. The best contributions keep the first release small, self-hosted, and safe by default.

## Development Principles

- Prefer a working Docker Compose MVP before broader deployment targets.
- Keep changes focused and easy to review.
- Avoid enterprise-only assumptions, paid-service dependencies, and hosted-SaaS defaults.
- Write clear docs when adding architecture, services, workers, or runbook behavior.
- Treat credentials and evidence artifacts as sensitive.

## Getting Started

Clone the repository and start the local stack:

```bash
git clone https://github.com/Operalith/qualora.git
cd qualora
docker compose up -d --build
make smoke
```

Development details live in [docs/development.md](docs/development.md).

## Pull Requests

Before opening a pull request:

- Keep the scope narrow.
- Update docs for behavior or architecture changes.
- Add or update tests when implementation code exists.
- Confirm secrets are not logged or committed.
- Explain any security-relevant behavior in the PR description.

## Code Style

Language-specific conventions will be added as the codebase grows.

Expected defaults:

- Go code should be formatted with `gofmt`.
- TypeScript should pass `npm run build` and `npm test` in `workers/browser`.
- Public APIs should be documented through OpenAPI where practical.

## Validation

Before opening a pull request, run:

```bash
make test
make lint
docker compose config
```

Run `make smoke` when changing the API, Docker Compose stack, queue behavior, browser worker, evidence handling, or reports.

## Reporting Security Issues

Do not report security vulnerabilities in public issues. Follow [SECURITY.md](SECURITY.md).

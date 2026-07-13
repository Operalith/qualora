# Release Process

Qualora v0.1.0-alpha is the first public alpha release. It should be published as an early, self-hosted MVP, not as a complete QA platform.

## Pre-Release Checklist

Run from the repository root:

```bash
make test
make lint
docker compose config
docker compose up -d --build
make smoke
docker compose down -v
```

Confirm:

- The API returns `{"status":"ok"}` from `/healthz`.
- The smoke script creates a project and run.
- The run reaches `completed`.
- The report includes screenshot and browser observation evidence.
- Screenshot evidence uses an `s3://qualora-evidence/...` URI when MinIO is healthy.
- Documentation does not claim unsupported UI, auth, API checks, login automation, or active security scanning.

## Tagging

```bash
git status --short
git add .
git commit -m "chore: prepare v0.1.0-alpha release"
git tag -a v0.1.0-alpha -m "v0.1.0-alpha"
git push origin main
git push origin v0.1.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.1.0-alpha
```

Use [release-notes/v0.1.0-alpha.md](release-notes/v0.1.0-alpha.md) as the release body.

## Version Scope

This release includes:

- Docker Compose MVP.
- Go control plane API.
- PostgreSQL metadata storage.
- Redis run queue.
- Playwright browser worker.
- MinIO screenshot evidence storage.
- Structured JSON reports.

This release does not include:

- Web UI.
- API worker.
- OpenAPI contract testing.
- Authentication.
- Login automation.
- Active security scanning.
- Helm/Kubernetes deployment.

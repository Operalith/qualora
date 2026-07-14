# Release Process

Qualora v0.3.0-alpha is the third public alpha release. It adds a minimal self-hosted web UI and self-contained HTML reports to the browser/API QA foundation.

## Pre-Release Checklist

Run from the repository root:

```bash
make test
make lint
docker compose config
docker compose up -d --build
make smoke
curl -s http://localhost:3000/healthz
docker compose down -v
```

Confirm:

- The API returns `{"status":"ok"}` from `/healthz`.
- The web UI is reachable at `http://localhost:3000`.
- The smoke script creates browser and API projects.
- Browser and API runs reach `completed`.
- Browser reports include screenshot and browser observation evidence.
- API reports include `api_observations`, `openapi_summary`, and `api_request` evidence.
- JSON report URLs work.
- HTML report URLs work and render a self-contained report.
- Screenshot evidence uses an `s3://qualora-evidence/...` URI when MinIO is healthy.
- Documentation does not claim unsupported auth, login automation, active security scanning, destructive API testing, schema fuzzing, or screenshot preview/download through the API.

## Tagging

```bash
git status --short
git add .
git commit -m "feat: add web UI and HTML reports for v0.3.0-alpha"
git tag -a v0.3.0-alpha -m "v0.3.0-alpha"
git push origin main
git push origin v0.3.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.3.0-alpha
```

Use [release-notes/v0.3.0-alpha.md](release-notes/v0.3.0-alpha.md) as the release body.

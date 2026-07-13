# Release Process

Qualora v0.2.0-alpha is the second public alpha release. It adds safe API and OpenAPI checks to the browser QA MVP.

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
- The smoke script creates browser and API projects.
- Browser and API runs reach `completed`.
- Browser reports include screenshot and browser observation evidence.
- API reports include `api_observations`, `openapi_summary`, and `api_request` evidence.
- Screenshot evidence uses an `s3://qualora-evidence/...` URI when MinIO is healthy.
- Documentation does not claim unsupported UI, auth, login automation, active security scanning, destructive API testing, or schema fuzzing.

## Tagging

```bash
git status --short
git add .
git commit -m "feat: add API worker for v0.2.0-alpha"
git tag -a v0.2.0-alpha -m "v0.2.0-alpha"
git push origin main
git push origin v0.2.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.2.0-alpha
```

Use [release-notes/v0.2.0-alpha.md](release-notes/v0.2.0-alpha.md) as the release body.

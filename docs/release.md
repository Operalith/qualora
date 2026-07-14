# Release Process

Qualora v0.5.0-alpha is the fifth public alpha release. It adds optional OpenAI-compatible AI provider management, provider testing, sanitized AI report analysis, AI report display, and a deterministic fake LLM smoke target.

## Pre-Release Checklist

Run from the repository root:

```bash
make test
make lint
docker compose config
docker compose up -d --build
make smoke
curl -s http://localhost:3000/healthz
curl -s http://localhost:8080/healthz
git diff --check
docker compose down -v
```

Confirm:

- The API returns `{"status":"ok"}` from `/healthz`.
- The web UI is reachable at `http://localhost:3000`.
- The smoke script starts the local `demo-web` and `mock-api` smoke services.
- The smoke script creates browser and API projects.
- Browser and API runs reach `completed`.
- Browser reports include screenshot and browser observation evidence.
- Screenshot evidence metadata includes filename, key, content type, size, and created timestamp.
- `GET /api/v1/evidence/{evidence_id}` returns the stored screenshot with an image content type.
- The smoke script creates and tests a fake OpenAI-compatible provider.
- AI analysis completes for the browser smoke run.
- JSON reports include `ai_analysis` when analysis has been generated.
- HTML reports include an AI Analysis section when analysis has been generated.
- API reports include `api_observations`, `openapi_summary`, and `api_request` evidence.
- JSON report URLs work.
- HTML report URLs work and render a self-contained report.
- Screenshot evidence uses an `s3://qualora-evidence/...` URI when MinIO is healthy.
- Documentation does not claim unsupported auth, login automation, active security scanning, destructive API testing, schema fuzzing, trace export, autonomous AI browser control, native Anthropic/Gemini support, or full browser/API test coverage.

## Tagging

```bash
git status --short
git add .
git commit -m "feat: add optional AI report analysis for v0.5.0-alpha"
git tag -a v0.5.0-alpha -m "v0.5.0-alpha"
git push origin main
git push origin v0.5.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.5.0-alpha
```

Use [release-notes/v0.5.0-alpha.md](release-notes/v0.5.0-alpha.md) as the release body.

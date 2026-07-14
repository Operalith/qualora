# Release Process

Qualora v0.7.0-alpha is the seventh public alpha release. It adds approved safe test plan execution, persisted execution step reports, execution HTML reports, web UI execution workflows, and deterministic smoke coverage while keeping AI-generated plans as suggestions and deterministic QA useful without AI.

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
- AI analysis completes for the browser and API smoke runs.
- AI test plan generation completes for the browser and API smoke runs.
- Test plan detail and export JSON URLs work.
- Safe test plan preview returns executable browser DSL steps for the browser smoke plan.
- Safe test plan execution reaches `completed`.
- Safe execution reports include scenarios, steps, screenshot evidence, browser observations, and HTML report output.
- JSON reports include `ai_analysis` when analysis has been generated.
- HTML reports include an AI Analysis section when analysis has been generated.
- JSON and HTML reports include related test plan references when a plan was generated from a run.
- API reports include `api_observations`, `openapi_summary`, and `api_request` evidence.
- JSON report URLs work.
- HTML report URLs work and render a self-contained report.
- Screenshot evidence uses an `s3://qualora-evidence/...` URI when MinIO is healthy.
- Documentation does not claim unsupported auth, login automation, active security scanning, destructive API testing, schema fuzzing, trace export, autonomous AI browser control, automatic/free-form execution of generated test plans, native Anthropic/Gemini support, or full browser/API test coverage.

## Tagging

```bash
git status --short
git add .
git commit -m "feat: add safe test plan execution for v0.7.0-alpha"
git tag -a v0.7.0-alpha -m "v0.7.0-alpha"
git push origin main
git push origin v0.7.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.7.0-alpha
```

Use [release-notes/v0.7.0-alpha.md](release-notes/v0.7.0-alpha.md) as the release body.

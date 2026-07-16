# Release Process

Qualora v0.14.0-alpha is the fourteenth public alpha release. It adds passive front-end quality checks for security, accessibility, and performance signals while keeping local auth, browser/API smoke, credential profiles, authorization checks, application discovery, AI analysis, Safe QA Runs, and approved safe plan execution intact.

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
- A fresh database requires first-run admin setup before project data is visible.
- Setup creates a local admin user and does not return the password or password hash.
- A second setup attempt is rejected after the admin user exists.
- Login succeeds with the local admin account and sets HTTP-only session plus CSRF cookies.
- `/api/v1/auth/me` reports the authenticated admin session.
- Protected endpoints reject unauthenticated requests.
- Logout clears the session and protected endpoints are rejected afterward.
- The smoke script starts the local `demo-web`, `demo-api`, and `fake-llm` smoke services.
- The smoke script creates browser and API projects.
- The smoke script creates a demo credential profile.
- The smoke script creates demo role credential profiles.
- Credential profile API responses do not include the raw demo username or password.
- The login check run reaches `completed` against `demo-web`.
- The authenticated browser smoke run reaches `completed` against `demo-web` `/dashboard`.
- Authorization checks are created for admin, readonly, customer-a, and customer-b demo roles.
- The authorization check run reaches `completed`.
- Authorization JSON and HTML reports work and do not contain demo passwords.
- Authorization reports include `authorization_observations` and screenshot evidence.
- Application discovery reaches `completed` against `demo-web`.
- Discovery JSON report and application map endpoints include pages, links, forms, findings, and evidence metadata.
- Discovery records skipped unsafe and external links.
- Discovery pages include screenshot evidence IDs and downloadable screenshot evidence.
- Discovery HTML report includes the application discovery summary, pages, skipped links, forms, findings, and safety notes.
- Passive quality check runs reach `completed` against `demo-web`.
- Quality JSON reports include security, accessibility, and performance finding counts.
- Quality HTML reports include the quality summary, findings, safety notes, and limitations.
- Quality reports do not contain demo passwords, cookies, auth headers, browser storage, request bodies, response bodies, full HTML, or raw credential values.
- Discovery-aware AI test plan generation completes from the completed discovery run.
- Discovery-generated test plans include `source_type=discovery`, `discovery_run_id`, and safe execution coverage metadata.
- Safe QA Run preview completes without automatically executing browser actions.
- Safe QA Run preview report includes discovery, optional quality checks, generated plan, safe execution preview, coverage, and safety notes.
- Explicit Safe QA Run execution starts only after `POST /api/v1/qa-runs/{qa_run_id}/execute` or `execute=true`.
- Safe QA Run execution reaches `completed`.
- Safe QA Run JSON and HTML reports include the linked execution report when execution has run.
- Safe QA Run reports do not contain demo passwords, cookies, tokens, browser storage, auth headers, full HTML, request bodies, or response bodies.
- Login and authenticated smoke JSON reports include `login_summary` and `login_observations`.
- Login and authenticated smoke HTML reports include the login summary.
- Login and authenticated smoke reports do not contain the demo password.
- Browser and safe API smoke runs reach `completed`.
- The demo OpenAPI spec imports with status `parsed`.
- Operation discovery includes skipped POST/DELETE/auth-required operations.
- The API smoke run records API result rows.
- The API smoke report includes a deterministic 5xx finding from `/broken`.
- JSON and HTML API smoke reports include API result tables.
- API smoke reports do not store response bodies.
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
- AI analysis and AI-assisted test planning still work for authenticated browser smoke reports without sending credentials.
- AI test planning from discovery uses sanitized application map data only and does not send screenshots/full HTML/cookies/browser storage/auth headers/tokens/credentials.
- API reports include `api_observations`, `openapi_summary`, `api_request` evidence, `api_summary`, and `api_results`.
- JSON report URLs work.
- HTML report URLs work and render a self-contained report.
- Screenshot evidence uses an `s3://qualora-evidence/...` URI when MinIO is healthy.
- Documentation does not claim unsupported multi-user management, password reset, SSO/OIDC/SAML, enterprise RBAC, multi-tenancy, authenticated API testing, arbitrary login automation, active security scanning, destructive API testing, schema fuzzing, trace export, autonomous AI browser control, automatic/free-form execution of generated test plans, native Anthropic/Gemini support, or full browser/API test coverage.

## Tagging

```bash
git status --short
git add .
git commit -m "feat: add passive quality checks for v0.14.0-alpha"
git tag -a v0.14.0-alpha -m "v0.14.0-alpha"
git push origin main
git push origin v0.14.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.14.0-alpha
```

Use [release-notes/v0.14.0-alpha.md](release-notes/v0.14.0-alpha.md) as the release body.

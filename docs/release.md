# Release Process

Qualora v0.17.0-alpha is the seventeenth public alpha release. It adds deterministic report intelligence, severity normalization, finding grouping/deduplication, noise classification, affected-page summaries, and executive summaries while keeping local auth, browser/API smoke, credential profiles, authorization checks, application discovery, passive quality checks, Safe Explorer, guided onboarding, AI analysis, Safe QA Runs, and approved safe plan execution intact.

## Pre-Release Checklist

Run from the repository root:

```bash
make test
make lint
docker compose config
docker compose up -d --build
make smoke
curl -fsS http://localhost:3000/healthz
curl -fsS http://localhost:8080/healthz
git diff --check
docker compose down -v
```

Confirm:

- The API returns `{"status":"ok"}` from `/healthz`.
- The web UI is reachable at `http://localhost:3000`.
- A fresh database requires first-run admin setup before project data is visible.
- Setup, login, logout, `/auth/me`, CSRF, and protected route checks pass.
- The dashboard shows the `v0.17.0-alpha` badge, quick-start actions, status indicators, recent projects, and recent Safe QA runs.
- The guided setup route `#/setup-project` renders the project basics, AI, login, OpenAPI, workflow, and results steps.
- `POST /api/v1/onboarding/project-setup` can create a project, optionally configure a demo AI provider, optionally create a credential profile, optionally import a demo OpenAPI spec, and start selected safe checks.
- The guided demo flow starts browser smoke, authenticated browser smoke, discovery, quality checks, Safe QA, and API smoke when all demo dependencies are configured.
- The guided setup response returns IDs, report links, skipped reasons, and safe metadata only.
- Guided setup responses and reports do not contain demo passwords, raw provider secrets, cookies, browser storage, authorization headers, or tokens.
- Project detail pages show the readiness checklist for frontend URL, AI provider, discovery, Safe Explorer, quality checks, credentials, OpenAPI, Safe QA, and reports.
- Project detail pages show the Interactive Safe Explorer card, run form, run list, and warning text.
- Safe Explorer runs complete against `demo-web`, execute at least one safe action, skip unsafe/external/POST/unsupported actions with reasons, and produce screenshot evidence.
- Safe Explorer JSON/HTML reports and trace endpoints work and do not expose demo passwords, cookies, browser storage, auth headers, or tokens.
- The reports landing page lists recent browser, API, discovery, Safe Explorer, quality, and Safe QA reports with status, high/medium counts, grouped counts, raw counts, and report links for recent reports.
- JSON reports include `executive_summary`, `severity_counts`, `grouped_findings`, `top_findings`, `top_affected_pages`, `noise_summary`, `raw_findings_count`, `deduplication_summary`, and `safety_limitations`.
- HTML reports include Executive Summary, Grouped Findings, Affected Pages, Noise / Repeated Findings, and the existing raw details.
- Existing browser smoke reports still include screenshot and browser observation evidence.
- Existing login and authenticated smoke reports still include `login_summary` and `login_observations`.
- Existing authorization checks and reports still work and remain deterministic.
- Existing discovery JSON/HTML reports and application maps still work.
- Existing quality check JSON/HTML reports still work and remain passive metadata checks.
- Existing API smoke reports still include `api_observations`, `openapi_summary`, `api_request`, `api_summary`, and `api_results`.
- Existing AI analysis, AI test planning, discovery-aware planning, safe plan preview/execution, and Safe QA Run flows still work.
- Documentation does not claim unsupported multi-user management, password reset, SSO/OIDC/SAML, enterprise RBAC, multi-tenancy, authenticated API testing, arbitrary login automation, active security scanning, destructive API testing, schema fuzzing, trace export, autonomous AI browser control, automatic/free-form execution of generated test plans, native Anthropic/Gemini support, or full browser/API test coverage.

## Tagging

```bash
git status --short
git add .
git commit -m "feat: add report intelligence for v0.17.0-alpha"
git tag -a v0.17.0-alpha -m "v0.17.0-alpha"
git push origin main
git push origin v0.17.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.17.0-alpha
```

Use [release-notes/v0.17.0-alpha.md](release-notes/v0.17.0-alpha.md) as the release body.

# Release Process

Qualora v0.24.0-alpha is the twenty-fourth public alpha release. It simplifies the authenticated experience with a focused dashboard and Project Cockpit, adds a near-live/replay Run Viewer for AI Browser Control and Safe Explorer, and provides an optional real OpenAI-compatible Demo Lab workflow.

## Pre-Release Checklist

Run from the repository root:

```bash
make test
make lint
docker compose config
docker compose up -d --build
make smoke
make showcase-smoke
bash -n scripts/run-demo-lab-real-llm.sh
scripts/test-real-llm-script.sh
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
- The dashboard shows the `v0.24.0-alpha` badge, `Run Demo Lab Showcase`, `Create Real Project`, `Open Reports`, and compact status indicators.
- Project details render the Project Cockpit with primary QA actions and preserve specialized features under Advanced Tools.
- AI Browser Control and Safe Explorer runs open in the near-live/replay Run Viewer.
- Run Viewer timelines include action metadata, latest screenshot links, AI suggestions and policy decisions where applicable, execution results, and blocked/skipped reasons.
- The dashboard and guided setup expose Demo Lab showcase copy, Demo Lab target defaults, and Fake LLM versus configured real-provider choices.
- Missing real-provider environment variables fail before services start, name only the missing variables, and never print the supplied API key.
- `demo-lab-web`, `demo-lab-api`, and `fake-llm` become healthy under the `demo-lab` profile.
- `make showcase-smoke` validates the full browser, API, auth, discovery, quality, form, AI, authorization, Safe QA, baseline, CI gate, and issue-export-preview workflow against Demo Lab.
- Demo Lab reports, CI output, issue previews, and logs do not expose fake passwords, bearer tokens, cookies, browser storage, auth headers, or session material.
- The guided setup route `#/setup-project` renders the project basics, AI, login, OpenAPI, workflow, and results steps.
- `POST /api/v1/onboarding/project-setup` can create a project, optionally configure a demo AI provider, optionally create a credential profile, optionally import a demo OpenAPI spec, and start selected safe checks.
- The guided demo flow starts browser smoke, authenticated browser smoke, discovery, quality checks, Safe QA, and API smoke when all demo dependencies are configured.
- The guided setup response returns IDs, report links, skipped reasons, and safe metadata only.
- Guided setup responses and reports do not contain demo passwords, raw provider secrets, cookies, browser storage, authorization headers, or tokens.
- Project detail pages show the readiness checklist for frontend URL, AI provider, discovery, Safe Explorer, quality checks, credentials, OpenAPI, Safe QA, and reports.
- Project detail pages show the Interactive Safe Explorer card, run form, run list, and warning text.
- Safe Explorer runs complete against `demo-web`, execute at least one safe action, skip unsafe/external/POST/unsupported actions with reasons, and produce screenshot evidence.
- Safe Explorer JSON/HTML reports and trace endpoints work and do not expose demo passwords, cookies, browser storage, auth headers, or tokens.
- AI Browser Control runs complete against `demo-web` using `fake-llm`, execute approved safe actions, record AI suggestions, policy decisions, sanitized observations, screenshot evidence, trace data, JSON reports, and HTML reports.
- The unsafe AI Browser Control smoke goal makes `fake-llm` propose a destructive route and Qualora records a policy block instead of executing it.
- The AI Browser Control safe-form smoke goal makes `fake-llm` propose `submit_safe_get_form` for the demo search form and Qualora executes only the policy-approved same-origin GET URL.
- The AI Browser Control unsafe-form smoke goal makes `fake-llm` propose an unsafe form action and Qualora records a policy block instead of submitting it.
- AI Browser Control reports do not expose demo passwords, cookies, browser storage, auth headers, tokens, screenshots to AI, full HTML, request bodies, or response bodies.
- Project detail pages show the Safe Form Testing card, run form, run list, and warning text.
- Safe Form Testing runs complete against `demo-web`, test safe same-origin GET search/filter forms, skip POST, login/password, dangerous, external, sensitive, and unsupported forms with reasons, and produce screenshot evidence.
- Safe Form Testing JSON/HTML reports include summary, classifications, tested forms, skipped forms, findings, evidence, report intelligence, safety notes, and limitations.
- Safe Form Testing reports and evidence do not contain demo passwords, raw form values, cookies, browser storage, auth headers, tokens, full HTML, request bodies, or response bodies.
- The reports landing page lists recent browser, API, discovery, Safe Explorer, AI Browser Control, Safe Form Testing, quality, and Safe QA reports with status, high/medium counts, grouped counts, raw counts, and report links for recent reports.
- JSON reports include `executive_summary`, `severity_counts`, `grouped_findings`, `top_findings`, `top_affected_pages`, `noise_summary`, `raw_findings_count`, `deduplication_summary`, and `safety_limitations`.
- HTML reports include Executive Summary, Grouped Findings, Affected Pages, Noise / Repeated Findings, and the existing raw details.
- A completed Safe QA report can be marked as the default `safe_qa` baseline.
- Creating a new default baseline unsets the previous default for the same project and report type.
- A second Safe QA report can be compared against the default baseline and returns deterministic new/fixed/unchanged grouped findings.
- Safe QA JSON and HTML reports show baseline/comparison/gate information when a default baseline exists.
- The quality gate endpoint returns pass/fail/warning JSON and compact `format=ci` JSON with an exit code.
- `scripts/qualora-ci-gate.sh` exits with the compact quality gate exit code.
- `POST /api/v1/projects/{project_id}/ci-runs` can reuse the latest completed Safe QA report, compare it with a baseline, evaluate the quality gate, persist `ci_runs`, and return exit code `0`.
- `scripts/qualora-ci-run.sh` logs in, starts or reuses a CI run, prints compact JSON, and exits with the Qualora CI exit code.
- Issue export configs can be created with encrypted GitHub/GitLab tokens, listed without raw tokens, and tested without creating issues.
- `POST /api/v1/reports/{report_type}/{report_id}/export-issues` defaults to dry-run issue previews from grouped sanitized findings.
- CI run output, issue export config responses, issue previews, JSON reports, and HTML reports do not contain the demo password, session cookies, CSRF tokens, provider secrets, tracker tokens, auth headers, browser storage, screenshots, full HTML, request bodies, or response bodies.
- The web UI renders Baselines & Regression cards, Set as baseline, Compare with baseline, Evaluate quality gate, status badges, and failed rule lists.
- The web UI renders CI Run and Issue Export sections on project pages and Export issues controls on Safe QA report pages.
- Existing browser smoke reports still include screenshot and browser observation evidence.
- Existing login and authenticated smoke reports still include `login_summary` and `login_observations`.
- Existing authorization checks and reports still work and remain deterministic.
- Existing discovery JSON/HTML reports and application maps still work.
- Existing quality check JSON/HTML reports still work and remain passive metadata checks.
- Existing API smoke reports still include `api_observations`, `openapi_summary`, `api_request`, `api_summary`, and `api_results`.
- API auth profiles can be created, listed, read, updated, deleted, and tested without returning bearer tokens, API keys, basic auth credentials, Authorization headers, or encrypted payloads.
- The demo API bearer profile using `demo-api-token` succeeds against `/private/profile`.
- Authenticated API smoke executes safe protected `GET` operations from the imported demo OpenAPI spec and records `api_auth`, `authenticated_operations`, and unauthenticated comparison status.
- Lightweight contract validation records the deterministic `/private/broken-contract` required-field mismatch without storing response bodies.
- Authenticated API smoke JSON reports, HTML reports, API result rows, AI analysis inputs/results, CI output, and issue export dry-run previews do not contain `demo-api-token`.
- Existing AI analysis, AI test planning, discovery-aware planning, safe plan preview/execution, and Safe QA Run flows still work.
- Documentation does not claim unsupported multi-user management, password reset, SSO/OIDC/SAML, enterprise RBAC, multi-tenancy, destructive API testing, full OpenAPI validation, schema fuzzing, trace export beyond existing alpha report traces, autonomous/direct AI browser control, automatic/free-form execution of generated test plans, native Anthropic/Gemini support, or full browser/API test coverage.

## Tagging

```bash
git status --short
git add .
git commit -m "feat: improve demo UX and browser run visibility for v0.24.0-alpha"
git tag -a v0.24.0-alpha -m "v0.24.0-alpha"
git push origin main
git push origin v0.24.0-alpha
```

## GitHub Release

Suggested title:

```text
Qualora v0.24.0-alpha
```

Use [release-notes/v0.24.0-alpha.md](release-notes/v0.24.0-alpha.md) as the release body.

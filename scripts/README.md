# Scripts

Developer automation will live here.

Scripts should be safe to run locally, documented, and avoid printing secrets.

Current scripts:

- `smoke.py`: end-to-end setup/login/logout, protected endpoint, guided project setup, dashboard/readiness/report UI text, browser, credential profile, login, authenticated smoke, application discovery, passive quality, role-aware authorization, OpenAPI import, safe API smoke, AI-analysis, AI test plan, Safe QA, Safe QA baseline/comparison/quality gate, native CI runs, both CI helper scripts, issue export config/test, dry-run issue export previews, and safe execution smoke test driver. It prints JSON and HTML report URLs, validates HTML report export, validates protected report/evidence access, validates API result rows, validates skipped unsafe API operations, validates quality finding counts, validates baseline comparison and compact CI gate output, validates screenshot evidence download, verifies password/token redaction, verifies AI analysis with the fake provider, and executes deterministic safe browser plans.
- `qualora-ci-gate.sh`: small HTTP-based CI helper that evaluates a quality gate for an existing report and exits with the API-provided exit code. It can log in with `QUALORA_EMAIL` and `QUALORA_PASSWORD`, or use an existing session cookie/CSRF token.
- `qualora-ci-run.sh`: pipeline-friendly helper that logs in, starts a native CI run, and exits with the API-provided exit code. Set `QUALORA_RUN_SAFE_QA=false` to reuse the latest completed Safe QA report without AI. Issue export is off by default; if enabled, `QUALORA_ISSUE_EXPORT_DRY_RUN=true` is the safe default.
- `demo-api/server.js`: deterministic local API and OpenAPI spec used by `make smoke`.
- `mock-api/server.js`: older deterministic local API retained for compatibility with earlier alpha API worker checks.
- `demo-web/server.js`: deterministic local frontend used by browser, login, authenticated smoke, application discovery, passive quality, authorization, Safe QA, and safe test plan smoke tests.
- `fake-llm/server.js`: deterministic OpenAI-compatible chat completions provider used by `make smoke`.
- `demo-lab/web`: dedicated showcase web target with authentication, roles, safe/unsafe forms, discovery actions, and passive quality fixtures.
- `demo-lab/api`: dedicated showcase OpenAPI target with public/authenticated safe reads, contract mismatch, server-error, and skipped mutation fixtures.
- `run-demo-lab.sh`: starts the `demo-lab` profile and runs the comprehensive end-to-end showcase without printing configured secrets.

Use `make showcase-smoke` to run the existing smoke assertions against Demo Lab, or `scripts/run-demo-lab.sh` to start the complete stack and validate it in one command.

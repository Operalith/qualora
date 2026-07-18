# Scripts

Developer automation will live here.

Scripts should be safe to run locally, documented, and avoid printing secrets.

Current scripts:

- `smoke.py`: end-to-end setup/login/logout, protected endpoint, guided project setup, dashboard/readiness/report UI text, browser, credential profile, login, authenticated smoke, application discovery, passive quality, role-aware authorization, OpenAPI import, safe API smoke, AI-analysis, AI test plan, Safe QA, Safe QA baseline/comparison/quality gate, and safe execution smoke test driver. It prints JSON and HTML report URLs, validates HTML report export, validates protected report/evidence access, validates API result rows, validates skipped unsafe API operations, validates quality finding counts, validates baseline comparison and compact CI gate output, validates screenshot evidence download, verifies password redaction, verifies AI analysis with the fake provider, and executes deterministic safe browser plans.
- `qualora-ci-gate.sh`: small HTTP-based CI helper that evaluates a report quality gate and exits with the API-provided exit code.
- `demo-api/server.js`: deterministic local API and OpenAPI spec used by `make smoke`.
- `mock-api/server.js`: older deterministic local API retained for compatibility with earlier alpha API worker checks.
- `demo-web/server.js`: deterministic local frontend used by browser, login, authenticated smoke, application discovery, passive quality, authorization, Safe QA, and safe test plan smoke tests.
- `fake-llm/server.js`: deterministic OpenAI-compatible chat completions provider used by `make smoke`.

# Scripts

Developer automation will live here.

Scripts should be safe to run locally, documented, and avoid printing secrets.

Current scripts:

- `smoke.py`: end-to-end deterministic smoke driver. It also verifies Project Cockpit/Run Viewer UI copy, observable browser links, screenshot evidence metadata/downloads, and secret redaction.
- `qualora-ci-gate.sh`: small HTTP-based CI helper that evaluates a quality gate for an existing report and exits with the API-provided exit code. It can log in with `QUALORA_EMAIL` and `QUALORA_PASSWORD`, or use an existing session cookie/CSRF token.
- `qualora-ci-run.sh`: pipeline-friendly helper that logs in, starts a native CI run, and exits with the API-provided exit code. Set `QUALORA_RUN_SAFE_QA=false` to reuse the latest completed Safe QA report without AI. Issue export is off by default; if enabled, `QUALORA_ISSUE_EXPORT_DRY_RUN=true` is the safe default.
- `demo-api/server.js`: deterministic local API and OpenAPI spec used by `make smoke`.
- `mock-api/server.js`: older deterministic local API retained for compatibility with earlier alpha API worker checks.
- `demo-web/server.js`: deterministic local frontend used by browser, login, authenticated smoke, application discovery, passive quality, authorization, Safe QA, and safe test plan smoke tests.
- `fake-llm/server.js`: deterministic OpenAI-compatible chat completions provider used by `make smoke`.
- `demo-lab/web`: dedicated showcase web target with authentication, roles, safe/unsafe forms, discovery actions, and passive quality fixtures.
- `demo-lab/api`: dedicated showcase OpenAPI target with public/authenticated safe reads, contract mismatch, server-error, and skipped mutation fixtures.
- `run-demo-lab.sh`: starts the `demo-lab` profile and runs the comprehensive end-to-end showcase without printing configured secrets.
- `run-demo-lab-real-llm.sh`: optional real OpenAI-compatible Demo Lab workflow. It validates required variables before starting services, never prints the API key, and prints Project Cockpit, Run Viewer, test-plan, and report links.
- `test-real-llm-script.sh`: validates missing-variable behavior and confirms validation output cannot expose a supplied API-key marker.

Use `make showcase-smoke` for deterministic Demo Lab assertions, `make demo-lab` for the full fake-LLM showcase, or `make demo-lab-real-llm` for the optional provider-backed walkthrough. See `docs/real-llm-demo.md` before making external model calls.

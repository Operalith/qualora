# Scripts

Developer automation will live here.

Scripts should be safe to run locally, documented, and avoid printing secrets.

Current scripts:

- `smoke.py`: end-to-end browser, credential profile, login, authenticated smoke, role-aware authorization, OpenAPI import, safe API smoke, AI-analysis, AI test plan, and safe execution smoke test driver. It prints JSON and HTML report URLs, validates HTML report export, validates API result rows, validates skipped unsafe API operations, validates screenshot evidence download, verifies password redaction, verifies AI analysis with the fake provider, and executes a deterministic safe browser plan.
- `demo-api/server.js`: deterministic local API and OpenAPI spec used by `make smoke`.
- `mock-api/server.js`: older deterministic local API retained for compatibility with earlier alpha API worker checks.
- `demo-web/server.js`: deterministic local frontend used by browser, login, authenticated smoke, authorization, and safe test plan smoke tests.
- `fake-llm/server.js`: deterministic OpenAI-compatible chat completions provider used by `make smoke`.

# Scripts

Developer automation will live here.

Scripts should be safe to run locally, documented, and avoid printing secrets.

Current scripts:

- `smoke.py`: end-to-end browser, API, AI-analysis, AI test plan, and safe execution smoke test driver. It prints JSON and HTML report URLs, validates HTML report export, validates screenshot evidence download, verifies AI analysis with the fake provider, and executes a deterministic safe browser plan.
- `mock-api/server.js`: deterministic local API used by `make smoke`.
- `demo-web/server.js`: deterministic local frontend used by browser smoke tests.
- `fake-llm/server.js`: deterministic OpenAI-compatible chat completions provider used by `make smoke`.

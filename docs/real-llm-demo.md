# Real LLM Demo Mode

Qualora's real LLM Demo Lab mode is an optional public-alpha walkthrough for an OpenAI-compatible chat completion provider. It is separate from deterministic smoke and CI, which continue to use `fake-llm`.

Real provider calls may incur cost. Review the provider model, pricing, rate limits, and data handling before running the script.

## Required Configuration

Set all required values in the current shell:

```bash
export QUALORA_REAL_LLM_NAME="Demo Provider"
export QUALORA_REAL_LLM_BASE_URL="https://api.example.com/v1"
export QUALORA_REAL_LLM_API_KEY="..."
export QUALORA_REAL_LLM_MODEL="model-name"
```

Optional settings:

```bash
export QUALORA_REAL_LLM_EXTRA_HEADERS_JSON='{"X-Provider-Header":"value"}'
export QUALORA_API_URL="http://localhost:8080"
export QUALORA_API_BASE_URL="http://localhost:8080"
export QUALORA_API_PORT="8080"
```

`QUALORA_API_URL` takes precedence over `QUALORA_API_BASE_URL`, which takes precedence over `QUALORA_API_PORT`.

Then run:

```bash
make demo-lab-real-llm
```

The script starts Qualora and the `demo-lab` profile, completes local setup/login if needed, creates or updates the named provider, verifies connectivity, creates or reuses the Demo Lab project, and runs:

- deterministic application discovery
- policy-gated AI Browser Control
- discovery-aware AI test planning
- a Safe QA preview

It prints links for the Qualora UI, Project Cockpit, Run Viewer, AI Browser Control HTML report, test plan, and Safe QA report. It never prints the API key.

## AI Data Boundary

Qualora sends only bounded goals and sanitized report/discovery/browser observation metadata required by the selected AI workflow.

Qualora never sends credentials, API keys, cookies, local or session storage, authorization headers, tokens, screenshots, full HTML, request bodies, response bodies, raw traces, provider secrets, or encrypted secret payloads to AI.

Screenshots remain local Qualora evidence and may be viewed through the authenticated Run Viewer. The model proposes typed actions; Qualora policy validates each action before Playwright executes it. The model never receives direct Playwright control.

## Validation And Troubleshooting

The script fails before starting services when a required variable is missing:

```bash
scripts/run-demo-lab-real-llm.sh
```

Provider connectivity is tested before any AI workflow starts. Confirm that the provider base URL is reachable from the `qualora-api` and browser-worker containers and that it implements OpenAI-compatible chat completions.

Use `make demo-lab` for repeatable local demos without external calls or model cost.

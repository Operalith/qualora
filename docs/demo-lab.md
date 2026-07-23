# Qualora Demo Lab

Qualora Demo Lab is a deterministic, local-only target application for public alpha walkthroughs, automated smoke checks, CI validation, screenshots, and videos. It is not a production application and does not replace testing a real target.

In v0.24, the dashboard starts the showcase directly, the Project Cockpit keeps the main actions visible, and Run Viewer shows AI Browser Control or Safe Explorer steps with local screenshot evidence, AI suggestions, policy decisions, execution results, and blocked reasons.

## Start And Validate

From the repository root, run:

```bash
scripts/run-demo-lab.sh
```

This starts the core Qualora stack, `demo-lab-web`, `demo-lab-api`, and `fake-llm`; performs local admin setup or login; runs the full showcase smoke; checks report redaction; and prints project/report links.

When Qualora is already running:

```bash
make showcase-smoke
```

The default path uses Fake LLM and remains deterministic. To use an optional real OpenAI-compatible provider, follow [real-llm-demo.md](real-llm-demo.md) and run:

```bash
make demo-lab-real-llm
```

Real provider calls may incur cost and are never required by smoke or CI.

Manual target URLs:

- Qualora UI: `http://localhost:3000`
- Demo Lab web: `http://localhost:18085`
- Demo Lab API: `http://localhost:18086`
- Demo Lab OpenAPI: `http://localhost:18086/openapi.yaml`

The container-network project values are `http://demo-lab-web:8080`, `http://demo-lab-api:8080`, and allowed hosts `demo-lab-web,demo-lab-api`.

## Demo Accounts

These accounts are fake local fixtures. Never reuse the credentials elsewhere.

| Role | Email | Password |
| --- | --- | --- |
| Admin | `admin@example.com` | `admin-password` |
| Readonly | `readonly@example.com` | `readonly-password` |
| Customer A | `customer-a@example.com` | `customer-a-password` |
| Customer B | `customer-b@example.com` | `customer-b-password` |

Stable login configuration:

- Login URL: `http://demo-lab-web:8080/login`
- Username selector: `input[name="email"]`
- Password selector: `input[name="password"]`
- Submit selector: `button[type="submit"]`
- Success URL contains: `/dashboard`
- Success text contains: `Welcome to Demo Lab`

The API bearer token is `demo-api-token`. Role-shaped variants are also accepted: `demo-api-token-admin`, `demo-api-token-readonly`, `demo-api-token-customer-a`, and `demo-api-token-customer-b`. All are local fixtures and must not be reused.

## What It Exercises

| Qualora workflow | Demo Lab fixture |
| --- | --- |
| Browser smoke and discovery | Stable Home, About, Status, Pricing, Search, Products, and quality routes |
| Authenticated browser smoke | Selector-based login and protected `/dashboard` |
| Authorization checks | `/admin`, `/reports`, and customer A/B invoice routes |
| Safe Explorer | Same-origin safe navigation plus unsafe, external, logout, and unsupported actions |
| Safe Form Testing | GET search/filter/sort forms plus skipped POST, external, password/token, upload, transfer, and delete forms |
| Passive quality checks | Missing CSP/header fixtures, `X-Powered-By`, console error, broken asset, accessibility gaps, source map, and bounded slow response |
| OpenAPI import and API smoke | Public health/catalog/status, path parameters, required-query skip, and skipped mutation operations |
| Authenticated API contracts | Profile/orders, intentional missing required fields, and deterministic 500 response |
| AI planning/control | Deterministic fake LLM plans, safe navigation/forms, screenshot/stop, and blocked destructive suggestion |
| Run Viewer | Near-live/replay step timeline, latest screenshot, AI suggestion, policy decision, execution result, and report links |
| Safe QA, baselines, and CI gates | Stable repeated run for unchanged comparison and pass gate |
| Issue export | Sanitized grouped-finding dry-run only |

Expected findings are intentional. Baseline mode includes passive missing-CSP and disclosure signals, accessibility metadata gaps, a broken same-origin asset, one console error, a bounded slow route, `/broken` and authenticated server errors, and one authenticated response schema mismatch. Qualora should keep raw findings visible while grouping repeated signals through report intelligence.

## Regression Mode

The default is `DEMO_LAB_MODE=baseline`. To enable deterministic additional failures:

```bash
DEMO_LAB_MODE=regressed docker compose --profile demo-lab up -d --build --force-recreate demo-lab-web demo-lab-api
```

Regressed mode adds a broken internal link, a new web 500 route, an extra console error, broader missing-header behavior, and a public API regression route. A practical walkthrough is:

1. Run Safe QA in baseline mode and set its report as the default baseline.
2. Recreate Demo Lab in regressed mode with the command above.
3. Run Safe QA again, compare it with the baseline, and evaluate the quality gate.
4. Recreate Demo Lab without `DEMO_LAB_MODE` to return to baseline mode.

Regression mode is deterministic and local-only. Exact comparison output depends on which project workflows and discovered pages are included in the Safe QA run.

## Screenshots And Videos

Start the profile, open `http://localhost:18085` for target footage, and use `http://localhost:3000` for Qualora workflow/report footage. Keep the services in baseline mode unless the walkthrough specifically demonstrates regression comparison. All fixture labels and routes are stable for repeatable capture.

## Safety

- No real users, payments, transfers, or external services exist.
- POST/mutating fixture handlers return `405` and persist nothing.
- Qualora must continue to skip mutating, destructive, logout, reset, upload, sensitive, and external actions.
- Demo credentials/tokens are fake but still checked for report, AI input, CI output, issue preview, and log leakage.
- Real LLM mode sends sanitized metadata and bounded goals only. It never sends credentials, cookies, tokens, auth headers, browser storage, screenshots, full HTML, request bodies, or response bodies.
- Demo Lab does not enable active scanning, payload attacks, fuzzing, or destructive testing.
- Do not expose Demo Lab or the default Qualora Compose credentials to untrusted networks.

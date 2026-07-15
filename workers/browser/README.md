# Browser Worker

Playwright worker for browser smoke tests, deterministic credential-profile login checks, authenticated browser smoke tests, explicit role-aware authorization checks, and approved safe test plan execution.

Responsibilities:

- Visit the configured frontend URL.
- Capture screenshots.
- Capture page title, final URL, initial status code, and body text length.
- Collect console errors and failed network requests.
- Block out-of-scope requests outside project allowed hosts.
- Block unsafe private, loopback, link-local, and metadata targets by default.
- Decrypt project-scoped credential profiles with `QUALORA_ENCRYPTION_KEY`.
- Fill only configured username/password selectors and click only the configured submit selector for login checks.
- Run authenticated browser smoke against one configured same-origin target path.
- Run role-aware authorization checks against one explicitly configured same-origin browser URL target.
- Classify authorization outcomes as allowed, denied, or unknown using status codes and configured success/denied text.
- Write findings and evidence metadata to PostgreSQL.
- Store screenshots in MinIO/S3 with a local filesystem fallback.
- Record screenshot metadata including filename, storage key, content type, size, storage backend, and created timestamp.
- Consume safe test plan execution jobs from `TEST_PLAN_EXECUTION_QUEUE`.
- Execute only persisted supported safe DSL actions after control-plane mapping.
- Persist execution step outcomes, screenshot evidence, browser observations, and findings.

All navigation and network activity must respect project allowed hosts.

The worker must never log or persist raw usernames, passwords, cookies, local storage, session storage, authorization headers, tokens, or browser storage contents.

Authorization checks must not crawl, submit arbitrary forms, run payloads, mutate state, or use AI browser control.

## Local Development

```bash
npm install
npm run build
npm run dev
```

The worker consumes Redis jobs from `RUN_QUEUE` and `TEST_PLAN_EXECUTION_QUEUE` and writes output directly to PostgreSQL for this MVP. Authorization check run jobs are carried on the browser run queue in v0.11.

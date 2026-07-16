# Browser Worker

Playwright worker for browser smoke tests, deterministic credential-profile login checks, authenticated browser smoke tests, safe application discovery, passive quality checks, explicit role-aware authorization checks, and approved safe test plan execution.

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
- Run bounded safe application discovery with same-origin defaults, allowed-host enforcement, screenshots, links, forms, fields, skip reasons, and deterministic findings.
- Run passive quality checks for security header/cookie/form metadata, basic accessibility heuristics, and performance/front-end resource observations.
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

Discovery must never submit forms, click arbitrary buttons, execute payloads, perform destructive actions, crawl external domains by default, or use autonomous AI browser control.

Quality checks must never submit forms, click arbitrary buttons, guess sensitive paths, execute payloads, fuzz inputs, perform active scanning, perform destructive actions, store cookie values/browser storage/full HTML/request bodies/response bodies, or use autonomous AI browser control.

Authorization checks must not crawl, submit arbitrary forms, run payloads, mutate state, or use AI browser control.

## Local Development

```bash
npm install
npm run build
npm run dev
```

The worker consumes Redis jobs from `RUN_QUEUE` and `TEST_PLAN_EXECUTION_QUEUE` and writes output directly to PostgreSQL for this MVP. Authorization check run jobs, discovery run jobs, and quality check run jobs are carried on the browser run queue.

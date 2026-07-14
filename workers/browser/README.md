# Browser Worker

Playwright worker for browser smoke tests and approved safe test plan execution.

Responsibilities:

- Visit the configured frontend URL.
- Capture screenshots.
- Capture page title, final URL, initial status code, and body text length.
- Collect console errors and failed network requests.
- Block out-of-scope requests outside project allowed hosts.
- Block unsafe private, loopback, link-local, and metadata targets by default.
- Write findings and evidence metadata to PostgreSQL.
- Store screenshots in MinIO/S3 with a local filesystem fallback.
- Record screenshot metadata including filename, storage key, content type, size, storage backend, and created timestamp.
- Consume safe test plan execution jobs from `TEST_PLAN_EXECUTION_QUEUE`.
- Execute only persisted supported safe DSL actions after control-plane mapping.
- Persist execution step outcomes, screenshot evidence, browser observations, and findings.

All navigation and network activity must respect project allowed hosts.

## Local Development

```bash
npm install
npm run build
npm run dev
```

The worker consumes Redis jobs from `RUN_QUEUE` and `TEST_PLAN_EXECUTION_QUEUE` and writes output directly to PostgreSQL for this MVP.

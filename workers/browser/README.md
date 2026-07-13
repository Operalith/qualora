# Browser Worker

Playwright worker for browser smoke tests.

Responsibilities:

- Visit the configured frontend URL.
- Capture screenshots.
- Collect console errors and failed network requests.
- Block out-of-scope requests outside project allowed hosts.
- Block unsafe private, loopback, link-local, and metadata targets by default.
- Write findings and evidence metadata to PostgreSQL.
- Store screenshots in MinIO/S3 with a local filesystem fallback.

All navigation and network activity must respect project allowed hosts.

## Local Development

```bash
npm install
npm run build
npm run dev
```

The worker consumes Redis jobs from `RUN_QUEUE` and writes run output directly to PostgreSQL for this MVP.

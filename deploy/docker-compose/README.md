# Docker Compose

The first supported self-hosted deployment path is the root-level `docker-compose.yml`.

From the repository root:

```bash
docker compose up -d --build
docker compose logs -f qualora-api qualora-web qualora-worker-browser qualora-worker-api
```

If local port `8080` is already in use:

```bash
QUALORA_API_PORT=18080 docker compose up -d --build
QUALORA_API_URL=http://localhost:18080 QUALORA_API_BASE_URL=http://localhost:18080 make smoke
```

The web UI is exposed at `http://localhost:3000` by default. Override it with `QUALORA_WEB_PORT`.

On a fresh database, open the web UI and complete first-run local admin setup before accessing projects and reports. After login, use `#/setup-project` for guided project setup or the dashboard `Run demo workflow` action for the local demo path. Project pages can start browser smoke, authenticated smoke, discovery, Interactive Safe Explorer, passive quality, Safe QA, native CI runs, and API smoke workflows when the required project settings are present. Safe QA report pages can set baselines, compare against baselines, evaluate quality gates, and dry-run sanitized issue export. The smoke script can also create the local admin automatically and exercise the guided demo flow for demo stacks.

The MVP Compose stack includes:

- `qualora-api`: Go control plane API.
- `qualora-web`: minimal React web UI.
- `qualora-worker-browser`: TypeScript/Playwright browser worker.
- `qualora-worker-api`: TypeScript API/OpenAPI worker.
- `postgres`: durable metadata.
- `redis`: browser and API run queues.
- `minio`: S3-compatible evidence storage.

The smoke profile also includes:

- `mock-api`: older deterministic local API retained for compatibility with earlier alpha API worker checks.
- `demo-api`: deterministic OpenAPI demo API used by safe API smoke tests.
- `demo-web`: deterministic local frontend used by browser, login, authenticated smoke, application discovery, Interactive Safe Explorer, passive quality, role-aware authorization, and safe test plan smoke tests.
- `fake-llm`: deterministic OpenAI-compatible provider used by AI and guided onboarding smoke tests.

The control plane receives the same MinIO/S3 configuration as the browser worker so authenticated `GET /api/v1/evidence/{evidence_id}` requests can stream screenshot evidence without exposing MinIO credentials to the web UI.

Set `QUALORA_ENCRYPTION_KEY` before storing real credential profiles or AI provider credentials. The default Compose value is intentionally insecure and only suitable for local demos.

Report intelligence in `v0.19.0-alpha` is computed inside `qualora-api` when reports are read. It adds executive summaries, severity counts, grouped findings, top findings, affected pages, noise summaries, and deduplication metadata to JSON/HTML reports without calling an AI provider.

Baselines, quality gates, and native CI runs in `v0.19.0-alpha` are also handled by `qualora-api`. Baselines persist grouped finding fingerprints and summary metadata in PostgreSQL; comparisons and gates are computed synchronously without starting workers or requiring AI. `scripts/qualora-ci-gate.sh` evaluates an existing report, while `scripts/qualora-ci-run.sh` starts or reuses a Safe QA workflow and returns a deterministic exit code. Both scripts can log in with `QUALORA_EMAIL` and `QUALORA_PASSWORD`.

Issue export in `v0.19.0-alpha` is optional. GitHub/GitLab tokens are encrypted with `QUALORA_ENCRYPTION_KEY`, responses expose only `token_configured`, and `POST /api/v1/reports/{report_type}/{report_id}/export-issues` defaults to dry-run previews from grouped sanitized findings.

Interactive Safe Explorer remains policy-gated in `v0.19.0-alpha`: it executes only safe classified same-origin navigation actions by default, records skipped unsafe/unsupported actions with reasons, and does not use AI to control the browser. It does not submit POST forms, click arbitrary buttons, run payloads, fuzz inputs, perform active scans, or perform destructive actions.

Auth-related local defaults:

- `QUALORA_SESSION_TTL_HOURS=12`
- `QUALORA_COOKIE_SECURE=false` for local HTTP. Set it to `true` behind HTTPS.
- `QUALORA_AUTH_DISABLED=false`

Keep this path working before adding Helm or other deployment targets.

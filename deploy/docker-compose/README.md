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

On a fresh database, open the web UI and complete first-run local admin setup before accessing projects and reports. The smoke script can also create the local admin automatically for demo stacks.

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
- `demo-web`: deterministic local frontend used by browser, login, authenticated smoke, application discovery, and role-aware authorization smoke tests.
- `fake-llm`: deterministic OpenAI-compatible provider used by AI smoke tests.

The control plane receives the same MinIO/S3 configuration as the browser worker so authenticated `GET /api/v1/evidence/{evidence_id}` requests can stream screenshot evidence without exposing MinIO credentials to the web UI.

Set `QUALORA_ENCRYPTION_KEY` before storing real AI provider credentials. The default Compose value is intentionally insecure and only suitable for local demos.

Auth-related local defaults:

- `QUALORA_SESSION_TTL_HOURS=12`
- `QUALORA_COOKIE_SECURE=false` for local HTTP. Set it to `true` behind HTTPS.
- `QUALORA_AUTH_DISABLED=false`

Keep this path working before adding Helm or other deployment targets.

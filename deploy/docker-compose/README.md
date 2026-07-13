# Docker Compose

The first supported self-hosted deployment path is the root-level `docker-compose.yml`.

From the repository root:

```bash
docker compose up -d --build
docker compose logs -f qualora-api qualora-worker-browser
```

The MVP Compose stack includes:

- `qualora-api`: Go control plane API.
- `qualora-worker-browser`: TypeScript/Playwright browser worker.
- `postgres`: durable metadata.
- `redis`: browser run queue.
- `minio`: S3-compatible evidence storage.

Keep this path working before adding Helm or other deployment targets.

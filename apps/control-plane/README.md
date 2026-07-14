# Control Plane

Go service for Qualora's API and orchestration layer.

Responsibilities:

- Project configuration.
- Test run lifecycle.
- Policy validation.
- Worker job scheduling.
- Metadata persistence.
- JSON and HTML report access.

The MVP delegates browser checks and API checks to workers through Redis queues.

The API performs project target validation, including `allowed_hosts` enforcement, blocked private/metadata targets by default, and DNS resolution checks for hostnames.

## Local Development

```bash
go test ./...
go run .
```

The service expects PostgreSQL and Redis. Defaults are suitable for local development:

- `DATABASE_URL=postgres://qualora:qualora@localhost:5432/qualora?sslmode=disable`
- `REDIS_ADDR=localhost:6379`
- `BROWSER_RUN_QUEUE=qualora:browser-runs`
- `API_RUN_QUEUE=qualora:api-runs`
- `CORS_ALLOWED_ORIGINS=http://localhost:3000`

Current report endpoints:

- `GET /api/v1/runs/{run_id}/report`
- `GET /api/v1/runs/{run_id}/report.html`

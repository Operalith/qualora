# Control Plane

Go service for Qualora's API and orchestration layer.

Responsibilities:

- Project configuration.
- Test run lifecycle.
- Policy validation.
- Worker job scheduling.
- Metadata persistence.
- Report and finding access.

The MVP delegates browser checks to the browser worker through a Redis queue.

The API performs project target validation, including `allowed_hosts` enforcement, blocked private/metadata targets by default, and DNS resolution checks for hostnames.

## Local Development

```bash
go test ./...
go run .
```

The service expects PostgreSQL and Redis. Defaults are suitable for local development:

- `DATABASE_URL=postgres://qualora:qualora@localhost:5432/qualora?sslmode=disable`
- `REDIS_ADDR=localhost:6379`
- `RUN_QUEUE=qualora:browser-runs`

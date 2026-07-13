# Development

This document covers local development for Qualora v0.1.0-alpha.

## Requirements

- Docker with Docker Compose.
- Go 1.22 or newer for control plane development.
- Node.js 22 or newer for browser worker development.
- Python 3 for the smoke script.

## Common Commands

```bash
make dev
make test
make lint
make compose-up
make compose-down
make logs
make smoke
```

Command behavior:

- `make dev`: starts the Docker Compose stack.
- `make test`: runs Go tests and browser worker TypeScript checks.
- `make lint`: runs the same checks plus `docker compose config`.
- `make compose-up`: runs `docker compose up -d --build`.
- `make compose-down`: runs `docker compose down`.
- `make logs`: tails API and browser worker logs.
- `make smoke`: creates an example project, starts a run, polls to completion, and prints the report.

## Start The Stack

```bash
docker compose up -d --build
```

If port `8080` is already in use:

```bash
QUALORA_API_PORT=18080 docker compose up -d --build
QUALORA_API_URL=http://localhost:18080 make smoke
```

## Run Tests

```bash
make test
make lint
```

Backend only:

```bash
cd apps/control-plane
go test ./...
go run .
```

Browser worker only:

```bash
cd workers/browser
npm ci
npm run build
npm test
```

## Smoke Test

With the Compose stack running:

```bash
make smoke
```

The smoke script targets `https://example.com` by default. Override it with:

```bash
QUALORA_TARGET_URL=https://example.com \
QUALORA_ALLOWED_HOST=example.com \
make smoke
```

For private or local targets, create projects manually with `allow_private_targets: true` only when testing systems you control.

## Clean Up

Stop containers:

```bash
docker compose down
```

Stop containers and delete local volumes:

```bash
docker compose down -v
```

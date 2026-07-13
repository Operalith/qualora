# Qualora MVP Architecture

This document defines the practical first-release shape of Qualora. It should be updated whenever the architecture changes.

## Goals

- Run locally with Docker Compose.
- Let users define a project with frontend URL, API base URL, optional OpenAPI URL, test credentials, allowed hosts, and a testing policy.
- Start a test run and collect browser, API, contract, and passive security evidence.
- Produce structured findings and reports.
- Keep the codebase modular enough for future workers without overbuilding the first implementation.

## Non-Goals

- Hosted SaaS assumptions.
- Multi-tenant billing, organizations, or enterprise RBAC.
- Kubernetes-first design.
- Aggressive or destructive security scanning.
- OWASP ZAP active scanning in the first release.

## Components

### Control Plane

The control plane is a Go service responsible for:

- Project CRUD.
- Test run lifecycle.
- Policy validation.
- Queueing worker jobs.
- Persisting run metadata, evidence metadata, findings, and report metadata.
- Exposing an OpenAPI-documented HTTP API.

The control plane should not directly perform heavy browser automation or long-running checks once workers exist.

### PostgreSQL

PostgreSQL stores durable data:

- Projects.
- Encrypted or sealed credential references.
- Allowed host policies.
- Test runs and job states.
- Findings.
- Report metadata.
- Evidence artifact metadata.

### Redis

Redis is used for:

- MVP job queueing.
- Short-lived run state.
- Worker coordination where needed.

Do not rely on Redis as the only source of durable run history.

### MinIO / S3-Compatible Storage

Evidence artifacts should be stored in S3-compatible storage:

- Screenshots.
- Playwright traces.
- Sanitized logs.
- Generated reports.

The control plane stores metadata and object references in PostgreSQL.

### Browser Worker

The browser worker runs Playwright smoke checks:

- Visit the configured frontend URL.
- Optionally authenticate with project test credentials.
- Capture screenshots.
- Capture traces when enabled.
- Record console errors.
- Record failed network requests.

All navigation and requests must be constrained by allowed hosts.

### API Worker

The API worker runs HTTP checks:

- Basic health checks against the API base URL.
- Configured endpoint checks.
- Optional OpenAPI contract checks when an OpenAPI URL is supplied.
- API error and response metadata collection.

Credentials and authorization headers must be redacted from logs and reports.

### Security Worker

The MVP security worker is passive and safe:

- Check security headers.
- Check cookie flags.
- Check mixed content signals.
- Check obvious TLS/configuration metadata where available.

It must not perform exploitation, brute force testing, destructive payloads, or broad scanning.

### Analyzer Worker

The analyzer worker converts raw evidence into structured findings:

- Normalize browser, API, contract, and passive security signals.
- Assign severity.
- Generate reproduction steps.
- Attach evidence references.
- Suggest recommendations.

### Report Engine

The report engine produces structured reports from findings and run metadata.

Expected report fields:

- Summary.
- Scope.
- Run metadata.
- Findings.
- Severity.
- Reproduction steps.
- Evidence references.
- Recommendations.

## Run Lifecycle

1. User creates or updates a project.
2. Control plane validates URLs, allowed hosts, credentials metadata, and policy.
3. User starts a test run.
4. Control plane creates a durable run record in PostgreSQL.
5. Control plane queues browser, API, and security jobs in Redis.
6. Workers execute jobs and write artifacts to MinIO/S3.
7. Workers write evidence metadata and raw signals through the control plane or a narrow storage interface.
8. Analyzer produces structured findings.
9. Report engine generates the run report.
10. Control plane marks the run complete or failed.

## Safety Model

Allowed hosts are mandatory. Every outbound browser navigation, API request, contract fetch, and passive security check must verify the destination before running.

Testing policy should eventually control:

- Maximum runtime.
- Maximum pages or endpoints.
- Whether authentication is allowed.
- Whether traces are captured.
- Whether passive security checks are enabled.
- Whether future active checks are permitted.

MVP default policy should be conservative.

## Credential Handling

For the local MVP, credentials may be persisted in PostgreSQL only through a dedicated secret abstraction that supports future replacement.

Requirements:

- Never log raw credential values.
- Redact known secret fields from errors and reports.
- Avoid returning secrets from API read endpoints.
- Keep the storage interface replaceable for Vault, Kubernetes Secrets, or cloud secret managers later.

## First Data Model Sketch

Initial entities:

- `projects`
- `project_credentials`
- `project_allowed_hosts`
- `test_runs`
- `run_jobs`
- `evidence_artifacts`
- `findings`
- `reports`

This is a sketch, not a migration contract.

## Deployment Path

First supported target:

- Docker Compose with control plane, PostgreSQL, Redis, MinIO, and workers.

Later target:

- Helm chart for Kubernetes after the Compose workflow is stable.

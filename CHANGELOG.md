# Changelog

All notable changes to Qualora will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses semantic versioning once stable releases begin.

## [v0.1.0-alpha] - 2026-07-13

### Added

- Go control plane API with health, project, run, and report endpoints.
- PostgreSQL migrations for projects, test runs, findings, and evidence.
- Redis-backed browser run queue.
- TypeScript/Node.js Playwright browser worker.
- Browser smoke checks for page title, screenshot capture, console errors, failed requests, and blocked out-of-scope requests.
- MinIO/S3 screenshot evidence storage with local filesystem fallback.
- Structured JSON report output.
- Docker Compose stack for local self-hosted use.
- Makefile with development, test, lint, Compose, logs, and smoke commands.
- Smoke script for `https://example.com`.
- OpenAPI contract for the current API.
- GitHub Actions CI workflow.
- Release, architecture, development, roadmap, and security model docs.

### Security

- Allowed-host validation for project targets.
- Default blocking for localhost, loopback, link-local, private IP literal targets, cloud metadata addresses, and public hostnames resolving to blocked IP ranges.
- Browser worker request blocking outside project `allowed_hosts`.
- Secret-like value redaction in worker logs.

### Known Limitations

- No web UI.
- No API worker or OpenAPI contract checks.
- No authentication.
- No login automation or credential storage.
- No active security scanning.
- No Helm/Kubernetes deployment.

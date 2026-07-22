# Qualora Demo Lab fixtures

This directory contains the deterministic, local-only target services used by the Qualora v0.23.0-alpha showcase:

- `web`: public pages, role-aware login/session routes, safe and unsafe forms, discovery actions, and passive quality fixtures.
- `api`: public and bearer-authenticated OpenAPI endpoints, skipped mutating operations, an intentional contract mismatch, and deterministic server errors.

Both services use only Node.js built-ins. They have no external dependencies, persist no user data, and disable all mutation fixtures. Use `scripts/run-demo-lab.sh` from the repository root for the full workflow, or `docker compose --profile demo-lab up -d --build` to inspect the targets manually.

Set `DEMO_LAB_MODE=regressed` before recreating the services to enable the local regression fixtures. See `docs/demo-lab.md` for users, URLs, expected findings, and safety boundaries.

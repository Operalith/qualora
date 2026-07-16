# Qualora Web

Minimal React/Vite web UI for Qualora v0.13.0-alpha.

The UI supports:

- Listing projects.
- First-run local admin setup.
- Local admin login, session refresh, and logout.
- Creating projects.
- Viewing project details and project runs.
- Importing OpenAPI specs for projects.
- Viewing API spec details, discovered operations, and skip reasons.
- Starting safe API smoke runs from imported specs.
- Managing credential profiles.
- Assigning role metadata to credential profiles.
- Testing deterministic login flows.
- Starting authenticated browser smoke runs.
- Creating explicit role-aware authorization checks.
- Starting authorization check runs.
- Viewing authorization check reports, results, findings, and evidence.
- Starting safe application discovery runs.
- Viewing discovery run reports, application maps, pages, links, forms, findings, and evidence.
- Generating discovery-aware AI test plans from sanitized application maps.
- Starting Safe QA Run previews, executing reviewed safe QA runs, and viewing Safe QA Run reports.
- Starting runs.
- Listing runs.
- Viewing run reports, findings, evidence metadata, browser metadata, API metadata, and worker job metadata.
- Opening the self-contained HTML report served by the control plane.
- Previewing and downloading screenshot evidence through the control-plane evidence endpoint.
- Configuring optional OpenAI-compatible AI providers.
- Testing AI provider connectivity.
- Running and viewing AI analysis for completed runs.
- Generating, listing, viewing, deleting, and exporting AI-assisted test plans.
- Previewing, starting, listing, and viewing approved safe test plan executions.

## Local Development

```bash
npm ci
npm run dev
```

The app defaults to `http://localhost:8080` for the API.

Override the API base URL during Vite development:

```bash
VITE_QUALORA_API_BASE_URL=http://localhost:18080 npm run dev
```

## Docker Runtime Config

The Docker image serves `/config.js` from `QUALORA_API_BASE_URL` at runtime:

```bash
QUALORA_API_BASE_URL=http://localhost:8080 docker compose up -d --build qualora-web
```

## Current Limitations

- Local authentication is intentionally minimal: one admin role, no password reset, no user management UI, no SSO/OIDC/SAML, and no multi-tenancy.
- Intended for trusted local/self-hosted alpha environments only.
- AI provider credentials should only be configured in trusted environments.
- AI-assisted test plans are suggestions and are never executed automatically.
- Safe test plan execution is limited to the supported non-destructive browser DSL.
- Safe QA Runs preview by default and execute only persisted safe DSL steps after explicit user action.
- Safe API smoke execution is read-only and skips mutating, authenticated, ambiguous, request-body, unresolved-parameter, and sensitive operations.
- Credential profiles are intended for trusted local/self-hosted test accounts and never return raw credentials.
- Authenticated browser smoke is limited to one configured login form and one same-origin target path.
- Authorization checks are explicit, read-only, and limited to deterministic browser URL checks.
- Application discovery is bounded, same-origin by default, and does not submit forms or click mutating controls.
- Evidence preview/download is limited to evidence records known to Qualora.
- No advanced filtering, pagination, or project editing yet.

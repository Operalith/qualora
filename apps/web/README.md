# Qualora Web

Minimal React/Vite web UI for Qualora v0.6.0-alpha.

The UI supports:

- Listing projects.
- Creating projects.
- Viewing project details and project runs.
- Starting runs.
- Listing runs.
- Viewing run reports, findings, evidence metadata, browser metadata, API metadata, and worker job metadata.
- Opening the self-contained HTML report served by the control plane.
- Previewing and downloading screenshot evidence through the control-plane evidence endpoint.
- Configuring optional OpenAI-compatible AI providers.
- Testing AI provider connectivity.
- Running and viewing AI analysis for completed runs.
- Generating, listing, viewing, deleting, and exporting AI-assisted test plans.

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

- No authentication or authorization.
- Intended for trusted local/self-hosted alpha environments only.
- AI provider credentials should only be configured in trusted environments.
- AI-assisted test plans are suggestions only and are not executed by Qualora.
- Evidence preview/download is limited to evidence records known to Qualora.
- No advanced filtering, pagination, or project editing yet.

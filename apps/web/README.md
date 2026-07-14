# Qualora Web

Minimal React/Vite web UI for Qualora v0.3.0-alpha.

The UI supports:

- Listing projects.
- Creating projects.
- Viewing project details and project runs.
- Starting runs.
- Listing runs.
- Viewing run reports, findings, evidence metadata, browser metadata, API metadata, and worker job metadata.
- Opening the self-contained HTML report served by the control plane.

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
- Evidence preview/download is not implemented yet; the UI displays metadata and URIs.
- No advanced filtering, pagination, or project editing yet.

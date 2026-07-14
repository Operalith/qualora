# Security Model

Qualora is security-adjacent automation. The v0.3.0-alpha safety model is intentionally conservative.

## Scope Rule

Only run Qualora against systems you own or are explicitly authorized to test.

Every project must define `allowed_hosts`. The control plane validates project URLs against that list, and workers block requests that fall outside the list.

## Default Target Blocking

Unless `allow_private_targets` is set to `true`, Qualora blocks:

- `localhost` and `.localhost`.
- `.local` hostnames.
- Loopback IPs.
- Link-local IPs.
- Private IP literal targets.
- Multicast and unspecified IPs.
- Common cloud metadata IPs and hostnames, including `169.254.169.254`, `100.100.100.200`, `metadata`, `metadata.google.internal`, `metadata.goog`, and `instance-data`.
- Public hostnames that resolve to blocked private, loopback, link-local, multicast, unspecified, or metadata IP addresses.

`allow_private_targets: true` is available for local and private test environments, but it should only be used for systems you control.

## Browser Request Enforcement

The browser worker routes Playwright requests through the host policy:

- Requests outside `allowed_hosts` are aborted.
- Blocked requests are recorded as browser observation evidence.
- Blocked requests can produce an informational finding.

## API Request Enforcement

The API worker validates `api_base_url`, `openapi_url`, and every OpenAPI endpoint URL against the same host policy.

Default API behavior:

- Safe baseline `GET` against `api_base_url`.
- OpenAPI document fetch from `openapi_url`.
- Safe OpenAPI methods only: `GET`, `HEAD`, and `OPTIONS`.
- Unsafe methods such as `POST`, `PUT`, `PATCH`, and `DELETE` are skipped.
- `destructive_actions=true` is not supported by the v0.3.0-alpha API worker.

## Web UI Exposure

The v0.3.0-alpha web UI has no authentication or authorization. It can create projects, start runs, and display report/evidence metadata through the control-plane API.

Use it only in trusted local or self-hosted environments. Do not expose `qualora-web` or `qualora-api` directly to untrusted networks without adding an external access-control layer.

## Secret Handling

The alpha does not implement login automation, authenticated API testing, or credential storage.

Current safeguards:

- API request logs do not include request bodies or query strings.
- Worker logs redact common token, password, secret, cookie, and authorization patterns.
- API evidence strips URL userinfo, query strings, and fragments.
- Screenshot and report artifacts should be treated as sensitive.

Future credential support must use a dedicated abstraction that can later support Vault, Kubernetes Secrets, or another secret manager.

## Non-Goals For This Alpha

- Exploit execution.
- Brute force testing.
- Destructive payloads.
- Broad crawling.
- Authenticated API testing.
- Schema fuzzing.
- OWASP ZAP integration.
- Active security scanning.

## Known Security Limitations

- DNS resolution checks are performed at validation/runtime, but DNS can change between checks.
- Browser screenshots can contain sensitive application data.
- API response metadata can reveal endpoint names and status behavior.
- MinIO uses local development credentials in Docker Compose.
- There is no API or web UI authentication in this alpha, so bind the API and UI only in trusted local environments.
- Screenshot preview/download through the control-plane API is not implemented yet; evidence metadata can still disclose target names, endpoint paths, and artifact URIs.

See [../SECURITY.md](../SECURITY.md) for vulnerability reporting.

# Security Model

Qualora is security-adjacent automation. The v0.5.0-alpha safety model is intentionally conservative.

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
- `destructive_actions=true` is not supported by the v0.5.0-alpha API worker.

## Web UI Exposure

The v0.5.0-alpha web UI has no authentication or authorization. It can create projects, start runs, configure AI providers, run AI analysis, and display report/evidence metadata through the control-plane API.

Use it only in trusted local or self-hosted environments. Do not expose `qualora-web` or `qualora-api` directly to untrusted networks without adding an external access-control layer.

## Secret Handling

The alpha does not implement login automation or authenticated API testing.

Current safeguards:

- API request logs do not include request bodies or query strings.
- Worker logs redact common token, password, secret, cookie, and authorization patterns.
- API evidence strips URL userinfo, query strings, and fragments.
- Evidence object downloads are served only for evidence records already known to Qualora; callers cannot provide arbitrary S3 keys or filesystem paths.
- AI provider API keys are encrypted at rest.
- AI provider extra headers are treated as sensitive and encrypted at rest.
- AI provider responses never include raw API keys or raw extra headers.
- Screenshot and report artifacts should be treated as sensitive.

The Docker Compose default `QUALORA_ENCRYPTION_KEY` is an insecure development fallback. Set a strong value before storing real provider credentials. Future credential support should keep the current abstraction and add Vault, Kubernetes Secrets, or another secret manager.

## AI Analysis Safety

AI is disabled until a provider is configured. Qualora works without AI.

The AI input builder sends sanitized structured report data only. By default it may include run status, summary counts, finding titles/categories/severities/summaries, safe evidence metadata, browser/API metadata, and job metadata.

The AI input builder does not send by default:

- Cookies.
- Authorization headers.
- Passwords.
- Tokens.
- API keys.
- Full request bodies.
- Full response bodies.
- Full HTML.
- Screenshots.
- Raw traces.
- Secret-looking query parameters.
- Sensitive headers.

Redaction is enabled by default and masks common bearer/basic auth values, API keys, passwords, access/refresh tokens, session IDs, cookies, and JWT-looking values. AI output is parsed as strict JSON and redacted before storage.

## Non-Goals For This Alpha

- Exploit execution.
- Brute force testing.
- Destructive payloads.
- Broad crawling.
- Authenticated API testing.
- Schema fuzzing.
- Autonomous AI browser control.
- OWASP ZAP integration.
- Active security scanning.

## Known Security Limitations

- DNS resolution checks are performed at validation/runtime, but DNS can change between checks.
- Browser screenshots can contain sensitive application data.
- API response metadata can reveal endpoint names and status behavior.
- MinIO uses local development credentials in Docker Compose.
- There is no API or web UI authentication in this alpha, so bind the API and UI only in trusted local environments.
- Screenshot preview/download through the control-plane API is available for stored evidence records and can expose sensitive application state to anyone with API access.
- Anyone with API/UI access can configure or use AI providers because this alpha has no authentication.
- AI analysis quality depends on the configured provider and the sanitized evidence available in the report.

See [../SECURITY.md](../SECURITY.md) for vulnerability reporting.

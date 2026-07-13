# Security Model

Qualora is security-adjacent automation. The v0.1.0-alpha safety model is intentionally conservative.

## Scope Rule

Only run Qualora against systems you own or are explicitly authorized to test.

Every project must define `allowed_hosts`. The control plane validates project URLs against that list, and the browser worker blocks browser requests that fall outside the list.

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

The browser worker routes Playwright requests through the same host policy:

- Requests outside `allowed_hosts` are aborted.
- Blocked requests are recorded as browser observation evidence.
- Blocked requests can produce an informational finding.

## Secret Handling

The alpha does not implement login automation or credential storage.

Current safeguards:

- API request logs do not include request bodies or query strings.
- Worker logs redact common token, password, secret, cookie, and authorization patterns.
- Screenshot and report artifacts should be treated as sensitive.

Future credential support must use a dedicated abstraction that can later support Vault, Kubernetes Secrets, or another secret manager.

## Non-Goals For This Alpha

- Exploit execution.
- Brute force testing.
- Destructive payloads.
- Broad crawling.
- OWASP ZAP integration.
- Active security scanning.

## Known Security Limitations

- DNS resolution checks are performed at validation/runtime, but DNS can change between checks.
- Browser screenshots can contain sensitive application data.
- MinIO uses local development credentials in Docker Compose.
- There is no API authentication in this alpha, so bind the API only in trusted local environments.

See [../SECURITY.md](../SECURITY.md) for vulnerability reporting.

# Security Policy

Qualora is a QA and security-adjacent automation tool. Its own safety model matters as much as the checks it performs.

## Supported Versions

Qualora is pre-release. No stable versions are supported yet.

| Version | Supported |
| --- | --- |
| v0.3.0-alpha | Best-effort alpha support |
| v0.2.0-alpha | Best-effort alpha support |
| v0.1.0-alpha | Best-effort alpha support |

## Reporting A Vulnerability

Please do not open public GitHub issues for vulnerabilities.

Preferred reporting process before the first public release:

1. Use GitHub private vulnerability reporting if it is enabled for the repository.
2. If unavailable, contact the maintainers privately. A permanent security contact should be confirmed before the first release.

Temporary contact placeholder: `security@operalith.com`

## Testing Scope

Only test systems you own or are explicitly authorized to test.

Qualora must respect project-level allowed hosts. Browser automation, API checks, passive security checks, artifact collection, and future integrations must all enforce that boundary.

The v0.3.0-alpha API and web UI do not include authentication. Expose them only in trusted local or self-hosted environments, or put them behind an external access-control layer.

See [docs/security-model.md](docs/security-model.md) for the current alpha safety model.

## Product Safety Requirements

- Safe checks by default.
- No destructive actions by default.
- No aggressive scanning in the MVP.
- No credential, token, cookie, or authorization-header logging.
- Sensitive values must be redacted from errors, traces, reports, and debug output.
- Evidence artifacts such as screenshots, traces, logs, and generated reports should be treated as sensitive.
- Credentials should be stored behind an abstraction that can later support Vault, Kubernetes Secrets, or other secret managers.

## Out Of Scope For The MVP

- Exploit execution.
- Brute force testing.
- Destructive payloads.
- Broad unauthenticated crawling.
- OWASP ZAP active scans.

OWASP ZAP integration may be added later with explicit policy controls and safe defaults.

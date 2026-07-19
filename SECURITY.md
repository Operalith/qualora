# Security Policy

Qualora is a QA and security-adjacent automation tool. Its own safety model matters as much as the checks it performs.

## Supported Versions

Qualora is pre-release. No stable versions are supported yet.

| Version | Supported |
| --- | --- |
| v0.20.0-alpha | Best-effort alpha support |
| v0.18.0-alpha | Best-effort alpha support |
| v0.17.0-alpha | Best-effort alpha support |
| v0.15.0-alpha | Best-effort alpha support |
| v0.14.0-alpha | Best-effort alpha support |
| v0.13.0-alpha | Best-effort alpha support |
| v0.12.0-alpha | Best-effort alpha support |
| v0.11.0-alpha | Best-effort alpha support |
| v0.10.0-alpha | Best-effort alpha support |
| v0.9.0-alpha | Best-effort alpha support |
| v0.8.0-alpha | Best-effort alpha support |
| v0.7.0-alpha | Best-effort alpha support |
| v0.6.0-alpha | Best-effort alpha support |
| v0.5.0-alpha | Best-effort alpha support |
| v0.4.0-alpha | Best-effort alpha support |
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

The v0.20.0-alpha API and web UI include local first-run admin authentication. This is intentionally minimal alpha authentication with one admin role, no password reset, no SSO/OIDC/SAML, no login rate limiting, and no audit log yet. Expose Qualora only in trusted local or self-hosted environments, or put it behind additional network access controls.

Report intelligence, baseline comparison, quality gates, CI run summaries, and issue export previews are deterministic and computed from already stored findings, safe result rows, grouped finding fingerprints, and metadata. They must not include credentials, cookies, local/session storage, auth headers, tokens, full HTML, screenshots, request bodies, response bodies, provider secrets, encrypted secret payloads, or raw AI prompts. Sensitive query values are redacted before URLs are used for grouping, fingerprints, or issue content.

See [docs/security-model.md](docs/security-model.md) for the current alpha safety model.

## Product Safety Requirements

- Safe checks by default.
- No destructive actions by default.
- No aggressive scanning in the MVP.
- No credential, token, cookie, or authorization-header logging.
- Sensitive values must be redacted from errors, traces, reports, and debug output.
- Evidence artifacts such as screenshots, traces, logs, and generated reports should be treated as sensitive.
- Evidence object downloads must only serve Qualora-owned evidence records and must not expose arbitrary S3 keys, filesystem paths, or object-store credentials.
- AI provider API keys and extra headers must not be returned by API responses and must be encrypted at rest.
- AI analysis must use sanitized report data, with redaction enabled by default.
- AI-assisted test planning must use sanitized project/run/report metadata, with redaction enabled by default.
- AI-generated test plans must remain reviewable suggestions and must not be executed automatically.
- Credential profile username/password values must be encrypted at rest and never returned raw by API responses.
- API auth profile bearer tokens, API keys, usernames, and passwords must be encrypted at rest and never returned raw by API responses.
- Authenticated API smoke must not expose Authorization headers, API keys, basic auth values, cookies, request bodies, response bodies, or raw token material in results, reports, AI input, CI output, or issue export.
- Local admin passwords must be hashed strongly and never returned by API responses.
- Session tokens must be stored hashed, delivered only as HTTP-only cookies, and never logged or returned in JSON.
- Mutating protected API requests must require CSRF validation.
- Deterministic login checks must only use configured selectors on the configured login form.
- Authenticated browser smoke must not expose cookies, session storage, local storage, authorization headers, tokens, or raw credentials.
- CI run output must not print local admin passwords, tracker tokens, provider secrets, credential profile secrets, cookies, or CSRF/session tokens.
- Issue export config tokens must be encrypted at rest and never returned raw or encrypted by API responses.
- Issue export must use grouped sanitized findings only and must not send screenshots, raw logs, full HTML, request bodies, response bodies, cookies, browser storage, auth headers, tokens, credentials, or provider secrets to trackers.
- Role-aware authorization checks must be explicit, deterministic, read-only, same-origin or allowed-host enforced, and limited to one configured target navigation after selector-based login.
- Authorization reports and AI input must not include passwords, raw usernames, cookies, session storage, local storage, authorization headers, tokens, screenshots, raw HTML, or browser storage contents by default.
- Application discovery must remain bounded, deterministic, same-origin by default, and allowed-host enforced.
- Application discovery must not submit forms, click arbitrary buttons, run payloads, perform destructive actions, crawl external domains by default, or use autonomous AI browser control.
- Application discovery reports and AI inputs must not include cookies, local/session storage, auth headers, tokens, credentials, full HTML, request bodies, or response bodies.
- Interactive Safe Explorer must remain deterministic, bounded, same-origin by default, and allowed-host enforced.
- Interactive Safe Explorer may inspect visible links, forms, buttons, submit controls, and inputs for metadata, but it must execute only policy-approved safe navigation actions.
- Interactive Safe Explorer must skip unsafe, external, unsupported, duplicate, sensitive-query, and policy-blocked actions with explicit reasons.
- Interactive Safe Explorer must not let AI choose or execute actions, submit POST forms, fill arbitrary forms, click arbitrary buttons, run payloads, fuzz inputs, perform active scanning, perform destructive actions, or crawl external domains by default.
- Interactive Safe Explorer reports and AI inputs must not include credentials, cookies, local/session storage, auth headers, tokens, full HTML, request bodies, or response bodies.
- Passive quality checks must remain read-only metadata checks. They must not submit forms, click arbitrary buttons, guess sensitive paths, run payloads, fuzz inputs, perform active scanning, perform destructive actions, or use autonomous AI browser control.
- Quality check reports and AI inputs must not include cookie values, browser storage, auth headers, tokens, credentials, full HTML, request bodies, or response bodies.
- Safe QA Runs must remain discovery-aware orchestration only: reviewable AI plans, deterministic preview, and explicit safe DSL execution without AI browser control, arbitrary clicks, form submission, active scanning, fuzzing, or destructive actions.
- Guided project setup must remain orchestration only. It may create safe configuration and start selected safe checks, but it must not add autonomous browser control, active scanning, fuzzing, arbitrary form submission, destructive behavior, or secret exposure.
- Baselines must store deterministic report fingerprints and summary metadata only.
- Report comparisons and quality gate evaluations must not start workers, send data to AI, hide raw findings, perform active scanning, fuzz inputs, execute payloads, or mutate tested systems.
- OpenAPI import must not execute API operations.
- Safe API smoke execution must remain read-only by default.
- Mutating, authenticated, request-body, unresolved-parameter, and sensitive API operations must be skipped unless a future explicit design changes this policy.
- API smoke results must not store request bodies or response bodies.
- Screenshots, full HTML, cookies, credentials, authorization headers, and full network bodies must not be sent to AI by default.
- Credentials should stay behind an abstraction that can later support Vault, Kubernetes Secrets, or other secret managers.

## Out Of Scope For The MVP

- Exploit execution.
- Brute force testing.
- Destructive payloads.
- Broad unauthenticated crawling.
- Arbitrary form submission or autonomous login automation.
- Uncontrolled application crawling.
- OWASP ZAP active scans.
- Autonomous AI browser control.
- Autonomous authorization attack generation or execution.

OWASP ZAP integration may be added later with explicit policy controls and safe defaults.

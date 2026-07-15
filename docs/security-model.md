# Security Model

Qualora is security-adjacent automation. The v0.9.0-alpha safety model is intentionally conservative.

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

The API worker and v0.9 control-plane API smoke executor validate `api_base_url`, `openapi_url`, imported OpenAPI URLs, OpenAPI server URLs, and every executed OpenAPI operation URL against the same host policy.

Default API behavior:

- Safe baseline `GET` against `api_base_url`.
- OpenAPI document fetch from `openapi_url`.
- Safe OpenAPI methods only: `GET`, `HEAD`, and `OPTIONS`.
- Unsafe methods such as `POST`, `PUT`, `PATCH`, and `DELETE` are skipped.
- Imported OpenAPI specs are parsed and classified before any API requests are executed.
- Auth-required operations are skipped in v0.9.
- Operations with required request bodies are skipped.
- Path parameters are skipped unless a safe `example`, `default`, or `enum` value exists.
- Required query parameters are sent only when a safe sample exists and the parameter name is not secret-like.
- API smoke execution does not store request bodies or response bodies.
- `destructive_actions=true` is not supported by the safe API smoke executor.

## Safe Test Plan Execution

AI test plans are suggestions and are never executed automatically. A user must explicitly preview and start safe execution.

The control plane maps model-generated plan JSON into a deterministic safe execution DSL. It queues only steps from scenarios that are:

- `automation_candidate=true`.
- `destructive=false`.
- `requires_authentication=false`.
- Not obviously login, payment, submit, upload, mutation, admin, exploit, SQLi, XSS, SSRF, brute-force, or destructive flows.

Supported browser actions:

- `goto`
- `assert_title_contains`
- `assert_url_contains`
- `assert_text_visible`
- `assert_element_visible`
- `assert_link_exists`
- `check_link_status`
- `capture_screenshot`
- `collect_browser_signals`
- `wait_for_load_state`
- `assert_no_console_errors`
- `assert_no_failed_requests`

Unsupported, ambiguous, authenticated, destructive, mutating, out-of-scope, and sensitive-query steps are persisted as skipped with reasons. The worker revalidates same-origin frontend URLs and `allowed_hosts` before navigation or link checks. The worker never executes model text as code and never performs POST/PUT/PATCH/DELETE actions from a generated plan.

## Credential Profiles And Login Checks

Credential profiles are project-scoped and intended for deterministic test accounts in trusted local/self-hosted environments. Username and password values are encrypted at rest using `QUALORA_ENCRYPTION_KEY`. API responses expose only safe metadata such as configured flags and a masked username display hint; they never return raw usernames or passwords.

Login automation is intentionally narrow:

- The configured `login_url` must pass `allowed_hosts` validation and match the project `frontend_url` origin.
- The browser worker fills only the configured username/password selectors.
- The browser worker clicks only the configured submit selector.
- Success is evaluated through configured URL/text criteria, with optional failure text detection.
- Authenticated browser smoke visits one configured relative same-origin target path after the login flow.
- No autonomous AI browser control is used.
- No arbitrary form submission, crawling, MFA handling, role switching, upload, payment, admin, or destructive action support is included.

Login evidence stores screenshots when configured plus metadata such as login status, safe URLs, page title, duration, console errors, failed requests, and blocked requests. It must not expose passwords, raw usernames, cookies, session storage, local storage, authorization headers, tokens, or browser storage contents.

## Web UI Exposure

The v0.9.0-alpha web UI has no authentication or authorization. It can create projects, manage credential profiles, test deterministic logins, start browser/API/authenticated smoke runs, import API specs, start safe API smoke runs, configure AI providers, run AI analysis, generate AI-assisted test plans, preview/start safe test plan executions, and display report/evidence metadata through the control-plane API.

Use it only in trusted local or self-hosted environments. Do not expose `qualora-web` or `qualora-api` directly to untrusted networks without adding an external access-control layer.

## Secret Handling

Current safeguards:

- API request logs do not include request bodies or query strings.
- Worker logs redact common token, password, secret, cookie, and authorization patterns.
- Credential profile username/password values are encrypted at rest and never returned raw.
- API evidence strips URL userinfo, query strings, and fragments.
- API smoke result rows store method, path, resolved URL, status, HTTP status, duration, content type, response size, errors, and skip reasons, but not request bodies or response bodies.
- Evidence object downloads are served only for evidence records already known to Qualora; callers cannot provide arbitrary S3 keys or filesystem paths.
- AI provider API keys are encrypted at rest.
- AI provider extra headers are treated as sensitive and encrypted at rest.
- AI provider responses never include raw API keys or raw extra headers.
- Screenshot and report artifacts should be treated as sensitive.

The Docker Compose default `QUALORA_ENCRYPTION_KEY` is an insecure development fallback. Set a strong value before storing real provider credentials or credential profiles. Future credential support should keep the current abstraction and add Vault, Kubernetes Secrets, or another secret manager.

## AI Safety

AI is disabled until a provider is configured. Qualora works without AI.

The AI input builder sends sanitized structured report data only. By default it may include run status, summary counts, finding titles/categories/severities/summaries, safe evidence metadata, browser/API/login metadata, API smoke result summaries, and job metadata.

The AI input builder does not send by default:

- Cookies.
- Authorization headers.
- Usernames.
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

AI-assisted test planning uses the same sanitized input path, plus optional user-provided product context. Do not put secrets, test credentials, cookies, API keys, or customer data in product context. Generated plans are stored as reviewable suggestions and are not executed automatically by Qualora. The v0.9 safe execution path can run only the approved deterministic browser DSL after explicit user action; it must not control the browser through free-form model text, call mutating APIs, submit forms, or perform unsupported generated steps.
AI-assisted test planning uses the same sanitized input path, plus optional user-provided product context. Do not put secrets, test credentials, cookies, API keys, or customer data in product context. Generated plans are stored as reviewable suggestions and are not executed automatically by Qualora. The v0.9 safe execution path can run only the approved deterministic browser DSL after explicit user action; it must not control the browser through free-form model text, call mutating APIs, submit forms, or perform unsupported generated steps.

## Non-Goals For This Alpha

- Exploit execution.
- Brute force testing.
- Destructive payloads.
- Broad crawling.
- Authenticated API testing.
- Schema fuzzing.
- Autonomous AI browser control.
- Automatic execution of generated AI test plans.
- Free-form AI-controlled browser automation.
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
- AI analysis and AI test plan quality depend on the configured provider and the sanitized evidence available in the report.

See [../SECURITY.md](../SECURITY.md) for vulnerability reporting.

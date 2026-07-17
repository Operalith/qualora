# Security Model

Qualora is security-adjacent automation. The v0.15.0-alpha safety model is intentionally conservative.

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

## Application Discovery

Application discovery in v0.15 is deterministic and safe by default. It is intended to build a lightweight application map, not to perform uncontrolled browser autonomy.

Discovery execution rules:

- Starts at the project `frontend_url` or a user-provided `start_url`.
- Defaults to `same_origin_only=true`, `max_pages=20`, and `max_depth=2`.
- Hard caps are `max_pages<=100` and `max_depth<=5`.
- Every navigation must pass `allowed_hosts`; same-origin discovery must stay on the project frontend origin.
- URL fragments are stripped and sensitive query parameter values are redacted before storage.
- Duplicate normalized URLs are not revisited.
- External, unsupported-scheme, non-HTML/download, and unsafe-looking links are skipped with recorded reasons.
- Forms and fields are recorded as metadata only.
- Screenshots and browser observation metadata may be stored as evidence.

Discovery does not:

- Submit forms.
- Click arbitrary buttons.
- Execute payloads.
- Use autonomous AI browser control.
- Perform destructive actions.
- Crawl external domains by default.
- Store full HTML, cookies, local/session storage, auth headers, tokens, credentials, request bodies, or response bodies.

## Passive Quality Checks

Quality checks in v0.15 are deterministic browser-worker observations. They are intended to surface obvious front-end quality issues, not to perform penetration testing, WCAG certification, Lighthouse audits, or exhaustive performance analysis.

Quality execution rules:

- Checks run against the project frontend URL, a selected completed discovery run, the latest completed discovery run, or a deterministic selector-authenticated session.
- Defaults are `max_pages=10`; the API cap is `max_pages<=50`.
- Every page visit must stay on the project frontend origin and pass `allowed_hosts`.
- Security checks use loaded page metadata only: response headers, cookie flags without values, forms, resource URLs, sensitive query parameter names, mixed-content observations, and obvious source-map exposure.
- Accessibility checks use basic document and element metadata: title, language, main landmark, image alt text, form labels, button names, and link text.
- Performance/front-end checks use navigation timing, console errors, failed resources, request counts, large JavaScript observations, and image dimension metadata.
- Quality evidence is metadata only and must not include cookie values, browser storage, authorization headers, tokens, credentials, request bodies, response bodies, full HTML, or screenshots by default.

Quality checks do not:

- Submit forms.
- Click arbitrary buttons.
- Guess sensitive paths.
- Run payloads.
- Fuzz inputs.
- Perform active scans.
- Perform destructive actions.
- Use autonomous AI browser control.

## API Request Enforcement

The API worker and v0.15 control-plane API smoke executor validate `api_base_url`, `openapi_url`, imported OpenAPI URLs, OpenAPI server URLs, and every executed OpenAPI operation URL against the same host policy.

Default API behavior:

- Safe baseline `GET` against `api_base_url`.
- OpenAPI document fetch from `openapi_url`.
- Safe OpenAPI methods only: `GET`, `HEAD`, and `OPTIONS`.
- Unsafe methods such as `POST`, `PUT`, `PATCH`, and `DELETE` are skipped.
- Imported OpenAPI specs are parsed and classified before any API requests are executed.
- Auth-required operations are skipped in the current alpha.
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

## Role-Aware Authorization Checks

Authorization checks are explicit, deterministic, and conservative. They are intended for dedicated test accounts and test data only.

Credential profiles can include role metadata such as `admin`, `readonly`, `customer-a`, or `customer-b`. The role metadata is descriptive and project-scoped; it does not grant access inside Qualora.

Authorization execution rules:

- A user must explicitly create each authorization check.
- Browser authorization checks log in with the configured actor credential profile.
- The worker navigates only to the configured `target_url` or path.
- Targets must stay on the project `frontend_url` origin and pass `allowed_hosts` validation.
- The worker compares the observed outcome with the configured expected outcome, `allowed` or `denied`.
- Denied outcomes are detected through HTTP `401`, `403`, `404`, or configured denied text.
- Allowed outcomes are detected through successful page load and optional success text.
- Ambiguous outcomes are recorded as unknown findings.
- Screenshots and `authorization_observations` metadata are recorded as evidence.

Authorization checks do not:

- Crawl the application.
- Submit arbitrary forms.
- Execute payloads.
- Fuzz inputs.
- Use POST/PUT/PATCH/DELETE.
- Perform active exploitation.
- Use autonomous AI browser control.
- Send credentials, cookies, storage, auth headers, or tokens to AI.

Authenticated API authorization testing is not fully supported in the current alpha. API-style authorization checks are skipped unless a future safe design adds explicit authenticated API support.

## Guided Onboarding

Guided project setup is a convenience layer over existing safe APIs. It can create a project, optionally configure an OpenAI-compatible provider, optionally create an encrypted credential profile, optionally import an OpenAPI spec, and start selected safe workflows.

Guided onboarding must keep these boundaries:

- Reject destructive project setup.
- Do not add autonomous browser control.
- Do not add arbitrary form submission.
- Do not add active scanning, fuzzing, exploitation, or destructive testing.
- Do not send credentials, provider secrets, cookies, browser storage, authorization headers, or tokens to AI.
- Do not return raw passwords, usernames, provider API keys, encrypted secret payloads, cookies, browser storage, authorization headers, or tokens.
- Return skipped workflow reasons when AI, credentials, OpenAPI specs, or target URLs are not configured.

## Web UI Exposure

The v0.15.0-alpha web UI and control-plane API require local authentication after first-run setup. On a fresh database, `POST /api/v1/setup/admin` creates the single local admin account. The setup route is rejected after a user exists. After setup, project data, credential profiles, AI provider configuration, reports, evidence, runs, API specs, test plans, discovery reports, and authorization reports require a valid local session.

Sessions use an HTTP-only `qualora_session` cookie. Mutating protected API requests must include a CSRF token from the `qualora_csrf` cookie in the `X-Qualora-CSRF` header. Health, setup status, first-run admin setup, login, logout, and session introspection endpoints are intentionally public.

This alpha still provides only one local admin role. It does not include user management, password reset, login rate limiting, audit logging, SSO/OIDC/SAML, enterprise RBAC, teams, or multi-tenancy. Use it only in trusted local or self-hosted environments. Do not expose `qualora-web` or `qualora-api` directly to untrusted networks without additional network controls.

## Secret Handling

Current safeguards:

- API request logs do not include request bodies or query strings.
- Local admin passwords are hashed with Argon2id and are never returned by API responses.
- Session tokens are stored hashed and are never returned in JSON responses.
- Worker logs redact common token, password, secret, cookie, and authorization patterns.
- Credential profile username/password values are encrypted at rest and never returned raw.
- Authorization check evidence stores role/profile labels and outcomes, not passwords, raw usernames, cookies, local/session storage, authorization headers, or tokens.
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

The AI input builder sends sanitized structured report data only. By default it may include run status, summary counts, finding titles/categories/severities/summaries, safe evidence metadata, browser/API/login/authorization metadata, quality check summaries and safe quality result metadata, API smoke result summaries, and job metadata. Discovery reports can be sent to AI test planning only through sanitized discovery-aware inputs in v0.15; those inputs are limited to discovery summaries, page paths/titles/statuses, form/link metadata, finding summaries, and evidence metadata.

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

AI-assisted test planning uses the same sanitized input path, plus optional user-provided product context. Do not put secrets, test credentials, cookies, API keys, or customer data in product context. Generated plans are stored as reviewable suggestions and are not executed automatically by Qualora. The v0.15 safe execution path can run only the approved deterministic browser DSL after explicit user action; it must not control the browser through free-form model text, call mutating APIs, submit forms, or perform unsupported generated steps. Authorization execution, application discovery, and guided login setup are deterministic and user-configured, not AI-generated.

## Safe QA Runs

Safe QA Runs in v0.15 orchestrate discovery, AI test planning, and safe test plan execution without changing the safety boundary.

Allowed behavior:

- Reuse or create a bounded application discovery run.
- Optionally run passive quality checks against discovered safe pages.
- Generate a reviewable AI test plan from sanitized project/report/discovery metadata.
- Persist safe execution coverage from the deterministic mapper.
- Stop after preview by default.
- Execute only persisted safe DSL steps after an explicit user request.
- Serve JSON/HTML reports that link discovery, plan, preview, optional execution, and safety metadata.

Safe QA Runs must not:

- Send credentials, cookies, local/session storage, auth headers, tokens, full HTML, screenshots, request bodies, or response bodies to AI.
- Give an LLM browser control.
- Execute free-form model text.
- Click arbitrary buttons.
- Submit forms.
- Crawl beyond the bounded discovery policy.
- Run payloads, active scans, fuzzing, mutation, uploads, payments, admin actions, or destructive actions.

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
- Local authentication is intentionally minimal: one admin role, no user management UI, no password reset, no rate limiting, and no audit log yet.
- Screenshot preview/download through the control-plane API is available for stored evidence records and can expose sensitive application state to the local admin.
- Anyone with the local admin session can configure or use AI providers.
- AI analysis and AI test plan quality depend on the configured provider and the sanitized evidence available in the report.
- Quality checks are heuristic metadata checks and can miss real security, accessibility, and performance issues.

See [../SECURITY.md](../SECURITY.md) for vulnerability reporting.

# Security Model

Qualora is security-adjacent automation. The v0.24.0-alpha safety model is intentionally conservative.

Qualora Demo Lab does not broaden that model. Its unsafe-looking links, POST forms, missing headers, broken assets, errors, and contract mismatches are inert local fixtures. Mutation handlers return `405`, no real data is stored, no external services are called, and existing allowlist and policy gates remain authoritative. The documented Demo Lab passwords and tokens are fake and must never be reused.

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

Application discovery in v0.22 is deterministic and safe by default. It is intended to build a lightweight application map, not to perform uncontrolled browser autonomy.

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

## Interactive Safe Explorer

Interactive Safe Explorer in v0.22 is deterministic and safe by default. It is intended to demonstrate bounded, human-understandable page exploration, not autonomous browser control.

Safe Explorer execution rules:

- Starts at the project `frontend_url` or a user-provided `start_url`.
- Defaults to `same_origin_only=true`, `max_steps=10`, `max_depth=2`, and `allow_get_forms=false`.
- Hard caps are `max_steps<=50` and `max_depth<=5`.
- Every navigation must pass `allowed_hosts`; same-origin runs must stay on the project frontend origin.
- Optional authentication uses only configured credential profile selectors and never AI.
- Visible links, forms, buttons, submit controls, and inputs may be inspected for metadata.
- Only classified safe navigation actions are executed by default.
- Unsafe, external, unsupported, duplicate, sensitive-query, and policy-blocked actions are skipped with reasons.
- GET forms are skipped unless `allow_get_forms=true`; POST/PUT/PATCH/DELETE-style forms are skipped.
- Screenshots, action metadata, browser observations, findings, and skip reasons may be stored as evidence/report data.

Safe Explorer does not:

- Let AI choose or execute actions.
- Submit POST forms.
- Fill arbitrary forms.
- Click arbitrary buttons.
- Execute payloads.
- Perform destructive actions.
- Crawl external domains by default.
- Store full HTML, cookies, local/session storage, auth headers, tokens, credentials, request bodies, or response bodies.
- Send credentials, cookies, browser storage, auth headers, tokens, screenshots, full HTML, request bodies, or response bodies to AI.

## Safe Form Testing

Safe Form Testing in v0.22 is deterministic and conservative. It is intended to validate simple non-mutating forms such as search, filter, sort, and navigation/query forms, not arbitrary workflow automation.

Safe Form Testing execution rules:

- Starts from a project `frontend_url`, a user-provided target URL, or pages from a completed discovery run.
- Every page and submitted URL must pass `allowed_hosts` and stay same-origin by default.
- Only `GET` forms classified as safe are eligible for execution.
- Safe classes are search, filter, sort, and navigation-like forms with same-origin actions.
- Bounded deterministic values are used, such as `demo` for search/query fields, first safe select options for filter/sort fields, small numbers, or stable dates.
- Submitted URLs are stored with sensitive query values redacted where applicable.
- Screenshots, form observations, form submission metadata, findings, classifications, decisions, and skip reasons may be stored.

Safe Form Testing skips:

- `POST`, `PUT`, `PATCH`, `DELETE`, and other mutating methods.
- Password fields, file fields, hidden sensitive fields, and required fields that cannot be safely filled.
- External action URLs.
- Sensitive parameter names such as password, token, secret, api_key, card, cvv, account, auth, and session.
- Login, password reset, payment, checkout, transfer, refund, delete, cancel, deactivate, upload, profile/account/admin mutation, destructive, unknown, and unsupported forms.

Safe Form Testing does not:

- Submit arbitrary forms.
- Submit POST/mutating forms.
- Fuzz inputs or generate payloads.
- Perform active security scanning.
- Execute destructive actions.
- Store raw form values, request bodies, response bodies, full HTML, cookies, local/session storage, auth headers, tokens, credentials, or browser storage.
- Send credentials, cookies, browser storage, auth headers, tokens, screenshots, full HTML, request bodies, response bodies, or raw form values to AI.

## Passive Quality Checks

Quality checks in v0.22 are deterministic browser-worker observations. They are intended to surface obvious front-end quality issues, not to perform penetration testing, WCAG certification, Lighthouse audits, or exhaustive performance analysis.

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

The API worker and v0.22 control-plane API smoke executor validate `api_base_url`, `openapi_url`, imported OpenAPI URLs, OpenAPI server URLs, and every executed OpenAPI operation URL against the same host policy.

API auth profile secrets are encrypted at rest with `QUALORA_ENCRYPTION_KEY`. API responses expose configured flags and safe display hints only, never raw bearer tokens, API keys, usernames, passwords, Authorization headers, or encrypted payloads. Authenticated API smoke may inject a selected API auth profile only into safe read-only requests for the configured project API target.

Default API behavior:

- Safe baseline `GET` against `api_base_url`.
- OpenAPI document fetch from `openapi_url`.
- Safe OpenAPI methods only: `GET`, `HEAD`, and `OPTIONS`.
- Unsafe methods such as `POST`, `PUT`, `PATCH`, and `DELETE` are skipped.
- Imported OpenAPI specs are parsed and classified before any API requests are executed.
- Auth-required operations are skipped unless a user explicitly starts an authenticated API smoke run with an enabled project API auth profile.
- Operations with required request bodies are skipped.
- Path parameters are skipped unless a safe `example`, `default`, or `enum` value exists.
- Required query parameters are sent only when a safe sample exists and the parameter name is not secret-like.
- API smoke execution does not store request bodies, response bodies, raw Authorization headers, bearer tokens, API keys, basic auth values, cookies, or secret query values.
- API auth material is not sent to AI providers, issue trackers, CI output, reports, or logs.
- Contract validation is lightweight and bounded to status codes, obvious content types, JSON parseability, and simple schema checks. It is not fuzzing and does not generate payloads.
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

The v0.24.0-alpha web UI and control-plane API require local authentication after first-run setup. On a fresh database, `POST /api/v1/setup/admin` creates the single local admin account. The setup route is rejected after a user exists. After setup, project data, credential profiles, API auth profiles, AI provider configuration, reports, evidence, runs, API specs, test plans, discovery reports, Safe Explorer reports, Safe Form Testing reports, authorization reports, CI runs, and issue export configs require a valid local session.

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

The Docker Compose default `QUALORA_ENCRYPTION_KEY` is an insecure development fallback. Set a strong value before storing real provider credentials, credential profiles, API auth profiles, or issue export tokens. Future credential support should keep the current abstraction and add Vault, Kubernetes Secrets, or another secret manager.

## AI Safety

AI is disabled until a provider is configured. Qualora works without AI.

The AI input builder sends sanitized structured report data only. By default it may include run status, summary counts, finding titles/categories/severities/summaries, safe evidence metadata, browser/API/login/authorization/form metadata, safe API auth summary metadata, quality check summaries and safe quality result metadata, API smoke result summaries, and job metadata. Discovery reports can be sent to AI test planning only through sanitized discovery-aware inputs in v0.22; those inputs are limited to discovery summaries, page paths/titles/statuses, form/link metadata, finding summaries, and evidence metadata. Safe Explorer does not send action execution context to AI and does not allow AI action choice.

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

## Report Intelligence Safety

Report intelligence in `v0.24.0-alpha` is deterministic and computed inside the control plane from already persisted finding, result, and safe evidence metadata. It normalizes severity, groups repeated findings, classifies noisy repeated signals, summarizes affected pages, and creates executive summaries without sending data to an AI provider.

Report intelligence must not include credentials, cookies, local storage, session storage, authorization headers, tokens, full HTML, screenshots, request bodies, response bodies, provider secrets, or encrypted secret payloads. URLs used for grouping are redacted for sensitive query names before fingerprints or report fields are produced. Raw findings remain available, so grouping must never be treated as deletion or suppression of evidence.

## Baselines, Comparisons, And Quality Gates

Baselines in `v0.24.0-alpha` are deterministic report snapshots. A baseline stores grouped finding fingerprints, severity counts, grouped finding counts, raw finding counts, and source report metadata for a known project report. It must not store credentials, cookies, local/session storage, authorization headers, tokens, screenshots, full HTML, request bodies, response bodies, provider secrets, encrypted secret payloads, or raw AI prompts.

Comparison is a read-only control-plane operation. It compares fingerprints from the baseline with fingerprints from the current report and classifies new, fixed, unchanged, severity-changed, and affected-scope-changed findings. It does not start a browser worker, API worker, security scan, AI call, payload, crawl, fuzzing run, or destructive action.

Quality gates evaluate comparison summaries and current severity counts. They are intended as alpha CI/release signals and do not replace human review. Gate evaluation must not hide raw findings, mutate project data, send data to AI, or execute new tests.

CI runs in `v0.24.0-alpha` orchestrate existing Safe QA, baseline comparison, and quality gate behavior. CI output must stay compact and sanitized. Scripts must not print `QUALORA_PASSWORD`, local admin session cookies, CSRF tokens, tracker tokens, provider secrets, credential profile secrets, cookies, browser storage, authorization headers, or raw target application credentials.

Issue export in `v0.24.0-alpha` is optional and uses grouped sanitized findings. Issue export configs store GitHub/GitLab tokens encrypted with `QUALORA_ENCRYPTION_KEY`; API/UI responses expose only `token_configured`. Issue titles and bodies must not contain credentials, cookies, local/session storage, auth headers, tokens, full HTML, screenshots, request bodies, response bodies, raw logs, provider secrets, encrypted secret payloads, or raw AI prompts. Dry-run is the default and should be used before actual tracker issue creation.

AI-assisted test planning uses the same sanitized input path, plus optional user-provided product context. Do not put secrets, test credentials, cookies, API keys, or customer data in product context. Generated plans are stored as reviewable suggestions and are not executed automatically by Qualora. The v0.22 safe execution path can run only the approved deterministic browser DSL after explicit user action; it must not control the browser through free-form model text, call mutating APIs, submit forms, or perform unsupported generated steps. Authorization execution, application discovery, Interactive Safe Explorer, Safe Form Testing, guided login setup, report intelligence, baselines, CI runs, and issue export previews are deterministic and user-configured, not AI-generated.

## Policy-Gated AI Browser Control

AI Browser Control in `v0.24.0-alpha` is AI-suggested but not direct AI browser control. The browser worker captures a sanitized observation, sends that observation plus the bounded user goal to an OpenAI-compatible provider, parses one strict JSON action, and runs the deterministic policy engine before Playwright can execute anything.

Allowed action types are limited to safe navigation, policy-approved safe GET form submission, assertions, metadata collection, screenshot capture, and `stop`: `goto`, `click_link`, `click_safe_navigation`, `submit_safe_get_form`, `assert_text_visible`, `assert_url_contains`, `assert_title_contains`, `capture_screenshot`, `collect_browser_signals`, and `stop`.

The sanitized AI input must not contain credentials, cookies, local/session storage, authorization headers, tokens, screenshots, full HTML, request bodies, response bodies, raw traces, provider secrets, encrypted secret payloads, or raw browser storage. Selector-based login may happen before a run when a credential profile is selected, but credentials and resulting browser session material are never sent to AI or included in reports.

The policy engine must reject external navigation when same-origin mode is enabled, disallowed hosts, sensitive query strings, destructive or mutating labels/paths, unsupported actions, duplicate loop targets, unsafe form submission, POST/mutating form submission, payload/fuzzing/exploit language, and arbitrary selectors that are not recognized as safe supported actions.

## Run Viewer And Real Provider Safety

The Run Viewer reads existing authenticated report, step, and evidence endpoints. It does not create a second browser-control path. Screenshot previews remain Qualora evidence and are never added to AI input. Viewer data must not expose credentials, cookies, local/session storage, authorization headers, tokens, full HTML, request bodies, response bodies, provider secrets, encrypted secret payloads, or raw browser storage.

Deterministic smoke, showcase smoke, and CI use Fake LLM. `scripts/run-demo-lab-real-llm.sh` is an explicit operator action for OpenAI-compatible providers and may incur provider cost. It validates required environment variable names before starting work, stores temporary session files with restricted permissions, and must never print `QUALORA_REAL_LLM_API_KEY` or secret extra-header values.

## Safe QA Runs

Safe QA Runs in v0.22 orchestrate discovery, AI test planning, and safe test plan execution without changing the safety boundary.

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
- Broad, mutating, fuzzed, or autonomous authenticated API testing.
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

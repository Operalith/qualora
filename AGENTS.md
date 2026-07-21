# AGENTS.md

Project-specific guidance for future Codex work in Qualora.

## Product Direction

Qualora is an open-source, self-hosted autonomous QA platform for web applications and APIs. Keep the MVP small, useful, and runnable on-prem with Docker Compose before adding Kubernetes, SaaS, enterprise, or marketplace features.

Default positioning: **Open-source AI-powered engineering tools for modern software operations.**

## Current Priorities

- Keep the `v0.22.0-alpha` Docker Compose MVP working.
- Backend/control plane first, with browser worker support.
- API worker support for safe API/OpenAPI checks.
- Imported OpenAPI specs, operation discovery, safe API smoke runs, and API result reports.
- Project-scoped encrypted API auth profiles and authenticated read-only API smoke checks.
- Lightweight OpenAPI status, content-type, JSON, and response schema contract validation.
- Minimal web UI support for local project/run/report workflows.
- Browser-only smoke run support and screenshot evidence preview/download.
- Optional OpenAI-compatible AI provider management and AI report analysis.
- Optional AI-assisted test planning as reviewable suggestions.
- Approved safe test plan execution for a small deterministic browser DSL.
- Project-scoped encrypted credential profiles and deterministic selector-based login checks.
- Authenticated browser smoke runs for one configured same-origin target path.
- Explicit role-aware authorization checks for configured browser URL targets.
- Safe deterministic application discovery and persistent application maps.
- Passive quality checks for security headers/cookies/forms, basic accessibility heuristics, and performance/front-end observations.
- Discovery-aware AI test plan generation from sanitized application maps.
- Safe QA Runs that preview first and execute only approved deterministic browser DSL steps.
- Interactive Safe Explorer for deterministic, bounded, safe navigation action exploration.
- Policy-gated AI Browser Control where AI proposes one typed action and Qualora executes only deterministic policy-approved safe actions.
- Safe Form Testing for deterministic classification and bounded same-origin GET search/filter/sort form execution.
- Guided project onboarding that creates a project, optionally configures AI, credentials, and OpenAPI specs, and starts selected safe checks.
- Dashboard, reports, and project readiness UI that make first-run workflows discoverable without hiding alpha limitations.
- Deterministic report intelligence with severity normalization, grouped findings, deduplication metadata, affected-page summaries, noise classification, and executive summaries.
- Deterministic Safe QA report baselines, regression comparisons, and CI quality gates.
- Native CI run orchestration for Safe QA baseline comparison and quality gates.
- Optional sanitized GitHub/GitLab issue export with encrypted tracker tokens and dry-run previews by default.
- Local first-run admin setup and session-protected API/web UI access.
- Docker Compose as the first deployment target.
- PostgreSQL for durable metadata.
- Redis for queues and short-lived run state.
- MinIO/S3-compatible storage for evidence artifacts.
- Playwright for real browser automation.
- Modular worker boundaries, but avoid premature distributed-system complexity.

## Repository Boundaries

- `apps/control-plane`: Go API and orchestration service.
- `apps/web`: Minimal React web UI for projects, runs, reports, findings, and evidence metadata.
- `workers/browser`: Playwright browser checks.
- `workers/api`: API and OpenAPI checks.
- `workers/security`: Passive, safe security checks.
- `workers/analyzer`: Evidence normalization and finding generation.
- `packages/report-engine`: Structured report generation.
- `packages/shared`: Shared schemas/utilities only when duplication becomes real.
- `api/openapi`: Internal API contracts.
- `deploy/docker-compose`: First supported deployment path.
- `deploy/helm`: Future Kubernetes packaging.

## Implementation Rules

- Prefer simple, explicit code over framework-heavy abstractions.
- Do not introduce paid SaaS assumptions.
- Do not add Kubernetes-only concepts before the Docker Compose path works.
- Keep the web UI focused on alpha workflows; do not add complex design systems, multi-user management, teams, billing, or SaaS assumptions without an explicit request.
- Do not introduce Temporal, OWASP ZAP, arbitrary login automation, or active security scanning in the MVP without an explicit request.
- Do not introduce autonomous AI browser control or native non-OpenAI-compatible provider SDKs without an explicit request.
- Do not execute AI-generated test plan steps automatically or as free-form model instructions.
- Do not let Safe QA Runs bypass review, safe execution mapping, or explicit user approval for execution.
- Do not use AI for login automation or authenticated browser control.
- Authorization checks must remain deterministic, explicitly user-configured, same-origin, allowlist-enforced, read-only, and limited to configured browser URL targets.
- Do not add crawling, fuzzing, payload execution, arbitrary form submission, destructive actions, or autonomous AI browser control to authorization checks.
- Application discovery must remain bounded, deterministic, same-origin by default, allowlist-enforced, and safe-link-only.
- Discovery must never submit forms, click arbitrary buttons, execute payloads, perform destructive actions, crawl external domains by default, or use autonomous AI browser control.
- Quality checks must remain passive, deterministic, metadata-only, same-origin, and allowlist-enforced.
- Quality checks must never submit forms, click arbitrary buttons, guess sensitive paths, execute payloads, fuzz inputs, perform active scans, perform destructive actions, crawl external domains by default, or use autonomous AI browser control.
- Safe test plan execution must remain explicit, previewable, same-origin, allowlist-enforced, non-destructive, and limited to the supported persisted DSL.
- Safe Explorer must remain deterministic, bounded, same-origin by default, allowlist-enforced, and limited to safe classified navigation actions.
- Safe Explorer must not use AI action choice, arbitrary clicking, arbitrary form submission, POST/mutating form execution, destructive actions, crawling external domains by default, active scanning, fuzzing, or payload execution.
- AI Browser Control must remain policy-gated: AI may propose one strict typed JSON action, but Qualora must validate it before Playwright executes anything.
- AI Browser Control must never become direct/free-form model control of Playwright.
- AI Browser Control may support only policy-approved `submit_safe_get_form` actions for observed, same-origin, GET, safe-classified forms with bounded non-sensitive values.
- AI Browser Control must not support arbitrary selectors, arbitrary clicking, arbitrary form submission, POST/mutating form execution, destructive actions, external crawling by default, active scanning, fuzzing, payload execution, or unsafe button clicks.
- Safe Form Testing must execute only same-origin safe GET forms such as search/filter/sort/navigation forms with bounded deterministic non-sensitive values.
- Safe Form Testing must skip POST/PUT/PATCH/DELETE forms, password/file fields, hidden sensitive fields, external actions, sensitive parameter names, login, payment, checkout, transfer, refund, delete, reset, cancel, deactivate, profile/account/admin mutation, destructive, unknown, and unsupported forms.
- Safe Form Testing must not store raw form values, cookies, local/session storage, auth headers, tokens, request bodies, response bodies, full HTML, or credentials in reports, evidence metadata, AI inputs, CI output, or issue export previews.
- Safe QA Runs must remain an orchestration layer over discovery, AI planning, and safe DSL execution; do not add arbitrary clicking, form submission, broad crawling, active scanning, fuzzing, or destructive actions.
- Guided setup must orchestrate existing safe capabilities; do not add new engines, autonomous browser control, active scanning, destructive behavior, or credential leakage through onboarding.
- Baseline comparison must stay deterministic and based on existing report intelligence fingerprints/grouped findings.
- Quality gates must work without AI and must not start workers, execute tests, hide raw findings, or mutate tested systems.
- CI runs must remain an orchestration layer over existing Safe QA, baseline comparison, and quality gate behavior; do not add new testing engines through CI mode.
- Issue export must remain optional, grouped-finding based, sanitized, and dry-run by default. Real tracker issue creation requires an explicit enabled config, encrypted token, and `dry_run=false`.
- API worker checks must stay safe by default: `GET`, `HEAD`, and `OPTIONS` only unless a later explicit policy supports more.
- Imported OpenAPI API smoke runs must stay read-only: skip mutating methods, required request bodies, unresolved parameters, sensitive paths/parameters, and external redirects.
- Auth-required OpenAPI operations may run only when an enabled project API auth profile is explicitly selected for an authenticated API smoke run.
- API auth profiles must use the existing encrypted secret abstraction and must never return raw bearer tokens, API keys, usernames, passwords, Authorization headers, encrypted payloads, or secret query values.
- Do not store API request bodies or response bodies in the current alpha API smoke path.
- Do not store API auth headers, raw token values, API keys, basic auth values, cookies, request bodies, or response bodies in API result rows, reports, AI inputs, CI output, or issue export previews.
- Keep worker contracts narrow and serializable.
- Prefer OpenAPI-first internal API design where practical.
- Keep report schemas structured enough for future UI/API consumers.
- Add tests around orchestration, host allowlisting, secret redaction, and report generation when those areas are implemented.
- Do not claim unsupported features in README, release notes, OpenAPI, or docs.
- Report intelligence must remain deterministic and must not hide or delete raw findings.

## Security And Safety Rules

- Never log test credentials, tokens, cookies, authorization headers, or secret values.
- Never log or return local admin passwords, password hashes, session tokens, CSRF tokens, credential profile secrets, or AI provider secrets.
- Local auth must remain simple unless explicitly expanded: one admin role, first-run setup, HTTP-only session cookie, CSRF protection for mutating protected API requests.
- Do not add SSO/OIDC/SAML, password reset, enterprise RBAC, teams, or multi-tenancy without an explicit request.
- Redact secrets in errors, traces, reports, and debug output.
- All browser, API, and security checks must enforce project allowed hosts.
- Security checks are passive and non-destructive by default.
- Do not add active exploitation, destructive payloads, brute force behavior, or broad crawling unless the user explicitly asks and a safe policy model exists.
- Treat screenshots, traces, and reports as sensitive artifacts.
- Evidence download endpoints must only serve database-backed Qualora evidence records and must never expose arbitrary S3 keys, local paths, or object-store credentials.
- Store credentials through a dedicated abstraction so local MVP storage can later move to Vault or Kubernetes Secrets.
- AI provider API keys and extra headers must be encrypted at rest and never returned raw by API responses.
- AI input must be built from sanitized report data only. Do not send screenshots, full HTML, cookies, credentials, authorization headers, or full network bodies to AI by default.
- Redaction must remain enabled by default for AI analysis and AI-assisted test planning.
- AI test planning must use sanitized project/run/report metadata only. Do not send screenshots, full HTML, cookies, credentials, authorization headers, raw traces, or full network bodies to AI by default.
- Discovery-aware AI test planning must use sanitized discovery summaries only. Do not send screenshots, full HTML, cookies, credentials, authorization headers, local/session storage, tokens, request bodies, or response bodies to AI by default.
- AI Browser Control must send only sanitized browser observations and bounded goals to AI. Do not send credentials, cookies, local/session storage, auth headers, tokens, screenshots, full HTML, request bodies, response bodies, raw traces, provider secrets, encrypted secret payloads, or browser storage.
- AI Browser Control policy decisions, suggestions, traces, reports, and findings must be safe to store and must not expose secrets or browser session material.
- AI Browser Control form suggestions must send only sanitized form metadata and must still pass deterministic policy before Playwright navigates to a generated GET URL.
- Safe Form Testing reports and evidence must include classifications, decisions, skip reasons, redacted submitted URLs, screenshots, findings, and bounded value summaries only; never raw submitted values or browser session material.
- Safe test plan execution must skip authenticated, destructive, mutating, submit/upload/admin, exploit, SQLi, XSS, SSRF, brute-force, out-of-scope, and unsupported actions with clear reasons.
- Safe API smoke must skip unsafe OpenAPI operations with clear reasons and must not send cookies, request bodies, response bodies, or secrets outside the configured allowed-host API target.
- Authenticated API smoke may inject API auth only from a selected API auth profile into safe read-only requests and must redact auth headers/tokens everywhere else.
- Authorization runs must never send credentials, cookies, local/session storage, auth headers, or tokens to AI or include them in evidence/reports.
- Discovery runs must never send credentials, cookies, local/session storage, auth headers, tokens, full HTML, request bodies, or response bodies to AI or include them in metadata/report fields.
- Quality checks must never send credentials, cookies, local/session storage, auth headers, tokens, full HTML, screenshots, request bodies, or response bodies to AI or include them in metadata/report fields.
- Baselines, comparisons, and quality gates must never include credentials, cookies, local/session storage, auth headers, tokens, full HTML, screenshots, request bodies, response bodies, provider secrets, encrypted secret payloads, or raw AI prompts.
- CI run summaries and scripts must never print or return local admin passwords, session cookies, CSRF tokens, credential profile secrets, provider secrets, tracker tokens, cookies, browser storage, authorization headers, full HTML, screenshots, request bodies, or response bodies.
- Issue export config APIs must never return raw tracker tokens. Issue previews and created issue bodies must be generated only from sanitized grouped findings and must not include credentials, cookies, local/session storage, auth headers, tokens, full HTML, screenshots, request bodies, response bodies, raw logs, provider secrets, encrypted secret payloads, or raw AI prompts.

## Contribution Style

- Update `README.md` or `docs/architecture/mvp.md` when changing product architecture.
- Keep PRs focused and explain user-facing behavior changes.
- When adding a service, include local run instructions and Docker Compose integration.
- When adding a worker, document its inputs, outputs, safety checks, and artifact behavior.

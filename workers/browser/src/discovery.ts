import type { FindingInput } from "./findings";

export type DiscoveryLinkDecision = {
  href: string;
  normalizedURL: string;
  linkText: string;
  sameOrigin: boolean;
  skipped: boolean;
  skipReason: string;
};

export type DiscoveryLinkPolicy = {
  sourceURL: string;
  frontendURL: string;
  allowedHosts: string[];
  sameOriginOnly: boolean;
};

export type ExtractedFormField = {
  field_name: string;
  field_type: string;
  placeholder: string;
  label: string;
  required: boolean;
};

export type ExtractedForm = {
  form_name: string;
  form_action: string;
  form_method: string;
  fields: ExtractedFormField[];
  submit_button_count: number;
};

export type DiscoveryFormSummary = {
  form_name: string;
  form_action: string;
  form_method: string;
  field_count: number;
  password_field_count: number;
  submit_button_count: number;
  classification: string;
  skipped_reason: string;
  fields: ExtractedFormField[];
};

export type DiscoveryPageFindingInput = {
  url: string;
  statusCode: number | null;
  loadError: string;
  bodyTextLength: number | null;
  consoleErrorCount: number;
  failedRequestCount: number;
  evidenceIds: string[];
};

const sensitiveQueryMarkers = [
  "access_token",
  "api_key",
  "apikey",
  "auth",
  "authorization",
  "credential",
  "jwt",
  "key",
  "password",
  "passwd",
  "secret",
  "session",
  "token"
];

const unsafeLinkMarkers = [
  "logout",
  "delete",
  "remove",
  "destroy",
  "reset",
  "payment",
  "transfer",
  "token",
  "password-reset",
  "reset-password",
  "forgot-password",
  "admin/delete",
  "admin/remove",
  "admin/destroy",
  "admin/reset",
  "admin/mutation"
];

const downloadExtensions = [
  ".7z",
  ".avi",
  ".csv",
  ".doc",
  ".docx",
  ".gif",
  ".gz",
  ".jpeg",
  ".jpg",
  ".mp3",
  ".mp4",
  ".pdf",
  ".png",
  ".tar",
  ".tgz",
  ".webp",
  ".xls",
  ".xlsx",
  ".zip"
];

export function normalizeDiscoveryURL(raw: string, baseURL?: string): string {
  const parsed = new URL(raw, baseURL);
  parsed.hash = "";
  parsed.protocol = parsed.protocol.toLowerCase();
  parsed.hostname = parsed.hostname.toLowerCase().replace(/\.$/, "");
  for (const key of Array.from(parsed.searchParams.keys())) {
    if (isSensitiveQueryName(key)) {
      parsed.searchParams.set(key, "[REDACTED]");
    }
  }
  parsed.searchParams.sort();
  return parsed.toString();
}

export function classifyDiscoveryLink(rawHref: string, linkText: string, policy: DiscoveryLinkPolicy): DiscoveryLinkDecision {
  const href = (rawHref || "").trim();
  const text = cleanText(linkText, 240);
  if (!href) {
    return emptyDecision(href, text, "empty_href");
  }

  let parsed: URL;
  try {
    parsed = new URL(href, policy.sourceURL);
  } catch {
    return emptyDecision(href, text, "invalid_url");
  }

  if (!["http:", "https:"].includes(parsed.protocol)) {
    return {
      href,
      normalizedURL: "",
      linkText: text,
      sameOrigin: false,
      skipped: true,
      skipReason: "unsupported_scheme"
    };
  }

  const normalizedURL = normalizeDiscoveryURL(parsed.toString());
  const source = new URL(policy.sourceURL);
  const frontend = new URL(policy.frontendURL);
  const sameOrigin = parsed.origin === source.origin && parsed.origin === frontend.origin;
  const hostAllowed = policy.allowedHosts.some((allowedHost) => hostMatches(parsed.hostname, allowedHost));

  if (policy.sameOriginOnly && !sameOrigin) {
    return skippedDecision(href, normalizedURL, text, sameOrigin, "external_link_skipped");
  }
  if (!hostAllowed) {
    return skippedDecision(href, normalizedURL, text, sameOrigin, "host_not_allowed");
  }
  if (looksLikeDownload(parsed)) {
    return skippedDecision(href, normalizedURL, text, sameOrigin, "non_html_resource");
  }
  if (looksUnsafe(parsed, text)) {
    return skippedDecision(href, normalizedURL, text, sameOrigin, "unsafe_link_skipped");
  }

  return {
    href,
    normalizedURL,
    linkText: text,
    sameOrigin,
    skipped: false,
    skipReason: ""
  };
}

export function summarizeDiscoveryForm(form: ExtractedForm): DiscoveryFormSummary {
  const fields = form.fields.map((field) => ({
    field_name: cleanText(field.field_name, 120),
    field_type: cleanText(field.field_type || "text", 40).toLowerCase(),
    placeholder: cleanText(field.placeholder, 160),
    label: cleanText(field.label, 160),
    required: Boolean(field.required)
  }));
  const passwordFieldCount = fields.filter((field) => field.field_type === "password").length;
  return {
    form_name: cleanText(form.form_name, 160),
    form_action: cleanText(form.form_action, 1000),
    form_method: cleanText(form.form_method || "get", 16).toLowerCase(),
    field_count: fields.length,
    password_field_count: passwordFieldCount,
    submit_button_count: Math.max(0, Number(form.submit_button_count || 0)),
    classification: passwordFieldCount > 0 ? "password_form" : fields.length > 0 ? "data_entry_form" : "empty_form",
    skipped_reason: "forms_are_not_submitted_by_discovery",
    fields
  };
}

export function buildDiscoveryPageFindings(input: DiscoveryPageFindingInput): FindingInput[] {
  const findings: FindingInput[] = [];
  if (input.loadError) {
    findings.push({
      title: "Discovery page load failed",
      severity: "high",
      category: "page_load_failure",
      confidence: "high",
      description: [`Summary: discovery could not load ${input.url}: ${input.loadError}`, `Steps to reproduce: open ${input.url} in a browser from the worker network.`].join("\n"),
      recommendation: "Verify that the route is reachable from the worker container and does not require unsupported interaction before initial load.",
      evidenceIds: input.evidenceIds
    });
  } else if (input.statusCode === 404) {
    findings.push({
      title: "Discovered internal page returned 404",
      severity: "medium",
      category: "not_found",
      confidence: "high",
      description: [`Summary: discovered page ${input.url} returned HTTP 404.`, `Steps to reproduce: open ${input.url} and confirm the route exists.`].join("\n"),
      recommendation: "Fix or remove links that point to missing internal routes.",
      evidenceIds: input.evidenceIds
    });
  } else if (input.statusCode !== null && input.statusCode >= 500) {
    findings.push({
      title: "Discovered page returned a server error",
      severity: "high",
      category: "server_error",
      confidence: "high",
      description: [`Summary: discovered page ${input.url} returned HTTP ${input.statusCode}.`, `Steps to reproduce: open ${input.url} and inspect the initial document response.`].join("\n"),
      recommendation: "Inspect the frontend service and upstream dependencies for server-side failures.",
      evidenceIds: input.evidenceIds
    });
  }

  if (input.consoleErrorCount > 0) {
    findings.push({
      title: "Console errors observed during discovery",
      severity: "medium",
      category: "console_error",
      confidence: "medium",
      description: [`Summary: ${input.consoleErrorCount} console error(s) were observed while loading ${input.url}.`, `Steps to reproduce: open ${input.url} and inspect browser console output during initial load.`].join("\n"),
      recommendation: "Review uncaught frontend exceptions and failed client-side initialization.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.failedRequestCount > 0) {
    findings.push({
      title: "Failed network requests observed during discovery",
      severity: "medium",
      category: "network_failure",
      confidence: "medium",
      description: [`Summary: ${input.failedRequestCount} failed network request(s) were observed while loading ${input.url}.`, `Steps to reproduce: open ${input.url} and inspect failed requests in browser developer tools.`].join("\n"),
      recommendation: "Ensure required first-party assets and APIs are reachable during page load.",
      evidenceIds: input.evidenceIds
    });
  }
  if (!input.loadError && input.statusCode !== null && input.statusCode < 400 && input.bodyTextLength === 0) {
    findings.push({
      title: "Discovered page appears empty",
      severity: "medium",
      category: "empty_page",
      confidence: "medium",
      description: [`Summary: ${input.url} loaded successfully but had no visible body text.`, `Steps to reproduce: open ${input.url} and inspect the rendered page body.`].join("\n"),
      recommendation: "Confirm the route renders meaningful visible content after initial load.",
      evidenceIds: input.evidenceIds
    });
  }
  return findings;
}

export function buildDiscoveryLinkFinding(decision: DiscoveryLinkDecision, evidenceIds: string[]): FindingInput | null {
  if (!decision.skipped) {
    return null;
  }
  if (decision.skipReason === "external_link_skipped") {
    return {
      title: "External link skipped during discovery",
      severity: "info",
      category: "external_link_skipped",
      confidence: "high",
      description: [`Summary: discovery skipped an external link to ${decision.normalizedURL || decision.href}.`, "Steps to reproduce: review the source page link list and compare the URL to the same-origin discovery policy."].join("\n"),
      recommendation: "Add required first-party hosts to allowed_hosts or keep external links out of discovery scope.",
      evidenceIds
    };
  }
  if (decision.skipReason === "unsafe_link_skipped") {
    return {
      title: "Unsafe-looking link skipped during discovery",
      severity: "low",
      category: "unsafe_link_skipped",
      confidence: "medium",
      description: [`Summary: discovery skipped a link that looked mutating or sensitive: ${decision.normalizedURL || decision.href}.`, "Steps to reproduce: inspect the link href/text and confirm whether it represents a safe navigation route."].join("\n"),
      recommendation: "Keep destructive or sensitive links out of automated discovery paths, or model them later with explicit safe policies.",
      evidenceIds
    };
  }
  return null;
}

export function buildDiscoveryFormFindings(form: DiscoveryFormSummary, pageURL: string, evidenceIds: string[]): FindingInput[] {
  const findings: FindingInput[] = [];
  const missingLabels = form.fields.filter((field) => !field.label && ["text", "email", "password", "search", "tel", "url"].includes(field.field_type)).length;
  if (missingLabels > 0) {
    findings.push({
      title: "Form field without visible label detected",
      severity: "low",
      category: "form_without_label",
      confidence: "medium",
      description: [`Summary: discovery found ${missingLabels} form field(s) without captured labels on ${pageURL}.`, "Steps to reproduce: inspect the form controls and verify accessible labels are present."].join("\n"),
      recommendation: "Add explicit labels or accessible names for form controls.",
      evidenceIds
    });
  }
  if (form.password_field_count > 0) {
    findings.push({
      title: "Password form detected",
      severity: "info",
      category: "password_form_detected",
      confidence: "high",
      description: [`Summary: discovery detected a password field on ${pageURL}.`, "Steps to reproduce: inspect the form metadata captured for the page."].join("\n"),
      recommendation: "Use a project-scoped credential profile for deterministic login checks; discovery will not submit this form.",
      evidenceIds
    });
  }
  return findings;
}

function emptyDecision(href: string, linkText: string, skipReason: string): DiscoveryLinkDecision {
  return {
    href,
    normalizedURL: "",
    linkText,
    sameOrigin: false,
    skipped: true,
    skipReason
  };
}

function skippedDecision(href: string, normalizedURL: string, linkText: string, sameOrigin: boolean, skipReason: string): DiscoveryLinkDecision {
  return {
    href,
    normalizedURL,
    linkText,
    sameOrigin,
    skipped: true,
    skipReason
  };
}

function isSensitiveQueryName(name: string): boolean {
  const normalized = name.toLowerCase().trim();
  return sensitiveQueryMarkers.some((marker) => normalized === marker || normalized.includes(marker));
}

function hostMatches(host: string, allowedHost: string): boolean {
  const normalizedHost = host.toLowerCase().replace(/\.$/, "");
  const allowed = allowedHost.toLowerCase().trim().replace(/\.$/, "");
  if (normalizedHost === allowed) {
    return true;
  }
  return allowed.startsWith("*.") && normalizedHost.endsWith(allowed.slice(1));
}

function looksLikeDownload(parsed: URL): boolean {
  const path = parsed.pathname.toLowerCase();
  return downloadExtensions.some((extension) => path.endsWith(extension));
}

function looksUnsafe(parsed: URL, linkText: string): boolean {
  const combined = `${parsed.pathname} ${parsed.search} ${linkText}`.toLowerCase().replace(/[_\s]+/g, "-");
  return unsafeLinkMarkers.some((marker) => combined.includes(marker));
}

function cleanText(value: string, maxLength: number): string {
  return String(value || "")
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, maxLength);
}

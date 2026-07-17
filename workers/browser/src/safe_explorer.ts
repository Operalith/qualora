import type { FindingInput } from "./findings";

export type SafeExplorerActionType = "link_navigation" | "button" | "form_get" | "form_post" | "input" | "unknown";
export type SafeExplorerSafety = "safe" | "unsafe" | "unsupported" | "unknown";
export type SafeExplorerDecision = "execute" | "skip";

export type ExtractedSafeExplorerAction = {
  actionType: SafeExplorerActionType;
  label: string;
  text: string;
  selectorHint: string;
  href: string;
  targetURL: string;
  method: string;
  fieldCount?: number;
  hasPasswordField?: boolean;
  hasFileField?: boolean;
  hasHiddenSensitiveField?: boolean;
};

export type ClassifiedSafeExplorerAction = ExtractedSafeExplorerAction & {
  normalizedURL: string;
  sameOrigin: boolean;
  safety: SafeExplorerSafety;
  decision: SafeExplorerDecision;
  skipReason: string;
};

export type SafeExplorerPolicy = {
  sourceURL: string;
  frontendURL: string;
  allowedHosts: string[];
  sameOriginOnly: boolean;
  allowGetForms: boolean;
};

export type SafeExplorerPageFindingInput = {
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

const dangerousActionMarkers = [
  "admin-delete",
  "admin-remove",
  "admin-destroy",
  "admin-reset",
  "admin-mutation",
  "cancel",
  "delete",
  "destroy",
  "logout",
  "payment",
  "refund",
  "remove",
  "reset",
  "transfer",
  "withdraw"
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

export function normalizeSafeExplorerURL(raw: string, baseURL?: string): string {
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

export function classifySafeExplorerAction(action: ExtractedSafeExplorerAction, policy: SafeExplorerPolicy): ClassifiedSafeExplorerAction {
  const cleaned = cleanAction(action);
  const target = cleaned.targetURL || cleaned.href;
  if (cleaned.actionType === "button" || cleaned.actionType === "input" || cleaned.actionType === "unknown") {
    return unsupported(cleaned, target, policy, looksDangerous(cleaned) ? "unsafe_action_skipped" : "unsupported_action", looksDangerous(cleaned) ? "unsafe" : "unsupported");
  }
  if (cleaned.actionType === "form_post" || (cleaned.method && !["", "get", "head", "options"].includes(cleaned.method.toLowerCase()))) {
    return unsupported(cleaned, target, policy, "form_method_not_safe", "unsafe");
  }
  if (cleaned.actionType === "form_get" && !policy.allowGetForms) {
    return unsupported(cleaned, target, policy, "get_forms_disabled", "unsupported");
  }
  if (cleaned.hasPasswordField || cleaned.hasFileField || cleaned.hasHiddenSensitiveField) {
    return unsupported(cleaned, target, policy, "unsafe_action_skipped", "unsafe");
  }
  if (!target) {
    return unsupported(cleaned, target, policy, "empty_target", "unsupported");
  }

  let parsed: URL;
  try {
    parsed = new URL(target, policy.sourceURL);
  } catch {
    return unsupported(cleaned, target, policy, "invalid_url", "unsupported");
  }

  if (!["http:", "https:"].includes(parsed.protocol)) {
    return skipped(cleaned, "", false, "unsupported_action", "unsupported");
  }

  const normalizedURL = normalizeSafeExplorerURL(parsed.toString());
  const sameOrigin = safeSameOrigin(parsed, policy.frontendURL);
  const hostAllowed = policy.allowedHosts.some((allowedHost) => hostMatches(parsed.hostname, allowedHost));

  if (policy.sameOriginOnly && !sameOrigin) {
    return skipped(cleaned, normalizedURL, sameOrigin, "external_action_skipped", "unsafe");
  }
  if (!hostAllowed) {
    return skipped(cleaned, normalizedURL, sameOrigin, "host_not_allowed", "unsafe");
  }
  if (hasSensitiveQuery(parsed)) {
    return skipped(cleaned, normalizedURL, sameOrigin, "sensitive_query_skipped", "unsafe");
  }
  if (looksLikeDownload(parsed)) {
    return skipped(cleaned, normalizedURL, sameOrigin, "non_html_resource", "unsupported");
  }
  if (looksDangerous(cleaned, parsed)) {
    return skipped(cleaned, normalizedURL, sameOrigin, "unsafe_action_skipped", "unsafe");
  }

  return {
    ...cleaned,
    targetURL: normalizedURL,
    normalizedURL,
    sameOrigin,
    safety: "safe",
    decision: "execute",
    skipReason: ""
  };
}

export function markSafeExplorerDuplicate(action: ClassifiedSafeExplorerAction): ClassifiedSafeExplorerAction {
  return {
    ...action,
    decision: "skip",
    skipReason: "duplicate_url"
  };
}

export function buildSafeExplorerPageFindings(input: SafeExplorerPageFindingInput): FindingInput[] {
  const findings: FindingInput[] = [];
  if (input.loadError) {
    findings.push({
      title: "Safe Explorer navigation failed",
      severity: "high",
      category: "explorer_navigation_failure",
      confidence: "high",
      description: [`Summary: Safe Explorer could not load ${input.url}: ${input.loadError}`, `Steps to reproduce: open ${input.url} from the worker network and verify it renders without unsupported interaction.`].join("\n"),
      recommendation: "Verify the target route is reachable, in scope, and does not require a flow Safe Explorer intentionally skips.",
      evidenceIds: input.evidenceIds
    });
  } else if (input.statusCode !== null && input.statusCode >= 500) {
    findings.push({
      title: "Safe Explorer observed a server error",
      severity: "high",
      category: "explorer_navigation_failure",
      confidence: "high",
      description: [`Summary: ${input.url} returned HTTP ${input.statusCode}.`, `Steps to reproduce: open ${input.url} and inspect the initial document response.`].join("\n"),
      recommendation: "Inspect the frontend service and upstream dependencies for server-side failures.",
      evidenceIds: input.evidenceIds
    });
  }

  if (input.consoleErrorCount > 0) {
    findings.push({
      title: "Console errors observed during Safe Explorer run",
      severity: "medium",
      category: "explorer_console_error",
      confidence: "medium",
      description: [`Summary: ${input.consoleErrorCount} console error(s) were observed while loading ${input.url}.`, `Steps to reproduce: open ${input.url} and inspect console output during page load.`].join("\n"),
      recommendation: "Review uncaught frontend exceptions and client-side initialization failures.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.failedRequestCount > 0) {
    findings.push({
      title: "Failed network requests observed during Safe Explorer run",
      severity: "medium",
      category: "explorer_network_failure",
      confidence: "medium",
      description: [`Summary: ${input.failedRequestCount} failed network request(s) were observed while loading ${input.url}.`, `Steps to reproduce: open ${input.url} and inspect failed requests in browser developer tools.`].join("\n"),
      recommendation: "Ensure required first-party assets and APIs are reachable during page load.",
      evidenceIds: input.evidenceIds
    });
  }
  if (!input.loadError && input.statusCode !== null && input.statusCode < 400 && input.bodyTextLength === 0) {
    findings.push({
      title: "Safe Explorer observed an empty page",
      severity: "medium",
      category: "explorer_empty_page",
      confidence: "medium",
      description: [`Summary: ${input.url} loaded successfully but had no visible body text.`, `Steps to reproduce: open ${input.url} and inspect the rendered page body.`].join("\n"),
      recommendation: "Confirm the route renders meaningful visible content after initial load.",
      evidenceIds: input.evidenceIds
    });
  }
  return findings;
}

export function buildSafeExplorerActionFinding(action: ClassifiedSafeExplorerAction, evidenceIds: string[]): FindingInput | null {
  if (action.decision !== "skip" || !action.skipReason) {
    return null;
  }
  const target = action.normalizedURL || action.targetURL || action.href || action.label;
  if (action.skipReason === "duplicate_url") {
    return {
      title: "Duplicate Safe Explorer target skipped",
      severity: "info",
      category: "explorer_duplicate_page",
      confidence: "high",
      description: [`Summary: Safe Explorer skipped a duplicate target: ${target}.`, "Steps to reproduce: review the action timeline and compare normalized URLs."].join("\n"),
      recommendation: "No action is required unless this route should appear as a distinct canonical URL.",
      evidenceIds
    };
  }
  if (action.skipReason === "external_action_skipped" || action.skipReason === "host_not_allowed") {
    return {
      title: "Out-of-scope action skipped by Safe Explorer",
      severity: "info",
      category: "explorer_external_action_skipped",
      confidence: "high",
      description: [`Summary: Safe Explorer skipped an out-of-scope action targeting ${target}.`, "Steps to reproduce: review the source page action list and compare the target to allowed_hosts and same_origin_only."].join("\n"),
      recommendation: "Add required first-party hosts to allowed_hosts only when they are explicitly in testing scope.",
      evidenceIds
    };
  }
  if (action.safety === "unsafe") {
    return {
      title: "Unsafe action skipped by Safe Explorer",
      severity: "low",
      category: "explorer_unsafe_action_skipped",
      confidence: "medium",
      description: [`Summary: Safe Explorer skipped an action that looked mutating, sensitive, or out of policy: ${target}.`, "Steps to reproduce: inspect the recorded action label, method, target, and skip reason."].join("\n"),
      recommendation: "Keep destructive or sensitive actions out of automated safe exploration paths, or model them later with explicit non-destructive fixtures.",
      evidenceIds
    };
  }
  if (action.skipReason === "get_forms_disabled" || action.skipReason === "max_depth_or_steps_reached") {
    return {
      title: "Action skipped by Safe Explorer policy",
      severity: "info",
      category: "explorer_policy_blocked",
      confidence: "high",
      description: [`Summary: Safe Explorer observed ${target} but policy prevented execution: ${action.skipReason}.`, "Steps to reproduce: review the run settings, action metadata, and policy limits."].join("\n"),
      recommendation: "Adjust Safe Explorer limits only when the route remains non-destructive, in scope, and safe to navigate.",
      evidenceIds
    };
  }
  return {
    title: "Unsupported action skipped by Safe Explorer",
    severity: "info",
    category: "explorer_unsupported_action",
    confidence: "high",
    description: [`Summary: Safe Explorer observed but did not execute unsupported action ${action.label || action.actionType}.`, `Skip reason: ${action.skipReason}.`].join("\n"),
    recommendation: "Use deterministic links for safe navigation coverage, or wait for future explicit support for this interaction type.",
    evidenceIds
  };
}

function cleanAction(action: ExtractedSafeExplorerAction): ExtractedSafeExplorerAction {
  return {
    actionType: action.actionType || "unknown",
    label: cleanText(action.label, 180),
    text: cleanText(action.text, 240),
    selectorHint: cleanText(action.selectorHint, 240),
    href: cleanText(action.href, 1000),
    targetURL: cleanText(action.targetURL, 1000),
    method: cleanText(action.method || "", 16).toLowerCase(),
    fieldCount: Math.max(0, Number(action.fieldCount || 0)),
    hasPasswordField: Boolean(action.hasPasswordField),
    hasFileField: Boolean(action.hasFileField),
    hasHiddenSensitiveField: Boolean(action.hasHiddenSensitiveField)
  };
}

function unsupported(
  action: ExtractedSafeExplorerAction,
  target: string,
  policy: SafeExplorerPolicy,
  skipReason: string,
  safety: SafeExplorerSafety
): ClassifiedSafeExplorerAction {
  let sameOrigin = false;
  let normalizedURL = "";
  if (target) {
    try {
      const parsed = new URL(target, policy.sourceURL);
      if (["http:", "https:"].includes(parsed.protocol)) {
        sameOrigin = safeSameOrigin(parsed, policy.frontendURL);
        normalizedURL = normalizeSafeExplorerURL(parsed.toString());
      }
    } catch {
      normalizedURL = "";
    }
  }
  return skipped(action, normalizedURL, sameOrigin, skipReason, safety);
}

function skipped(
  action: ExtractedSafeExplorerAction,
  normalizedURL: string,
  sameOrigin: boolean,
  skipReason: string,
  safety: SafeExplorerSafety
): ClassifiedSafeExplorerAction {
  return {
    ...action,
    targetURL: normalizedURL || action.targetURL || action.href,
    normalizedURL,
    sameOrigin,
    safety,
    decision: "skip",
    skipReason
  };
}

function isSensitiveQueryName(name: string): boolean {
  const normalized = name.toLowerCase().trim();
  return sensitiveQueryMarkers.some((marker) => normalized === marker || normalized.includes(marker));
}

function hasSensitiveQuery(parsed: URL): boolean {
  return Array.from(parsed.searchParams.keys()).some(isSensitiveQueryName);
}

function hostMatches(host: string, allowedHost: string): boolean {
  const normalizedHost = host.toLowerCase().replace(/\.$/, "");
  const allowed = allowedHost.toLowerCase().trim().replace(/\.$/, "");
  if (normalizedHost === allowed) {
    return true;
  }
  return allowed.startsWith("*.") && normalizedHost.endsWith(allowed.slice(1));
}

function safeSameOrigin(parsed: URL, frontendURL: string): boolean {
  try {
    return parsed.origin === new URL(frontendURL).origin;
  } catch {
    return false;
  }
}

function looksLikeDownload(parsed: URL): boolean {
  const path = parsed.pathname.toLowerCase();
  return downloadExtensions.some((extension) => path.endsWith(extension));
}

function looksDangerous(action: ExtractedSafeExplorerAction, parsed?: URL): boolean {
  const combined = `${parsed?.pathname || ""} ${parsed?.search || ""} ${action.href} ${action.targetURL} ${action.label} ${action.text}`
    .toLowerCase()
    .replace(/[_\s/]+/g, "-");
  return dangerousActionMarkers.some((marker) => combined.includes(marker));
}

function cleanText(value: string, maxLength: number): string {
  return String(value || "")
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, maxLength);
}

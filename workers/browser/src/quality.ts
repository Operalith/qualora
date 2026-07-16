export type QualityCategory = "security" | "accessibility" | "performance";
export type QualitySeverity = "critical" | "high" | "medium" | "low" | "info";

export type QualityCheckOptions = {
  includeSecurity: boolean;
  includeAccessibility: boolean;
  includePerformance: boolean;
};

export type QualityCookie = {
  name: string;
  secure: boolean;
  httpOnly: boolean;
  sameSite: string;
};

export type QualityResource = {
  url: string;
  status: number | null;
  contentType: string;
  resourceType: string;
  contentLength: number | null;
};

export type QualityPageSnapshot = {
  targetURL: string;
  finalURL: string;
  title: string;
  statusCode: number | null;
  headers: Record<string, string>;
  isHTTPS: boolean;
  bodyTextLength: number | null;
  loadDurationMS: number | null;
  loadError: string;
  consoleErrorCount: number;
  failedRequestCount: number;
  blockedRequestCount: number;
  cookies: QualityCookie[];
  resources: QualityResource[];
  forms: {
    total: number;
    passwordForms: number;
    passwordAutocompleteIssues: number;
    insecureActions: number;
  };
  accessibility: {
    hasTitle: boolean;
    htmlLang: string;
    hasMainLandmark: boolean;
    imagesTotal: number;
    imagesMissingAlt: number;
    inputsTotal: number;
    inputsMissingLabels: number;
    buttonsTotal: number;
    buttonsMissingNames: number;
    linksTotal: number;
    linksMissingNames: number;
    imagesMissingDimensions: number;
  };
};

export type QualityResult = {
  category: QualityCategory;
  ruleID: string;
  severity: QualitySeverity;
  title: string;
  description: string;
  recommendation: string;
  url: string;
  evidence: Record<string, unknown>;
};

export function buildQualityResults(snapshot: QualityPageSnapshot, options: QualityCheckOptions): QualityResult[] {
  const results: QualityResult[] = [];
  if (options.includeSecurity) {
    results.push(...buildSecurityResults(snapshot));
  }
  if (options.includeAccessibility) {
    results.push(...buildAccessibilityResults(snapshot));
  }
  if (options.includePerformance) {
    results.push(...buildPerformanceResults(snapshot));
  }
  return results;
}

function buildSecurityResults(snapshot: QualityPageSnapshot): QualityResult[] {
  const results: QualityResult[] = [];
  const header = (name: string) => snapshot.headers[name.toLowerCase()] || "";
  const csp = header("content-security-policy");
  if (!csp) {
    results.push(result(snapshot, "security", "missing_csp", "medium", "Content Security Policy header is missing", "The page response did not include a Content-Security-Policy header.", "Add a restrictive Content-Security-Policy header appropriate for the application.", { header: "content-security-policy" }));
  }
  if (!header("x-frame-options") && !csp.toLowerCase().includes("frame-ancestors")) {
    results.push(result(snapshot, "security", "missing_frame_protection", "medium", "Clickjacking protection is not obvious", "The page did not include X-Frame-Options and CSP frame-ancestors was not observed.", "Set CSP frame-ancestors or X-Frame-Options to restrict where the page can be framed.", { headers_checked: ["x-frame-options", "content-security-policy"] }));
  }
  if (!header("x-content-type-options")) {
    results.push(result(snapshot, "security", "missing_x_content_type_options", "low", "X-Content-Type-Options header is missing", "The page response did not include X-Content-Type-Options.", "Set X-Content-Type-Options: nosniff on HTML and asset responses where practical.", { header: "x-content-type-options" }));
  }
  if (!header("referrer-policy")) {
    results.push(result(snapshot, "security", "missing_referrer_policy", "low", "Referrer-Policy header is missing", "The page response did not include a Referrer-Policy header.", "Set a Referrer-Policy that matches the product privacy and analytics requirements.", { header: "referrer-policy" }));
  }
  if (!header("permissions-policy")) {
    results.push(result(snapshot, "security", "missing_permissions_policy", "info", "Permissions-Policy header is missing", "The page response did not include a Permissions-Policy header.", "Consider setting Permissions-Policy to disable unused powerful browser features.", { header: "permissions-policy" }));
  }
  if (snapshot.isHTTPS && !header("strict-transport-security")) {
    results.push(result(snapshot, "security", "missing_hsts", "medium", "HSTS header is missing", "The HTTPS page response did not include Strict-Transport-Security.", "Enable HSTS after confirming HTTPS is deployed consistently for the host.", { header: "strict-transport-security" }));
  }
  for (const name of ["server", "x-powered-by"]) {
    const value = header(name);
    if (value) {
      results.push(result(snapshot, "security", `${name.replaceAll("-", "_")}_disclosure`, "low", "Server technology disclosure observed", `The response includes a ${name} header.`, "Remove or minimize technology-identifying headers where possible.", { header: name, value: truncate(value, 120) }));
    }
  }
  if (!snapshot.isHTTPS && snapshot.forms.passwordForms > 0) {
    results.push(result(snapshot, "security", "password_form_over_http", "high", "Password form is served over HTTP", "A visible password form was observed on a non-HTTPS page.", "Serve login and password collection pages over HTTPS only.", { password_forms: snapshot.forms.passwordForms }));
  }
  if (snapshot.forms.insecureActions > 0) {
    results.push(result(snapshot, "security", "insecure_form_action", "high", "Form posts to an insecure action", "One or more forms use an HTTP action from an HTTPS page or target.", "Use HTTPS form actions and keep credential flows same-origin where possible.", { insecure_form_actions: snapshot.forms.insecureActions }));
  }
  if (snapshot.forms.passwordAutocompleteIssues > 0) {
    results.push(result(snapshot, "security", "password_autocomplete_unconfigured", "info", "Password autocomplete hints are incomplete", "One or more password inputs are missing a useful autocomplete value.", "Set autocomplete values such as current-password or new-password for password fields.", { affected_password_fields: snapshot.forms.passwordAutocompleteIssues }));
  }
  if (hasSensitiveQuery(snapshot.finalURL || snapshot.targetURL)) {
    results.push(result(snapshot, "security", "sensitive_query_parameter", "medium", "Sensitive-looking query parameter observed", "The page URL contains a sensitive-looking query parameter name.", "Avoid putting tokens, passwords, keys, or session identifiers in URLs.", { url_query_redacted: true }));
  }
  const sourceMaps = snapshot.resources.filter((resource) => resource.url.endsWith(".map") && (resource.status === null || resource.status < 400));
  if (sourceMaps.length > 0) {
    results.push(result(snapshot, "security", "source_map_exposed", "low", "Source map file is publicly reachable", "One or more source map resources were requested successfully.", "Confirm source maps are intended for this environment or restrict them in production.", { count: sourceMaps.length, examples: sourceMaps.slice(0, 5).map((resource) => resource.url) }));
  }
  const mixed = snapshot.isHTTPS ? snapshot.resources.filter((resource) => resource.url.startsWith("http://")) : [];
  if (mixed.length > 0) {
    results.push(result(snapshot, "security", "mixed_content", "high", "Mixed content was observed", "The HTTPS page requested one or more HTTP resources.", "Serve page dependencies over HTTPS and remove insecure resource URLs.", { count: mixed.length, examples: mixed.slice(0, 5).map((resource) => resource.url) }));
  }
  const cookieIssues = snapshot.cookies.filter((cookie) => !cookie.httpOnly || (snapshot.isHTTPS && !cookie.secure) || !cookie.sameSite);
  if (cookieIssues.length > 0) {
    results.push(result(snapshot, "security", "cookie_flags_incomplete", "medium", "Cookie security flags are incomplete", "One or more visible cookie metadata records are missing HttpOnly, Secure, or SameSite attributes.", "Set HttpOnly, Secure, and SameSite attributes for session and sensitive cookies where appropriate.", { cookies: cookieIssues.slice(0, 10) }));
  }
  return results;
}

function buildAccessibilityResults(snapshot: QualityPageSnapshot): QualityResult[] {
  const results: QualityResult[] = [];
  const a11y = snapshot.accessibility;
  if (!a11y.hasTitle) {
    results.push(result(snapshot, "accessibility", "missing_title", "low", "Document title is missing", "The page did not expose a non-empty document title.", "Add a concise page title that identifies the view or workflow.", {}));
  }
  if (!a11y.htmlLang) {
    results.push(result(snapshot, "accessibility", "missing_html_lang", "low", "HTML lang attribute is missing", "The root HTML element does not declare a language.", "Set the lang attribute on the html element.", {}));
  }
  if (!a11y.hasMainLandmark) {
    results.push(result(snapshot, "accessibility", "missing_main_landmark", "info", "Main landmark is missing", "The page did not expose a main element or role=main landmark.", "Add a single main landmark around the primary page content.", {}));
  }
  if (a11y.imagesMissingAlt > 0) {
    results.push(result(snapshot, "accessibility", "images_missing_alt", "medium", "Images are missing alt text", "One or more visible images are missing alt text.", "Add meaningful alt text or mark decorative images with an empty alt attribute.", { total_images: a11y.imagesTotal, affected_images: a11y.imagesMissingAlt }));
  }
  if (a11y.inputsMissingLabels > 0) {
    results.push(result(snapshot, "accessibility", "inputs_missing_labels", "medium", "Form controls are missing labels", "One or more inputs, selects, or textareas are missing accessible labels.", "Associate labels with form controls using label, aria-label, or aria-labelledby.", { total_inputs: a11y.inputsTotal, affected_inputs: a11y.inputsMissingLabels }));
  }
  if (a11y.buttonsMissingNames > 0) {
    results.push(result(snapshot, "accessibility", "buttons_missing_names", "medium", "Buttons are missing accessible names", "One or more buttons do not expose text or an ARIA label.", "Give icon-only and empty buttons a clear accessible name.", { total_buttons: a11y.buttonsTotal, affected_buttons: a11y.buttonsMissingNames }));
  }
  if (a11y.linksMissingNames > 0) {
    results.push(result(snapshot, "accessibility", "links_missing_names", "low", "Links are missing accessible text", "One or more links do not expose visible text or an ARIA label.", "Give links descriptive text or accessible labels.", { total_links: a11y.linksTotal, affected_links: a11y.linksMissingNames }));
  }
  return results;
}

function buildPerformanceResults(snapshot: QualityPageSnapshot): QualityResult[] {
  const results: QualityResult[] = [];
  if (snapshot.loadError) {
    results.push(result(snapshot, "performance", "page_load_failed", "high", "Page navigation failed", "The browser could not complete navigation for the checked page.", "Verify the route is reachable from the worker network and allowed by project scope.", { load_error: snapshot.loadError }));
  }
  if (snapshot.statusCode !== null && snapshot.statusCode >= 500) {
    results.push(result(snapshot, "performance", "document_server_error", "high", "Document returned a server error", `The initial document returned HTTP ${snapshot.statusCode}.`, "Inspect the frontend route and upstream dependencies for server-side failures.", { status_code: snapshot.statusCode }));
  } else if (snapshot.statusCode !== null && snapshot.statusCode >= 400) {
    results.push(result(snapshot, "performance", "document_client_error", "medium", "Document returned an error status", `The initial document returned HTTP ${snapshot.statusCode}.`, "Verify the route exists and does not require unsupported setup before initial load.", { status_code: snapshot.statusCode }));
  }
  if (snapshot.loadDurationMS !== null && snapshot.loadDurationMS > 7000) {
    results.push(result(snapshot, "performance", "page_load_very_slow", "high", "Page load was very slow", `The page took ${snapshot.loadDurationMS}ms to reach the observed load point.`, "Profile backend response time, asset loading, and client-side initialization.", { load_duration_ms: snapshot.loadDurationMS }));
  } else if (snapshot.loadDurationMS !== null && snapshot.loadDurationMS > 3000) {
    results.push(result(snapshot, "performance", "page_load_slow", "medium", "Page load was slow", `The page took ${snapshot.loadDurationMS}ms to reach the observed load point.`, "Review large assets, blocking scripts, and slow first-party dependencies.", { load_duration_ms: snapshot.loadDurationMS }));
  }
  if (snapshot.consoleErrorCount > 0) {
    results.push(result(snapshot, "performance", "console_errors", "medium", "Console errors were observed", `${snapshot.consoleErrorCount} console error(s) were observed while loading the page.`, "Fix uncaught frontend exceptions and failed client-side initialization.", { console_error_count: snapshot.consoleErrorCount }));
  }
  if (snapshot.failedRequestCount > 0) {
    results.push(result(snapshot, "performance", "failed_network_requests", "medium", "Failed network requests were observed", `${snapshot.failedRequestCount} request(s) failed while loading the page.`, "Ensure required first-party assets and APIs are available from the worker network.", { failed_request_count: snapshot.failedRequestCount }));
  }
  const failedResources = snapshot.resources.filter((resource) => resource.status !== null && resource.status >= 400);
  if (failedResources.length > 0) {
    results.push(result(snapshot, "performance", "failed_resource_status", "medium", "Page resources returned error statuses", "One or more page resources returned HTTP 4xx or 5xx statuses.", "Fix missing or failing first-party assets and API calls used during initial page load.", { count: failedResources.length, examples: failedResources.slice(0, 5).map((resource) => ({ url: resource.url, status: resource.status, type: resource.resourceType })) }));
  }
  if (snapshot.blockedRequestCount > 0) {
    results.push(result(snapshot, "performance", "out_of_scope_requests_blocked", "info", "Out-of-scope requests were blocked", `${snapshot.blockedRequestCount} request(s) were blocked by allowed-host enforcement.`, "Review third-party dependencies and add only intentional hosts to allowed_hosts.", { blocked_request_count: snapshot.blockedRequestCount }));
  }
  const scripts = snapshot.resources.filter((resource) => resource.resourceType === "script" && (resource.contentLength ?? 0) > 250_000);
  if (scripts.length > 0) {
    results.push(result(snapshot, "performance", "large_javascript_bundle", "low", "Large JavaScript resource observed", "One or more JavaScript resources exceeded 250KB by Content-Length.", "Split, compress, or lazy-load large JavaScript bundles where practical.", { count: scripts.length, examples: scripts.slice(0, 5).map((resource) => ({ url: resource.url, bytes: resource.contentLength })) }));
  }
  if (snapshot.resources.length > 50) {
    results.push(result(snapshot, "performance", "many_network_requests", "low", "Many network requests observed", `The page loaded ${snapshot.resources.length} observed resources.`, "Review request count and defer non-critical resources.", { resource_count: snapshot.resources.length }));
  }
  if (snapshot.accessibility.imagesMissingDimensions > 0) {
    results.push(result(snapshot, "performance", "images_missing_dimensions", "info", "Images are missing dimensions", "One or more visible images do not include width and height attributes.", "Set image dimensions or use layout-reserving CSS to reduce layout shifts.", { affected_images: snapshot.accessibility.imagesMissingDimensions }));
  }
  if (snapshot.statusCode !== null && snapshot.statusCode < 400 && snapshot.bodyTextLength === 0) {
    results.push(result(snapshot, "performance", "empty_body", "low", "Page body appears empty", "The page loaded successfully but visible body text was empty.", "Verify client-side rendering and loading states for the route.", { body_text_length: snapshot.bodyTextLength }));
  }
  return results;
}

function result(
  snapshot: QualityPageSnapshot,
  category: QualityCategory,
  ruleID: string,
  severity: QualitySeverity,
  title: string,
  description: string,
  recommendation: string,
  evidence: Record<string, unknown>
): QualityResult {
  return {
    category,
    ruleID,
    severity,
    title,
    description,
    recommendation,
    url: snapshot.finalURL || snapshot.targetURL,
    evidence: {
      target_url: snapshot.targetURL,
      final_url: snapshot.finalURL,
      page_title: snapshot.title,
      status_code: snapshot.statusCode,
      ...evidence
    }
  };
}

function hasSensitiveQuery(raw: string): boolean {
  try {
    const parsed = new URL(raw);
    for (const name of parsed.searchParams.keys()) {
      const normalized = name.toLowerCase();
      if (["token", "password", "secret", "api_key", "apikey", "session", "auth", "authorization", "jwt", "key"].some((marker) => normalized.includes(marker))) {
        return true;
      }
    }
  } catch {
    return false;
  }
  return false;
}

function truncate(value: string, max: number): string {
  return value.length > max ? `${value.slice(0, max)}...` : value;
}

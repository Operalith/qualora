import type { Page } from "playwright";
import type { FindingInput } from "./findings";

export type SafeFormField = {
  name: string;
  type: string;
  label: string;
  placeholder: string;
  required: boolean;
  hidden: boolean;
  options: string[];
};

export type ExtractedFormCandidate = {
  selectorHint: string;
  pageURL: string;
  actionURL: string;
  method: string;
  label: string;
  fieldCount: number;
  passwordFieldCount: number;
  fileFieldCount: number;
  hiddenSensitiveFieldCount: number;
  submitButtonCount: number;
  fields: SafeFormField[];
};

export type FormTestPolicy = {
  sourceURL: string;
  frontendURL: string;
  allowedHosts: string[];
  sameOriginOnly: boolean;
  safeGetOnly: boolean;
};

export type ClassifiedFormCandidate = ExtractedFormCandidate & {
  normalizedActionURL: string;
  sameOrigin: boolean;
  classification:
    | "search"
    | "filter"
    | "sort"
    | "navigation"
    | "newsletter"
    | "contact"
    | "login"
    | "password"
    | "payment"
    | "profile_update"
    | "upload"
    | "admin_mutation"
    | "destructive"
    | "unknown";
  safety: "safe" | "unsafe" | "unsupported" | "unknown";
  decision: "test" | "skip";
  skipReason: string;
  testValues: Record<string, string>;
};

export type FormExecutionFindingInput = {
  result: ClassifiedFormCandidate;
  submittedURL: string;
  finalURL: string;
  statusCode: number | null;
  loadError: string;
  bodyTextLength: number | null;
  consoleErrorCount: number;
  failedRequestCount: number;
  evidenceIds: string[];
};

const dangerousFormMarkers = [
  "admin",
  "billing",
  "card",
  "cart",
  "checkout",
  "confirm",
  "contact",
  "delete",
  "destroy",
  "download",
  "invoice/pay",
  "login",
  "logout",
  "mutate",
  "order",
  "password",
  "payment",
  "profile",
  "purchase",
  "refund",
  "remove",
  "reset",
  "signup",
  "subscribe",
  "transfer",
  "upload",
  "withdraw"
];

const sensitiveFieldMarkers = [
  "access_token",
  "amount",
  "api_key",
  "apikey",
  "auth",
  "authorization",
  "card",
  "cookie",
  "credential",
  "csrf",
  "cvv",
  "email",
  "jwt",
  "mfa",
  "otp",
  "password",
  "passwd",
  "phone",
  "secret",
  "session",
  "ssn",
  "token",
  "username"
];

const safeFieldNames = /^(q|query|search|term|keyword|keywords|filter|sort|order|category|tag|type|status|page|limit|view)$/i;
const safeTextFieldTypes = new Set(["", "text", "search"]);
const safeChoiceFieldTypes = new Set(["select-one", "select-multiple", "checkbox", "radio"]);

export async function extractFormCandidates(page: Page): Promise<ExtractedFormCandidate[]> {
  return page.evaluate(() => {
    const clean = (value: string | null | undefined, max = 240) => String(value || "").replace(/\s+/g, " ").trim().slice(0, max);
    const visible = (element: Element) => {
      const style = window.getComputedStyle(element);
      return style.visibility !== "hidden" && style.display !== "none" && element.getClientRects().length > 0;
    };
    const selectorHint = (element: Element, index: number): string => {
      const tag = element.tagName.toLowerCase();
      const id = element.getAttribute("id");
      if (id) {
        return `${tag}#${id}`.slice(0, 240);
      }
      const name = element.getAttribute("name");
      if (name) {
        return `${tag}[name="${name}"]`.slice(0, 240);
      }
      const aria = element.getAttribute("aria-label");
      if (aria) {
        return `${tag}[aria-label="${clean(aria, 80)}"]`.slice(0, 240);
      }
      return `${tag}:nth-of-type(${index + 1})`;
    };
    const labelFor = (field: HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement): string => {
      const id = field.getAttribute("id");
      if (id) {
        const explicit = document.querySelector(`label[for="${CSS.escape(id)}"]`);
        if (explicit?.textContent) {
          return clean(explicit.textContent, 160);
        }
      }
      const wrapping = field.closest("label");
      if (wrapping?.textContent) {
        return clean(wrapping.textContent, 160);
      }
      return clean(field.getAttribute("aria-label") || field.getAttribute("placeholder") || field.getAttribute("name") || "", 160);
    };
    const fieldName = (field: HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement, index: number): string => {
      return clean(field.getAttribute("name") || field.getAttribute("id") || `field_${index + 1}`, 120);
    };
    const fieldType = (field: HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement): string => {
      if (field instanceof HTMLSelectElement) {
        return field.multiple ? "select-multiple" : "select-one";
      }
      if (field instanceof HTMLTextAreaElement) {
        return "textarea";
      }
      return clean(field.getAttribute("type") || "text", 40).toLowerCase();
    };
    const fieldOptions = (field: HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement): string[] => {
      if (field instanceof HTMLSelectElement) {
        return Array.from(field.options)
          .map((option) => clean(option.value || option.textContent, 80))
          .filter(Boolean)
          .slice(0, 12);
      }
      if (field instanceof HTMLInputElement && (field.type === "checkbox" || field.type === "radio")) {
        return [clean(field.value || "on", 80)];
      }
      return [];
    };
    const isSensitiveHidden = (field: SafeFormField) => {
      return field.hidden && /token|secret|password|auth|session|csrf|key/.test(`${field.name} ${field.label}`.toLowerCase());
    };

    return Array.from(document.querySelectorAll<HTMLFormElement>("form"))
      .filter(visible)
      .map((form, formIndex) => {
        const fields = Array.from(form.querySelectorAll<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>("input, select, textarea"))
          .filter((field) => {
            const type = fieldType(field);
            return type !== "submit" && type !== "button" && type !== "reset" && type !== "image";
          })
          .slice(0, 30)
          .map((field, fieldIndex) => {
            const type = fieldType(field);
            return {
              name: fieldName(field, fieldIndex),
              type,
              label: labelFor(field),
              placeholder: clean(field.getAttribute("placeholder"), 160),
              required: Boolean((field as HTMLInputElement).required),
              hidden: type === "hidden" || !visible(field),
              options: fieldOptions(field)
            };
          });
        const labels = fields
          .map((field) => field.label || field.name)
          .filter(Boolean)
          .slice(0, 3)
          .join(", ");
        const method = clean(form.getAttribute("method") || "GET", 16).toUpperCase();
        const label = clean(form.getAttribute("aria-label") || form.getAttribute("name") || labels || "Form", 180);
        const passwordFieldCount = fields.filter((field) => field.type === "password").length;
        const fileFieldCount = fields.filter((field) => field.type === "file").length;
        return {
          selectorHint: selectorHint(form, formIndex),
          pageURL: clean(window.location.href, 600),
          actionURL: clean(form.action || form.getAttribute("action") || window.location.href, 600),
          method,
          label,
          fieldCount: fields.length,
          passwordFieldCount,
          fileFieldCount,
          hiddenSensitiveFieldCount: fields.filter(isSensitiveHidden).length,
          submitButtonCount: form.querySelectorAll("button, input[type='submit'], input[type='image']").length,
          fields
        };
      })
      .slice(0, 80);
  });
}

export function classifyFormCandidate(form: ExtractedFormCandidate, policy: FormTestPolicy): ClassifiedFormCandidate {
  const cleaned = cleanFormCandidate(form);
  const method = cleaned.method.toLowerCase() || "get";
  const normalizedActionURL = normalizeFormURL(cleaned.actionURL || policy.sourceURL, policy.sourceURL);
  let parsed: URL;
  try {
    parsed = new URL(normalizedActionURL);
  } catch {
    return skipped(cleaned, "", false, "unknown", "unsupported", "invalid_action_url");
  }

  const sameOrigin = sameOriginURL(parsed, policy.frontendURL || policy.sourceURL);
  const hostAllowed = policy.allowedHosts.some((allowedHost) => hostMatches(parsed.hostname, allowedHost));
  const classification = classifyFormPurpose(cleaned, parsed);
  const values = buildSafeTestValues(cleaned, classification);
  const sensitiveField = cleaned.fields.find((field) => isSensitiveField(field));

  if (policy.safeGetOnly && method !== "get") {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsafe", "form_method_not_safe");
  }
  if (method !== "get") {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsafe", "form_method_not_safe");
  }
  if (!["http:", "https:"].includes(parsed.protocol)) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsupported", "unsupported_scheme");
  }
  if (policy.sameOriginOnly && !sameOrigin) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsafe", "external_action_skipped");
  }
  if (!hostAllowed) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsafe", "host_not_allowed");
  }
  if (cleaned.passwordFieldCount > 0 || cleaned.fileFieldCount > 0 || cleaned.hiddenSensitiveFieldCount > 0 || sensitiveField) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsafe", "sensitive_form_skipped");
  }
  if (looksDangerousForm(cleaned, parsed)) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsafe", "unsafe_form_skipped");
  }
  if (hasSensitiveQuery(parsed)) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsafe", "sensitive_query_skipped");
  }
  if (!["search", "filter", "sort", "navigation"].includes(classification)) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsupported", "classification_not_safe");
  }
  if (cleaned.fieldCount > 0 && Object.keys(values).length === 0) {
    return skipped(cleaned, normalizedActionURL, sameOrigin, classification, "unsupported", "no_safe_test_values");
  }
  return {
    ...cleaned,
    normalizedActionURL,
    sameOrigin,
    classification,
    safety: "safe",
    decision: "test",
    skipReason: "",
    testValues: values
  };
}

export function buildSubmittedFormURL(form: ClassifiedFormCandidate): string {
  const target = new URL(form.normalizedActionURL || form.actionURL, form.pageURL);
  for (const [name, value] of Object.entries(form.testValues)) {
    if (!name || isSensitiveName(name) || looksDangerousValue(value)) {
      continue;
    }
    target.searchParams.set(name, value);
  }
  target.searchParams.sort();
  return redactSensitiveFormURL(target).toString();
}

export function formValuesSummary(form: ClassifiedFormCandidate): Record<string, unknown> {
  return {
    classification: form.classification,
    fields_populated: Object.keys(form.testValues).length,
    values_are_synthetic: true,
    raw_values_stored: false,
    fields: Object.keys(form.testValues)
      .sort()
      .slice(0, 12)
      .map((name) => ({
        name,
        value_kind: "bounded_benign_value"
      }))
  };
}

export function buildFormTestFindings(input: FormExecutionFindingInput): FindingInput[] {
  const findings: FindingInput[] = [];
  const target = input.submittedURL || input.result.normalizedActionURL || input.result.pageURL;
  if (input.loadError) {
    findings.push({
      title: "Safe GET form submission failed",
      severity: "high",
      category: "form_submission_failure",
      confidence: "high",
      description: [`Summary: Safe Form Testing could not load the submitted GET form result for ${target}: ${input.loadError}.`, `Steps to reproduce: open ${input.result.pageURL}, submit the ${input.result.classification} form with benign values, and inspect the result page.`].join("\n"),
      recommendation: "Verify the form action route is reachable and handles benign query parameters without browser navigation failures.",
      evidenceIds: input.evidenceIds
    });
  } else if (input.statusCode !== null && input.statusCode >= 500) {
    findings.push({
      title: "Safe GET form returned a server error",
      severity: "high",
      category: "form_server_error",
      confidence: "high",
      description: [`Summary: Safe GET form result ${target} returned HTTP ${input.statusCode}.`, `Steps to reproduce: submit the recorded safe GET form with the same synthetic benign values.`].join("\n"),
      recommendation: "Inspect the form handler and upstream services for unhandled query-parameter errors.",
      evidenceIds: input.evidenceIds
    });
  } else if (input.statusCode !== null && input.statusCode >= 400) {
    findings.push({
      title: "Safe GET form returned a client error",
      severity: "medium",
      category: "form_client_error",
      confidence: "medium",
      description: [`Summary: Safe GET form result ${target} returned HTTP ${input.statusCode}.`, `Steps to reproduce: submit the recorded safe GET form with the same synthetic benign values.`].join("\n"),
      recommendation: "Verify the form action, supported query parameters, and routing behavior.",
      evidenceIds: input.evidenceIds
    });
  }
  if (!input.loadError && input.statusCode !== null && input.statusCode < 400 && (input.bodyTextLength ?? 0) < 20) {
    findings.push({
      title: "Safe GET form result page looked empty",
      severity: "low",
      category: "form_low_value_result",
      confidence: "medium",
      description: [`Summary: Safe GET form result ${target} loaded but had very little visible text.`, "Steps to reproduce: submit the recorded safe GET form and inspect the rendered result page."].join("\n"),
      recommendation: "Confirm that the form result page renders useful content for benign search/filter inputs.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.consoleErrorCount > 0) {
    findings.push({
      title: "Console errors observed after safe form submission",
      severity: "medium",
      category: "form_console_error",
      confidence: "medium",
      description: `${input.consoleErrorCount} console error(s) were observed after submitting a safe GET form.`,
      recommendation: "Review frontend runtime errors on the form result page.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.failedRequestCount > 0) {
    findings.push({
      title: "Network failures observed after safe form submission",
      severity: "medium",
      category: "form_network_failure",
      confidence: "medium",
      description: `${input.failedRequestCount} failed network request(s) were observed after submitting a safe GET form.`,
      recommendation: "Inspect failed first-party resources and API calls on the form result page.",
      evidenceIds: input.evidenceIds
    });
  }
  return findings;
}

export function buildFormSkipFinding(result: ClassifiedFormCandidate, evidenceIds: string[]): FindingInput | null {
  if (result.decision !== "skip") {
    return null;
  }
  if (result.safety === "unsafe") {
    return {
      title: "Unsafe form skipped by Safe Form Testing",
      severity: "low",
      category: "form_unsafe_skipped",
      confidence: "high",
      description: [`Summary: Qualora skipped a form on ${result.pageURL} because ${result.skipReason}.`, "Steps to reproduce: review the form action, method, fields, and safety decision in the evidence metadata."].join("\n"),
      recommendation: "Keep destructive, sensitive, credential, payment, upload, and mutating forms out of safe automated testing paths unless a future explicit non-destructive fixture policy is configured.",
      evidenceIds
    };
  }
  if (result.safety === "unsupported") {
    return {
      title: "Unsupported form skipped by Safe Form Testing",
      severity: "info",
      category: "form_unsupported_skipped",
      confidence: "high",
      description: [`Summary: Qualora skipped a form on ${result.pageURL} because ${result.skipReason}.`, "Steps to reproduce: review the form classification and supported safe GET form policy."].join("\n"),
      recommendation: "No action is required unless this form should be modeled as a safe same-origin GET search, filter, sort, or navigation flow.",
      evidenceIds
    };
  }
  return null;
}

export function buildNoSafeFormsFinding(pageURL: string, evidenceIds: string[]): FindingInput {
  return {
    title: "No safe GET forms were available to test",
    severity: "info",
    category: "form_no_safe_candidate",
    confidence: "high",
    description: `Safe Form Testing did not find a same-origin GET search, filter, sort, or navigation form with non-sensitive fields starting from ${pageURL}.`,
    recommendation: "If safe read-only forms exist, ensure they use GET, have non-sensitive field names, and stay within project allowed hosts.",
    evidenceIds
  };
}

function cleanFormCandidate(form: ExtractedFormCandidate): ExtractedFormCandidate {
  return {
    selectorHint: clean(form.selectorHint, 240),
    pageURL: sanitizeURL(form.pageURL),
    actionURL: sanitizeURL(form.actionURL),
    method: clean(form.method || "GET", 16).toLowerCase(),
    label: clean(form.label, 180),
    fieldCount: Math.max(0, Number(form.fieldCount || 0)),
    passwordFieldCount: Math.max(0, Number(form.passwordFieldCount || 0)),
    fileFieldCount: Math.max(0, Number(form.fileFieldCount || 0)),
    hiddenSensitiveFieldCount: Math.max(0, Number(form.hiddenSensitiveFieldCount || 0)),
    submitButtonCount: Math.max(0, Number(form.submitButtonCount || 0)),
    fields: form.fields.slice(0, 30).map((field) => ({
      name: clean(field.name, 120),
      type: clean(field.type || "text", 40).toLowerCase(),
      label: clean(field.label, 160),
      placeholder: clean(field.placeholder, 160),
      required: Boolean(field.required),
      hidden: Boolean(field.hidden),
      options: field.options.map((option) => clean(option, 80)).filter(Boolean).slice(0, 12)
    }))
  };
}

function classifyFormPurpose(form: ExtractedFormCandidate, target: URL): ClassifiedFormCandidate["classification"] {
  const haystack = `${form.label} ${form.selectorHint} ${target.pathname} ${target.search} ${form.fields
    .map((field) => `${field.name} ${field.label} ${field.placeholder}`)
    .join(" ")}`.toLowerCase();
  if (/\b(delete|destroy|remove|reset|cancel|deactivate|logout|withdraw)\b/.test(haystack)) {
    return "destructive";
  }
  if (/\b(payment|checkout|pay|refund|transfer|billing|invoice\/pay|purchase|order|cart|card|cvv|iban)\b/.test(haystack)) {
    return "payment";
  }
  if (/\b(admin|administrator|manage users|role|permission)\b/.test(haystack)) {
    return "admin_mutation";
  }
  if (/\b(upload|file|attachment|avatar)\b/.test(haystack)) {
    return "upload";
  }
  if (/\b(password|passwd|secret|mfa|otp)\b/.test(haystack)) {
    return "password";
  }
  if (/\b(login|log in|signin|sign in|auth|session|credential|username)\b/.test(haystack)) {
    return "login";
  }
  if (/\b(profile|account|settings|preferences|address|phone)\b/.test(haystack)) {
    return "profile_update";
  }
  if (/\b(contact|support|message|feedback)\b/.test(haystack)) {
    return "contact";
  }
  if (/\b(newsletter|subscribe|subscription|email signup)\b/.test(haystack)) {
    return "newsletter";
  }
  if (/\b(search|query|keyword|term|find)\b/.test(haystack)) {
    return "search";
  }
  if (/\b(filter|category|tag|status|type|facet|products?)\b/.test(haystack)) {
    return "filter";
  }
  if (/\b(sort|order)\b/.test(haystack)) {
    return "sort";
  }
  if (form.fieldCount === 0) {
    return "navigation";
  }
  return "unknown";
}

function buildSafeTestValues(form: ExtractedFormCandidate, classification: ClassifiedFormCandidate["classification"]): Record<string, string> {
  const values: Record<string, string> = {};
  for (const field of form.fields) {
    if (field.hidden || isSensitiveField(field) || !isSupportedField(field)) {
      continue;
    }
    const name = field.name;
    if (!name || !safeFieldNames.test(name)) {
      continue;
    }
    if (safeTextFieldTypes.has(field.type)) {
      values[name] = classification === "search" ? "demo" : "all";
      continue;
    }
    if (safeChoiceFieldTypes.has(field.type)) {
      const option = field.options.find((candidate) => candidate && !isSensitiveName(candidate) && !looksDangerousValue(candidate));
      values[name] = option || "all";
    }
  }
  return values;
}

function skipped(
  form: ExtractedFormCandidate,
  normalizedActionURL: string,
  sameOrigin: boolean,
  classification: ClassifiedFormCandidate["classification"],
  safety: ClassifiedFormCandidate["safety"],
  skipReason: string
): ClassifiedFormCandidate {
  return {
    ...form,
    normalizedActionURL,
    sameOrigin,
    classification,
    safety,
    decision: "skip",
    skipReason,
    testValues: {}
  };
}

function normalizeFormURL(raw: string, baseURL?: string): string {
  const parsed = new URL(raw, baseURL);
  parsed.hash = "";
  parsed.protocol = parsed.protocol.toLowerCase();
  parsed.hostname = parsed.hostname.toLowerCase().replace(/\.$/, "");
  parsed.searchParams.sort();
  return redactSensitiveFormURL(parsed).toString();
}

function redactSensitiveFormURL(url: URL): URL {
  const clone = new URL(url.toString());
  for (const key of Array.from(clone.searchParams.keys())) {
    if (isSensitiveName(key)) {
      clone.searchParams.set(key, "[REDACTED]");
    }
  }
  return clone;
}

function sameOriginURL(url: URL, reference: string): boolean {
  try {
    const parsed = new URL(reference);
    return url.protocol === parsed.protocol && url.hostname === parsed.hostname && normalizedPort(url) === normalizedPort(parsed);
  } catch {
    return false;
  }
}

function hostMatches(hostname: string, allowedHost: string): boolean {
  const host = hostname.toLowerCase().replace(/\.$/, "");
  const allowed = String(allowedHost || "").toLowerCase().replace(/\.$/, "");
  return host === allowed || host.endsWith(`.${allowed}`);
}

function normalizedPort(url: URL): string {
  if (url.port) {
    return url.port;
  }
  return url.protocol === "https:" ? "443" : "80";
}

function hasSensitiveQuery(url: URL): boolean {
  for (const key of url.searchParams.keys()) {
    if (isSensitiveName(key)) {
      return true;
    }
  }
  return false;
}

function isSensitiveField(field: SafeFormField): boolean {
  return field.type === "password" || field.type === "file" || isSensitiveName(`${field.name} ${field.label} ${field.placeholder}`);
}

function isSupportedField(field: SafeFormField): boolean {
  return safeTextFieldTypes.has(field.type) || safeChoiceFieldTypes.has(field.type);
}

function isSensitiveName(value: string): boolean {
  const normalized = value.toLowerCase();
  return sensitiveFieldMarkers.some((marker) => normalized.includes(marker));
}

function looksDangerousForm(form: ExtractedFormCandidate, target: URL): boolean {
  return dangerousFormMarkers.some((marker) =>
    `${form.label} ${form.selectorHint} ${target.pathname} ${target.search} ${form.actionURL} ${form.fields
      .map((field) => `${field.name} ${field.label}`)
      .join(" ")}`.toLowerCase().includes(marker)
  );
}

function looksDangerousValue(value: string): boolean {
  const normalized = value.toLowerCase();
  return dangerousFormMarkers.some((marker) => normalized.includes(marker)) || isSensitiveName(normalized) || normalized.length > 80;
}

function sanitizeURL(value: string): string {
  try {
    const parsed = new URL(value);
    return redactSensitiveFormURL(parsed).toString().slice(0, 1000);
  } catch {
    return clean(value, 1000);
  }
}

function clean(value: unknown, max = 240): string {
  return String(value || "").replace(/\s+/g, " ").trim().slice(0, max);
}

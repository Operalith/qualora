import type { FindingInput } from "./findings";
import {
  classifySafeExplorerAction,
  normalizeSafeExplorerURL,
  type ClassifiedSafeExplorerAction,
  type ExtractedSafeExplorerAction,
  type SafeExplorerPolicy
} from "./safe_explorer";

export type AIBrowserActionType =
  | "goto"
  | "click_link"
  | "click_safe_navigation"
  | "assert_text_visible"
  | "assert_url_contains"
  | "assert_title_contains"
  | "capture_screenshot"
  | "collect_browser_signals"
  | "submit_safe_get_form"
  | "stop";

export type AIBrowserAction = {
  type: AIBrowserActionType;
  target_url?: string;
  path?: string;
  link_text?: string;
  selector_hint?: string;
  form_selector_hint?: string;
  field_values?: Record<string, string>;
  label?: string;
  text?: string;
  reason?: string;
};

export type AIBrowserSuggestion = {
  rationale: string;
  action: AIBrowserAction;
  expected_result?: string;
  risk_assessment?: string;
};

export type AIBrowserObservation = {
  project_name?: string;
  goal?: string;
  current_url: string;
  current_path: string;
  page_title: string;
  visible_text_snippets: string[];
  headings: string[];
  safe_candidate_links: Array<{ text: string; path: string; target_url: string; same_origin: boolean; safety: string; selector_hint: string }>;
  candidate_buttons: Array<{ label: string; safety: string; selector_hint: string }>;
  forms: Array<{
    method: string;
    field_count: number;
    password_field_count: number;
    classification: string;
    safety: string;
    selector_hint: string;
    action_url: string;
    fields: Array<{ name: string; type: string; label: string }>;
  }>;
  console_error_count: number;
  failed_request_count: number;
  previous_steps: Array<{ step_index: number; page: string; action: string; decision: string; result: string }>;
};

export type AIBrowserPolicyInput = {
  suggestion: AIBrowserSuggestion | null;
  observation: AIBrowserObservation;
  policy: SafeExplorerPolicy;
  startURL: string;
  depth: number;
  maxDepth: number;
  visited: Set<string>;
};

export type AIBrowserPolicyDecision = {
  decision: "approved" | "blocked" | "unsupported" | "invalid" | "skipped";
  reason: string;
  action: AIBrowserAction | null;
  targetURL: string;
  selectorHint: string;
  label: string;
};

const supportedActionTypes = new Set<AIBrowserActionType>([
  "goto",
  "click_link",
  "click_safe_navigation",
  "assert_text_visible",
  "assert_url_contains",
  "assert_title_contains",
  "capture_screenshot",
  "collect_browser_signals",
  "submit_safe_get_form",
  "stop"
]);

const dangerousMarkers = /(admin|cancel|deactivate|delete|destroy|logout|mutat|password change|pay|payment|refund|remove|reset|submit|transfer|upload|withdraw)/i;
const sensitiveQueryMarkers = /(access[_-]?token|api[_-]?key|apikey|auth|authorization|credential|jwt|password|passwd|secret|session|token)/i;

export function parseAIBrowserSuggestion(raw: string): { suggestion: AIBrowserSuggestion | null; error: string } {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return { suggestion: null, error: "AI response was not valid JSON" };
  }
  if (!parsed || typeof parsed !== "object") {
    return { suggestion: null, error: "AI response must be a JSON object" };
  }
  const root = parsed as Record<string, unknown>;
  const action = root.action;
  if (!action || typeof action !== "object") {
    return { suggestion: null, error: "AI response is missing action object" };
  }
  const actionObject = action as Record<string, unknown>;
  const type = cleanString(actionObject.type, 80) as AIBrowserActionType;
  if (!supportedActionTypes.has(type)) {
    return { suggestion: null, error: `unsupported action type ${type || "missing"}` };
  }
  const suggestion: AIBrowserSuggestion = {
    rationale: cleanString(root.rationale, 500),
    expected_result: cleanString(root.expected_result, 500),
    risk_assessment: cleanString(root.risk_assessment, 120),
    action: {
      type,
      target_url: cleanString(actionObject.target_url, 500),
      path: cleanString(actionObject.path, 500),
      link_text: cleanString(actionObject.link_text, 180),
      selector_hint: cleanString(actionObject.selector_hint, 240),
      form_selector_hint: cleanString(actionObject.form_selector_hint, 240),
      field_values: cleanFieldValues(actionObject.field_values),
      label: cleanString(actionObject.label, 180),
      text: cleanString(actionObject.text, 500),
      reason: cleanString(actionObject.reason, 500)
    }
  };
  const missing = requiredFieldMissing(suggestion.action);
  if (missing) {
    return { suggestion: null, error: missing };
  }
  return { suggestion, error: "" };
}

export function evaluateAIBrowserPolicy(input: AIBrowserPolicyInput): AIBrowserPolicyDecision {
  const suggestion = input.suggestion;
  if (!suggestion) {
    return invalid("missing AI suggestion", null);
  }
  const action = suggestion.action;
  if (!action || !supportedActionTypes.has(action.type)) {
    return unsupported(`unsupported action type ${action?.type || "missing"}`, action ?? null);
  }
  const explicitLabel = cleanString(action.label || action.link_text || action.text || action.reason, 240);
  const label = explicitLabel || action.type;
  if (action.type !== "submit_safe_get_form" && looksDangerous(label)) {
    return blocked("action label looks destructive or mutating", action);
  }
  if (action.type === "stop") {
    return { decision: "approved", reason: "stop requested by AI", action, targetURL: "", selectorHint: "", label };
  }
  if (["capture_screenshot", "collect_browser_signals"].includes(action.type)) {
    return { decision: "approved", reason: "safe metadata action", action, targetURL: "", selectorHint: "", label };
  }
  if (["assert_text_visible", "assert_url_contains", "assert_title_contains"].includes(action.type)) {
    if (!cleanString(action.text, 500)) {
      return invalid("assertion action is missing text", action);
    }
    if (looksDangerous(action.text || "")) {
      return blocked("assertion text looks destructive or sensitive", action);
    }
    return { decision: "approved", reason: "safe assertion action", action, targetURL: "", selectorHint: "", label };
  }

  if (input.depth >= input.maxDepth) {
    return blocked("max_depth reached", action);
  }
  if (action.type === "submit_safe_get_form") {
    return evaluateSafeGetFormPolicy(input, action, explicitLabel);
  }
  const target = action.target_url || action.path || "";
  if (!target) {
    return invalid("navigation action is missing target_url or path", action);
  }
  let targetURL: string;
  try {
    targetURL = normalizeSafeExplorerURL(target, input.observation.current_url || input.startURL);
  } catch {
    return invalid("navigation target URL is invalid", action);
  }
  let parsed: URL;
  try {
    parsed = new URL(targetURL);
  } catch {
    return invalid("navigation target URL is invalid", action);
  }
  if (hasSensitiveQuery(parsed)) {
    return blocked("navigation target contains sensitive query parameters", action, targetURL);
  }
  if (looksDangerous(parsed.pathname) || looksDangerous(target)) {
    return blocked("navigation target path looks destructive or mutating", action, targetURL);
  }
  if (input.visited.has(targetURL)) {
    return blocked("loop detected for previously visited URL", action, targetURL);
  }

  const observed = findObservedCandidate(input.observation, action, targetURL);
  if (action.type === "click_link" || action.type === "click_safe_navigation") {
    if (!observed) {
      return blocked("navigation action did not match an observed safe candidate", action, targetURL);
    }
    if (action.selector_hint && action.selector_hint !== observed.selector_hint) {
      return blocked("selector hint did not match an observed safe candidate", action, targetURL);
    }
  }

  const explorerAction: ExtractedSafeExplorerAction = {
    actionType: "link_navigation",
    label: label || observed?.text || targetURL,
    text: action.link_text || observed?.text || label,
    selectorHint: action.selector_hint || observed?.selector_hint || "",
    href: targetURL,
    targetURL,
    method: "GET"
  };
  const classified = classifySafeExplorerAction(explorerAction, input.policy);
  if (classified.decision !== "execute") {
    return policyFromSafeExplorer(classified, action);
  }
  return {
    decision: "approved",
    reason: "policy approved safe same-origin navigation",
    action,
    targetURL: classified.normalizedURL || targetURL,
    selectorHint: action.selector_hint || observed?.selector_hint || "",
    label: label || observed?.text || ""
  };
}

export function buildAIBrowserFinding(input: {
  decision: AIBrowserPolicyDecision;
  executionStatus: string;
  executionError?: string;
  consoleErrorCount: number;
  failedRequestCount: number;
  evidenceIds: string[];
}): FindingInput[] {
  const findings: FindingInput[] = [];
  if (input.decision.decision === "blocked") {
    findings.push({
      title: "AI Browser policy blocked an action",
      severity: looksDangerous(input.decision.label + " " + input.decision.targetURL) ? "medium" : "low",
      category: "ai_browser_policy_block",
      confidence: "high",
      description: `AI suggested ${input.decision.action?.type || "unknown"} but policy blocked it: ${input.decision.reason}.`,
      recommendation: "Review the suggested action and keep AI Browser Control constrained to observed safe same-origin navigation.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.decision.decision === "invalid") {
    findings.push({
      title: "AI Browser invalid action",
      severity: "low",
      category: "ai_browser_invalid_action",
      confidence: "high",
      description: `AI Browser Control rejected an invalid suggestion: ${input.decision.reason}.`,
      recommendation: "Use an OpenAI-compatible provider that returns strict JSON matching Qualora's typed action schema.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.decision.decision === "unsupported") {
    findings.push({
      title: "AI Browser unsupported action",
      severity: "info",
      category: "ai_browser_unsupported_action",
      confidence: "high",
      description: `AI Browser Control skipped an unsupported action: ${input.decision.reason}.`,
      recommendation: "Keep prompts and provider behavior limited to the supported safe action schema.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.executionStatus === "failed" || input.executionStatus === "error") {
    findings.push({
      title: "AI Browser navigation or assertion failed",
      severity: "medium",
      category: input.decision.action?.type?.startsWith("assert_") ? "ai_browser_assertion_failure" : "ai_browser_navigation_failure",
      confidence: "medium",
      description: `AI Browser Control could not execute the approved action: ${input.executionError || "execution failed"}.`,
      recommendation: "Verify the target page is stable and the proposed action remains within the supported safe browser action subset.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.consoleErrorCount > 0) {
    findings.push({
      title: "Console errors observed during AI Browser Control",
      severity: "medium",
      category: "ai_browser_console_error",
      confidence: "medium",
      description: `${input.consoleErrorCount} console error(s) were observed while evaluating an AI Browser step.`,
      recommendation: "Review frontend runtime errors on the affected page.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.failedRequestCount > 0) {
    findings.push({
      title: "Network failures observed during AI Browser Control",
      severity: "medium",
      category: "ai_browser_network_failure",
      confidence: "medium",
      description: `${input.failedRequestCount} failed network request(s) were observed while evaluating an AI Browser step.`,
      recommendation: "Inspect failed first-party resources and API calls for the affected page.",
      evidenceIds: input.evidenceIds
    });
  }
  if (input.decision.reason.includes("loop detected")) {
    findings.push({
      title: "AI Browser loop detected",
      severity: "low",
      category: "ai_browser_loop_detected",
      confidence: "high",
      description: "AI Browser Control stopped a repeated navigation target.",
      recommendation: "Refine the goal or provider prompt to explore a new observed safe page.",
      evidenceIds: input.evidenceIds
    });
  }
  return findings;
}

export function redactObservationValue(input: string): string {
  return cleanString(input, 1000)
    .replace(/(authorization|password|passwd|token|secret|api[_-]?key|cookie|session)=([^&\s]+)/gi, "$1=[REDACTED]")
    .replace(/(Bearer|Basic)\s+[A-Za-z0-9._~+/=-]+/gi, "$1 [REDACTED]");
}

function requiredFieldMissing(action: AIBrowserAction): string {
  if (["goto", "click_link"].includes(action.type) && !(action.target_url || action.path)) {
    return `${action.type} requires target_url or path`;
  }
  if (action.type === "click_safe_navigation" && !action.selector_hint) {
    return "click_safe_navigation requires selector_hint";
  }
  if (action.type === "submit_safe_get_form") {
    if (!(action.form_selector_hint || action.selector_hint)) {
      return "submit_safe_get_form requires form_selector_hint";
    }
    if (!action.field_values || Object.keys(action.field_values).length === 0) {
      return "submit_safe_get_form requires field_values";
    }
  }
  if (["assert_text_visible", "assert_url_contains", "assert_title_contains"].includes(action.type) && !action.text) {
    return `${action.type} requires text`;
  }
  return "";
}

function evaluateSafeGetFormPolicy(input: AIBrowserPolicyInput, action: AIBrowserAction, label: string): AIBrowserPolicyDecision {
  const selector = cleanString(action.form_selector_hint || action.selector_hint, 240);
  const observed = input.observation.forms.find((form) => form.selector_hint === selector);
  if (!observed) {
    return blocked("form action did not match an observed safe form candidate", action);
  }
  if (observed.method.toLowerCase() !== "get") {
    return blocked("form method is not safe GET", action);
  }
  if (observed.safety !== "safe") {
    return blocked("observed form was not classified safe", action);
  }
  if (!["search", "filter", "sort", "navigation"].includes(observed.classification)) {
    return blocked("observed form classification is not eligible for safe submission", action);
  }

  const fieldValues = action.field_values || {};
  const allowedFields = new Map(observed.fields.map((field) => [field.name, field]));
  const submitted = new URL(observed.action_url || input.observation.current_url, input.observation.current_url);
  for (const [rawName, rawValue] of Object.entries(fieldValues)) {
    const name = cleanString(rawName, 120);
    const value = cleanString(rawValue, 80);
    const observedField = allowedFields.get(name);
    if (!observedField) {
      return blocked(`form field ${name || "missing"} was not in the sanitized observation`, action);
    }
    if (hasSensitiveQueryName(name) || hasSensitiveQueryName(value) || looksDangerous(name) || looksDangerous(value)) {
      return blocked("form field name or value looked sensitive, destructive, or mutating", action);
    }
    if (!["", "text", "search", "select-one", "select-multiple", "checkbox", "radio"].includes(observedField.type)) {
      return blocked(`form field ${name} uses unsupported type ${observedField.type}`, action);
    }
    submitted.searchParams.set(name, value);
  }
  submitted.hash = "";
  submitted.searchParams.sort();
  if (hasSensitiveQuery(submitted)) {
    return blocked("submitted form URL contains sensitive query parameters", action, submitted.toString());
  }
  if (looksDangerous(submitted.pathname) || (label && looksDangerous(label))) {
    return blocked("submitted form path or label looks destructive or mutating", action, submitted.toString());
  }
  const targetURL = submitted.toString();
  if (input.visited.has(targetURL)) {
    return blocked("loop detected for previously visited URL", action, targetURL);
  }

  const explorerAction: ExtractedSafeExplorerAction = {
    actionType: "form_get",
    label: label || observed.classification,
    text: label || observed.classification,
    selectorHint: selector,
    href: observed.action_url,
    targetURL,
    method: "GET",
    fieldCount: observed.field_count,
    hasPasswordField: observed.password_field_count > 0,
    hasFileField: false,
    hasHiddenSensitiveField: false
  };
  const classified = classifySafeExplorerAction(explorerAction, { ...input.policy, allowGetForms: true });
  if (classified.decision !== "execute") {
    return policyFromSafeExplorer(classified, action);
  }
  return {
    decision: "approved",
    reason: "policy approved safe same-origin GET form submission",
    action,
    targetURL: classified.normalizedURL || targetURL,
    selectorHint: selector,
    label: label || observed.classification
  };
}

function findObservedCandidate(observation: AIBrowserObservation, action: AIBrowserAction, targetURL: string) {
  const parsed = new URL(targetURL);
  const path = parsed.pathname + parsed.search;
  return observation.safe_candidate_links.find((candidate) => {
    if (candidate.target_url === targetURL || candidate.path === path) {
      return true;
    }
    if (action.link_text && candidate.text.toLowerCase() === action.link_text.toLowerCase()) {
      return true;
    }
    return Boolean(action.selector_hint && candidate.selector_hint === action.selector_hint);
  });
}

function policyFromSafeExplorer(classified: ClassifiedSafeExplorerAction, action: AIBrowserAction): AIBrowserPolicyDecision {
  if (classified.safety === "unsupported") {
    return unsupported(classified.skipReason || "unsupported safe explorer action", action, classified.normalizedURL);
  }
  return blocked(classified.skipReason || "safe explorer policy blocked action", action, classified.normalizedURL);
}

function invalid(reason: string, action: AIBrowserAction | null, targetURL = ""): AIBrowserPolicyDecision {
  return { decision: "invalid", reason, action, targetURL, selectorHint: action?.selector_hint || "", label: action?.label || action?.link_text || action?.type || "" };
}

function unsupported(reason: string, action: AIBrowserAction | null, targetURL = ""): AIBrowserPolicyDecision {
  return { decision: "unsupported", reason, action, targetURL, selectorHint: action?.selector_hint || "", label: action?.label || action?.link_text || action?.type || "" };
}

function blocked(reason: string, action: AIBrowserAction, targetURL = ""): AIBrowserPolicyDecision {
  return { decision: "blocked", reason, action, targetURL, selectorHint: action.selector_hint || "", label: action.label || action.link_text || action.text || action.type || "" };
}

function cleanString(value: unknown, max = 240): string {
  return String(value || "").replace(/\s+/g, " ").trim().slice(0, max);
}

function cleanFieldValues(value: unknown): Record<string, string> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  const out: Record<string, string> = {};
  for (const [key, raw] of Object.entries(value as Record<string, unknown>).slice(0, 8)) {
    const name = cleanString(key, 120);
    const fieldValue = cleanString(raw, 80);
    if (name && fieldValue) {
      out[name] = fieldValue;
    }
  }
  return out;
}

function looksDangerous(value: string): boolean {
  return dangerousMarkers.test(value);
}

function hasSensitiveQuery(url: URL): boolean {
  for (const key of url.searchParams.keys()) {
    if (sensitiveQueryMarkers.test(key)) {
      return true;
    }
  }
  return false;
}

function hasSensitiveQueryName(value: string): boolean {
  return sensitiveQueryMarkers.test(value);
}

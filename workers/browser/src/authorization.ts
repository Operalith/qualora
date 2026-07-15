import type { BrowserResult, FindingInput } from "./findings";

export type AuthorizationExpectedOutcome = "allowed" | "denied";
export type AuthorizationActualOutcome = "allowed" | "denied" | "unknown";
export type AuthorizationResultStatus = "passed" | "failed" | "skipped" | "error";

export type AuthorizationOutcomeInput = {
  statusCode: number | null;
  bodyText: string;
  successTextContains: string;
  deniedTextContains: string;
  loadError: string;
  timedOut: boolean;
};

export type AuthorizationFindingInput = {
  checkName: string;
  actorName: string;
  actorRoleName: string;
  targetURL: string;
  expectedOutcome: AuthorizationExpectedOutcome;
  actualOutcome: AuthorizationActualOutcome;
  resultStatus: AuthorizationResultStatus;
  errorMessage: string;
  skipReason: string;
  timedOut: boolean;
  consoleErrors: BrowserResult["consoleErrors"];
  failedRequests: BrowserResult["failedRequests"];
  blockedRequests: BrowserResult["blockedRequests"];
};

export function classifyAuthorizationOutcome(input: AuthorizationOutcomeInput): AuthorizationActualOutcome {
  if (input.loadError || input.timedOut) {
    return "unknown";
  }
  if (input.statusCode !== null && [401, 403, 404].includes(input.statusCode)) {
    return "denied";
  }
  const body = input.bodyText.toLowerCase();
  if (input.deniedTextContains && body.includes(input.deniedTextContains.toLowerCase())) {
    return "denied";
  }
  if (input.statusCode !== null && input.statusCode >= 200 && input.statusCode < 300) {
    if (!input.successTextContains || body.includes(input.successTextContains.toLowerCase())) {
      return "allowed";
    }
    return "unknown";
  }
  return "unknown";
}

export function compareAuthorizationOutcome(
  expectedOutcome: AuthorizationExpectedOutcome,
  actualOutcome: AuthorizationActualOutcome
): AuthorizationResultStatus {
  if (actualOutcome === "unknown") {
    return "failed";
  }
  return expectedOutcome === actualOutcome ? "passed" : "failed";
}

export function buildAuthorizationFindings(input: AuthorizationFindingInput, evidenceIds: string[]): FindingInput[] {
  const findings: FindingInput[] = [];
  const actor = input.actorRoleName || input.actorName || "configured actor";
  const target = input.targetURL || "configured target";

  if (input.skipReason) {
    findings.push({
      title: "Authorization target blocked",
      severity: "low",
      category: "authorization_target_blocked",
      confidence: "high",
      description: [
        `Summary: Qualora skipped authorization check ${input.checkName} because the target was outside the configured safe scope: ${input.skipReason}`,
        `Steps to reproduce: review the authorization check target ${target} and the project's allowed_hosts policy.`
      ].join("\n"),
      recommendation: "Keep authorization check targets on the configured frontend origin and within allowed_hosts.",
      evidenceIds
    });
    return findings;
  }

  if (input.errorMessage && /login/i.test(input.errorMessage)) {
    findings.push({
      title: "Authorization login failed",
      severity: "high",
      category: "authorization_login_failure",
      confidence: "high",
      description: [
        `Summary: Qualora could not log in as ${actor} before running authorization check ${input.checkName}: ${input.errorMessage}`,
        "Steps to reproduce: use the configured login URL and selectors with the actor test credential profile, then navigate to the authorization target."
      ].join("\n"),
      recommendation: "Verify the actor credential profile, selector configuration, and deterministic success criteria.",
      evidenceIds
    });
    return findings;
  }

  if (input.timedOut) {
    findings.push({
      title: "Authorization check timed out",
      severity: "high",
      category: "authorization_check_timeout",
      confidence: "high",
      description: [
        `Summary: authorization check ${input.checkName} timed out while testing ${target} as ${actor}.`,
        "Steps to reproduce: log in with the configured actor test account and open the configured authorization target."
      ].join("\n"),
      recommendation: "Verify target availability and keep authorization targets stable for deterministic smoke checks.",
      evidenceIds
    });
  } else if (input.actualOutcome === "unknown") {
    findings.push({
      title: "Authorization outcome was unclear",
      severity: "medium",
      category: "authorization_check_unknown",
      confidence: "medium",
      description: [
        `Summary: authorization check ${input.checkName} could not classify access to ${target} as clearly allowed or denied for ${actor}.`,
        "Steps to reproduce: log in with the configured actor test account and inspect the configured authorization target, success text, and denied text."
      ].join("\n"),
      recommendation: "Add explicit success_text_contains or denied_text_contains criteria for this authorization check.",
      evidenceIds
    });
  } else if (input.expectedOutcome === "denied" && input.actualOutcome === "allowed") {
    findings.push({
      title: "Authorization bypass detected",
      severity: "critical",
      category: "authorization_bypass",
      confidence: "high",
      description: [
        `Summary: ${actor} was expected to be denied but Qualora observed allowed access to ${target}.`,
        "Steps to reproduce: log in with the configured actor test account and open the configured authorization target. Do not use production credentials."
      ].join("\n"),
      recommendation: "Review server-side authorization checks for this resource and ensure access decisions do not depend only on client-side UI controls.",
      evidenceIds
    });
  } else if (input.expectedOutcome === "allowed" && input.actualOutcome === "denied") {
    findings.push({
      title: "Expected access was denied",
      severity: "high",
      category: "unexpected_access_denied",
      confidence: "high",
      description: [
        `Summary: ${actor} was expected to be allowed but Qualora observed denied access to ${target}.`,
        "Steps to reproduce: log in with the configured actor test account and open the configured authorization target."
      ].join("\n"),
      recommendation: "Review role/resource permissions and route authorization rules for the configured actor.",
      evidenceIds
    });
  }

  if (input.consoleErrors.length > 0) {
    findings.push({
      title: "Console error detected during authorization check",
      severity: "medium",
      category: "authorization_console_error",
      confidence: "medium",
      description: `Summary: the browser observed ${input.consoleErrors.length} console error(s) during authorization check ${input.checkName}.`,
      recommendation: "Review frontend console errors on the login and target routes.",
      evidenceIds
    });
  }
  if (input.failedRequests.length > 0) {
    findings.push({
      title: "Failed network request detected during authorization check",
      severity: "medium",
      category: "authorization_network_failure",
      confidence: "medium",
      description: `Summary: the browser observed ${input.failedRequests.length} failed network request(s) during authorization check ${input.checkName}.`,
      recommendation: "Inspect failed requests and ensure first-party dependencies are reachable from the worker.",
      evidenceIds
    });
  }
  if (input.blockedRequests.length > 0) {
    findings.push({
      title: "Out-of-scope browser request blocked during authorization check",
      severity: "info",
      category: "authorization_target_blocked",
      confidence: "high",
      description: `Summary: Qualora blocked ${input.blockedRequests.length} request(s) outside the project's allowed_hosts policy during authorization check ${input.checkName}.`,
      recommendation: "Add required first-party hosts to allowed_hosts or remove unexpected third-party dependencies from the authorization target.",
      evidenceIds
    });
  }

  return findings;
}

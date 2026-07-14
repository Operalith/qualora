export type FindingInput = {
  title: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  category: string;
  confidence: "high" | "medium" | "low";
  description: string;
  recommendation: string;
  evidenceIds: string[];
};

export type BrowserResult = {
  targetURL: string;
  finalURL: string;
  pageTitle: string;
  statusCode: number | null;
  bodyTextLength: number | null;
  loadError: string;
  timedOut: boolean;
  consoleErrors: Array<{ type: string; text: string; location: string }>;
  failedRequests: Array<{ url: string; method: string; failure: string }>;
  blockedRequests: Array<{ url: string; reason: string }>;
  screenshot: Buffer | null;
};

export function buildFindings(result: BrowserResult, evidenceIds: string[]): FindingInput[] {
  const findings: FindingInput[] = [];

  if (result.loadError) {
    findings.push({
      title: result.timedOut ? "Page load timed out" : "Page load failed",
      severity: "high",
      category: "frontend",
      confidence: "high",
      description: [
        `Summary: the target page did not complete the initial browser load: ${result.loadError}`,
        `Steps to reproduce: open ${result.targetURL} in a browser from the worker network and wait for the initial document load.`
      ].join("\n"),
      recommendation: "Verify that the frontend URL is reachable from the worker container and that the application serves a valid page.",
      evidenceIds
    });
  } else if (result.statusCode !== null && result.statusCode >= 500) {
    findings.push({
      title: "Server error while loading page",
      severity: "high",
      category: "frontend",
      confidence: "high",
      description: [
        `Summary: the target page returned HTTP ${result.statusCode}.`,
        `Steps to reproduce: open ${result.targetURL} and inspect the initial document response.`
      ].join("\n"),
      recommendation: "Inspect the frontend service and upstream dependencies for server-side errors.",
      evidenceIds
    });
  } else if (result.statusCode !== null && (result.statusCode < 200 || result.statusCode >= 400)) {
    findings.push({
      title: "Non-success status while loading page",
      severity: "medium",
      category: "frontend",
      confidence: "high",
      description: [
        `Summary: the target page returned HTTP ${result.statusCode}.`,
        `Steps to reproduce: open ${result.targetURL} and confirm the configured frontend route returns a successful document response.`
      ].join("\n"),
      recommendation: "Confirm that the configured frontend URL is correct and reachable within the allowed test scope.",
      evidenceIds
    });
  }

  if (result.consoleErrors.length > 0) {
    findings.push({
      title: "Console error detected",
      severity: "medium",
      category: "frontend",
      confidence: "medium",
      description: [
        `Summary: the browser observed ${result.consoleErrors.length} console error(s) during page load.`,
        `Steps to reproduce: open ${result.finalURL || result.targetURL}, then inspect browser console errors during initial load.`
      ].join("\n"),
      recommendation: "Review browser console errors and fix uncaught frontend exceptions or failed client-side initialization.",
      evidenceIds
    });
  }

  if (result.failedRequests.length > 0) {
    findings.push({
      title: "Failed network request detected",
      severity: "medium",
      category: "frontend",
      confidence: "medium",
      description: [
        `Summary: the browser observed ${result.failedRequests.length} failed network request(s) within the allowed scope.`,
        `Steps to reproduce: open ${result.finalURL || result.targetURL} and review failed requests in browser network tools.`
      ].join("\n"),
      recommendation: "Inspect failed requests and ensure required assets, APIs, and dependencies are available during page load.",
      evidenceIds
    });
  }

  if (!result.loadError && result.statusCode !== null && result.statusCode < 400) {
    const emptyTitle = result.pageTitle.trim() === "";
    const emptyBody = result.bodyTextLength === 0;
    if (emptyTitle || emptyBody) {
      findings.push({
        title: emptyTitle && emptyBody ? "Loaded page appears empty" : "Loaded page has sparse browser metadata",
        severity: emptyTitle && emptyBody ? "medium" : "low",
        category: "frontend",
        confidence: "medium",
        description: [
          `Summary: the page loaded with ${emptyTitle ? "an empty title" : "a title"} and ${emptyBody ? "an empty body" : "body text"}.`,
          `Steps to reproduce: open ${result.finalURL || result.targetURL} and inspect the rendered title and body content.`
        ].join("\n"),
        recommendation: "Verify that the configured URL points to a meaningful application route and that the initial render contains visible content.",
        evidenceIds
      });
    }
  }

  if (result.blockedRequests.length > 0) {
    findings.push({
      title: "Out-of-scope browser request blocked",
      severity: "info",
      category: "scope",
      confidence: "high",
      description: [
        `Summary: the browser blocked ${result.blockedRequests.length} request(s) outside the project's allowed hosts.`,
        "Steps to reproduce: load the page and compare requested hosts against the project's allowed_hosts policy."
      ].join("\n"),
      recommendation: "Add required first-party hosts to allowed_hosts or remove unexpected third-party dependencies from the smoke path.",
      evidenceIds
    });
  }

  return findings;
}

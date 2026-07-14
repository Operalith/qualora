import assert from "node:assert/strict";
import test from "node:test";
import { buildFindings, type BrowserResult } from "./findings";

function result(overrides: Partial<BrowserResult> = {}): BrowserResult {
  return {
    targetURL: "http://demo-web:8080/",
    finalURL: "http://demo-web:8080/",
    pageTitle: "Qualora Demo Web",
    statusCode: 200,
    bodyTextLength: 42,
    loadError: "",
    timedOut: false,
    consoleErrors: [],
    failedRequests: [],
    blockedRequests: [],
    screenshot: null,
    ...overrides
  };
}

test("buildFindings classifies browser timeouts as high severity", () => {
  const findings = buildFindings(result({ loadError: "Timeout 30000ms exceeded", timedOut: true }), ["evidence-id"]);

  assert.equal(findings[0].title, "Page load timed out");
  assert.equal(findings[0].severity, "high");
  assert.match(findings[0].description, /Steps to reproduce/);
});

test("buildFindings classifies console errors as medium severity", () => {
  const findings = buildFindings(
    result({
      consoleErrors: [{ type: "error", text: "ReferenceError: app is not defined", location: "http://demo-web:8080/:1:1" }]
    }),
    ["evidence-id"]
  );

  assert.ok(findings.some((finding) => finding.title === "Console error detected" && finding.severity === "medium"));
});

test("buildFindings creates a high severity finding for 5xx document responses", () => {
  const findings = buildFindings(result({ statusCode: 503 }), ["evidence-id"]);

  assert.ok(findings.some((finding) => finding.title === "Server error while loading page" && finding.severity === "high"));
});

test("buildFindings creates empty page findings for sparse loaded pages", () => {
  const findings = buildFindings(result({ pageTitle: "", bodyTextLength: 0 }), ["evidence-id"]);

  assert.ok(findings.some((finding) => finding.title === "Loaded page appears empty" && finding.severity === "medium"));
});

test("buildFindings records blocked out-of-scope requests as informational", () => {
  const findings = buildFindings(
    result({
      blockedRequests: [{ url: "https://cdn.example.net/app.js", reason: "host cdn.example.net is not present in allowed_hosts" }]
    }),
    ["evidence-id"]
  );

  assert.ok(findings.some((finding) => finding.title === "Out-of-scope browser request blocked" && finding.severity === "info"));
});

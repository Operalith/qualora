import assert from "node:assert/strict";
import test from "node:test";
import { buildAuthorizationFindings, classifyAuthorizationOutcome, compareAuthorizationOutcome } from "./authorization";

test("classifyAuthorizationOutcome treats denied status codes as denied", () => {
  assert.equal(
    classifyAuthorizationOutcome({
      statusCode: 403,
      bodyText: "Access denied",
      successTextContains: "",
      deniedTextContains: "",
      loadError: "",
      timedOut: false
    }),
    "denied"
  );
});

test("classifyAuthorizationOutcome requires configured success text when provided", () => {
  assert.equal(
    classifyAuthorizationOutcome({
      statusCode: 200,
      bodyText: "Invoice for Customer A",
      successTextContains: "Invoice for Customer A",
      deniedTextContains: "Access denied",
      loadError: "",
      timedOut: false
    }),
    "allowed"
  );
  assert.equal(
    classifyAuthorizationOutcome({
      statusCode: 200,
      bodyText: "Unexpected page",
      successTextContains: "Invoice for Customer A",
      deniedTextContains: "Access denied",
      loadError: "",
      timedOut: false
    }),
    "unknown"
  );
});

test("compareAuthorizationOutcome passes expected denied and expected allowed cases", () => {
  assert.equal(compareAuthorizationOutcome("denied", "denied"), "passed");
  assert.equal(compareAuthorizationOutcome("allowed", "allowed"), "passed");
  assert.equal(compareAuthorizationOutcome("denied", "allowed"), "failed");
  assert.equal(compareAuthorizationOutcome("allowed", "unknown"), "failed");
});

test("buildAuthorizationFindings creates critical bypass finding for denied expectation", () => {
  const findings = buildAuthorizationFindings(
    {
      checkName: "Customer B denied customer A invoice",
      actorName: "Customer B",
      actorRoleName: "customer-b",
      targetURL: "http://demo-web:8080/customers/a/invoice",
      expectedOutcome: "denied",
      actualOutcome: "allowed",
      resultStatus: "failed",
      errorMessage: "",
      skipReason: "",
      timedOut: false,
      consoleErrors: [],
      failedRequests: [],
      blockedRequests: []
    },
    ["evidence-id"]
  );

  assert.equal(findings[0].category, "authorization_bypass");
  assert.equal(findings[0].severity, "critical");
  assert.match(findings[0].description, /Steps to reproduce/);
});

test("buildAuthorizationFindings creates login failure finding without secret values", () => {
  const findings = buildAuthorizationFindings(
    {
      checkName: "Admin route",
      actorName: "Admin",
      actorRoleName: "admin",
      targetURL: "http://demo-web:8080/admin",
      expectedOutcome: "allowed",
      actualOutcome: "unknown",
      resultStatus: "error",
      errorMessage: "login failed for configured actor",
      skipReason: "",
      timedOut: false,
      consoleErrors: [],
      failedRequests: [],
      blockedRequests: []
    },
    ["evidence-id"]
  );

  const raw = JSON.stringify(findings);
  assert.equal(findings[0].category, "authorization_login_failure");
  assert.doesNotMatch(raw, /password|admin-password/);
});

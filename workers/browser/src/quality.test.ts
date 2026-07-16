import assert from "node:assert/strict";
import test from "node:test";
import { buildQualityResults, type QualityPageSnapshot } from "./quality";

function baseSnapshot(overrides: Partial<QualityPageSnapshot> = {}): QualityPageSnapshot {
  return {
    targetURL: "https://app.example.com/",
    finalURL: "https://app.example.com/",
    title: "Demo App",
    statusCode: 200,
    headers: {
      "content-security-policy": "default-src 'self'; frame-ancestors 'none'",
      "x-content-type-options": "nosniff",
      "referrer-policy": "strict-origin-when-cross-origin",
      "permissions-policy": "camera=()",
      "strict-transport-security": "max-age=31536000"
    },
    isHTTPS: true,
    bodyTextLength: 120,
    loadDurationMS: 250,
    loadError: "",
    consoleErrorCount: 0,
    failedRequestCount: 0,
    blockedRequestCount: 0,
    cookies: [],
    resources: [],
    forms: {
      total: 0,
      passwordForms: 0,
      passwordAutocompleteIssues: 0,
      insecureActions: 0
    },
    accessibility: {
      hasTitle: true,
      htmlLang: "en",
      hasMainLandmark: true,
      imagesTotal: 0,
      imagesMissingAlt: 0,
      inputsTotal: 0,
      inputsMissingLabels: 0,
      buttonsTotal: 0,
      buttonsMissingNames: 0,
      linksTotal: 0,
      linksMissingNames: 0,
      imagesMissingDimensions: 0
    },
    ...overrides
  };
}

test("buildQualityResults creates passive security findings for missing headers", () => {
  const results = buildQualityResults(baseSnapshot({ headers: {} }), {
    includeSecurity: true,
    includeAccessibility: false,
    includePerformance: false
  });
  const ruleIDs = results.map((result) => result.ruleID);
  assert.ok(ruleIDs.includes("missing_csp"));
  assert.ok(ruleIDs.includes("missing_frame_protection"));
  assert.ok(ruleIDs.includes("missing_hsts"));
});

test("buildQualityResults creates accessibility findings for obvious metadata gaps", () => {
  const results = buildQualityResults(
    baseSnapshot({
      title: "",
      accessibility: {
        hasTitle: false,
        htmlLang: "",
        hasMainLandmark: false,
        imagesTotal: 2,
        imagesMissingAlt: 1,
        inputsTotal: 3,
        inputsMissingLabels: 2,
        buttonsTotal: 2,
        buttonsMissingNames: 1,
        linksTotal: 4,
        linksMissingNames: 1,
        imagesMissingDimensions: 0
      }
    }),
    {
      includeSecurity: false,
      includeAccessibility: true,
      includePerformance: false
    }
  );
  assert.deepEqual(
    results.map((result) => result.ruleID).sort(),
    ["buttons_missing_names", "images_missing_alt", "inputs_missing_labels", "links_missing_names", "missing_html_lang", "missing_main_landmark", "missing_title"].sort()
  );
});

test("buildQualityResults creates performance findings without response bodies", () => {
  const results = buildQualityResults(
    baseSnapshot({
      loadDurationMS: 4500,
      consoleErrorCount: 1,
      failedRequestCount: 2,
      resources: [
        {
          url: "https://app.example.com/assets/app.js",
          status: 200,
          contentType: "application/javascript",
          resourceType: "script",
          contentLength: 400_000
        }
      ]
    }),
    {
      includeSecurity: false,
      includeAccessibility: false,
      includePerformance: true
    }
  );
  const ruleIDs = results.map((result) => result.ruleID);
  assert.ok(ruleIDs.includes("page_load_slow"));
  assert.ok(ruleIDs.includes("console_errors"));
  assert.ok(ruleIDs.includes("failed_network_requests"));
  assert.ok(ruleIDs.includes("large_javascript_bundle"));
  assert.equal(JSON.stringify(results).includes("password"), false);
});

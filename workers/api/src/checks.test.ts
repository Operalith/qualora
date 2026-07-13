import assert from "node:assert/strict";
import test from "node:test";
import {
  buildFindingsForObservation,
  endpointURL,
  parseOpenAPIDocument,
  safeMethodsOnly,
  validateTargetURL
} from "./checks";

test("safeMethodsOnly filters unsafe methods by default", () => {
  assert.deepEqual(safeMethodsOnly(["get", "post", "HEAD", "delete", "options"], false), ["GET", "HEAD", "OPTIONS"]);
});

test("safeMethodsOnly can include unsafe methods only when explicitly enabled", () => {
  assert.deepEqual(safeMethodsOnly(["get", "post", "patch"], true), ["GET", "POST", "PATCH"]);
});

test("parseOpenAPIDocument parses safe OpenAPI 3 operations", () => {
  const summary = parseOpenAPIDocument(
    JSON.stringify({
      openapi: "3.0.3",
      paths: {
        "/health": {
          get: {
            responses: {
              "200": {
                description: "ok",
                content: {
                  "application/json": {}
                }
              }
            }
          },
          post: {
            responses: {
              "201": { description: "created" }
            }
          }
        }
      }
    }),
    "application/json"
  );

  assert.equal(summary.version, "3.0.3");
  assert.equal(summary.pathCount, 1);
  assert.equal(summary.operationCount, 2);
  assert.equal(summary.safeOperationCount, 1);
  assert.equal(summary.skippedUnsafeOperationCount, 1);
  assert.equal(summary.endpoints[0].method, "GET");
  assert.deepEqual(summary.endpoints[0].expectedStatuses, ["200"]);
  assert.deepEqual(summary.endpoints[0].expectedContentTypes, ["application/json"]);
});

test("parseOpenAPIDocument rejects non OpenAPI 3 documents", () => {
  assert.throws(() => parseOpenAPIDocument(JSON.stringify({ swagger: "2.0", paths: {} }), "application/json"), /OpenAPI 3/);
});

test("buildFindingsForObservation creates 5xx findings", () => {
  const findings = buildFindingsForObservation(
    {
      method: "GET",
      url: "https://api.example.com/health",
      statusCode: 500,
      contentType: "application/json",
      responseTimeMs: 123,
      error: "",
      expectedStatuses: ["200"],
      expectedContentTypes: ["application/json"]
    },
    ["evidence-id"]
  );

  assert.ok(findings.some((finding) => finding.title === "API endpoint returned 5xx"));
  assert.ok(findings.some((finding) => finding.title === "API endpoint returned unexpected status code"));
});

test("endpointURL joins base path and endpoint path", () => {
  assert.equal(endpointURL("https://api.example.com/v1", "/health"), "https://api.example.com/v1/health");
});

test("validateTargetURL rejects OpenAPI URL outside allowed hosts", async () => {
  await assert.rejects(() => validateTargetURL("https://evil.example.net/openapi.json", ["api.example.com"], false), /not present in allowed_hosts/);
});

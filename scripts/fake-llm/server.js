const http = require("node:http");

const analysis = {
  executive_summary: "The deterministic Qualora run completed and the observed findings should be reviewed before release.",
  technical_summary: "This fake provider analyzed only the sanitized run summary, findings, and evidence metadata supplied by Qualora.",
  risk_level: "medium",
  likely_causes: ["Application behavior or API responses differed from the expected smoke-test baseline."],
  recommended_actions: ["Review the listed findings and inspect the captured evidence metadata.", "Rerun the smoke test after applying fixes."],
  suggested_next_tests: ["Add targeted regression checks for the affected route or endpoint.", "Run the API/OpenAPI smoke checks against the same project."],
  confidence: 0.76,
  limitations: ["Fake provider output is deterministic.", "No screenshots, full HTML, cookies, credentials, or response bodies were analyzed."]
};

const testPlan = {
  title: "Qualora deterministic alpha test plan",
  summary: "A conservative reviewable plan generated from sanitized project and run metadata.",
  assumptions: [
    "Only sanitized project configuration, findings, and evidence metadata were available.",
    "Generated steps are suggestions and are not executed automatically by Qualora."
  ],
  coverage_goals: [
    "Confirm the primary frontend or API smoke path remains reachable.",
    "Review deterministic findings and evidence metadata for regression candidates.",
    "Prioritize safe, non-destructive checks suitable for an alpha release."
  ],
  scenarios: [
    {
      id: "scenario-01",
      name: "Baseline target availability",
      type: "smoke",
      priority: "high",
      risk: "medium",
      description: "Verify that configured frontend and API targets respond successfully from the Qualora environment.",
      preconditions: ["Project targets and allowed hosts are configured.", "A completed Qualora run is available when possible."],
      steps: [
        {
          order: 1,
          action: "Open the project target through the existing browser or API smoke workflow.",
          target: "Configured frontend URL or API base URL",
          data: "",
          expected_result: "The target is reachable and does not return an obvious 5xx response."
        },
        {
          order: 2,
          action: "Review collected findings and evidence metadata.",
          target: "Qualora run report",
          data: "",
          expected_result: "Findings are understood and mapped to concrete follow-up checks."
        }
      ],
      assertions: ["No unexpected critical availability finding is present.", "Evidence metadata is present for the executed worker."],
      test_data_needed: ["A safe test project target."],
      automation_candidate: true,
      destructive: false,
      requires_authentication: false,
      related_findings: [],
      tags: ["alpha", "safe", "smoke"]
    },
    {
      id: "scenario-02",
      name: "API contract visibility",
      type: "api",
      priority: "medium",
      risk: "medium",
      description: "Review safe OpenAPI-derived checks and confirm documented read-only endpoints behave consistently.",
      preconditions: ["An OpenAPI URL is configured or API observation evidence exists."],
      steps: [
        {
          order: 1,
          action: "Inspect the OpenAPI summary evidence.",
          target: "openapi_summary evidence",
          data: "",
          expected_result: "Safe methods and skipped unsafe methods are clearly reported."
        },
        {
          order: 2,
          action: "Compare failed endpoints with declared responses.",
          target: "api_observations evidence",
          data: "",
          expected_result: "Unexpected status codes are identified for manual review."
        }
      ],
      assertions: ["Unsafe methods are not called by default.", "5xx and unexpected safe-method statuses are reviewable."],
      test_data_needed: ["Published OpenAPI document when available."],
      automation_candidate: true,
      destructive: false,
      requires_authentication: false,
      related_findings: [],
      tags: ["api", "openapi", "safe-methods"]
    }
  ],
  suggested_next_instrumentation: [
    "Add authenticated test account support before planning logged-in journeys.",
    "Capture richer endpoint labels and page metadata to improve future plans."
  ],
  limitations: [
    "The fake provider does not inspect screenshots, raw traces, cookies, credentials, full HTML, or response bodies.",
    "This plan is deterministic test data for smoke validation."
  ]
};

const server = http.createServer((req, res) => {
  if (req.method === "GET" && (req.url === "/health" || req.url === "/")) {
    writeJSON(res, 200, { status: "ok" });
    return;
  }

  if (req.method === "POST" && req.url === "/v1/chat/completions") {
    readBody(req)
      .then((body) => {
        const request = body ? JSON.parse(body) : {};
        const content = Array.isArray(request.messages)
          ? request.messages.map((message) => String(message.content || "")).join("\n").toLowerCase()
          : "";
        const isTestPlan = content.includes("test plan") || content.includes("test planning");
        const payload = isTestPlan ? testPlan : analysis;
        writeJSON(res, 200, {
          id: "chatcmpl-qualora-fake",
          object: "chat.completion",
          model: request.model || "qualora-fake-analyst",
          choices: [
            {
              index: 0,
              message: {
                role: "assistant",
                content: JSON.stringify(payload)
              },
              finish_reason: "stop"
            }
          ],
          usage: {
            prompt_tokens: 120,
            completion_tokens: isTestPlan ? 240 : 90,
            total_tokens: isTestPlan ? 360 : 210
          }
        });
      })
      .catch(() => writeJSON(res, 400, { error: "invalid_json" }));
    return;
  }

  writeJSON(res, 404, { error: "not_found" });
});

function readBody(req) {
  return new Promise((resolve, reject) => {
    let body = "";
    req.setEncoding("utf8");
    req.on("data", (chunk) => {
      body += chunk;
      if (body.length > 1024 * 1024) {
        reject(new Error("request_too_large"));
        req.destroy();
      }
    });
    req.on("end", () => resolve(body));
    req.on("error", reject);
  });
}

function writeJSON(res, statusCode, payload) {
  res.writeHead(statusCode, { "content-type": "application/json" });
  res.end(JSON.stringify(payload));
}

server.listen(8080, "0.0.0.0", () => {
  process.stdout.write("qualora fake llm listening on 8080\n");
});

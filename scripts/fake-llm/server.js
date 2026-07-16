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
    "Generated steps are suggestions until a user approves safe execution."
  ],
  coverage_goals: [
    "Confirm the demo homepage renders stable public text.",
    "Confirm same-origin public links are visible and reachable.",
    "Prioritize safe, non-destructive checks suitable for an alpha release."
  ],
  scenarios: [
    {
      id: "scenario-01",
      name: "Homepage public smoke checks",
      type: "smoke",
      priority: "high",
      risk: "medium",
      description: "Verify that the public demo homepage renders stable content and same-origin navigation.",
      preconditions: ["The project frontend URL points at the Qualora demo web target."],
      steps: [
        {
          order: 1,
          action: "goto",
          target: "/",
          data: "",
          expected_result: "The homepage loads."
        },
        {
          order: 2,
          action: "assert_title_contains",
          target: "Qualora Demo Web",
          data: "",
          expected_result: "The page title identifies the demo web target."
        },
        {
          order: 3,
          action: "assert_text_visible",
          target: "Self-hosted QA automation demo",
          data: "",
          expected_result: "The demo description is visible."
        },
        {
          order: 4,
          action: "assert_link_exists",
          target: "/status",
          data: "",
          expected_result: "The status link exists on the page."
        },
        {
          order: 5,
          action: "capture_screenshot",
          target: "",
          data: "",
          expected_result: "Screenshot evidence is captured."
        },
        {
          order: 6,
          action: "collect_browser_signals",
          target: "",
          data: "",
          expected_result: "Browser observations are recorded."
        }
      ],
      assertions: ["The homepage title and body text are visible.", "Same-origin links are discoverable."],
      test_data_needed: [],
      automation_candidate: true,
      destructive: false,
      requires_authentication: false,
      related_findings: [],
      tags: ["alpha", "safe", "smoke", "browser"]
    },
    {
      id: "scenario-02",
      name: "Status page public checks",
      type: "regression",
      priority: "medium",
      risk: "medium",
      description: "Verify the public status page text and same-origin about link without any mutating interaction.",
      preconditions: ["The status page is linked from the homepage."],
      steps: [
        {
          order: 1,
          action: "goto",
          target: "/status",
          data: "",
          expected_result: "The status page loads."
        },
        {
          order: 2,
          action: "assert_url_contains",
          target: "/status",
          data: "",
          expected_result: "The browser remains on the status route."
        },
        {
          order: 3,
          action: "assert_text_visible",
          target: "System status: OK",
          data: "",
          expected_result: "The status text is visible."
        },
        {
          order: 4,
          action: "assert_link_exists",
          target: "/about",
          data: "",
          expected_result: "The about link exists on the page."
        },
        {
          order: 5,
          action: "check_link_status",
          target: "/about",
          data: "",
          expected_result: "The about link returns a successful status."
        },
        {
          order: 6,
          action: "assert_no_console_errors",
          target: "",
          data: "",
          expected_result: "No console errors are observed."
        },
        {
          order: 7,
          action: "assert_no_failed_requests",
          target: "",
          data: "",
          expected_result: "No failed network requests are observed."
        }
      ],
      assertions: ["The status route remains public and readable.", "Same-origin link checks use safe methods only."],
      test_data_needed: [],
      automation_candidate: true,
      destructive: false,
      requires_authentication: false,
      related_findings: [],
      tags: ["alpha", "safe", "regression", "browser"]
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

const discoveryTestPlan = {
  title: "Discovery-aware safe QA plan",
  summary: "A conservative plan generated from a sanitized Qualora discovery map using only safe executable browser DSL actions.",
  assumptions: [
    "Only pages and links from the provided discovery map were considered.",
    "Credentials, cookies, browser storage, screenshots, full HTML, and response bodies were not available to the fake provider."
  ],
  coverage_goals: [
    "Exercise discovered public routes with safe navigation and assertions.",
    "Check discovered same-origin links without submitting forms or mutating state.",
    "Collect browser observations for discovered pages."
  ],
  scenarios: [
    {
      id: "discovery-home-smoke",
      name: "Discovered home route safe smoke",
      type: "smoke",
      priority: "high",
      risk: "medium",
      description: "Verify a discovered public entry route loads and exposes stable text.",
      preconditions: ["A completed Qualora discovery map includes the public home route."],
      steps: [
        { order: 1, action: "goto", target: "/", data: "", expected_result: "The discovered home route loads." },
        { order: 2, action: "assert_title_contains", target: "Qualora Demo Web", data: "", expected_result: "The title matches the discovered demo application." },
        { order: 3, action: "assert_text_visible", target: "Self-hosted QA automation demo", data: "", expected_result: "The discovered page text is visible." },
        { order: 4, action: "collect_browser_signals", target: "", data: "", expected_result: "Console and network observations are collected." }
      ],
      assertions: ["The discovered home route is reachable and renders expected public text."],
      test_data_needed: [],
      automation_candidate: true,
      destructive: false,
      requires_authentication: false,
      related_findings: [],
      tags: ["generated_from_discovery", "safe_executable_candidate", "smoke"]
    },
    {
      id: "discovery-status-link",
      name: "Discovered status link remains safe and reachable",
      type: "regression",
      priority: "medium",
      risk: "medium",
      description: "Verify a discovered same-origin status route and link target with safe read-only actions.",
      preconditions: ["The discovery map includes /status or a same-origin status link."],
      steps: [
        { order: 1, action: "goto", target: "/status", data: "", expected_result: "The status route loads." },
        { order: 2, action: "assert_url_contains", target: "/status", data: "", expected_result: "The browser remains on the discovered status route." },
        { order: 3, action: "assert_text_visible", target: "System status: OK", data: "", expected_result: "The status text is visible." },
        { order: 4, action: "check_link_status", target: "/about", data: "", expected_result: "A discovered same-origin link responds successfully." },
        { order: 5, action: "assert_no_console_errors", target: "", data: "", expected_result: "No console errors are observed." },
        { order: 6, action: "assert_no_failed_requests", target: "", data: "", expected_result: "No failed network requests are observed." }
      ],
      assertions: ["The discovered status route is reachable.", "Browser observations stay clean for the safe check."],
      test_data_needed: [],
      automation_candidate: true,
      destructive: false,
      requires_authentication: false,
      related_findings: [],
      tags: ["generated_from_discovery", "safe_executable_candidate", "regression"]
    }
  ],
  suggested_next_instrumentation: [
    "Add semantic route labels to the discovery map for richer future planning.",
    "Review skipped links before broadening safe QA coverage."
  ],
  limitations: [
    "The fake provider only returns deterministic discovery-aware smoke data.",
    "No authenticated, mutating, or arbitrary form-submission flows are generated."
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
        const isDiscoveryAware = /"discovery_map"\s*:\s*\{/.test(content) || content.includes("generated_from_discovery");
        const payload = isTestPlan ? (isDiscoveryAware ? discoveryTestPlan : testPlan) : analysis;
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

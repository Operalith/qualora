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

const server = http.createServer((req, res) => {
  if (req.method === "GET" && (req.url === "/health" || req.url === "/")) {
    writeJSON(res, 200, { status: "ok" });
    return;
  }

  if (req.method === "POST" && req.url === "/v1/chat/completions") {
    readBody(req)
      .then((body) => {
        const request = body ? JSON.parse(body) : {};
        writeJSON(res, 200, {
          id: "chatcmpl-qualora-fake",
          object: "chat.completion",
          model: request.model || "qualora-fake-analyst",
          choices: [
            {
              index: 0,
              message: {
                role: "assistant",
                content: JSON.stringify(analysis)
              },
              finish_reason: "stop"
            }
          ],
          usage: {
            prompt_tokens: 120,
            completion_tokens: 90,
            total_tokens: 210
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

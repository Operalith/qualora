const { readFileSync } = require("node:fs");
const http = require("node:http");
const { join } = require("node:path");

const port = Number(process.env.PORT || "8080");
const openapi = readFileSync(join(__dirname, "openapi.yaml"), "utf8");
const users = [
  { id: "1", name: "Ada Lovelace", role: "qa-lead" },
  { id: "2", name: "Grace Hopper", role: "api-reviewer" }
];

const server = http.createServer((req, res) => {
  const url = new URL(req.url || "/", `http://${req.headers.host || "localhost"}`);

  if (req.method === "GET" && (url.pathname === "/" || url.pathname === "/health")) {
    writeJSON(res, 200, { status: "ok" });
    return;
  }

  if (req.method === "GET" && url.pathname === "/status") {
    writeJSON(res, 200, { service: "demo-api", ready: true, mode: "deterministic" });
    return;
  }

  if (req.method === "GET" && url.pathname === "/users") {
    writeJSON(res, 200, { users });
    return;
  }

  const userMatch = url.pathname.match(/^\/users\/([^/]+)$/);
  if (req.method === "GET" && userMatch) {
    const user = users.find((item) => item.id === userMatch[1]);
    if (!user) {
      writeJSON(res, 404, { error: "not_found" });
      return;
    }
    writeJSON(res, 200, user);
    return;
  }

  if (req.method === "GET" && url.pathname === "/broken") {
    writeJSON(res, 500, { error: "deterministic_failure" });
    return;
  }

  if (req.method === "GET" && url.pathname === "/profile") {
    writeJSON(res, 401, { error: "authentication_required" });
    return;
  }

  if (req.method === "GET" && url.pathname === "/openapi.yaml") {
    res.writeHead(200, { "content-type": "application/yaml; charset=utf-8" });
    res.end(openapi);
    return;
  }

  writeJSON(res, 404, { error: "not_found" });
});

function writeJSON(res, statusCode, payload) {
  res.writeHead(statusCode, { "content-type": "application/json; charset=utf-8" });
  res.end(JSON.stringify(payload));
}

server.listen(port, "0.0.0.0", () => {
  process.stdout.write(`qualora demo api listening on ${port}\n`);
});

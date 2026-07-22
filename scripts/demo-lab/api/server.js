const { readFileSync } = require("node:fs");
const http = require("node:http");
const { join } = require("node:path");

const port = Number(process.env.PORT || "8080");
const mode = process.env.DEMO_LAB_MODE === "regressed" ? "regressed" : "baseline";
const openapi = readFileSync(join(__dirname, "openapi.yaml"), "utf8");
const acceptedTokens = new Map([
  ["demo-api-token", "demo"],
  ["demo-api-token-admin", "admin"],
  ["demo-api-token-readonly", "readonly"],
  ["demo-api-token-customer-a", "customer-a"],
  ["demo-api-token-customer-b", "customer-b"]
]);
const catalog = [
  { id: "book-1", name: "Reliable Systems", category: "books", price: 42 },
  { id: "tool-1", name: "QA Checklist", category: "tools", price: 12 }
];
const orders = [
  { id: "ord-1", status: "paid", total: 42.5 },
  { id: "ord-2", status: "processing", total: 19.99 }
];
const legacyUsers = [
  { id: "1", name: "Ada Lovelace", role: "qa-lead" },
  { id: "2", name: "Grace Hopper", role: "api-reviewer" }
];

const server = http.createServer((request, response) => {
  const url = new URL(request.url || "/", `http://${request.headers.host || "demo-lab-api"}`);

  if (request.method === "GET" && (url.pathname === "/" || url.pathname === "/health")) {
    writeJSON(response, 200, { status: "ok", service: "demo-lab-api", mode });
    return;
  }
  if (request.method === "GET" && url.pathname === "/openapi.yaml") {
    response.writeHead(200, { "content-type": "application/yaml; charset=utf-8", "x-powered-by": "Qualora Demo Lab API" });
    response.end(openapi);
    return;
  }
  if (request.method === "GET" && url.pathname === "/public/health") {
    writeJSON(response, 200, { status: "ok", visibility: "public" });
    return;
  }
  if (request.method === "GET" && url.pathname === "/public/catalog") {
    writeJSON(response, 200, { items: catalog });
    return;
  }
  if (request.method === "GET" && url.pathname === "/public/status") {
    writeJSON(response, 200, { service: "demo-lab-api", ready: true, mode });
    return;
  }
  if (request.method === "GET" && url.pathname === "/public/search") {
    const query = String(url.searchParams.get("q") || "").slice(0, 80);
    if (!query) {
      writeJSON(response, 400, { error: "query_required", message: "q is required" });
      return;
    }
    writeJSON(response, 200, { query, items: catalog.filter((item) => item.name.toLowerCase().includes(query.toLowerCase())) });
    return;
  }
  if (request.method === "GET" && url.pathname === "/status") {
    writeJSON(response, 200, { service: "demo-lab-api", ready: true, mode });
    return;
  }
  if (request.method === "GET" && url.pathname === "/users") {
    writeJSON(response, 200, { users: legacyUsers });
    return;
  }
  const legacyUserMatch = url.pathname.match(/^\/users\/([^/]+)$/);
  if (request.method === "GET" && legacyUserMatch) {
    const user = legacyUsers.find((item) => item.id === legacyUserMatch[1]);
    writeJSON(response, user ? 200 : 404, user || { error: "not_found" });
    return;
  }
  if (request.method === "GET" && url.pathname === "/broken") {
    writeJSON(response, 500, { error: "deterministic_failure" });
    return;
  }
  if (request.method === "GET" && url.pathname === "/profile") {
    writeJSON(response, 401, { error: "authentication_required" });
    return;
  }

  const role = bearerRole(request);
  if (url.pathname.startsWith("/private/") && !role) {
    writeJSON(response, 401, { error: "authentication_required" });
    return;
  }
  if (request.method === "GET" && url.pathname === "/private/profile") {
    writeJSON(response, 200, { id: "demo-user", email: "demo@example.test", role });
    return;
  }
  if (request.method === "GET" && url.pathname === "/private/orders") {
    writeJSON(response, 200, { orders });
    return;
  }
  const orderMatch = url.pathname.match(/^\/private\/orders\/([^/]+)$/);
  if (request.method === "GET" && orderMatch) {
    const order = orders.find((item) => item.id === orderMatch[1]);
    writeJSON(response, order ? 200 : 404, order || { error: "not_found", message: "Order not found" });
    return;
  }
  if (request.method === "GET" && url.pathname === "/private/broken-contract") {
    writeJSON(response, 200, { id: "demo-user" });
    return;
  }
  if (request.method === "GET" && url.pathname === "/private/server-error") {
    writeJSON(response, 500, { error: "deterministic_private_failure", message: "Intentional Demo Lab server error" });
    return;
  }
  if (request.method === "GET" && url.pathname === "/public/regression" && mode === "regressed") {
    writeJSON(response, 500, { error: "regressed_mode_failure" });
    return;
  }
  if (["POST", "PUT", "PATCH", "DELETE"].includes(request.method || "")) {
    writeJSON(response, 405, { error: "demo_mutation_disabled" });
    return;
  }

  writeJSON(response, 404, { error: "not_found" });
});

function bearerRole(request) {
  const header = String(request.headers.authorization || "");
  if (!header.startsWith("Bearer ")) return "";
  return acceptedTokens.get(header.slice("Bearer ".length)) || "";
}

function writeJSON(response, statusCode, payload) {
  response.writeHead(statusCode, { "content-type": "application/json; charset=utf-8", "x-powered-by": "Qualora Demo Lab API" });
  response.end(JSON.stringify(payload));
}

server.listen(port, "0.0.0.0", () => {
  process.stdout.write(`qualora demo lab api listening on ${port} in ${mode} mode\n`);
});

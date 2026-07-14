import { createReadStream, existsSync, statSync } from "node:fs";
import { createServer } from "node:http";
import { extname, join, normalize } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(fileURLToPath(new URL(".", import.meta.url)), "dist");
const port = Number(process.env.PORT || "3000");
const apiBaseUrl = process.env.QUALORA_API_BASE_URL || "http://localhost:8080";

const contentTypes = new Map([
  [".html", "text/html; charset=utf-8"],
  [".js", "text/javascript; charset=utf-8"],
  [".css", "text/css; charset=utf-8"],
  [".json", "application/json; charset=utf-8"],
  [".svg", "image/svg+xml"],
  [".ico", "image/x-icon"]
]);

createServer((request, response) => {
  const url = new URL(request.url || "/", `http://${request.headers.host || "localhost"}`);

  if (url.pathname === "/healthz") {
    response.writeHead(200, { "Content-Type": "application/json; charset=utf-8" });
    response.end(JSON.stringify({ status: "ok" }));
    return;
  }

  if (url.pathname === "/config.js") {
    response.writeHead(200, {
      "Content-Type": "text/javascript; charset=utf-8",
      "Cache-Control": "no-store"
    });
    response.end(`window.__QUALORA_CONFIG__ = ${JSON.stringify({ apiBaseUrl })};\n`);
    return;
  }

  const requested = normalize(url.pathname).replace(/^(\.\.(\/|\\|$))+/, "");
  let filePath = join(root, requested === "/" ? "index.html" : requested);
  if (!filePath.startsWith(root)) {
    response.writeHead(403);
    response.end("Forbidden");
    return;
  }

  if (!existsSync(filePath) || !statSync(filePath).isFile()) {
    filePath = join(root, "index.html");
  }

  response.writeHead(200, {
    "Content-Type": contentTypes.get(extname(filePath)) || "application/octet-stream"
  });
  createReadStream(filePath).pipe(response);
}).listen(port, "0.0.0.0", () => {
  process.stdout.write(`qualora-web listening on ${port}\n`);
});

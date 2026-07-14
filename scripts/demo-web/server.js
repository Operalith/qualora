const http = require("node:http");

const port = Number(process.env.PORT || "8080");

const server = http.createServer((request, response) => {
  const url = new URL(request.url, "http://demo-web");

  if (url.pathname === "/health") {
    response.writeHead(200, { "content-type": "application/json" });
    response.end(JSON.stringify({ status: "ok" }));
    return;
  }

  if (url.pathname === "/app.js") {
    response.writeHead(200, { "content-type": "application/javascript" });
    response.end("document.documentElement.dataset.qualoraDemo='ready';");
    return;
  }

  if (url.pathname === "/status") {
    writePage(response, "Qualora Demo Web Status", "System status: OK", "Status information for the deterministic Qualora demo web target.");
    return;
  }

  if (url.pathname === "/about") {
    writePage(response, "About Qualora Demo Web", "About Qualora", "This page gives the safe test plan executor stable public text to verify.");
    return;
  }

  if (url.pathname !== "/") {
    response.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    response.end("Not found");
    return;
  }

  writePage(
    response,
    "Qualora Demo Web",
    "Qualora Demo Web",
    "Self-hosted QA automation demo for browser smoke tests and approved safe test plan execution."
  );
});

function writePage(response, title, heading, body) {
  response.writeHead(200, { "content-type": "text/html; charset=utf-8" });
  response.end(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${escapeHTML(title)}</title>
  <style>
    body {
      margin: 0;
      background: #f5f7fa;
      color: #17202c;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    main {
      display: grid;
      gap: 12px;
      margin: 0 auto;
      max-width: 760px;
      padding: 64px 24px;
    }
    h1 {
      font-size: 36px;
      line-height: 1.15;
      margin: 0;
    }
    p {
      color: #667085;
      font-size: 16px;
      margin: 0;
    }
    nav {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
    }
    a {
      color: #0d5b57;
      font-weight: 700;
    }
  </style>
</head>
<body>
  <main>
    <nav aria-label="Demo navigation">
      <a href="/">Home</a>
      <a href="/status">Status</a>
      <a href="/about">About</a>
    </nav>
    <h1>${escapeHTML(heading)}</h1>
    <p>${escapeHTML(body)}</p>
  </main>
  <script src="/app.js"></script>
</body>
</html>`);
}

function escapeHTML(value) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

server.listen(port, "0.0.0.0", () => {
  console.log(`qualora demo web listening on ${port}`);
});

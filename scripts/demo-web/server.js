const http = require("node:http");

const port = Number(process.env.PORT || "8080");

const server = http.createServer((request, response) => {
  if (request.url === "/health") {
    response.writeHead(200, { "content-type": "application/json" });
    response.end(JSON.stringify({ status: "ok" }));
    return;
  }

  if (request.url === "/app.js") {
    response.writeHead(200, { "content-type": "application/javascript" });
    response.end("document.documentElement.dataset.qualoraDemo='ready';");
    return;
  }

  response.writeHead(200, { "content-type": "text/html; charset=utf-8" });
  response.end(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora Demo Web</title>
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
  </style>
</head>
<body>
  <main>
    <h1>Qualora Demo Web</h1>
    <p>This deterministic local page is used by the browser smoke test.</p>
  </main>
  <script src="/app.js"></script>
</body>
</html>`);
});

server.listen(port, "0.0.0.0", () => {
  console.log(`qualora demo web listening on ${port}`);
});

const http = require("node:http");

const port = Number(process.env.PORT || "8080");
const demoUsername = "demo@example.com";
const demoPassword = "demo-password";

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

  if (url.pathname === "/login" && request.method === "GET") {
    writeLoginPage(response, "");
    return;
  }

  if (url.pathname === "/login" && request.method === "POST") {
    readForm(request)
      .then((form) => {
        if (form.get("username") === demoUsername && form.get("password") === demoPassword) {
          response.writeHead(303, {
            "set-cookie": "qualora_demo_session=authenticated; HttpOnly; SameSite=Lax; Path=/",
            location: "/dashboard"
          });
          response.end();
          return;
        }
        writeLoginPage(response, "Invalid credentials");
      })
      .catch(() => {
        response.writeHead(400, { "content-type": "text/plain; charset=utf-8" });
        response.end("Bad request");
      });
    return;
  }

  if (url.pathname === "/dashboard") {
    if (!isAuthenticated(request)) {
      response.writeHead(302, { location: "/login" });
      response.end();
      return;
    }
    writePage(
      response,
      "Qualora Demo Dashboard",
      "Welcome to the Qualora demo dashboard",
      "Authenticated area"
    );
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

function writeLoginPage(response, error) {
  response.writeHead(error ? 401 : 200, { "content-type": "text/html; charset=utf-8" });
  response.end(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora Demo Login</title>
  <style>
    body {
      margin: 0;
      background: #f5f7fa;
      color: #17202c;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    main {
      display: grid;
      gap: 14px;
      margin: 0 auto;
      max-width: 440px;
      padding: 64px 24px;
    }
    label {
      display: grid;
      gap: 6px;
      font-weight: 700;
    }
    input {
      border: 1px solid #c8d1dc;
      border-radius: 6px;
      font: inherit;
      padding: 10px 12px;
    }
    button {
      background: #0d5b57;
      border: 0;
      border-radius: 6px;
      color: white;
      cursor: pointer;
      font: inherit;
      font-weight: 700;
      padding: 11px 14px;
    }
    .error {
      color: #b42318;
      font-weight: 700;
    }
  </style>
</head>
<body>
  <main>
    <h1>Qualora Demo Login</h1>
    ${error ? `<p class="error">${escapeHTML(error)}</p>` : ""}
    <form method="post" action="/login">
      <label>
        Username
        <input id="username" name="username" type="email" autocomplete="username" required>
      </label>
      <label>
        Password
        <input id="password" name="password" type="password" autocomplete="current-password" required>
      </label>
      <button id="login-submit" type="submit">Sign in</button>
    </form>
  </main>
</body>
</html>`);
}

function readForm(request) {
  return new Promise((resolve, reject) => {
    let body = "";
    request.on("data", (chunk) => {
      body += chunk;
      if (body.length > 8192) {
        reject(new Error("form body too large"));
        request.destroy();
      }
    });
    request.on("end", () => resolve(new URLSearchParams(body)));
    request.on("error", reject);
  });
}

function isAuthenticated(request) {
  const cookie = request.headers.cookie || "";
  return cookie.split(";").map((part) => part.trim()).includes("qualora_demo_session=authenticated");
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

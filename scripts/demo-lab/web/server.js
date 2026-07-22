const http = require("node:http");

const port = Number(process.env.PORT || "8080");
const mode = process.env.DEMO_LAB_MODE === "regressed" ? "regressed" : "baseline";
const accounts = {
  "admin@example.com": { password: "admin-password", role: "admin", label: "Demo Lab Admin" },
  "readonly@example.com": { password: "readonly-password", role: "readonly", label: "Demo Lab Readonly" },
  "customer-a@example.com": { password: "customer-a-password", role: "customer-a", label: "Customer A" },
  "customer-b@example.com": { password: "customer-b-password", role: "customer-b", label: "Customer B" }
};
const safeHeadPaths = new Set([
  "/",
  "/about",
  "/status",
  "/pricing",
  "/search",
  "/products",
  "/quality/console-error",
  "/quality/broken-asset",
  "/quality/a11y",
  "/quality/missing-headers",
  "/quality/slow"
]);

const server = http.createServer((request, response) => {
  const url = new URL(request.url || "/", `http://${request.headers.host || "demo-lab-web"}`);

  if (request.method === "HEAD" && safeHeadPaths.has(url.pathname)) {
    response.writeHead(200, pageHeaders(url.pathname, { "content-type": "text/html; charset=utf-8" }));
    response.end();
    return;
  }

  if (request.method === "GET" && url.pathname === "/health") {
    writeJSON(response, 200, { status: "ok", service: "demo-lab-web", mode });
    return;
  }

  if (request.method === "GET" && url.pathname === "/assets/lab-mark.svg") {
    response.writeHead(200, pageHeaders(url.pathname, { "content-type": "image/svg+xml; charset=utf-8" }));
    response.end('<svg xmlns="http://www.w3.org/2000/svg" width="72" height="72" viewBox="0 0 72 72"><rect width="72" height="72" rx="14" fill="#173f3a"/><path d="M25 17h22v7l-6 8v8l13 17H18l13-17v-8l-6-8z" fill="#d9f99d"/><path d="M27 48h18" stroke="#173f3a" stroke-width="4" stroke-linecap="round"/></svg>');
    return;
  }

  if (request.method === "GET" && url.pathname === "/assets/app.js") {
    response.writeHead(200, pageHeaders(url.pathname, { "content-type": "application/javascript; charset=utf-8" }));
    response.end("document.documentElement.dataset.demoLab='ready';\n//# sourceMappingURL=/assets/app.js.map");
    return;
  }

  if (request.method === "GET" && url.pathname === "/assets/app.js.map") {
    writeJSON(response, 200, { version: 3, sources: ["app.js"], names: [], mappings: "" });
    return;
  }

  if (request.method === "GET" && url.pathname === "/login") {
    writeLoginPage(response, "");
    return;
  }

  if (request.method === "POST" && url.pathname === "/login") {
    readForm(request)
      .then((form) => {
        const email = String(form.get("email") || "");
        const account = accounts[email];
        if (!account || form.get("password") !== account.password) {
          writeLoginPage(response, "Invalid credentials");
          return;
        }
        response.writeHead(303, pageHeaders(url.pathname, {
          location: "/dashboard",
          "set-cookie": `demo_lab_session=${encodeURIComponent(account.role)}; HttpOnly; SameSite=Lax; Path=/`
        }));
        response.end();
      })
      .catch(() => writeText(response, 400, "Bad request"));
    return;
  }

  if (request.method === "GET" && url.pathname === "/dashboard") {
    const account = requireLogin(request, response);
    if (!account) return;
    writePage(response, url.pathname, "Qualora Demo Lab Dashboard", "Welcome to Demo Lab", `Authenticated area for ${account.label} (${account.role}).`, { account });
    return;
  }

  if (request.method === "GET" && url.pathname === "/admin") {
    const account = requireRole(request, response, ["admin"]);
    if (!account) return;
    writePage(response, url.pathname, "Demo Lab Admin", "Admin console", "Administrative settings used for deterministic authorization checks.", { account });
    return;
  }

  if (request.method === "GET" && url.pathname === "/reports") {
    const account = requireRole(request, response, ["admin", "readonly"]);
    if (!account) return;
    writePage(response, url.pathname, "Demo Lab Reports", "Reports center", "Readonly report content for admin and readonly roles.", { account });
    return;
  }

  if (request.method === "GET" && url.pathname === "/customers/a/invoice") {
    const account = requireRole(request, response, ["admin", "customer-a"]);
    if (!account) return;
    writePage(response, url.pathname, "Customer A Invoice", "Invoice for Customer A", "Customer A invoice total: $42.00.", { account });
    return;
  }

  if (request.method === "GET" && url.pathname === "/customers/b/invoice") {
    const account = requireRole(request, response, ["admin", "customer-b"]);
    if (!account) return;
    writePage(response, url.pathname, "Customer B Invoice", "Invoice for Customer B", "Customer B invoice total: $84.00.", { account });
    return;
  }

  if (request.method === "GET" && url.pathname === "/logout") {
    response.writeHead(200, pageHeaders(url.pathname, {
      "content-type": "text/html; charset=utf-8",
      "set-cookie": "demo_lab_session=; Max-Age=0; HttpOnly; SameSite=Lax; Path=/"
    }));
    response.end(simplePage("Signed out", "This route is intentionally classified as unsafe by Qualora explorers."));
    return;
  }

  if (request.method === "GET" && url.pathname === "/search") {
    const query = cleanValue(url.searchParams.get("q"), 80);
    writePage(response, url.pathname, "Demo Lab Search", query ? `Search results for ${query}` : "Search", query ? `Found deterministic showcase result for ${query}.` : "Use the safe GET search form to find demo content.");
    return;
  }

  if (request.method === "GET" && url.pathname === "/products") {
    const category = cleanValue(url.searchParams.get("category"), 80) || "all";
    const sort = cleanValue(url.searchParams.get("sort"), 80);
    const detail = sort ? `Products in ${category}, sorted by ${sort}.` : `Products in ${category}`;
    writePage(response, url.pathname, "Demo Lab Products", `Products in ${category}`, detail);
    return;
  }

  if (request.method === "GET" && url.pathname === "/quality/console-error") {
    writePage(response, url.pathname, "Demo Lab Console Fixture", "Console error fixture", "This page emits one deterministic console error.", { consoleError: true });
    return;
  }

  if (request.method === "GET" && url.pathname === "/quality/broken-asset") {
    writePage(response, url.pathname, "Demo Lab Broken Asset", "Broken asset fixture", "This page requests one missing same-origin script.", { brokenAsset: true });
    return;
  }

  if (request.method === "GET" && url.pathname === "/quality/a11y") {
    writePage(response, url.pathname, "Demo Lab Accessibility Fixture", "Accessibility fixture", "This page contains obvious, deterministic accessibility metadata gaps.", { accessibilityIssues: true });
    return;
  }

  if (request.method === "GET" && url.pathname === "/quality/missing-headers") {
    writePage(response, url.pathname, "Demo Lab Header Fixture", "Missing headers fixture", "This route intentionally omits selected passive response security headers.");
    return;
  }

  if (request.method === "GET" && url.pathname === "/quality/slow") {
    setTimeout(() => {
      writePage(response, url.pathname, "Demo Lab Slow Fixture", "Slow response fixture", "This bounded delay is deterministic and exists only for passive performance checks.");
    }, Number(process.env.DEMO_LAB_SLOW_MS || "1200"));
    return;
  }

  if (request.method === "GET" && url.pathname === "/regression/server-error" && mode === "regressed") {
    writeText(response, 500, "Deterministic regressed-mode server error");
    return;
  }

  if (request.method === "GET" && ["/about", "/status", "/pricing"].includes(url.pathname)) {
    const pages = {
      "/about": ["About Qualora Demo Lab", "About Qualora Demo Lab", "A deterministic target for safe end-to-end Qualora demonstrations."],
      "/status": ["Demo Lab Status", "System status: OK", `All local showcase fixtures are ready in ${mode} mode.`],
      "/pricing": ["Demo Lab Pricing", "Pricing", "Demo Lab is local, open-source showcase infrastructure with no billing or real payments."]
    };
    writePage(response, url.pathname, ...pages[url.pathname]);
    return;
  }

  if (request.method === "GET" && ["/delete-account", "/transfer", "/reset-password"].includes(url.pathname)) {
    writePage(response, url.pathname, "Unsafe Demo Lab Fixture", url.pathname === "/transfer" ? "Transfer money" : "Unsafe action fixture", "Qualora should classify this route as unsafe and skip it.");
    return;
  }

  if (request.method === "POST" && ["/contact", "/delete-account", "/transfer", "/reset-password", "/upload"].includes(url.pathname)) {
    writeText(response, 405, "Intentional demo mutation endpoint disabled");
    return;
  }

  if (request.method === "GET" && url.pathname === "/") {
    writePage(response, url.pathname, "Qualora Demo Web | Qualora Demo Lab", "Qualora Demo Lab", "Self-hosted QA automation demo for realistic, deterministic end-to-end validation.", { home: true, regressed: mode === "regressed" });
    return;
  }

  writeText(response, 404, "Not found");
});

function writePage(response, path, title, heading, body, options = {}) {
  response.writeHead(200, pageHeaders(path, { "content-type": "text/html; charset=utf-8" }));
  response.end(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${escapeHTML(title)}</title>
  <style>
    :root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    * { box-sizing: border-box; }
    body { margin: 0; background: #f4f7f6; color: #18302d; }
    header { align-items: center; background: #173f3a; color: #fff; display: flex; justify-content: space-between; padding: 16px max(24px, calc((100vw - 1120px) / 2)); }
    .brand { align-items: center; color: #fff; display: flex; font-size: 17px; font-weight: 800; gap: 10px; text-decoration: none; }
    .brand img { height: 34px; width: 34px; }
    nav { display: flex; flex-wrap: wrap; gap: 16px; }
    nav a { color: #dff6ee; font-size: 14px; font-weight: 700; }
    main { margin: 0 auto; max-width: 1120px; min-height: calc(100vh - 138px); padding: 54px 24px 72px; }
    .hero { align-items: start; display: grid; gap: 20px; grid-template-columns: minmax(0, 1.5fr) minmax(260px, .7fr); }
    .eyebrow { color: #36736b; font-size: 12px; font-weight: 800; margin: 0 0 10px; text-transform: uppercase; }
    h1 { font-size: clamp(36px, 7vw, 64px); line-height: 1.02; margin: 0; max-width: 760px; }
    .lead { color: #546b67; font-size: 18px; line-height: 1.6; margin: 18px 0 0; max-width: 720px; }
    .status-band { background: #fff; border: 1px solid #d6e1de; border-radius: 8px; display: grid; gap: 14px; padding: 20px; }
    .status-band strong { font-size: 26px; }
    .status-band span { color: #667b77; font-size: 13px; }
    .grid { display: grid; gap: 16px; grid-template-columns: repeat(3, minmax(0, 1fr)); margin-top: 36px; }
    .panel { background: #fff; border: 1px solid #d6e1de; border-radius: 8px; padding: 20px; }
    .panel h2 { font-size: 17px; margin: 0 0 8px; }
    .panel p { color: #667b77; line-height: 1.5; margin: 0; }
    .forms { border-top: 1px solid #d6e1de; display: grid; gap: 20px; grid-template-columns: repeat(2, minmax(0, 1fr)); margin-top: 42px; padding-top: 32px; }
    form { align-content: start; background: #fff; border: 1px solid #d6e1de; border-radius: 8px; display: grid; gap: 12px; padding: 18px; }
    label { display: grid; font-size: 13px; font-weight: 750; gap: 6px; }
    input, select, textarea { border: 1px solid #afbfbb; border-radius: 5px; font: inherit; padding: 9px 10px; }
    button { background: #173f3a; border: 0; border-radius: 5px; color: #fff; cursor: pointer; font: inherit; font-weight: 750; min-height: 38px; padding: 9px 13px; }
    .secondary { background: #e7efed; color: #173f3a; }
    .danger { background: #a83232; }
    .fixture { margin-top: 28px; }
    nav.fixture { line-height: 1.7; }
    nav.fixture a { color: #245f57; }
    footer { border-top: 1px solid #d6e1de; color: #667b77; font-size: 13px; padding: 22px 24px; text-align: center; }
    @media (max-width: 760px) { header { align-items: flex-start; flex-direction: column; gap: 14px; } .hero, .grid, .forms { grid-template-columns: 1fr; } h1 { font-size: 40px; } }
  </style>
</head>
<body>
  <header>
    <a class="brand" href="/"><img src="/assets/lab-mark.svg" alt="">Qualora Demo Lab</a>
    <nav aria-label="Demo Lab navigation">
      <a href="/about">About</a><a href="/status">Status</a><a href="/pricing">Pricing</a><a href="/search?q=demo">Search</a><a href="/products?category=books">Products</a><a href="/login">Login</a>
    </nav>
  </header>
  <main>
    <div class="hero">
      <div><p class="eyebrow">Deterministic showcase target</p><h1>${escapeHTML(heading)}</h1><p class="lead">${escapeHTML(body)}</p></div>
      <div class="status-band"><span>Current fixture mode</span><strong>${escapeHTML(mode)}</strong><span>No real data or external services</span></div>
    </div>
    ${options.home ? homeContent(options.regressed) : ""}
    ${options.accessibilityIssues ? accessibilityFixture() : ""}
    ${options.brokenAsset ? '<script src="/assets/intentionally-missing.js"></script>' : ""}
    ${options.consoleError ? "<script>console.error('Qualora Demo Lab deterministic console error');</script>" : ""}
    ${options.account ? `<div class="panel fixture"><h2>Signed-in role</h2><p>${escapeHTML(options.account.role)}</p></div>` : ""}
  </main>
  <footer>Qualora Demo Lab is a local-only showcase target. Intentional issues are safe and documented.</footer>
  <script src="/assets/app.js"></script>
</body>
</html>`);
}

function homeContent(regressed) {
  return `
    <div class="grid">
      <div class="panel"><h2>Browser workflows</h2><p>Stable public, authenticated, role-aware, and intentionally unsafe routes.</p></div>
      <div class="panel"><h2>Quality fixtures</h2><p>Passive security, accessibility, console, network, and bounded performance signals.</p></div>
      <div class="panel"><h2>API contracts</h2><p>Public and bearer-authenticated OpenAPI operations with one intentional mismatch.</p></div>
    </div>
    <nav class="fixture" aria-label="Discovery and safety fixtures">
      <a href="/quality/console-error">Console error</a>
      <a href="/quality/broken-asset">Broken asset</a>
      <a href="/quality/a11y">Accessibility</a>
      <a href="/quality/missing-headers">Missing headers</a>
      <a href="/quality/slow">Slow page</a>
      <a href="/dashboard">Dashboard</a>
      <a href="/admin">Admin</a>
      <a href="/reports">Reports</a>
      <a href="/delete-account">Delete account</a>
      <a href="/transfer">Transfer money</a>
      <a href="/logout">Logout</a>
      <a href="/reset-password">Reset password</a>
      <a href="https://example.com/external">External docs</a>
      ${regressed ? '<a href="/regression/broken-link">New broken internal link</a><a href="/regression/server-error">New server error</a>' : ""}
    </nav>
    <div class="forms">
      <form id="site-search" method="get" action="/search" aria-label="Site search"><label for="search-query">Search</label><input id="search-query" name="q" type="search" placeholder="Search Demo Lab"><button type="submit">Search</button></form>
      <form id="product-filter" method="get" action="/products" aria-label="Product filter"><label for="product-category">Category</label><select id="product-category" name="category"><option value="all">All</option><option value="books">Books</option><option value="tools">Tools</option></select><button type="submit">Filter</button></form>
      <form id="product-sort" method="get" action="/products" aria-label="Product sort"><label for="product-sort-value">Sort</label><select id="product-sort-value" name="sort"><option value="price">Price</option><option value="name">Name</option></select><button type="submit">Sort</button></form>
      <form id="contact-form" method="post" action="/contact" aria-label="Contact form"><label for="contact-message">Message</label><textarea id="contact-message" name="message"></textarea><button type="submit">Send message</button></form>
      <form id="delete-account-form" method="post" action="/delete-account" aria-label="Delete account"><button class="danger" type="submit">Delete account</button></form>
      <form id="transfer-form" method="post" action="/transfer" aria-label="Transfer money"><label for="amount">Amount</label><input id="amount" name="amount" type="number"><button class="danger" type="submit">Transfer</button></form>
      <form id="external-form" method="get" action="https://example.com/external" aria-label="External search"><label for="external-query">Query</label><input id="external-query" name="q"><button type="submit">External search</button></form>
      <form id="password-reset-form" method="post" action="/reset-password" aria-label="Password reset"><input name="reset_token" type="hidden" value="fixture-only"><label for="new-password">New password</label><input id="new-password" name="password" type="password"><button type="submit">Reset password</button></form>
      <form id="upload-form" method="post" action="/upload" enctype="multipart/form-data" aria-label="File upload"><label for="upload-file">File</label><input id="upload-file" name="file" type="file"><button type="submit">Upload</button></form>
    </div>
    <div class="forms"><button class="secondary" id="open-details" type="button">Open details</button><button class="danger" id="danger-delete" type="button">Delete account</button></div>
    ${regressed ? "<script>console.error('Qualora Demo Lab regressed mode console error');</script>" : ""}`;
}

function accessibilityFixture() {
  return '<div class="panel fixture"><img src="/assets/lab-mark.svg"><input type="text" name="unlabelled"><button type="button" aria-hidden="true"></button><a href="/about"></a></div>';
}

function writeLoginPage(response, error) {
  response.writeHead(error ? 401 : 200, pageHeaders("/login", { "content-type": "text/html; charset=utf-8" }));
  response.end(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Qualora Demo Lab Login</title><style>body{background:#f4f7f6;color:#18302d;font-family:system-ui;margin:0}main{display:grid;gap:16px;margin:60px auto;max-width:430px;padding:24px}form{background:#fff;border:1px solid #d6e1de;border-radius:8px;display:grid;gap:14px;padding:24px}label{display:grid;font-weight:700;gap:6px}input{border:1px solid #afbfbb;border-radius:5px;font:inherit;padding:10px}button{background:#173f3a;border:0;border-radius:5px;color:#fff;font:inherit;font-weight:750;padding:11px}.error{color:#a83232;font-weight:700}</style></head><body><main><a href="/">Qualora Demo Lab</a><h1>Sign in to Demo Lab</h1>${error ? `<p class="error">${escapeHTML(error)}</p>` : ""}<form method="post" action="/login"><label for="username">Email</label><input id="username" name="email" type="email" autocomplete="username" required><label for="password">Password</label><input id="password" name="password" type="password" autocomplete="current-password" required><button id="login-submit" type="submit">Sign in</button></form></main></body></html>`);
}

function requireLogin(request, response) {
  const account = accountFromRequest(request);
  if (account) return account;
  response.writeHead(302, pageHeaders("/dashboard", { location: "/login" }));
  response.end();
  return null;
}

function requireRole(request, response, roles) {
  const account = requireLogin(request, response);
  if (!account) return null;
  if (roles.includes(account.role)) return account;
  response.writeHead(403, pageHeaders("/denied", { "content-type": "text/html; charset=utf-8" }));
  response.end(simplePage("Access denied", `${account.label} (${account.role}) is not allowed to view this resource.`));
  return null;
}

function accountFromRequest(request) {
  const cookie = request.headers.cookie || "";
  const session = cookie.split(";").map((part) => part.trim()).find((part) => part.startsWith("demo_lab_session="));
  if (!session) return null;
  const role = decodeURIComponent(session.split("=").slice(1).join("="));
  return Object.values(accounts).find((account) => account.role === role) || null;
}

function pageHeaders(path, headers = {}) {
  const base = { "x-powered-by": "Qualora Demo Lab", "x-frame-options": "DENY", "referrer-policy": "no-referrer" };
  if (path !== "/quality/missing-headers" && mode !== "regressed") base["x-content-type-options"] = "nosniff";
  return { ...base, ...headers };
}

function writeJSON(response, statusCode, payload) {
  response.writeHead(statusCode, pageHeaders("/api", { "content-type": "application/json; charset=utf-8" }));
  response.end(JSON.stringify(payload));
}

function writeText(response, statusCode, body) {
  response.writeHead(statusCode, pageHeaders("/text", { "content-type": "text/plain; charset=utf-8" }));
  response.end(body);
}

function simplePage(heading, body) {
  return `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>${escapeHTML(heading)}</title></head><body><main><h1>${escapeHTML(heading)}</h1><p>${escapeHTML(body)}</p></main></body></html>`;
}

function readForm(request) {
  return new Promise((resolve, reject) => {
    let body = "";
    request.on("data", (chunk) => {
      body += chunk;
      if (body.length > 8192) reject(new Error("form body too large"));
    });
    request.on("end", () => resolve(new URLSearchParams(body)));
    request.on("error", reject);
  });
}

function cleanValue(value, max) {
  return String(value || "").replace(/[\r\n<>]/g, " ").trim().slice(0, max);
}

function escapeHTML(value) {
  return String(value).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

server.listen(port, "0.0.0.0", () => {
  process.stdout.write(`qualora demo lab web listening on ${port} in ${mode} mode\n`);
});

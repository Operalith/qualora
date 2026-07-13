const http = require("node:http");

const openapi = {
  openapi: "3.0.3",
  info: {
    title: "Qualora Mock API",
    version: "1.0.0"
  },
  paths: {
    "/health": {
      get: {
        responses: {
          "200": {
            description: "Health response",
            content: {
              "application/json": {}
            }
          }
        }
      }
    },
    "/status": {
      get: {
        responses: {
          "200": {
            description: "Status response",
            content: {
              "application/json": {}
            }
          }
        }
      }
    },
    "/items": {
      post: {
        responses: {
          "201": {
            description: "Created item"
          }
        }
      }
    }
  }
};

const server = http.createServer((req, res) => {
  if (req.url === "/health" || req.url === "/") {
    writeJSON(res, 200, { status: "ok" });
    return;
  }
  if (req.url === "/status") {
    writeJSON(res, 200, { service: "mock-api", ready: true });
    return;
  }
  if (req.url === "/openapi.json") {
    writeJSON(res, 200, openapi);
    return;
  }

  writeJSON(res, 404, { error: "not_found" });
});

function writeJSON(res, statusCode, payload) {
  res.writeHead(statusCode, { "content-type": "application/json" });
  res.end(JSON.stringify(payload));
}

server.listen(8080, "0.0.0.0", () => {
  process.stdout.write("qualora mock api listening on 8080\n");
});

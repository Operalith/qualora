FROM node:22-bookworm-slim

WORKDIR /app

COPY scripts/demo-api/server.js ./
COPY scripts/demo-api/openapi.yaml ./

EXPOSE 8080

CMD ["node", "server.js"]

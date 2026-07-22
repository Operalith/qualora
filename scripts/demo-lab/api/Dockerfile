FROM node:22-alpine

WORKDIR /app

COPY scripts/demo-lab/api/server.js ./server.js
COPY scripts/demo-lab/api/openapi.yaml ./openapi.yaml

EXPOSE 8080

CMD ["node", "server.js"]

FROM node:22-bookworm-slim AS build

WORKDIR /app

COPY apps/web/package*.json ./
RUN npm ci

COPY apps/web/ ./
RUN npm run build

FROM node:22-bookworm-slim

WORKDIR /app

ENV PORT=3000
ENV QUALORA_API_BASE_URL=http://localhost:8080

COPY --from=build /app/dist ./dist
COPY apps/web/server.mjs ./server.mjs

EXPOSE 3000

CMD ["node", "server.mjs"]

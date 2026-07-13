FROM node:22-bookworm-slim

WORKDIR /app

COPY workers/api/package*.json ./
RUN npm ci

COPY workers/api/ ./
RUN npm run build

CMD ["npm", "start"]

FROM node:22-bookworm-slim

WORKDIR /app

COPY workers/browser/package*.json ./
RUN npm ci
RUN npx playwright install --with-deps chromium

COPY workers/browser/ ./
RUN npm run build

CMD ["npm", "start"]

.PHONY: dev test lint compose-up compose-down logs smoke

dev: compose-up

test:
	cd apps/control-plane && go test ./...
	cd apps/web && npm ci && npm run build
	cd workers/browser && npm ci && npm test
	cd workers/api && npm ci && npm test

lint:
	cd apps/control-plane && go test ./...
	cd apps/web && npm ci && npm run lint
	cd workers/browser && npm ci && npm run lint
	cd workers/api && npm ci && npm run lint
	docker compose config >/dev/null

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down

logs:
	docker compose logs -f qualora-api qualora-web qualora-worker-browser qualora-worker-api

smoke:
	docker compose --profile smoke up -d --build --force-recreate mock-api demo-web fake-llm
	python3 scripts/smoke.py

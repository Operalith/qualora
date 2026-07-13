.PHONY: dev test lint compose-up compose-down logs smoke

dev: compose-up

test:
	cd apps/control-plane && go test ./...
	cd workers/browser && npm install && npm test

lint:
	cd apps/control-plane && go test ./...
	cd workers/browser && npm install && npm run lint

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down

logs:
	docker compose logs -f qualora-api qualora-worker-browser

smoke:
	python3 scripts/smoke.py

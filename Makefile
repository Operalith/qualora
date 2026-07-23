.PHONY: dev test lint compose-up compose-down logs smoke showcase-smoke demo-lab demo-lab-real-llm

dev: compose-up

test:
	cd apps/control-plane && go test ./...
	cd apps/web && npm ci && npm run build
	cd workers/browser && npm ci && npm test
	cd workers/api && npm ci && npm test
	scripts/test-real-llm-script.sh

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
	docker compose --profile smoke up -d --build --force-recreate demo-api demo-web fake-llm
	python3 scripts/smoke.py

showcase-smoke:
	docker compose --profile demo-lab up -d --build demo-lab-web demo-lab-api fake-llm
	@QUALORA_TARGET_URL=http://demo-lab-web:8080 \
	QUALORA_ALLOWED_HOST=demo-lab-web \
	QUALORA_DEMO_USERNAME=admin@example.com \
	QUALORA_DEMO_PASSWORD=admin-password \
	QUALORA_LOGIN_USERNAME_SELECTOR='input[name="email"]' \
	QUALORA_LOGIN_SUCCESS_TEXT='Welcome to Demo Lab' \
	DEMO_WEB_HEALTH_URL=http://localhost:$${DEMO_LAB_WEB_PORT:-18085}/health \
	QUALORA_API_SMOKE_URL=http://demo-lab-api:8080 \
	QUALORA_API_SMOKE_OPENAPI_URL=http://demo-lab-api:8080/openapi.yaml \
	QUALORA_API_SMOKE_ALLOWED_HOST=demo-lab-api \
	DEMO_API_HEALTH_URL=http://localhost:$${DEMO_LAB_API_PORT:-18086}/health \
	python3 scripts/smoke.py

demo-lab:
	scripts/run-demo-lab.sh

demo-lab-real-llm:
	scripts/run-demo-lab-real-llm.sh

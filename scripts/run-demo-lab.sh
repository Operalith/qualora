#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

cd "$REPO_DIR"

echo "Starting Qualora and the Demo Lab showcase services..."
docker compose --profile demo-lab up -d --build

echo "Running the full deterministic Demo Lab validation..."
make --no-print-directory showcase-smoke

echo
echo "Demo Lab validation completed."
echo "Qualora UI: http://localhost:${QUALORA_WEB_PORT:-3000}"
echo "Demo Lab web: http://localhost:${DEMO_LAB_WEB_PORT:-18085}"
echo "Demo Lab API: http://localhost:${DEMO_LAB_API_PORT:-18086}"
echo "Demo Lab OpenAPI: http://localhost:${DEMO_LAB_API_PORT:-18086}/openapi.yaml"
echo "Run-specific project and report links are printed above."

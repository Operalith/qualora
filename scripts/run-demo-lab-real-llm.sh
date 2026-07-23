#!/usr/bin/env bash
set -euo pipefail

required_vars=(
  QUALORA_REAL_LLM_NAME
  QUALORA_REAL_LLM_BASE_URL
  QUALORA_REAL_LLM_API_KEY
  QUALORA_REAL_LLM_MODEL
)
missing=()
for name in "${required_vars[@]}"; do
  if [[ -z "${!name:-}" ]]; then
    missing+=("${name}")
  fi
done
if (( ${#missing[@]} > 0 )); then
  printf 'Missing required environment variables: %s\n' "${missing[*]}" >&2
  echo "Set the real OpenAI-compatible provider values and run again. No services were started." >&2
  exit 2
fi

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPO_DIR="$(CDPATH= cd -- "${SCRIPT_DIR}/.." && pwd)"
API_URL="${QUALORA_API_URL:-${QUALORA_API_BASE_URL:-http://localhost:${QUALORA_API_PORT:-8080}}}"
API_URL="${API_URL%/}"
WEB_URL="${QUALORA_WEB_URL:-http://localhost:${QUALORA_WEB_PORT:-3000}}"
ADMIN_EMAIL="${QUALORA_ADMIN_EMAIL:-admin@qualora.local}"
ADMIN_PASSWORD="${QUALORA_ADMIN_PASSWORD:-qualora-admin-password}"
PROJECT_NAME="${QUALORA_DEMO_LAB_PROJECT_NAME:-Qualora Demo Lab Real LLM}"
TIMEOUT_SECONDS="${QUALORA_REAL_LLM_TIMEOUT_SECONDS:-900}"
EXTRA_HEADERS_JSON="${QUALORA_REAL_LLM_EXTRA_HEADERS_JSON:-{}}"

COOKIE_JAR="$(mktemp)"
PAYLOAD_FILE="$(mktemp)"
chmod 600 "${COOKIE_JAR}" "${PAYLOAD_FILE}"
cleanup() {
  rm -f "${COOKIE_JAR}" "${PAYLOAD_FILE}"
}
trap cleanup EXIT

csrf_token() {
  awk '$6 == "qualora_csrf" { value=$7 } END { print value }' "${COOKIE_JAR}"
}

json_field() {
  local field="$1"
  python3 -c 'import json,sys; value=json.load(sys.stdin); print(value.get(sys.argv[1], ""))' "${field}"
}

api_request() {
  local method="$1"
  local path="$2"
  local payload_file="${3:-}"
  local headers=(-H "Accept: application/json")
  local csrf
  csrf="$(csrf_token)"
  if [[ -n "${payload_file}" ]]; then
    headers+=(-H "Content-Type: application/json")
  fi
  if [[ "${method}" != "GET" && "${method}" != "HEAD" && "${method}" != "OPTIONS" && -n "${csrf}" ]]; then
    headers+=(-H "X-Qualora-CSRF: ${csrf}")
  fi
  if [[ -n "${payload_file}" ]]; then
    curl -fsS -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" "${headers[@]}" -X "${method}" --data-binary "@${payload_file}" "${API_URL}${path}"
  else
    curl -fsS -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" "${headers[@]}" -X "${method}" "${API_URL}${path}"
  fi
}

write_payload() {
  PAYLOAD_ENV="$1" python3 - <<'PY' > "${PAYLOAD_FILE}"
import json
import os

print(json.dumps(json.loads(os.environ["PAYLOAD_ENV"])))
PY
}

wait_for_status() {
  local path="$1"
  local label="$2"
  local deadline=$((SECONDS + TIMEOUT_SECONDS))
  local response status
  while (( SECONDS < deadline )); do
    response="$(api_request GET "${path}")"
    status="$(printf '%s' "${response}" | json_field status)"
    printf '%s status: %s\n' "${label}" "${status}"
    case "${status}" in
      completed|passed)
        return 0
        ;;
      failed|error|canceled)
        echo "${label} did not complete successfully." >&2
        return 1
        ;;
    esac
    sleep 2
  done
  echo "${label} timed out after ${TIMEOUT_SECONDS} seconds." >&2
  return 1
}

cd "${REPO_DIR}"

echo "Starting Qualora and Demo Lab for optional real LLM mode..."
docker compose --profile demo-lab up -d --build

for _ in $(seq 1 90); do
  if curl -fsS "${API_URL}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
curl -fsS "${API_URL}/healthz" >/dev/null

setup_status="$(curl -fsS -H "Accept: application/json" "${API_URL}/api/v1/setup/status")"
setup_required="$(printf '%s' "${setup_status}" | python3 -c 'import json,sys; print("true" if json.load(sys.stdin).get("setup_required") else "false")')"
if [[ "${setup_required}" == "true" ]]; then
  QUALORA_ADMIN_EMAIL_VALUE="${ADMIN_EMAIL}" QUALORA_ADMIN_PASSWORD_VALUE="${ADMIN_PASSWORD}" python3 - <<'PY' > "${PAYLOAD_FILE}"
import json
import os

print(json.dumps({
    "display_name": "Qualora Admin",
    "email": os.environ["QUALORA_ADMIN_EMAIL_VALUE"],
    "password": os.environ["QUALORA_ADMIN_PASSWORD_VALUE"],
    "confirm_password": os.environ["QUALORA_ADMIN_PASSWORD_VALUE"],
}))
PY
  curl -fsS -c "${COOKIE_JAR}" -H "Accept: application/json" -H "Content-Type: application/json" \
    -X POST --data-binary "@${PAYLOAD_FILE}" "${API_URL}/api/v1/setup/admin" >/dev/null
  echo "Created the local Qualora admin."
fi

QUALORA_ADMIN_EMAIL_VALUE="${ADMIN_EMAIL}" QUALORA_ADMIN_PASSWORD_VALUE="${ADMIN_PASSWORD}" python3 - <<'PY' > "${PAYLOAD_FILE}"
import json
import os

print(json.dumps({
    "email": os.environ["QUALORA_ADMIN_EMAIL_VALUE"],
    "password": os.environ["QUALORA_ADMIN_PASSWORD_VALUE"],
}))
PY
curl -fsS -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" -H "Accept: application/json" -H "Content-Type: application/json" \
  -X POST --data-binary "@${PAYLOAD_FILE}" "${API_URL}/api/v1/auth/login" >/dev/null
echo "Authenticated to the local Qualora API."

providers_json="$(api_request GET /api/v1/ai/providers)"
provider_id="$(QUALORA_PROVIDERS_JSON="${providers_json}" QUALORA_PROVIDER_NAME="${QUALORA_REAL_LLM_NAME}" python3 - <<'PY'
import json
import os

for provider in json.loads(os.environ["QUALORA_PROVIDERS_JSON"]).get("providers", []):
    if provider.get("name") == os.environ["QUALORA_PROVIDER_NAME"]:
        print(provider.get("id", ""))
        break
PY
)"

QUALORA_EXTRA_HEADERS_VALUE="${EXTRA_HEADERS_JSON}" \
QUALORA_PROVIDER_NAME="${QUALORA_REAL_LLM_NAME}" \
QUALORA_PROVIDER_BASE_URL="${QUALORA_REAL_LLM_BASE_URL}" \
QUALORA_PROVIDER_API_KEY="${QUALORA_REAL_LLM_API_KEY}" \
QUALORA_PROVIDER_MODEL="${QUALORA_REAL_LLM_MODEL}" python3 - <<'PY' > "${PAYLOAD_FILE}"
import json
import os

headers = json.loads(os.environ["QUALORA_EXTRA_HEADERS_VALUE"])
if not isinstance(headers, dict):
    raise SystemExit("QUALORA_REAL_LLM_EXTRA_HEADERS_JSON must be a JSON object")
print(json.dumps({
    "name": os.environ["QUALORA_PROVIDER_NAME"],
    "preset": "custom",
    "type": "openai-compatible",
    "base_url": os.environ["QUALORA_PROVIDER_BASE_URL"],
    "model": os.environ["QUALORA_PROVIDER_MODEL"],
    "api_key": os.environ["QUALORA_PROVIDER_API_KEY"],
    "extra_headers": {str(key): str(value) for key, value in headers.items()},
    "temperature": 0.2,
    "max_output_tokens": 1600,
    "timeout_seconds": 60,
    "send_screenshots": False,
    "send_html": False,
    "send_network_bodies": False,
    "redaction_enabled": True,
    "is_default": True,
}))
PY

if [[ -n "${provider_id}" ]]; then
  provider_json="$(api_request PUT "/api/v1/ai/providers/${provider_id}" "${PAYLOAD_FILE}")"
  echo "Reused and updated the configured real LLM provider."
else
  provider_json="$(api_request POST /api/v1/ai/providers "${PAYLOAD_FILE}")"
  provider_id="$(printf '%s' "${provider_json}" | json_field id)"
  echo "Created the configured real LLM provider."
fi
api_request POST "/api/v1/ai/providers/${provider_id}/test" >/dev/null
echo "Real LLM provider connectivity test passed."

projects_json="$(api_request GET /api/v1/projects)"
project_id="$(QUALORA_PROJECTS_JSON="${projects_json}" QUALORA_PROJECT_NAME="${PROJECT_NAME}" python3 - <<'PY'
import json
import os

for project in json.loads(os.environ["QUALORA_PROJECTS_JSON"]).get("projects", []):
    if project.get("name") == os.environ["QUALORA_PROJECT_NAME"]:
        print(project.get("id", ""))
        break
PY
)"
if [[ -z "${project_id}" ]]; then
  write_payload "$(QUALORA_PROJECT_NAME_VALUE="${PROJECT_NAME}" python3 - <<'PY'
import json
import os
print(json.dumps({
    "name": os.environ["QUALORA_PROJECT_NAME_VALUE"],
    "frontend_url": "http://demo-lab-web:8080",
    "api_base_url": "http://demo-lab-api:8080",
    "openapi_url": "http://demo-lab-api:8080/openapi.yaml",
    "allowed_hosts": ["demo-lab-web", "demo-lab-api"],
    "security_mode": "passive",
    "destructive_actions": False,
    "allow_private_targets": True,
}))
PY
)"
  project_json="$(api_request POST /api/v1/projects "${PAYLOAD_FILE}")"
  project_id="$(printf '%s' "${project_json}" | json_field id)"
  echo "Created the Demo Lab real LLM project."
else
  echo "Reused the Demo Lab real LLM project."
fi

write_payload '{"start_url":"http://demo-lab-web:8080","max_pages":12,"max_depth":2,"same_origin_only":true}'
discovery_json="$(api_request POST "/api/v1/projects/${project_id}/discovery-runs" "${PAYLOAD_FILE}")"
discovery_id="$(printf '%s' "${discovery_json}" | json_field id)"
wait_for_status "/api/v1/discovery-runs/${discovery_id}" "Discovery"

QUALORA_PROVIDER_ID="${provider_id}" python3 - <<'PY' > "${PAYLOAD_FILE}"
import json
import os

print(json.dumps({
    "provider_id": os.environ["QUALORA_PROVIDER_ID"],
    "goal": "Explore the main public Demo Lab pages safely, collect visual evidence, and stop.",
    "start_url": "http://demo-lab-web:8080",
    "max_steps": 8,
    "max_depth": 2,
    "same_origin_only": True,
}))
PY
ai_browser_json="$(api_request POST "/api/v1/projects/${project_id}/ai-browser-control-runs" "${PAYLOAD_FILE}")"
ai_browser_id="$(printf '%s' "${ai_browser_json}" | json_field id)"
wait_for_status "/api/v1/ai-browser-control-runs/${ai_browser_id}" "AI Browser Control"

QUALORA_PROVIDER_ID="${provider_id}" QUALORA_DISCOVERY_ID="${discovery_id}" python3 - <<'PY' > "${PAYLOAD_FILE}"
import json
import os

print(json.dumps({
    "provider_id": os.environ["QUALORA_PROVIDER_ID"],
    "discovery_run_id": os.environ["QUALORA_DISCOVERY_ID"],
    "include_discovery_map": True,
    "execution_mode": "safe_executable",
    "max_pages_from_discovery": 12,
    "product_context": "Demo Lab real LLM showcase using sanitized discovery metadata.",
    "focus_areas": ["smoke", "functional", "regression"],
    "max_scenarios": 8,
}))
PY
test_plan_json="$(api_request POST "/api/v1/projects/${project_id}/ai-test-plans" "${PAYLOAD_FILE}")"
test_plan_id="$(printf '%s' "${test_plan_json}" | json_field id)"
echo "AI test plan completed."

QUALORA_PROVIDER_ID="${provider_id}" QUALORA_DISCOVERY_ID="${discovery_id}" python3 - <<'PY' > "${PAYLOAD_FILE}"
import json
import os

print(json.dumps({
    "mode": "safe",
    "provider_id": os.environ["QUALORA_PROVIDER_ID"],
    "use_existing_discovery_run_id": os.environ["QUALORA_DISCOVERY_ID"],
    "execute": False,
    "max_pages": 12,
    "max_depth": 2,
    "max_scenarios": 8,
    "include_quality_checks": True,
    "quality_max_pages": 10,
    "quality_include_security": True,
    "quality_include_accessibility": True,
    "quality_include_performance": True,
    "focus_areas": ["smoke", "functional", "regression"],
    "product_context": "Demo Lab real LLM Safe QA preview using sanitized metadata.",
}))
PY
qa_json="$(api_request POST "/api/v1/projects/${project_id}/qa-runs" "${PAYLOAD_FILE}")"
qa_id="$(printf '%s' "${qa_json}" | json_field id)"
wait_for_status "/api/v1/qa-runs/${qa_id}" "Safe QA"

echo
echo "Real LLM Demo Lab workflow completed."
echo "Qualora UI: ${WEB_URL}"
echo "Project Cockpit: ${WEB_URL}/#/projects/${project_id}"
echo "AI Browser Control Run Viewer: ${WEB_URL}/#/run-viewer/ai-browser-control/${ai_browser_id}"
echo "AI Browser Control HTML report: ${API_URL}/api/v1/ai-browser-control-runs/${ai_browser_id}/report.html"
echo "AI test plan: ${WEB_URL}/#/test-plans/${test_plan_id}"
echo "Safe QA report: ${WEB_URL}/#/qa-runs/${qa_id}"

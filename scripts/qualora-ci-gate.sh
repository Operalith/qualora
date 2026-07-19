#!/usr/bin/env bash
set -euo pipefail

API_URL="${QUALORA_API_URL:-http://localhost:8080}"
PROJECT_ID="${QUALORA_PROJECT_ID:-${1:-}}"
REPORT_ID="${QUALORA_REPORT_ID:-${2:-}}"
REPORT_TYPE="${QUALORA_REPORT_TYPE:-safe_qa}"
BASELINE_ID="${QUALORA_BASELINE_ID:-}"
SESSION_COOKIE="${QUALORA_SESSION_COOKIE:-}"
CSRF_TOKEN="${QUALORA_CSRF_TOKEN:-}"
EMAIL="${QUALORA_EMAIL:-}"
PASSWORD="${QUALORA_PASSWORD:-}"
COOKIE_JAR=""

cleanup() {
  if [[ -n "${COOKIE_JAR}" ]]; then
    rm -f "${COOKIE_JAR}"
  fi
}
trap cleanup EXIT

if [[ -z "${PROJECT_ID}" || -z "${REPORT_ID}" ]]; then
  echo "usage: QUALORA_SESSION_COOKIE=... QUALORA_CSRF_TOKEN=... $0 <project_id> <report_id>" >&2
  echo "env: QUALORA_API_URL, QUALORA_REPORT_TYPE, QUALORA_BASELINE_ID" >&2
  exit 2
fi

if [[ -z "${SESSION_COOKIE}" && -n "${EMAIL}" && -n "${PASSWORD}" ]]; then
  COOKIE_JAR="$(mktemp)"
  login_payload="$(QUALORA_EMAIL_VALUE="${EMAIL}" QUALORA_PASSWORD_VALUE="${PASSWORD}" python3 - <<'PY'
import json
import os

print(json.dumps({"email": os.environ["QUALORA_EMAIL_VALUE"], "password": os.environ["QUALORA_PASSWORD_VALUE"]}))
PY
)"
  curl -fsS -c "${COOKIE_JAR}" -H "Accept: application/json" -H "Content-Type: application/json" -X POST --data "${login_payload}" "${API_URL%/}/api/v1/auth/login" >/dev/null
  CSRF_TOKEN="$(awk '$6 == "qualora_csrf" { value=$7 } END { print value }' "${COOKIE_JAR}")"
fi

payload="$(python3 - <<PY
import json
import os

body = {
    "report_type": "${REPORT_TYPE}",
    "current_report_id": "${REPORT_ID}",
    "use_default_baseline": not bool(os.environ.get("QUALORA_BASELINE_ID", "")),
    "format": "ci",
}
if os.environ.get("QUALORA_BASELINE_ID", ""):
    body["baseline_id"] = os.environ["QUALORA_BASELINE_ID"]
print(json.dumps(body))
PY
)"

headers=(-H "Accept: application/json" -H "Content-Type: application/json")
if [[ -n "${CSRF_TOKEN}" ]]; then
  headers+=(-H "X-Qualora-CSRF: ${CSRF_TOKEN}")
fi

cookie_args=()
if [[ -n "${COOKIE_JAR}" ]]; then
  cookie_args=(-b "${COOKIE_JAR}" -c "${COOKIE_JAR}")
elif [[ -n "${SESSION_COOKIE}" || -n "${CSRF_TOKEN}" ]]; then
  cookie_args=(-b "qualora_session=${SESSION_COOKIE}; qualora_csrf=${CSRF_TOKEN}")
fi

response="$(curl -fsS \
  "${headers[@]}" \
  "${cookie_args[@]}" \
  -X POST \
  --data "${payload}" \
  "${API_URL%/}/api/v1/projects/${PROJECT_ID}/quality-gates/evaluate?format=ci")"

echo "${response}"
QUALORA_CI_GATE_RESPONSE="${response}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["QUALORA_CI_GATE_RESPONSE"])
sys.exit(int(data.get("exit_code", 1)))
PY

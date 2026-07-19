#!/usr/bin/env bash
set -euo pipefail

QUALORA_URL="${QUALORA_URL:-${QUALORA_API_URL:-http://localhost:8080}}"
QUALORA_URL="${QUALORA_URL%/}"
EMAIL="${QUALORA_EMAIL:-}"
PASSWORD="${QUALORA_PASSWORD:-}"
PROJECT_ID="${QUALORA_PROJECT_ID:-${1:-}}"
PROJECT_NAME="${QUALORA_PROJECT_NAME:-}"
FRONTEND_URL="${QUALORA_FRONTEND_URL:-}"
BASELINE_ID="${QUALORA_BASELINE_ID:-}"
USE_LATEST_BASELINE="${QUALORA_USE_LATEST_BASELINE:-true}"
RUN_SAFE_QA="${QUALORA_RUN_SAFE_QA:-true}"
TIMEOUT_SECONDS="${QUALORA_TIMEOUT_SECONDS:-900}"
EXPORT_ISSUES="${QUALORA_EXPORT_ISSUES:-false}"
ISSUE_EXPORT_DRY_RUN="${QUALORA_ISSUE_EXPORT_DRY_RUN:-true}"
ISSUE_EXPORT_CONFIG_ID="${QUALORA_ISSUE_EXPORT_CONFIG_ID:-}"

if [[ -z "${EMAIL}" || -z "${PASSWORD}" ]]; then
  echo "QUALORA_EMAIL and QUALORA_PASSWORD are required" >&2
  exit 2
fi

COOKIE_JAR="$(mktemp)"
cleanup() {
  rm -f "${COOKIE_JAR}"
}
trap cleanup EXIT

csrf_token() {
  awk '$6 == "qualora_csrf" { value=$7 } END { print value }' "${COOKIE_JAR}"
}

api_request() {
  local method="$1"
  local path="$2"
  local payload="${3:-}"
  local csrf
  csrf="$(csrf_token)"
  local headers=(-H "Accept: application/json")
  if [[ -n "${payload}" ]]; then
    headers+=(-H "Content-Type: application/json")
  fi
  if [[ "${method}" != "GET" && "${method}" != "HEAD" && "${method}" != "OPTIONS" && -n "${csrf}" ]]; then
    headers+=(-H "X-Qualora-CSRF: ${csrf}")
  fi
  if [[ -n "${payload}" ]]; then
    curl -fsS -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" "${headers[@]}" -X "${method}" --data "${payload}" "${QUALORA_URL}${path}"
  else
    curl -fsS -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" "${headers[@]}" -X "${method}" "${QUALORA_URL}${path}"
  fi
}

login_payload="$(QUALORA_EMAIL_VALUE="${EMAIL}" QUALORA_PASSWORD_VALUE="${PASSWORD}" python3 - <<'PY'
import json
import os

print(json.dumps({"email": os.environ["QUALORA_EMAIL_VALUE"], "password": os.environ["QUALORA_PASSWORD_VALUE"]}))
PY
)"
curl -fsS -c "${COOKIE_JAR}" -H "Accept: application/json" -H "Content-Type: application/json" -X POST --data "${login_payload}" "${QUALORA_URL}/api/v1/auth/login" >/dev/null

if [[ -z "${PROJECT_ID}" ]]; then
  if [[ -z "${PROJECT_NAME}" ]]; then
    echo "QUALORA_PROJECT_ID or QUALORA_PROJECT_NAME is required" >&2
    exit 2
  fi
  projects_json="$(api_request GET /api/v1/projects)"
  PROJECT_ID="$(QUALORA_PROJECTS_JSON="${projects_json}" QUALORA_PROJECT_NAME_VALUE="${PROJECT_NAME}" python3 - <<'PY'
import json
import os

projects = json.loads(os.environ["QUALORA_PROJECTS_JSON"]).get("projects", [])
for project in projects:
    if project.get("name") == os.environ["QUALORA_PROJECT_NAME_VALUE"]:
        print(project.get("id", ""))
        break
PY
)"
fi

if [[ -z "${PROJECT_ID}" && -n "${PROJECT_NAME}" && -n "${FRONTEND_URL}" ]]; then
  create_payload="$(QUALORA_PROJECT_NAME_VALUE="${PROJECT_NAME}" QUALORA_FRONTEND_URL_VALUE="${FRONTEND_URL}" python3 - <<'PY'
import json
import os
from urllib.parse import urlparse

frontend = os.environ["QUALORA_FRONTEND_URL_VALUE"]
host = urlparse(frontend).hostname or ""
print(json.dumps({
    "name": os.environ["QUALORA_PROJECT_NAME_VALUE"],
    "frontend_url": frontend,
    "api_base_url": "",
    "openapi_url": "",
    "allowed_hosts": [host] if host else [],
    "security_mode": "passive",
    "destructive_actions": False,
    "allow_private_targets": True,
}))
PY
)"
  created_project="$(api_request POST /api/v1/projects "${create_payload}")"
  PROJECT_ID="$(QUALORA_PROJECT_JSON="${created_project}" python3 - <<'PY'
import json
import os

print(json.loads(os.environ["QUALORA_PROJECT_JSON"]).get("id", ""))
PY
)"
fi

if [[ -z "${PROJECT_ID}" ]]; then
  echo "project could not be resolved; set QUALORA_PROJECT_ID or provide QUALORA_PROJECT_NAME and QUALORA_FRONTEND_URL" >&2
  exit 2
fi

ci_payload="$(QUALORA_BASELINE_ID_VALUE="${BASELINE_ID}" QUALORA_USE_LATEST_BASELINE_VALUE="${USE_LATEST_BASELINE}" QUALORA_RUN_SAFE_QA_VALUE="${RUN_SAFE_QA}" QUALORA_TIMEOUT_SECONDS_VALUE="${TIMEOUT_SECONDS}" QUALORA_EXPORT_ISSUES_VALUE="${EXPORT_ISSUES}" QUALORA_ISSUE_EXPORT_DRY_RUN_VALUE="${ISSUE_EXPORT_DRY_RUN}" QUALORA_ISSUE_EXPORT_CONFIG_ID_VALUE="${ISSUE_EXPORT_CONFIG_ID}" python3 - <<'PY'
import json
import os

def env_bool(name, default):
    raw = os.environ.get(name, "")
    if raw == "":
        return default
    return raw.lower() in ("1", "true", "yes", "on")

body = {
    "mode": "safe_qa",
    "use_latest_baseline": env_bool("QUALORA_USE_LATEST_BASELINE_VALUE", True),
    "run_safe_qa": env_bool("QUALORA_RUN_SAFE_QA_VALUE", True),
    "include_quality_checks": True,
    "execute_safe_plan": True,
    "timeout_seconds": int(os.environ.get("QUALORA_TIMEOUT_SECONDS_VALUE") or "900"),
    "export_issues": env_bool("QUALORA_EXPORT_ISSUES_VALUE", False),
    "issue_export_dry_run": env_bool("QUALORA_ISSUE_EXPORT_DRY_RUN_VALUE", True),
}
if os.environ.get("QUALORA_BASELINE_ID_VALUE"):
    body["baseline_id"] = os.environ["QUALORA_BASELINE_ID_VALUE"]
if os.environ.get("QUALORA_ISSUE_EXPORT_CONFIG_ID_VALUE"):
    body["issue_export_config_id"] = os.environ["QUALORA_ISSUE_EXPORT_CONFIG_ID_VALUE"]
print(json.dumps(body))
PY
)"

response="$(api_request POST "/api/v1/projects/${PROJECT_ID}/ci-runs" "${ci_payload}")"
QUALORA_CI_RUN_RESPONSE="${response}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["QUALORA_CI_RUN_RESPONSE"])
gate = data.get("quality_gate_result") or {}
summary = data.get("comparison_summary") or {}
compact = {
    "status": data.get("status"),
    "exit_code": data.get("exit_code", 1),
    "ci_run_id": data.get("ci_run_id"),
    "qa_run_id": data.get("qa_run_id"),
    "report_url": data.get("report_url"),
    "html_report_url": data.get("html_report_url"),
    "new_critical": summary.get("new_critical", 0),
    "new_high": summary.get("new_high", 0),
    "failed_rules": gate.get("failed_rules", []),
    "error_message": data.get("error_message", ""),
}
print(json.dumps(compact, indent=2, sort_keys=True))
sys.exit(int(data.get("exit_code", 1)))
PY

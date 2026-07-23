#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
SCRIPT="${SCRIPT_DIR}/run-demo-lab-real-llm.sh"
MARKER="qualora-real-key-must-not-print"

set +e
output=$(
  QUALORA_REAL_LLM_NAME="Validation Provider" \
  QUALORA_REAL_LLM_BASE_URL="https://example.invalid/v1" \
  QUALORA_REAL_LLM_API_KEY="${MARKER}" \
  "${SCRIPT}" 2>&1
)
code=$?
set -e

if [ "${code}" -ne 2 ]; then
  echo "expected missing-variable validation to exit 2, got ${code}" >&2
  exit 1
fi
if ! printf '%s' "${output}" | grep -q "QUALORA_REAL_LLM_MODEL"; then
  echo "missing-variable validation did not identify QUALORA_REAL_LLM_MODEL" >&2
  exit 1
fi
if printf '%s' "${output}" | grep -q "${MARKER}"; then
  echo "real LLM validation output exposed the API key" >&2
  exit 1
fi

echo "real LLM script validation passed"

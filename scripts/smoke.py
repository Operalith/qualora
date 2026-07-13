#!/usr/bin/env python3
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request


API_URL = os.environ.get("QUALORA_API_URL", "http://localhost:8080").rstrip("/")
TARGET_URL = os.environ.get("QUALORA_TARGET_URL", "https://example.com")
DEFAULT_ALLOWED_HOST = urllib.parse.urlparse(TARGET_URL).hostname or "example.com"
ALLOWED_HOST = os.environ.get("QUALORA_ALLOWED_HOST", DEFAULT_ALLOWED_HOST)
TIMEOUT_SECONDS = int(os.environ.get("QUALORA_SMOKE_TIMEOUT_SECONDS", "120"))


def request(method, path, payload=None):
    body = None
    headers = {"Accept": "application/json"}
    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(
        f"{API_URL}{path}",
        data=body,
        headers=headers,
        method=method,
    )
    try:
        with urllib.request.urlopen(req, timeout=20) as response:
            text = response.read().decode("utf-8")
            return json.loads(text) if text else {}
    except urllib.error.HTTPError as exc:
        text = exc.read().decode("utf-8")
        raise RuntimeError(f"{method} {path} failed with HTTP {exc.code}: {text}") from exc


def main():
    project = request(
        "POST",
        "/api/v1/projects",
        {
            "name": "Qualora Smoke Target",
            "frontend_url": TARGET_URL,
            "api_base_url": "",
            "openapi_url": "",
            "allowed_hosts": [ALLOWED_HOST],
            "security_mode": "passive",
            "destructive_actions": False,
        },
    )
    print(f"created project: {project['id']}")

    run = request("POST", f"/api/v1/projects/{project['id']}/runs")
    run_id = run["id"]
    print(f"started run: {run_id}")

    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/runs/{run_id}")
        status = current["status"]
        print(f"run status: {status}")
        if status in ("completed", "failed", "canceled"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/runs/{run_id}/report")
    print(json.dumps(report, indent=2))

    if report["status"] != "completed":
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request


API_URL = os.environ.get("QUALORA_API_URL", "http://localhost:8080").rstrip("/")
WEB_URL = os.environ.get("QUALORA_WEB_URL", "http://localhost:3000").rstrip("/")
BROWSER_TARGET_URL = os.environ.get("QUALORA_TARGET_URL", "https://example.com")
BROWSER_ALLOWED_HOST = os.environ.get(
    "QUALORA_ALLOWED_HOST",
    urllib.parse.urlparse(BROWSER_TARGET_URL).hostname or "example.com",
)
API_SMOKE_URL = os.environ.get("QUALORA_API_SMOKE_URL", "http://mock-api:8080")
API_SMOKE_OPENAPI_URL = os.environ.get(
    "QUALORA_API_SMOKE_OPENAPI_URL",
    "http://mock-api:8080/openapi.json",
)
API_SMOKE_ALLOWED_HOST = os.environ.get("QUALORA_API_SMOKE_ALLOWED_HOST", "mock-api")
MOCK_API_HEALTH_URL = os.environ.get("MOCK_API_HEALTH_URL", "http://localhost:18081/health")
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


def fetch_text(path):
    req = urllib.request.Request(f"{API_URL}{path}", headers={"Accept": "text/html"}, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=20) as response:
            return response.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        text = exc.read().decode("utf-8")
        raise RuntimeError(f"GET {path} failed with HTTP {exc.code}: {text}") from exc


def wait_for_url(url, timeout_seconds=30):
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=5) as response:
                if response.status < 500:
                    return
        except Exception:
            time.sleep(1)
    raise RuntimeError(f"{url} did not become ready within {timeout_seconds} seconds")


def create_project(payload):
    project = request("POST", "/api/v1/projects", payload)
    print(f"created project: {project['id']} ({project['name']})")
    return project


def run_project(project):
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
        raise RuntimeError(f"run {run_id} finished with status {report['status']}")
    print(f"JSON report: {API_URL}/api/v1/runs/{run_id}/report")
    print(f"HTML report: {API_URL}/api/v1/runs/{run_id}/report.html")
    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    if "Qualora HTML report" not in html:
        raise RuntimeError(f"run {run_id} HTML report did not include the expected title")
    return report


def main():
    print(f"Web UI: {WEB_URL}")

    print("== Browser smoke ==")
    browser_project = create_project(
        {
            "name": "Qualora Browser Smoke Target",
            "frontend_url": BROWSER_TARGET_URL,
            "api_base_url": "",
            "openapi_url": "",
            "allowed_hosts": [BROWSER_ALLOWED_HOST],
            "security_mode": "passive",
            "destructive_actions": False,
        }
    )
    run_project(browser_project)

    print("== API smoke ==")
    wait_for_url(MOCK_API_HEALTH_URL)
    api_project = create_project(
        {
            "name": "Qualora API Smoke Target",
            "frontend_url": "",
            "api_base_url": API_SMOKE_URL,
            "openapi_url": API_SMOKE_OPENAPI_URL,
            "allowed_hosts": [API_SMOKE_ALLOWED_HOST],
            "security_mode": "passive",
            "destructive_actions": False,
            "allow_private_targets": True,
        }
    )
    run_project(api_project)

    return 0


if __name__ == "__main__":
    sys.exit(main())

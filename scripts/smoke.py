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
BROWSER_TARGET_URL = os.environ.get("QUALORA_TARGET_URL", "http://demo-web:8080")
BROWSER_ALLOWED_HOST = os.environ.get(
    "QUALORA_ALLOWED_HOST",
    urllib.parse.urlparse(BROWSER_TARGET_URL).hostname or "demo-web",
)
DEMO_WEB_HEALTH_URL = os.environ.get("DEMO_WEB_HEALTH_URL", "http://localhost:18082/health")
API_SMOKE_URL = os.environ.get("QUALORA_API_SMOKE_URL", "http://mock-api:8080")
API_SMOKE_OPENAPI_URL = os.environ.get(
    "QUALORA_API_SMOKE_OPENAPI_URL",
    "http://mock-api:8080/openapi.json",
)
API_SMOKE_ALLOWED_HOST = os.environ.get("QUALORA_API_SMOKE_ALLOWED_HOST", "mock-api")
MOCK_API_HEALTH_URL = os.environ.get("MOCK_API_HEALTH_URL", "http://localhost:18081/health")
FAKE_LLM_BASE_URL = os.environ.get("QUALORA_FAKE_LLM_URL", "http://fake-llm:8080/v1")
FAKE_LLM_HEALTH_URL = os.environ.get("FAKE_LLM_HEALTH_URL", "http://localhost:18083/health")
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


def fetch_binary(path):
    req = urllib.request.Request(f"{API_URL}{path}", headers={"Accept": "*/*"}, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=20) as response:
            return response.headers, response.read()
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


def create_ai_provider():
    provider = request(
        "POST",
        "/api/v1/ai/providers",
        {
            "name": "Qualora Fake LLM",
            "preset": "custom",
            "type": "openai-compatible",
            "base_url": FAKE_LLM_BASE_URL,
            "model": "qualora-fake-analyst",
            "api_key": "fake-key",
            "extra_headers": {},
            "temperature": 0.2,
            "max_output_tokens": 1200,
            "timeout_seconds": 10,
            "send_screenshots": False,
            "send_html": False,
            "send_network_bodies": False,
            "redaction_enabled": True,
            "is_default": True,
        },
    )
    print(f"created AI provider: {provider['id']} ({provider['name']})")
    if provider.get("api_key_encrypted") or provider.get("api_key"):
        raise RuntimeError("AI provider response exposed an API key")
    if not provider.get("api_key_configured"):
        raise RuntimeError("AI provider did not report configured API key")
    return provider


def test_ai_provider(provider):
    result = request("POST", f"/api/v1/ai/providers/{provider['id']}/test")
    print(f"AI provider test: {json.dumps(result, indent=2)}")
    if not result.get("success"):
        raise RuntimeError(f"AI provider test failed: {result}")


def run_ai_analysis(report, provider):
    run_id = report["run_id"]
    analysis = request("POST", f"/api/v1/runs/{run_id}/ai-analysis", {"provider_id": provider["id"]})
    print(f"AI analysis: {json.dumps(analysis, indent=2)}")
    if analysis.get("status") != "completed":
        raise RuntimeError(f"AI analysis did not complete: {analysis}")
    if analysis.get("risk_level") != "medium":
        raise RuntimeError(f"AI analysis risk level was unexpected: {analysis}")

    updated_report = request("GET", f"/api/v1/runs/{run_id}/report")
    ai_analysis = updated_report.get("ai_analysis")
    if not ai_analysis or ai_analysis.get("status") != "completed":
        raise RuntimeError("JSON report did not include completed AI analysis")
    print(f"AI JSON report: {API_URL}/api/v1/runs/{run_id}/report")
    print(f"AI HTML report: {API_URL}/api/v1/runs/{run_id}/report.html")

    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    if "AI Analysis" not in html or "fake provider" not in html:
        raise RuntimeError("HTML report did not include the fake AI analysis")
    return updated_report


def generate_ai_test_plan(project, report, provider):
    run_id = report["run_id"]
    plan = request(
        "POST",
        f"/api/v1/projects/{project['id']}/ai-test-plans",
        {
            "provider_id": provider["id"],
            "run_id": run_id,
            "product_context": "Smoke demo context. password=should-not-leak",
            "focus_areas": ["smoke", "functional", "api", "regression"],
            "max_scenarios": 8,
        },
    )
    print(f"AI test plan: {json.dumps(plan, indent=2)}")
    if plan.get("status") != "completed":
        raise RuntimeError(f"AI test plan did not complete: {plan}")
    if int(plan.get("total_scenarios") or 0) < 1:
        raise RuntimeError(f"AI test plan did not include scenarios: {plan}")
    if "should-not-leak" in json.dumps(plan):
        raise RuntimeError("AI test plan response exposed redaction smoke text")

    plans = request("GET", f"/api/v1/projects/{project['id']}/test-plans").get("test_plans", [])
    if not any(item.get("id") == plan["id"] for item in plans):
        raise RuntimeError("project test plan list did not include generated plan")

    fetched = request("GET", f"/api/v1/test-plans/{plan['id']}")
    if fetched.get("id") != plan["id"] or fetched.get("status") != "completed":
        raise RuntimeError(f"test plan detail did not match generated plan: {fetched}")

    exported = request("GET", f"/api/v1/test-plans/{plan['id']}/export.json")
    if not exported.get("scenarios"):
        raise RuntimeError(f"test plan export did not include scenarios: {exported}")

    updated_report = request("GET", f"/api/v1/runs/{run_id}/report")
    if not any(item.get("id") == plan["id"] for item in updated_report.get("test_plans", [])):
        raise RuntimeError("JSON report did not include related AI test plan")

    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    if "Related AI Test Plans" not in html or "Qualora deterministic alpha test plan" not in html:
        raise RuntimeError("HTML report did not include the related AI test plan")

    print(f"AI test plan detail: {API_URL}/api/v1/test-plans/{plan['id']}")
    print(f"AI test plan export: {API_URL}/api/v1/test-plans/{plan['id']}/export.json")
    print(f"Web test plan detail: {WEB_URL}/#/test-plans/{plan['id']}")
    return plan


def preview_test_plan_execution(plan):
    preview = request(
        "POST",
        f"/api/v1/test-plans/{plan['id']}/executions",
        {
            "max_scenarios": 5,
            "max_steps_per_scenario": 10,
            "dry_run": True,
        },
    )
    print(f"safe execution preview: {json.dumps(preview, indent=2)}")
    if not preview.get("dry_run"):
        raise RuntimeError("safe execution preview did not preserve dry_run=true")
    if int(preview.get("executable_steps") or 0) < 1:
        raise RuntimeError(f"safe execution preview had no executable steps: {preview}")
    if int(preview.get("skipped_steps") or 0) != 0:
        raise RuntimeError(f"fake safe execution plan unexpectedly skipped steps: {preview}")
    return preview


def execute_test_plan(plan):
    detail = request(
        "POST",
        f"/api/v1/test-plans/{plan['id']}/executions",
        {
            "max_scenarios": 5,
            "max_steps_per_scenario": 10,
            "dry_run": False,
        },
    )
    execution_id = detail["execution"]["id"]
    print(f"started safe test plan execution: {execution_id}")

    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/test-plan-executions/{execution_id}")
        status = current["execution"]["status"]
        print(f"safe execution status: {status}")
        if status in ("completed", "failed", "canceled", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"safe execution {execution_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/test-plan-executions/{execution_id}/report")
    print(f"safe execution report: {json.dumps(report, indent=2)}")
    if report["execution"]["status"] != "completed":
        raise RuntimeError(f"safe execution finished with status {report['execution']['status']}")
    if int(report["execution"].get("passed_steps") or 0) < 1:
        raise RuntimeError("safe execution did not pass any steps")
    if not report.get("scenarios"):
        raise RuntimeError("safe execution report did not include scenarios")

    evidence = report.get("evidence", [])
    types = {item.get("type") for item in evidence}
    if "screenshot" not in types or "browser_observations" not in types:
        raise RuntimeError(f"safe execution report missed expected evidence types: {types}")
    screenshot = next(item for item in evidence if item.get("type") == "screenshot")
    headers, body = fetch_binary(f"/api/v1/evidence/{screenshot['id']}")
    content_type = headers.get("content-type", "")
    if "image/png" not in content_type or not body.startswith(b"\x89PNG"):
        raise RuntimeError("safe execution screenshot evidence was not downloadable PNG data")

    html = fetch_text(f"/api/v1/test-plan-executions/{execution_id}/report.html")
    if "Qualora safe test plan execution report" not in html or "Homepage public smoke checks" not in html:
        raise RuntimeError("safe execution HTML report did not include expected content")

    listed = request("GET", f"/api/v1/test-plans/{plan['id']}/executions").get("executions", [])
    if not any(item.get("id") == execution_id for item in listed):
        raise RuntimeError("test plan execution list did not include completed execution")

    print(f"safe execution JSON report: {API_URL}/api/v1/test-plan-executions/{execution_id}/report")
    print(f"safe execution HTML report: {API_URL}/api/v1/test-plan-executions/{execution_id}/report.html")
    print(f"Web safe execution detail: {WEB_URL}/#/test-plan-executions/{execution_id}")
    return report


def run_project(project, run_path=None):
    path = run_path or f"/api/v1/projects/{project['id']}/runs"
    run = request("POST", path)
    run_id = run["id"]
    print(f"started run: {run_id}")

    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/runs/{run_id}")
        status = current["status"]
        print(f"run status: {status}")
        if status in ("completed", "passed", "failed", "canceled", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/runs/{run_id}/report")
    print(json.dumps(report, indent=2))
    if report["status"] not in ("completed", "passed"):
        raise RuntimeError(f"run {run_id} finished with status {report['status']}")
    print(f"JSON report: {API_URL}/api/v1/runs/{run_id}/report")
    print(f"HTML report: {API_URL}/api/v1/runs/{run_id}/report.html")
    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    if "Qualora HTML report" not in html:
        raise RuntimeError(f"run {run_id} HTML report did not include the expected title")
    return report


def assert_browser_report(report):
    evidence = report.get("evidence", [])
    types = {item.get("type") for item in evidence}
    if "browser_observations" not in types:
        raise RuntimeError("browser report did not include browser_observations evidence")
    screenshots = [item for item in evidence if item.get("type") == "screenshot"]
    if not screenshots:
        raise RuntimeError("browser report did not include screenshot evidence")

    screenshot = screenshots[0]
    metadata = screenshot.get("metadata", {})
    if metadata.get("content_type") != "image/png":
        raise RuntimeError(f"screenshot content type metadata was unexpected: {metadata}")
    if int(metadata.get("size_bytes") or 0) <= 0:
        raise RuntimeError(f"screenshot size metadata was unexpected: {metadata}")

    headers, body = fetch_binary(f"/api/v1/evidence/{screenshot['id']}")
    content_type = headers.get("content-type", "")
    if "image/png" not in content_type:
        raise RuntimeError(f"downloaded screenshot content type was unexpected: {content_type}")
    if not body.startswith(b"\x89PNG"):
        raise RuntimeError("downloaded screenshot did not look like a PNG")


def main():
    print(f"Web UI: {WEB_URL}")

    print("== AI provider smoke ==")
    wait_for_url(FAKE_LLM_HEALTH_URL)
    provider = create_ai_provider()
    test_ai_provider(provider)

    print("== Browser smoke ==")
    wait_for_url(DEMO_WEB_HEALTH_URL)
    browser_project = create_project(
        {
            "name": "Qualora Browser Smoke Target",
            "frontend_url": BROWSER_TARGET_URL,
            "api_base_url": "",
            "openapi_url": "",
            "allowed_hosts": [BROWSER_ALLOWED_HOST],
            "security_mode": "passive",
            "destructive_actions": False,
            "allow_private_targets": True,
        }
    )
    browser_report = run_project(browser_project, f"/api/v1/projects/{browser_project['id']}/browser-smoke-runs")
    assert_browser_report(browser_report)
    browser_report = run_ai_analysis(browser_report, provider)
    browser_plan = generate_ai_test_plan(browser_project, browser_report, provider)
    preview_test_plan_execution(browser_plan)
    execute_test_plan(browser_plan)

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
    api_report = run_project(api_project)
    api_report = run_ai_analysis(api_report, provider)
    generate_ai_test_plan(api_project, api_report, provider)

    return 0


if __name__ == "__main__":
    sys.exit(main())

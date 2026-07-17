#!/usr/bin/env python3
import json
import os
import sys
import time
import http.cookiejar
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
DEMO_USERNAME = os.environ.get("QUALORA_DEMO_USERNAME", "demo@example.com")
DEMO_PASSWORD = os.environ.get("QUALORA_DEMO_PASSWORD", "demo-password")
ROLE_CREDENTIALS = [
    ("Qualora Demo Admin", "admin@example.com", "admin-password", "admin", "Demo Admin"),
    ("Qualora Demo Readonly", "readonly@example.com", "readonly-password", "readonly", "Demo Readonly"),
    ("Qualora Demo Customer A", "customer-a@example.com", "customer-a-password", "customer-a", "Customer A"),
    ("Qualora Demo Customer B", "customer-b@example.com", "customer-b-password", "customer-b", "Customer B"),
]
DEMO_WEB_HEALTH_URL = os.environ.get("DEMO_WEB_HEALTH_URL", "http://localhost:18082/health")
API_SMOKE_URL = os.environ.get("QUALORA_API_SMOKE_URL", "http://demo-api:8080")
API_SMOKE_OPENAPI_URL = os.environ.get(
    "QUALORA_API_SMOKE_OPENAPI_URL",
    "http://demo-api:8080/openapi.yaml",
)
API_SMOKE_ALLOWED_HOST = os.environ.get("QUALORA_API_SMOKE_ALLOWED_HOST", "demo-api")
DEMO_API_HEALTH_URL = os.environ.get("DEMO_API_HEALTH_URL", "http://localhost:18084/health")
FAKE_LLM_BASE_URL = os.environ.get("QUALORA_FAKE_LLM_URL", "http://fake-llm:8080/v1")
FAKE_LLM_HEALTH_URL = os.environ.get("FAKE_LLM_HEALTH_URL", "http://localhost:18083/health")
TIMEOUT_SECONDS = int(os.environ.get("QUALORA_SMOKE_TIMEOUT_SECONDS", "120"))
QUALORA_ADMIN_EMAIL = os.environ.get("QUALORA_ADMIN_EMAIL", "admin@qualora.local")
QUALORA_ADMIN_PASSWORD = os.environ.get("QUALORA_ADMIN_PASSWORD", "qualora-admin-password")
QUALORA_ADMIN_NAME = os.environ.get("QUALORA_ADMIN_NAME", "Qualora Admin")
COOKIE_JAR = http.cookiejar.CookieJar()
OPENER = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(COOKIE_JAR))


def request(method, path, payload=None):
    body = None
    headers = {"Accept": "application/json"}
    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    csrf = csrf_token()
    if method.upper() not in ("GET", "HEAD", "OPTIONS") and csrf:
        headers["X-Qualora-CSRF"] = csrf

    req = urllib.request.Request(
        f"{API_URL}{path}",
        data=body,
        headers=headers,
        method=method,
    )
    try:
        with OPENER.open(req, timeout=20) as response:
            text = response.read().decode("utf-8")
            return json.loads(text) if text else {}
    except urllib.error.HTTPError as exc:
        text = exc.read().decode("utf-8")
        raise RuntimeError(f"{method} {path} failed with HTTP {exc.code}: {text}") from exc


def fetch_text(path):
    req = urllib.request.Request(f"{API_URL}{path}", headers={"Accept": "text/html"}, method="GET")
    try:
        with OPENER.open(req, timeout=20) as response:
            return response.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        text = exc.read().decode("utf-8")
        raise RuntimeError(f"GET {path} failed with HTTP {exc.code}: {text}") from exc


def fetch_web_text(path):
    req = urllib.request.Request(f"{WEB_URL}{path}", headers={"Accept": "text/html,*/*"}, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=20) as response:
            return response.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        text = exc.read().decode("utf-8")
        raise RuntimeError(f"GET {WEB_URL}{path} failed with HTTP {exc.code}: {text}") from exc


def fetch_binary(path):
    req = urllib.request.Request(f"{API_URL}{path}", headers={"Accept": "*/*"}, method="GET")
    try:
        with OPENER.open(req, timeout=20) as response:
            return response.headers, response.read()
    except urllib.error.HTTPError as exc:
        text = exc.read().decode("utf-8")
        raise RuntimeError(f"GET {path} failed with HTTP {exc.code}: {text}") from exc


def public_request(method, path, payload=None):
    body = None
    headers = {"Accept": "application/json"}
    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(f"{API_URL}{path}", data=body, headers=headers, method=method)
    with urllib.request.urlopen(req, timeout=20) as response:
        text = response.read().decode("utf-8")
        return json.loads(text) if text else {}


def expect_http_error(method, path, status):
    try:
        public_request(method, path)
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8")
        if exc.code != status:
            raise RuntimeError(f"{method} {path} returned HTTP {exc.code}, expected {status}: {body}") from exc
        return body
    raise RuntimeError(f"{method} {path} unexpectedly succeeded; expected HTTP {status}")


def csrf_token():
    for cookie in COOKIE_JAR:
        if cookie.name == "qualora_csrf":
            return cookie.value
    return ""


def login_admin():
    logged_in = request("POST", "/api/v1/auth/login", {"email": QUALORA_ADMIN_EMAIL, "password": QUALORA_ADMIN_PASSWORD})
    if logged_in.get("user", {}).get("email") != QUALORA_ADMIN_EMAIL:
        raise RuntimeError(f"login response did not include sanitized admin user: {logged_in}")
    if not csrf_token():
        raise RuntimeError("login did not set a CSRF cookie")
    print(f"logged in as {logged_in['user']['email']}")
    return logged_in


def setup_and_login():
    status = public_request("GET", "/api/v1/setup/status")
    print(f"setup status: {json.dumps(status, indent=2)}")
    if "0.15.0-alpha" not in status.get("version", ""):
        raise RuntimeError(f"unexpected setup status version: {status}")
    expect_http_error("GET", "/api/v1/projects", 401)
    print("protected endpoint rejects unauthenticated requests")

    if status.get("setup_required"):
        created = request(
            "POST",
            "/api/v1/setup/admin",
            {
                "display_name": QUALORA_ADMIN_NAME,
                "email": QUALORA_ADMIN_EMAIL,
                "password": QUALORA_ADMIN_PASSWORD,
                "confirm_password": QUALORA_ADMIN_PASSWORD,
            },
        )
        print(f"created first admin: {created['user']['email']}")
        if created["user"].get("password_hash") or created.get("password"):
            raise RuntimeError("setup response exposed password material")
        second = expect_http_error("POST", "/api/v1/setup/admin", 409)
        if "setup_complete" not in second:
            raise RuntimeError(f"second setup call did not report setup_complete: {second}")
    else:
        login_admin()

    me_response = request("GET", "/api/v1/auth/me")
    if not me_response.get("authenticated") or me_response.get("user", {}).get("email") != QUALORA_ADMIN_EMAIL:
        raise RuntimeError(f"auth/me did not report the admin session: {me_response}")
    protected = request("GET", "/api/v1/projects")
    if "projects" not in protected:
        raise RuntimeError(f"protected endpoint did not work after login: {protected}")

    request("POST", "/api/v1/auth/logout", {})
    logged_out = request("GET", "/api/v1/auth/me")
    if logged_out.get("authenticated"):
        raise RuntimeError(f"auth/me stayed authenticated after logout: {logged_out}")
    expect_http_error("GET", "/api/v1/projects", 401)
    print("logout cleared the session and protected endpoints reject again")

    login_admin()


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


def assert_no_demo_secret(value, label):
    text = value if isinstance(value, str) else json.dumps(value, sort_keys=True)
    secrets = [DEMO_USERNAME, DEMO_PASSWORD]
    for _, username, password, _, _ in ROLE_CREDENTIALS:
        secrets.extend([username, password])
    for secret in secrets:
        if secret and secret in text:
            raise RuntimeError(f"{label} exposed demo credential secret")


def create_credential_profile(project):
    login_url = f"{BROWSER_TARGET_URL.rstrip('/')}/login"
    profile = request(
        "POST",
        f"/api/v1/projects/{project['id']}/credential-profiles",
        {
            "name": "Qualora Demo Login",
            "type": "username_password",
            "username": DEMO_USERNAME,
            "password": DEMO_PASSWORD,
            "login_url": login_url,
            "username_selector": "#username",
            "password_selector": "#password",
            "submit_selector": "#login-submit",
            "success_url_contains": "/dashboard",
            "success_text_contains": "Authenticated area",
            "failure_text_contains": "Invalid credentials",
            "post_login_wait_ms": 100,
            "is_default": True,
        },
    )
    print(f"created credential profile: {profile['id']} ({profile['name']})")
    assert_no_demo_secret(profile, "credential profile response")
    if not profile.get("username_configured") or not profile.get("password_configured"):
        raise RuntimeError(f"credential profile did not report configured secrets: {profile}")
    if profile.get("username_display_hint") == DEMO_USERNAME:
        raise RuntimeError("credential profile returned the raw username as display hint")

    profiles = request("GET", f"/api/v1/projects/{project['id']}/credential-profiles").get("credential_profiles", [])
    assert_no_demo_secret(profiles, "credential profile list")
    if not any(item.get("id") == profile["id"] for item in profiles):
        raise RuntimeError("credential profile list did not include created profile")

    fetched = request("GET", f"/api/v1/credential-profiles/{profile['id']}")
    assert_no_demo_secret(fetched, "credential profile detail")
    if fetched.get("id") != profile["id"]:
        raise RuntimeError("credential profile detail did not match created profile")
    return profile


def create_role_credential_profile(project, name, username, password, role_name, subject_label, is_default=False):
    login_url = f"{BROWSER_TARGET_URL.rstrip('/')}/login"
    profile = request(
        "POST",
        f"/api/v1/projects/{project['id']}/credential-profiles",
        {
            "name": name,
            "type": "username_password",
            "role_name": role_name,
            "role_description": f"Deterministic demo role {role_name}",
            "subject_label": subject_label,
            "username": username,
            "password": password,
            "login_url": login_url,
            "username_selector": "#username",
            "password_selector": "#password",
            "submit_selector": "#login-submit",
            "success_url_contains": "/dashboard",
            "success_text_contains": "Authenticated area",
            "failure_text_contains": "Invalid credentials",
            "post_login_wait_ms": 100,
            "is_default": is_default,
        },
    )
    print(f"created role credential profile: {profile['id']} ({profile['name']} role={profile.get('role_name')})")
    assert_no_demo_secret(profile, "role credential profile response")
    if profile.get("role_name") != role_name:
        raise RuntimeError(f"credential profile did not preserve role_name: {profile}")
    return profile


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
    assert_no_demo_secret(analysis, "AI analysis response")
    if analysis.get("status") != "completed":
        raise RuntimeError(f"AI analysis did not complete: {analysis}")
    if analysis.get("risk_level") != "medium":
        raise RuntimeError(f"AI analysis risk level was unexpected: {analysis}")

    updated_report = request("GET", f"/api/v1/runs/{run_id}/report")
    assert_no_demo_secret(updated_report, "AI JSON report")
    ai_analysis = updated_report.get("ai_analysis")
    if not ai_analysis or ai_analysis.get("status") != "completed":
        raise RuntimeError("JSON report did not include completed AI analysis")
    print(f"AI JSON report: {API_URL}/api/v1/runs/{run_id}/report")
    print(f"AI HTML report: {API_URL}/api/v1/runs/{run_id}/report.html")

    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    assert_no_demo_secret(html, "AI HTML report")
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
    assert_no_demo_secret(plan, "AI test plan response")
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
    assert_no_demo_secret(exported, "AI test plan export")
    if not exported.get("scenarios"):
        raise RuntimeError(f"test plan export did not include scenarios: {exported}")

    updated_report = request("GET", f"/api/v1/runs/{run_id}/report")
    if not any(item.get("id") == plan["id"] for item in updated_report.get("test_plans", [])):
        raise RuntimeError("JSON report did not include related AI test plan")

    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    assert_no_demo_secret(html, "AI test plan HTML report")
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


def import_demo_api_spec(project):
    detail = request(
        "POST",
        f"/api/v1/projects/{project['id']}/api-specs",
        {
            "name": "Qualora Demo API",
            "source_type": "url",
            "source_url": API_SMOKE_OPENAPI_URL,
        },
    )
    print(f"API spec import: {json.dumps(detail, indent=2)}")
    spec = detail["spec"]
    if spec.get("status") != "parsed":
        raise RuntimeError(f"demo API spec did not parse: {detail}")
    if int(spec.get("operation_count") or 0) < 6:
        raise RuntimeError(f"demo API spec discovered too few operations: {detail}")
    if int(spec.get("safe_operation_count") or 0) < 4:
        raise RuntimeError(f"demo API spec discovered too few safe operations: {detail}")
    if int(spec.get("skipped_operation_count") or 0) < 2:
        raise RuntimeError(f"demo API spec did not skip unsafe operations: {detail}")

    listed = request("GET", f"/api/v1/projects/{project['id']}/api-specs").get("api_specs", [])
    if not any(item.get("id") == spec["id"] for item in listed):
        raise RuntimeError("project API spec list did not include imported spec")

    fetched = request("GET", f"/api/v1/api-specs/{spec['id']}")
    if fetched["spec"]["id"] != spec["id"]:
        raise RuntimeError("API spec detail did not match imported spec")

    operations = request("GET", f"/api/v1/api-specs/{spec['id']}/operations").get("operations", [])
    if not operations:
        raise RuntimeError("API operations endpoint returned no operations")
    if not any(item.get("method") == "POST" and not item.get("safe_to_execute") for item in operations):
        raise RuntimeError("POST operation was not skipped")
    if not any(item.get("method") == "DELETE" and not item.get("safe_to_execute") for item in operations):
        raise RuntimeError("DELETE operation was not skipped")
    if not any(item.get("path") == "/profile" and "auth" in (item.get("skip_reason") or "") for item in operations):
        raise RuntimeError("auth-required operation was not skipped")
    if not any(item.get("path") == "/users/{id}" and item.get("safe_to_execute") for item in operations):
        raise RuntimeError("path parameter operation with safe example was not executable")

    print(f"API spec detail: {API_URL}/api/v1/api-specs/{spec['id']}")
    print(f"Web API spec detail: {WEB_URL}/#/api-specs/{spec['id']}")
    return spec


def run_api_smoke(spec):
    run = request("POST", f"/api/v1/api-specs/{spec['id']}/api-smoke-runs")
    run_id = run["id"]
    print(f"started API smoke run: {run_id}")

    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/runs/{run_id}")
        status = current["status"]
        print(f"API smoke status: {status}")
        if status in ("completed", "passed", "failed", "canceled", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"API smoke run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/runs/{run_id}/report")
    print(f"API smoke report: {json.dumps(report, indent=2)}")
    if report.get("run_type") != "api_smoke":
        raise RuntimeError(f"API smoke report had unexpected run type: {report.get('run_type')}")
    if report.get("status") != "completed":
        raise RuntimeError(f"API smoke run did not complete: {report}")
    api_results = report.get("api_results") or []
    if not api_results:
        raise RuntimeError("API smoke report did not include api_results")
    if not any(item.get("status") == "skipped" and item.get("method") in ("POST", "DELETE") for item in api_results):
        raise RuntimeError("API smoke report did not include skipped unsafe operations")
    broken = [item for item in api_results if item.get("path") == "/broken"]
    if not broken or broken[0].get("http_status") != 500 or broken[0].get("status") != "failed":
        raise RuntimeError(f"API smoke did not record deterministic /broken failure: {broken}")
    if not any("5xx" in item.get("title", "") for item in report.get("findings", [])):
        raise RuntimeError("API smoke report did not include deterministic 5xx finding")
    if "deterministic_failure" in json.dumps(report):
        raise RuntimeError("API smoke report exposed response body content")

    api_results_endpoint = request("GET", f"/api/v1/runs/{run_id}/api-results").get("api_results", [])
    if len(api_results_endpoint) != len(api_results):
        raise RuntimeError("API results endpoint did not match report results")

    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    if "API Smoke Results" not in html or "/broken" not in html:
        raise RuntimeError("API smoke HTML report did not include expected API result content")
    if "deterministic_failure" in html:
        raise RuntimeError("API smoke HTML report exposed response body content")

    print(f"API smoke JSON report: {API_URL}/api/v1/runs/{run_id}/report")
    print(f"API smoke HTML report: {API_URL}/api/v1/runs/{run_id}/report.html")
    print(f"Web API smoke report: {WEB_URL}/#/runs/{run_id}")
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


def wait_for_run_report(run_id, label):
    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/runs/{run_id}")
        status = current["status"]
        print(f"{label} status: {status}")
        if status in ("completed", "passed", "failed", "canceled", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"{label} run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/runs/{run_id}/report")
    print(f"{label} report: {json.dumps(report, indent=2)}")
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


def assert_login_report(report, expected_type, expect_authenticated_target, expected_profile_name="Qualora Demo Login"):
    assert_no_demo_secret(report, f"{expected_type} JSON report")
    if report.get("run_type") != expected_type:
        raise RuntimeError(f"login report had unexpected run type: {report.get('run_type')}")
    if report.get("status") != "completed":
        raise RuntimeError(f"login report did not complete: {report}")
    summary = report.get("login_summary") or {}
    if summary.get("login_status") != "passed":
        raise RuntimeError(f"login summary did not report passed login: {summary}")
    if "dashboard" not in (summary.get("login_final_url") or ""):
        raise RuntimeError(f"login final URL did not reach dashboard: {summary}")
    if summary.get("credential_profile_name") != expected_profile_name:
        raise RuntimeError(f"login summary did not include the safe profile name: {summary}")

    evidence = report.get("evidence", [])
    types = {item.get("type") for item in evidence}
    if "login_observations" not in types:
        raise RuntimeError(f"login report missed login_observations evidence: {types}")
    login_evidence = next(item for item in evidence if item.get("type") == "login_observations")
    metadata = login_evidence.get("metadata", {})
    if metadata.get("success") is not True or metadata.get("login_status") != "passed":
        raise RuntimeError(f"login evidence metadata did not report success: {metadata}")
    if expect_authenticated_target and "dashboard" not in (metadata.get("authenticated_target_url") or ""):
        raise RuntimeError(f"authenticated target URL was missing from login evidence: {metadata}")

    if expect_authenticated_target:
        if "browser_observations" not in types:
            raise RuntimeError(f"authenticated smoke report missed browser observations: {types}")
        browser_observations = [item for item in evidence if item.get("type") == "browser_observations"]
        if not any(item.get("metadata", {}).get("authenticated") is True for item in browser_observations):
            raise RuntimeError("authenticated smoke report did not mark browser observations as authenticated")
        assert_browser_report(report)


def test_credential_profile_login(profile, expected_profile_name="Qualora Demo Login"):
    run = request("POST", f"/api/v1/credential-profiles/{profile['id']}/test-login", {})
    run_id = run["id"]
    print(f"started login check run: {run_id}")
    report = wait_for_run_report(run_id, "login check")
    assert_login_report(report, "login_check", False, expected_profile_name)
    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    assert_no_demo_secret(html, "login HTML report")
    if "Login Summary" not in html or expected_profile_name not in html:
        raise RuntimeError("login HTML report did not include login summary")
    print(f"login check JSON report: {API_URL}/api/v1/runs/{run_id}/report")
    print(f"login check HTML report: {API_URL}/api/v1/runs/{run_id}/report.html")
    print(f"Web login check report: {WEB_URL}/#/runs/{run_id}")
    return report


def run_authenticated_browser_smoke(project, profile):
    run = request(
        "POST",
        f"/api/v1/projects/{project['id']}/authenticated-browser-smoke-runs",
        {
            "credential_profile_id": profile["id"],
            "target_path": "/dashboard",
            "capture_screenshot": True,
            "max_duration_seconds": 30,
        },
    )
    run_id = run["id"]
    print(f"started authenticated browser smoke run: {run_id}")
    report = wait_for_run_report(run_id, "authenticated browser smoke")
    assert_login_report(report, "authenticated_browser_smoke", True)
    html = fetch_text(f"/api/v1/runs/{run_id}/report.html")
    assert_no_demo_secret(html, "authenticated browser smoke HTML report")
    if "Login Summary" not in html or "Authenticated Target" not in html:
        raise RuntimeError("authenticated browser smoke HTML report did not include login summary")
    print(f"authenticated browser smoke JSON report: {API_URL}/api/v1/runs/{run_id}/report")
    print(f"authenticated browser smoke HTML report: {API_URL}/api/v1/runs/{run_id}/report.html")
    print(f"Web authenticated browser smoke report: {WEB_URL}/#/runs/{run_id}")
    return report


def run_application_discovery(project, profile=None):
    payload = {
        "start_url": BROWSER_TARGET_URL,
        "max_pages": 12,
        "max_depth": 2,
        "same_origin_only": True,
    }
    if profile:
        payload["credential_profile_id"] = profile["id"]
    run = request("POST", f"/api/v1/projects/{project['id']}/discovery-runs", payload)
    run_id = run["id"]
    print(f"started application discovery run: {run_id}")

    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/discovery-runs/{run_id}")
        status = current["status"]
        print(f"discovery status: {status}")
        if status in ("completed", "failed", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"discovery run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    app_map = request("GET", f"/api/v1/discovery-runs/{run_id}/map")
    report = request("GET", f"/api/v1/discovery-runs/{run_id}/report")
    print(f"discovery report: {json.dumps(report, indent=2)}")
    assert_no_demo_secret(report, "discovery JSON report")
    if report["run"]["status"] != "completed":
        raise RuntimeError(f"discovery run did not complete: {report['run']}")
    if int(report["summary"].get("total_pages") or 0) <= 1:
        raise RuntimeError(f"discovery found too few pages: {report['summary']}")
    if int(report["summary"].get("total_links") or 0) <= 1:
        raise RuntimeError(f"discovery found too few links: {report['summary']}")
    if int(report["summary"].get("total_forms") or 0) < 1:
        raise RuntimeError(f"discovery did not find forms: {report['summary']}")
    if not any(link.get("skip_reason") == "unsafe_link_skipped" for link in report.get("links", [])):
        raise RuntimeError("discovery did not record an unsafe skipped link")
    if not any(link.get("skip_reason") == "external_link_skipped" for link in report.get("links", [])):
        raise RuntimeError("discovery did not record an external skipped link")
    if not any(page.get("screenshot_evidence_id") for page in report.get("pages", [])):
        raise RuntimeError("discovery pages did not include screenshot evidence IDs")
    evidence = report.get("evidence") or []
    if "screenshot" not in {item.get("type") for item in evidence}:
        raise RuntimeError("discovery report did not include screenshot evidence")
    screenshot = next(item for item in evidence if item.get("type") == "screenshot")
    headers, body = fetch_binary(f"/api/v1/evidence/{screenshot['id']}")
    if "image/png" not in headers.get("content-type", "") or not body.startswith(b"\x89PNG"):
        raise RuntimeError("discovery screenshot evidence was not downloadable PNG data")
    if app_map.get("summary", {}).get("total_pages") != report.get("summary", {}).get("total_pages"):
        raise RuntimeError("discovery map summary did not match report summary")

    html = fetch_text(f"/api/v1/discovery-runs/{run_id}/report.html")
    assert_no_demo_secret(html, "discovery HTML report")
    if "Qualora application discovery report" not in html or "Skipped Links" not in html:
        raise RuntimeError("discovery HTML report did not include expected content")

    print(f"discovery JSON report: {API_URL}/api/v1/discovery-runs/{run_id}/report")
    print(f"discovery HTML report: {API_URL}/api/v1/discovery-runs/{run_id}/report.html")
    print(f"discovery map: {API_URL}/api/v1/discovery-runs/{run_id}/map")
    print(f"Web discovery report: {WEB_URL}/#/discovery-runs/{run_id}")
    return report


def run_quality_check(project, discovery_report, profile=None):
    payload = {
        "use_latest_discovery": True,
        "target_url": BROWSER_TARGET_URL,
        "max_pages": 10,
        "include_security": True,
        "include_accessibility": True,
        "include_performance": True,
    }
    if profile:
        payload["credential_profile_id"] = profile["id"]
    run = request("POST", f"/api/v1/projects/{project['id']}/quality-check-runs", payload)
    run_id = run["id"]
    print(f"started quality check run: {run_id}")

    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/quality-check-runs/{run_id}")
        status = current["status"]
        print(f"quality check status: {status}")
        if status in ("completed", "failed", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"quality check run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/quality-check-runs/{run_id}/report")
    print(f"quality check report: {json.dumps(report, indent=2)}")
    assert_no_demo_secret(report, "quality check JSON report")
    if report["run"]["status"] != "completed":
        raise RuntimeError(f"quality check run did not complete: {report['run']}")
    if report["run"].get("discovery_run_id") != discovery_report["run"]["id"]:
        raise RuntimeError(f"quality check did not use latest discovery run: {report['run']}")
    summary = report.get("summary") or {}
    if int(summary.get("total_pages") or 0) < 1:
        raise RuntimeError(f"quality check did not check any pages: {summary}")
    if int(summary.get("security_findings") or 0) < 1:
        raise RuntimeError(f"quality check did not record passive security findings: {summary}")
    if int(summary.get("accessibility_findings") or 0) < 1:
        raise RuntimeError(f"quality check did not record accessibility findings: {summary}")
    if int(summary.get("performance_findings") or 0) < 1:
        raise RuntimeError(f"quality check did not record performance/front-end findings: {summary}")
    categories = {item.get("category") for item in report.get("results", [])}
    if not {"security", "accessibility", "performance"}.issubset(categories):
        raise RuntimeError(f"quality check report missed expected categories: {categories}")

    html = fetch_text(f"/api/v1/quality-check-runs/{run_id}/report.html")
    assert_no_demo_secret(html, "quality check HTML report")
    if "Qualora quality report" not in html or "Quality Findings" not in html:
        raise RuntimeError("quality check HTML report did not include expected content")

    listed = request("GET", f"/api/v1/projects/{project['id']}/quality-check-runs").get("quality_check_runs", [])
    if not any(item.get("id") == run_id for item in listed):
        raise RuntimeError("project quality check list did not include completed run")

    print(f"quality check JSON report: {API_URL}/api/v1/quality-check-runs/{run_id}/report")
    print(f"quality check HTML report: {API_URL}/api/v1/quality-check-runs/{run_id}/report.html")
    print(f"Web quality report: {WEB_URL}/#/quality-check-runs/{run_id}")
    return report


def assert_quality_ui_bundle():
    index = fetch_web_text("/")
    asset_paths = []
    for marker in ("src=\"", "href=\""):
        parts = index.split(marker)
        for part in parts[1:]:
            candidate = part.split("\"", 1)[0]
            if candidate.startswith("/assets/"):
                asset_paths.append(candidate)
    bundle_text = index
    for path in sorted(set(asset_paths)):
        if path.endswith((".js", ".css")):
            bundle_text += "\n" + fetch_web_text(path)
    for expected in (
        "Quality Checks",
        "Start quality checks",
        "Quality Report",
        "Include passive quality checks",
        "Create project with guided setup",
        "Run demo workflow",
        "Project Readiness",
        "Guided Project Setup",
        "Reports",
    ):
        if expected not in bundle_text:
            raise RuntimeError(f"web UI bundle did not include expected v0.15 UI text: {expected}")
    print("web UI bundle includes guided onboarding, readiness, reports, and Quality Checks screens")


def wait_for_discovery_report(run_id, label):
    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/discovery-runs/{run_id}")
        status = current["status"]
        print(f"{label} discovery status: {status}")
        if status in ("completed", "failed", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"{label} discovery run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/discovery-runs/{run_id}/report")
    assert_no_demo_secret(report, f"{label} discovery JSON report")
    if report["run"]["status"] != "completed":
        raise RuntimeError(f"{label} discovery did not complete: {report['run']}")
    if int(report.get("summary", {}).get("total_pages") or 0) < 1:
        raise RuntimeError(f"{label} discovery did not record pages: {report.get('summary')}")
    html = fetch_text(f"/api/v1/discovery-runs/{run_id}/report.html")
    assert_no_demo_secret(html, f"{label} discovery HTML report")
    if "Qualora application discovery report" not in html:
        raise RuntimeError(f"{label} discovery HTML report did not render")
    return report


def wait_for_quality_report(run_id, label):
    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/quality-check-runs/{run_id}")
        status = current["status"]
        print(f"{label} quality status: {status}")
        if status in ("completed", "failed", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"{label} quality run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/quality-check-runs/{run_id}/report")
    assert_no_demo_secret(report, f"{label} quality JSON report")
    if report["run"]["status"] != "completed":
        raise RuntimeError(f"{label} quality run did not complete: {report['run']}")
    if int(report.get("summary", {}).get("total_findings") or 0) < 1:
        raise RuntimeError(f"{label} quality report did not include findings: {report.get('summary')}")
    html = fetch_text(f"/api/v1/quality-check-runs/{run_id}/report.html")
    assert_no_demo_secret(html, f"{label} quality HTML report")
    if "Qualora quality report" not in html:
        raise RuntimeError(f"{label} quality HTML report did not render")
    return report


def run_guided_project_setup():
    print("== Guided project setup smoke ==")
    setup = request(
        "POST",
        "/api/v1/onboarding/project-setup",
        {
            "project": {
                "name": "Qualora Guided Demo Target",
                "frontend_url": BROWSER_TARGET_URL,
                "api_base_url": API_SMOKE_URL,
                "openapi_url": "",
                "allowed_hosts": [BROWSER_ALLOWED_HOST, API_SMOKE_ALLOWED_HOST],
                "security_mode": "passive",
                "destructive_actions": False,
                "allow_private_targets": True,
            },
            "ai": {"mode": "demo"},
            "credential": {
                "mode": "create",
                "profile": {
                    "name": "Qualora Guided Demo Login",
                    "type": "username_password",
                    "username": DEMO_USERNAME,
                    "password": DEMO_PASSWORD,
                    "login_url": f"{BROWSER_TARGET_URL.rstrip('/')}/login",
                    "username_selector": "#username",
                    "password_selector": "#password",
                    "submit_selector": "#login-submit",
                    "success_url_contains": "/dashboard",
                    "success_text_contains": "Authenticated area",
                    "failure_text_contains": "Invalid credentials",
                    "post_login_wait_ms": 100,
                    "is_default": True,
                },
            },
            "api_spec": {"mode": "demo"},
            "workflow": {
                "browser_smoke": True,
                "discovery": True,
                "quality_checks": True,
                "safe_qa_run": True,
                "execute_safe_qa": False,
                "api_smoke": True,
                "authenticated_smoke": True,
            },
        },
    )
    print(f"guided setup response: {json.dumps(setup, indent=2)}")
    assert_no_demo_secret(setup, "guided setup response")
    project = setup.get("project") or {}
    started = setup.get("started") or {}
    if not project.get("id"):
        raise RuntimeError(f"guided setup did not return project id: {setup}")
    required = [
        "browser_smoke_run_id",
        "authenticated_smoke_run_id",
        "discovery_run_id",
        "quality_check_run_id",
        "safe_qa_run_id",
        "api_smoke_run_id",
        "ai_provider_id",
        "credential_profile_id",
        "api_spec_id",
    ]
    missing = [key for key in required if not started.get(key)]
    if missing:
        raise RuntimeError(f"guided setup did not start/create expected resources {missing}: {setup}")

    browser_report = wait_for_run_report(started["browser_smoke_run_id"], "guided browser smoke")
    assert_browser_report(browser_report)
    authenticated_report = wait_for_run_report(started["authenticated_smoke_run_id"], "guided authenticated smoke")
    assert_login_report(authenticated_report, "authenticated_browser_smoke", True, "Qualora Guided Demo Login")
    wait_for_discovery_report(started["discovery_run_id"], "guided setup")
    wait_for_quality_report(started["quality_check_run_id"], "guided setup")
    wait_for_qa_run(started["safe_qa_run_id"], "guided setup", expect_quality=True)

    api_report = request("GET", f"/api/v1/runs/{started['api_smoke_run_id']}/report")
    assert_no_demo_secret(api_report, "guided API smoke report")
    if api_report.get("status") != "completed" or not api_report.get("api_summary"):
        raise RuntimeError(f"guided API smoke report was not complete: {api_report}")
    html = fetch_text(f"/api/v1/runs/{started['api_smoke_run_id']}/report.html")
    if "API Smoke Results" not in html or "/broken" not in html:
        raise RuntimeError("guided API smoke HTML report did not include API results")

    project_detail = request("GET", f"/api/v1/projects/{project['id']}")
    if project_detail.get("name") != "Qualora Guided Demo Target":
        raise RuntimeError(f"guided project detail did not load: {project_detail}")
    print(f"guided project: {WEB_URL}/#/projects/{project['id']}")
    print(f"guided Safe QA report: {WEB_URL}/#/qa-runs/{started['safe_qa_run_id']}")
    print(f"guided reports index: {WEB_URL}/#/reports")
    return setup


def generate_discovery_ai_test_plan(project, discovery_report, provider):
    discovery_run_id = discovery_report["run"]["id"]
    plan = request(
        "POST",
        f"/api/v1/projects/{project['id']}/ai-test-plans",
        {
            "provider_id": provider["id"],
            "discovery_run_id": discovery_run_id,
            "include_discovery_map": True,
            "execution_mode": "safe_executable",
            "max_pages_from_discovery": 12,
            "product_context": "Discovery-aware smoke context. password=should-not-leak",
            "focus_areas": ["smoke", "functional", "regression"],
            "max_scenarios": 10,
        },
    )
    print(f"discovery-aware AI test plan: {json.dumps(plan, indent=2)}")
    assert_no_demo_secret(plan, "discovery-aware AI test plan response")
    rendered = json.dumps(plan, sort_keys=True)
    if "should-not-leak" in rendered:
        raise RuntimeError("discovery-aware AI test plan exposed redaction smoke text")
    if plan.get("status") != "completed":
        raise RuntimeError(f"discovery-aware AI test plan did not complete: {plan}")
    if plan.get("source_type") != "discovery" or plan.get("discovery_run_id") != discovery_run_id:
        raise RuntimeError(f"discovery-aware AI test plan was not linked to discovery: {plan}")
    coverage = plan.get("execution_coverage") or {}
    if int(coverage.get("executable_steps") or 0) < 1:
        raise RuntimeError(f"discovery-aware AI test plan did not record executable coverage: {coverage}")
    scenarios = (plan.get("plan_json") or {}).get("scenarios") or []
    tags = {tag for scenario in scenarios for tag in scenario.get("tags", [])}
    if "generated_from_discovery" not in tags or "safe_executable_candidate" not in tags:
        raise RuntimeError(f"discovery-aware AI test plan did not include discovery/safe tags: {tags}")

    fetched = request("GET", f"/api/v1/test-plans/{plan['id']}")
    if fetched.get("source_type") != "discovery" or fetched.get("discovery_run_id") != discovery_run_id:
        raise RuntimeError(f"discovery-aware test plan detail lost source metadata: {fetched}")
    print(f"Discovery-aware AI test plan detail: {API_URL}/api/v1/test-plans/{plan['id']}")
    print(f"Web discovery-aware test plan: {WEB_URL}/#/test-plans/{plan['id']}")
    return plan


def wait_for_qa_run(qa_run_id, label, require_execution=False, expect_quality=False):
    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        current = request("GET", f"/api/v1/qa-runs/{qa_run_id}")
        status = current["status"]
        print(f"{label} status: {status}")
        if status in ("completed", "failed", "error", "canceled"):
            if not require_execution or current.get("test_plan_execution_id"):
                break
        time.sleep(2)
    else:
        raise RuntimeError(f"{label} QA run {qa_run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/qa-runs/{qa_run_id}/report")
    print(f"{label} report: {json.dumps(report, indent=2)}")
    assert_no_demo_secret(report, f"{label} JSON report")
    if report["run"]["status"] != "completed":
        raise RuntimeError(f"{label} QA run did not complete: {report['run']}")
    if not report.get("discovery_run") or not report.get("test_plan") or not report.get("execution_preview"):
        raise RuntimeError(f"{label} QA report missed discovery, plan, or preview: {report}")
    if int(report["execution_preview"].get("executable_steps") or 0) < 1:
        raise RuntimeError(f"{label} QA preview had no executable steps: {report['execution_preview']}")
    if expect_quality:
        quality_summary = report.get("quality_summary") or {}
        if int(quality_summary.get("total_findings") or 0) < 1:
            raise RuntimeError(f"{label} QA report did not include quality findings: {quality_summary}")
        if not report.get("quality_check_run"):
            raise RuntimeError(f"{label} QA report missed the linked quality check run")
        if not report.get("quality_results"):
            raise RuntimeError(f"{label} QA report missed quality result rows")

    html = fetch_text(f"/api/v1/qa-runs/{qa_run_id}/report.html")
    assert_no_demo_secret(html, f"{label} HTML report")
    if "Qualora safe QA report" not in html or "Safe Execution Preview" not in html:
        raise RuntimeError(f"{label} QA HTML report did not include expected content")
    if expect_quality and "Quality Checks" not in html:
        raise RuntimeError(f"{label} QA HTML report did not include quality checks")

    print(f"{label} QA JSON report: {API_URL}/api/v1/qa-runs/{qa_run_id}/report")
    print(f"{label} QA HTML report: {API_URL}/api/v1/qa-runs/{qa_run_id}/report.html")
    print(f"Web {label} QA report: {WEB_URL}/#/qa-runs/{qa_run_id}")
    return report


def run_safe_qa_preview(project, discovery_report, provider):
    qa_run = request(
        "POST",
        f"/api/v1/projects/{project['id']}/qa-runs",
        {
            "mode": "safe",
            "provider_id": provider["id"],
            "use_existing_discovery_run_id": discovery_report["run"]["id"],
            "execute": False,
            "max_pages": 12,
            "max_depth": 2,
            "max_scenarios": 10,
            "include_quality_checks": True,
            "quality_max_pages": 10,
            "quality_include_security": True,
            "quality_include_accessibility": True,
            "quality_include_performance": True,
            "focus_areas": ["smoke", "functional", "regression"],
            "product_context": "One-click safe QA preview. password=should-not-leak",
        },
    )
    qa_run_id = qa_run["id"]
    print(f"started safe QA preview: {qa_run_id}")
    report = wait_for_qa_run(qa_run_id, "safe QA preview", expect_quality=True)
    if report["run"].get("test_plan_execution_id"):
        raise RuntimeError(f"safe QA preview unexpectedly executed a plan: {report['run']}")
    return report


def execute_previewed_qa_run(preview_report):
    qa_run_id = preview_report["run"]["id"]
    accepted = request("POST", f"/api/v1/qa-runs/{qa_run_id}/execute", {})
    print(f"accepted safe QA preview execution: {json.dumps(accepted, indent=2)}")
    report = wait_for_qa_run(qa_run_id, "safe QA executed preview", require_execution=True, expect_quality=True)
    execution_report = report.get("execution_report")
    if not execution_report:
        raise RuntimeError(f"executed QA report did not include execution report: {report}")
    if execution_report["execution"].get("status") != "completed":
        raise RuntimeError(f"QA execution did not complete: {execution_report['execution']}")
    if int(execution_report["execution"].get("passed_steps") or 0) < 1:
        raise RuntimeError("QA execution did not pass any safe steps")
    evidence_types = {item.get("type") for item in report.get("evidence", [])}
    if "screenshot" not in evidence_types or "browser_observations" not in evidence_types:
        raise RuntimeError(f"executed QA report missed expected evidence types: {evidence_types}")
    return report


def create_authorization_check(project, profile, name, target_path, expected_outcome, success_text="", denied_text="Access denied"):
    check = request(
        "POST",
        f"/api/v1/projects/{project['id']}/authorization-checks",
        {
            "name": name,
            "description": "Deterministic demo authorization check",
            "type": "browser_url",
            "resource_label": target_path,
            "actor_credential_profile_id": profile["id"],
            "expected_outcome": expected_outcome,
            "target_url": target_path,
            "success_text_contains": success_text,
            "denied_text_contains": denied_text,
            "enabled": True,
        },
    )
    print(f"created authorization check: {check['id']} ({check['name']})")
    assert_no_demo_secret(check, "authorization check response")
    if check.get("expected_outcome") != expected_outcome:
        raise RuntimeError(f"authorization check expected outcome was not preserved: {check}")
    return check


def run_authorization_checks(project, checks):
    run = request(
        "POST",
        f"/api/v1/projects/{project['id']}/authorization-check-runs",
        {"check_ids": [check["id"] for check in checks], "max_checks": 10},
    )
    run_id = run["id"]
    print(f"started authorization check run: {run_id}")

    deadline = time.time() + TIMEOUT_SECONDS
    while time.time() < deadline:
        detail = request("GET", f"/api/v1/authorization-check-runs/{run_id}")
        status = detail["run"]["status"]
        print(f"authorization check status: {status}")
        if status in ("completed", "failed", "error"):
            break
        time.sleep(2)
    else:
        raise RuntimeError(f"authorization check run {run_id} did not finish within {TIMEOUT_SECONDS} seconds")

    report = request("GET", f"/api/v1/authorization-check-runs/{run_id}/report")
    print(f"authorization check report: {json.dumps(report, indent=2)}")
    assert_no_demo_secret(report, "authorization JSON report")
    if report["run"]["status"] != "completed":
        raise RuntimeError(f"authorization run did not complete: {report}")
    if int(report["run"].get("passed_checks") or 0) != len(checks):
        raise RuntimeError(f"authorization run did not pass all expected checks: {report['run']}")
    if report["run"].get("failed_checks") or report["run"].get("skipped_checks"):
        raise RuntimeError(f"authorization run had failed/skipped checks: {report['run']}")
    results = report.get("results") or []
    if len(results) != len(checks):
        raise RuntimeError(f"authorization report did not include all results: {results}")
    if not all(item.get("status") == "passed" for item in results):
        raise RuntimeError(f"authorization results were not all passed: {results}")
    evidence = report.get("evidence") or []
    types = {item.get("type") for item in evidence}
    if "screenshot" not in types or "authorization_observations" not in types:
        raise RuntimeError(f"authorization report missed expected evidence types: {types}")
    screenshot = next(item for item in evidence if item.get("type") == "screenshot")
    expect_http_error("GET", f"/api/v1/authorization-check-runs/{run_id}/report", 401)
    expect_http_error("GET", f"/api/v1/evidence/{screenshot['id']}", 401)
    print("authorization report and evidence reject unauthenticated requests")
    headers, body = fetch_binary(f"/api/v1/evidence/{screenshot['id']}")
    if "image/png" not in headers.get("content-type", "") or not body.startswith(b"\x89PNG"):
        raise RuntimeError("authorization screenshot evidence was not downloadable PNG data")

    html = fetch_text(f"/api/v1/authorization-check-runs/{run_id}/report.html")
    assert_no_demo_secret(html, "authorization HTML report")
    if "Qualora role-aware authorization report" not in html or "Check Results" not in html:
        raise RuntimeError("authorization HTML report did not include expected content")

    print(f"authorization JSON report: {API_URL}/api/v1/authorization-check-runs/{run_id}/report")
    print(f"authorization HTML report: {API_URL}/api/v1/authorization-check-runs/{run_id}/report.html")
    print(f"Web authorization report: {WEB_URL}/#/authorization-check-runs/{run_id}")
    return report


def main():
    print(f"Web UI: {WEB_URL}")
    wait_for_url(f"{API_URL}/healthz")
    wait_for_url(f"{WEB_URL}/healthz")
    setup_and_login()
    assert_quality_ui_bundle()

    print("== AI provider smoke ==")
    wait_for_url(FAKE_LLM_HEALTH_URL)
    provider = create_ai_provider()
    test_ai_provider(provider)
    run_guided_project_setup()

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
    credential_profile = create_credential_profile(browser_project)
    login_report = test_credential_profile_login(credential_profile)
    run_ai_analysis(login_report, provider)
    authenticated_report = run_authenticated_browser_smoke(browser_project, credential_profile)
    authenticated_report = run_ai_analysis(authenticated_report, provider)
    generate_ai_test_plan(browser_project, authenticated_report, provider)
    discovery_report = run_application_discovery(browser_project)
    run_quality_check(browser_project, discovery_report, credential_profile)
    discovery_plan = generate_discovery_ai_test_plan(browser_project, discovery_report, provider)
    preview_test_plan_execution(discovery_plan)
    qa_preview_report = run_safe_qa_preview(browser_project, discovery_report, provider)
    execute_previewed_qa_run(qa_preview_report)

    role_profiles = {
        role_name: create_role_credential_profile(browser_project, name, username, password, role_name, subject_label)
        for name, username, password, role_name, subject_label in ROLE_CREDENTIALS
    }
    test_credential_profile_login(role_profiles["admin"], "Qualora Demo Admin")
    test_credential_profile_login(role_profiles["readonly"], "Qualora Demo Readonly")
    authorization_checks = [
        create_authorization_check(
            browser_project,
            role_profiles["admin"],
            "Admin can access admin route",
            "/admin",
            "allowed",
            success_text="Admin console",
        ),
        create_authorization_check(
            browser_project,
            role_profiles["readonly"],
            "Readonly is denied admin route",
            "/admin",
            "denied",
        ),
        create_authorization_check(
            browser_project,
            role_profiles["customer-a"],
            "Customer A can access own invoice",
            "/customers/a/invoice",
            "allowed",
            success_text="Invoice for Customer A",
        ),
        create_authorization_check(
            browser_project,
            role_profiles["customer-b"],
            "Customer B is denied Customer A invoice",
            "/customers/a/invoice",
            "denied",
        ),
    ]
    run_authorization_checks(browser_project, authorization_checks)

    browser_report = run_project(browser_project, f"/api/v1/projects/{browser_project['id']}/browser-smoke-runs")
    assert_browser_report(browser_report)
    browser_report = run_ai_analysis(browser_report, provider)
    browser_plan = generate_ai_test_plan(browser_project, browser_report, provider)
    preview_test_plan_execution(browser_plan)
    execute_test_plan(browser_plan)

    print("== API smoke ==")
    wait_for_url(DEMO_API_HEALTH_URL)
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
    api_spec = import_demo_api_spec(api_project)
    api_report = run_api_smoke(api_spec)
    api_report = run_ai_analysis(api_report, provider)
    generate_ai_test_plan(api_project, api_report, provider)

    return 0


if __name__ == "__main__":
    sys.exit(main())

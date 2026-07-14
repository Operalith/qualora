import {
  CreateBucketCommand,
  HeadBucketCommand,
  PutObjectCommand,
  S3Client
} from "@aws-sdk/client-s3";
import dns from "node:dns/promises";
import { randomUUID } from "node:crypto";
import { promises as fs } from "node:fs";
import net from "node:net";
import path from "node:path";
import Redis from "ioredis";
import { Pool } from "pg";
import { chromium, type Page } from "playwright";
import { buildFindings, type BrowserResult, type FindingInput } from "./findings";

type Config = {
  databaseUrl: string;
  redisUrl: string;
  queueName: string;
  planExecutionQueueName: string;
  evidenceDir: string;
  s3Endpoint: string;
  s3Region: string;
  s3Bucket: string;
  s3AccessKeyId: string;
  s3SecretAccessKey: string;
  s3ForcePathStyle: boolean;
};

type BrowserRunJob = {
  job_id: string;
  run_id: string;
  project_id: string;
};

type TestPlanExecutionJob = {
  execution_id: string;
};

type Project = {
  id: string;
  frontend_url: string;
  allowed_hosts: string[];
  allow_private_targets: boolean;
};

type TestPlanExecutionContext = {
  id: string;
  project: Project;
};

type ExecutionScenario = {
  id: string;
  name: string;
  steps: ExecutionStep[];
};

type ExecutionStep = {
  id: string;
  scenario_execution_id: string;
  step_order: number;
  mapped_action: string;
  target: string;
  expected_result: string;
};

type BrowserSignals = {
  consoleErrors: BrowserResult["consoleErrors"];
  failedRequests: BrowserResult["failedRequests"];
  blockedRequests: BrowserResult["blockedRequests"];
  blockedURLs: Set<string>;
};

type EvidenceOwner = {
  runID?: string;
  executionID?: string;
};

type FindingOwner = EvidenceOwner & {
  scenarioExecutionID?: string;
  stepExecutionID?: string;
};

type StoredEvidenceObject = {
  uri: string;
  filename: string;
  key: string;
  contentType: string;
  sizeBytes: number;
  createdAt: string;
  storage: "s3" | "local";
};

const config = loadConfig();
const pool = new Pool({ connectionString: config.databaseUrl });
const redis = new Redis(config.redisUrl, { maxRetriesPerRequest: null });
const s3 = new S3Client({
  endpoint: config.s3Endpoint,
  region: config.s3Region,
  forcePathStyle: config.s3ForcePathStyle,
  credentials: {
    accessKeyId: config.s3AccessKeyId,
    secretAccessKey: config.s3SecretAccessKey
  }
});

let stopping = false;

process.on("SIGTERM", () => {
  stopping = true;
});
process.on("SIGINT", () => {
  stopping = true;
});

main().catch(async (error) => {
  log("worker_fatal", { error: sanitizeText(String(error)) });
  await shutdown();
  process.exit(1);
});

async function main(): Promise<void> {
  await ensureS3Bucket();
  log("worker_started", { browser_queue: config.queueName, test_plan_execution_queue: config.planExecutionQueueName });

  while (!stopping) {
    const item = await redis.blpop(config.queueName, config.planExecutionQueueName, 5);
    if (!item) {
      continue;
    }

    const queue = item[0];
    const payload = item[1];
    if (queue === config.planExecutionQueueName) {
      let job: TestPlanExecutionJob;
      try {
        job = JSON.parse(payload) as TestPlanExecutionJob;
      } catch {
        log("invalid_test_plan_execution_payload", {});
        continue;
      }
      await handleTestPlanExecutionJob(job);
    } else {
      let job: BrowserRunJob;
      try {
        job = JSON.parse(payload) as BrowserRunJob;
      } catch {
        log("invalid_job_payload", {});
        continue;
      }
      await handleJob(job);
    }
  }

  await shutdown();
}

async function shutdown(): Promise<void> {
  redis.disconnect();
  await pool.end();
}

async function handleJob(job: BrowserRunJob): Promise<void> {
  log("run_started", { run_id: job.run_id, project_id: job.project_id });

  try {
    await markJobRunning(job);

    const project = await getProject(job.project_id);
    const scopeCheck = await validateTargetURL(project.frontend_url, project.allowed_hosts, project.allow_private_targets);
    if (!scopeCheck.ok) {
      throw new Error(scopeCheck.reason);
    }

    const result = await runBrowserCheck(project);
    const evidenceIds: string[] = [];

    if (result.screenshot) {
      const screenshotObject = await storeScreenshot("runs", job.run_id, result.screenshot);
      const screenshotEvidenceID = await insertEvidence({ runID: job.run_id }, "screenshot", screenshotObject.uri, {
        filename: screenshotObject.filename,
        key: screenshotObject.key,
        content_type: screenshotObject.contentType,
        size_bytes: screenshotObject.sizeBytes,
        created_at: screenshotObject.createdAt,
        storage: screenshotObject.storage,
        target_url: result.targetURL,
        final_url: result.finalURL,
        page_title: result.pageTitle,
        status_code: result.statusCode
      });
      evidenceIds.push(screenshotEvidenceID);
    }

    const observationsEvidenceID = await insertEvidence({ runID: job.run_id }, "browser_observations", "inline://browser-observations", {
      target_url: result.targetURL,
      final_url: result.finalURL,
      page_title: result.pageTitle,
      status_code: result.statusCode,
      body_text_length: result.bodyTextLength,
      load_error: result.loadError,
      timed_out: result.timedOut,
      console_errors: result.consoleErrors,
      failed_requests: result.failedRequests,
      blocked_requests: result.blockedRequests
    });
    evidenceIds.push(observationsEvidenceID);

    const findings = buildFindings(result, evidenceIds);
    for (const finding of findings) {
      await insertFinding({ runID: job.run_id }, finding);
    }

    await finishJob(job, "completed", "", result.pageTitle);
    log("run_completed", { run_id: job.run_id, findings: findings.length });
  } catch (error) {
    const message = sanitizeText(error instanceof Error ? error.message : String(error));
    await insertFinding({ runID: job.run_id }, {
      title: "Browser smoke test failed",
      severity: "high",
      category: "frontend",
      confidence: "medium",
      description: "The browser worker could not complete the smoke test.",
      recommendation: "Verify the target URL, allowed hosts, network access from the worker container, and application availability.",
      evidenceIds: []
    }).catch(() => undefined);
    await finishJob(job, "failed", message, "").catch(() => undefined);
    log("run_failed", { run_id: job.run_id, error: message });
  }
}

async function markJobRunning(job: BrowserRunJob): Promise<void> {
  if (!job.job_id) {
    throw new Error("browser job is missing job_id");
  }
  await pool.query(
    `UPDATE run_jobs
     SET status = 'running', started_at = COALESCE(started_at, now()), updated_at = now()
     WHERE id = $1 AND run_id = $2`,
    [job.job_id, job.run_id]
  );
  await pool.query(
    `UPDATE test_runs
     SET status = 'running', started_at = COALESCE(started_at, now()), updated_at = now()
     WHERE id = $1`,
    [job.run_id]
  );
}

async function finishJob(job: BrowserRunJob, status: "completed" | "failed", errorMessage: string, pageTitle: string): Promise<void> {
  await pool.query(
    `UPDATE run_jobs
     SET status = $3, error_message = $4, completed_at = now(), updated_at = now()
     WHERE id = $1 AND run_id = $2`,
    [job.job_id, job.run_id, status, errorMessage]
  );
  if (pageTitle) {
    await pool.query(
      `UPDATE test_runs
       SET page_title = $2, updated_at = now()
       WHERE id = $1`,
      [job.run_id, pageTitle]
    );
  }
  await pool.query(`SELECT refresh_test_run_status($1)`, [job.run_id]);
}

async function handleTestPlanExecutionJob(job: TestPlanExecutionJob): Promise<void> {
  log("test_plan_execution_started", { execution_id: job.execution_id });

  try {
    if (!job.execution_id) {
      throw new Error("test plan execution job is missing execution_id");
    }
    const execution = await getTestPlanExecutionContext(job.execution_id);
    const scopeCheck = await validateTargetURL(
      execution.project.frontend_url,
      execution.project.allowed_hosts,
      execution.project.allow_private_targets
    );
    if (!scopeCheck.ok) {
      throw new Error(scopeCheck.reason);
    }

    await markTestPlanExecutionRunning(execution.id);
    const scenarios = await getQueuedExecutionScenarios(execution.id);
    if (scenarios.length === 0) {
      await refreshTestPlanExecutionStatus(execution.id);
      log("test_plan_execution_completed", { execution_id: execution.id, scenarios: 0 });
      return;
    }

    await runSafeTestPlanExecution(execution, scenarios);
    const status = await refreshTestPlanExecutionStatus(execution.id);
    log("test_plan_execution_completed", { execution_id: execution.id, status });
  } catch (error) {
    const message = sanitizeText(error instanceof Error ? error.message : String(error));
    if (job.execution_id) {
      await failTestPlanExecution(job.execution_id, message).catch(() => undefined);
      await insertFinding(
        { executionID: job.execution_id },
        {
          title: "Safe test plan execution failed",
          severity: "high",
          category: "test_plan_execution",
          confidence: "medium",
          description: "The browser worker could not complete the approved safe test plan execution.",
          recommendation: "Verify the project frontend URL, allowed hosts, worker network access, and mapped test plan steps.",
          evidenceIds: []
        }
      ).catch(() => undefined);
    }
    log("test_plan_execution_failed", { execution_id: job.execution_id, error: message });
  }
}

async function getTestPlanExecutionContext(executionID: string): Promise<TestPlanExecutionContext> {
  const result = await pool.query(
    `SELECT e.id, p.id AS project_id, p.frontend_url, p.allowed_hosts, p.allow_private_targets
     FROM test_plan_executions e
     JOIN projects p ON p.id = e.project_id
     WHERE e.id = $1`,
    [executionID]
  );
  if (result.rowCount !== 1) {
    throw new Error("test plan execution was not found");
  }

  const row = result.rows[0] as {
    id: string;
    project_id: string;
    frontend_url: string;
    allowed_hosts: string[] | string;
    allow_private_targets: boolean;
  };
  return {
    id: row.id,
    project: {
      id: row.project_id,
      frontend_url: row.frontend_url,
      allowed_hosts: Array.isArray(row.allowed_hosts) ? row.allowed_hosts : JSON.parse(row.allowed_hosts),
      allow_private_targets: row.allow_private_targets
    }
  };
}

async function getQueuedExecutionScenarios(executionID: string): Promise<ExecutionScenario[]> {
  const scenarioRows = await pool.query(
    `SELECT id, name
     FROM test_plan_execution_scenarios
     WHERE execution_id = $1 AND status = 'queued'
     ORDER BY created_at ASC`,
    [executionID]
  );
  const scenarios = scenarioRows.rows.map((row) => ({
    id: row.id as string,
    name: row.name as string,
    steps: [] as ExecutionStep[]
  }));
  const scenarioIndex = new Map(scenarios.map((scenario, index) => [scenario.id, index]));

  const stepRows = await pool.query(
    `SELECT id, scenario_execution_id, step_order, mapped_action, target, expected_result
     FROM test_plan_execution_steps
     WHERE execution_id = $1 AND status = 'queued'
     ORDER BY created_at ASC, step_order ASC`,
    [executionID]
  );
  for (const row of stepRows.rows) {
    const step: ExecutionStep = {
      id: row.id,
      scenario_execution_id: row.scenario_execution_id,
      step_order: Number(row.step_order),
      mapped_action: row.mapped_action,
      target: row.target,
      expected_result: row.expected_result
    };
    const index = scenarioIndex.get(step.scenario_execution_id);
    if (index !== undefined) {
      scenarios[index].steps.push(step);
    }
  }
  return scenarios.filter((scenario) => scenario.steps.length > 0);
}

async function markTestPlanExecutionRunning(executionID: string): Promise<void> {
  await pool.query(
    `UPDATE test_plan_executions
     SET status = 'running', started_at = COALESCE(started_at, now()), updated_at = now()
     WHERE id = $1`,
    [executionID]
  );
}

async function failTestPlanExecution(executionID: string, message: string): Promise<void> {
  await pool.query(
    `UPDATE test_plan_execution_steps
     SET status = 'error', error_message = $2, updated_at = now()
     WHERE execution_id = $1 AND status IN ('queued', 'running')`,
    [executionID, message]
  );
  await pool.query(
    `UPDATE test_plan_execution_scenarios
     SET status = 'error', completed_at = COALESCE(completed_at, now()), updated_at = now()
     WHERE execution_id = $1 AND status IN ('queued', 'running')`,
    [executionID]
  );
  await pool.query(
    `UPDATE test_plan_executions
     SET status = 'failed', error_message = $2, completed_at = COALESCE(completed_at, now()), updated_at = now()
     WHERE id = $1`,
    [executionID, message]
  );
  await refreshTestPlanExecutionStatus(executionID).catch(() => undefined);
}

async function refreshTestPlanExecutionStatus(executionID: string): Promise<string> {
  const result = await pool.query(`SELECT refresh_test_plan_execution_status($1) AS status`, [executionID]);
  return String(result.rows[0]?.status ?? "");
}

async function markScenarioRunning(scenarioID: string): Promise<void> {
  await pool.query(
    `UPDATE test_plan_execution_scenarios
     SET status = 'running', started_at = COALESCE(started_at, now()), updated_at = now()
     WHERE id = $1`,
    [scenarioID]
  );
}

async function finishScenario(scenarioID: string, status: "passed" | "failed"): Promise<void> {
  await pool.query(
    `UPDATE test_plan_execution_scenarios
     SET status = $2, completed_at = now(), updated_at = now()
     WHERE id = $1`,
    [scenarioID, status]
  );
}

async function markStepRunning(stepID: string): Promise<void> {
  await pool.query(
    `UPDATE test_plan_execution_steps
     SET status = 'running', updated_at = now()
     WHERE id = $1`,
    [stepID]
  );
}

async function finishStep(
  stepID: string,
  status: "passed" | "failed" | "error",
  actualResult: string,
  errorMessage: string,
  durationMS: number,
  evidenceID: string | null
): Promise<void> {
  await pool.query(
    `UPDATE test_plan_execution_steps
     SET status = $2, actual_result = $3, error_message = $4, duration_ms = $5, evidence_id = $6, updated_at = now()
     WHERE id = $1`,
    [stepID, status, actualResult, errorMessage, durationMS, evidenceID]
  );
}

async function runSafeTestPlanExecution(execution: TestPlanExecutionContext, scenarios: ExecutionScenario[]): Promise<void> {
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox"]
  });
  const context = await browser.newContext({
    ignoreHTTPSErrors: false,
    viewport: { width: 1365, height: 768 }
  });
  const page = await context.newPage();
  const signals = createBrowserSignals(page);
  await installAllowedHostRoutes(page, execution.project, signals);

  try {
    for (const scenario of scenarios) {
      await markScenarioRunning(scenario.id);
      let failed = false;
      for (const step of scenario.steps) {
        const startedAt = Date.now();
        await markStepRunning(step.id);
        try {
          const result = await executeExecutionStep(page, execution.project, execution.id, step, signals);
          await finishStep(step.id, "passed", result.actualResult, "", Date.now() - startedAt, result.evidenceID);
        } catch (error) {
          failed = true;
          const message = sanitizeText(error instanceof Error ? error.message : String(error));
          const evidenceID = await captureStepFailureEvidence(page, execution.id, scenario.id, step, message).catch(() => null);
          await insertFinding(
            { executionID: execution.id, scenarioExecutionID: scenario.id, stepExecutionID: step.id },
            {
              title: "Safe test plan step failed",
              severity: "medium",
              category: "test_plan_execution",
              confidence: "high",
              description: `Step ${step.step_order} (${step.mapped_action}) failed: ${message}`,
              recommendation: "Review the step target, page state, and captured evidence. Keep the test plan limited to supported safe actions.",
              evidenceIds: evidenceID ? [evidenceID] : []
            }
          );
          await finishStep(step.id, "failed", "", message, Date.now() - startedAt, evidenceID);
        }
      }
      await finishScenario(scenario.id, failed ? "failed" : "passed");
      await refreshTestPlanExecutionStatus(execution.id);
    }
  } finally {
    await browser.close();
  }
}

async function executeExecutionStep(
  page: Page,
  project: Project,
  executionID: string,
  step: ExecutionStep,
  signals: BrowserSignals
): Promise<{ actualResult: string; evidenceID: string | null }> {
  switch (step.mapped_action) {
    case "goto": {
      const target = await safeExecutionTarget(project, step.target);
      const response = await page.goto(target, { waitUntil: "domcontentloaded", timeout: 30000 });
      await page.waitForLoadState("networkidle", { timeout: 5000 }).catch(() => undefined);
      const status = response ? response.status() : null;
      if (status !== null && status >= 400) {
        throw new Error(`navigation returned HTTP ${status}`);
      }
      return { actualResult: `navigated to ${sanitizeURL(page.url())}`, evidenceID: null };
    }
    case "wait_for_load_state": {
      await page.waitForLoadState("networkidle", { timeout: 10000 }).catch(async () => {
        await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
      });
      return { actualResult: "page load state reached", evidenceID: null };
    }
    case "assert_title_contains": {
      const title = sanitizeText(await page.title());
      assertIncludes(title, step.target, "page title");
      return { actualResult: `page title contains ${step.target}`, evidenceID: null };
    }
    case "assert_url_contains": {
      const currentURL = sanitizeURL(page.url());
      assertIncludes(currentURL, step.target, "current URL");
      return { actualResult: `current URL contains ${step.target}`, evidenceID: null };
    }
    case "assert_text_visible": {
      await page.getByText(step.target, { exact: false }).first().waitFor({ state: "visible", timeout: 10000 });
      return { actualResult: `text is visible: ${step.target}`, evidenceID: null };
    }
    case "assert_element_visible": {
      await page.locator(step.target).first().waitFor({ state: "visible", timeout: 10000 });
      return { actualResult: `element is visible: ${step.target}`, evidenceID: null };
    }
    case "assert_link_exists": {
      const target = await safeExecutionTarget(project, step.target);
      const exists = await page.locator("a[href]").evaluateAll((links, expected) => {
        return links.some((link) => (link as HTMLAnchorElement).href === expected);
      }, target);
      if (!exists) {
        throw new Error(`link not found for ${target}`);
      }
      return { actualResult: `link exists: ${sanitizeURL(target)}`, evidenceID: null };
    }
    case "check_link_status": {
      const target = await safeExecutionTarget(project, step.target);
      let method = "HEAD";
      let response = await page.request.head(target, { timeout: 10000 }).catch(() => null);
      if (!response || response.status() === 405) {
        method = "GET";
        response = await page.request.get(target, { timeout: 10000 });
      }
      if (response.status() >= 400) {
        throw new Error(`link returned HTTP ${response.status()}`);
      }
      const evidenceID = await insertEvidence({ executionID }, "api_request", "inline://test-plan-link-check", {
        method,
        url: sanitizeURL(target),
        status_code: response.status(),
        content_type: response.headers()["content-type"] ?? "",
        safe_methods_only: true
      });
      return { actualResult: `link returned HTTP ${response.status()}`, evidenceID };
    }
    case "capture_screenshot": {
      const screenshot = await captureScreenshot(page);
      if (!screenshot) {
        throw new Error("screenshot could not be captured");
      }
      const object = await storeScreenshot("test-plan-executions", executionID, screenshot);
      const evidenceID = await insertEvidence({ executionID }, "screenshot", object.uri, {
        filename: object.filename,
        key: object.key,
        content_type: object.contentType,
        size_bytes: object.sizeBytes,
        created_at: object.createdAt,
        storage: object.storage,
        current_url: sanitizeURL(page.url()),
        page_title: sanitizeText(await page.title().catch(() => ""))
      });
      return { actualResult: "screenshot captured", evidenceID };
    }
    case "collect_browser_signals": {
      const evidenceID = await insertEvidence({ executionID }, "browser_observations", "inline://test-plan-browser-observations", {
        current_url: sanitizeURL(page.url()),
        page_title: sanitizeText(await page.title().catch(() => "")),
        console_errors: signals.consoleErrors.slice(0, 50),
        failed_requests: signals.failedRequests.slice(0, 50),
        blocked_requests: signals.blockedRequests.slice(0, 50)
      });
      return { actualResult: "browser signals collected", evidenceID };
    }
    case "assert_no_console_errors": {
      if (signals.consoleErrors.length > 0) {
        throw new Error(`${signals.consoleErrors.length} console error(s) were observed`);
      }
      return { actualResult: "no console errors observed", evidenceID: null };
    }
    case "assert_no_failed_requests": {
      if (signals.failedRequests.length > 0) {
        throw new Error(`${signals.failedRequests.length} failed request(s) were observed`);
      }
      return { actualResult: "no failed requests observed", evidenceID: null };
    }
    default:
      throw new Error(`unsupported execution action ${step.mapped_action}`);
  }
}

async function captureStepFailureEvidence(
  page: Page,
  executionID: string,
  scenarioID: string,
  step: ExecutionStep,
  message: string
): Promise<string | null> {
  const screenshot = await captureScreenshot(page);
  if (!screenshot) {
    return null;
  }
  const object = await storeScreenshot("test-plan-executions", executionID, screenshot);
  return insertEvidence({ executionID }, "screenshot", object.uri, {
    filename: object.filename,
    key: object.key,
    content_type: object.contentType,
    size_bytes: object.sizeBytes,
    created_at: object.createdAt,
    storage: object.storage,
    scenario_execution_id: scenarioID,
    step_execution_id: step.id,
    step_order: step.step_order,
    action: step.mapped_action,
    error_message: message,
    current_url: sanitizeURL(page.url()),
    page_title: sanitizeText(await page.title().catch(() => ""))
  });
}

function createBrowserSignals(page: Page): BrowserSignals {
  const signals: BrowserSignals = {
    consoleErrors: [],
    failedRequests: [],
    blockedRequests: [],
    blockedURLs: new Set<string>()
  };

  page.on("console", (message) => {
    if (message.type() !== "error") {
      return;
    }
    const location = message.location();
    signals.consoleErrors.push({
      type: message.type(),
      text: sanitizeText(message.text()),
      location: sanitizeText(`${location.url}:${location.lineNumber}:${location.columnNumber}`)
    });
  });

  page.on("requestfailed", (request) => {
    const url = sanitizeURL(request.url());
    if (signals.blockedURLs.has(url)) {
      return;
    }
    signals.failedRequests.push({
      url,
      method: request.method(),
      failure: sanitizeText(request.failure()?.errorText ?? "request failed")
    });
  });

  return signals;
}

async function installAllowedHostRoutes(page: Page, project: Project, signals: BrowserSignals): Promise<void> {
  await page.route("**/*", async (route) => {
    const requestURL = route.request().url();
    if (!requestURL.startsWith("http://") && !requestURL.startsWith("https://")) {
      await route.continue();
      return;
    }

    const allowed = await validateTargetURL(requestURL, project.allowed_hosts, project.allow_private_targets);
    if (!allowed.ok) {
      const sanitized = sanitizeURL(requestURL);
      signals.blockedURLs.add(sanitized);
      signals.blockedRequests.push({ url: sanitized, reason: allowed.reason });
      await route.abort("blockedbyclient");
      return;
    }

    await route.continue();
  });
}

async function safeExecutionTarget(project: Project, raw: string): Promise<string> {
  let root: URL;
  let target: URL;
  try {
    root = new URL(project.frontend_url);
    target = new URL(raw, root);
  } catch {
    throw new Error("execution target URL is invalid");
  }
  target.hash = "";
  if (target.protocol !== "http:" && target.protocol !== "https:") {
    throw new Error("execution target must use http or https");
  }
  if (target.origin !== root.origin) {
    throw new Error("execution target must stay on the project frontend origin");
  }
  if (hasSensitiveTargetQuery(target)) {
    throw new Error("execution target query contains sensitive parameter names");
  }
  const allowed = await validateTargetURL(target.toString(), project.allowed_hosts, project.allow_private_targets);
  if (!allowed.ok) {
    throw new Error(allowed.reason);
  }
  return target.toString();
}

function hasSensitiveTargetQuery(target: URL): boolean {
  const sensitiveNames = [
    "access_token",
    "api_key",
    "apikey",
    "auth",
    "authorization",
    "credential",
    "jwt",
    "key",
    "password",
    "passwd",
    "secret",
    "session",
    "token"
  ];
  for (const name of target.searchParams.keys()) {
    const normalized = name.toLowerCase();
    if (sensitiveNames.includes(normalized) || sensitiveNames.some((sensitive) => normalized.includes(sensitive))) {
      return true;
    }
  }
  return false;
}

function assertIncludes(actual: string, expected: string, label: string): void {
  if (!actual.toLowerCase().includes(expected.toLowerCase())) {
    throw new Error(`${label} did not contain ${expected}`);
  }
}

async function getProject(projectID: string): Promise<Project> {
  const result = await pool.query(
    `SELECT id, frontend_url, allowed_hosts, allow_private_targets
     FROM projects
     WHERE id = $1`,
    [projectID]
  );
  if (result.rowCount !== 1) {
    throw new Error("project was not found");
  }

  const row = result.rows[0] as {
    id: string;
    frontend_url: string;
    allowed_hosts: string[] | string;
    allow_private_targets: boolean;
  };

  return {
    id: row.id,
    frontend_url: row.frontend_url,
    allowed_hosts: Array.isArray(row.allowed_hosts) ? row.allowed_hosts : JSON.parse(row.allowed_hosts),
    allow_private_targets: row.allow_private_targets
  };
}

async function runBrowserCheck(project: Project): Promise<BrowserResult> {
  const targetURL = sanitizeURL(project.frontend_url);
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox"]
  });
  const context = await browser.newContext({
    ignoreHTTPSErrors: false,
    viewport: { width: 1365, height: 768 }
  });
  const page = await context.newPage();

  const consoleErrors: BrowserResult["consoleErrors"] = [];
  const failedRequests: BrowserResult["failedRequests"] = [];
  const blockedRequests: BrowserResult["blockedRequests"] = [];
  const blockedURLs = new Set<string>();

  page.on("console", (message) => {
    if (message.type() !== "error") {
      return;
    }
    const location = message.location();
    consoleErrors.push({
      type: message.type(),
      text: sanitizeText(message.text()),
      location: sanitizeText(`${location.url}:${location.lineNumber}:${location.columnNumber}`)
    });
  });

  page.on("requestfailed", (request) => {
    const url = sanitizeURL(request.url());
    if (blockedURLs.has(url)) {
      return;
    }
    failedRequests.push({
      url,
      method: request.method(),
      failure: sanitizeText(request.failure()?.errorText ?? "request failed")
    });
  });

  await page.route("**/*", async (route) => {
    const requestURL = route.request().url();
    if (!requestURL.startsWith("http://") && !requestURL.startsWith("https://")) {
      await route.continue();
      return;
    }

    const allowed = await validateTargetURL(requestURL, project.allowed_hosts, project.allow_private_targets);
    if (!allowed.ok) {
      const sanitized = sanitizeURL(requestURL);
      blockedURLs.add(sanitized);
      blockedRequests.push({ url: sanitized, reason: allowed.reason });
      await route.abort("blockedbyclient");
      return;
    }

    await route.continue();
  });

  let pageTitle = "";
  let statusCode: number | null = null;
  let finalURL = targetURL;
  let bodyTextLength: number | null = null;
  let loadError = "";
  let screenshot: Buffer | null = null;

  try {
    const response = await page.goto(project.frontend_url, {
      waitUntil: "domcontentloaded",
      timeout: 30000
    });
    statusCode = response ? response.status() : null;
    await page.waitForLoadState("networkidle", { timeout: 5000 }).catch(() => undefined);
    pageTitle = sanitizeText(await page.title().catch(() => ""));
    finalURL = sanitizeURL(page.url());
    bodyTextLength = await page
      .evaluate(() => (document.body?.innerText ?? "").trim().length)
      .catch(() => null);
  } catch (error) {
    loadError = sanitizeText(error instanceof Error ? error.message : String(error));
    finalURL = sanitizeURL(page.url());
    bodyTextLength = await page
      .evaluate(() => (document.body?.innerText ?? "").trim().length)
      .catch(() => null);
  }

  screenshot = await captureScreenshot(page);
  await browser.close();

  return {
    targetURL,
    finalURL,
    pageTitle,
    statusCode,
    bodyTextLength,
    loadError,
    timedOut: /\btimeout\b|timed out/i.test(loadError),
    consoleErrors: consoleErrors.slice(0, 50),
    failedRequests: failedRequests.slice(0, 50),
    blockedRequests: blockedRequests.slice(0, 50),
    screenshot
  };
}

async function captureScreenshot(page: Page): Promise<Buffer | null> {
  try {
    return await page.screenshot({ fullPage: true, type: "png" });
  } catch {
    return null;
  }
}

async function insertEvidence(owner: EvidenceOwner, type: string, uri: string, metadata: Record<string, unknown>): Promise<string> {
  const id = randomUUID();
  await pool.query(
    `INSERT INTO evidence (id, run_id, test_plan_execution_id, type, uri, metadata)
     VALUES ($1, $2, $3, $4, $5, $6)`,
    [id, owner.runID ?? null, owner.executionID ?? null, type, uri, JSON.stringify(metadata)]
  );
  return id;
}

async function insertFinding(owner: FindingOwner, finding: FindingInput): Promise<void> {
  await pool.query(
    `INSERT INTO findings (
       id, run_id, test_plan_execution_id, scenario_execution_id, step_execution_id,
       title, severity, category, confidence, description, recommendation, evidence_ids
     )
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
    [
      randomUUID(),
      owner.runID ?? null,
      owner.executionID ?? null,
      owner.scenarioExecutionID ?? null,
      owner.stepExecutionID ?? null,
      finding.title,
      finding.severity,
      finding.category,
      finding.confidence,
      finding.description,
      finding.recommendation,
      JSON.stringify(finding.evidenceIds)
    ]
  );
}

async function ensureS3Bucket(): Promise<void> {
  try {
    await s3.send(new HeadBucketCommand({ Bucket: config.s3Bucket }));
  } catch {
    try {
      await s3.send(new CreateBucketCommand({ Bucket: config.s3Bucket }));
    } catch (error) {
      log("s3_bucket_create_failed", { error: sanitizeText(error instanceof Error ? error.message : String(error)) });
    }
  }
}

async function storeScreenshot(ownerKind: "runs" | "test-plan-executions", ownerID: string, screenshot: Buffer): Promise<StoredEvidenceObject> {
  const filename = `${Date.now()}-${randomUUID()}.png`;
  const key = `${ownerKind}/${ownerID}/screenshots/${filename}`;
  const createdAt = new Date().toISOString();
  const contentType = "image/png";

  try {
    await putScreenshotObject(key, screenshot);
    return {
      uri: `s3://${config.s3Bucket}/${key}`,
      filename,
      key,
      contentType,
      sizeBytes: screenshot.byteLength,
      createdAt,
      storage: "s3"
    };
  } catch (error) {
    try {
      await ensureS3Bucket();
      await putScreenshotObject(key, screenshot);
      return {
        uri: `s3://${config.s3Bucket}/${key}`,
        filename,
        key,
        contentType,
        sizeBytes: screenshot.byteLength,
        createdAt,
        storage: "s3"
      };
    } catch {
      log("s3_put_failed_using_local_fallback", { error: sanitizeText(error instanceof Error ? error.message : String(error)) });
    }

    const localPath = path.join(config.evidenceDir, key);
    await fs.mkdir(path.dirname(localPath), { recursive: true });
    await fs.writeFile(localPath, screenshot);
    return {
      uri: `file://${localPath}`,
      filename,
      key,
      contentType,
      sizeBytes: screenshot.byteLength,
      createdAt,
      storage: "local"
    };
  }
}

async function putScreenshotObject(key: string, screenshot: Buffer): Promise<void> {
  await s3.send(
    new PutObjectCommand({
      Bucket: config.s3Bucket,
      Key: key,
      Body: screenshot,
      ContentType: "image/png"
    })
  );
}

async function validateTargetURL(
  raw: string,
  allowedHosts: string[],
  allowPrivateTargets: boolean
): Promise<{ ok: true } | { ok: false; reason: string }> {
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    return { ok: false, reason: "URL is invalid" };
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return { ok: false, reason: "only http and https URLs are supported" };
  }

  const host = parsed.hostname.toLowerCase().replace(/\.$/, "");
  if (!host) {
    return { ok: false, reason: "host is required" };
  }
  if (isBlockedHost(host, allowPrivateTargets)) {
    return { ok: false, reason: `host ${host} is blocked by the default safety policy` };
  }
  if (!allowedHosts.some((allowedHost) => hostAllowed(host, allowedHost))) {
    return { ok: false, reason: `host ${host} is not present in allowed_hosts` };
  }
  const resolved = await validateResolvedTarget(host, allowPrivateTargets);
  if (!resolved.ok) {
    return resolved;
  }
  return { ok: true };
}

function hostAllowed(host: string, allowedHost: string): boolean {
  const normalized = allowedHost.toLowerCase().trim().replace(/\.$/, "");
  if (host === normalized) {
    return true;
  }
  if (normalized.startsWith("*.") && host.endsWith(normalized.slice(1))) {
    return true;
  }
  return false;
}

function isBlockedHost(host: string, allowPrivateTargets: boolean): boolean {
  if (allowPrivateTargets) {
    return false;
  }

  if (host === "localhost" || host.endsWith(".localhost") || host.endsWith(".local")) {
    return true;
  }
  if (
    host === "metadata" ||
    host === "metadata.google.internal" ||
    host === "metadata.goog" ||
    host === "instance-data" ||
    host === "169.254.169.254" ||
    host === "100.100.100.200"
  ) {
    return true;
  }

  if (net.isIP(host) === 4) {
    return isBlockedIPv4(host);
  }
  if (net.isIP(host) === 6) {
    return isBlockedIPv6(host);
  }
  return false;
}

async function validateResolvedTarget(
  host: string,
  allowPrivateTargets: boolean
): Promise<{ ok: true } | { ok: false; reason: string }> {
  if (allowPrivateTargets || net.isIP(host) !== 0) {
    return { ok: true };
  }

  let records: Array<{ address: string }>;
  try {
    records = await dns.lookup(host, { all: true, verbatim: true });
  } catch {
    return { ok: false, reason: `host ${host} could not be resolved by DNS` };
  }

  if (records.length === 0) {
    return { ok: false, reason: `host ${host} did not resolve to any IP addresses` };
  }

  for (const record of records) {
    if (isBlockedHost(record.address, false)) {
      return {
        ok: false,
        reason: `host ${host} resolves to a blocked private, loopback, link-local, multicast, unspecified, or metadata IP address`
      };
    }
  }

  return { ok: true };
}

function isBlockedIPv4(host: string): boolean {
  const parts = host.split(".").map((part) => Number(part));
  if (parts.length !== 4 || parts.some((part) => Number.isNaN(part))) {
    return true;
  }
  const [a, b] = parts;
  return (
    a === 0 ||
    a === 10 ||
    a === 127 ||
    (a === 169 && b === 254) ||
    (a === 172 && b >= 16 && b <= 31) ||
    (a === 192 && b === 168)
  );
}

function isBlockedIPv6(host: string): boolean {
  const normalized = host.toLowerCase();
  if (normalized.startsWith("::ffff:")) {
    return isBlockedIPv4(normalized.replace("::ffff:", ""));
  }
  return normalized === "::1" || normalized === "::" || normalized.startsWith("fc") || normalized.startsWith("fd") || normalized.startsWith("fe80");
}

function sanitizeURL(raw: string): string {
  try {
    const parsed = new URL(raw);
    parsed.username = "";
    parsed.password = "";
    parsed.search = "";
    parsed.hash = "";
    return parsed.toString();
  } catch {
    return sanitizeText(raw);
  }
}

function sanitizeText(input: string): string {
  return input
    .replace(/(authorization|password|passwd|token|secret|api[_-]?key|cookie|session)=([^&\s]+)/gi, "$1=[REDACTED]")
    .replace(/(Bearer|Basic)\s+[A-Za-z0-9._~+/=-]+/gi, "$1 [REDACTED]")
    .slice(0, 1000);
}

function loadConfig(): Config {
  return {
    databaseUrl: env("DATABASE_URL", "postgres://qualora:qualora@localhost:5432/qualora?sslmode=disable"),
    redisUrl: env("REDIS_URL", "redis://localhost:6379"),
    queueName: env("RUN_QUEUE", "qualora:browser-runs"),
    planExecutionQueueName: env("TEST_PLAN_EXECUTION_QUEUE", "qualora:test-plan-executions"),
    evidenceDir: env("EVIDENCE_DIR", "/tmp/qualora-evidence"),
    s3Endpoint: env("S3_ENDPOINT", "http://localhost:9000"),
    s3Region: env("S3_REGION", "us-east-1"),
    s3Bucket: env("S3_BUCKET", "qualora-evidence"),
    s3AccessKeyId: env("S3_ACCESS_KEY_ID", "qualora"),
    s3SecretAccessKey: env("S3_SECRET_ACCESS_KEY", "qualora-dev-secret"),
    s3ForcePathStyle: env("S3_FORCE_PATH_STYLE", "true") === "true"
  };
}

function env(key: string, fallback: string): string {
  return process.env[key] || fallback;
}

function log(message: string, fields: Record<string, unknown>): void {
  process.stdout.write(`${JSON.stringify({ level: "info", message, ...fields })}\n`);
}

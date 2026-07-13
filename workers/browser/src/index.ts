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

type Config = {
  databaseUrl: string;
  redisUrl: string;
  queueName: string;
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

type Project = {
  id: string;
  frontend_url: string;
  allowed_hosts: string[];
  allow_private_targets: boolean;
};

type FindingInput = {
  title: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  category: string;
  confidence: "high" | "medium" | "low";
  description: string;
  recommendation: string;
  evidenceIds: string[];
};

type BrowserResult = {
  pageTitle: string;
  statusCode: number | null;
  loadError: string;
  consoleErrors: Array<{ type: string; text: string; location: string }>;
  failedRequests: Array<{ url: string; method: string; failure: string }>;
  blockedRequests: Array<{ url: string; reason: string }>;
  screenshot: Buffer | null;
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
  log("worker_started", { queue: config.queueName });

  while (!stopping) {
    const item = await redis.blpop(config.queueName, 5);
    if (!item) {
      continue;
    }

    const payload = item[1];
    let job: BrowserRunJob;
    try {
      job = JSON.parse(payload) as BrowserRunJob;
    } catch {
      log("invalid_job_payload", {});
      continue;
    }

    await handleJob(job);
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
      const screenshotURI = await storeScreenshot(job.run_id, result.screenshot);
      const screenshotEvidenceID = await insertEvidence(job.run_id, "screenshot", screenshotURI, {
        page_title: result.pageTitle,
        status_code: result.statusCode
      });
      evidenceIds.push(screenshotEvidenceID);
    }

    const observationsEvidenceID = await insertEvidence(job.run_id, "browser_observations", "inline://browser-observations", {
      page_title: result.pageTitle,
      status_code: result.statusCode,
      load_error: result.loadError,
      console_errors: result.consoleErrors,
      failed_requests: result.failedRequests,
      blocked_requests: result.blockedRequests
    });
    evidenceIds.push(observationsEvidenceID);

    const findings = buildFindings(result, evidenceIds);
    for (const finding of findings) {
      await insertFinding(job.run_id, finding);
    }

    await finishJob(job, "completed", "", result.pageTitle);
    log("run_completed", { run_id: job.run_id, findings: findings.length });
  } catch (error) {
    const message = sanitizeText(error instanceof Error ? error.message : String(error));
    await insertFinding(job.run_id, {
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
  } catch (error) {
    loadError = sanitizeText(error instanceof Error ? error.message : String(error));
  }

  screenshot = await captureScreenshot(page);
  await browser.close();

  return {
    pageTitle,
    statusCode,
    loadError,
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

function buildFindings(result: BrowserResult, evidenceIds: string[]): FindingInput[] {
  const findings: FindingInput[] = [];

  if (result.loadError) {
    findings.push({
      title: "Page load failed",
      severity: "high",
      category: "frontend",
      confidence: "high",
      description: `The target page did not complete the initial browser load: ${result.loadError}`,
      recommendation: "Verify that the frontend URL is reachable from the worker container and that the application serves a valid page.",
      evidenceIds
    });
  } else if (result.statusCode !== null && result.statusCode >= 500) {
    findings.push({
      title: "Server error while loading page",
      severity: "high",
      category: "frontend",
      confidence: "high",
      description: `The target page returned HTTP ${result.statusCode}.`,
      recommendation: "Inspect the frontend service and upstream dependencies for server-side errors.",
      evidenceIds
    });
  } else if (result.statusCode !== null && result.statusCode >= 400) {
    findings.push({
      title: "Client error while loading page",
      severity: "medium",
      category: "frontend",
      confidence: "high",
      description: `The target page returned HTTP ${result.statusCode}.`,
      recommendation: "Confirm that the configured frontend URL is correct and publicly reachable within the allowed test scope.",
      evidenceIds
    });
  }

  if (result.consoleErrors.length > 0) {
    findings.push({
      title: "Console error detected",
      severity: "low",
      category: "frontend",
      confidence: "medium",
      description: `The browser observed ${result.consoleErrors.length} console error(s) during page load.`,
      recommendation: "Review browser console errors and fix uncaught frontend exceptions or failed client-side initialization.",
      evidenceIds
    });
  }

  if (result.failedRequests.length > 0) {
    findings.push({
      title: "Failed network request detected",
      severity: "medium",
      category: "frontend",
      confidence: "medium",
      description: `The browser observed ${result.failedRequests.length} failed network request(s) within the allowed scope.`,
      recommendation: "Inspect failed requests and ensure required assets, APIs, and dependencies are available during page load.",
      evidenceIds
    });
  }

  if (result.blockedRequests.length > 0) {
    findings.push({
      title: "Out-of-scope browser request blocked",
      severity: "info",
      category: "scope",
      confidence: "high",
      description: `The browser blocked ${result.blockedRequests.length} request(s) outside the project's allowed hosts.`,
      recommendation: "Add required first-party hosts to allowed_hosts or remove unexpected third-party dependencies from the smoke path.",
      evidenceIds
    });
  }

  return findings;
}

async function insertEvidence(runID: string, type: string, uri: string, metadata: Record<string, unknown>): Promise<string> {
  const id = randomUUID();
  await pool.query(
    `INSERT INTO evidence (id, run_id, type, uri, metadata)
     VALUES ($1, $2, $3, $4, $5)`,
    [id, runID, type, uri, JSON.stringify(metadata)]
  );
  return id;
}

async function insertFinding(runID: string, finding: FindingInput): Promise<void> {
  await pool.query(
    `INSERT INTO findings (id, run_id, title, severity, category, confidence, description, recommendation, evidence_ids)
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
    [
      randomUUID(),
      runID,
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

async function storeScreenshot(runID: string, screenshot: Buffer): Promise<string> {
  const key = `runs/${runID}/screenshots/${Date.now()}.png`;

  try {
    await putScreenshotObject(key, screenshot);
    return `s3://${config.s3Bucket}/${key}`;
  } catch (error) {
    try {
      await ensureS3Bucket();
      await putScreenshotObject(key, screenshot);
      return `s3://${config.s3Bucket}/${key}`;
    } catch {
      log("s3_put_failed_using_local_fallback", { error: sanitizeText(error instanceof Error ? error.message : String(error)) });
    }

    const localDir = path.join(config.evidenceDir, "runs", runID, "screenshots");
    await fs.mkdir(localDir, { recursive: true });
    const localPath = path.join(localDir, `${Date.now()}.png`);
    await fs.writeFile(localPath, screenshot);
    return `file://${localPath}`;
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
    evidenceDir: env("EVIDENCE_DIR", "/tmp/qualora-evidence"),
    s3Endpoint: env("S3_ENDPOINT", "http://localhost:9000"),
    s3Region: env("S3_REGION", "us-east-1"),
    s3Bucket: env("S3_BUCKET", "qualora-evidence"),
    s3AccessKeyId: env("S3_ACCESS_KEY_ID", "qualora"),
    s3SecretAccessKey: env("S3_SECRET_ACCESS_KEY", "qualora-secret"),
    s3ForcePathStyle: env("S3_FORCE_PATH_STYLE", "true") === "true"
  };
}

function env(key: string, fallback: string): string {
  return process.env[key] || fallback;
}

function log(message: string, fields: Record<string, unknown>): void {
  process.stdout.write(`${JSON.stringify({ level: "info", message, ...fields })}\n`);
}

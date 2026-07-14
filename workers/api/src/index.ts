import { randomUUID } from "node:crypto";
import Redis from "ioredis";
import { Pool } from "pg";
import {
  buildFindingsForObservation,
  buildStackTraceFinding,
  endpointURL,
  looksLikeStackTrace,
  parseOpenAPIDocument,
  sanitizeText,
  sanitizeURL,
  validateTargetURL,
  type EndpointCheck,
  type EvidenceInput,
  type FindingInput,
  type OpenAPISummary,
  type Project,
  type RequestObservation
} from "./checks";

type Config = {
  databaseUrl: string;
  redisUrl: string;
  queueName: string;
  requestTimeoutMs: number;
  maxOpenAPIEndpoints: number;
};

type APIRunJob = {
  job_id: string;
  run_id: string;
  project_id: string;
};

type RequestResult = RequestObservation & {
  bodyPreview: string;
  rawBody: string;
};

const config = loadConfig();
const pool = new Pool({ connectionString: config.databaseUrl });
const redis = new Redis(config.redisUrl, { maxRetriesPerRequest: null });

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
  log("worker_started", { queue: config.queueName });

  while (!stopping) {
    const item = await redis.blpop(config.queueName, 5);
    if (!item) {
      continue;
    }

    let job: APIRunJob;
    try {
      job = JSON.parse(item[1]) as APIRunJob;
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

async function handleJob(job: APIRunJob): Promise<void> {
  log("run_started", { run_id: job.run_id, project_id: job.project_id });

  try {
    await markJobRunning(job);
    const project = await getProject(job.project_id);

    if (project.destructive_actions) {
      throw new Error("destructive_actions=true is not supported by the v0.4.0-alpha API worker");
    }

    const result = await runAPIChecks(project);
    const evidenceIDs: string[] = [];

    for (const evidence of result.evidence) {
      evidenceIDs.push(await insertEvidence(job.run_id, evidence));
    }
    for (const finding of result.findings) {
      await insertFinding(job.run_id, { ...finding, evidenceIds: finding.evidenceIds.length > 0 ? finding.evidenceIds : evidenceIDs });
    }

    await finishJob(job, "completed", "");
    log("run_completed", { run_id: job.run_id, findings: result.findings.length, evidence: result.evidence.length });
  } catch (error) {
    const message = sanitizeText(error instanceof Error ? error.message : String(error));
    await insertFinding(job.run_id, {
      title: "API checks failed",
      severity: "high",
      category: "api",
      confidence: "medium",
      description: "The API worker could not complete the API checks.",
      recommendation: "Verify the API URLs, allowed hosts, network access from the worker container, and API availability.",
      evidenceIds: []
    }).catch(() => undefined);
    await finishJob(job, "failed", message).catch(() => undefined);
    log("run_failed", { run_id: job.run_id, error: message });
  }
}

async function runAPIChecks(project: Project): Promise<{ evidence: EvidenceInput[]; findings: FindingInput[] }> {
  const evidence: EvidenceInput[] = [];
  const findings: FindingInput[] = [];
  const observations: RequestResult[] = [];
  let checkedEndpoints = 0;
  let failedEndpoints = 0;

  if (project.api_base_url) {
    await validateTargetURL(project.api_base_url, project.allowed_hosts, project.allow_private_targets);
    const observation = await requestEndpoint("GET", project.api_base_url, [], []);
    observations.push(observation);
    checkedEndpoints++;
    if (observation.error || (observation.statusCode !== null && observation.statusCode >= 400)) {
      failedEndpoints++;
    }
    evidence.push(apiRequestEvidence(observation));
    findings.push(...buildFindingsForObservation(observation, []));
    if (observation.bodyPreview && looksLikeStackTrace(observation.bodyPreview, observation.contentType)) {
      findings.push(buildStackTraceFinding(observation.url, []));
    }
  }

  if (project.openapi_url) {
    await validateTargetURL(project.openapi_url, project.allowed_hosts, project.allow_private_targets);
    const specObservation = await requestEndpoint("GET", project.openapi_url, ["200"], ["application/json", "application/yaml", "text/yaml", "application/x-yaml"]);
    observations.push(specObservation);
    evidence.push(apiRequestEvidence(specObservation));
    if (specObservation.error || specObservation.statusCode === null || specObservation.statusCode >= 400) {
      findings.push({
        title: "OpenAPI document unreachable",
        severity: "high",
        category: "contract",
        confidence: "high",
        description: `The OpenAPI document could not be fetched successfully from ${sanitizeURL(project.openapi_url)}.`,
        recommendation: "Verify the OpenAPI URL, allowed hosts, and API availability.",
        evidenceIds: []
      });
    } else {
      try {
        const summary = parseOpenAPIDocument(specObservation.rawBody, specObservation.contentType);
        const endpointChecks = buildEndpointChecks(project, summary).slice(0, config.maxOpenAPIEndpoints);
        checkedEndpoints += endpointChecks.length;
        evidence.push({
          type: "openapi_summary",
          uri: "inline://openapi-summary",
          metadata: {
            openapi_url: sanitizeURL(project.openapi_url),
            version: summary.version,
            paths: summary.pathCount,
            operations: summary.operationCount,
            safe_operations: summary.safeOperationCount,
            skipped_unsafe_operations: summary.skippedUnsafeOperationCount,
            checked_endpoints: endpointChecks.length,
            safe_methods_only: true
          }
        });

        for (const endpoint of endpointChecks) {
          await validateTargetURL(endpoint.url, project.allowed_hosts, project.allow_private_targets);
          const observation = await requestEndpoint(endpoint.method, endpoint.url, endpoint.expectedStatuses, endpoint.expectedContentTypes);
          observations.push(observation);
          evidence.push(apiRequestEvidence(observation));
          if (observation.error || (observation.statusCode !== null && observation.statusCode >= 400)) {
            failedEndpoints++;
          }
          findings.push(...buildFindingsForObservation(observation, []));
          if (observation.bodyPreview && looksLikeStackTrace(observation.bodyPreview, observation.contentType)) {
            findings.push(buildStackTraceFinding(observation.url, []));
          }
        }
      } catch (error) {
        findings.push({
          title: "Invalid OpenAPI document",
          severity: "medium",
          category: "contract",
          confidence: "high",
          description: sanitizeText(error instanceof Error ? error.message : String(error)),
          recommendation: "Publish a valid OpenAPI 3.x JSON or YAML document.",
          evidenceIds: []
        });
      }
    }
  }

  evidence.unshift({
    type: "api_observations",
    uri: "inline://api-observations",
    metadata: {
      api_base_url: project.api_base_url ? sanitizeURL(project.api_base_url) : "",
      openapi_url: project.openapi_url ? sanitizeURL(project.openapi_url) : "",
      checked_endpoints: checkedEndpoints,
      failed_endpoints: failedEndpoints,
      safe_methods_only: true,
      observations: observations.map(({ bodyPreview: _bodyPreview, rawBody: _rawBody, ...observation }) => observation)
    }
  });

  return { evidence, findings };
}

function buildEndpointChecks(project: Project, summary: OpenAPISummary): EndpointCheck[] {
  const baseURL = project.api_base_url || new URL(project.openapi_url).origin;
  return summary.endpoints.map((endpoint) => ({
    ...endpoint,
    url: endpointURL(baseURL, endpoint.path)
  }));
}

async function requestEndpoint(
  method: string,
  url: string,
  expectedStatuses: string[],
  expectedContentTypes: string[]
): Promise<RequestResult> {
  const started = Date.now();
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), config.requestTimeoutMs);

  try {
    const response = await fetch(url, {
      method,
      redirect: "manual",
      signal: controller.signal,
      headers: {
        "User-Agent": "Qualora API Worker v0.4.0-alpha",
        Accept: "application/json, application/yaml, text/yaml, text/plain, */*"
      }
    });
    const contentType = response.headers.get("content-type") ?? "";
    const rawBody = method === "HEAD" ? "" : (await response.text()).slice(0, 200_000);
    const bodyPreview = sanitizeText(rawBody);
    return {
      method,
      url: sanitizeURL(url),
      statusCode: response.status,
      contentType,
      responseTimeMs: Date.now() - started,
      error: "",
      expectedStatuses,
      expectedContentTypes,
      bodyPreview,
      rawBody
    };
  } catch (error) {
    return {
      method,
      url: sanitizeURL(url),
      statusCode: null,
      contentType: "",
      responseTimeMs: Date.now() - started,
      error: sanitizeText(error instanceof Error ? error.message : String(error)),
      expectedStatuses,
      expectedContentTypes,
      bodyPreview: "",
      rawBody: ""
    };
  } finally {
    clearTimeout(timeout);
  }
}

function apiRequestEvidence(observation: RequestObservation): EvidenceInput {
  const { bodyPreview: _bodyPreview, rawBody: _rawBody, ...metadata } = observation as RequestResult;
  return {
    type: "api_request",
    uri: observation.url,
    metadata
  };
}

async function getProject(projectID: string): Promise<Project> {
  const result = await pool.query(
    `SELECT id, api_base_url, openapi_url, allowed_hosts, destructive_actions, allow_private_targets
     FROM projects
     WHERE id = $1`,
    [projectID]
  );
  if (result.rowCount !== 1) {
    throw new Error("project was not found");
  }

  const row = result.rows[0] as {
    id: string;
    api_base_url: string;
    openapi_url: string;
    allowed_hosts: string[] | string;
    destructive_actions: boolean;
    allow_private_targets: boolean;
  };

  return {
    id: row.id,
    api_base_url: row.api_base_url,
    openapi_url: row.openapi_url,
    allowed_hosts: Array.isArray(row.allowed_hosts) ? row.allowed_hosts : JSON.parse(row.allowed_hosts),
    destructive_actions: row.destructive_actions,
    allow_private_targets: row.allow_private_targets
  };
}

async function markJobRunning(job: APIRunJob): Promise<void> {
  if (!job.job_id) {
    throw new Error("api job is missing job_id");
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

async function finishJob(job: APIRunJob, status: "completed" | "failed", errorMessage: string): Promise<void> {
  await pool.query(
    `UPDATE run_jobs
     SET status = $3, error_message = $4, completed_at = now(), updated_at = now()
     WHERE id = $1 AND run_id = $2`,
    [job.job_id, job.run_id, status, errorMessage]
  );
  await pool.query(`SELECT refresh_test_run_status($1)`, [job.run_id]);
}

async function insertEvidence(runID: string, evidence: EvidenceInput): Promise<string> {
  const id = randomUUID();
  await pool.query(
    `INSERT INTO evidence (id, run_id, type, uri, metadata)
     VALUES ($1, $2, $3, $4, $5)`,
    [id, runID, evidence.type, evidence.uri, JSON.stringify(evidence.metadata)]
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

function loadConfig(): Config {
  return {
    databaseUrl: env("DATABASE_URL", "postgres://qualora:qualora@localhost:5432/qualora?sslmode=disable"),
    redisUrl: env("REDIS_URL", "redis://localhost:6379"),
    queueName: env("API_RUN_QUEUE", "qualora:api-runs"),
    requestTimeoutMs: Number(env("API_REQUEST_TIMEOUT_MS", "10000")),
    maxOpenAPIEndpoints: Number(env("MAX_OPENAPI_ENDPOINTS", "25"))
  };
}

function env(key: string, fallback: string): string {
  return process.env[key] || fallback;
}

function log(message: string, fields: Record<string, unknown>): void {
  process.stdout.write(`${JSON.stringify({ level: "info", message, ...fields })}\n`);
}

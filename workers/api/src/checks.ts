import dns from "node:dns/promises";
import net from "node:net";
import { parse as parseYAML } from "yaml";

export type Project = {
  id: string;
  api_base_url: string;
  openapi_url: string;
  allowed_hosts: string[];
  destructive_actions: boolean;
  allow_private_targets: boolean;
};

export type FindingInput = {
  title: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  category: string;
  confidence: "high" | "medium" | "low";
  description: string;
  recommendation: string;
  evidenceIds: string[];
};

export type EvidenceInput = {
  type: "api_observations" | "openapi_summary" | "api_request";
  uri: string;
  metadata: Record<string, unknown>;
};

export type EndpointCheck = {
  method: string;
  path: string;
  url: string;
  expectedStatuses: string[];
  expectedContentTypes: string[];
};

export type RequestObservation = {
  method: string;
  url: string;
  statusCode: number | null;
  contentType: string;
  responseTimeMs: number;
  error: string;
  expectedStatuses: string[];
  expectedContentTypes: string[];
};

export type OpenAPISummary = {
  version: string;
  pathCount: number;
  operationCount: number;
  safeOperationCount: number;
  skippedUnsafeOperationCount: number;
  endpoints: EndpointCheck[];
};

export type APICheckResult = {
  evidence: EvidenceInput[];
  findings: FindingInput[];
};

const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);
const UNSAFE_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);
const STACK_TRACE_PATTERNS = [
  /exception/i,
  /stack trace/i,
  /traceback \(most recent call last\)/i,
  /\bat\s+\S+\s+\(.+:\d+:\d+\)/,
  /java\.lang\./i,
  /goroutine \d+ \[/i
];

export function safeMethodsOnly(methods: string[], destructiveActions: boolean): string[] {
  return methods
    .map((method) => method.toUpperCase())
    .filter((method) => SAFE_METHODS.has(method) || (destructiveActions && UNSAFE_METHODS.has(method)));
}

export function buildFindingsForObservation(observation: RequestObservation, evidenceIds: string[]): FindingInput[] {
  const findings: FindingInput[] = [];

  if (observation.error) {
    findings.push({
      title: "API endpoint unreachable",
      severity: "high",
      category: "api",
      confidence: "high",
      description: `${observation.method} ${observation.url} could not be reached: ${observation.error}`,
      recommendation: "Verify DNS, TLS, networking, service availability, and the configured API URL.",
      evidenceIds
    });
    return findings;
  }

  if (observation.statusCode !== null && observation.statusCode >= 500) {
    findings.push({
      title: "API endpoint returned 5xx",
      severity: "high",
      category: "api",
      confidence: "high",
      description: `${observation.method} ${observation.url} returned HTTP ${observation.statusCode}.`,
      recommendation: "Inspect the API service logs and upstream dependencies for server-side failures.",
      evidenceIds
    });
  }

  if (
    observation.expectedStatuses.length > 0 &&
    observation.statusCode !== null &&
    !statusMatchesDeclaredResponses(observation.statusCode, observation.expectedStatuses)
  ) {
    findings.push({
      title: "API endpoint returned unexpected status code",
      severity: observation.statusCode >= 500 ? "high" : "medium",
      category: "contract",
      confidence: "medium",
      description: `${observation.method} ${observation.url} returned HTTP ${observation.statusCode}, which is not declared in the OpenAPI responses.`,
      recommendation: "Update the OpenAPI document or adjust the endpoint behavior to match the documented responses.",
      evidenceIds
    });
  }

  if (
    observation.expectedContentTypes.length > 0 &&
    observation.contentType &&
    !contentTypeMatches(observation.contentType, observation.expectedContentTypes)
  ) {
    findings.push({
      title: "API endpoint returned unexpected content type",
      severity: "low",
      category: "contract",
      confidence: "medium",
      description: `${observation.method} ${observation.url} returned content type ${observation.contentType}, which does not obviously match the OpenAPI response content types.`,
      recommendation: "Verify the endpoint response content type or update the OpenAPI response content definitions.",
      evidenceIds
    });
  }

  return findings;
}

export function buildStackTraceFinding(url: string, evidenceIds: string[]): FindingInput {
  return {
    title: "API response appears to expose a stack trace",
    severity: "medium",
    category: "api",
    confidence: "medium",
    description: `The response body from ${url} appears to contain stack trace or exception details.`,
    recommendation: "Return sanitized error responses to clients and keep stack traces in server-side logs.",
    evidenceIds
  };
}

export function parseOpenAPIDocument(raw: string, contentType: string): OpenAPISummary {
  let doc: unknown;
  try {
    if (contentType.includes("json") || raw.trimStart().startsWith("{")) {
      doc = JSON.parse(raw);
    } else {
      doc = parseYAML(raw);
    }
  } catch (error) {
    throw new Error(`OpenAPI document could not be parsed: ${error instanceof Error ? error.message : String(error)}`);
  }

  if (!isRecord(doc)) {
    throw new Error("OpenAPI document must be an object");
  }
  const version = typeof doc.openapi === "string" ? doc.openapi : "";
  if (!version.startsWith("3.")) {
    throw new Error("only OpenAPI 3.x documents are supported in this alpha");
  }
  if (!isRecord(doc.paths)) {
    throw new Error("OpenAPI document must include a paths object");
  }

  const endpoints: EndpointCheck[] = [];
  let operationCount = 0;
  let skippedUnsafeOperationCount = 0;

  for (const [pathName, pathItem] of Object.entries(doc.paths)) {
    if (!isRecord(pathItem)) {
      continue;
    }
    for (const [methodName, operation] of Object.entries(pathItem)) {
      const method = methodName.toUpperCase();
      if (!SAFE_METHODS.has(method) && !UNSAFE_METHODS.has(method)) {
        continue;
      }
      operationCount++;
      if (!SAFE_METHODS.has(method)) {
        skippedUnsafeOperationCount++;
        continue;
      }
      const expectedStatuses = extractExpectedStatuses(operation);
      const expectedContentTypes = extractExpectedContentTypes(operation);
      endpoints.push({
        method,
        path: pathName,
        url: "",
        expectedStatuses,
        expectedContentTypes
      });
    }
  }

  return {
    version,
    pathCount: Object.keys(doc.paths).length,
    operationCount,
    safeOperationCount: endpoints.length,
    skippedUnsafeOperationCount,
    endpoints
  };
}

export function endpointURL(baseURL: string, endpointPath: string): string {
  const base = new URL(baseURL);
  const normalizedBasePath = base.pathname.endsWith("/") ? base.pathname.slice(0, -1) : base.pathname;
  const normalizedEndpointPath = endpointPath.startsWith("/") ? endpointPath : `/${endpointPath}`;
  base.pathname = `${normalizedBasePath}${normalizedEndpointPath}`;
  base.search = "";
  base.hash = "";
  return base.toString();
}

export function looksLikeStackTrace(body: string, contentType: string): boolean {
  if (!body || body.length > 200_000) {
    return false;
  }
  const lowerContentType = contentType.toLowerCase();
  if (!lowerContentType.includes("text") && !lowerContentType.includes("json") && !lowerContentType.includes("html")) {
    return false;
  }
  return STACK_TRACE_PATTERNS.some((pattern) => pattern.test(body));
}

export async function validateTargetURL(raw: string, allowedHosts: string[], allowPrivateTargets: boolean): Promise<void> {
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    throw new Error("URL is invalid");
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new Error("only http and https URLs are supported");
  }

  const host = parsed.hostname.toLowerCase().replace(/\.$/, "");
  if (!host) {
    throw new Error("host is required");
  }
  if (isBlockedHost(host, allowPrivateTargets)) {
    throw new Error(`host ${host} is blocked by the default safety policy`);
  }
  if (!allowedHosts.some((allowedHost) => hostAllowed(host, allowedHost))) {
    throw new Error(`host ${host} is not present in allowed_hosts`);
  }
  await validateResolvedTarget(host, allowPrivateTargets);
}

export function sanitizeText(input: string): string {
  return input
    .replace(/(authorization|password|passwd|token|secret|api[_-]?key|cookie|session)=([^&\s]+)/gi, "$1=[REDACTED]")
    .replace(/(Bearer|Basic)\s+[A-Za-z0-9._~+/=-]+/gi, "$1 [REDACTED]")
    .slice(0, 1000);
}

export function sanitizeURL(raw: string): string {
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

function extractExpectedStatuses(operation: unknown): string[] {
  if (!isRecord(operation) || !isRecord(operation.responses)) {
    return [];
  }
  return Object.keys(operation.responses);
}

function extractExpectedContentTypes(operation: unknown): string[] {
  if (!isRecord(operation) || !isRecord(operation.responses)) {
    return [];
  }
  const contentTypes = new Set<string>();
  for (const response of Object.values(operation.responses)) {
    if (!isRecord(response) || !isRecord(response.content)) {
      continue;
    }
    for (const contentType of Object.keys(response.content)) {
      contentTypes.add(contentType.toLowerCase());
    }
  }
  return Array.from(contentTypes);
}

function statusMatchesDeclaredResponses(statusCode: number, expectedStatuses: string[]): boolean {
  const status = String(statusCode);
  return expectedStatuses.some((expected) => {
    if (expected === status || expected === "default") {
      return true;
    }
    return expected.length === 3 && expected.endsWith("XX") && expected[0] === status[0];
  });
}

function contentTypeMatches(actual: string, expectedContentTypes: string[]): boolean {
  const normalizedActual = actual.split(";")[0].trim().toLowerCase();
  return expectedContentTypes.some((expected) => {
    const normalizedExpected = expected.split(";")[0].trim().toLowerCase();
    return normalizedActual === normalizedExpected || normalizedActual.endsWith(`+${normalizedExpected.split("/").pop() ?? ""}`);
  });
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
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

async function validateResolvedTarget(host: string, allowPrivateTargets: boolean): Promise<void> {
  if (allowPrivateTargets || net.isIP(host) !== 0) {
    return;
  }

  let records: Array<{ address: string }>;
  try {
    records = await dns.lookup(host, { all: true, verbatim: true });
  } catch {
    throw new Error(`host ${host} could not be resolved by DNS`);
  }

  if (records.length === 0) {
    throw new Error(`host ${host} did not resolve to any IP addresses`);
  }

  for (const record of records) {
    if (isBlockedHost(record.address, false)) {
      throw new Error(`host ${host} resolves to a blocked private, loopback, link-local, multicast, unspecified, or metadata IP address`);
    }
  }
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

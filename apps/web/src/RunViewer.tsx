import { useCallback, useEffect, useMemo, useState } from "react";
import {
  API_BASE_URL,
  aiBrowserControlHTMLReportURL,
  evidenceDownloadURL,
  getAIBrowserControlReport,
  getSafeExplorerReport,
  safeExplorerHTMLReportURL
} from "./api";
import type {
  AIBrowserControlReport,
  AIBrowserControlStep,
  SafeExplorerReport,
  SafeExplorerStep
} from "./types";

export type RunViewerType = "ai-browser-control" | "safe-explorer";

type ViewerStep = {
  id: string;
  index: number;
  pageURL: string;
  actionType: string;
  actionLabel: string;
  policyDecision: string;
  executionStatus: string;
  reason: string;
  aiSuggestion?: Record<string, unknown>;
  sanitizedObservation?: Record<string, unknown>;
  screenshotEvidenceID?: string;
  consoleErrors: number;
  failedRequests: number;
};

type ViewerModel = {
  type: RunViewerType;
  typeLabel: string;
  runID: string;
  status: string;
  projectID: string;
  projectName: string;
  startURL: string;
  startedAt?: string;
  completedAt?: string;
  steps: ViewerStep[];
  findingsCount: number;
  jsonURL: string;
  htmlURL: string;
  reportURL: string;
};

export function RunViewer({ runType, runID }: { runType: RunViewerType; runID: string }) {
  const [model, setModel] = useState<ViewerModel>();
  const [selectedStepID, setSelectedStepID] = useState("");
  const [followLatest, setFollowLatest] = useState(true);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    setError("");
    try {
      const next = runType === "ai-browser-control"
        ? aiBrowserViewerModel(await getAIBrowserControlReport(runID))
        : safeExplorerViewerModel(await getSafeExplorerReport(runID));
      setModel(next);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : String(caught));
    } finally {
      setLoading(false);
    }
  }, [runID, runType]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!model || !isActiveStatus(model.status)) {
      return undefined;
    }
    const timer = window.setInterval(() => void refresh(), 2500);
    return () => window.clearInterval(timer);
  }, [model, refresh]);

  useEffect(() => {
    const latestStepID = model?.steps.at(-1)?.id;
    if (latestStepID && (followLatest || !model?.steps.some((step) => step.id === selectedStepID))) {
      setSelectedStepID(latestStepID);
    }
  }, [followLatest, model, selectedStepID]);

  const selectedStep = useMemo(
    () => model?.steps.find((step) => step.id === selectedStepID) || model?.steps.at(-1),
    [model, selectedStepID]
  );

  if (loading && !model) {
    return <div className="viewer-loading">Loading Run Viewer...</div>;
  }
  if (error && !model) {
    return <div className="notice danger">{error}</div>;
  }
  if (!model) {
    return <div className="notice danger">Run Viewer could not load this browser run.</div>;
  }

  const latestStep = model.steps.at(-1);
  const running = isActiveStatus(model.status);

  return (
    <div className="run-viewer">
      <section className="run-viewer-header">
        <div>
          <p className="eyebrow">Observable browser run</p>
          <h2>Run Viewer</h2>
          <p className="muted">
            <span className={`status ${model.status}`}>{running ? "Running..." : model.status}</span>
            {" "}{model.typeLabel} for {model.projectName}
          </p>
        </div>
        <div className="button-row">
          <a className="button secondary-link" href={`#/projects/${model.projectID}`}>Back to Project Cockpit</a>
          <a className="button secondary-link" href={model.reportURL}>Open report</a>
          <a className="button secondary-link" href={model.jsonURL} target="_blank" rel="noreferrer">Raw JSON</a>
          <a className="button" href={model.htmlURL} target="_blank" rel="noreferrer">HTML report</a>
        </div>
      </section>

      {error && <div className="notice danger">{error}</div>}

      <div className="viewer-status-grid">
        <ViewerMetric label="Run status" value={running ? "Running..." : model.status} />
        <ViewerMetric label="Run type" value={model.typeLabel} />
        <ViewerMetric label="Current/latest step" value={latestStep ? `${latestStep.index + 1} of ${model.steps.length}` : "Waiting for first step"} />
        <ViewerMetric label="Console errors" value={String(model.steps.reduce((total, step) => total + step.consoleErrors, 0))} />
        <ViewerMetric label="Failed requests" value={String(model.steps.reduce((total, step) => total + step.failedRequests, 0))} />
        <ViewerMetric label="Findings" value={String(model.findingsCount)} />
      </div>
      <div className="viewer-context">
        <p><strong>Start URL:</strong> <code>{model.startURL}</code></p>
        <p><strong>Started:</strong> {formatViewerDate(model.startedAt)}</p>
        <p><strong>Completed:</strong> {model.completedAt ? formatViewerDate(model.completedAt) : "Not completed"}</p>
      </div>

      {model.type === "ai-browser-control" && (
        <div className="notice info">
          AI proposes one action at a time. Qualora validates it through policy before Playwright executes anything.
        </div>
      )}

      <div className="run-viewer-layout">
        <section className="viewer-timeline-panel">
          <div className="section-heading">
            <div>
              <h2>Step timeline</h2>
              <p>{running ? "Near-live updates every 2.5 seconds." : "Replay the recorded browser activity step by step."}</p>
            </div>
            <div className="button-row compact">
              {running && !followLatest && (
                <button type="button" className="secondary" onClick={() => setFollowLatest(true)}>Follow latest</button>
              )}
              <button type="button" className="secondary" onClick={() => void refresh()}>Refresh</button>
            </div>
          </div>
          {model.steps.length === 0 ? (
            <div className="empty-state">
              <h3>Waiting for browser activity</h3>
              <p>The viewer will update when the worker records its first safe step.</p>
            </div>
          ) : (
            <div className="viewer-step-list">
              {model.steps.map((step) => (
                <button
                  key={step.id}
                  type="button"
                  className={`viewer-step ${selectedStep?.id === step.id ? "selected" : ""}`}
                  onClick={() => {
                    setSelectedStepID(step.id);
                    setFollowLatest(step.id === latestStep?.id);
                  }}
                >
                  <span className="viewer-step-number">{step.index + 1}</span>
                  <span>
                    <strong>{step.actionLabel || step.actionType || "Observe page"}</strong>
                    <small>{step.executionStatus} · {step.policyDecision || "observed"}</small>
                  </span>
                </button>
              ))}
            </div>
          )}
        </section>

        <section className="viewer-detail-panel">
          <div className="section-heading">
            <div>
              <h2>Latest screenshot</h2>
              <p>{selectedStep ? `Step ${selectedStep.index + 1}` : "No step selected"}</p>
            </div>
            {selectedStep?.screenshotEvidenceID && (
              <a href={evidenceDownloadURL(selectedStep.screenshotEvidenceID)} target="_blank" rel="noreferrer">View screenshot</a>
            )}
          </div>
          {selectedStep?.screenshotEvidenceID ? (
            <a className="viewer-screenshot-link" href={evidenceDownloadURL(selectedStep.screenshotEvidenceID)} target="_blank" rel="noreferrer">
              <img src={evidenceDownloadURL(selectedStep.screenshotEvidenceID)} alt={`Browser step ${selectedStep.index + 1}`} />
            </a>
          ) : (
            <div className="viewer-screenshot-empty">No screenshot recorded for this step.</div>
          )}

          {selectedStep && (
            <>
              <div className="viewer-step-detail">
                <ViewerMetric label="Action type" value={selectedStep.actionType || "observe"} />
                <ViewerMetric label="Action label/path" value={selectedStep.actionLabel || selectedStep.pageURL || "Not recorded"} />
                <div className="viewer-metric">
                  <span>Policy decision</span>
                  <strong><span className={`status ${policyStatusClass(selectedStep.policyDecision)}`}>{selectedStep.policyDecision || "observed"}</span></strong>
                </div>
                <ViewerMetric label="Execution result" value={selectedStep.executionStatus || "observed"} />
              </div>
              <p className="viewer-url"><strong>Page:</strong> <code>{selectedStep.pageURL}</code></p>
              {selectedStep.reason && <div className="notice info"><strong>Blocked/skipped reason:</strong> {selectedStep.reason}</div>}
              {selectedStep.aiSuggestion && (
                <details className="viewer-json" open>
                  <summary>AI suggestion</summary>
                  <pre>{JSON.stringify(selectedStep.aiSuggestion, null, 2)}</pre>
                </details>
              )}
              {selectedStep.sanitizedObservation && (
                <details className="viewer-json">
                  <summary>Sanitized observation summary</summary>
                  <pre>{JSON.stringify(selectedStep.sanitizedObservation, null, 2)}</pre>
                </details>
              )}
            </>
          )}
        </section>
      </div>
    </div>
  );
}

function aiBrowserViewerModel(report: AIBrowserControlReport): ViewerModel {
  return {
    type: "ai-browser-control",
    typeLabel: "AI Browser Control",
    runID: report.run.id,
    status: report.run.status,
    projectID: report.project.id,
    projectName: report.project.name,
    startURL: report.run.start_url,
    startedAt: report.run.started_at,
    completedAt: report.run.completed_at,
    steps: report.steps.map(aiBrowserStep),
    findingsCount: report.findings.length,
    jsonURL: `${API_BASE_URL}/api/v1/ai-browser-control-runs/${report.run.id}/report`,
    htmlURL: aiBrowserControlHTMLReportURL(report.run.id),
    reportURL: `#/ai-browser-control-runs/${report.run.id}`
  };
}

function aiBrowserStep(step: AIBrowserControlStep): ViewerStep {
  return {
    id: step.id,
    index: step.step_index,
    pageURL: step.final_url || step.action_target_url || step.normalized_url || step.page_url,
    actionType: step.action_type || suggestionActionType(step.ai_suggestion),
    actionLabel: step.action_label || suggestionLabel(step.ai_suggestion),
    policyDecision: step.policy_decision,
    executionStatus: step.execution_status,
    reason: step.policy_reason || "",
    aiSuggestion: step.ai_suggestion,
    sanitizedObservation: step.sanitized_observation,
    screenshotEvidenceID: step.screenshot_evidence_id,
    consoleErrors: step.console_error_count,
    failedRequests: step.failed_request_count
  };
}

function safeExplorerViewerModel(report: SafeExplorerReport): ViewerModel {
  return {
    type: "safe-explorer",
    typeLabel: "Safe Explorer",
    runID: report.run.id,
    status: report.run.status,
    projectID: report.project.id,
    projectName: report.project.name,
    startURL: report.run.start_url,
    startedAt: report.run.started_at,
    completedAt: report.run.completed_at,
    steps: report.steps.map(safeExplorerStep),
    findingsCount: report.findings.length,
    jsonURL: `${API_BASE_URL}/api/v1/safe-explorer-runs/${report.run.id}/report`,
    htmlURL: safeExplorerHTMLReportURL(report.run.id),
    reportURL: `#/safe-explorer-runs/${report.run.id}`
  };
}

function safeExplorerStep(step: SafeExplorerStep): ViewerStep {
  return {
    id: step.id,
    index: step.step_index,
    pageURL: step.final_url || step.action_target_url || step.normalized_url || step.page_url,
    actionType: step.action_type || "observe",
    actionLabel: step.action_label || step.page_title || "",
    policyDecision: step.action_decision,
    executionStatus: step.result_status,
    reason: step.skip_reason || "",
    screenshotEvidenceID: step.screenshot_evidence_id,
    consoleErrors: step.console_error_count,
    failedRequests: step.failed_request_count
  };
}

function suggestionActionType(suggestion: Record<string, unknown>): string {
  const action = suggestion.action;
  return action && typeof action === "object" ? String((action as Record<string, unknown>).type || "") : "";
}

function suggestionLabel(suggestion: Record<string, unknown>): string {
  const action = suggestion.action;
  if (action && typeof action === "object") {
    const value = action as Record<string, unknown>;
    return String(value.label || value.link_text || value.target_url || "");
  }
  return String(suggestion.rationale || "");
}

function ViewerMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="viewer-metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function isActiveStatus(status: string): boolean {
  return ["queued", "pending", "running", "processing"].includes(status);
}

function policyStatusClass(decision: string): string {
  if (decision === "approved" || decision === "executed" || decision === "observed") {
    return "completed";
  }
  if (decision === "blocked" || decision === "invalid") {
    return "failed";
  }
  return "pending";
}

function formatViewerDate(value?: string): string {
  if (!value) {
    return "Not started";
  }
  return new Date(value).toLocaleString();
}

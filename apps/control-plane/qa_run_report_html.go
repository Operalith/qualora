package main

import (
	"html/template"
	"io"
)

type qaRunHTMLReportData struct {
	Report  *QARunReport
	Summary ReportSummary
}

var qaRunHTMLReportTemplate = template.Must(template.New("qa-run-report").Funcs(template.FuncMap{
	"json":               prettyJSON,
	"formatTime":         formatReportTime,
	"add":                func(left int, right int) int { return left + right },
	"intValue":           optionalIntValue,
	"reportIntelligence": reportIntelligenceHTML,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Report.Project.Name }} - Qualora safe QA report</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f6f7f9;
      --panel: #ffffff;
      --text: #18202a;
      --muted: #5c6675;
      --line: #d8dee8;
      --strong: #0d5b57;
      --critical: #7f1d1d;
      --high: #b42318;
      --medium: #b54708;
      --low: #175cd3;
      --info: #475467;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      background: var(--bg);
      color: var(--text);
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 14px;
      line-height: 1.5;
    }
    main { max-width: 1120px; margin: 0 auto; padding: 32px 20px 48px; }
    header { margin-bottom: 24px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 28px; }
    h2 { font-size: 18px; margin-bottom: 12px; }
    h3 { font-size: 15px; margin-bottom: 8px; }
    section, .metric {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 16px;
    }
    .subtle { color: var(--muted); }
    .grid { display: grid; gap: 16px; }
    .grid.two { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .grid.six { grid-template-columns: repeat(6, minmax(0, 1fr)); }
    .metric span { display: block; color: var(--muted); font-size: 12px; text-transform: uppercase; }
    .metric strong { display: block; margin-top: 4px; font-size: 22px; }
    .status {
      display: inline-block;
      border: 1px solid var(--line);
      border-radius: 999px;
      padding: 2px 8px;
      background: #eef8f7;
      color: var(--strong);
      font-weight: 700;
    }
    table { width: 100%; border-collapse: collapse; }
    th, td { border-bottom: 1px solid var(--line); padding: 10px 8px; text-align: left; vertical-align: top; }
    th { color: var(--muted); font-size: 12px; text-transform: uppercase; }
    tr:last-child td { border-bottom: 0; }
    code, pre {
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
      font-size: 12px;
    }
    pre {
      margin: 0;
      white-space: pre-wrap;
      word-break: break-word;
      background: #f1f4f8;
      border: 1px solid var(--line);
      border-radius: 6px;
      padding: 10px;
      max-height: 320px;
      overflow: auto;
    }
    .severity-critical { color: var(--critical); font-weight: 700; }
    .severity-high { color: var(--high); font-weight: 700; }
    .severity-medium { color: var(--medium); font-weight: 700; }
    .severity-low { color: var(--low); font-weight: 700; }
    .severity-info { color: var(--info); font-weight: 700; }
    @media (max-width: 760px) {
      main { padding: 20px 12px 32px; }
      .grid.two, .grid.six { grid-template-columns: 1fr; }
      table { display: block; overflow-x: auto; }
    }
  </style>
</head>
<body>
<main>
  <header>
    <p class="subtle">Qualora safe QA report</p>
    <h1>{{ .Report.Project.Name }}</h1>
    <p><span class="status">{{ .Report.Run.Status }}</span> <span class="subtle">QA run {{ .Report.Run.ID }} generated {{ formatTime .Report.GeneratedAt }}</span></p>
  </header>

  <section>
    <h2>Run Summary</h2>
    <div class="grid two">
      <div>
        <p><strong>Project ID:</strong> <code>{{ .Report.Project.ID }}</code></p>
        <p><strong>QA run ID:</strong> <code>{{ .Report.Run.ID }}</code></p>
        <p><strong>Mode:</strong> {{ .Report.Run.Mode }}</p>
      </div>
      <div>
        <p><strong>Created:</strong> {{ formatTime .Report.Run.CreatedAt }}</p>
        <p><strong>Started:</strong> {{ formatTime .Report.Run.StartedAt }}</p>
        <p><strong>Completed:</strong> {{ formatTime .Report.Run.CompletedAt }}</p>
      </div>
    </div>
    {{ if .Report.Run.ErrorMessage }}<p class="severity-high"><strong>Error:</strong> {{ .Report.Run.ErrorMessage }}</p>{{ end }}
  </section>

  <div class="grid six" style="margin: 16px 0;">
    <div class="metric"><span>Findings</span><strong>{{ .Summary.TotalFindings }}</strong></div>
    <div class="metric"><span>Critical</span><strong>{{ .Summary.Critical }}</strong></div>
    <div class="metric"><span>High</span><strong>{{ .Summary.High }}</strong></div>
    <div class="metric"><span>Medium</span><strong>{{ .Summary.Medium }}</strong></div>
    <div class="metric"><span>Low</span><strong>{{ .Summary.Low }}</strong></div>
    <div class="metric"><span>Info</span><strong>{{ .Summary.Info }}</strong></div>
  </div>

  {{ reportIntelligence .Report.ReportIntelligence }}

  <section style="margin-bottom: 16px;">
    <h2>Baseline & Regression</h2>
    {{ if .Report.Baseline }}
    <p><strong>Baseline:</strong> {{ .Report.Baseline.Name }} <span class="subtle">({{ .Report.Baseline.ReportType }} report {{ .Report.Baseline.ReportID }})</span></p>
    {{ with .Report.Comparison }}
    <div class="grid six">
      <div class="metric"><span>Status</span><strong>{{ .Status }}</strong></div>
      <div class="metric"><span>New</span><strong>{{ .Summary.NewFindingsCount }}</strong></div>
      <div class="metric"><span>Fixed</span><strong>{{ .Summary.FixedFindingsCount }}</strong></div>
      <div class="metric"><span>Unchanged</span><strong>{{ .Summary.UnchangedFindingsCount }}</strong></div>
      <div class="metric"><span>New High+</span><strong>{{ add .Summary.NewCritical .Summary.NewHigh }}</strong></div>
      <div class="metric"><span>Fixed High+</span><strong>{{ add .Summary.FixedCritical .Summary.FixedHigh }}</strong></div>
    </div>
    <p>{{ .Recommendation }}</p>
    {{ end }}
    {{ with .Report.QualityGate }}
    <p><strong>Quality gate:</strong> <span class="status">{{ .Status }}</span> <strong>CI exit code:</strong> {{ .CIExitCode }}</p>
    {{ if .FailedRules }}
    <p><strong>Failed rules:</strong></p>
    <ul>{{ range .FailedRules }}<li><code>{{ . }}</code></li>{{ end }}</ul>
    {{ end }}
    {{ if .Warnings }}
    <p><strong>Warnings:</strong></p>
    <ul>{{ range .Warnings }}<li><code>{{ . }}</code></li>{{ end }}</ul>
    {{ end }}
    {{ end }}
    {{ else }}
    <p class="subtle">{{ .Report.BaselineMessage }}</p>
    {{ end }}
  </section>

  {{ with .Report.DiscoverySummary }}
  <section style="margin-bottom: 16px;">
    <h2>Discovery</h2>
    <div class="grid six">
      <div class="metric"><span>Pages</span><strong>{{ .TotalPages }}</strong></div>
      <div class="metric"><span>Links</span><strong>{{ .TotalLinks }}</strong></div>
      <div class="metric"><span>Forms</span><strong>{{ .TotalForms }}</strong></div>
      <div class="metric"><span>Skipped</span><strong>{{ .SkippedLinks }}</strong></div>
      <div class="metric"><span>Console</span><strong>{{ .TotalConsoleErrors }}</strong></div>
      <div class="metric"><span>Failed Req</span><strong>{{ .TotalFailedRequests }}</strong></div>
    </div>
  </section>
  {{ end }}

  {{ with .Report.QualitySummary }}
  <section style="margin-bottom: 16px;">
    <h2>Quality Checks</h2>
    <div class="grid six">
      <div class="metric"><span>Pages</span><strong>{{ .TotalPages }}</strong></div>
      <div class="metric"><span>Findings</span><strong>{{ .TotalFindings }}</strong></div>
      <div class="metric"><span>Security</span><strong>{{ .SecurityFindings }}</strong></div>
      <div class="metric"><span>A11y</span><strong>{{ .AccessibilityFindings }}</strong></div>
      <div class="metric"><span>Performance</span><strong>{{ .PerformanceFindings }}</strong></div>
      <div class="metric"><span>High+</span><strong>{{ add .Critical .High }}</strong></div>
    </div>
  </section>
  {{ end }}

  {{ if .Report.QualityResults }}
  <section style="margin-bottom: 16px;">
    <h2>Quality Findings</h2>
    <table>
      <thead>
        <tr><th>Severity</th><th>Category</th><th>Rule</th><th>Title</th><th>URL</th><th>Recommendation</th></tr>
      </thead>
      <tbody>
        {{ range .Report.QualityResults }}
        <tr>
          <td class="severity-{{ .Severity }}">{{ .Severity }}</td>
          <td>{{ .Category }}</td>
          <td>{{ .RuleID }}</td>
          <td>{{ .Title }}</td>
          <td><code>{{ .URL }}</code></td>
          <td>{{ .Recommendation }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
  </section>
  {{ end }}

  {{ with .Report.TestPlan }}
  <section style="margin-bottom: 16px;">
    <h2>AI Test Plan</h2>
    <p><strong>{{ .Title }}</strong></p>
    <p>{{ .Summary }}</p>
    <p><strong>Status:</strong> <span class="status">{{ .Status }}</span> <strong>Risk:</strong> <span class="severity-{{ .RiskLevel }}">{{ .RiskLevel }}</span> <strong>Scenarios:</strong> {{ .TotalScenarios }}</p>
    <pre>{{ json .ExecutionCoverage }}</pre>
  </section>
  {{ end }}

  {{ with .Report.ExecutionPreview }}
  <section style="margin-bottom: 16px;">
    <h2>Safe Execution Preview</h2>
    <div class="grid six">
      <div class="metric"><span>Scenarios</span><strong>{{ .TotalScenarios }}</strong></div>
      <div class="metric"><span>Executable</span><strong>{{ .ExecutableScenarios }}</strong></div>
      <div class="metric"><span>Skipped</span><strong>{{ .SkippedScenarios }}</strong></div>
      <div class="metric"><span>Steps</span><strong>{{ .TotalSteps }}</strong></div>
      <div class="metric"><span>Safe Steps</span><strong>{{ .ExecutableSteps }}</strong></div>
      <div class="metric"><span>Skipped Steps</span><strong>{{ .SkippedSteps }}</strong></div>
    </div>
  </section>
  {{ end }}

  {{ if .Report.APISmokeRun }}
  <section style="margin-bottom: 16px;">
    <h2>Authenticated API Smoke & Contract Validation</h2>
    {{ with .Report.APISpec }}
    <p><strong>Spec:</strong> {{ .Name }} {{ if .ParsedVersion }}<span class="subtle">({{ .ParsedVersion }})</span>{{ end }}</p>
    {{ end }}
    {{ with .Report.APIAuth }}
    <p><strong>Auth mode:</strong> <span class="status">{{ .AuthMode }}</span> {{ if .ProfileName }}<strong>Profile:</strong> {{ .ProfileName }}{{ end }}</p>
    {{ end }}
    {{ with .Report.APISummary }}
    <div class="grid six">
      <div class="metric"><span>Executed</span><strong>{{ .ExecutedOperations }}</strong></div>
      <div class="metric"><span>Authenticated</span><strong>{{ .AuthenticatedOperations }}</strong></div>
      <div class="metric"><span>Contract Passed</span><strong>{{ .ContractPassed }}</strong></div>
      <div class="metric"><span>Contract Failed</span><strong>{{ .ContractFailed }}</strong></div>
      <div class="metric"><span>Schema Errors</span><strong>{{ .SchemaValidationErrorCount }}</strong></div>
      <div class="metric"><span>Skipped</span><strong>{{ .SkippedOperations }}</strong></div>
    </div>
    {{ end }}
    <p class="subtle">API auth secrets, auth headers, request bodies, and response bodies are not stored or included in this report.</p>
    {{ if .Report.APIResults }}
    <table>
      <thead><tr><th>Status</th><th>Method</th><th>Path</th><th>HTTP</th><th>Contract</th><th>Reason/Error</th></tr></thead>
      <tbody>
      {{ range .Report.APIResults }}
        <tr>
          <td><span class="status">{{ .Status }}</span></td>
          <td><code>{{ .Method }}</code></td>
          <td><code>{{ .Path }}</code></td>
          <td>{{ if .HTTPStatus }}{{ intValue .HTTPStatus }}{{ else }}<span class="subtle">n/a</span>{{ end }}</td>
          <td>{{ .ContractValidationStatus }}</td>
          <td>{{ if .SkippedReason }}{{ .SkippedReason }}{{ else }}{{ .ErrorMessage }}{{ end }}</td>
        </tr>
      {{ end }}
      </tbody>
    </table>
    {{ end }}
  </section>
  {{ end }}

  {{ with .Report.ExecutionReport }}
  <section style="margin-bottom: 16px;">
    <h2>Execution Result</h2>
    <div class="grid six">
      <div class="metric"><span>Status</span><strong>{{ .Execution.Status }}</strong></div>
      <div class="metric"><span>Total</span><strong>{{ .Execution.TotalScenarios }}</strong></div>
      <div class="metric"><span>Passed</span><strong>{{ .Execution.PassedScenarios }}</strong></div>
      <div class="metric"><span>Failed</span><strong>{{ .Execution.FailedScenarios }}</strong></div>
      <div class="metric"><span>Skipped</span><strong>{{ .Execution.SkippedScenarios }}</strong></div>
      <div class="metric"><span>Steps</span><strong>{{ .Execution.TotalSteps }}</strong></div>
    </div>
  </section>
  {{ end }}

  <section style="margin-bottom: 16px;">
    <h2>Findings</h2>
    {{ if .Report.Findings }}
    <table>
      <thead>
        <tr><th>Severity</th><th>Title</th><th>Category</th><th>Description</th><th>Recommendation</th></tr>
      </thead>
      <tbody>
        {{ range .Report.Findings }}
        <tr>
          <td class="severity-{{ .Severity }}">{{ .Severity }}</td>
          <td>{{ .Title }}</td>
          <td>{{ .Category }}</td>
          <td>{{ .Description }}</td>
          <td>{{ .Recommendation }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}
    <p class="subtle">No findings were recorded for this safe QA run.</p>
    {{ end }}
  </section>

  <section style="margin-bottom: 16px;">
    <h2>Evidence</h2>
    {{ if .Report.Evidence }}
    <table>
      <thead>
        <tr><th>Type</th><th>URI</th><th>Metadata</th></tr>
      </thead>
      <tbody>
        {{ range .Report.Evidence }}
        <tr>
          <td>{{ .Type }}</td>
          <td><code>{{ .URI }}</code></td>
          <td><pre>{{ json .Metadata }}</pre></td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}
    <p class="subtle">No evidence was recorded for this safe QA run.</p>
    {{ end }}
  </section>

  <section>
    <h2>Safety Scope</h2>
    <h3>Notes</h3>
    <ul>{{ range .Report.SafetyNotes }}<li>{{ . }}</li>{{ end }}</ul>
    <h3>Limitations</h3>
    <ul>{{ range .Report.Limitations }}<li>{{ . }}</li>{{ end }}</ul>
  </section>
</main>
</body>
</html>`))

func RenderQARunHTMLReport(w io.Writer, report *QARunReport) error {
	return qaRunHTMLReportTemplate.Execute(w, qaRunHTMLReportData{
		Report:  report,
		Summary: summarizeFindings(report.Findings),
	})
}

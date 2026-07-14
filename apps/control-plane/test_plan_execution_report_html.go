package main

import (
	"html/template"
	"io"
)

type testPlanExecutionHTMLReportData struct {
	Report  *TestPlanExecutionReport
	Summary ReportSummary
}

var testPlanExecutionHTMLReportTemplate = template.Must(template.New("test-plan-execution-report").Funcs(template.FuncMap{
	"json":       prettyJSON,
	"formatTime": formatReportTime,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora test plan execution - {{ .Report.Project.Name }}</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f6f7f9;
      --panel: #ffffff;
      --text: #18202a;
      --muted: #5c6675;
      --line: #d8dee8;
      --strong: #0d5b57;
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
    .severity-critical, .severity-high { color: var(--high); font-weight: 700; }
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
    <p class="subtle">Qualora safe test plan execution report</p>
    <h1>{{ .Report.TestPlan.Title }}</h1>
    <p><span class="status">{{ .Report.Execution.Status }}</span> <span class="subtle">{{ .Report.Project.Name }} generated {{ formatTime .Report.GeneratedAt }}</span></p>
  </header>

  <section>
    <h2>Execution Summary</h2>
    <div class="grid two">
      <div>
        <p><strong>Execution ID:</strong> <code>{{ .Report.Execution.ID }}</code></p>
        <p><strong>Test plan ID:</strong> <code>{{ .Report.TestPlan.ID }}</code></p>
        <p><strong>Project ID:</strong> <code>{{ .Report.Project.ID }}</code></p>
      </div>
      <div>
        <p><strong>Created:</strong> {{ formatTime .Report.Execution.CreatedAt }}</p>
        <p><strong>Started:</strong> {{ formatTime .Report.Execution.StartedAt }}</p>
        <p><strong>Completed:</strong> {{ formatTime .Report.Execution.CompletedAt }}</p>
      </div>
    </div>
  </section>

  <div class="grid six" style="margin: 16px 0;">
    <div class="metric"><span>Scenarios</span><strong>{{ .Report.Execution.TotalScenarios }}</strong></div>
    <div class="metric"><span>Passed</span><strong>{{ .Report.Execution.PassedScenarios }}</strong></div>
    <div class="metric"><span>Failed</span><strong>{{ .Report.Execution.FailedScenarios }}</strong></div>
    <div class="metric"><span>Skipped</span><strong>{{ .Report.Execution.SkippedScenarios }}</strong></div>
    <div class="metric"><span>Steps</span><strong>{{ .Report.Execution.TotalSteps }}</strong></div>
    <div class="metric"><span>Findings</span><strong>{{ .Summary.TotalFindings }}</strong></div>
  </div>

  <section style="margin-bottom: 16px;">
    <h2>Safety Scope</h2>
    <div class="grid two">
      <div>
        <p><strong>Executed steps:</strong> {{ .Report.SafetySummary.ExecutedSteps }}</p>
        <p><strong>Skipped unsafe steps:</strong> {{ .Report.SafetySummary.SkippedUnsafeSteps }}</p>
      </div>
      <div>
        <p><strong>Skipped unsupported steps:</strong> {{ .Report.SafetySummary.SkippedUnsupportedSteps }}</p>
        <p><strong>Skipped scenarios:</strong> {{ .Report.SafetySummary.SkippedScenarios }}</p>
      </div>
    </div>
    <p class="subtle">Qualora executes only approved, safe, non-destructive browser actions from the supported DSL. Unsupported, ambiguous, authenticated, or destructive steps are skipped.</p>
  </section>

  <section>
    <h2>Scenarios and Steps</h2>
    {{ if .Report.Scenarios }}
      {{ range .Report.Scenarios }}
      <h3>{{ .Name }} <span class="status">{{ .Status }}</span></h3>
      {{ if .SkipReason }}<p class="subtle">{{ .SkipReason }}</p>{{ end }}
      <table style="margin-bottom: 16px;">
        <thead>
          <tr><th>#</th><th>Action</th><th>Target</th><th>Status</th><th>Result</th></tr>
        </thead>
        <tbody>
          {{ range .Steps }}
          <tr>
            <td>{{ .StepOrder }}</td>
            <td><code>{{ .MappedAction }}</code></td>
            <td><code>{{ .Target }}</code></td>
            <td>{{ .Status }}</td>
            <td>{{ if .ActualResult }}{{ .ActualResult }}{{ else if .SkipReason }}{{ .SkipReason }}{{ else if .ErrorMessage }}{{ .ErrorMessage }}{{ else }}<span class="subtle">Not recorded</span>{{ end }}</td>
          </tr>
          {{ end }}
        </tbody>
      </table>
      {{ end }}
    {{ else }}
    <p class="subtle">No execution scenarios were recorded.</p>
    {{ end }}
  </section>

  <section style="margin-top: 16px;">
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
    <p class="subtle">No findings were recorded for this execution.</p>
    {{ end }}
  </section>

  <section style="margin-top: 16px;">
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
    <p class="subtle">No evidence was recorded for this execution.</p>
    {{ end }}
  </section>
</main>
</body>
</html>`))

func RenderTestPlanExecutionHTMLReport(w io.Writer, report *TestPlanExecutionReport) error {
	return testPlanExecutionHTMLReportTemplate.Execute(w, testPlanExecutionHTMLReportData{
		Report:  report,
		Summary: summarizeFindings(report.Findings),
	})
}

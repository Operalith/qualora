package main

import (
	"html/template"
	"io"
	"time"
)

type aiBrowserControlHTMLReportData struct {
	Report      *AIBrowserControlReport
	GeneratedAt string
}

var aiBrowserControlHTMLReportTemplate = template.Must(template.New("ai-browser-control-report").Funcs(template.FuncMap{
	"formatTime":         formatReportTime,
	"intValue":           optionalIntValue,
	"reportIntelligence": reportIntelligenceHTML,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora AI Browser Control report - {{ .Report.Project.Name }}</title>
  <style>
    :root { color-scheme: light; --bg:#f7f8fb; --panel:#fff; --text:#192232; --muted:#647084; --line:#d8dee8; --ok:#067647; --warn:#b54708; --danger:#b42318; --info:#344054; --brand:#0b615e; }
    * { box-sizing: border-box; }
    body { margin:0; background:var(--bg); color:var(--text); font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; font-size:14px; line-height:1.5; }
    main { max-width:1200px; margin:0 auto; padding:32px 20px 48px; }
    h1,h2,h3 { margin:0; line-height:1.2; }
    h1 { font-size:28px; }
    h2 { font-size:18px; margin-bottom:12px; }
    section,.metric { background:var(--panel); border:1px solid var(--line); border-radius:8px; padding:16px; }
    section { margin-top:16px; }
    .subtle { color:var(--muted); }
    .status { display:inline-block; border:1px solid var(--line); border-radius:999px; padding:2px 8px; background:#eef8f7; color:var(--brand); font-weight:700; }
    .grid { display:grid; gap:16px; }
    .grid.two { grid-template-columns:repeat(2,minmax(0,1fr)); }
    .grid.seven { grid-template-columns:repeat(7,minmax(0,1fr)); }
    .metric span { display:block; color:var(--muted); font-size:12px; text-transform:uppercase; }
    .metric strong { display:block; margin-top:4px; font-size:22px; }
    table { width:100%; border-collapse:collapse; }
    th,td { border-bottom:1px solid var(--line); padding:10px 8px; text-align:left; vertical-align:top; }
    th { color:var(--muted); font-size:12px; text-transform:uppercase; }
    tr:last-child td { border-bottom:0; }
    code,pre { font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono",monospace; font-size:12px; word-break:break-word; white-space:pre-wrap; }
    .approved,.executed { color:var(--ok); font-weight:700; }
    .blocked,.failed,.error { color:var(--danger); font-weight:700; }
    .unsupported,.invalid,.skipped { color:var(--warn); font-weight:700; }
    .severity-high,.severity-critical { color:var(--danger); font-weight:700; }
    .severity-medium { color:var(--warn); font-weight:700; }
    .severity-low { color:#175cd3; font-weight:700; }
    .severity-info { color:var(--info); font-weight:700; }
    @media (max-width: 900px) { main{padding:20px 12px 32px;} .grid.two,.grid.seven{grid-template-columns:1fr;} table{display:block; overflow-x:auto;} }
  </style>
</head>
<body>
<main>
  <header>
    <p class="subtle">Qualora Policy-Gated AI Browser Control report</p>
    <h1>{{ .Report.Project.Name }}</h1>
    <p><span class="status">{{ .Report.Run.Status }}</span> <span class="subtle">Run {{ .Report.Run.ID }} generated {{ .GeneratedAt }}</span></p>
  </header>

  <section>
    <h2>Run Summary</h2>
    <div class="grid two">
      <div>
        <p><strong>Goal:</strong> {{ if .Report.Run.Goal }}{{ .Report.Run.Goal }}{{ else }}<span class="subtle">No goal supplied</span>{{ end }}</p>
        <p><strong>Start URL:</strong> <code>{{ .Report.Run.StartURL }}</code></p>
        <p><strong>Provider:</strong> {{ .Report.Run.ProviderName }} <code>{{ .Report.Run.ProviderID }}</code></p>
      </div>
      <div>
        <p><strong>Created:</strong> {{ formatTime .Report.Run.CreatedAt }}</p>
        <p><strong>Started:</strong> {{ formatTime .Report.Run.StartedAt }}</p>
        <p><strong>Completed:</strong> {{ formatTime .Report.Run.CompletedAt }}</p>
      </div>
    </div>
    {{ if .Report.Run.ErrorMessage }}<p><strong>Error:</strong> {{ .Report.Run.ErrorMessage }}</p>{{ end }}
  </section>

  <div class="grid seven" style="margin-top:16px;">
    <div class="metric"><span>Steps</span><strong>{{ .Report.Summary.TotalSteps }}</strong></div>
    <div class="metric"><span>AI Suggestions</span><strong>{{ .Report.Summary.TotalAISuggestions }}</strong></div>
    <div class="metric"><span>Approved</span><strong>{{ .Report.Summary.ActionsApproved }}</strong></div>
    <div class="metric"><span>Executed</span><strong>{{ .Report.Summary.ActionsExecuted }}</strong></div>
    <div class="metric"><span>Skipped</span><strong>{{ .Report.Summary.ActionsSkipped }}</strong></div>
    <div class="metric"><span>Policy Blocks</span><strong>{{ .Report.Summary.PolicyBlocks }}</strong></div>
    <div class="metric"><span>Findings</span><strong>{{ .Report.Summary.Findings }}</strong></div>
  </div>

  {{ reportIntelligence .Report.ReportIntelligence }}

  <section>
    <h2>Policy Model</h2>
    <p>AI proposes exactly one typed action from sanitized observations. Qualora validates that action through deterministic policy checks before the browser worker executes anything.</p>
    <ul>
      {{ range .Report.SafetyNotes }}<li>{{ . }}</li>{{ end }}
    </ul>
  </section>

  <section>
    <h2>Step Timeline</h2>
    {{ if .Report.Steps }}
    <table>
      <thead><tr><th>#</th><th>Page</th><th>AI Suggestion</th><th>Policy Decision</th><th>Execution</th><th>Reason</th><th>Evidence</th></tr></thead>
      <tbody>
        {{ range .Report.Steps }}
        <tr>
          <td>{{ .StepIndex }}</td>
          <td><strong>{{ .PageTitle }}</strong><br><code>{{ .NormalizedURL }}</code></td>
          <td><code>{{ .ActionType }}</code><br>{{ .ActionLabel }}<br><code>{{ .ActionTargetURL }}</code></td>
          <td><span class="{{ .PolicyDecision }}">{{ .PolicyDecision }}</span></td>
          <td><span class="{{ .ExecutionStatus }}">{{ .ExecutionStatus }}</span></td>
          <td>{{ if .PolicyReason }}{{ .PolicyReason }}{{ else }}<span class="subtle">None</span>{{ end }}</td>
          <td>{{ if .ScreenshotEvidenceID }}<code>{{ .ScreenshotEvidenceID }}</code>{{ else }}<span class="subtle">None</span>{{ end }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No steps recorded yet.</p>{{ end }}
  </section>

  <section>
    <h2>Findings</h2>
    {{ if .Report.Findings }}
    <table>
      <thead><tr><th>Severity</th><th>Category</th><th>Title</th><th>Description</th><th>Recommendation</th></tr></thead>
      <tbody>
        {{ range .Report.Findings }}
        <tr>
          <td class="severity-{{ .Severity }}">{{ .Severity }}</td>
          <td>{{ .Category }}</td>
          <td>{{ .Title }}</td>
          <td>{{ .Description }}</td>
          <td>{{ .Recommendation }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No findings were recorded for this run.</p>{{ end }}
  </section>

  <section>
    <h2>Evidence</h2>
    {{ if .Report.Evidence }}
    <table>
      <thead><tr><th>Type</th><th>URI</th><th>Metadata</th></tr></thead>
      <tbody>
        {{ range .Report.Evidence }}
        <tr>
          <td>{{ .Type }}</td>
          <td><code>{{ .URI }}</code></td>
          <td><pre>{{ printf "%#v" .Metadata }}</pre></td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No evidence recorded yet.</p>{{ end }}
  </section>

  <section>
    <h2>Limitations</h2>
    <ul>
      {{ range .Report.Limitations }}<li>{{ . }}</li>{{ end }}
    </ul>
  </section>
</main>
</body>
</html>`))

func RenderAIBrowserControlHTMLReport(w io.Writer, report *AIBrowserControlReport, generatedAt time.Time) error {
	return aiBrowserControlHTMLReportTemplate.Execute(w, aiBrowserControlHTMLReportData{
		Report:      report,
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
	})
}

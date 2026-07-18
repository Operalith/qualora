package main

import (
	"html/template"
	"io"
	"time"
)

type safeExplorerHTMLReportData struct {
	Report      *SafeExplorerReport
	GeneratedAt string
}

var safeExplorerHTMLReportTemplate = template.Must(template.New("safe-explorer-report").Funcs(template.FuncMap{
	"formatTime":         formatReportTime,
	"intValue":           optionalIntValue,
	"reportIntelligence": reportIntelligenceHTML,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora Safe Explorer report - {{ .Report.Project.Name }}</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f7f8fb;
      --panel: #ffffff;
      --text: #192232;
      --muted: #647084;
      --line: #d8dee8;
      --strong: #0b615e;
      --danger: #b42318;
      --warn: #b54708;
      --ok: #067647;
      --info: #344054;
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
    main { max-width: 1200px; margin: 0 auto; padding: 32px 20px 48px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 28px; }
    h2 { font-size: 18px; margin-bottom: 12px; }
    h3 { font-size: 15px; margin: 14px 0 8px; }
    section, .metric {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 16px;
    }
    section { margin-top: 16px; }
    .subtle { color: var(--muted); }
    .status {
      display: inline-block;
      border: 1px solid var(--line);
      border-radius: 999px;
      padding: 2px 8px;
      background: #eef8f7;
      color: var(--strong);
      font-weight: 700;
    }
    .grid { display: grid; gap: 16px; }
    .grid.two { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .grid.six { grid-template-columns: repeat(6, minmax(0, 1fr)); }
    .metric span { display: block; color: var(--muted); font-size: 12px; text-transform: uppercase; }
    .metric strong { display: block; margin-top: 4px; font-size: 22px; }
    table { width: 100%; border-collapse: collapse; }
    th, td { border-bottom: 1px solid var(--line); padding: 10px 8px; text-align: left; vertical-align: top; }
    th { color: var(--muted); font-size: 12px; text-transform: uppercase; }
    tr:last-child td { border-bottom: 0; }
    code {
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
      font-size: 12px;
      word-break: break-word;
    }
    ul { margin: 8px 0 0 18px; padding: 0; }
    .safe { color: var(--ok); font-weight: 700; }
    .unsafe { color: var(--danger); font-weight: 700; }
    .unsupported, .unknown { color: var(--warn); font-weight: 700; }
    .severity-high, .severity-critical { color: var(--danger); font-weight: 700; }
    .severity-medium { color: var(--warn); font-weight: 700; }
    .severity-low { color: #175cd3; font-weight: 700; }
    .severity-info { color: var(--info); font-weight: 700; }
    @media (max-width: 860px) {
      main { padding: 20px 12px 32px; }
      .grid.two, .grid.six { grid-template-columns: 1fr; }
      table { display: block; overflow-x: auto; }
    }
  </style>
</head>
<body>
<main>
  <header>
    <p class="subtle">Qualora Interactive Safe Explorer report</p>
    <h1>{{ .Report.Project.Name }}</h1>
    <p><span class="status">{{ .Report.Run.Status }}</span> <span class="subtle">Run {{ .Report.Run.ID }} generated {{ .GeneratedAt }}</span></p>
  </header>

  <section>
    <h2>Run Summary</h2>
    <div class="grid two">
      <div>
        <p><strong>Start URL:</strong> <code>{{ .Report.Run.StartURL }}</code></p>
        <p><strong>Project ID:</strong> <code>{{ .Report.Project.ID }}</code></p>
        <p><strong>Created:</strong> {{ formatTime .Report.Run.CreatedAt }}</p>
      </div>
      <div>
        <p><strong>Started:</strong> {{ formatTime .Report.Run.StartedAt }}</p>
        <p><strong>Completed:</strong> {{ formatTime .Report.Run.CompletedAt }}</p>
        <p><strong>Error:</strong> {{ if .Report.Run.ErrorMessage }}{{ .Report.Run.ErrorMessage }}{{ else }}<span class="subtle">None recorded</span>{{ end }}</p>
      </div>
    </div>
  </section>

  <div class="grid six" style="margin-top: 16px;">
    <div class="metric"><span>Steps</span><strong>{{ .Report.Summary.TotalSteps }}</strong></div>
    <div class="metric"><span>Pages</span><strong>{{ .Report.Summary.TotalPagesObserved }}</strong></div>
    <div class="metric"><span>Detected</span><strong>{{ .Report.Summary.TotalActionsDetected }}</strong></div>
    <div class="metric"><span>Executed</span><strong>{{ .Report.Summary.TotalActionsExecuted }}</strong></div>
    <div class="metric"><span>Skipped</span><strong>{{ .Report.Summary.TotalActionsSkipped }}</strong></div>
    <div class="metric"><span>Findings</span><strong>{{ .Report.Summary.TotalFindings }}</strong></div>
  </div>

  {{ reportIntelligence .Report.ReportIntelligence }}

  <section>
    <h2>Settings</h2>
    <div class="grid two">
      <div>
        <p><strong>Max steps:</strong> {{ .Report.Run.MaxSteps }}</p>
        <p><strong>Max depth:</strong> {{ .Report.Run.MaxDepth }}</p>
        <p><strong>Same origin only:</strong> {{ .Report.Run.SameOriginOnly }}</p>
      </div>
      <div>
        <p><strong>Allow GET forms:</strong> {{ .Report.Run.AllowGetForms }}</p>
        <p><strong>Forms submitted:</strong> false</p>
        <p><strong>AI action selection:</strong> false</p>
      </div>
    </div>
  </section>

  <section>
    <h2>Timeline</h2>
    {{ if .Report.Steps }}
    <table>
      <thead><tr><th>#</th><th>Decision</th><th>Depth</th><th>Status</th><th>Title / Action</th><th>URL</th><th>Evidence</th></tr></thead>
      <tbody>
        {{ range .Report.Steps }}
        <tr>
          <td>{{ .StepIndex }}</td>
          <td>{{ .ActionDecision }}</td>
          <td>{{ .Depth }}</td>
          <td>{{ .ResultStatus }} {{ intValue .HTTPStatus }}</td>
          <td>{{ if .ActionLabel }}{{ .ActionLabel }}{{ else }}{{ .PageTitle }}{{ end }}</td>
          <td><code>{{ if .ActionTargetURL }}{{ .ActionTargetURL }}{{ else }}{{ .NormalizedURL }}{{ end }}</code></td>
          <td>{{ if .ScreenshotEvidenceID }}<code>{{ .ScreenshotEvidenceID }}</code>{{ else }}<span class="subtle">None</span>{{ end }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No steps recorded yet.</p>{{ end }}
  </section>

  <section>
    <h2>Actions</h2>
    {{ if .Report.Actions }}
    <table>
      <thead><tr><th>Decision</th><th>Safety</th><th>Type</th><th>Label</th><th>Target</th><th>Skip reason</th></tr></thead>
      <tbody>
        {{ range .Report.Actions }}
        <tr>
          <td>{{ .Decision }}</td>
          <td><span class="{{ .Safety }}">{{ .Safety }}</span></td>
          <td>{{ .ActionType }}</td>
          <td>{{ .Label }}</td>
          <td><code>{{ .TargetURL }}</code></td>
          <td>{{ if .SkipReason }}{{ .SkipReason }}{{ else }}<span class="subtle">None</span>{{ end }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No actions recorded yet.</p>{{ end }}
  </section>

  <section>
    <h2>Findings</h2>
    {{ if .Report.Findings }}
    <table>
      <thead><tr><th>Severity</th><th>Title</th><th>Category</th><th>Description</th><th>Recommendation</th></tr></thead>
      <tbody>
        {{ range .Report.Findings }}
        <tr>
          <td><span class="severity-{{ .Severity }}">{{ .Severity }}</span></td>
          <td>{{ .Title }}</td>
          <td>{{ .Category }}</td>
          <td>{{ .Description }}</td>
          <td>{{ .Recommendation }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No findings recorded.</p>{{ end }}
  </section>

  <section>
    <h2>Evidence Metadata</h2>
    {{ if .Report.Evidence }}
    <table>
      <thead><tr><th>Type</th><th>URI</th><th>ID</th></tr></thead>
      <tbody>
        {{ range .Report.Evidence }}
        <tr>
          <td>{{ .Type }}</td>
          <td><code>{{ .URI }}</code></td>
          <td><code>{{ .ID }}</code></td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No evidence recorded yet.</p>{{ end }}
  </section>

  <section>
    <h2>Safety Notes</h2>
    <ul>{{ range .Report.SafetyNotes }}<li>{{ . }}</li>{{ end }}</ul>
    <h3>Known Limitations</h3>
    <ul>{{ range .Report.Limitations }}<li>{{ . }}</li>{{ end }}</ul>
  </section>
</main>
</body>
</html>`))

func RenderSafeExplorerHTMLReport(w io.Writer, report *SafeExplorerReport, generatedAt time.Time) error {
	return safeExplorerHTMLReportTemplate.Execute(w, safeExplorerHTMLReportData{
		Report:      report,
		GeneratedAt: generatedAt.Format(time.RFC3339),
	})
}

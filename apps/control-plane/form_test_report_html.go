package main

import (
	"html/template"
	"io"
	"time"
)

type formTestHTMLReportData struct {
	Report      *FormTestReport
	GeneratedAt string
}

var formTestHTMLReportTemplate = template.Must(template.New("form-test-report").Funcs(template.FuncMap{
	"json":               prettyJSON,
	"formatTime":         formatReportTime,
	"reportIntelligence": reportIntelligenceHTML,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora safe form report - {{ .Report.Project.Name }}</title>
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
    code, pre {
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
      font-size: 12px;
      word-break: break-word;
    }
    pre {
      margin: 0;
      white-space: pre-wrap;
      background: #f1f4f8;
      border: 1px solid var(--line);
      border-radius: 6px;
      padding: 10px;
      max-height: 260px;
      overflow: auto;
    }
    ul { margin: 8px 0 0 18px; padding: 0; }
    .severity-critical { color: var(--critical); font-weight: 700; }
    .severity-high { color: var(--high); font-weight: 700; }
    .severity-medium { color: var(--medium); font-weight: 700; }
    .severity-low { color: var(--low); font-weight: 700; }
    .severity-info { color: var(--info); font-weight: 700; }
    @media (max-width: 820px) {
      main { padding: 20px 12px 32px; }
      .grid.two, .grid.six { grid-template-columns: 1fr; }
      table { display: block; overflow-x: auto; }
    }
  </style>
</head>
<body>
<main>
  <header>
    <p class="subtle">Qualora safe form report</p>
    <h1>{{ .Report.Project.Name }}</h1>
    <p><span class="status">{{ .Report.Run.Status }}</span> <span class="subtle">Form test run {{ .Report.Run.ID }} generated {{ .GeneratedAt }}</span></p>
  </header>

  <section>
    <h2>Run Summary</h2>
    <div class="grid two">
      <div>
        <p><strong>Project ID:</strong> <code>{{ .Report.Project.ID }}</code></p>
        <p><strong>Run ID:</strong> <code>{{ .Report.Run.ID }}</code></p>
        <p><strong>Target URL:</strong> <code>{{ .Report.Run.TargetURL }}</code></p>
        {{ if .Report.Run.DiscoveryRunID }}<p><strong>Discovery run:</strong> <code>{{ .Report.Run.DiscoveryRunID }}</code></p>{{ end }}
      </div>
      <div>
        <p><strong>Created:</strong> {{ formatTime .Report.Run.CreatedAt }}</p>
        <p><strong>Started:</strong> {{ formatTime .Report.Run.StartedAt }}</p>
        <p><strong>Completed:</strong> {{ formatTime .Report.Run.CompletedAt }}</p>
        <p><strong>Error:</strong> {{ if .Report.Run.ErrorMessage }}{{ .Report.Run.ErrorMessage }}{{ else }}<span class="subtle">None recorded</span>{{ end }}</p>
      </div>
    </div>
  </section>

  <div class="grid six" style="margin-top: 16px;">
    <div class="metric"><span>Detected</span><strong>{{ .Report.Summary.FormsDetected }}</strong></div>
    <div class="metric"><span>Safe</span><strong>{{ .Report.Summary.FormsClassifiedSafe }}</strong></div>
    <div class="metric"><span>Tested</span><strong>{{ .Report.Summary.FormsTested }}</strong></div>
    <div class="metric"><span>Skipped</span><strong>{{ .Report.Summary.FormsSkipped }}</strong></div>
    <div class="metric"><span>Findings</span><strong>{{ .Report.Summary.Findings }}</strong></div>
    <div class="metric"><span>Screenshots</span><strong>{{ .Report.Summary.Screenshots }}</strong></div>
  </div>

  {{ reportIntelligence .Report.ReportIntelligence }}

  <section>
    <h2>Form Results</h2>
    {{ if .Report.Results }}
    <table>
      <thead><tr><th>Decision</th><th>Safety</th><th>Method</th><th>Classification</th><th>Page</th><th>Action</th><th>Status</th><th>Skip Reason</th></tr></thead>
      <tbody>
        {{ range .Report.Results }}
        <tr>
          <td>{{ .Decision }}</td>
          <td>{{ .Safety }}</td>
          <td>{{ .FormMethod }}</td>
          <td>{{ .Classification }}</td>
          <td><code>{{ .PageURL }}</code></td>
          <td><code>{{ .FormAction }}</code></td>
          <td>{{ if .HTTPStatus }}{{ .HTTPStatus }}{{ else }}<span class="subtle">n/a</span>{{ end }}</td>
          <td>{{ if .SkipReason }}{{ .SkipReason }}{{ else }}<span class="subtle">None</span>{{ end }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No form results were recorded.</p>{{ end }}
  </section>

  <section>
    <h2>Findings</h2>
    {{ if .Report.Findings }}
    <table>
      <thead><tr><th>Severity</th><th>Category</th><th>Title</th><th>Description</th><th>Recommendation</th></tr></thead>
      <tbody>
        {{ range .Report.Findings }}
        <tr>
          <td><span class="severity-{{ .Severity }}">{{ .Severity }}</span></td>
          <td>{{ .Category }}</td>
          <td>{{ .Title }}</td>
          <td>{{ .Description }}</td>
          <td>{{ .Recommendation }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No findings were recorded.</p>{{ end }}
  </section>

  <section>
    <h2>Evidence Metadata</h2>
    {{ if .Report.Evidence }}
    <table>
      <thead><tr><th>Type</th><th>URI</th><th>Metadata</th></tr></thead>
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
    {{ else }}<p class="subtle">No evidence metadata was recorded.</p>{{ end }}
  </section>

  <section>
    <h2>Safety Scope</h2>
    <h3>Notes</h3>
    <ul>{{ range .Report.SafetyNotes }}<li>{{ . }}</li>{{ end }}</ul>
    <h3>Known Limitations</h3>
    <ul>{{ range .Report.Limitations }}<li>{{ . }}</li>{{ end }}</ul>
  </section>
</main>
</body>
</html>`))

func RenderFormTestHTMLReport(w io.Writer, report *FormTestReport, generatedAt time.Time) error {
	return formTestHTMLReportTemplate.Execute(w, formTestHTMLReportData{
		Report:      report,
		GeneratedAt: generatedAt.Format(time.RFC3339),
	})
}

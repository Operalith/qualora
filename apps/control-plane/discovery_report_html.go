package main

import (
	"html/template"
	"io"
	"time"
)

type discoveryHTMLReportData struct {
	Report      *DiscoveryReport
	GeneratedAt string
}

var discoveryHTMLReportTemplate = template.Must(template.New("discovery-report").Funcs(template.FuncMap{
	"formatTime": formatReportTime,
	"intValue":   optionalIntValue,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora discovery report - {{ .Report.Project.Name }}</title>
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
    main { max-width: 1180px; margin: 0 auto; padding: 32px 20px 48px; }
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
    <p class="subtle">Qualora application discovery report</p>
    <h1>{{ .Report.Project.Name }}</h1>
    <p><span class="status">{{ .Report.Run.Status }}</span> <span class="subtle">Discovery run {{ .Report.Run.ID }} generated {{ .GeneratedAt }}</span></p>
  </header>

  <section>
    <h2>Executive Summary</h2>
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
    <div class="metric"><span>Pages</span><strong>{{ .Report.Summary.TotalPages }}</strong></div>
    <div class="metric"><span>Links</span><strong>{{ .Report.Summary.TotalLinks }}</strong></div>
    <div class="metric"><span>Forms</span><strong>{{ .Report.Summary.TotalForms }}</strong></div>
    <div class="metric"><span>Findings</span><strong>{{ .Report.Summary.TotalFindings }}</strong></div>
    <div class="metric"><span>Console Errors</span><strong>{{ .Report.Summary.TotalConsoleErrors }}</strong></div>
    <div class="metric"><span>Failed Requests</span><strong>{{ .Report.Summary.TotalFailedRequests }}</strong></div>
  </div>

  <section>
    <h2>Discovery Settings</h2>
    <div class="grid two">
      <div>
        <p><strong>Max pages:</strong> {{ .Report.Run.MaxPages }}</p>
        <p><strong>Max depth:</strong> {{ .Report.Run.MaxDepth }}</p>
        <p><strong>Same origin only:</strong> {{ .Report.Run.SameOriginOnly }}</p>
      </div>
      <div>
        <p><strong>Forms submitted:</strong> false</p>
        <p><strong>Destructive actions:</strong> false</p>
        <p><strong>Autonomous AI browser control:</strong> false</p>
      </div>
    </div>
  </section>

  <section>
    <h2>Discovered Pages</h2>
    {{ if .Report.Pages }}
    <table>
      <thead><tr><th>Depth</th><th>Status</th><th>Title</th><th>URL</th><th>Signals</th><th>Screenshot Evidence</th></tr></thead>
      <tbody>
        {{ range .Report.Pages }}
        <tr>
          <td>{{ .Depth }}</td>
          <td>{{ intValue .HTTPStatus }}</td>
          <td>{{ .Title }}</td>
          <td><code>{{ .NormalizedURL }}</code></td>
          <td>{{ .ConsoleErrorCount }} console errors, {{ .FailedRequestCount }} failed requests</td>
          <td>{{ if .ScreenshotEvidenceID }}<code>{{ .ScreenshotEvidenceID }}</code>{{ else }}<span class="subtle">None</span>{{ end }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No pages recorded yet.</p>{{ end }}
  </section>

  <section>
    <h2>Forms</h2>
    {{ if .Report.Forms }}
    <table>
      <thead><tr><th>Method</th><th>Action</th><th>Fields</th><th>Password Fields</th><th>Classification</th></tr></thead>
      <tbody>
        {{ range .Report.Forms }}
        <tr>
          <td>{{ .FormMethod }}</td>
          <td><code>{{ .FormAction }}</code></td>
          <td>{{ .FieldCount }}</td>
          <td>{{ .PasswordFieldCount }}</td>
          <td>{{ .Classification }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No forms recorded.</p>{{ end }}
  </section>

  <section>
    <h2>Skipped Links</h2>
    {{ if .Report.Links }}
    <table>
      <thead><tr><th>Reason</th><th>Text</th><th>Href</th><th>Normalized URL</th></tr></thead>
      <tbody>
        {{ range .Report.Links }}{{ if .Skipped }}
        <tr>
          <td>{{ .SkipReason }}</td>
          <td>{{ .LinkText }}</td>
          <td><code>{{ .Href }}</code></td>
          <td><code>{{ .NormalizedURL }}</code></td>
        </tr>
        {{ end }}{{ end }}
      </tbody>
    </table>
    {{ else }}<p class="subtle">No links recorded.</p>{{ end }}
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
    <h2>Safety Notes</h2>
    <ul>{{ range .Report.SafetyNotes }}<li>{{ . }}</li>{{ end }}</ul>
    <h3>Known Limitations</h3>
    <ul>{{ range .Report.Limitations }}<li>{{ . }}</li>{{ end }}</ul>
  </section>
</main>
</body>
</html>`))

func RenderDiscoveryHTMLReport(w io.Writer, report *DiscoveryReport, generatedAt time.Time) error {
	return discoveryHTMLReportTemplate.Execute(w, discoveryHTMLReportData{
		Report:      report,
		GeneratedAt: generatedAt.Format(time.RFC3339),
	})
}

package main

import (
	"encoding/json"
	"html/template"
	"io"
	"time"
)

type htmlReportData struct {
	Project     *Project
	Run         *TestRun
	Report      *Report
	GeneratedAt string
}

var htmlReportTemplate = template.Must(template.New("html-report").Funcs(template.FuncMap{
	"json":       prettyJSON,
	"formatTime": formatReportTime,
	"jsonField":  jsonField,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qualora report - {{ .Project.Name }}</title>
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
    .subtle { color: var(--muted); }
    .grid { display: grid; gap: 16px; }
    .grid.two { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .grid.six { grid-template-columns: repeat(6, minmax(0, 1fr)); }
    section, .metric {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 16px;
    }
    .metric span { display: block; color: var(--muted); font-size: 12px; text-transform: uppercase; }
    .metric strong { display: block; margin-top: 4px; font-size: 22px; }
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
    .status {
      display: inline-block;
      border: 1px solid var(--line);
      border-radius: 999px;
      padding: 2px 8px;
      background: #eef8f7;
      color: var(--strong);
      font-weight: 700;
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
    <p class="subtle">Qualora HTML report</p>
    <h1>{{ .Project.Name }}</h1>
    <p><span class="status">{{ .Report.Status }}</span> <span class="subtle">Run {{ .Run.ID }} generated {{ .GeneratedAt }}</span></p>
  </header>

  <section>
    <h2>Run Summary</h2>
    <div class="grid two">
      <div>
        <p><strong>Project ID:</strong> <code>{{ .Project.ID }}</code></p>
        <p><strong>Run ID:</strong> <code>{{ .Run.ID }}</code></p>
        <p><strong>Created:</strong> {{ formatTime .Run.CreatedAt }}</p>
      </div>
      <div>
        <p><strong>Started:</strong> {{ formatTime .Run.StartedAt }}</p>
        <p><strong>Completed:</strong> {{ formatTime .Run.CompletedAt }}</p>
        <p><strong>Page title:</strong> {{ if .Run.PageTitle }}{{ .Run.PageTitle }}{{ else }}<span class="subtle">Not captured</span>{{ end }}</p>
      </div>
    </div>
  </section>

  <div class="grid six" style="margin: 16px 0;">
    <div class="metric"><span>Total</span><strong>{{ .Report.Summary.TotalFindings }}</strong></div>
    <div class="metric"><span>Critical</span><strong>{{ .Report.Summary.Critical }}</strong></div>
    <div class="metric"><span>High</span><strong>{{ .Report.Summary.High }}</strong></div>
    <div class="metric"><span>Medium</span><strong>{{ .Report.Summary.Medium }}</strong></div>
    <div class="metric"><span>Low</span><strong>{{ .Report.Summary.Low }}</strong></div>
    <div class="metric"><span>Info</span><strong>{{ .Report.Summary.Info }}</strong></div>
  </div>

  {{ with .Report.AIAnalysis }}
  <section style="margin-bottom: 16px;">
    <h2>AI Analysis</h2>
    <div class="grid two">
      <div>
        <p><strong>Status:</strong> <span class="status">{{ .Status }}</span></p>
        <p><strong>Risk level:</strong> {{ if .RiskLevel }}<span class="severity-{{ .RiskLevel }}">{{ .RiskLevel }}</span>{{ else }}<span class="subtle">Not set</span>{{ end }}</p>
        <p><strong>Provider:</strong> {{ if .ProviderName }}{{ .ProviderName }}{{ else }}<span class="subtle">Not available</span>{{ end }}</p>
        <p><strong>Model:</strong> <code>{{ .Model }}</code></p>
      </div>
      <div>
        <p><strong>Prompt tokens:</strong> {{ .PromptTokens }}</p>
        <p><strong>Completion tokens:</strong> {{ .CompletionTokens }}</p>
        <p><strong>Total tokens:</strong> {{ .TotalTokens }}</p>
      </div>
    </div>
    {{ if .ErrorMessage }}<p class="severity-high"><strong>Error:</strong> {{ .ErrorMessage }}</p>{{ end }}
    {{ if .ExecutiveSummary }}<h3>Executive Summary</h3><p>{{ .ExecutiveSummary }}</p>{{ end }}
    {{ if .TechnicalSummary }}<h3>Technical Summary</h3><p>{{ .TechnicalSummary }}</p>{{ end }}
    <h3>Recommendations</h3>
    <pre>{{ jsonField .AnalysisJSON "recommended_actions" }}</pre>
    <h3>Suggested Next Tests</h3>
    <pre>{{ jsonField .AnalysisJSON "suggested_next_tests" }}</pre>
    <h3>Limitations</h3>
    <pre>{{ jsonField .AnalysisJSON "limitations" }}</pre>
  </section>
  {{ end }}

  {{ if .Report.TestPlans }}
  <section style="margin-bottom: 16px;">
    <h2>Related AI Test Plans</h2>
    <table>
      <thead>
        <tr><th>Title</th><th>Status</th><th>Risk</th><th>Scenarios</th><th>Created</th><th>ID</th></tr>
      </thead>
      <tbody>
        {{ range .Report.TestPlans }}
        <tr>
          <td>{{ .Title }}</td>
          <td>{{ .Status }}</td>
          <td>{{ .RiskLevel }}</td>
          <td>{{ .TotalScenarios }}</td>
          <td>{{ formatTime .CreatedAt }}</td>
          <td><code>{{ .ID }}</code></td>
        </tr>
        {{ end }}
      </tbody>
    </table>
  </section>
  {{ end }}

  <section>
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
    <p class="subtle">No findings were recorded for this run.</p>
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
    <p class="subtle">No evidence was recorded for this run.</p>
    {{ end }}
  </section>

  <section style="margin-top: 16px;">
    <h2>Run Metadata</h2>
    <pre>{{ json .Report.Metadata }}</pre>
  </section>
</main>
</body>
</html>`))

func RenderHTMLReport(w io.Writer, project *Project, run *TestRun, report *Report, generatedAt time.Time) error {
	return htmlReportTemplate.Execute(w, htmlReportData{
		Project:     project,
		Run:         run,
		Report:      report,
		GeneratedAt: generatedAt.Format(time.RFC3339),
	})
}

func prettyJSON(value any) string {
	if value == nil {
		return "{}"
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func jsonField(value any, key string) string {
	fields, ok := value.(map[string]any)
	if !ok {
		return "[]"
	}
	item, ok := fields[key]
	if !ok || item == nil {
		return "[]"
	}
	return prettyJSON(item)
}

func formatReportTime(value any) string {
	switch typed := value.(type) {
	case time.Time:
		return typed.Format(time.RFC3339)
	case *time.Time:
		if typed == nil {
			return "Not set"
		}
		return typed.Format(time.RFC3339)
	default:
		return "Not set"
	}
}

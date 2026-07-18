package main

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

func reportIntelligenceHTML(intelligence ReportIntelligence) template.HTML {
	if intelligence.ExecutiveSummary.Headline == "" && intelligence.RawFindingsCount == 0 && len(intelligence.GroupedFindings) == 0 {
		return ""
	}

	var out strings.Builder
	exec := intelligence.ExecutiveSummary
	out.WriteString(`<section style="margin-bottom: 16px;">`)
	out.WriteString(`<h2>Executive Summary</h2>`)
	out.WriteString(`<div class="grid two">`)
	out.WriteString(`<div>`)
	out.WriteString(`<p><strong>Overall status:</strong> <span class="status">` + html.EscapeString(exec.OverallStatus) + `</span></p>`)
	out.WriteString(`<p>` + html.EscapeString(exec.Headline) + `</p>`)
	out.WriteString(`<p class="subtle">Grouped findings are deterministic summaries. Raw findings remain available below.</p>`)
	out.WriteString(`</div>`)
	out.WriteString(`<div>`)
	out.WriteString(`<p><strong>Raw findings:</strong> ` + fmt.Sprint(intelligence.RawFindingsCount) + `</p>`)
	out.WriteString(`<p><strong>Grouped findings:</strong> ` + fmt.Sprint(len(intelligence.GroupedFindings)) + `</p>`)
	out.WriteString(`<p><strong>Duplicates reduced:</strong> ` + fmt.Sprint(intelligence.DeduplicationSummary.DuplicateFindingsReduced) + `</p>`)
	out.WriteString(`</div>`)
	out.WriteString(`</div>`)
	out.WriteString(`<div class="grid six" style="margin-top: 12px;">`)
	out.WriteString(metricHTML("Critical", fmt.Sprint(intelligence.SeverityCounts.Critical)))
	out.WriteString(metricHTML("High", fmt.Sprint(intelligence.SeverityCounts.High)))
	out.WriteString(metricHTML("Medium", fmt.Sprint(intelligence.SeverityCounts.Medium)))
	out.WriteString(metricHTML("Low", fmt.Sprint(intelligence.SeverityCounts.Low)))
	out.WriteString(metricHTML("Info", fmt.Sprint(intelligence.SeverityCounts.Info)))
	out.WriteString(metricHTML("Noisy", fmt.Sprint(intelligence.NoiseSummary.NoisyRepeated)))
	out.WriteString(`</div>`)
	writeStringList(&out, "Recommended Next Actions", exec.RecommendedNextActions)
	writeStringList(&out, "Safety Limitations", intelligence.SafetyLimitations)
	out.WriteString(`</section>`)

	writeGroupedFindingsHTML(&out, "Top Findings", intelligence.TopFindings)
	writeGroupedFindingsHTML(&out, "Grouped Findings", intelligence.GroupedFindings)
	writeAffectedPagesHTML(&out, intelligence.TopAffectedPages)
	writeNoiseSummaryHTML(&out, intelligence.NoiseSummary)
	return template.HTML(out.String())
}

func metricHTML(label string, value string) string {
	return `<div class="metric"><span>` + html.EscapeString(label) + `</span><strong>` + html.EscapeString(value) + `</strong></div>`
}

func writeStringList(out *strings.Builder, title string, values []string) {
	out.WriteString(`<h3 style="margin-top: 16px;">` + html.EscapeString(title) + `</h3>`)
	if len(values) == 0 {
		out.WriteString(`<p class="subtle">None recorded.</p>`)
		return
	}
	out.WriteString(`<ul>`)
	for _, value := range values {
		out.WriteString(`<li>` + html.EscapeString(value) + `</li>`)
	}
	out.WriteString(`</ul>`)
}

func writeGroupedFindingsHTML(out *strings.Builder, title string, groups []GroupedFinding) {
	out.WriteString(`<section style="margin-bottom: 16px;">`)
	out.WriteString(`<h2>` + html.EscapeString(title) + `</h2>`)
	if len(groups) == 0 {
		out.WriteString(`<p class="subtle">No grouped findings were generated.</p></section>`)
		return
	}
	out.WriteString(`<table><thead><tr><th>Severity</th><th>Finding</th><th>Occurrences</th><th>Sources</th><th>Noise</th><th>Affected paths</th><th>Recommendation</th></tr></thead><tbody>`)
	for _, group := range groups {
		out.WriteString(`<tr>`)
		out.WriteString(`<td><span class="severity-` + html.EscapeString(group.NormalizedSeverity) + `">` + html.EscapeString(group.NormalizedSeverity) + `</span></td>`)
		out.WriteString(`<td><strong>` + html.EscapeString(group.Title) + `</strong><br><span class="subtle">` + html.EscapeString(group.Category) + `</span></td>`)
		out.WriteString(`<td>` + fmt.Sprint(group.OccurrencesCount) + `</td>`)
		out.WriteString(`<td>` + html.EscapeString(strings.Join(group.Sources, ", ")) + `</td>`)
		out.WriteString(`<td>` + html.EscapeString(noiseLabel(group.NoiseLevel)) + `</td>`)
		out.WriteString(`<td>` + html.EscapeString(strings.Join(group.AffectedPaths, ", ")) + `</td>`)
		out.WriteString(`<td>` + html.EscapeString(group.Recommendation) + `</td>`)
		out.WriteString(`</tr>`)
	}
	out.WriteString(`</tbody></table></section>`)
}

func writeAffectedPagesHTML(out *strings.Builder, pages []AffectedPageSummary) {
	out.WriteString(`<section style="margin-bottom: 16px;">`)
	out.WriteString(`<h2>Affected Pages</h2>`)
	if len(pages) == 0 {
		out.WriteString(`<p class="subtle">No affected page summary was available.</p></section>`)
		return
	}
	out.WriteString(`<table><thead><tr><th>Path</th><th>Highest severity</th><th>Grouped occurrences</th></tr></thead><tbody>`)
	for _, page := range pages {
		label := page.Path
		if label == "" {
			label = page.URL
		}
		out.WriteString(`<tr><td><code>` + html.EscapeString(label) + `</code></td>`)
		out.WriteString(`<td><span class="severity-` + html.EscapeString(page.HighestSeverity) + `">` + html.EscapeString(page.HighestSeverity) + `</span></td>`)
		out.WriteString(`<td>` + fmt.Sprint(page.FindingsCount) + `</td></tr>`)
	}
	out.WriteString(`</tbody></table></section>`)
}

func writeNoiseSummaryHTML(out *strings.Builder, summary NoiseSummary) {
	out.WriteString(`<section style="margin-bottom: 16px;">`)
	out.WriteString(`<h2>Noise / Repeated Findings</h2>`)
	out.WriteString(`<div class="grid six">`)
	out.WriteString(metricHTML("High signal", fmt.Sprint(summary.HighSignal)))
	out.WriteString(metricHTML("Needs attention", fmt.Sprint(summary.NeedsAttention)))
	out.WriteString(metricHTML("Informational", fmt.Sprint(summary.Informational)))
	out.WriteString(metricHTML("Low noise", fmt.Sprint(summary.LowNoise)))
	out.WriteString(metricHTML("Medium noise", fmt.Sprint(summary.MediumNoise)))
	out.WriteString(metricHTML("High noise", fmt.Sprint(summary.HighNoise)))
	out.WriteString(`</div></section>`)
}

func noiseLabel(noise string) string {
	switch noise {
	case reportNoiseHigh:
		return "Noisy / repeated"
	case reportNoiseMedium:
		return "Informational"
	default:
		return "High signal"
	}
}

package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jlgore/hartea/internal/har"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Generator struct {
	harFiles   []*har.HAR
	analyzers  []*har.Analyzer
	comparison *har.Comparison
}

type Report struct {
	GeneratedAt time.Time       `json:"generated_at"`
	Files       []string        `json:"files"`
	Summary     ReportSummary   `json:"summary"`
	Metrics     []*har.Metrics  `json:"metrics"`
	Comparison  *har.Comparison `json:"comparison,omitempty"`
	Entries     []har.Entry     `json:"entries,omitempty"`
}

type ReportSummary struct {
	TotalFiles      int     `json:"total_files"`
	TotalRequests   int     `json:"total_requests"`
	TotalErrors     int     `json:"total_errors"`
	AverageLoadTime float64 `json:"average_load_time"`
	AverageTTFB     float64 `json:"average_ttfb"`
	TotalTransferMB float64 `json:"total_transfer_mb"`
}

func NewGenerator(harFiles []*har.HAR, analyzers []*har.Analyzer, comparison *har.Comparison) *Generator {
	return &Generator{
		harFiles:   harFiles,
		analyzers:  analyzers,
		comparison: comparison,
	}
}

func (g *Generator) GenerateReport(includeEntries bool) *Report {
	// Calculate summary metrics
	summary := g.calculateSummary()

	// Collect all metrics
	metrics := make([]*har.Metrics, len(g.analyzers))
	for i, analyzer := range g.analyzers {
		metrics[i] = analyzer.CalculateMetrics()
	}

	// File names
	fileNames := make([]string, len(g.harFiles))
	for i := range g.harFiles {
		fileNames[i] = fmt.Sprintf("File %d", i+1)
	}

	report := &Report{
		GeneratedAt: time.Now(),
		Files:       fileNames,
		Summary:     summary,
		Metrics:     metrics,
		Comparison:  g.comparison,
	}

	// Include entries if requested (for detailed analysis)
	if includeEntries && len(g.harFiles) > 0 {
		report.Entries = g.harFiles[0].Log.Entries
	}

	return report
}

func (g *Generator) calculateSummary() ReportSummary {
	summary := ReportSummary{
		TotalFiles: len(g.harFiles),
	}

	if len(g.analyzers) == 0 {
		return summary
	}

	var totalRequests, totalErrors int
	var totalLoadTime, totalTTFB, totalTransferBytes float64

	for _, analyzer := range g.analyzers {
		metrics := analyzer.CalculateMetrics()
		totalRequests += metrics.TotalRequests
		totalErrors += metrics.ErrorRequests
		totalLoadTime += metrics.PageLoadTime
		totalTTFB += metrics.TTFB
		totalTransferBytes += float64(metrics.TotalSize)
	}

	fileCount := float64(len(g.analyzers))
	summary.TotalRequests = totalRequests
	summary.TotalErrors = totalErrors
	summary.AverageLoadTime = totalLoadTime / fileCount
	summary.AverageTTFB = totalTTFB / fileCount
	summary.TotalTransferMB = totalTransferBytes / (1024 * 1024) // Convert to MB

	return summary
}

func (g *Generator) ExportJSON(filename string, includeEntries bool) error {
	report := g.GenerateReport(includeEntries)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func (g *Generator) ExportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	headers := []string{
		"File", "Total Load Time (ms)", "TTFB (ms)", "DNS Time (ms)",
		"Connect Time (ms)", "SSL Time (ms)", "Total Requests",
		"Error Requests", "Third-party Requests", "Cache Hit Ratio (%)",
		"Total Size (MB)",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write metrics for each file
	for i, analyzer := range g.analyzers {
		metrics := analyzer.CalculateMetrics()
		record := []string{
			fmt.Sprintf("File %d", i+1),
			fmt.Sprintf("%.1f", metrics.PageLoadTime),
			fmt.Sprintf("%.1f", metrics.TTFB),
			fmt.Sprintf("%.1f", metrics.DNSTime),
			fmt.Sprintf("%.1f", metrics.ConnectTime),
			fmt.Sprintf("%.1f", metrics.SSLTime),
			fmt.Sprintf("%d", metrics.TotalRequests),
			fmt.Sprintf("%d", metrics.ErrorRequests),
			fmt.Sprintf("%d", metrics.ThirdPartyRequests),
			fmt.Sprintf("%.1f", metrics.CacheHitRatio),
			fmt.Sprintf("%.2f", float64(metrics.TotalSize)/(1024*1024)),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

func (g *Generator) ExportHTML(filename string) error {
	report := g.GenerateReport(false)

	html := g.generateHTMLContent(report)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(html); err != nil {
		return fmt.Errorf("failed to write HTML content: %w", err)
	}

	return nil
}

func (g *Generator) generateHTMLContent(report *Report) string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hartea Analysis Report - Charting Yer Digital Seas</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1, h2, h3 {
            color: #333;
            margin-top: 30px;
        }
        h1 {
            border-bottom: 3px solid #007acc;
            padding-bottom: 10px;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .metric-card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 6px;
            border-left: 4px solid #007acc;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            color: #007acc;
        }
        .metric-label {
            color: #666;
            font-size: 14px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            font-weight: 600;
            color: #333;
        }
        tr:hover {
            background-color: #f8f9fa;
        }
        .improvement {
            color: #28a745;
            font-weight: bold;
        }
        .regression {
            color: #dc3545;
            font-weight: bold;
        }
        .unchanged {
            color: #6c757d;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            color: #666;
            font-size: 14px;
        }
        .status-good { color: #28a745; }
        .status-warning { color: #ffc107; }
        .status-danger { color: #dc3545; }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚓ Hartea Analysis Report - Ahoy Matey!</h1>
        <p><strong>Generated:</strong> ` + report.GeneratedAt.Format("January 2, 2006 at 3:04 PM") + `</p>
        <p><strong>Files Analyzed:</strong> ` + strings.Join(report.Files, ", ") + `</p>`)

	// Summary section
	html.WriteString(`
        <h2>📊 Executive Summary</h2>
        <div class="summary">
            <div class="metric-card">
                <div class="metric-value">` + fmt.Sprintf("%d", report.Summary.TotalFiles) + `</div>
                <div class="metric-label">Files Analyzed</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">` + fmt.Sprintf("%d", report.Summary.TotalRequests) + `</div>
                <div class="metric-label">Total Requests</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">` + fmt.Sprintf("%.1fms", report.Summary.AverageLoadTime) + `</div>
                <div class="metric-label">Average Load Time</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">` + fmt.Sprintf("%.1fms", report.Summary.AverageTTFB) + `</div>
                <div class="metric-label">Average TTFB</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">` + fmt.Sprintf("%.2fMB", report.Summary.TotalTransferMB) + `</div>
                <div class="metric-label">Total Transfer Size</div>
            </div>
            <div class="metric-card">
                <div class="metric-value ` + getErrorStatusClass(report.Summary.TotalErrors) + `">` + fmt.Sprintf("%d", report.Summary.TotalErrors) + `</div>
                <div class="metric-label">Total Errors</div>
            </div>
        </div>`)

	// Detailed metrics
	html.WriteString(`
        <h2>📈 Detailed Metrics</h2>
        <table>
            <thead>
                <tr>
                    <th>File</th>
                    <th>Load Time</th>
                    <th>TTFB</th>
                    <th>Requests</th>
                    <th>Errors</th>
                    <th>Cache Hit %</th>
                    <th>Size (MB)</th>
                </tr>
            </thead>
            <tbody>`)

	for i, metrics := range report.Metrics {
		statusClass := getLoadTimeStatusClass(metrics.PageLoadTime)
		ttfbClass := getTTFBStatusClass(metrics.TTFB)
		errorClass := getErrorStatusClass(metrics.ErrorRequests)

		html.WriteString(fmt.Sprintf(`
                <tr>
                    <td><strong>%s</strong></td>
                    <td class="%s">%.1fms</td>
                    <td class="%s">%.1fms</td>
                    <td>%d</td>
                    <td class="%s">%d</td>
                    <td>%.1f%%</td>
                    <td>%.2f</td>
                </tr>`,
			report.Files[i],
			statusClass, metrics.PageLoadTime,
			ttfbClass, metrics.TTFB,
			metrics.TotalRequests,
			errorClass, metrics.ErrorRequests,
			metrics.CacheHitRatio,
			float64(metrics.TotalSize)/(1024*1024)))
	}

	html.WriteString(`
            </tbody>
        </table>`)

	// Comparison section (if available)
	if report.Comparison != nil {
		html.WriteString(`
        <h2>🔄 Performance Comparison</h2>
        <p><strong>Summary:</strong> ` + fmt.Sprintf("%d improvements, %d regressions, %d unchanged",
			report.Comparison.Summary.BetterCount,
			report.Comparison.Summary.WorseCount,
			report.Comparison.Summary.UnchangedCount) + `</p>
        
        <table>
            <thead>
                <tr>
                    <th>Metric</th>`)

		for i, file := range report.Comparison.Files {
			if i == 0 {
				html.WriteString(`<th>` + file + ` (Base)</th>`)
			} else {
				html.WriteString(`<th>` + file + `</th>`)
			}
		}

		html.WriteString(`
                </tr>
            </thead>
            <tbody>`)

		for _, diff := range report.Comparison.Differences {
			html.WriteString(`<tr><td><strong>` + diff.Name + `</strong></td>`)

			for i, value := range diff.Values {
				if i == 0 {
					html.WriteString(`<td>` + fmt.Sprintf("%v", value) + `</td>`)
				} else {
					change := diff.Changes[i]
					improvement := diff.Improvements[i]
					class := "unchanged"
					if change != "Baseline" && change != "No change" {
						if improvement {
							class = "improvement"
							change += " ✅"
						} else {
							class = "regression"
							change += " ⚠️"
						}
					}
					html.WriteString(`<td>` + fmt.Sprintf("%v", value) + ` <span class="` + class + `">(` + change + `)</span></td>`)
				}
			}

			html.WriteString(`</tr>`)
		}

		html.WriteString(`
            </tbody>
        </table>`)
	}

	// Footer
	html.WriteString(`
        <div class="footer">
            <p>Generated by <strong>Hartea</strong> - Charting the performance seas, one treasure at a time</p>
            <p>Report includes Core Web Vitals, network timing analysis, and performance recommendations.</p>
        </div>
    </div>
</body>
</html>`)

	return html.String()
}

func getLoadTimeStatusClass(loadTime float64) string {
	if loadTime <= 1500 {
		return "status-good"
	} else if loadTime <= 3000 {
		return "status-warning"
	}
	return "status-danger"
}

func getTTFBStatusClass(ttfb float64) string {
	if ttfb <= 200 {
		return "status-good"
	} else if ttfb <= 800 {
		return "status-warning"
	}
	return "status-danger"
}

func getErrorStatusClass(errors int) string {
	if errors == 0 {
		return "status-good"
	} else if errors <= 5 {
		return "status-warning"
	}
	return "status-danger"
}

func (g *Generator) ExportPDF(filename string) error {
	// First generate HTML
	htmlFile := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".html"
	if err := g.ExportHTML(htmlFile); err != nil {
		return fmt.Errorf("failed to generate HTML for PDF: %w", err)
	}

	// Convert HTML to PDF using gofpdf (native approach)
	return g.convertHTMLToPDF(htmlFile, filename)
}

func (g *Generator) convertHTMLToPDF(htmlFile, pdfFile string) error {
	// For this implementation, we'll create a native PDF report
	// rather than converting HTML, which gives us better control
	report := g.GenerateReport(false)
	return g.generateNativePDF(report, pdfFile)
}

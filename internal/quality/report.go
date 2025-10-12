package quality

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"time"
)

// ReportData holds all data needed to generate a comparison report
type ReportData struct {
	GeneratedAt  time.Time
	ImageReports []ImageReport
	Summaries    []BenchmarkSummary
}

// ImageReport holds comparison results for a single source image
type ImageReport struct {
	ImageName   string
	ImageWidth  int
	ImageHeight int
	Results     []*ComparisonResult
}

// GenerateHTMLReport creates an HTML comparison report
func GenerateHTMLReport(data ReportData, w io.Writer) error {
	tmpl := template.Must(template.New("report").Funcs(template.FuncMap{
		"formatBytes":    formatBytes,
		"formatDuration": formatDuration,
		"formatFloat":    formatFloat,
		"toBase64":       toBase64,
		"bestSSIM": func(results []*ComparisonResult) float64 {
			best := 0.0
			for _, r := range results {
				if r.Metrics.SSIM > best {
					best = r.Metrics.SSIM
				}
			}
			return best
		},
		"isBest": func(value, best float64) bool {
			return value >= best-0.001 // Within 0.1% is considered "best"
		},
	}).Parse(htmlTemplate))

	return tmpl.Execute(w, data)
}

// Helper functions for template

func formatBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	kb := float64(b) / 1024.0
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024.0
	return fmt.Sprintf("%.2f MB", mb)
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0f μs", float64(d.Microseconds()))
	}
	return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/1000.0)
}

func formatFloat(f float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, f)
}

func toBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Thumbnail Quality Comparison Report</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 2em;
        }
        .metadata {
            color: #7f8c8d;
            margin-bottom: 30px;
            font-size: 0.9em;
        }
        h2 {
            color: #34495e;
            margin-top: 40px;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #3498db;
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        .summary-card {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 6px;
            padding: 15px;
        }
        .summary-card h3 {
            color: #495057;
            font-size: 1em;
            margin-bottom: 12px;
        }
        .metric-row {
            display: flex;
            justify-content: space-between;
            padding: 6px 0;
            border-bottom: 1px solid #e9ecef;
        }
        .metric-row:last-child {
            border-bottom: none;
        }
        .metric-label {
            color: #6c757d;
            font-size: 0.85em;
        }
        .metric-value {
            font-weight: 600;
            color: #212529;
        }
        .best-value {
            color: #28a745;
        }
        .image-comparison {
            margin-bottom: 60px;
            page-break-inside: avoid;
        }
        .image-comparison h3 {
            color: #2c3e50;
            margin-bottom: 15px;
        }
        .comparison-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 20px;
        }
        .result-card {
            border: 1px solid #dee2e6;
            border-radius: 6px;
            overflow: hidden;
            background: white;
        }
        .result-card.best {
            border: 2px solid #28a745;
            box-shadow: 0 0 10px rgba(40, 167, 69, 0.2);
        }
        .result-header {
            background: #f8f9fa;
            padding: 10px 15px;
            font-weight: 600;
            color: #495057;
            font-size: 0.9em;
        }
        .result-card.best .result-header {
            background: #28a745;
            color: white;
        }
        .result-image {
            width: 100%;
            height: auto;
            display: block;
            background: #f8f9fa;
        }
        .result-metrics {
            padding: 15px;
        }
        .result-metrics table {
            width: 100%;
            font-size: 0.85em;
        }
        .result-metrics td {
            padding: 4px 0;
        }
        .result-metrics td:first-child {
            color: #6c757d;
            width: 60%;
        }
        .result-metrics td:last-child {
            text-align: right;
            font-weight: 600;
        }
        .legend {
            background: #e7f3ff;
            border-left: 4px solid #3498db;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 4px;
        }
        .legend h4 {
            margin-bottom: 10px;
            color: #2c3e50;
        }
        .legend ul {
            margin-left: 20px;
        }
        .legend li {
            margin-bottom: 5px;
            color: #555;
        }
        @media print {
            body {
                background: white;
            }
            .container {
                box-shadow: none;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Thumbnail Quality Comparison Report</h1>
        <div class="metadata">
            Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}
            <br>
            Images tested: {{len .ImageReports}}
            {{if .Summaries}}
            | Approaches compared: {{len .Summaries}}
            {{end}}
        </div>

        <div class="legend">
            <h4>Metrics Explained</h4>
            <ul>
                <li><strong>SSIM</strong> (Structural Similarity): 0-1, higher is better. Measures perceptual similarity.</li>
                <li><strong>PSNR</strong> (Peak Signal-to-Noise Ratio): Measured in dB, higher is better (typically 20-50 dB).</li>
                <li><strong>Sharpness</strong>: Laplacian variance, higher values indicate sharper images.</li>
                <li><strong>Size</strong>: File size in bytes after JPEG compression.</li>
                <li><strong>Time</strong>: Processing time to generate the thumbnail.</li>
            </ul>
        </div>

        {{if .Summaries}}
        <h2>Summary Statistics</h2>
        <div class="summary-grid">
            {{range .Summaries}}
            <div class="summary-card">
                <h3>{{.ApproachName}}</h3>
                <div class="metric-row">
                    <span class="metric-label">Avg SSIM</span>
                    <span class="metric-value">{{formatFloat .AvgSSIM 4}}</span>
                </div>
                <div class="metric-row">
                    <span class="metric-label">Avg PSNR</span>
                    <span class="metric-value">{{formatFloat .AvgPSNR 2}} dB</span>
                </div>
                <div class="metric-row">
                    <span class="metric-label">Avg Sharpness</span>
                    <span class="metric-value">{{formatFloat .AvgSharpness 0}}</span>
                </div>
                <div class="metric-row">
                    <span class="metric-label">Avg Size</span>
                    <span class="metric-value">{{formatBytes .AvgFileSize}}</span>
                </div>
                <div class="metric-row">
                    <span class="metric-label">Avg Time</span>
                    <span class="metric-value">{{formatDuration .AvgProcessingTime}}</span>
                </div>
                <div class="metric-row">
                    <span class="metric-label">Images</span>
                    <span class="metric-value">{{.ImageCount}}</span>
                </div>
            </div>
            {{end}}
        </div>
        {{end}}

        <h2>Detailed Comparisons</h2>
        {{range .ImageReports}}
        <div class="image-comparison">
            <h3>{{.ImageName}} ({{.ImageWidth}}×{{.ImageHeight}})</h3>
            <div class="comparison-grid">
                {{$bestSSIM := bestSSIM .Results}}
                {{range .Results}}
                <div class="result-card{{if isBest .Metrics.SSIM $bestSSIM}} best{{end}}">
                    <div class="result-header">
                        {{.Config.Name}}
                        {{if isBest .Metrics.SSIM $bestSSIM}}★ Best SSIM{{end}}
                    </div>
                    <img src="data:image/jpeg;base64,{{toBase64 .ThumbnailData}}"
                         alt="{{.Config.Name}}"
                         class="result-image"
                         width="{{.WidthPx}}"
                         height="{{.HeightPx}}">
                    <div class="result-metrics">
                        <table>
                            <tr>
                                <td>SSIM</td>
                                <td{{if isBest .Metrics.SSIM $bestSSIM}} class="best-value"{{end}}>
                                    {{formatFloat .Metrics.SSIM 4}}
                                </td>
                            </tr>
                            <tr>
                                <td>PSNR</td>
                                <td>{{formatFloat .Metrics.PSNR 2}} dB</td>
                            </tr>
                            <tr>
                                <td>Sharpness</td>
                                <td>{{formatFloat .Metrics.Sharpness 0}}</td>
                            </tr>
                            <tr>
                                <td>Size</td>
                                <td>{{formatBytes .ThumbnailSize}}</td>
                            </tr>
                            <tr>
                                <td>Time</td>
                                <td>{{formatDuration .ProcessingTime}}</td>
                            </tr>
                            <tr>
                                <td>Dimensions</td>
                                <td>{{.WidthPx}}×{{.HeightPx}}</td>
                            </tr>
                        </table>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
    </div>
</body>
</html>`

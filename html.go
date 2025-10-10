package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/log"
)

// handleRoot handles the root HTML page.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("path", r.URL.Path).Str("method", r.Method).Msg("Request received")

	result, err := checkUmbrella()
	if err != nil {
		log.Error().Err(err).Msg("Failed to check umbrella status")
		http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
		return
	}

	// Build SVG rainfall chart for periods
	barWidth := 30
	barGap := 10
	chartHeight := 80
	chartWidth := len(result.Periods)*(barWidth+barGap) - barGap
	svgBars := ""
	for i, p := range result.Periods {
		// Height for likelihood (blue) and volume (gray)
		lh := int(float64(chartHeight) * float64(p.Likelihood) / 100.0)
		vh := int(float64(chartHeight) * p.Volume / 10.0) // scale 10mm = full height
		x := i * (barWidth + barGap)
		svgBars += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#3498db" rx="4"/>
`, x, chartHeight-lh, barWidth/2-2, lh)
		svgBars += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#888" rx="4"/>
`, x+barWidth/2+2, chartHeight-vh, barWidth/2-2, vh)
	}
	svgChart := fmt.Sprintf(`<svg width="%d" height="%d" style="margin:1rem 0 0.5rem 0;">%s</svg>`, chartWidth, chartHeight+20, svgBars)

	// Prepare JS array of start times for x-axis labels (skip empty)
	startTimes := "["
	for _, p := range result.Periods {
		if p.StartTime == "" {
			continue
		}
		if startTimes != "[" {
			startTimes += ","
		}
		startTimes += fmt.Sprintf("'%s'", p.StartTime)
	}
	startTimes += "]"

	// Template data struct
	type htmlPageData struct {
		Color        string
		Icon         string
		Title        string
		Chance       int
		Volume       float64
		SVGChart     template.HTML
		Info         string
		StartTimesJS template.JS
	}

	var umbrellaTemplate = template.Must(template.New("umbrella").Parse(`<!DOCTYPE html>
           <html>
           <head>
               <meta charset="utf-8">
               <meta name="viewport" content="width=device-width, initial-scale=1">
               <title>Umbrella Check</title>
               <style>
                   body {
                       font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
                       display: flex;
                       justify-content: center;
                       align-items: center;
                       min-height: 100vh;
                       margin: 0;
                       background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                   }
                   .card {
                       background: white;
                       border-radius: 20px;
                       padding: 3rem;
                       box-shadow: 0 20px 60px rgba(0,0,0,0.3);
                       text-align: center;
                       max-width: 400px;
                   }
                   .icon {
                       font-size: 5rem;
                       margin-bottom: 1rem;
                   }
                   h1 {
                       margin: 0 0 0.5rem 0;
                       color: #333;
                       font-size: 2rem;
                   }
                   .stats {
                       font-size: 1.5rem;
                       font-weight: bold;
                       color: {{.Color}};
                       margin: 1rem 0;
                   }
                   .info {
                       color: #666;
                       font-size: 1rem;
                       margin-top: 1rem;
                       line-height: 1.5;
                   }
                   .chart-labels {
                       display: flex;
                       justify-content: center;
                       gap: 1.5rem;
                       font-size: 0.95rem;
                       color: #333;
                       margin-bottom: 0.5rem;
                   }
                   .x-labels {
                       display: flex;
                       justify-content: center;
                       gap: 18px;
                       font-size: 0.85rem;
                       color: #333;
                       margin-top: 0.2rem;
                   }
               </style>
           </head>
           <body>
               <div class="card">
                   <div class="icon">{{.Icon}}</div>
                   <h1>{{.Title}}</h1>
                   <div class="stats">{{.Chance}}% chance · {{printf "%.1f" .Volume}}mm</div>
                   <div class="chart-labels"><span>Likelihood</span><span>Volume</span></div>
                   {{.SVGChart}}
                   <div id="xlabels" class="x-labels"></div>
                   <div class="info">{{.Info}}</div>
               </div>
               <script>
               // Render x-axis labels as local time
               const startTimes = {{.StartTimesJS}};
               const container = document.getElementById('xlabels');
               if (container) {
                   startTimes.forEach((utc, i) => {
                       const d = new Date(utc);
                       const span = document.createElement('span');
                       span.textContent = d.toLocaleString(undefined, { hour: '2-digit' });
                       container.appendChild(span);
                   });
               }
               </script>
           </body>
           </html>`))

	data := htmlPageData{
		Color:        map[bool]string{true: "#e74c3c", false: "#27ae60"}[result.NeedUmbrella],
		Icon:         map[bool]string{true: "☔", false: "☀️"}[result.NeedUmbrella],
		Title:        map[bool]string{true: "Take an umbrella!", false: "No umbrella needed"}[result.NeedUmbrella],
		Chance:       result.PrecipitationChance,
		Volume:       result.PrecipitationVolumeMax,
		SVGChart:     template.HTML(svgChart),
		Info:         map[bool]string{true: "High likelihood and volume of rain", false: "Low likelihood or volume of rain"}[result.NeedUmbrella],
		StartTimesJS: template.JS(startTimes),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := umbrellaTemplate.Execute(w, data); err != nil {
		log.Error().Err(err).Msg("Failed to render template")
	}
}

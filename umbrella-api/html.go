package main

import (
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


	// Template data struct
	type htmlPageData struct {
		Color  string
		Icon   string
		Title  string
		Chance int
		Volume float64
		Info   string
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
               </style>
           </head>
           <body>
               <div class="card">
                   <div class="icon">{{.Icon}}</div>
                   <h1>{{.Title}}</h1>
                   <div class="stats">{{.Chance}}% chance · {{printf "%.1f" .Volume}}mm</div>
                   <div class="info">{{.Info}}</div>
               </div>
           </body>
           </html>`))

	data := htmlPageData{
		Color:  map[bool]string{true: "#e74c3c", false: "#27ae60"}[result.NeedUmbrella],
		Icon:   map[bool]string{true: "☔", false: "☀️"}[result.NeedUmbrella],
		Title:  map[bool]string{true: "Take an umbrella!", false: "No umbrella needed"}[result.NeedUmbrella],
		Chance: result.PrecipitationChance,
		Volume: result.PrecipitationVolumeMax,
		Info:   map[bool]string{true: "High likelihood and volume of rain", false: "Low likelihood or volume of rain"}[result.NeedUmbrella],
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := umbrellaTemplate.Execute(w, data); err != nil {
		log.Error().Err(err).Msg("Failed to render template")
	}
}

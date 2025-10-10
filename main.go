package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// XML structures for parsing BOM data
type Forecast struct {
	XMLName xml.Name `xml:"product"`
	Areas   []Area   `xml:"forecast>area"`
}

type Area struct {
	AAC             string           `xml:"aac,attr"`
	ForecastPeriods []ForecastPeriod `xml:"forecast-period"`
}

type ForecastPeriod struct {
	Index     int               `xml:"index,attr"`
	StartTime string            `xml:"start-time,attr"`
	Texts     []ForecastText    `xml:"text"`
	Elements  []ForecastElement `xml:"element"`
}

type ForecastText struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type ForecastElement struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// UmbrellaResponse is the JSON response structure
type UmbrellaResponse struct {
	NeedUmbrella           bool         `json:"need_umbrella"`
	PrecipitationChance    int          `json:"precipitation_chance_percent"`
	PrecipitationVolumeMax float64      `json:"precipitation_volume_mm"`
	SumProduct             float64      `json:"sum_product"`
	Periods                []RainPeriod `json:"periods"`
	Location               string       `json:"location"`
	Timestamp              string       `json:"timestamp"`
}

type RainPeriod struct {
	Likelihood int     `json:"likelihood"`
	Volume     float64 `json:"volume"`
	StartTime  string  `json:"start_time"`
}

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/umbrella", handleAPI)

	log.Info().Str("port", port).Msg("Starting umbrella API server")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

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
		// Add text label for period index
		svgBars += fmt.Sprintf(`<text x="%d" y="%d" font-size="10" text-anchor="middle" fill="#333">%d</text>
`, x+barWidth/2, chartHeight+12, i+1)
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
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
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
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
            color: %s;
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
        <div class="icon">%s</div>
        <h1>%s</h1>
	<div class="stats">%d%% chance · %.1fmm</div>
	<div class="chart-labels"><span>Likelihood</span><span>Volume</span></div>
	%s
	<div id="xlabels" class="x-labels"></div>
	<div class="info">%s</div>
    </div>
    <script>
    // Render x-axis labels as local time
    const startTimes = %s;
    const container = document.getElementById('xlabels');
    if (container) {
      startTimes.forEach((utc, i) => {
        const d = new Date(utc);
        const span = document.createElement('span');
        span.textContent = d.toLocaleString(undefined, { hour: '2-digit', minute: '2-digit' });
        container.appendChild(span);
      });
    }
    </script>
</body>
</html>`,
		map[bool]string{true: "#e74c3c", false: "#27ae60"}[result.NeedUmbrella],
		map[bool]string{true: "☔", false: "☀️"}[result.NeedUmbrella],
		map[bool]string{true: "Take an umbrella!", false: "No umbrella needed"}[result.NeedUmbrella],
		result.PrecipitationChance,
		result.PrecipitationVolumeMax,
		svgChart,
		map[bool]string{true: "High likelihood and volume of rain", false: "Low likelihood or volume of rain"}[result.NeedUmbrella],
		startTimes,
	)

	fmt.Fprint(w, html)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("path", r.URL.Path).Str("method", r.Method).Msg("API request received")

	// Check for threshold override in query param
	thresholdStr := r.URL.Query().Get("threshold")
	var result *UmbrellaResponse
	var err error
	if thresholdStr != "" {
		threshold, errParse := strconv.ParseFloat(thresholdStr, 64)
		if errParse != nil {
			log.Warn().Str("threshold", thresholdStr).Msg("Invalid threshold param, using default")
			result, err = checkUmbrella()
		} else {
			result, err = checkUmbrella(threshold)
		}
	} else {
		result, err = checkUmbrella()
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check umbrella status")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to fetch weather data"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Use encoding/json to marshal the full struct including periods
	w.Header().Set("Content-Type", "application/json")
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to encode response"}`)
		return
	}
	w.Write(jsonBytes)
}

// checkUmbrella computes the sum-product of (rain likelihood x rain amount) across all forecast periods for NSW_PT131.
// If the sum-product exceeds the threshold, an umbrella is needed.
func checkUmbrella(threshold ...float64) (*UmbrellaResponse, error) {
	log.Info().Msg("Fetching weather data from BOM")

	// Default threshold
	sumProductThreshold := 20.0
	if len(threshold) > 0 {
		sumProductThreshold = threshold[0]
	}

	// Fetch XML from HTTP
	xmlData, err := fetchFromHTTP("http://www.bom.gov.au/fwo/IDN11060.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	log.Debug().Int("bytes", len(xmlData)).Msg("Downloaded XML data")

	// Parse XML
	var forecast Forecast
	if err := xml.Unmarshal(xmlData, &forecast); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	var sumProduct float64
	var precipChanceMax int
	var precipVolumeMax float64
	var periods []RainPeriod
	foundArea := false

	for _, area := range forecast.Areas {
		if area.AAC == "NSW_PT131" {
			foundArea = true
			for _, period := range area.ForecastPeriods {
				var chance int
				var volume float64
				// Extract precipitation chance
				for _, text := range period.Texts {
					if text.Type == "probability_of_precipitation" {
						valueStr := strings.TrimSpace(text.Value)
						valueStr = strings.TrimSuffix(valueStr, "%")
						chance, err = strconv.Atoi(valueStr)
						if err != nil {
							chance = 0 // treat as 0 if parse fails
						}
						if chance > precipChanceMax {
							precipChanceMax = chance
						}
						break
					}
				}
				// Extract precipitation volume (parse "X to Y mm" format)
				for _, element := range period.Elements {
					if element.Type == "precipitation_range" {
						valueStr := strings.TrimSpace(element.Value)
						valueStr = strings.TrimSuffix(valueStr, " mm")
						parts := strings.Split(valueStr, " to ")
						if len(parts) == 2 {
							volume, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
							if err != nil {
								volume = 0.0 // treat as 0 if parse fails
							}
							if volume > precipVolumeMax {
								precipVolumeMax = volume
							}
						}
						break
					}
				}
				sumProduct += float64(chance) * volume / 100.0 // scale chance to 0-1
				// Use period.StartTime (already UTC)
				periods = append(periods, RainPeriod{Likelihood: chance, Volume: volume, StartTime: period.StartTime})
			}
			break
		}
	}

	if !foundArea {
		return nil, fmt.Errorf("NSW_PT131 area not found")
	}

	needUmbrella := sumProduct > sumProductThreshold

	log.Info().
		Float64("sum_product", sumProduct).
		Float64("threshold", sumProductThreshold).
		Bool("need_umbrella", needUmbrella).
		Msg("Weather check complete")

	return &UmbrellaResponse{
		NeedUmbrella:           needUmbrella,
		PrecipitationChance:    precipChanceMax,
		PrecipitationVolumeMax: precipVolumeMax,
		SumProduct:             sumProduct,
		Periods:                periods,
		Location:               "NSW_PT131",
		Timestamp:              time.Now().Format(time.RFC3339),
	}, nil
}

// fetchFromHTTP is a package-level variable to allow mocking in tests
var fetchFromHTTP = realFetchFromHTTP

func realFetchFromHTTP(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set browser-like headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

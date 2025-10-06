package main

import (
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
	AAC      string         `xml:"aac,attr"`
	Elements []ForecastText `xml:"forecast-period>text"`
}

type ForecastText struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// UmbrellaResponse is the JSON response structure
type UmbrellaResponse struct {
	NeedUmbrella         bool   `json:"need_umbrella"`
	PrecipitationChance  int    `json:"precipitation_chance_percent"`
	Location             string `json:"location"`
	Timestamp            string `json:"timestamp"`
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
        .chance {
            font-size: 3rem;
            font-weight: bold;
            color: %s;
            margin: 1rem 0;
        }
        .info {
            color: #666;
            font-size: 0.9rem;
            margin-top: 1rem;
        }
        .location {
            color: #888;
            font-size: 0.85rem;
            margin-top: 0.5rem;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">%s</div>
        <h1>%s</h1>
        <div class="chance">%d%%</div>
        <div class="info">chance of rain</div>
        <div class="location">%s</div>
    </div>
</body>
</html>`,
		map[bool]string{true: "#e74c3c", false: "#27ae60"}[result.NeedUmbrella],
		map[bool]string{true: "☔", false: "☀️"}[result.NeedUmbrella],
		map[bool]string{true: "Take an umbrella!", false: "No umbrella needed"}[result.NeedUmbrella],
		result.PrecipitationChance,
		result.Location,
	)

	fmt.Fprint(w, html)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("path", r.URL.Path).Str("method", r.Method).Msg("API request received")

	result, err := checkUmbrella()
	if err != nil {
		log.Error().Err(err).Msg("Failed to check umbrella status")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to fetch weather data"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"need_umbrella": %t, "precipitation_chance_percent": %d, "location": "%s", "timestamp": "%s"}`,
		result.NeedUmbrella,
		result.PrecipitationChance,
		result.Location,
		result.Timestamp,
	)
}

func checkUmbrella() (*UmbrellaResponse, error) {
	log.Info().Msg("Fetching weather data from BOM")

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

	// Find NSW_PT131 area
	var precipChance int
	found := false
	for _, area := range forecast.Areas {
		if area.AAC == "NSW_PT131" {
			for _, element := range area.Elements {
				if element.Type == "probability_of_precipitation" {
					// Parse percentage value (e.g., "15%")
					valueStr := strings.TrimSpace(element.Value)
					valueStr = strings.TrimSuffix(valueStr, "%")
					precipChance, err = strconv.Atoi(valueStr)
					if err != nil {
						return nil, fmt.Errorf("failed to parse precipitation value: %w", err)
					}
					found = true
					break
				}
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("NSW_PT131 area or precipitation data not found")
	}

	log.Info().Int("precipitation_chance", precipChance).Bool("need_umbrella", precipChance > 5).Msg("Weather check complete")

	return &UmbrellaResponse{
		NeedUmbrella:        precipChance > 5,
		PrecipitationChance: precipChance,
		Location:            "NSW_PT131",
		Timestamp:           time.Now().Format(time.RFC3339),
	}, nil
}

func fetchFromHTTP(url string) ([]byte, error) {
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

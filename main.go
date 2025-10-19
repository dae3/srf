package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// timeNow is a variable that can be mocked in tests
var timeNow = time.Now

// XML structures for parsing BOM data

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

func checkUmbrella(threshold ...float64) (*UmbrellaResponse, error) {
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

	var sumProduct float64
	var precipChanceMax int
	var precipVolumeMax float64
	var periods []RainPeriod
	foundArea := false

	// Find tomorrow's date
	now := timeNow()
	tomorrow := now.AddDate(0, 0, 1)
	tomorrowDateStr := tomorrow.Format("2006-01-02")

	for _, area := range forecast.Areas {
		if area.AAC == "NSW_PT131" {
			foundArea = true

			// Find the forecast period for tomorrow
			var tomorrowPeriod *ForecastPeriod
			log.Debug().Str("tomorrow_date", tomorrowDateStr).Msg("Looking for tomorrow's forecast period")
			for _, period := range area.ForecastPeriods {
				// Parse the start time to check if it's tomorrow
				startTime, err := time.Parse(time.RFC3339, period.StartTime)
				if err != nil {
					log.Debug().Str("start_time", period.StartTime).Err(err).Msg("Failed to parse start time")
					continue // skip if we can't parse the time
				}

				periodDateStr := startTime.Format("2006-01-02")
				log.Debug().Str("period_date", periodDateStr).Str("start_time", period.StartTime).Msg("Checking period")

				// Check if this period is for tomorrow
				if periodDateStr == tomorrowDateStr {
					log.Debug().Str("start_time", period.StartTime).Msg("Found tomorrow's period")
					tomorrowPeriod = &period
					break
				}
			}

			if tomorrowPeriod != nil {
				var chance int
				var volume float64

				// Extract precipitation chance
				for _, text := range tomorrowPeriod.Texts {
					if text.Type == "probability_of_precipitation" {
						valueStr := strings.TrimSpace(text.Value)
						valueStr = strings.TrimSuffix(valueStr, "%")
						chance, err = strconv.Atoi(valueStr)
						if err != nil {
							chance = 0 // treat as 0 if parse fails
						}
						precipChanceMax = chance
						break
					}
				}

				// Extract precipitation volume (parse "X to Y mm" format)
				for _, element := range tomorrowPeriod.Elements {
					if element.Type == "precipitation_range" {
						valueStr := strings.TrimSpace(element.Value)
						valueStr = strings.TrimSuffix(valueStr, " mm")
						parts := strings.Split(valueStr, " to ")
						if len(parts) == 2 {
							volume, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
							if err != nil {
								volume = 0.0 // treat as 0 if parse fails
							}
							precipVolumeMax = volume
						}
						break
					}
				}

				sumProduct = float64(chance) * volume / 100.0 // scale chance to 0-1
				periods = append(periods, RainPeriod{Likelihood: chance, Volume: volume, StartTime: tomorrowPeriod.StartTime})
			}
			break
		}
	}

	if !foundArea {
		return nil, fmt.Errorf("NSW_PT131 area not found")
	}

	// Default threshold
	sumProductThreshold := 20.0
	if len(threshold) > 0 {
		sumProductThreshold = threshold[0]
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
		Timestamp:              timeNow().Format(time.RFC3339),
	}, nil
}

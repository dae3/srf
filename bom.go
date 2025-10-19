package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
	StartTime string            `xml:"start-time-local,attr"`
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

type RainPeriod struct {
	Likelihood int     `json:"likelihood"`
	Volume     float64 `json:"volume"`
	StartTime  string  `json:"start_time"`
}

type UmbrellaResponse struct {
	NeedUmbrella           bool         `json:"need_umbrella"`
	PrecipitationChance    int          `json:"precipitation_chance_percent"`
	PrecipitationVolumeMax float64      `json:"precipitation_volume_mm"`
	SumProduct             float64      `json:"sum_product"`
	Periods                []RainPeriod `json:"periods"`
	Location               string       `json:"location"`
	Timestamp              string       `json:"timestamp"`
}

// timeNow is a variable that can be mocked in tests
var timeNow = time.Now

func checkUmbrella(threshold ...float64) (*UmbrellaResponse, error) {
	// Fetch XML from HTTP
	xmlData, err := fetchFromHTTP("http://www.bom.gov.au/fwo/IDN11060.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

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
			for _, period := range area.ForecastPeriods {
				// Parse the start time to check if it's tomorrow
				startTime, err := time.Parse(time.RFC3339, period.StartTime)
				if err != nil {
					continue // skip if we can't parse the time
				}

				periodDateStr := startTime.Format("2006-01-02")

				// Check if this period is for tomorrow
				if periodDateStr == tomorrowDateStr {
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

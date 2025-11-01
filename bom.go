package main

import (
	"encoding/xml"
	"fmt"
	"time"
)

// timeNow is a variable that can be mocked in tests
var timeNow = time.Now

// bomClient is a package-level variable to allow mocking in tests
var bomClient = NewBOMClient("http://www.bom.gov.au/fwo/IDN11060.xml")

func checkUmbrella(threshold ...float64) (*UmbrellaResponse, error) {
	// Fetch XML from HTTP
	xmlData, err := bomClient.Fetch()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	// Parse XML
	var forecast Forecast
	if err := xml.Unmarshal(xmlData, &forecast); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Find tomorrow's date
	now := timeNow()
	tomorrow := now.AddDate(0, 0, 1)

	// Parse the forecast for the specific location and date
	response, err := forecast.Parse("NSW_PT131", tomorrow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse forecast: %w", err)
	}

	// Default threshold for precipitation chance
	chanceThreshold := 50.0
	if len(threshold) > 0 {
		chanceThreshold = threshold[0]
	}

	response.NeedUmbrella = float64(response.PrecipitationChance) > chanceThreshold
	response.Location = "NSW_PT131"
	response.Timestamp = timeNow().Format(time.RFC3339)

	return response, nil
}

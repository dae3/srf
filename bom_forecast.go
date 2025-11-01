package main

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Forecast represents the top-level structure of the BOM XML forecast data.
type Forecast struct {
	XMLName xml.Name `xml:"product"`
	Areas   []Area   `xml:"forecast>area"`
}

// Area holds the forecast data for a specific region.
type Area struct {
	AAC             string           `xml:"aac,attr"`
	ForecastPeriods []ForecastPeriod `xml:"forecast-period"`
}

// ForecastPeriod contains the forecast for a specific time period.
type ForecastPeriod struct {
	Index     int               `xml:"index,attr"`
	StartTime string            `xml:"start-time-local,attr"`
	Texts     []ForecastText    `xml:"text"`
	Elements  []ForecastElement `xml:"element"`
}

// ForecastText is a textual element within a forecast period.
type ForecastText struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// ForecastElement is a numerical or ranged element within a forecast period.
type ForecastElement struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// Parse processes the raw XML data and extracts the relevant forecast information.
func (f *Forecast) Parse(locationAAC string, targetDate time.Time) (*UmbrellaResponse, error) {
	targetDateStr := targetDate.Format("2006-01-02")
	var area *Area
	for i := range f.Areas {
		if f.Areas[i].AAC == locationAAC {
			area = &f.Areas[i]
			break
		}
	}

	if area == nil {
		return nil, fmt.Errorf("%s area not found", locationAAC)
	}

	var tomorrowPeriod *ForecastPeriod
	for i := range area.ForecastPeriods {
		startTime, err := time.Parse(time.RFC3339, area.ForecastPeriods[i].StartTime)
		if err != nil {
			continue // Skip if time parsing fails
		}
		if startTime.Format("2006-01-02") == targetDateStr {
			tomorrowPeriod = &area.ForecastPeriods[i]
			break
		}
	}

	if tomorrowPeriod == nil {
		return nil, fmt.Errorf("forecast for %s not found for date %s", locationAAC, targetDateStr)
	}

	return parsePeriod(tomorrowPeriod)
}

func parsePeriod(period *ForecastPeriod) (*UmbrellaResponse, error) {
	var parsed UmbrellaResponse
	var err error

	for _, text := range period.Texts {
		if text.Type == "probability_of_precipitation" {
			valueStr := strings.TrimSuffix(strings.TrimSpace(text.Value), "%")
			parsed.PrecipitationChance, err = strconv.Atoi(valueStr)
			if err != nil {
				parsed.PrecipitationChance = 0
			}
		}
	}

	for _, element := range period.Elements {
		switch element.Type {
		case "precipitation_range":
			valueStr := strings.TrimSuffix(strings.TrimSpace(element.Value), " mm")
			parts := strings.Split(valueStr, " to ")
			if len(parts) == 2 {
				parsed.PrecipitationVolumeMax, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err != nil {
					parsed.PrecipitationVolumeMax = 0.0
				}
			}
		case "air_temperature_minimum":
			parsed.MinTemp, err = strconv.Atoi(strings.TrimSpace(element.Value))
			if err != nil {
				parsed.MinTemp = 0
			}
		case "air_temperature_maximum":
			parsed.MaxTemp, err = strconv.Atoi(strings.TrimSpace(element.Value))
			if err != nil {
				parsed.MaxTemp = 0
			}
		case "wind_speed_kilometres":
			parsed.WindSpeed, err = strconv.Atoi(strings.TrimSpace(element.Value))
			if err != nil {
				parsed.WindSpeed = 0
			}
		}
	}

	return &parsed, nil
}

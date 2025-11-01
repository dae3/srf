package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

// UmbrellaResponse represents the JSON response for the umbrella check.
type UmbrellaResponse struct {
	NeedUmbrella           bool    `json:"need_umbrella"`
	PrecipitationChance    int     `json:"precipitation_chance_percent"`
	PrecipitationVolumeMax float64 `json:"precipitation_volume_mm"`
	Location               string  `json:"location"`
	Timestamp              string  `json:"timestamp"`
	MinTemp                int     `json:"min_temp"`
	MaxTemp                int     `json:"max_temp"`
	WindSpeed              int     `json:"wind_speed_kmh"`
}

// TempResponse represents the JSON response for the temperature check.
type TempResponse struct {
	MinTemp int `json:"min_temp"`
	MaxTemp int `json:"max_temp"`
}

// handleAPI handles the /api/umbrella endpoint.
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
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to encode response"}`)
		return
	}
	w.Write(jsonBytes)
}

// handleTemp handles the /api/temp endpoint.
func handleTemp(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("path", r.URL.Path).Str("method", r.Method).Msg("API request received")

	result, err := checkUmbrella()
	if err != nil {
		log.Error().Err(err).Msg("Failed to check umbrella status")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to fetch weather data"}`)
		return
	}

	tempResponse := TempResponse{
		MinTemp: result.MinTemp,
		MaxTemp: result.MaxTemp,
	}

	w.Header().Set("Content-Type", "application/json")
	jsonBytes, err := json.Marshal(tempResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to encode response"}`)
		return
	}
	w.Write(jsonBytes)
}

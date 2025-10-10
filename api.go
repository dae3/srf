package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

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

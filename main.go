package main

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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


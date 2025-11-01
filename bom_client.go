package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// BOMClient is responsible for fetching data from the BOM website.
type BOMClient struct {
	URL    string
	Client *http.Client
}

// NewBOMClient creates a new BOM client with default settings.
func NewBOMClient(url string) *BOMClient {
	return &BOMClient{
		URL: url,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Fetch retrieves the raw XML data from the BOM endpoint.
func (c *BOMClient) Fetch() ([]byte, error) {
	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := c.Client.Do(req)
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

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// override fetchFromHTTP for tests
func mockFetchFromHTTP(url string) ([]byte, error) {
	// This will be overridden in individual tests
	return []byte(""), nil
}

func TestCheckUmbrella_DefaultThreshold(t *testing.T) {
	// Mock time to make tomorrow predictable
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	
	// Patch fetchFromHTTP
	oldFetch := fetchFromHTTP
	fetchFromHTTP = func(url string) ([]byte, error) {
		// Generate XML with tomorrow's date (2025-10-11)
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time="2025-10-10T00:00:00Z">
				<text type="probability_of_precipitation">30%</text>
				<element type="precipitation_range">0 to 2 mm</element>
			</forecast-period>
			<forecast-period start-time="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
		return []byte(xml), nil
	}
	defer func() { fetchFromHTTP = oldFetch }()

	// Mock timeNow to return our fixed time
	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	resp, err := checkUmbrella()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.NeedUmbrella {
		t.Errorf("expected NeedUmbrella to be true, got false")
	}
	if resp.PrecipitationChance != 80 {
		t.Errorf("expected PrecipitationChance to be 80, got %d", resp.PrecipitationChance)
	}
	if resp.PrecipitationVolumeMax != 8.0 {
		t.Errorf("expected PrecipitationVolumeMax to be 8.0, got %f", resp.PrecipitationVolumeMax)
	}
	if len(resp.Periods) != 1 {
		t.Errorf("expected 1 period (tomorrow only), got %d", len(resp.Periods))
	} else {
		// Verify it's tomorrow's period
		if resp.Periods[0].StartTime != "2025-10-11T00:00:00Z" {
			t.Errorf("expected tomorrow's start time, got %s", resp.Periods[0].StartTime)
		}
	}
}

func TestCheckUmbrella_CustomThreshold(t *testing.T) {
	// Mock time to make tomorrow predictable
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	
	oldFetch := fetchFromHTTP
	fetchFromHTTP = func(url string) ([]byte, error) {
		// Generate XML with tomorrow's date (2025-10-11)
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
		return []byte(xml), nil
	}
	defer func() { fetchFromHTTP = oldFetch }()

	// Mock timeNow to return our fixed time
	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	resp, err := checkUmbrella(100.0) // very high threshold
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.NeedUmbrella {
		t.Errorf("expected NeedUmbrella to be false with high threshold, got true")
	}
}

func TestAPIUmbrellaHandler(t *testing.T) {
	// Mock time to make tomorrow predictable
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	
	oldFetch := fetchFromHTTP
	fetchFromHTTP = func(url string) ([]byte, error) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
		return []byte(xml), nil
	}
	defer func() { fetchFromHTTP = oldFetch }()

	// Mock timeNow to return our fixed time
	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	req := httptest.NewRequest("GET", "/api/umbrella", nil)
	rw := httptest.NewRecorder()
	handleAPI(rw, req)
	res := rw.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "need_umbrella") {
		t.Errorf("response missing need_umbrella: %s", body)
	}
	if !strings.Contains(body, "sum_product") {
		t.Errorf("response missing sum_product: %s", body)
	}
	if !strings.Contains(body, "periods") {
		t.Errorf("response missing periods array: %s", body)
	}
	// Check for expected period values (only tomorrow's period now)
	if !strings.Contains(body, "\"likelihood\":80") || !strings.Contains(body, "\"volume\":8") {
		t.Errorf("expected tomorrow's period values not found: %s", body)
	}
}

func TestAPIUmbrellaHandler_ThresholdParam(t *testing.T) {
	// Mock time to make tomorrow predictable
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	
	oldFetch := fetchFromHTTP
	fetchFromHTTP = func(url string) ([]byte, error) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
		return []byte(xml), nil
	}
	defer func() { fetchFromHTTP = oldFetch }()

	// Mock timeNow to return our fixed time
	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	// Use a high threshold to ensure NeedUmbrella is false
	req := httptest.NewRequest("GET", "/api/umbrella?threshold=100.0", nil)
	rw := httptest.NewRecorder()
	handleAPI(rw, req)
	res := rw.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	body := rw.Body.String()
	if !strings.Contains(body, "need_umbrella") {
		t.Errorf("response missing need_umbrella: %s", body)
	}
	if strings.Contains(body, "\"need_umbrella\":true") {
		t.Errorf("expected need_umbrella false with high threshold, got true: %s", body)
	}
	if !strings.Contains(body, "sum_product") {
		t.Errorf("response missing sum_product: %s", body)
	}
	if !strings.Contains(body, "periods") {
		t.Errorf("response missing periods array: %s", body)
	}
}

func TestCheckUmbrella_OnlyTomorrowPeriod(t *testing.T) {
	// Mock time to make tomorrow predictable
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	
	oldFetch := fetchFromHTTP
	fetchFromHTTP = func(url string) ([]byte, error) {
		// XML with multiple periods - today, tomorrow, and day after
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time="2025-10-10T00:00:00Z">
				<text type="probability_of_precipitation">90%</text>
				<element type="precipitation_range">10 to 20 mm</element>
			</forecast-period>
			<forecast-period start-time="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">30%</text>
				<element type="precipitation_range">0 to 2 mm</element>
			</forecast-period>
			<forecast-period start-time="2025-10-12T00:00:00Z">
				<text type="probability_of_precipitation">70%</text>
				<element type="precipitation_range">5 to 15 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
		return []byte(xml), nil
	}
	defer func() { fetchFromHTTP = oldFetch }()

	// Mock timeNow to return our fixed time
	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	resp, err := checkUmbrella()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Should only process tomorrow's period (30%, 2mm), not today's (90%, 20mm) or day after's (70%, 15mm)
	if resp.PrecipitationChance != 30 {
		t.Errorf("expected PrecipitationChance to be 30 (tomorrow's), got %d", resp.PrecipitationChance)
	}
	if resp.PrecipitationVolumeMax != 2.0 {
		t.Errorf("expected PrecipitationVolumeMax to be 2.0 (tomorrow's), got %f", resp.PrecipitationVolumeMax)
	}
	if len(resp.Periods) != 1 {
		t.Errorf("expected exactly 1 period (tomorrow only), got %d", len(resp.Periods))
	} else {
		if resp.Periods[0].StartTime != "2025-10-11T00:00:00Z" {
			t.Errorf("expected tomorrow's start time, got %s", resp.Periods[0].StartTime)
		}
	}
	
	// Sum product should be 30% * 2mm / 100 = 0.6, which is < 20.0 threshold
	expectedSumProduct := 0.6
	if resp.SumProduct != expectedSumProduct {
		t.Errorf("expected SumProduct to be %f, got %f", expectedSumProduct, resp.SumProduct)
	}
	if resp.NeedUmbrella {
		t.Errorf("expected NeedUmbrella to be false (sum product %f < 20.0), got true", resp.SumProduct)
	}
}

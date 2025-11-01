package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(xmlResponse string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, xmlResponse)
	}))
}

func TestCheckUmbrella_DefaultThreshold(t *testing.T) {
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time-local="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
				<element type="air_temperature_minimum">15</element>
				<element type="air_temperature_maximum">25</element>
				<element type="wind_speed_kilometres">30</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
	server := newTestServer(xml)
	defer server.Close()

	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	originalBomClient := bomClient
	bomClient = NewBOMClient(server.URL)
	defer func() { bomClient = originalBomClient }()

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
	if resp.MinTemp != 15 {
		t.Errorf("expected MinTemp to be 15, got %d", resp.MinTemp)
	}
	if resp.MaxTemp != 25 {
		t.Errorf("expected MaxTemp to be 25, got %d", resp.MaxTemp)
	}
	if resp.WindSpeed != 30 {
		t.Errorf("expected WindSpeed to be 30, got %d", resp.WindSpeed)
	}
}

func TestCheckUmbrella_CustomThreshold(t *testing.T) {
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time-local="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
	server := newTestServer(xml)
	defer server.Close()

	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	originalBomClient := bomClient
	bomClient = NewBOMClient(server.URL)
	defer func() { bomClient = originalBomClient }()

	resp, err := checkUmbrella(100.0) // very high threshold
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.NeedUmbrella {
		t.Errorf("expected NeedUmbrella to be false with high threshold, got true")
	}
}

func TestAPIUmbrellaHandler(t *testing.T) {
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time-local="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
	server := newTestServer(xml)
	defer server.Close()

	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	originalBomClient := bomClient
	bomClient = NewBOMClient(server.URL)
	defer func() { bomClient = originalBomClient }()

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
}

func TestAPIUmbrellaHandler_ThresholdParam(t *testing.T) {
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time-local="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">80%</text>
				<element type="precipitation_range">2 to 8 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
	server := newTestServer(xml)
	defer server.Close()

	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	originalBomClient := bomClient
	bomClient = NewBOMClient(server.URL)
	defer func() { bomClient = originalBomClient }()

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
}

func TestCheckUmbrella_OnlyTomorrowPeriod(t *testing.T) {
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time-local="2025-10-10T00:00:00Z">
				<text type="probability_of_precipitation">90%</text>
				<element type="precipitation_range">10 to 20 mm</element>
			</forecast-period>
			<forecast-period start-time-local="2025-10-11T00:00:00Z">
				<text type="probability_of_precipitation">30%</text>
				<element type="precipitation_range">0 to 2 mm</element>
			</forecast-period>
			<forecast-period start-time-local="2025-10-12T00:00:00Z">
				<text type="probability_of_precipitation">70%</text>
				<element type="precipitation_range">5 to 15 mm</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
	server := newTestServer(xml)
	defer server.Close()

	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	originalBomClient := bomClient
	bomClient = NewBOMClient(server.URL)
	defer func() { bomClient = originalBomClient }()

	resp, err := checkUmbrella()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.PrecipitationChance != 30 {
		t.Errorf("expected PrecipitationChance to be 30 (tomorrow's), got %d", resp.PrecipitationChance)
	}
	if resp.PrecipitationVolumeMax != 2.0 {
		t.Errorf("expected PrecipitationVolumeMax to be 2.0 (tomorrow's), got %f", resp.PrecipitationVolumeMax)
	}

	if resp.NeedUmbrella {
		t.Errorf("expected NeedUmbrella to be false (30%% < 50%% threshold), got true")
	}
}

func TestAPITempHandler(t *testing.T) {
	now := time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC)
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<product>
	<forecast>
		<area aac="NSW_PT131">
			<forecast-period start-time-local="2025-10-11T00:00:00Z">
				<element type="air_temperature_minimum">15</element>
				<element type="air_temperature_maximum">25</element>
			</forecast-period>
		</area>
	</forecast>
</product>`
	server := newTestServer(xml)
	defer server.Close()

	originalTimeNow := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = originalTimeNow }()

	originalBomClient := bomClient
	bomClient = NewBOMClient(server.URL)
	defer func() { bomClient = originalBomClient }()

	req := httptest.NewRequest("GET", "/api/temp", nil)
	rw := httptest.NewRecorder()
	handleTemp(rw, req)
	res := rw.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	body := rw.Body.String()
	if !strings.Contains(body, `"min_temp":15`) {
		t.Errorf("response missing min_temp:15: %s", body)
	}
	if !strings.Contains(body, `"max_temp":25`) {
		t.Errorf("response missing max_temp:25: %s", body)
	}
}

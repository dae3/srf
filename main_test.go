package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mock XML data for BOM
const mockBOMXML = `<?xml version="1.0" encoding="UTF-8"?>
<product>
  <forecast>
    <area aac="NSW_PT131">
      <forecast-period>
        <text type="probability_of_precipitation">80%</text>
        <element type="precipitation_range">2 to 8 mm</element>
      </forecast-period>
      <forecast-period>
        <text type="probability_of_precipitation">40%</text>
        <element type="precipitation_range">0 to 2 mm</element>
      </forecast-period>
    </area>
  </forecast>
</product>`

// override fetchFromHTTP for tests
func mockFetchFromHTTP(url string) ([]byte, error) {
	return []byte(mockBOMXML), nil
}

func TestCheckUmbrella_DefaultThreshold(t *testing.T) {
	// Patch fetchFromHTTP
	oldFetch := fetchFromHTTP
	fetchFromHTTP = mockFetchFromHTTP
	defer func() { fetchFromHTTP = oldFetch }()

	resp, err := checkUmbrella()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
       if resp.NeedUmbrella {
	       t.Errorf("expected NeedUmbrella false, got true")
       }
	if resp.SumProduct <= 0 {
		t.Errorf("expected positive sum product, got %v", resp.SumProduct)
	}
}

func TestCheckUmbrella_CustomThreshold(t *testing.T) {
	oldFetch := fetchFromHTTP
	fetchFromHTTP = mockFetchFromHTTP
	defer func() { fetchFromHTTP = oldFetch }()

	resp, err := checkUmbrella(100.0) // very high threshold
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.NeedUmbrella {
		t.Errorf("expected NeedUmbrella false with high threshold, got true")
	}
}

func TestAPIUmbrellaHandler(t *testing.T) {
	oldFetch := fetchFromHTTP
	fetchFromHTTP = mockFetchFromHTTP
	defer func() { fetchFromHTTP = oldFetch }()

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
}

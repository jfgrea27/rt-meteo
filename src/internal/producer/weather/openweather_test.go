package weather

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetCurrentWeatherRaw(t *testing.T) {
	t.Run("successful request returns raw JSON", func(t *testing.T) {
		expected := map[string]interface{}{
			"dt": float64(1700000000),
			"main": map[string]interface{}{
				"temp":     float64(18.0),
				"pressure": float64(1015.0),
				"humidity": float64(65.0),
			},
			"wind": map[string]interface{}{
				"speed": float64(6.0),
			},
			"weather": []interface{}{
				map[string]interface{}{
					"description": "few clouds",
				},
			},
			"name": "London",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		originalAPI := GET_CURRENT_WEATHER_API
		GET_CURRENT_WEATHER_API = server.URL + "?appid=%s&lat=%f&lon=%f"
		defer func() { GET_CURRENT_WEATHER_API = originalAPI }()

		svc := &OpenWeatherService{
			apiKey: "test-key",
			client: server.Client(),
			log:    slog.Default(),
		}

		raw, err := svc.GetCurrentWeatherRaw("London")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it's valid JSON
		var got map[string]interface{}
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("raw response is not valid JSON: %v", err)
		}

		// Verify key fields are present
		if got["name"] != "London" {
			t.Errorf("name = %v, want London", got["name"])
		}
		main, _ := got["main"].(map[string]interface{})
		if main["temp"] != float64(18.0) {
			t.Errorf("temp = %v, want 18.0", main["temp"])
		}
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		originalAPI := GET_CURRENT_WEATHER_API
		GET_CURRENT_WEATHER_API = server.URL + "?appid=%s&lat=%f&lon=%f"
		defer func() { GET_CURRENT_WEATHER_API = originalAPI }()

		svc := &OpenWeatherService{
			apiKey: "bad-key",
			client: server.Client(),
			log:    slog.Default(),
		}

		_, err := svc.GetCurrentWeatherRaw("London")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetHistoricalWeatherRaw(t *testing.T) {
	t.Run("successful request returns raw JSON", func(t *testing.T) {
		expected := map[string]interface{}{
			"list": []interface{}{
				map[string]interface{}{
					"dt": float64(1700000000),
					"main": map[string]interface{}{
						"temp": float64(12.0),
					},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		originalAPI := GET_HISTORICAL_WEATHER_API
		GET_HISTORICAL_WEATHER_API = server.URL + "?appid=%s&lat=%f&lon=%f&type=hour&start=%d&end=%d"
		defer func() { GET_HISTORICAL_WEATHER_API = originalAPI }()

		svc := &OpenWeatherService{
			apiKey: "test-key",
			client: server.Client(),
			log:    slog.Default(),
		}

		from := time.Unix(1699990000, 0)
		to := time.Unix(1700000000, 0)
		raw, err := svc.GetHistoricalWeatherRaw("London", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var got map[string]interface{}
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("raw response is not valid JSON: %v", err)
		}

		list, ok := got["list"].([]interface{})
		if !ok || len(list) != 1 {
			t.Errorf("expected list with 1 entry, got %v", got["list"])
		}
	})
}

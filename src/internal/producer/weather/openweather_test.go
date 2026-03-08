package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConvertOpenWeatherCurrent(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		body := map[string]interface{}{
			"current": map[string]interface{}{
				"dt":         float64(1700000000),
				"temp":       float64(15.5),
				"pressure":   float64(1013.0),
				"humidity":   float64(72.0),
				"wind_speed": float64(5.3),
				"uvi":        float64(3.2),
				"weather": []interface{}{
					map[string]interface{}{
						"description": "clear sky",
					},
				},
			},
		}

		resp, err := convertOpenWeatherCurrent(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Time != time.Unix(1700000000, 0) {
			t.Errorf("Time = %v, want %v", resp.Time, time.Unix(1700000000, 0))
		}
		if resp.Temperature != 15.5 {
			t.Errorf("Temperature = %v, want 15.5", resp.Temperature)
		}
		if resp.Pressure != 1013.0 {
			t.Errorf("Pressure = %v, want 1013.0", resp.Pressure)
		}
		if resp.Humidity != 72.0 {
			t.Errorf("Humidity = %v, want 72.0", resp.Humidity)
		}
		if resp.WindSpeed != 5.3 {
			t.Errorf("WindSpeed = %v, want 5.3", resp.WindSpeed)
		}
		if resp.UV != 3.2 {
			t.Errorf("UV = %v, want 3.2", resp.UV)
		}
		if resp.Description != "clear sky" {
			t.Errorf("Description = %q, want %q", resp.Description, "clear sky")
		}
	})

	t.Run("missing current field", func(t *testing.T) {
		body := map[string]interface{}{}

		_, err := convertOpenWeatherCurrent(body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing dt field", func(t *testing.T) {
		body := map[string]interface{}{
			"current": map[string]interface{}{
				"temp": float64(15.0),
			},
		}

		_, err := convertOpenWeatherCurrent(body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing temp field", func(t *testing.T) {
		body := map[string]interface{}{
			"current": map[string]interface{}{
				"dt": float64(1700000000),
			},
		}

		_, err := convertOpenWeatherCurrent(body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing optional fields", func(t *testing.T) {
		body := map[string]interface{}{
			"current": map[string]interface{}{
				"dt":   float64(1700000000),
				"temp": float64(20.0),
			},
		}

		resp, err := convertOpenWeatherCurrent(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Pressure != 0 {
			t.Errorf("Pressure = %v, want 0", resp.Pressure)
		}
		if resp.Description != "" {
			t.Errorf("Description = %q, want empty", resp.Description)
		}
	})
}

func TestConvertOpenWeatherHistorical(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		body := map[string]interface{}{
			"list": []interface{}{
				map[string]interface{}{
					"dt": float64(1700000000),
					"main": map[string]interface{}{
						"temp":     float64(14.0),
						"pressure": float64(1010.0),
						"humidity": float64(80.0),
					},
					"wind": map[string]interface{}{
						"speed": float64(4.1),
					},
					"weather": []interface{}{
						map[string]interface{}{
							"description": "light rain",
						},
					},
				},
				map[string]interface{}{
					"dt": float64(1700003600),
					"main": map[string]interface{}{
						"temp":     float64(13.0),
						"pressure": float64(1011.0),
						"humidity": float64(85.0),
					},
					"wind": map[string]interface{}{
						"speed": float64(3.5),
					},
					"weather": []interface{}{
						map[string]interface{}{
							"description": "overcast clouds",
						},
					},
				},
			},
		}

		resp, err := convertOpenWeatherHistorical("London", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.City != "London" {
			t.Errorf("City = %q, want %q", resp.City, "London")
		}
		if len(resp.Entries) != 2 {
			t.Fatalf("len(Entries) = %d, want 2", len(resp.Entries))
		}

		e := resp.Entries[0]
		if e.City != "London" {
			t.Errorf("Entries[0].City = %q, want %q", e.City, "London")
		}
		if e.Temperature != 14.0 {
			t.Errorf("Entries[0].Temperature = %v, want 14.0", e.Temperature)
		}
		if e.Description != "light rain" {
			t.Errorf("Entries[0].Description = %q, want %q", e.Description, "light rain")
		}

		e2 := resp.Entries[1]
		if e2.Temperature != 13.0 {
			t.Errorf("Entries[1].Temperature = %v, want 13.0", e2.Temperature)
		}
	})

	t.Run("missing list field", func(t *testing.T) {
		body := map[string]interface{}{}

		_, err := convertOpenWeatherHistorical("London", body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("skips entries without dt", func(t *testing.T) {
		body := map[string]interface{}{
			"list": []interface{}{
				map[string]interface{}{
					"main": map[string]interface{}{
						"temp": float64(14.0),
					},
				},
				map[string]interface{}{
					"dt": float64(1700000000),
					"main": map[string]interface{}{
						"temp": float64(15.0),
					},
				},
			},
		}

		resp, err := convertOpenWeatherHistorical("Paris", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Entries) != 1 {
			t.Fatalf("len(Entries) = %d, want 1", len(resp.Entries))
		}
		if resp.Entries[0].Temperature != 15.0 {
			t.Errorf("Temperature = %v, want 15.0", resp.Entries[0].Temperature)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		body := map[string]interface{}{
			"list": []interface{}{},
		}

		resp, err := convertOpenWeatherHistorical("London", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Entries) != 0 {
			t.Errorf("len(Entries) = %d, want 0", len(resp.Entries))
		}
	})
}

func TestGetCurrentWeather(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"current": map[string]interface{}{
					"dt":         float64(1700000000),
					"temp":       float64(18.0),
					"pressure":   float64(1015.0),
					"humidity":   float64(65.0),
					"wind_speed": float64(6.0),
					"uvi":        float64(2.0),
					"weather": []interface{}{
						map[string]interface{}{
							"description": "few clouds",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Override the API URL to point to test server
		originalAPI := GET_CURRENT_WEATHER_API
		GET_CURRENT_WEATHER_API = server.URL + "?appid=%s&lat=%f&lon=%f"
		defer func() { GET_CURRENT_WEATHER_API = originalAPI }()

		svc := &OpenWeatherService{
			apiKey: "test-key",
			client: server.Client(),
		}

		resp, err := svc.GetCurrentWeather("London")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Temperature != 18.0 {
			t.Errorf("Temperature = %v, want 18.0", resp.Temperature)
		}
		if resp.Description != "few clouds" {
			t.Errorf("Description = %q, want %q", resp.Description, "few clouds")
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
		}

		_, err := svc.GetCurrentWeather("London")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetHistoricalWeather(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"list": []interface{}{
					map[string]interface{}{
						"dt": float64(1700000000),
						"main": map[string]interface{}{
							"temp":     float64(12.0),
							"pressure": float64(1008.0),
							"humidity": float64(90.0),
						},
						"wind": map[string]interface{}{
							"speed": float64(7.0),
						},
						"weather": []interface{}{
							map[string]interface{}{
								"description": "heavy rain",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		originalAPI := GET_HISTORICAL_WEATHER_API
		GET_HISTORICAL_WEATHER_API = server.URL + "?appid=%s&lat=%f&lon=%f&type=hour&start=%d&end=%d"
		defer func() { GET_HISTORICAL_WEATHER_API = originalAPI }()

		svc := &OpenWeatherService{
			apiKey: "test-key",
			client: server.Client(),
		}

		from := time.Unix(1699990000, 0)
		to := time.Unix(1700000000, 0)

		resp, err := svc.GetHistoricalWeather("London", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.City != "London" {
			t.Errorf("City = %q, want %q", resp.City, "London")
		}
		if len(resp.Entries) != 1 {
			t.Fatalf("len(Entries) = %d, want 1", len(resp.Entries))
		}
		if resp.Entries[0].Temperature != 12.0 {
			t.Errorf("Temperature = %v, want 12.0", resp.Entries[0].Temperature)
		}
	})
}

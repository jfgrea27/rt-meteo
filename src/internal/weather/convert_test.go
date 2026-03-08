package weather

import (
	"encoding/json"
	"testing"
	"time"
)

func TestConvertOpenWeatherCurrent(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		body := map[string]interface{}{
			"dt": float64(1700000000),
			"main": map[string]interface{}{
				"temp":     float64(15.5),
				"pressure": float64(1013.0),
				"humidity": float64(72.0),
			},
			"wind": map[string]interface{}{
				"speed": float64(5.3),
			},
			"weather": []interface{}{
				map[string]interface{}{
					"description": "clear sky",
				},
			},
			"name": "London",
		}

		resp, err := ConvertOpenWeatherCurrent(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Time != time.Unix(1700000000, 0) {
			t.Errorf("Time = %v, want %v", resp.Time, time.Unix(1700000000, 0))
		}
		if resp.City != "London" {
			t.Errorf("City = %q, want %q", resp.City, "London")
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
		if resp.Description != "clear sky" {
			t.Errorf("Description = %q, want %q", resp.Description, "clear sky")
		}
	})

	t.Run("missing dt field", func(t *testing.T) {
		body := map[string]interface{}{
			"main": map[string]interface{}{"temp": float64(15.0)},
		}
		_, err := ConvertOpenWeatherCurrent(body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing main field", func(t *testing.T) {
		body := map[string]interface{}{
			"dt": float64(1700000000),
		}
		_, err := ConvertOpenWeatherCurrent(body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing temp field", func(t *testing.T) {
		body := map[string]interface{}{
			"dt":   float64(1700000000),
			"main": map[string]interface{}{"pressure": float64(1013.0)},
		}
		_, err := ConvertOpenWeatherCurrent(body)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing optional fields", func(t *testing.T) {
		body := map[string]interface{}{
			"dt":   float64(1700000000),
			"main": map[string]interface{}{"temp": float64(20.0)},
		}

		resp, err := ConvertOpenWeatherCurrent(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Pressure != 0 {
			t.Errorf("Pressure = %v, want 0", resp.Pressure)
		}
		if resp.WindSpeed != 0 {
			t.Errorf("WindSpeed = %v, want 0", resp.WindSpeed)
		}
		if resp.Description != "" {
			t.Errorf("Description = %q, want empty", resp.Description)
		}
		if resp.City != "" {
			t.Errorf("City = %q, want empty", resp.City)
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
					"wind": map[string]interface{}{"speed": float64(4.1)},
					"weather": []interface{}{
						map[string]interface{}{"description": "light rain"},
					},
				},
				map[string]interface{}{
					"dt": float64(1700003600),
					"main": map[string]interface{}{
						"temp":     float64(13.0),
						"pressure": float64(1011.0),
						"humidity": float64(85.0),
					},
					"wind": map[string]interface{}{"speed": float64(3.5)},
					"weather": []interface{}{
						map[string]interface{}{"description": "overcast clouds"},
					},
				},
			},
		}

		resp, err := ConvertOpenWeatherHistorical("London", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.City != "London" {
			t.Errorf("City = %q, want %q", resp.City, "London")
		}
		if len(resp.Entries) != 2 {
			t.Fatalf("len(Entries) = %d, want 2", len(resp.Entries))
		}
		if resp.Entries[0].Temperature != 14.0 {
			t.Errorf("Entries[0].Temperature = %v, want 14.0", resp.Entries[0].Temperature)
		}
		if resp.Entries[0].Description != "light rain" {
			t.Errorf("Entries[0].Description = %q, want %q", resp.Entries[0].Description, "light rain")
		}
	})

	t.Run("missing list field", func(t *testing.T) {
		_, err := ConvertOpenWeatherHistorical("London", map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("skips entries without dt", func(t *testing.T) {
		body := map[string]interface{}{
			"list": []interface{}{
				map[string]interface{}{
					"main": map[string]interface{}{"temp": float64(14.0)},
				},
				map[string]interface{}{
					"dt":   float64(1700000000),
					"main": map[string]interface{}{"temp": float64(15.0)},
				},
			},
		}

		resp, err := ConvertOpenWeatherHistorical("Paris", body)
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
		resp, err := ConvertOpenWeatherHistorical("London", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Entries) != 0 {
			t.Errorf("len(Entries) = %d, want 0", len(resp.Entries))
		}
	})
}

func TestConvertCurrentWeather(t *testing.T) {
	t.Run("openweather provider", func(t *testing.T) {
		raw, _ := json.Marshal(map[string]interface{}{
			"dt":   float64(1700000000),
			"main": map[string]interface{}{"temp": float64(20.0)},
			"name": "Paris",
		})

		resp, err := ConvertCurrentWeather(OpenWeather, raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Temperature != 20.0 {
			t.Errorf("Temperature = %v, want 20.0", resp.Temperature)
		}
		if resp.City != "Paris" {
			t.Errorf("City = %q, want %q", resp.City, "Paris")
		}
	})

	t.Run("unsupported provider", func(t *testing.T) {
		_, err := ConvertCurrentWeather("unknown", []byte(`{}`))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := ConvertCurrentWeather(OpenWeather, []byte(`not json`))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

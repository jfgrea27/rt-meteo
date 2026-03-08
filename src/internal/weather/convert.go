package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func ConvertOpenWeatherCurrent(body map[string]interface{}) (*CurrentWeatherResponse, error) {
	timestamp, ok := body["dt"].(float64)
	if !ok {
		return nil, errors.New("missing or invalid 'dt' field")
	}

	main, ok := body["main"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing or invalid 'main' field")
	}

	temperature, ok := main["temp"].(float64)
	if !ok {
		return nil, errors.New("missing or invalid 'temp' field")
	}

	pressure, _ := main["pressure"].(float64)
	humidity, _ := main["humidity"].(float64)

	var windSpeed float64
	if wind, ok := body["wind"].(map[string]interface{}); ok {
		windSpeed, _ = wind["speed"].(float64)
	}

	var description string
	if weatherArr, ok := body["weather"].([]interface{}); ok && len(weatherArr) > 0 {
		if weatherObj, ok := weatherArr[0].(map[string]interface{}); ok {
			description, _ = weatherObj["description"].(string)
		}
	}

	var cityName string
	if name, ok := body["name"].(string); ok {
		cityName = name
	}

	resp := &CurrentWeatherResponse{
		Time:        time.Unix(int64(timestamp), 0),
		City:        City(cityName),
		Temperature: float32(temperature),
		Pressure:    float32(pressure),
		Humidity:    float32(humidity),
		WindSpeed:   float32(windSpeed),
		Description: description,
	}

	return resp, nil
}

func ConvertOpenWeatherHistorical(city City, body map[string]interface{}) (*HistoricalWeatherResponse, error) {
	list, ok := body["list"].([]interface{})
	if !ok {
		return nil, errors.New("missing or invalid 'list' field")
	}

	entries := make([]WeatherEntry, 0, len(list))
	for _, item := range list {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		timestamp, ok := entry["dt"].(float64)
		if !ok {
			continue
		}

		main, _ := entry["main"].(map[string]interface{})
		temperature, _ := main["temp"].(float64)
		pressure, _ := main["pressure"].(float64)
		humidity, _ := main["humidity"].(float64)

		wind, _ := entry["wind"].(map[string]interface{})
		windSpeed, _ := wind["speed"].(float64)

		var description string
		if weatherArr, ok := entry["weather"].([]interface{}); ok && len(weatherArr) > 0 {
			if weatherObj, ok := weatherArr[0].(map[string]interface{}); ok {
				description, _ = weatherObj["description"].(string)
			}
		}

		entries = append(entries, WeatherEntry{
			Time:        time.Unix(int64(timestamp), 0),
			City:        city,
			Temperature: float32(temperature),
			Pressure:    float32(pressure),
			Humidity:    float32(humidity),
			WindSpeed:   float32(windSpeed),
			Description: description,
		})
	}

	return &HistoricalWeatherResponse{
		City:    city,
		Entries: entries,
	}, nil
}

func ConvertCurrentWeather(provider Provider, raw []byte) (*CurrentWeatherResponse, error) {
	switch provider {
	case OpenWeather:
		var body map[string]interface{}
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw content: %w", err)
		}
		return ConvertOpenWeatherCurrent(body)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

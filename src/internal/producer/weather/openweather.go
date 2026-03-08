package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	GET_CURRENT_WEATHER_API    = "https://api.openweather.org/data/3.0/onecall?appid=%s&lat=%f&lon=%f"
	GET_HISTORICAL_WEATHER_API = "https://history.openweather.org/data/2.5/history/city?appid=%s&lat=%f&lon=%f&type=hour&start=%d&end=%d"
)

type OpenWeatherService struct {
	apiKey string
	client *http.Client
}

func convertOpenWeatherCurrent(body map[string]interface{}) (*CurrentWeatherResponse, error) {
	curr, ok := body["current"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing or invalid 'current' field")
	}

	timestamp, ok := curr["dt"].(float64)
	if !ok {
		return nil, errors.New("missing or invalid 'dt' field")
	}

	temperature, ok := curr["temp"].(float64)
	if !ok {
		return nil, errors.New("missing or invalid 'temp' field")
	}

	pressure, _ := curr["pressure"].(float64)
	humidity, _ := curr["humidity"].(float64)
	windSpeed, _ := curr["wind_speed"].(float64)
	uvi, _ := curr["uvi"].(float64)

	var description string
	if weatherArr, ok := curr["weather"].([]interface{}); ok && len(weatherArr) > 0 {
		if weatherObj, ok := weatherArr[0].(map[string]interface{}); ok {
			description, _ = weatherObj["description"].(string)
		}
	}

	resp := &CurrentWeatherResponse{
		Time:        time.Unix(int64(timestamp), 0),
		Temperature: float32(temperature),
		Pressure:    float32(pressure),
		Humidity:    float32(humidity),
		WindSpeed:   float32(windSpeed),
		UV:          float32(uvi),
		Description: description,
	}

	return resp, nil
}

func (s *OpenWeatherService) GetCurrentWeather(c City) (*CurrentWeatherResponse, error) {

	coord := CITY_COORDINATES[c]

	api := fmt.Sprintf(GET_CURRENT_WEATHER_API, s.apiKey, coord.Lat, coord.Lon)

	resp, err := s.client.Get(api)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d != %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	res, err := convertOpenWeatherCurrent(body)
	if err != nil {
		return nil, fmt.Errorf("error constructing CurrentWeatherResponse: %w", err)
	}
	return res, nil
}

func convertOpenWeatherHistorical(city City, body map[string]interface{}) (*HistoricalWeatherResponse, error) {
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

func (s *OpenWeatherService) GetHistoricalWeather(city City, from time.Time, to time.Time) (*HistoricalWeatherResponse, error) {
	coord := CITY_COORDINATES[city]

	api := fmt.Sprintf(GET_HISTORICAL_WEATHER_API, s.apiKey, coord.Lat, coord.Lon, from.Unix(), to.Unix())

	resp, err := s.client.Get(api)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code %d != %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	res, err := convertOpenWeatherHistorical(city, body)
	if err != nil {
		return nil, fmt.Errorf("error constructing HistoricalWeatherResponse: %w", err)
	}

	return res, nil
}

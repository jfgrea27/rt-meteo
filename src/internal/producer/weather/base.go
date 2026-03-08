package weather

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jfgrea27/rt-meteo/internal/utils"
)

type WeatherService interface {
	GetCurrentWeather(c City) (*CurrentWeatherResponse, error)
	GetHistoricalWeather(c City, from time.Time, to time.Time) (*HistoricalWeatherResponse, error)
}

func ConstructWeatherService(p WeatherProvider) WeatherService {
	var svc WeatherService
	switch p {
	case OpenWeather:
		svc = &OpenWeatherService{
			apiKey: utils.GetEnvVar("OPEN_WEATHER_API_KEY", false),
			client: &http.Client{},
		}
	default:
		panic(fmt.Sprintf("%s is not a valid weather provider", p))
	}
	return svc
}

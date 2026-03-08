package weather

type WeatherProvider string

const (
	OpenWeather WeatherProvider = "openweather"
)

var CITY_COORDINATES = map[City]Coordinate{
	"London": {
		Lat: 51.5073219,
		Lon: -0.1276474,
	},
	"Paris": {
		Lat: 48.8588897,
		Lon: 2.3200410217200766,
	},
}

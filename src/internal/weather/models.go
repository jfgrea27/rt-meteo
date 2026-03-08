package weather

import "time"

type City string

type Coordinate struct {
	Lat float32
	Lon float32
}

type WeatherEntry struct {
	Time time.Time
	City City

	Temperature float32
	Pressure    float32
	Humidity    float32
	WindSpeed   float32
	UV          float32
	Description string
}

type CurrentWeatherResponse = WeatherEntry

type HistoricalWeatherResponse struct {
	City    City
	Entries []WeatherEntry
}

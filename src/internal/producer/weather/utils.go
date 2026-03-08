package weather

func AggregateCurrentWeather(svc WeatherService) []*CurrentWeatherResponse {
	weathers := make([]*CurrentWeatherResponse, len(CITY_COORDINATES))

	for city := range CITY_COORDINATES {
		weather, err := svc.GetCurrentWeather(city)

		if err != nil {
			panic(err)
		}
		weathers = append(weathers, weather)
	}
	return weathers
}

package main

import (
	"context"
	"time"

	"github.com/jfgrea27/rt-meteo/internal/producer/weather"
	"github.com/jfgrea27/rt-meteo/internal/queue"
	"github.com/jfgrea27/rt-meteo/internal/utils"
)

func main() {
	// TODO: set up logging

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// setup weather service
	weatherProvider := weather.WeatherProvider(utils.GetEnvVar("WEATHER_PROVIDER", true))
	weatherSvc := weather.ConstructWeatherService(weatherProvider)

	//setup queue service
	queueProvider := queue.QueueProvider(utils.GetEnvVar("QUEUE_PROVIDER", true))
	queueName := utils.GetEnvVar("QUEUE_NAME", true)
	queueSvc := queue.ConstructQueueService(queueProvider)

	weathers := weather.AggregateCurrentWeather(weatherSvc)

	queueSvc.Produce(ctx, weathers, queueName)
}

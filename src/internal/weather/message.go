package weather

import "encoding/json"

type WeatherMessage struct {
	Provider Provider        `json:"provider"`
	Content  json.RawMessage `json:"content"`
}

package weather_station

import (
	"net/http"
	"weather_station/config"
)

type Environment struct {
	Temperature int64
	Pressure    int64
	Humidity    int32
}

type WeatherStation interface {
	ServeHTTP(w http.ResponseWriter, _ *http.Request)
	Start()
	Init(config *config.Config)
}

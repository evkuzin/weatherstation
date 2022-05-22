package weather_station

import (
	"github.com/evkuzin/weatherstation/config"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Environment struct {
	Temperature int64
	Pressure    int64
	Humidity    int32
	Time        time.Time
}

type WeatherStation interface {
	ServeHTTP(w http.ResponseWriter, _ *http.Request)
	Start()
	Init(config *config.Config, logger *logrus.Logger) error
}

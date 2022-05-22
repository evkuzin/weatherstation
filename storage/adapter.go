package storage

import (
	"github.com/evkuzin/weatherstation/config"
	"github.com/evkuzin/weatherstation/weather_station"
	"time"
)

type Adapter interface {
	Init(config *config.Config) error
	Put(event *weather_station.Environment) error
	GetEvents(t time.Duration) []weather_station.Environment
}

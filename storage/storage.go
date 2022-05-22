package storage

import (
	"time"
	"weather_station/weather_station"
)

type Storage struct {
}

func (s *Storage) GetStats(time time.Duration) {

}

func (s *Storage) Put(event *weather_station.Environment) {
}

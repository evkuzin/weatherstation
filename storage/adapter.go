package storage

import (
	"context"
	"weather_station/config"
)

type adapter interface {
	Init(ctx context.Context, config *config.Config) error
	Put(ctx context.Context, event *Event) error
}

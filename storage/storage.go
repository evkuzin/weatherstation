package storage

import (
	"fmt"
	"github.com/evkuzin/weatherstation/config"
	"github.com/evkuzin/weatherstation/weather_station"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type Environment struct {
	temperature int64 `gorm:"primaryKey"`
	pressure    int64
	humidity    int32
	time        time.Time
}

type Storage struct {
	db *gorm.DB
}

func (s *Storage) Init(config *config.Config) error {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.Database,
		config.Database.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return err
	}
	s.db = db
	err = s.db.AutoMigrate(&Environment{})
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Put(event *weather_station.Environment) error {
	tx := s.db.Create(&Environment{
		temperature: event.Temperature,
		pressure:    event.Pressure,
		humidity:    event.Humidity,
		time:        event.Time,
	})
	return tx.Error
}

func (s *Storage) GetEvents(t time.Duration) []weather_station.Environment {
	var events []weather_station.Environment
	s.db.Where("time > (?)", t).Find(&events)
	return events
}

func NewStorage() Adapter {
	return &Storage{}
}

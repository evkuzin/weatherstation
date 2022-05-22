package main

import (
	"github.com/evkuzin/weatherstation/config"
	"github.com/evkuzin/weatherstation/weather_station/impl"
	"net/http"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    &logrus.TextFormatter{},
		Level:        logrus.DebugLevel,
		ReportCaller: true,
	}
	wg := &sync.WaitGroup{}
	conf, err := config.NewConfig("config.yaml")
	if err != nil {
		logger.Errorf("cannot parse config: %s", err)
		os.Exit(1)
	}
	ws := impl.NewWeatherStation()
	err = ws.Init(conf, logger)
	if err != nil {
		logger.Errorf("cannot init weather station: %s", err)
		os.Exit(1)
	}
	logger.Info("peripheral init complete...")
	wg.Add(1)
	go ws.Start()

	err = http.ListenAndServe(":8080", ws)
	if err != nil {
		logger.Warnf("Cannot start stats server. %v", err.Error())
	}
	wg.Wait()
	logger.Info("all threads killed, shutdown...")
}

package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    &logrus.TextFormatter{},
		Level:        logrus.InfoLevel,
		ReportCaller: true,
	}
	wg := &sync.WaitGroup{}

	sensor, err := PeripheralInitialisation(logger)
	if err != nil {
		logger.Errorf("cannot init periph: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		err := sensor.Halt()
		if err != nil {
			logger.Warnf("Error during shutdown sensor: %v", err.Error())
		}
	}()
	ch := make(chan struct{}, 1)
	ws := NewWeatherStation(sensor, logger, ch, wg)
	wg.Add(1)
	ws.Start()

	err = http.ListenAndServe(":8080", ws)
	if err != nil {
		logger.Warnf("Cannot start stats server. %v", err.Error())
	}
	wg.Wait()
	logger.Info("all threads killed, shutdown...")
}

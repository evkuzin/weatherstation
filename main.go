package main

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

func main() {
	logger := &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    &logrus.TextFormatter{},
		Level:        logrus.InfoLevel,
		ReportCaller: true,
	}
	wg := &sync.WaitGroup{}

	// Make sure periph is initialized.
	state, err := host.Init()
	if err != nil {
		logger.Debugf("failed to initialize periph: %v", err)
	}

	// Prints the loaded driver.
	logger.Debugf("Using drivers:\n")
	for _, driver := range state.Loaded {
		logger.Debugf("- %s\n", driver)
	}

	// Prints the driver that were skipped as irrelevant on the platform.
	logger.Debugf("Drivers skipped:\n")
	for _, failure := range state.Skipped {
		logger.Debugf("- %s: %s\n", failure.D, failure.Err)
	}

	// Having drivers failing to load may not require process termination. It
	// is possible to continue to run in partial failure mode.
	logger.Debugf("Drivers failed to load:\n")
	for _, failure := range state.Failed {
		logger.Debugf("- %s: %v\n", failure.D, failure.Err)
	}

	// Open default I2C bus
	bus, err := i2creg.Open("")
	if err != nil {
		logger.Debugf("cannot open a bus")
		logger.Debugf(err.Error())
		os.Exit(1)
	} else {
		logger.Debugf("I2C bus open call successful. Got: %v", bus.String())
	}
	defer bus.Close()
	stopCh := make(chan struct{})
	wg.Add(1)

	// Open a handle to a bme280/bmp280 connected on the I²C bus using Indoor navigation:
	// continuous sampling at 40ms with filter F16, pressure
	// O16x, temperature O2x, humidity O1x, filter F16. Power consumption 633µA.
	// RMS noise: 0.2Pa / 1.7cm.
	sensor, err := bmxx80.NewI2C(bus, 0x76, &bmxx80.Opts{
		Temperature: bmxx80.O2x,
		Pressure:    bmxx80.O16x,
		Humidity:    bmxx80.O1x,
		Filter:      bmxx80.F16,
	})
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		err := sensor.Halt()
		if err != nil {
			logger.Warnf("Error during shutdown sensor: %v", err.Error())
		}
	}()

	demeter := NewWeatherStation(sensor, logger, stopCh, wg)
	wg.Add(1)
	demeter.Start()

	// err = http.ListenAndServe(":8080", demeter)
	// if err != nil {
	//	 logger.Warnf("Cannot start stats server. %v", err.Error())
	// }
	wg.Wait()
	logger.Info("all threads killed, shutdown...")
}
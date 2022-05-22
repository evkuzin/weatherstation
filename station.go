package main

import (
	"fmt"
	"sync"
	"time"

	"periph.io/x/conn/v3/physic"

	"github.com/sirupsen/logrus"
	"periph.io/x/devices/v3/bmxx80"
)

const (
	scanFreq             = 2 * time.Second
	temperatureThreshold = 37
	temperatureZone      = 0.1
	logFrequency         = time.Second * 10
	// ZeroCelsius static variable for Celsius Kelvin conversion
	ZeroCelsius = 237.7
	aConstant   = 17.27
)

var (
	heaterInitLoopTime = 50 * time.Millisecond
)

type circularBuffer struct {
	values   []physic.Env
	position int
}

func (b *circularBuffer) add(env physic.Env) {
	if b.position < len(b.values) {
		b.values[b.position] = env
		b.position++
	} else {
		copy(b.values[:len(b.values)-1], b.values[1:])
		b.values[b.position-1] = env
	}
}

func (b *circularBuffer) get() []physic.Env {
	return b.values
}

// Demeter struct contains info about all environment
// relay controls and logging
type Demeter struct {
	Environment *bmxx80.Dev
	Logger      *logrus.Logger
	stop        chan struct{}
	wg          *sync.WaitGroup
	lastValues  *circularBuffer
	inertness   time.Duration
	envSpeed    float64
}

//
//func (d *Demeter) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
//	// create a new line instance
//	line := charts.NewLine()
//	// set some global options like Title/Legend/ToolTip or anything else
//	line.SetGlobalOptions(
//		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
//		charts.WithTitleOpts(opts.Title{
//			Title: "Temperature",
//		}))
//	samples := d.Monitoring.GetStats(time.Minute * 5)
//	xTime := make([]int64, len(*samples))
//	yTemperature := make([]opts.LineData, len(*samples))
//	var symbol string
//	for i, sample := range *samples {
//		if sample.HeaterState {
//			symbol = "circle"
//		} else {
//			symbol = "diamond"
//		}
//		xTime[i] = sample.Time.Unix()
//		yTemperature[i] = opts.LineData{Value: sample.Temperature, Symbol: symbol}
//	}
//	line.SetXAxis(xTime).
//		AddSeries("Temperature", yTemperature)
//	err := line.Render(w)
//	d.Logger.Infof("build graph based on %d last metrics", len(*samples))
//	if err != nil {
//		d.Logger.Infof("Unable to render graph. %v", err.Error())
//	}
//}

// Start is the main daemon loop
func (d *Demeter) Start() {
	defer d.wg.Done()
	d.Logger.Info("Weather station starting...")

	envCh, err := d.Environment.SenseContinuous(scanFreq)
	if err != nil {
		d.Logger.Fatalf("Cannot read from device: %v", err.Error())
		return
	}
	defer func(Environment *bmxx80.Dev) {
		err := Environment.Halt()
		if err != nil {
			fmt.Errorf("error: %s", err.Error())
		}
	}(d.Environment)
	var env physic.Env
	for {
		select {
		case <-d.stop:
			d.Logger.Info("Stopping Demeter")
			return
		case env = <-envCh:
			fmt.Printf(
				"Temperature: %s\nHumidity: %s\nPressure: %s\n",
				env.Temperature,
				env.Humidity,
				env.Pressure,
			)
		}
	}
}

func (d *Demeter) updateLastValues(env physic.Env) {
	d.lastValues.add(env)
}

// NewWeatherStation return a new instance of a Demeter daemon
func NewWeatherStation(
	sensor *bmxx80.Dev,
	logger *logrus.Logger,
	stopChan chan struct{},
	wg *sync.WaitGroup,
) *Demeter {

	return &Demeter{
		Environment: sensor,
		Logger:      logger,
		stop:        stopChan,
		wg:          wg,
	}
}

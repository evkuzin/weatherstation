package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"net/http"
	"sync"
	"time"
	"weather_station/storage"

	"periph.io/x/conn/v3/physic"

	"github.com/sirupsen/logrus"
	"periph.io/x/devices/v3/bmxx80"
)

const (
	scanFreq = 2 * time.Second
)

// WeatherStation struct contains info about all environment
// relay controls and logging
type WeatherStation struct {
	Environment *bmxx80.Dev
	Logger      *logrus.Logger
	stop        chan struct{}
	wg          *sync.WaitGroup
	inertness   time.Duration
	envSpeed    float64
	Storage     *storage.Storage
}

func (ws *WeatherStation) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title: "Temperature",
		}))
	samples := ws.Storage.GetStats(time.Minute * 5)
	xTime := make([]int64, len(*samples))
	yTemperature := make([]opts.LineData, len(*samples))
	var symbol string
	for i, sample := range *samples {
		if sample.HeaterState {
			symbol = "circle"
		} else {
			symbol = "diamond"
		}
		xTime[i] = sample.Time.Unix()
		yTemperature[i] = opts.LineData{Value: sample.Temperature, Symbol: symbol}
	}
	line.SetXAxis(xTime).
		AddSeries("Temperature", yTemperature)
	err := line.Render(w)
	ws.Logger.Infof("build graph based on %ws last metrics", len(*samples))
	if err != nil {
		ws.Logger.Infof("Unable to render graph. %v", err.Error())
	}
}

// Start is the main daemon loop
func (ws *WeatherStation) Start() {
	defer ws.wg.Done()
	ws.Logger.Info("Weather station starting...")

	envCh, err := ws.Environment.SenseContinuous(scanFreq)
	if err != nil {
		ws.Logger.Fatalf("Cannot read from device: %v", err.Error())
		return
	}
	defer func(Environment *bmxx80.Dev) {
		err := Environment.Halt()
		if err != nil {
			logrus.Errorf("error: %s", err.Error())
		}
	}(ws.Environment)
	var env physic.Env
	for {
		select {
		case <-ws.stop:
			ws.Logger.Info("Stopping WeatherStation")
			return
		case env = <-envCh:
			fmt.Printf(
				"Temperature: %s\nHumidity: %s\nPressure: %s\n\n",
				env.Temperature,
				env.Humidity,
				env.Pressure,
			)
		}
	}
}

// NewWeatherStation return a new instance of a WeatherStation daemon
func NewWeatherStation(
	sensor *bmxx80.Dev,
	logger *logrus.Logger,
	stopChan chan struct{},
	wg *sync.WaitGroup,
) *WeatherStation {

	return &WeatherStation{
		Environment: sensor,
		Logger:      logger,
		stop:        stopChan,
		wg:          wg,
	}
}

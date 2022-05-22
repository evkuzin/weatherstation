package impl

import (
	"fmt"
	"github.com/evkuzin/weatherstation/config"
	"github.com/evkuzin/weatherstation/storage"
	"github.com/evkuzin/weatherstation/weather_station"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"net/http"
	"sync"
	"time"

	"periph.io/x/conn/v3/physic"

	"github.com/sirupsen/logrus"
	"periph.io/x/devices/v3/bmxx80"
)

const (
	scanFreq = 2 * time.Second
)

// weatherStationImpl struct contains info about all environment
// relay controls and logging
type weatherStationImpl struct {
	sensor  *bmxx80.Dev
	logger  *logrus.Logger
	stop    chan struct{}
	wg      *sync.WaitGroup
	Storage storage.Adapter
}

func (ws *weatherStationImpl) Init(config *config.Config, logger *logrus.Logger) error {
	sensor, err := peripheralInitialisation(logger)
	if err != nil {
		return err
	}
	ws.sensor = sensor
	ws.logger = logger
	ws.Storage = storage.NewStorage()
	err = ws.Storage.Init(config)
	if err != nil {
		return err
	}
	return nil
}

func (ws *weatherStationImpl) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title: "Temperature",
		}))
	samples := ws.Storage.GetEvents(time.Minute * 5)
	xTime := make([]int64, len(samples))
	yTemperature := make([]opts.LineData, len(samples))
	//var symbol string
	//for i, sample := range samples {
	//	if sample.HeaterState {
	//		symbol = "circle"
	//	} else {
	//		symbol = "diamond"
	//	}
	//	xTime[i] = sample.Time.Unix()
	//	yTemperature[i] = opts.LineData{Value: sample.Temperature, Symbol: symbol}
	//}
	line.SetXAxis(xTime).
		AddSeries("Temperature", yTemperature)
	err := line.Render(w)
	ws.logger.Infof("build graph based on %ws last metrics", len(samples))
	if err != nil {
		ws.logger.Infof("Unable to render graph. %v", err.Error())
	}
}

// Start is the main daemon loop
func (ws *weatherStationImpl) Start() {
	defer ws.wg.Done()
	ws.logger.Info("Weather station starting...")

	envCh, err := ws.sensor.SenseContinuous(scanFreq)
	if err != nil {
		ws.logger.Fatalf("Cannot read from device: %v", err.Error())
		return
	}
	defer func(Environment *bmxx80.Dev) {
		err := Environment.Halt()
		if err != nil {
			logrus.Errorf("error: %s", err.Error())
		}
	}(ws.sensor)
	var env physic.Env
	for {
		select {
		case <-ws.stop:
			ws.logger.Info("Stopping weatherStationImpl")
			return
		case env = <-envCh:
			fmt.Printf(
				"Temperature: %s\nHumidity: %s\nPressure: %s\n\n",
				env.Temperature,
				env.Humidity,
				env.Pressure,
			)
			err := ws.Storage.Put(&weather_station.Environment{
				Temperature: int64(env.Temperature),
				Pressure:    int64(env.Pressure),
				Humidity:    int32(env.Humidity),
				Time:        time.Now(),
			})
			if err != nil {
				ws.logger.Warnf("cannot write to storage: %s", err.Error())
			}
		}
	}
}

// NewWeatherStation return a new instance of a weatherStationImpl daemon
func NewWeatherStation() weather_station.WeatherStation {

	return &weatherStationImpl{}
}

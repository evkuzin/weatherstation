package impl

import (
	"github.com/evkuzin/weatherstation/config"
	"github.com/evkuzin/weatherstation/storage"
	"github.com/evkuzin/weatherstation/weather_station"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"net/http"
	"periph.io/x/conn/v3/i2c"
	"sync"
	"time"

	"periph.io/x/conn/v3/physic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	bus     i2c.BusCloser
	tg      *tgbotapi.BotAPI
}

func (ws *weatherStationImpl) Init(config *config.Config, logger *logrus.Logger) error {
	ws.logger = logger

	bus, sensor, err := peripheralInitialisation(logger)
	if err != nil {
		return err
	}
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Key)
	if err != nil {
		return err
	}
	bot.Debug = config.Telegram.Debug
	ws.logger.Infof("Telegram authorized on account %s", bot.Self.UserName)
	go ws.telegramStart()
	ws.sensor = sensor
	ws.bus = bus
	ws.tg = bot
	ws.Storage = storage.NewStorage()
	err = ws.Storage.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Start is the main daemon loop
func (ws *weatherStationImpl) Start() {
	defer func(bus i2c.BusCloser) {
		err := bus.Close()
		if err != nil {
			ws.logger.Errorf("error: %s", err.Error())
		}
	}(ws.bus)
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
			ws.logger.Debugf(
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

func (ws *weatherStationImpl) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// create a new line instance
	line := charts.NewLine()
	samples := ws.Storage.GetEvents(time.Hour * 5)
	xTime := make([]time.Time, len(samples))
	yPressure := make([]opts.LineData, len(samples))
	var maxY int64
	var minY int64
	if len(samples) > 0 {
		maxY = samples[0].Pressure
		minY = samples[0].Pressure
	}
	for i, sample := range samples {
		xTime[i] = sample.Time
		yPressure[i] = opts.LineData{Value: tokHPa(sample.Pressure)}
		if maxY < sample.Pressure {
			maxY = sample.Pressure
		}
		if minY > sample.Pressure {
			minY = sample.Pressure
		}
	}
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithDataZoomOpts(opts.DataZoom{}),
		charts.WithTitleOpts(opts.Title{
			Title: "Pressure graph",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Min: float64(minY) - 0.005,
			Max: float64(maxY) + 0.005,
		}))

	//var symbol string
	//for i, sample := range samples {
	//	if sample.HeaterState {
	//		symbol = "circle"
	//	} else {
	//		symbol = "diamond"
	//	}
	//	xTime[i] = sample.Time.Unix()
	//	yPressure[i] = opts.LineData{Value: sample.Temperature, Symbol: symbol}
	//}
	line.SetXAxis(xTime).
		AddSeries("Pressure", yPressure)
	err := line.Render(w)
	ws.logger.Infof("build graph based on %ws last metrics", len(samples))
	if err != nil {
		ws.logger.Infof("Unable to render graph. %v", err.Error())
	}
}

// NewWeatherStation return a new instance of a WeatherStation daemon
func NewWeatherStation() weather_station.WeatherStation {
	return &weatherStationImpl{}
}

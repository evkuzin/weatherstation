package impl

import (
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"os"
	"periph.io/x/conn/v3/physic"
	"time"
)

func (ws *weatherStationImpl) telegramStart() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := ws.tg.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			ws.logger.Infof("[%s]: '%s'", update.Message.From.UserName, update.Message.Text)
			switch update.Message.Text {
			case "/avg":
				avg12h, err := ws.Storage.GetAvg(time.Hour * 12)
				if err != nil {
					ws.logger.Warnf("cannot get average: %s", err)
				}
				avg6h, err := ws.Storage.GetAvg(time.Hour * 6)
				if err != nil {
					ws.logger.Warnf("cannot get average: %s", err)
				}
				avg1h, err := ws.Storage.GetAvg(time.Hour)
				if err != nil {
					ws.logger.Warnf("cannot get average: %s", err)
				}
				avg1m, err := ws.Storage.GetAvg(time.Minute)
				if err != nil {
					ws.logger.Warnf("cannot get average: %s", err)
				}

				msgText := fmt.Sprintf("12h avg: %s\n6h avg: %s\n1h avg: %s\nCurrent: %s\n",
					physic.Pressure(avg12h),
					physic.Pressure(avg6h),
					physic.Pressure(avg1h),
					physic.Pressure(avg1m),
				)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				msg.Text = msgText
				msg.ReplyToMessageID = update.Message.MessageID

				_, err = ws.tg.Send(msg)
				if err != nil {
					ws.logger.Warnf("error: %s", err)
				}
			case "/graph":
				file, err := ioutil.TempFile("/tmp", "weather-station-*.html")
				if err != nil {
					ws.logger.Warnf("cannot write temporary htmlFile: %s", err)
				}
				ws.createGraph(file)
				ws.logger.Infof("htmlFile %s with metrics ready to send to %s", file.Name(), update.Message.From.UserName)
				msg := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FilePath(file.Name()))
				_, err = ws.tg.Send(msg)
				if err != nil {
					ws.logger.Warnf("error: %s", err)
				}
				err = os.Remove(file.Name())
				if err != nil {
					ws.logger.Warnf("cannot remove temp htmlFile: %s", err)
				}
			case "/pdf":
				htmlFile, err := ioutil.TempFile("/tmp", "weather-station-*.html")
				if err != nil {
					ws.logger.Warnf("cannot write temporary htmlFile: %s", err)
				}
				pdfFile, err := ioutil.TempFile("/tmp", "weather-station-*.pdf")
				if err != nil {
					ws.logger.Warnf("cannot write temporary pdfFile: %s", err)
				}
				ws.createGraph(htmlFile)
				pdfg, err := wkhtmltopdf.NewPDFGenerator()
				if err != nil {
					ws.logger.Warnf("cannot create new pdf generator: %s", err)
				}

				// Set global options
				pdfg.Dpi.Set(300)
				pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)

				// Create a new input page from an URL
				page := wkhtmltopdf.NewPage(htmlFile.Name())

				page.Zoom.Set(0.95)

				pdfg.AddPage(page)

				err = pdfg.Create()
				if err != nil {
					ws.logger.Warnf("cannot create new pdf: %s", err)
				}

				// Write buffer contents to htmlFile on disk
				err = pdfg.WriteFile(pdfFile.Name())
				if err != nil {
					ws.logger.Warnf("cannot write to new pdf: %s", err)
				}
				ws.logger.Infof("pdfFile %s with metrics ready to send to %s", pdfFile.Name(), update.Message.From.UserName)
				msg := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FilePath(pdfFile.Name()))
				_, err = ws.tg.Send(msg)
				if err != nil {
					ws.logger.Warnf("error: %s", err)
				}
				err = os.Remove(htmlFile.Name())
				if err != nil {
					ws.logger.Warnf("cannot remove temp htmlFile: %s", err)
				}
				err = os.Remove(pdfFile.Name())
				if err != nil {
					ws.logger.Warnf("cannot remove temp htmlFile: %s", err)
				}
			case "/png":
				pngFile, err := ioutil.TempFile("/tmp", "weather-station-*.png")
				if err != nil {
					ws.logger.Warnf("cannot write temporary png: %s", err)
				}

				ws.createPngGraph(pngFile)
				ws.logger.Infof("png %s with metrics ready to send to %s", pngFile.Name(), update.Message.From.UserName)
				msg := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FilePath(pngFile.Name()))
				_, err = ws.tg.Send(msg)
				if err != nil {
					ws.logger.Warnf("error: %s", err)
				}
				err = os.Remove(pngFile.Name())
				if err != nil {
					ws.logger.Warnf("cannot remove temp pngFile: %s", err)
				}
			default:
				msgText := "Unsupported.\nPlease choose any command from the menu."
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				msg.Text = msgText
				msg.ReplyToMessageID = update.Message.MessageID

				_, err := ws.tg.Send(msg)
				if err != nil {
					ws.logger.Warnf("error: %s", err)
				}
			}

		}
	}
}

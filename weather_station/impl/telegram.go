package impl

import (
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"os"
	"periph.io/x/conn/v3/physic"
	"time"
)

var buttons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("avg stats"),
		tgbotapi.NewKeyboardButton("graph html"),
		tgbotapi.NewKeyboardButton("graph pdf"),
	),
)

func (ws *weatherStationImpl) telegramStart() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := ws.tg.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			ws.logger.Infof("[%s]: '%s'", update.Message.From.UserName, update.Message.Text)
			switch update.Message.Text {
			case "avg stats":
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
				msg.ReplyMarkup = buttons

				_, err = ws.tg.Send(msg)
				if err != nil {
					ws.logger.Warnf("error: %s", err)
				}
			case "graph html":
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
			case "graph pdf":
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
					log.Fatal(err)
				}

				// Set global options
				pdfg.Dpi.Set(300)
				pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)
				pdfg.Grayscale.Set(true)

				// Create a new input page from an URL
				page := wkhtmltopdf.NewPage(htmlFile.Name())

				page.Zoom.Set(0.95)

				pdfg.AddPage(page)

				err = pdfg.Create()
				if err != nil {
					log.Fatal(err)
				}

				// Write buffer contents to htmlFile on disk
				err = pdfg.WriteFile(pdfFile.Name())
				if err != nil {
					log.Fatal(err)
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
			default:
				msgText := "Unsupported.\nPlease press any button."
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				msg.Text = msgText
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ReplyMarkup = buttons

				_, err := ws.tg.Send(msg)
				if err != nil {
					ws.logger.Warnf("error: %s", err)
				}
			}

		}
	}
}

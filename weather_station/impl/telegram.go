package impl

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"periph.io/x/conn/v3/physic"
	"time"
)

var buttons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("avg stats"),
	),
)

func (ws *weatherStationImpl) telegramStart() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := ws.tg.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			ws.logger.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
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
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ReplyMarkup = buttons

			_, err = ws.tg.Send(msg)
			if err != nil {
				ws.logger.Warnf("error: %s", err)
			}
		}
	}
}

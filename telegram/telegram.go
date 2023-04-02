package telegram

import (
	"log"
	"time"

	"github.com/ashwinath/money-tracker-telegram/processor"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram struct {
	bot              *tgbotapi.BotAPI
	allowedUser      string
	processorManager *processor.ProcessorManager
}

func New(
	apiKey string,
	isDebug bool,
	allowedUser string,
	processorManager *processor.ProcessorManager,
) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(apiKey)

	if err != nil {
		return nil, err
	}

	bot.Debug = isDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Telegram{
		bot:              bot,
		allowedUser:      allowedUser,
		processorManager: processorManager,
	}, nil
}

func (t *Telegram) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message.From.UserName == t.allowedUser && update.Message != nil { // If we got a message
			log.Printf("[%s to bot] %s", update.Message.From.UserName, update.Message.Text)

			reply := t.processorManager.ProcessMessage(update.Message.Text, time.Unix(int64(update.Message.Date), 0))
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, *reply)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = "Markdown"
			log.Printf("[bot to %s] %s", update.Message.From.UserName, *reply)

			_, err := t.bot.Send(msg)
			if err != nil {
				log.Printf("[bot to %s] error: %s", update.Message.From.UserName, err)
			}
		}
	}
}

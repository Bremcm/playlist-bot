package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	builder PlaylistBuilder
}

func New(token string, builder PlaylistBuilder) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	log.Printf("autorized as @%s", api.Self.UserName)
	return &Bot{api: api, builder: builder}, nil
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	log.Printf("[%s], %s", msg.From.UserName, msg.Text)

	var reply string
	switch msg.Command() {
	case "start":
		reply = "Привет! Я соберу тебе плейлист.\n" +
			"Пришли от 1 до 5 треков в формате «Исполнитель — Название», " +
			"по одному в сообщении. Когда закончишь — отправь /done."
	default:
		reply = "ты написал: " + msg.Text
	}

	out := tgbotapi.NewMessage(msg.Chat.ID, reply)
	if _, err := b.api.Send(out); err != nil {
		log.Printf("send message: %v", err)
	}
}

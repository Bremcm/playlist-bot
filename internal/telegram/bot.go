package telegram

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Bremcm/playlist-bot/internal/session"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	builder PlaylistBuilder
	session *session.Store
}

func New(token string, builder PlaylistBuilder, session *session.Store) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	log.Printf("autorized as @%s", api.Self.UserName)
	return &Bot{api: api, builder: builder, session: session}, nil
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
	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	chatID := msg.Chat.ID

	switch msg.Command() {
	case "start":
		b.reply(chatID, "Привет! Я соберу тебе плейлист.\n"+
			"Пришли от 1 до 5 треков в формате «Исполнитель — Название», "+
			"по одному в сообщении. Когда закончишь — отправь /done.")
		return

	case "done":
		b.handleDone(chatID)
		return
	}

	track, err := parseTrack(msg.Text)
	if err != nil {
		b.reply(chatID, "Не понял трек. Пришли в формате «Исполнитель — Название», например:\nMadonna — Frozen")
		return
	}

	count := b.session.Add(chatID, track)
	if count >= 5 {
		b.reply(chatID, "Принял. У тебя 5 треков — это максимум. Отправляю /done автоматически…")
		b.handleDone(chatID)
		return
	}
	b.reply(chatID, "Принял: "+track.Artist+" — "+track.Name+". Всего треков: "+itoa(count)+". Ещё? Или /done.")
}

func (b *Bot) handleDone(chatID int64) {
	seeds := b.session.Get(chatID)
	if len(seeds) == 0 {
		b.reply(chatID, "Ты ещё не прислал ни одного трека. Пришли хотя бы один в формате «Исполнитель — Название».")
		return
	}
	b.reply(chatID, "Собираю плейлист…")

	limit := playlistSize(len(seeds))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := b.builder.Build(ctx, seeds, limit)
	if err != nil {
		b.reply(chatID, "Что-то пошло не так при сборке плейлиста. Попробуй ещё раз.")
		log.Printf("build error: %v", err)
		return
	}

	b.session.Clear(chatID)

	var sb strings.Builder
	sb.WriteString("🎵 Твой плейлист:\n\n")
	for i, t := range result.Playlist {
		sb.WriteString(itoa(i+1) + ". " + t.Artist + " — " + t.Name + "\n")
	}
	if len(result.Failed) > 0 {
		sb.WriteString("\nПо этим трекам похожих не нашлось:\n")
		for _, t := range result.Failed {
			sb.WriteString("• " + t.Artist + " — " + t.Name + "\n")
		}
	}
	b.reply(chatID, sb.String())
}

func (b *Bot) reply(chatID int64, text string) {
	out := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(out); err != nil {
		log.Printf("send message: %v", err)
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

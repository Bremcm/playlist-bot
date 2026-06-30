package telegram

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Bremcm/playlist-bot/internal/models"
	"github.com/Bremcm/playlist-bot/internal/session"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	builder  PlaylistBuilder
	searcher TrackSearcher
	sessions *session.Store
}

func New(token string, builder PlaylistBuilder, searcher TrackSearcher, sessions *session.Store) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	log.Printf("autorized as @%s", api.Self.UserName)
	return &Bot{api: api, builder: builder, searcher: searcher, sessions: sessions}, nil
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

	parsed, err := parseTrack(msg.Text)
	if err != nil {
		b.reply(chatID, "Не понял трек. Пришли в формате «Исполнитель — Название», например:\nMadonna — Frozen")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	track, corrected, ok := b.resolveTrack(ctx, parsed)
	if !ok {
		b.reply(chatID, "Не уверен, что нашёл такой трек: "+parsed.Artist+" — "+parsed.Name+
			".\nПроверь название или пришли по-другому.")
		return
	}

	count := b.sessions.Add(chatID, track)

	prefix := "Принял: "
	if corrected {
		prefix = "Распознал как: "
	}

	if count >= 5 {
		b.reply(chatID, prefix+track.Artist+" — "+track.Name+
			"\nУ тебя 5 треков — это максимум. Собираю плейлист…")
		b.handleDone(chatID)
		return
	}
	b.reply(chatID, prefix+track.Artist+" — "+track.Name+
		". Всего треков: "+itoa(count)+". Ещё? Или /done.")
}

func (b *Bot) handleDone(chatID int64) {
	seeds := b.sessions.Get(chatID)
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

	b.sessions.Clear(chatID)

	var sb strings.Builder
	sb.WriteString("🎵 Твой плейлист:\n\n")
	for i, t := range result.Playlist {
		sb.WriteString(itoa(i + 1))
		sb.WriteString(". ")
		sb.WriteString(t.Artist)
		sb.WriteString(" — ")
		sb.WriteString(t.Name)
		sb.WriteString("\n")
	}
	if len(result.Failed) > 0 {
		sb.WriteString("\nПо этим трекам похожих не нашлось:\n")
		for _, t := range result.Failed {
			sb.WriteString("• ")
			sb.WriteString(t.Artist)
			sb.WriteString(" — ")
			sb.WriteString(t.Name)
			sb.WriteString("\n")
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

func (b *Bot) resolveTrack(ctx context.Context, parsed models.Track) (models.Track, bool, bool) {
	query := parsed.Artist + " " + parsed.Name

	results, err := b.searcher.Search(ctx, query, 5)
	if err != nil {
		return parsed, false, true
	}
	if len(results) == 0 {
		return parsed, false, false
	}

	best := results[0]

	if isCloseTrack(parsed, best) {
		corrected := best.Artist != parsed.Artist || best.Name != parsed.Name
		return best, corrected, true
	}

	return parsed, false, false
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

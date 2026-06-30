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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	track, corrected, ok := b.resolveTrack(ctx, parsed)
	if !ok {
		b.reply(chatID, "Не уверен, что нашёл такой трек: "+parsed.Artist+" — "+parsed.Name+
			".\nПроверь название или пришли по-другому.")
		return
	}

	candidates, err := b.builder.FetchCandidates(ctx, track)
	if err != nil {
		b.reply(chatID, "Не получилось проверить трек, попробуй ещё раз.")
		log.Printf("fetch candidates: %v", err)
		return
	}

	count := b.sessions.Add(chatID, session.Entry{
		Seed:       track,
		Candidates: candidates,
	})

	prefix := "Принял: "
	if corrected {
		prefix = "Распознал как: "
	}
	line := prefix + track.Artist + " — " + track.Name

	if len(candidates) == 0 {
		line += "\n⚠️ У Last.fm нет похожих на этот трек — он не попадёт в плейлист."
	}

	if count >= 5 {
		b.reply(chatID, line+"\nУ тебя 5 треков — это максимум. Собираю плейлист…")
		b.handleDone(chatID)
		return
	}
	b.reply(chatID, line+"\nВсего треков: "+itoa(count)+". Ещё? Или /done.")
}

func (b *Bot) handleDone(chatID int64) {
	entries := b.sessions.Entries(chatID)
	if len(entries) == 0 {
		b.reply(chatID, "Ты ещё не прислал ни одного трека. Пришли хотя бы один в формате «Исполнитель — Название».")
		return
	}

	var seeds []models.Track
	var allCandidates []models.Candidate
	var failed []models.Track

	for _, e := range entries {
		seeds = append(seeds, e.Seed)
		if len(e.Candidates) == 0 {
			failed = append(failed, e.Seed)
			continue
		}
		allCandidates = append(allCandidates, e.Candidates...)
	}

	limit := playlistSize(len(seeds))
	result := b.builder.Rank(seeds, allCandidates, limit)

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
	if len(failed) > 0 {
		sb.WriteString("\nПо этим трекам похожих не нашлось:\n")
		for _, t := range failed {
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
		best.Name = stripArtistPrefix(best.Name, best.Artist)
		corrected := best.Artist != parsed.Artist || best.Name != parsed.Name
		return best, corrected, true
	}

	return parsed, false, false
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

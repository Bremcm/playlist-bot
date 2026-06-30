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
		switch {
		case update.Message != nil:
			b.handleMessage(update.Message)
		case update.CallbackQuery != nil:
			b.handleCallback(update.CallbackQuery)
		}
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

	res := b.resolveTrack(ctx, parsed)

	switch res.outcome {
	case outcomeUnknown:
		b.reply(chatID, "Не уверен, что нашёл такой трек: "+parsed.Artist+" — "+parsed.Name+
			".\nПроверь название или пришли по-другому.")
		return

	case outcomeChoose:
		b.sessions.SetPending(chatID, res.options)
		b.sendOptions(chatID, parsed, res.options)
		return
	}

	b.acceptTrack(chatID, res.track, res.corrected)
}

func (b *Bot) handleCallback(cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	data := cq.Data

	b.api.Request(tgbotapi.NewCallback(cq.ID, ""))

	b.removeKeyboard(cq.Message)

	const prefix = "choose:"
	if !strings.HasPrefix(data, prefix) {
		return
	}

	arg := strings.TrimPrefix(data, prefix)

	if arg == "cancel" {
		b.sessions.ClearPending(chatID)
		b.reply(chatID, "Хорошо, пропустим этот трек. Пришли другой или /done.")
		return
	}

	index, err := strconv.Atoi(arg)
	if err != nil {
		return
	}

	track, ok := b.sessions.TakePending(chatID, index)
	if !ok {
		b.reply(chatID, "Этот выбор уже неактуален. Пришли трек заново.")
		return
	}
	b.acceptTrack(chatID, track, true)
}

func (b *Bot) acceptTrack(chatID int64, track models.Track, corrected bool) {
	if b.sessions.HasSeed(chatID, track) {
		b.reply(chatID, "Этот трек уже в списке: "+track.Artist+" — "+track.Name+". Пришли другой или /done.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

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

func (b *Bot) removeKeyboard(msg *tgbotapi.Message) {
	empty := tgbotapi.NewEditMessageReplyMarkup(
		msg.Chat.ID,
		msg.MessageID,
		tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}},
	)
	b.api.Request(empty)
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

func (b *Bot) sendOptions(chatID int64, parsed models.Track, options []models.Track) {
	text := "Не нашёл точно «" + parsed.Artist + " — " + parsed.Name + "». Может, ты имел в виду:"

	var rows [][]tgbotapi.InlineKeyboardButton
	for i, opt := range options {
		label := opt.Artist + " — " + opt.Name
		data := "choose:" + itoa(i)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	cancelBtn := tgbotapi.NewInlineKeyboardButtonData("Ничего из этого", "choose:cancel")
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(cancelBtn))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	out := tgbotapi.NewMessage(chatID, text)
	out.ReplyMarkup = keyboard
	if _, err := b.api.Send(out); err != nil {
		log.Printf("send options: %v", err)
	}
}

func (b *Bot) reply(chatID int64, text string) {
	out := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(out); err != nil {
		log.Printf("send message: %v", err)
	}
}

type resolveOutcome int

const (
	outcomeAccept resolveOutcome = iota
	outcomeChoose
	outcomeUnknown
)

type resolveResult struct {
	outcome   resolveOutcome
	track     models.Track
	corrected bool
	options   []models.Track
}

func (b *Bot) resolveTrack(ctx context.Context, parsed models.Track) resolveResult {
	query := parsed.Artist + " " + parsed.Name

	results, err := b.searcher.Search(ctx, query, 5)
	if err != nil {
		return resolveResult{outcome: outcomeAccept, track: parsed, corrected: false}
	}
	if len(results) == 0 {
		return resolveResult{outcome: outcomeUnknown}
	}

	best := results[0]

	if isCloseTrack(parsed, best) {
		best.Name = stripArtistPrefix(best.Name, best.Artist)
		corrected := best.Artist != parsed.Artist || best.Name != parsed.Name
		return resolveResult{outcome: outcomeAccept, track: best, corrected: corrected}
	}

	options := make([]models.Track, 0, 3)
	for i, r := range results {
		if i >= 3 {
			break
		}
		r.Name = stripArtistPrefix(r.Name, r.Artist)
		options = append(options, r)
	}
	return resolveResult{outcome: outcomeChoose, options: options}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

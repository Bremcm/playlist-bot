package main

import (
	"log"
	"os"

	"github.com/Bremcm/playlist-bot/internal/lastfm"
	"github.com/Bremcm/playlist-bot/internal/recommender"
	"github.com/Bremcm/playlist-bot/internal/session"
	"github.com/Bremcm/playlist-bot/internal/telegram"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, relying on real env vars")
	}
	lastfmKey := os.Getenv("LASTFM_API_KEY")
	if lastfmKey == "" {
		log.Fatal("LASTFM_API_KEY is not set")
	}
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	fetcher := lastfm.New(lastfmKey)
	rec := recommender.New(fetcher)
	session := session.New()

	bot, err := telegram.New(token, rec, session)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}
	bot.Run()
}

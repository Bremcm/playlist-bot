package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/Bremcm/playlist-bot/internal/lastfm"
	"github.com/Bremcm/playlist-bot/internal/models"
	"github.com/Bremcm/playlist-bot/internal/recommender"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, relying on real env vars")
	}

	apiKey := os.Getenv("LASTFM_API_KEY")
	if apiKey == "" {
		log.Fatal("LASTFM_API_KEY is not set")
	}

	fetcher := lastfm.New(apiKey)
	rec := recommender.New(fetcher)

	seeds := []models.Track{
		{Artist: "Cher", Name: "Believe"},
		{Artist: "Madonna", Name: "Frozen"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := rec.Build(ctx, seeds, 10)
	if err != nil {
		log.Fatalf("build playlist: %v", err)
	}

	fmt.Println("Рекомендованный плейлист:")
	for i, t := range result.Playlist {
		fmt.Printf("%2d. %s — %s\n", i+1, t.Artist, t.Name)
	}

	if len(result.Failed) > 0 {
		fmt.Println("\nПо этим трекам похожих не нашлось:")
		for _, t := range result.Failed {
			fmt.Printf("  - %s — %s\n", t.Artist, t.Name)
		}
	}
}

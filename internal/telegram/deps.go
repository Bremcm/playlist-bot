package telegram

import (
	"context"

	"github.com/Bremcm/playlist-bot/internal/models"
	"github.com/Bremcm/playlist-bot/internal/recommender"
)

type PlaylistBuilder interface {
	Build(ctx context.Context, seeds []models.Track, limit int) (recommender.BuildResult, error)
}

type TrackSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]models.Track, error)
}

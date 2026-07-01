package telegram

import (
	"context"

	"github.com/Bremcm/playlist-bot/internal/models"
	"github.com/Bremcm/playlist-bot/internal/recommender"
	"github.com/Bremcm/playlist-bot/internal/storage"
)

type PlaylistBuilder interface {
	FetchCandidates(ctx context.Context, seed models.Track) ([]models.Candidate, error)
	Rank(seeds []models.Track, candidates []models.Candidate, limit int) recommender.BuildResult
}

type TrackSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]models.Track, error)
}

type HistoryStore interface {
	SaveRequest(ctx context.Context, chatID int64, seeds, playlist []models.Track) error
	History(ctx context.Context, chatID int64, limit int) ([]storage.HistoryItem, error)
}

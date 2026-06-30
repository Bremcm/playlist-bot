package telegram

import (
	"context"

	"github.com/Bremcm/playlist-bot/internal/models"
	"github.com/Bremcm/playlist-bot/internal/recommender"
)

type PlaylistBuilder interface {
	FetchCandidates(ctx context.Context, seed models.Track) ([]models.Candidate, error)
	Rank(seeds []models.Track, candidates []models.Candidate, limit int) recommender.BuildResult
}

type TrackSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]models.Track, error)
}

package recommender

import (
	"context"

	"github.com/Bremcm/playlist-bot/internal/models"
)

type BuildResult struct {
	Playlist []models.Track
	Failed   []models.Track
}

func (r *Recommender) FetchCandidates(ctx context.Context, seed models.Track) ([]models.Candidate, error) {
	return r.fetcher.GetSimilar(ctx, seed)
}

func (r *Recommender) Rank(seeds []models.Track, candidates []models.Candidate, limit int) BuildResult {
	return BuildResult{
		Playlist: rank(seeds, candidates, limit),
		Failed:   nil,
	}
}

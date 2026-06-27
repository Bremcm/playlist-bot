package recommender

import (
	"context"

	"github.com/Bremcm/playlist-bot/internal/models"
)

func (r *Recommender) Build(ctx context.Context, seeds []models.Track, limit int) ([]models.Track, error) {
	var all []models.Candidate

	for _, seed := range seeds {
		candidates, err := r.fetcher.GetSimilar(ctx, seed)
		if err != nil {
			return nil, err
		}
		all = append(all, candidates...)
	}

	return rank(seeds, all, limit), nil
}

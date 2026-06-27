package recommender

import (
	"context"

	"github.com/Bremcm/playlist-bot/internal/models"
)

type BuildResult struct {
	Playlist []models.Track
	Failed   []models.Track
}

func (r *Recommender) Build(ctx context.Context, seeds []models.Track, limit int) (BuildResult, error) {
	var all []models.Candidate
	var failed []models.Track

	for _, seed := range seeds {
		candidates, err := r.fetcher.GetSimilar(ctx, seed)
		if err != nil {
			failed = append(failed, seed)
			continue
		}
		if len(candidates) == 0 {
			failed = append(failed, seed)
			continue
		}
		all = append(all, candidates...)
	}

	return BuildResult{
		Playlist: rank(seeds, all, limit),
		Failed:   failed,
	}, nil
}

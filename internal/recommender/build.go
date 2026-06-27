package recommender

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/Bremcm/playlist-bot/internal/models"
)

type BuildResult struct {
	Playlist []models.Track
	Failed   []models.Track
}

func (r *Recommender) Build(ctx context.Context, seeds []models.Track, limit int) (BuildResult, error) {
	var (
		mu     sync.Mutex
		all    []models.Candidate
		failed []models.Track
	)

	g, ctx := errgroup.WithContext(ctx)

	for _, seed := range seeds {
		g.Go(func() error {
			candidates, err := r.fetcher.GetSimilar(ctx, seed)

			mu.Lock()
			defer mu.Unlock()

			if err != nil || len(candidates) == 0 {
				failed = append(failed, seed)
				return nil
			}
			all = append(all, candidates...)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return BuildResult{}, err
	}

	return BuildResult{
		Playlist: rank(seeds, all, limit),
		Failed:   failed,
	}, nil
}

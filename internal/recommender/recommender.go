package recommender

import (
	"context"

	"github.com/Bremcm/playlist-bot/internal/models"
)

type SimilarFetcher interface {
	GetSimilar(ctx context.Context, seed models.Track) ([]models.Candidate, error)
}

type Recommender struct {
	fetcher SimilarFetcher
}

func New(fetcher SimilarFetcher) *Recommender {
	return &Recommender{fetcher: fetcher}
}

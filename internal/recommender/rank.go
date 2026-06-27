package recommender

import (
	"sort"

	"github.com/Bremcm/playlist-bot/internal/models"
)

func rank(seeds []models.Track, candidates []models.Candidate, limit int) []models.Track {
	seedSet := make(map[models.Track]bool, len(seeds))
	for _, s := range seeds {
		seedSet[s] = true
	}

	scores := make(map[models.Track]float64)
	for _, c := range candidates {
		if seedSet[c.Track] {
			continue
		}
		scores[c.Track] += c.Match
	}

	type scored struct {
		track models.Track
		score float64
	}
	ranked := make([]scored, 0, len(scores))
	for t, s := range scores {
		ranked = append(ranked, scored{track: t, score: s})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score != ranked[j].score {
			return ranked[i].score > ranked[j].score
		}
		if ranked[i].track.Artist != ranked[j].track.Artist {
			return ranked[i].track.Artist < ranked[j].track.Artist
		}
		return ranked[i].track.Name < ranked[j].track.Name
	})

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	result := make([]models.Track, len(ranked))
	for i, r := range ranked {
		result[i] = r.track
	}
	return result
}

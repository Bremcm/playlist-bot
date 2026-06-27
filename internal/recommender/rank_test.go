package recommender

import (
	"reflect"
	"testing"

	"github.com/Bremcm/playlist-bot/internal/models"
)

func TestRank(t *testing.T) {
	seedA := models.Track{Artist: "Artist A", Name: "Seed A"}
	seedB := models.Track{Artist: "Artist B", Name: "Seed B"}

	// Кандидат, похожий на ОБА сида (балл должен сложиться: 0.5 + 0.5 = 1.0).
	shared := models.Track{Artist: "Shared", Name: "Popular"}
	// Кандидат, похожий только на один сид (балл 0.9).
	single := models.Track{Artist: "Single", Name: "Lonely"}

	tests := []struct {
		name       string
		seeds      []models.Track
		candidates []models.Candidate
		limit      int
		want       []models.Track
	}{
		{
			name:  "общий кандидат обгоняет одиночного за счёт суммы баллов",
			seeds: []models.Track{seedA, seedB},
			candidates: []models.Candidate{
				{Track: shared, Match: 0.5},
				{Track: single, Match: 0.9},
				{Track: shared, Match: 0.5},
			},
			limit: 10,
			want:  []models.Track{shared, single},
		},
		{
			name:  "сиды исключаются из результата",
			seeds: []models.Track{seedA, seedB},
			candidates: []models.Candidate{
				{Track: seedA, Match: 1.0},
				{Track: single, Match: 0.3},
			},
			limit: 10,
			want:  []models.Track{single},
		},
		{
			name:  "limit обрезает результат",
			seeds: []models.Track{seedA},
			candidates: []models.Candidate{
				{Track: shared, Match: 0.9},
				{Track: single, Match: 0.8},
			},
			limit: 1,
			want:  []models.Track{shared},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rank(tt.seeds, tt.candidates, tt.limit)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rank() = %v, want %v", got, tt.want)
			}
		})
	}
}

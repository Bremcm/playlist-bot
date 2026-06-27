package lastfm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Bremcm/playlist-bot/internal/models"
)

const baseURL = "https://ws.audioscrobbler.com/2.0/"

// Client ходит в Last.fm. Конструктор возвращает конкретный *Client
// (accept interfaces, return structs).
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New создаёт клиент с заданным ключом и разумным таймаутом.
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetSimilar реализует recommender.SimilarFetcher: по одному треку
// возвращает список похожих кандидатов. Метод не знает про recommender —
// просто совпадает с его интерфейсом по сигнатуре. Так работает
// неявная имплементация интерфейсов в Go.
func (c *Client) GetSimilar(ctx context.Context, seed models.Track) ([]models.Candidate, error) {
	params := url.Values{}
	params.Set("method", "track.getsimilar")
	params.Set("artist", seed.Artist)
	params.Set("track", seed.Name)
	params.Set("api_key", c.apiKey)
	params.Set("format", "json")
	params.Set("autocorrect", "1")
	params.Set("limit", "30")

	reqURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("lastfm: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lastfm: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lastfm: unexpected status %d", resp.StatusCode)
	}

	var parsed similarResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("lastfm: decode response: %w", err)
	}

	candidates := make([]models.Candidate, 0, len(parsed.SimilarTracks.Track))
	for _, t := range parsed.SimilarTracks.Track {
		candidates = append(candidates, models.Candidate{
			Track: models.Track{
				Artist: t.Artist.Name,
				Name:   t.Name,
			},
			Match: t.Match,
		})
	}

	return candidates, nil
}

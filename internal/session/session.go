package session

import (
	"sync"

	"github.com/Bremcm/playlist-bot/internal/models"
)

type Store struct {
	mu     sync.Mutex
	tracks map[int64][]models.Track
}

func New() *Store {
	return &Store{
		tracks: make(map[int64][]models.Track),
	}
}

func (s *Store) Add(chatID int64, t models.Track) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tracks[chatID] = append(s.tracks[chatID], t)
	return len(s.tracks[chatID])
}

func (s *Store) Get(chatID int64) []models.Track {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.tracks[chatID]
}

func (s *Store) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tracks, chatID)
}

func (s *Store) Count(chatID int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.tracks[chatID])
}

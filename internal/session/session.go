package session

import (
	"sync"

	"github.com/Bremcm/playlist-bot/internal/models"
)

type Entry struct {
	Seed       models.Track
	Candidates []models.Candidate
}

type Store struct {
	mu      sync.Mutex
	entries map[int64][]Entry
	pending map[int64][]models.Track
}

func New() *Store {
	return &Store{
		entries: make(map[int64][]Entry),
		pending: make(map[int64][]models.Track),
	}
}

func (s *Store) Add(chatID int64, e Entry) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[chatID] = append(s.entries[chatID], e)
	return len(s.entries[chatID])
}

func (s *Store) Entries(chatID int64) []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.entries[chatID]
}

func (s *Store) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, chatID)
}

func (s *Store) Count(chatID int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.entries[chatID])
}

func (s *Store) SetPending(chatID int64, options []models.Track) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending[chatID] = options
}

func (s *Store) TakePending(chatID int64, index int) (models.Track, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	options, exists := s.pending[chatID]
	if !exists || index < 0 || index >= len(options) {
		return models.Track{}, false
	}

	chosen := options[index]
	delete(s.pending, chatID)
	return chosen, true
}

func (s *Store) ClearPending(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pending, chatID)
}

func (s *Store) HasSeed(chatID int64, seed models.Track) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.entries[chatID] {
		if e.Seed == seed {
			return true
		}
	}
	return false
}

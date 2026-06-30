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
}

func New() *Store {
	return &Store{
		entries: make(map[int64][]Entry),
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

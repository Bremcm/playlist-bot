package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Bremcm/playlist-bot/internal/models"
)

func (s *Store) SaveRequest(ctx context.Context, chatID int64, seeds, playlist []models.Track) error {
	seedsJSON, err := json.Marshal(seeds)
	if err != nil {
		return fmt.Errorf("storage: marshal seeds: %w", err)
	}
	playlistJSON, err := json.Marshal(playlist)
	if err != nil {
		return fmt.Errorf("storage: marshal playlist: %w", err)
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO requests (chat_id, seeds, playlist) VALUES ($1, $2, $3)`,
		chatID, seedsJSON, playlistJSON,
	)
	if err != nil {
		return fmt.Errorf("storage: insert request: %w", err)
	}
	return nil
}

type HistoryItem struct {
	Seeds    []models.Track
	Playlist []models.Track
}

func (s *Store) History(ctx context.Context, chatID int64, limit int) ([]HistoryItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT seeds, playlist FROM requests
		 WHERE chat_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		chatID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("storage: query history: %w", err)
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var seedsJSON, playlistJSON []byte
		if err := rows.Scan(&seedsJSON, &playlistJSON); err != nil {
			return nil, fmt.Errorf("storage: scan row: %w", err)
		}

		var item HistoryItem
		if err := json.Unmarshal(seedsJSON, &item.Seeds); err != nil {
			return nil, fmt.Errorf("storage: unmarshal seeds: %w", err)
		}
		if err := json.Unmarshal(playlistJSON, &item.Playlist); err != nil {
			return nil, fmt.Errorf("storage: unmarshal playlist: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: rows error: %w", err)
	}

	return items, nil
}

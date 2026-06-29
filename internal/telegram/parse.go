package telegram

import (
	"errors"
	"strings"

	"github.com/Bremcm/playlist-bot/internal/models"
)

var errInvalidFormat = errors.New("invalid track format")

func parseTrack(line string) (models.Track, error) {
	normalized := strings.NewReplacer("—", "-", "–", "-").Replace(line)

	parts := strings.SplitN(normalized, "-", 2)
	if len(parts) != 2 {
		return models.Track{}, errInvalidFormat
	}

	artist := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	if artist == "" || name == "" {
		return models.Track{}, errInvalidFormat
	}
	return models.Track{Artist: artist, Name: name}, nil
}

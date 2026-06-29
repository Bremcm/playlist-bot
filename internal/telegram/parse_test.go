package telegram

import (
	"testing"

	"github.com/Bremcm/playlist-bot/internal/models"
)

func TestParseTrack(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    models.Track
		wantErr bool
	}{
		{
			name:  "длинное тире",
			input: "Madonna — Frozen",
			want:  models.Track{Artist: "Madonna", Name: "Frozen"},
		},
		{
			name:  "обычный дефис",
			input: "Cher - Believe",
			want:  models.Track{Artist: "Cher", Name: "Believe"},
		},
		{
			name:  "лишние пробелы",
			input: "  ABBA   —   SOS  ",
			want:  models.Track{Artist: "ABBA", Name: "SOS"},
		},
		{
			name:  "дефис в названии трека",
			input: "AC/DC - Rock-N-Roll",
			want:  models.Track{Artist: "AC/DC", Name: "Rock-N-Roll"},
		},
		{
			name:    "нет разделителя",
			input:   "Madonna",
			wantErr: true,
		},
		{
			name:    "пустой исполнитель",
			input:   "- Frozen",
			wantErr: true,
		},
		{
			name:    "пустая строка",
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTrack(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTrack(%q): ожидалась ошибка, но её нет", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseTrack(%q): неожиданная ошибка: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("parseTrack(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

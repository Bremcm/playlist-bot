package telegram

import "testing"

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"madona", "madonna", 1},
		{"frozn", "frozen", 1},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
	}
	for _, tt := range tests {
		if got := levenshtein(tt.a, tt.b); got != tt.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsCloseMatch(t *testing.T) {
	tests := []struct {
		name          string
		guess, target string
		want          bool
	}{
		{"мелкая опечатка", "Madona", "Madonna", true},
		{"регистр не важен", "madonna", "Madonna", true},
		{"точное совпадение", "Frozen", "Frozen", true},
		{"совсем другое", "qwerty", "Madonna", false},
		{"далёкий ввод", "мадона ледяная", "Madonna Frozen", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCloseMatch(tt.guess, tt.target); got != tt.want {
				t.Errorf("isCloseMatch(%q, %q) = %v, want %v", tt.guess, tt.target, got, tt.want)
			}
		})
	}
}

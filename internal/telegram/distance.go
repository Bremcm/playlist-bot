package telegram

import (
	"strings"

	"github.com/Bremcm/playlist-bot/internal/models"
)

func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)

	la := len(ra)
	lb := len(rb)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	curr := make([]int, lb+1)
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min3(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func stripArtistPrefix(name, artist string) string {
	lowerName := strings.ToLower(name)
	lowerArtist := strings.ToLower(artist)

	prefix := lowerArtist + " - "

	if strings.HasPrefix(lowerName, prefix) {
		return strings.TrimSpace(name[len(prefix):])
	}
	return name
}

func isCloseTrack(a, b models.Track) bool {
	if !isCloseMatch(a.Artist, b.Artist) {
		return false
	}

	nameA := stripArtistPrefix(a.Name, a.Artist)
	nameB := stripArtistPrefix(b.Name, b.Artist)

	return isCloseMatch(nameA, nameB)
}

func isCloseMatch(guess, target string) bool {
	g := strings.ToLower(strings.TrimSpace(guess))
	t := strings.ToLower(strings.TrimSpace(target))

	if g == t {
		return true
	}

	dist := levenshtein(g, t)
	longer := len(([]rune(g)))
	if l := len([]rune(t)); l > longer {
		longer = l
	}
	if longer == 0 {
		return true
	}

	ratio := float64(dist) / float64(longer)
	return ratio <= 0.30

}

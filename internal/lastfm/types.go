package lastfm

type similarResponse struct {
	SimilarTracks similarTracks `json:"similartracks"`
}

type similarTracks struct {
	Track []apiTrack `json:"track"`
}

type apiTrack struct {
	Name   string    `json:"name"`
	Match  float64   `json:"match"`
	Artist apiArtist `json:"artist"`
}

type apiArtist struct {
	Name string `json:"name"`
}

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

type searchResponce struct {
	Results searchResults `json:"results"`
}
type searchResults struct {
	TrackMatches trackMatches `json:"trackmatches"`
}

type trackMatches struct {
	Track []searchTrack `json:"track"`
}
type searchTrack struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

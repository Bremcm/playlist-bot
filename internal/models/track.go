package models

type Track struct {
	Artist string
	Name   string
}

type Candidate struct {
	Track Track
	Match float64
}

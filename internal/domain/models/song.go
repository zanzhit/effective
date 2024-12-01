package models

type SongData struct {
	ID          int    `json:"id,omitempty" db:"id"`
	Group       string `json:"group,omitempty" db:"group"`
	Song        string `json:"song,omitempty" db:"song"`
	ReleaseDate string `json:"releaseDate,omitempty" db:"release_date"`
	Text        string `json:"text,omitempty" db:"lyrics"`
	Link        string `json:"link,omitempty" db:"link"`
}

type UpdateSongData struct {
	Group       *string `json:"group,omitempty"`
	Song        *string `json:"song,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
	Text        *string `json:"text,omitempty"`
	Link        *string `json:"link,omitempty"`
}

type FilterSongData struct {
	Group       *string `json:"group,omitempty"`
	Song        *string `json:"song,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
	Page        int
	PerPage     int
}

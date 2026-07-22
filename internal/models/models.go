package models
// Package models defines the data structures used to decode the
// Groupie Trackers API responses and the view models handed to templates.
type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

type LocationEntry struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

type LocationIndex struct {
    Index []LocationEntry `json:"index"`
}

type DateEntry struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

type DatesIndex struct {
	Index []DateEntry `json:"index"`
}

// RelationEntry links an artist ID to a map of location -> dates played there.
type RelationEntry struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// RelationsIndex wraps the "index" array the relation endpoint returns.
type RelationsIndex struct {
	Index []RelationEntry `json:"index"`
}

// ArtistDetail is the aggregated view model passed to the artist detail page.
type ArtistDetail struct {
	Artist         Artist
	DatesLocations map[string][]string
}

// SearchResult is a single suggestion returned by the live search endpoint.
type SearchResult struct {
	ArtistID int    `json:"artistId"`
	Label    string `json:"label"`
	Type     string `json:"type"`
}

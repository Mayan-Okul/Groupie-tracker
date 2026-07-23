// Package models defines the data structures used to decode the
// Groupie Trackers API responses and the view models handed to templates.
package models

// Artist represents a single band/artist as returned by the /api/artists endpoint.
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

// LocationEntry is one artist's list of concert locations, from /api/locations.
type LocationEntry struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

// LocationsIndex wraps the "index" array the locations endpoint returns.
type LocationsIndex struct {
	Index []LocationEntry `json:"index"`
}

// DateEntry is one artist's list of concert dates, from /api/dates.
type DateEntry struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// DatesIndex wraps the "index" array the dates endpoint returns.
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
	AllDates       []string // raw list from the /dates endpoint, shown separately from relation data
}

// SearchResult is a single suggestion returned by the live search endpoint.
type SearchResult struct {
	ArtistID int    `json:"artistId"`
	Label    string `json:"label"` // text shown to the user, e.g. "Queen (band)"
	Type     string `json:"type"`  // artist/band, member, location, first album, creation date
}
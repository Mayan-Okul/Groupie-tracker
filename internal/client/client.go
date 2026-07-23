// Package client is responsible for talking to the Groupie Trackers API
// and exposing the merged result through a thread-safe in-memory Store.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"Groupie-tracker/internal/models"
)

const (
	baseURL        = "https://groupietrackers.herokuapp.com/api"
	artistsURL     = baseURL + "/artists"
	locationsURL   = baseURL + "/locations"
	datesURL       = baseURL + "/dates"
	relationURL    = baseURL + "/relation"
	requestTimeout = 10 * time.Second
)

var httpClient = &http.Client{Timeout: requestTimeout}

// getJSON performs a GET request against url and decodes the JSON body into out.
func getJSON(url string, out interface{}) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, url, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}

func fetchArtists() ([]models.Artist, error) {
	var artists []models.Artist
	if err := getJSON(artistsURL, &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func fetchRelations() (models.RelationsIndex, error) {
	var rel models.RelationsIndex
	if err := getJSON(relationURL, &rel); err != nil {
		return models.RelationsIndex{}, err
	}
	return rel, nil
}

func fetchLocations() (models.LocationsIndex, error) {
	var locs models.LocationsIndex
	if err := getJSON(locationsURL, &locs); err != nil {
		return models.LocationsIndex{}, err
	}
	return locs, nil
}

func fetchDates() (models.DatesIndex, error) {
	var dates models.DatesIndex
	if err := getJSON(datesURL, &dates); err != nil {
		return models.DatesIndex{}, err
	}
	return dates, nil
}

// Store holds the fully-loaded, merged dataset in memory so every HTTP
// request served by this app is answered from RAM instead of hitting the
// upstream API each time. It is safe for concurrent reads/writes.
type Store struct {
	mu        sync.RWMutex
	artists   []models.Artist
	byID      map[int]models.Artist
	relations map[int]map[string][]string
	locations map[int][]string
	dates     map[int][]string
	loadedAt  time.Time
}

// NewStore builds an empty store. Call Refresh to populate it.
func NewStore() *Store {
	return &Store{
		byID:      make(map[int]models.Artist),
		relations: make(map[int]map[string][]string),
	}
}

// Refresh fetches all four API resources and atomically swaps them into the
// store. If any request fails, the store keeps serving its previous (stale)
// data rather than being left half-updated or empty.
func (s *Store) Refresh() error {
	artists, err := fetchArtists()
	if err != nil {
		return fmt.Errorf("fetch artists: %w", err)
	}
	relations, err := fetchRelations()
	if err != nil {
		return fmt.Errorf("fetch relations: %w", err)
	}
	locs, err := fetchLocations()
	if err != nil {
		return fmt.Errorf("fetch locations: %w", err)
	}
	dates, err := fetchDates()
	if err != nil {
		return fmt.Errorf("fetch dates: %w", err)
	}

	byID := make(map[int]models.Artist, len(artists))
	for _, a := range artists {
		byID[a.ID] = a
	}

	relByID := make(map[int]map[string][]string, len(relations.Index))
	for _, r := range relations.Index {
		relByID[r.ID] = r.DatesLocations
	}

	locByID := make(map[int][]string, len(locs.Index))
	for _, l := range locs.Index {
		locByID[l.ID] = l.Locations
	}

	dateByID := make(map[int][]string, len(dates.Index))
	for _, d := range dates.Index {
		dateByID[d.ID] = d.Dates
	}

	s.mu.Lock()
	s.artists = artists
	s.byID = byID
	s.relations = relByID
	s.locations = locByID
	s.dates = dateByID
	s.loadedAt = time.Now()
	s.mu.Unlock()
	return nil
}

// Ready reports whether the store has ever been successfully populated.
func (s *Store) Ready() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.artists) > 0
}

// Artists returns a copy of the artist slice.
func (s *Store) Artists() []models.Artist {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Artist, len(s.artists))
	copy(out, s.artists)
	return out
}

// ArtistByID returns a single artist and whether it was found.
func (s *Store) ArtistByID(id int) (models.Artist, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.byID[id]
	return a, ok
}

// RelationsFor returns the location->dates map for an artist ID.
func (s *Store) RelationsFor(id int) map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.relations[id]
}

// LocationsFor returns the raw list of concert locations for an artist ID.
func (s *Store) LocationsFor(id int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.locations[id]
}

// DatesFor returns the raw list of concert dates for an artist ID.
func (s *Store) DatesFor(id int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dates[id]
}
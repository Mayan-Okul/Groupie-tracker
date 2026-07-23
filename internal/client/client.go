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
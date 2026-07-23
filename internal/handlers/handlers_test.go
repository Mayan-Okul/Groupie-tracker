package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Groupie-tracker/internal/models"
)

// mockStore is a fixed, in-memory stand-in for client.Store so handler
// tests never touch the network.
type mockStore struct {
	ready     bool
	artists   []models.Artist
	relations map[int]map[string][]string
	locations map[int][]string
	dates     map[int][]string
}

func (m *mockStore) Ready() bool              { return m.ready }
func (m *mockStore) Artists() []models.Artist { return m.artists }
func (m *mockStore) ArtistByID(id int) (models.Artist, bool) {
	for _, a := range m.artists {
		if a.ID == id {
			return a, true
		}
	}
	return models.Artist{}, false
}
func (m *mockStore) RelationsFor(id int) map[string][]string { return m.relations[id] }
func (m *mockStore) LocationsFor(id int) []string             { return m.locations[id] }
func (m *mockStore) DatesFor(id int) []string                 { return m.dates[id] }

func newTestHandler(t *testing.T, ready bool) *Handler {
	t.Helper()
	store := &mockStore{
		ready: ready,
		artists: []models.Artist{
			{ID: 1, Name: "Queen", Image: "img.jpg", Members: []string{"Freddie Mercury", "Brian May"}, CreationDate: 1970, FirstAlbum: "01-07-1973"},
			{ID: 2, Name: "Nirvana", Image: "img2.jpg", Members: []string{"Kurt Cobain"}, CreationDate: 1987, FirstAlbum: "15-06-1989"},
		},
		relations: map[int]map[string][]string{
			1: {"london-uk": {"01-01-2020"}},
		},
		locations: map[int][]string{
			1: {"london-uk"},
		},
	}
	h, err := NewFromFiles(store,
		"../../templates/index.html",
		"../../templates/artist.html",
		"../../templates/error.html",
	)
	if err != nil {
		t.Fatalf("failed to build handler: %v", err)
	}
	return h
}

func TestIndex_OK(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.Index(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestIndex_UnknownPath(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	w := httptest.NewRecorder()

	h.Index(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestIndex_DataNotReady(t *testing.T) {
	h := newTestHandler(t, false)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.Index(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestArtistDetail_OK(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/artist?id=1", nil)
	w := httptest.NewRecorder()

	h.ArtistDetail(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestArtistDetail_NotFound(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/artist?id=999", nil)
	w := httptest.NewRecorder()

	h.ArtistDetail(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestArtistDetail_InvalidID(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/artist?id=abc", nil)
	w := httptest.NewRecorder()

	h.ArtistDetail(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSearch_MatchesArtistName(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=queen", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var results []models.SearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for 'queen'")
	}
	found := false
	for _, r := range results {
		if r.ArtistID == 1 && r.Type == "artist/band" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected a result matching Queen by name")
	}
}

func TestSearch_MatchesMember(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=cobain", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	var results []models.SearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if len(results) != 1 || results[0].ArtistID != 2 {
		t.Fatalf("expected exactly one result pointing at Nirvana, got %+v", results)
	}
}

func TestSearch_EmptyQueryReturnsEmptyList(t *testing.T) {
	h := newTestHandler(t, true)
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	var results []models.SearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results for empty query, got %d", len(results))
	}
}

func TestPrettyLocation(t *testing.T) {
	got := prettyLocation("san_francisco-usa")
	want := "San Francisco-usa"
	if got != want {
		t.Fatalf("prettyLocation() = %q, want %q", got, want)
	}
}
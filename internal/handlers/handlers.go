// Package handlers wires the in-memory data Store to HTTP endpoints:
// the artist grid, the artist detail page, and a live search API used
// by the client-side JavaScript event feature.
package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"Groupie-tracker/internal/models"
)

// DataStore is the subset of client.Store that handlers depend on.
// Defining it as an interface keeps this package testable without
// hitting the real network.
type DataStore interface {
	Ready() bool
	Artists() []models.Artist
	ArtistByID(id int) (models.Artist, bool)
	RelationsFor(id int) map[string][]string
	LocationsFor(id int) []string
	DatesFor(id int) []string
}

// Handler bundles the store and parsed templates needed to serve every route.
type Handler struct {
	store     DataStore
	templates *template.Template
}

// NewFromFiles parses the given template file paths and returns a Handler.
func NewFromFiles(store DataStore, patterns ...string) (*Handler, error) {
	tmpl, err := template.ParseFiles(patterns...)
	if err != nil {
		return nil, err
	}
	return &Handler{store: store, templates: tmpl}, nil
}

// Recover wraps a handler so that if it panics — a nil pointer, an index
// out of range, anything unexpected — the server catches it, logs it, and
// returns a 500 to that one request instead of crashing the whole process.
func (h *Handler) Recover(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered on %s %s: %v", r.Method, r.URL.Path, rec)
				h.renderError(w, http.StatusInternalServerError, "Something went wrong on our end.")
			}
		}()
		next(w, r)
	}
}

// Index renders the home page: a searchable grid of every artist.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.renderError(w, http.StatusNotFound, "Page not found.")
		return
	}
	if r.Method != http.MethodGet {
		h.renderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}
	if !h.store.Ready() {
		h.renderError(w, http.StatusServiceUnavailable, "Data is still loading, please refresh in a few seconds.")
		return
	}

	artists := h.store.Artists()
	sort.Slice(artists, func(i, j int) bool { return artists[i].Name < artists[j].Name })

	h.render(w, "index.html", map[string]interface{}{
		"Title":   "Groupie Trackers",
		"Artists": artists,
	})
}

// ArtistDetail renders the page for one artist, merging in their concert
// locations, dates, and the combined location->dates relation data.
func (h *Handler) ArtistDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.renderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		h.renderError(w, http.StatusBadRequest, "Invalid artist id.")
		return
	}

	artist, ok := h.store.ArtistByID(id)
	if !ok {
		h.renderError(w, http.StatusNotFound, "That artist doesn't exist.")
		return
	}

	detail := models.ArtistDetail{
		Artist:         artist,
		DatesLocations: h.store.RelationsFor(id),
		AllDates:       h.store.DatesFor(id),
	}

	h.render(w, "artist.html", map[string]interface{}{
		"Title":  artist.Name,
		"Detail": detail,
	})
}

// Search is the client-server "event" feature: the browser fires a request
// on every keystroke (debounced client-side), the server matches the query
// against artist/member/location/first-album/creation-date fields, and
// responds with JSON that the page renders without a reload.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	if !h.store.Ready() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "data still loading"})
		return
	}

	query := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
	results := []models.SearchResult{}

	if query != "" {
		for _, a := range h.store.Artists() {
			if strings.Contains(strings.ToLower(a.Name), query) {
				results = append(results, models.SearchResult{ArtistID: a.ID, Label: a.Name, Type: "artist/band"})
			}
			for _, m := range a.Members {
				if strings.Contains(strings.ToLower(m), query) {
					results = append(results, models.SearchResult{ArtistID: a.ID, Label: m + " — " + a.Name, Type: "member"})
				}
			}
			if strings.Contains(strings.ToLower(a.FirstAlbum), query) {
				results = append(results, models.SearchResult{ArtistID: a.ID, Label: a.FirstAlbum + " — " + a.Name, Type: "first album"})
			}
			if strings.Contains(strconv.Itoa(a.CreationDate), query) {
				results = append(results, models.SearchResult{ArtistID: a.ID, Label: strconv.Itoa(a.CreationDate) + " — " + a.Name, Type: "creation date"})
			}
			for _, loc := range h.store.LocationsFor(a.ID) {
				normalized := strings.ReplaceAll(strings.ToLower(loc), "_", " ")
				if strings.Contains(normalized, query) {
					results = append(results, models.SearchResult{ArtistID: a.ID, Label: prettyLocation(loc) + " — " + a.Name, Type: "location"})
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Label < results[j].Label })
	if len(results) > 25 {
		results = results[:25]
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("search encode error: %v", err)
	}
}

// prettyLocation turns the API's "san_francisco-usa" style strings into
// "San Francisco-usa" for display. strings.Title is deprecated, so this
// does the word-capitalization by hand.
func prettyLocation(loc string) string {
	loc = strings.ReplaceAll(loc, "_", " ")
	words := strings.Fields(loc)
	for i, w := range words {
		if w == "" {
			continue
		}
		r := []rune(w)
		r[0] = []rune(strings.ToUpper(string(r[0])))[0]
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

// render executes a named template and writes it to w. Template execution
// errors are logged and turned into a 500 rather than a half-written page.
func (h *Handler) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template %s error: %v", name, err)
		h.renderError(w, http.StatusInternalServerError, "Failed to render page.")
	}
}

// renderError writes a plain, dependency-free error page so it can never
// itself fail even if the main templates are broken.
func (h *Handler) renderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if h.templates != nil && h.templates.Lookup("error.html") != nil {
		err := h.templates.ExecuteTemplate(w, "error.html", map[string]interface{}{
			"Code":    status,
			"Message": message,
		})
		if err == nil {
			return
		}
	}
	w.Write([]byte("<h1>" + strconv.Itoa(status) + "</h1><p>" + message + "</p>"))
}
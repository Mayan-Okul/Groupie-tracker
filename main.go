// Command groupie-tracker runs the Groupie Trackers web server.
// It loads artist/location/date/relation data from the public API into
// memory, then serves an HTML site plus a small JSON search API.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"Groupie-tracker/internal/client"
	"Groupie-tracker/internal/handlers"
)

const (
	addr            = ":8080"
	refreshInterval = 30 * time.Minute
)

func main() {
	store := client.NewStore()

	log.Println("loading data from Groupie Trackers API...")
	if err := store.Refresh(); err != nil {
		// The site can still start (handlers report 503 until data loads),
		// but we make noise so it's obvious something's wrong.
		log.Printf("initial data load failed: %v", err)
	} else {
		log.Println("data loaded successfully")
	}

	// Keep the in-memory data reasonably fresh without blocking any request.
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := store.Refresh(); err != nil {
				log.Printf("background refresh failed: %v", err)
			}
		}
	}()

	h, err := handlers.NewFromFiles(store,
		"templates/index.html",
		"templates/artist.html",
		"templates/error.html",
	)
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Recover(h.Index))
	mux.HandleFunc("/artist", h.Recover(h.ArtistDetail))
	mux.HandleFunc("/api/search", h.Recover(h.Search))

	staticDir := http.Dir("static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticDir)))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("serving on http://localhost%s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Println("server stopped:", err)
		os.Exit(1)
	}
}
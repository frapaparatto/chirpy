package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/frapaparatto/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	const rootDir = "./app/"
	const port = "8080"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		// do something
		return
	}

	queries := database.New(db)

	cfg := &Config{
		dbQueries: queries,
	}

	mux := http.NewServeMux()

	var m http.Handler = http.HandlerFunc(cfg.metricsHandler)
	var r http.Handler = http.HandlerFunc(cfg.resetMetrics)
	var v http.Handler = http.HandlerFunc(validationHandler)
	var h http.Handler = http.HandlerFunc(healthHandler)

	// FileServer Handler
	f := http.StripPrefix("/app", http.FileServer(http.Dir(rootDir)))
	mux.Handle("/app/", cfg.middlewareMetricsInc(f))

	// API handler
	mux.Handle("GET /api/healthz", h)
	mux.Handle("POST /api/validate_chirp", v)

	// Admin handlers
	mux.Handle("GET /admin/metrics", m)
	mux.Handle("POST /admin/reset", r)

	s := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Serving files from %s on port: %s\n", rootDir, port)
	log.Fatal(s.ListenAndServe())
}

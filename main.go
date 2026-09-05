package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/frapaparatto/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	const rootDir = "./app/"
	const port = "8080"

	dbURL := os.Getenv("DB_URL")
	ptf := os.Getenv("PLATFORM")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(dbConn)

	cfg := &Config{
		db:       dbQueries,
		platform: ptf,
	}

	mux := http.NewServeMux()

	// FileServer Handler
	f := http.StripPrefix("/app", http.FileServer(http.Dir(rootDir)))
	mux.Handle("/app/", cfg.middlewareMetricsInc(f))

	// API handler
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /api/chirps", cfg.handleChirpList)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirp)
	mux.HandleFunc("POST /api/chirps", cfg.handleChirpCreation)
	mux.HandleFunc("POST /api/users", cfg.handleUserCreation)

	// Admin handlers
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)

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

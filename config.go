package main

import (
	"sync/atomic"

	"github.com/frapaparatto/chirpy/internal/database"
)

type Config struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

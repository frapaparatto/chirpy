package main

import (
	"sync/atomic"

	"github.com/frapaparatto/http-server-go/internal/database"
)

type Config struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

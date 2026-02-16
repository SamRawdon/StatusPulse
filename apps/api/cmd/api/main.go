package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"statuspulse/apps/api/internal/config"
	"statuspulse/apps/api/internal/db"
	httpapi "statuspulse/apps/api/internal/http"
	"statuspulse/apps/api/internal/store"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	s := &store.Store{DB: pool}
	api := &httpapi.API{Store: s}

	log.Printf("api listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, httpapi.Router(api, cfg.CorsOrigin)))
}

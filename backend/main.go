package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"media-sequencer/config"
	"media-sequencer/db"
	"media-sequencer/handlers"
	"media-sequencer/seed"
	"media-sequencer/services"
)

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var store db.Store
	if cfg.MongoURI != "" {
		mongoStore, err := db.NewMongoStore(ctx, cfg.MongoURI, cfg.MongoDBName)
		if err != nil {
			log.Printf("Failed to connect to MongoDB (%v). Falling back to MemoryStore.", err)
			store = db.NewMemoryStore()
		} else {
			log.Println("Connected successfully to MongoDB")
			store = mongoStore
		}
	} else {
		log.Println("MONGO_URI not provided. Using in-memory store.")
		store = db.NewMemoryStore()
	}

	// Seed initial windows A, B, C
	if err := seed.SeedInitialData(context.Background(), store); err != nil {
		log.Printf("Warning: failed to seed initial data: %v", err)
	} else {
		log.Println("Windows A, B, C seeded successfully")
	}

	syncManager := services.NewSyncManager(store)
	h := handlers.NewHandler(store, syncManager, cfg.SyncDurationSeconds)

	mux := http.NewServeMux()
	
	// Router pattern without method lock to allow CORS OPTIONS preflight
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/state", h.GetState)
	mux.HandleFunc("/api/windows/{id}/playlist", h.GetPlaylist)
	mux.HandleFunc("/api/windows/{id}/media", h.AddMedia)
	mux.HandleFunc("/api/sync", h.TriggerSync)

	addr := ":" + cfg.Port
	fmt.Printf("Server running on port %s...\n", cfg.Port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}

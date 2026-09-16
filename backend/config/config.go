package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	MongoURI            string
	MongoDBName         string
	SyncDurationSeconds int
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "media_sequencer"
	}

	syncDurStr := os.Getenv("SYNC_DURATION_SECONDS")
	syncDur := 5
	if syncDurStr != "" {
		if val, err := strconv.Atoi(syncDurStr); err == nil && val > 0 {
			syncDur = val
		}
	}

	return &Config{
		Port:                port,
		MongoURI:            mongoURI,
		MongoDBName:         dbName,
		SyncDurationSeconds: syncDur,
	}
}

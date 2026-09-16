package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                string
	MongoURI            string
	MongoDBName         string
	SyncDurationSeconds int
}

// loadDotEnv helper loads key-value pairs from .env file into environment if not already set
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func LoadConfig() *Config {
	// Try loading .env from current directory or parent
	loadDotEnv(".env")
	loadDotEnv("../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = os.Getenv("MONGODB_URI")
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = os.Getenv("DB_NAME")
	}
	if dbName == "" {
		dbName = "eva2"
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

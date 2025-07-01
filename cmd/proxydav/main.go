package main

import (
	"log"
	"os"

	"proxydav/internal/config"
	"proxydav/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Load file entries
	files, err := config.LoadFileEntries(cfg.ConfigFile)
	if err != nil {
		log.Fatalf("Failed to load file entries: %v", err)
	}

	if len(files) == 0 {
		log.Println("Warning: No files configured")
	} else {
		log.Printf("Loaded %d file entries from %s", len(files), cfg.ConfigFile)
	}

	// Create and start server
	srv := server.New(cfg, files)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}

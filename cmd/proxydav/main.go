package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"proxydav/internal/config"
	"proxydav/internal/server"
	"proxydav/pkg/types"
)

// Build information (set by linker flags during build)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Handle version flag
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	// Load configuration (this will parse all flags including version)
	cfg := config.Load()

	if showVersion {
		fmt.Printf("ProxyDAV %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Start with empty filesystem - files can only be added via API
	files := []types.FileEntry{}
	log.Println("Starting with empty filesystem. Use the REST API to add files.")

	// Create and start server
	srv := server.New(cfg, files)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}

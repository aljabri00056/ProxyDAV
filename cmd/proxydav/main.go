package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"proxydav/internal/config"
	"proxydav/internal/server"
)

// Build information (set by linker flags during build)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	cfg := config.Load()

	if showVersion {
		fmt.Printf("ProxyDAV %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Println("Starting. Files will be automatically loaded from database.")

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}

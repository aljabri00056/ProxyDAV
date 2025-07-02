package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"proxydav/internal/config"
	"proxydav/internal/server"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	fs := flag.NewFlagSet("proxydav", flag.ExitOnError)

	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "Show version information")

	cfg := config.Load(fs)

	if showVersion {
		fmt.Println()
		fmt.Println("🌐 ProxyDAV - Virtual WebDAV Server")
		fmt.Printf("📦 Version: %s\n", version)
		if commit != "unknown" {
			fmt.Printf("🔗 Commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("📅 Built: %s\n", date)
		}
		fmt.Println()
		os.Exit(0)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Configuration validation failed: %v", err)
	}

	log.Println("🚀 Starting ProxyDAV server...")

	for {
		srv, err := server.New(cfg)
		if err != nil {
			log.Fatalf("❌ Failed to create server: %v", err)
		}

		err = srv.Start()
		if err == nil {
			// Normal shutdown
			break
		} else if errors.Is(err, server.ErrRestart) {
			log.Println("🔄 Restarting server...")
			// Reload configuration for restart
			cfg = config.Reload()
			if err := cfg.Validate(); err != nil {
				log.Fatalf("❌ Configuration validation failed on restart: %v", err)
			}
			continue
		} else {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}
}

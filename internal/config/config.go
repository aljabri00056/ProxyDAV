package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration options for the ProxyDAV server
type Config struct {
	Port         int    `json:"port"`
	CacheTTL     int    `json:"cache_ttl_seconds"`
	UseRedirect  bool   `json:"use_redirect"`
	AuthEnabled  bool   `json:"auth_enabled"`
	AuthUser     string `json:"auth_user"`
	AuthPass     string `json:"auth_pass"`
	MaxCacheSize int    `json:"max_cache_size"`
	HealthPath   string `json:"health_path"`
}

// Load loads configuration from environment variables, command line flags, and defaults
func Load() *Config {
	config := &Config{
		Port:         8080,
		CacheTTL:     3600,
		UseRedirect:  false,
		AuthEnabled:  false,
		AuthUser:     "",
		AuthPass:     "",
		MaxCacheSize: 1000,
		HealthPath:   "/health",
	}

	// Parse command line flags
	flag.IntVar(&config.Port, "port", config.Port, "Port to listen on")
	flag.IntVar(&config.CacheTTL, "cache-ttl", config.CacheTTL, "Cache TTL in seconds")
	flag.BoolVar(&config.UseRedirect, "redirect", config.UseRedirect, "Use 302 redirects instead of proxying content")
	flag.BoolVar(&config.AuthEnabled, "auth", config.AuthEnabled, "Enable HTTP Basic authentication")
	flag.StringVar(&config.AuthUser, "user", config.AuthUser, "Username for authentication")
	flag.StringVar(&config.AuthPass, "pass", config.AuthPass, "Password for authentication")
	flag.IntVar(&config.MaxCacheSize, "max-cache", config.MaxCacheSize, "Maximum number of items in cache")
	flag.StringVar(&config.HealthPath, "health-path", config.HealthPath, "Path for health check endpoint")
	flag.Parse()

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}
	if ttl := os.Getenv("CACHE_TTL"); ttl != "" {
		if t, err := strconv.Atoi(ttl); err == nil {
			config.CacheTTL = t
		}
	}
	if redirect := os.Getenv("USE_REDIRECT"); redirect == "true" {
		config.UseRedirect = true
	}
	if auth := os.Getenv("AUTH_ENABLED"); auth == "true" {
		config.AuthEnabled = true
	}
	if user := os.Getenv("AUTH_USER"); user != "" {
		config.AuthUser = user
	}
	if pass := os.Getenv("AUTH_PASS"); pass != "" {
		config.AuthPass = pass
	}
	if maxCache := os.Getenv("MAX_CACHE_SIZE"); maxCache != "" {
		if m, err := strconv.Atoi(maxCache); err == nil {
			config.MaxCacheSize = m
		}
	}
	if healthPath := os.Getenv("HEALTH_PATH"); healthPath != "" {
		config.HealthPath = healthPath
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.CacheTTL < 0 {
		return fmt.Errorf("cache TTL cannot be negative")
	}
	if c.AuthEnabled && (c.AuthUser == "" || c.AuthPass == "") {
		return fmt.Errorf("authentication requires both username and password")
	}
	if c.MaxCacheSize < 0 {
		return fmt.Errorf("max cache size cannot be negative")
	}
	return nil
}

package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        int    `json:"port"`
	UseRedirect bool   `json:"use_redirect"`
	AuthEnabled bool   `json:"auth_enabled"`
	AuthUser    string `json:"auth_user"`
	AuthPass    string `json:"auth_pass"`
	DataDir     string `json:"data_dir"`
}

func Load() *Config {
	config := &Config{
		Port:        8080,
		UseRedirect: false,
		AuthEnabled: false,
		AuthUser:    "",
		AuthPass:    "",
		DataDir:     "./proxydavData",
	}

	flag.IntVar(&config.Port, "port", config.Port, "Port to listen on")
	flag.BoolVar(&config.UseRedirect, "redirect", config.UseRedirect, "Use 302 redirects instead of proxying content")
	flag.BoolVar(&config.AuthEnabled, "auth", config.AuthEnabled, "Enable HTTP Basic authentication")
	flag.StringVar(&config.AuthUser, "user", config.AuthUser, "Username for authentication")
	flag.StringVar(&config.AuthPass, "pass", config.AuthPass, "Password for authentication")
	flag.StringVar(&config.DataDir, "data-dir", config.DataDir, "Directory for persistent data storage")
	flag.Parse()

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
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
	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		config.DataDir = dataDir
	}

	return config
}

func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.AuthEnabled && (c.AuthUser == "" || c.AuthPass == "") {
		return fmt.Errorf("authentication requires both username and password")
	}
	if c.DataDir == "" {
		return fmt.Errorf("data directory cannot be empty")
	}
	return nil
}

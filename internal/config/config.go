package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ConfigUpdater interface {
	UpdateConfig(newConfig *Config) error
	GetConfig() *Config
}

type Config struct {
	Port        int    `json:"port"`
	UseRedirect bool   `json:"use_redirect"`
	AuthEnabled bool   `json:"auth_enabled"`
	AuthUser    string `json:"auth_user"`
	AuthPass    string `json:"auth_pass"`
	DataDir     string `json:"data_dir"`
}

func Load(fs *flag.FlagSet) *Config {
	config := &Config{
		Port:        8080,
		DataDir:     "./proxydavData",
		UseRedirect: false,
		AuthEnabled: false,
		AuthUser:    "",
		AuthPass:    "",
	}

	fs.IntVar(&config.Port, "port", config.Port, "Port to listen on")
	fs.StringVar(&config.DataDir, "data-dir", config.DataDir, "Directory for persistent data storage")
	fs.BoolVar(&config.UseRedirect, "redirect", config.UseRedirect, "Use 302 redirects instead of proxying content")
	fs.BoolVar(&config.AuthEnabled, "auth", config.AuthEnabled, "Enable HTTP Basic authentication")
	fs.StringVar(&config.AuthUser, "user", config.AuthUser, "Username for authentication")
	fs.StringVar(&config.AuthPass, "pass", config.AuthPass, "Password for authentication")
	fs.Parse(os.Args[1:])

	return loadFromEnv(config)
}

func Reload() *Config {
	config := &Config{
		Port:        8080,
		UseRedirect: false,
		AuthEnabled: false,
		AuthUser:    "",
		AuthPass:    "",
		DataDir:     "./proxydavData",
	}

	if f := flag.Lookup("port"); f != nil {
		if p, err := strconv.Atoi(f.Value.String()); err == nil {
			config.Port = p
		}
	}
	if f := flag.Lookup("redirect"); f != nil {
		config.UseRedirect = f.Value.String() == "true"
	}
	if f := flag.Lookup("auth"); f != nil {
		config.AuthEnabled = f.Value.String() == "true"
	}
	if f := flag.Lookup("user"); f != nil {
		config.AuthUser = f.Value.String()
	}
	if f := flag.Lookup("pass"); f != nil {
		config.AuthPass = f.Value.String()
	}
	if f := flag.Lookup("data-dir"); f != nil {
		config.DataDir = f.Value.String()
	}

	return loadFromEnv(config)
}

func loadFromEnv(config *Config) *Config {
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

type ConfigStore interface {
	GetConfig() (map[string]interface{}, error)
	SetConfig(config map[string]interface{}) error
}

func (c *Config) SaveToStore(store ConfigStore) error {
	configMap := map[string]interface{}{
		"port":         c.Port,
		"use_redirect": c.UseRedirect,
		"auth_enabled": c.AuthEnabled,
		"auth_user":    c.AuthUser,
		"auth_pass":    c.AuthPass,
		"data_dir":     c.DataDir,
	}

	return store.SetConfig(configMap)
}

func LoadFromStore(store ConfigStore) (*Config, error) {
	configMap, err := store.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config from store: %w", err)
	}

	if configMap == nil {
		return nil, nil // No config found
	}

	config := &Config{
		Port:        8080,
		UseRedirect: false,
		AuthEnabled: false,
		AuthUser:    "",
		AuthPass:    "",
		DataDir:     "./proxydavData",
	}

	if port, ok := configMap["port"].(float64); ok {
		config.Port = int(port)
	}
	if useRedirect, ok := configMap["use_redirect"].(bool); ok {
		config.UseRedirect = useRedirect
	}
	if authEnabled, ok := configMap["auth_enabled"].(bool); ok {
		config.AuthEnabled = authEnabled
	}
	if authUser, ok := configMap["auth_user"].(string); ok {
		config.AuthUser = authUser
	}
	if authPass, ok := configMap["auth_pass"].(string); ok {
		config.AuthPass = authPass
	}
	if dataDir, ok := configMap["data_dir"].(string); ok {
		config.DataDir = dataDir
	}

	return config, nil
}

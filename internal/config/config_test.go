package config

import (
	"os"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	// Create a temporary test file for valid tests
	tmpFile, err := os.CreateTemp("", "test-mapping-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	content := `[{"path": "/test.txt", "url": "https://example.com/test.txt"}]`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Port:        8080,
				MappingFile: tmpFile.Name(),
				CacheTTL:    3600,
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: Config{
				Port:        0,
				MappingFile: tmpFile.Name(),
				CacheTTL:    3600,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: Config{
				Port:        99999,
				MappingFile: tmpFile.Name(),
				CacheTTL:    3600,
			},
			wantErr: true,
		},
		{
			name: "empty mapping file",
			config: Config{
				Port:        8080,
				MappingFile: "",
				CacheTTL:    3600,
			},
			wantErr: true,
		},
		{
			name: "non-existent mapping file",
			config: Config{
				Port:        8080,
				MappingFile: "non-existent.json",
				CacheTTL:    3600,
			},
			wantErr: true,
		},
		{
			name: "negative cache TTL",
			config: Config{
				Port:        8080,
				MappingFile: tmpFile.Name(),
				CacheTTL:    -1,
			},
			wantErr: true,
		},
		{
			name: "auth enabled without credentials",
			config: Config{
				Port:        8080,
				MappingFile: tmpFile.Name(),
				CacheTTL:    3600,
				AuthEnabled: true,
			},
			wantErr: true,
		},
		{
			name: "auth enabled with credentials",
			config: Config{
				Port:        8080,
				MappingFile: tmpFile.Name(),
				CacheTTL:    3600,
				AuthEnabled: true,
				AuthUser:    "user",
				AuthPass:    "pass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFileEntries(t *testing.T) {
	// Create a temporary test file
	content := `[
		{
			"path": "/test/file.txt",
			"url": "https://example.com/file.txt"
		}
	]`

	tmpFile, err := os.CreateTemp("", "test-mapping-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading valid file
	entries, err := LoadFileEntries(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadFileEntries() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Path != "/test/file.txt" {
		t.Errorf("Expected path '/test/file.txt', got '%s'", entries[0].Path)
	}

	if entries[0].URL != "https://example.com/file.txt" {
		t.Errorf("Expected URL 'https://example.com/file.txt', got '%s'", entries[0].URL)
	}

	// Test loading non-existent file
	_, err = LoadFileEntries("non-existent.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

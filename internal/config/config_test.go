package config

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Port:    8080,
				DataDir: "./proxydavData",
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: Config{
				Port:    0,
				DataDir: "./proxydavData",
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: Config{
				Port:    99999,
				DataDir: "./proxydavData",
			},
			wantErr: true,
		},
		{
			name: "empty data directory",
			config: Config{
				Port:    8080,
				DataDir: "",
			},
			wantErr: true,
		},
		{
			name: "auth enabled without credentials",
			config: Config{
				Port:        8080,
				DataDir:     "./proxydavData",
				AuthEnabled: true,
			},
			wantErr: true,
		},
		{
			name: "auth enabled with credentials",
			config: Config{
				Port:        8080,
				DataDir:     "./proxydavData",
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

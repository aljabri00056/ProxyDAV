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
				Port:     8080,
				CacheTTL: 3600,
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: Config{
				Port:     0,
				CacheTTL: 3600,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: Config{
				Port:     99999,
				CacheTTL: 3600,
			},
			wantErr: true,
		},
		{
			name: "negative cache TTL",
			config: Config{
				Port:     8080,
				CacheTTL: -1,
			},
			wantErr: true,
		},
		{
			name: "auth enabled without credentials",
			config: Config{
				Port:        8080,
				CacheTTL:    3600,
				AuthEnabled: true,
			},
			wantErr: true,
		},
		{
			name: "auth enabled with credentials",
			config: Config{
				Port:        8080,
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

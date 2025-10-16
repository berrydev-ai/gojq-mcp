package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expected    *Config
		expectError bool
	}{
		{
			name: "config with auth_token",
			configYAML: `data_path: /data
transport: http
port: 9000
auth_token: my-secret-token
`,
			expected: &Config{
				DataPath:  "/data",
				Transport: "http",
				Port:      9000,
				AuthToken: "my-secret-token",
			},
			expectError: false,
		},
		{
			name: "minimal config with defaults",
			configYAML: `
data_path: /data
`,
			expected: &Config{
				DataPath:  "/data",
				Transport: "stdio",
				Port:      8080,
			},
			expectError: false,
		},
		{
			name:        "invalid YAML",
			configYAML:  `invalid: yaml: [content`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "nonexistent file",
			configYAML:  "",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.configYAML != "" {
				// Create temporary config file
				tempDir := t.TempDir()
				configPath = filepath.Join(tempDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
				require.NoError(t, err)
			} else {
				configPath = "/nonexistent/file.yaml"
			}

			config, err := LoadConfig(configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, config)
			}
		})
	}
}

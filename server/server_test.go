package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/berrydev-ai/gojq-mcp/config"
	"github.com/berrydev-ai/gojq-mcp/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupMCPServer(t *testing.T) {
	// Create temporary directory for registry
	tempDir := t.TempDir()
	fileRegistry, err := registry.NewFileRegistry(tempDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataPath:  tempDir,
		Transport: "stdio",
		Port:      8080,
		Prompts: []config.PromptConfig{
			{
				Name:        "test_prompt",
				Description: "A test prompt",
				Arguments: []config.PromptArgumentConfig{
					{
						Name:        "arg1",
						Description: "First argument",
						Required:    true,
					},
				},
			},
		},
	}

	s, err := SetupMCPServer(cfg, fileRegistry)
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// Test that server was created successfully
	// Note: MCP-Go server API doesn't expose ListTools/ListPrompts methods
	// We can only test that the function doesn't error
}

func TestSetupMCPServerWithTools(t *testing.T) {
	// Create temporary directory with test file
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "test.json")
	jsonContent := `{"name": "test", "value": 42}`

	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	require.NoError(t, err)

	fileRegistry, err := registry.NewFileRegistry(tempDir)
	require.NoError(t, err)

	cfg := &config.Config{
		DataPath:  tempDir,
		Transport: "stdio",
		Port:      8080,
	}

	s, err := SetupMCPServer(cfg, fileRegistry)
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// Test that the server was created successfully
	// We can't easily test the MCP tool execution without MCP client setup
	// But we can verify the server creation doesn't fail
}

package jq

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteJQ(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		input     interface{}
		expected  string
		expectErr bool
	}{
		{
			name:   "identity filter",
			filter: ".",
			input: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expected:  "{\n  \"age\": 30,\n  \"name\": \"John\"\n}",
			expectErr: false,
		},
		{
			name:   "field extraction",
			filter: ".name",
			input: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expected:  "\"John\"",
			expectErr: false,
		},
		{
			name:      "invalid filter",
			filter:    "invalid[",
			input:     map[string]interface{}{},
			expected:  "",
			expectErr: true,
		},
		{
			name:   "array result",
			filter: ".[]",
			input: []interface{}{
				"item1",
				"item2",
			},
			expected:  "[\n  \"item1\",\n  \"item2\"\n]",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteJQ(tt.filter, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExecuteJQMultiFiles(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		input     []interface{}
		expected  string
		expectErr bool
	}{
		{
			name:   "collect all inputs",
			filter: "[inputs]",
			input: []interface{}{
				map[string]interface{}{"name": "file1"},
				map[string]interface{}{"name": "file2"},
			},
			expected:  "[\n  {\n    \"name\": \"file1\"\n  },\n  {\n    \"name\": \"file2\"\n  }\n]",
			expectErr: false,
		},
		{
			name:   "process each input",
			filter: "inputs | .name",
			input: []interface{}{
				map[string]interface{}{"name": "file1"},
				map[string]interface{}{"name": "file2"},
			},
			expected:  "[\n  \"file1\",\n  \"file2\"\n]",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteJQMultiFiles(tt.filter, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProcessJQQuery(t *testing.T) {
	// Create temporary directory with test files
	tempDir := t.TempDir()

	// Create test JSON files
	file1 := filepath.Join(tempDir, "test1.json")
	file2 := filepath.Join(tempDir, "test2.json")

	err := os.WriteFile(file1, []byte(`{"name": "file1", "value": 1}`), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(`{"name": "file2", "value": 2}`), 0644)
	require.NoError(t, err)

	tests := []struct {
		name      string
		filter    string
		patterns  []string
		expected  string
		expectErr bool
	}{
		{
			name:     "single file query",
			filter:   ".name",
			patterns: []string{"test1.json"},
			expected: "\"file1\"",
			expectErr: false,
		},
		{
			name:     "multiple files with inputs",
			filter:   "inputs | .name",
			patterns: []string{"test1.json", "test2.json"},
			expected:  "[\n  \"file1\",\n  \"file2\"\n]",
			expectErr: false,
		},
		{
			name:      "no files found",
			filter:    ".",
			patterns:  []string{"nonexistent.json"},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "invalid file path",
			filter:    ".",
			patterns:  []string{"/outside/data/directory.json"},
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessJQQuery(tt.filter, tt.patterns, tempDir)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExpandGlobPatterns(t *testing.T) {
	// Create temporary directory with test files
	tempDir := t.TempDir()

	// Create test files
	files := []string{"test1.json", "test2.json", "subdir/test3.json"}
	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(`{}`), 0644)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		patterns []string
		expected []string
	}{
		{
			name:     "single pattern",
			patterns: []string{filepath.Join(tempDir, "*.json")},
			expected: []string{
				filepath.Join(tempDir, "test1.json"),
				filepath.Join(tempDir, "test2.json"),
			},
		},
		{
			name:     "recursive pattern",
			patterns: []string{filepath.Join(tempDir, "**/*.json")},
			expected: []string{
				filepath.Join(tempDir, "subdir/test3.json"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandGlobPatterns(tt.patterns)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

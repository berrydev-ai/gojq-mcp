package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileRegistry(t *testing.T) {
	// Create temporary directory with test files
	tempDir := t.TempDir()

	// Create some test JSON files
	testFiles := []string{"test1.json", "test2.json", "subdir/test3.json"}
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(`{"test": "data"}`), 0644)
		require.NoError(t, err)
	}

	// Create non-JSON file (should be ignored)
	err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("not json"), 0644)
	require.NoError(t, err)

	registry, err := NewFileRegistry(tempDir)
	require.NoError(t, err)
	assert.NotNil(t, registry)

	files := registry.GetFiles()
	assert.Len(t, files, 3) // Only JSON files should be included

	// Check that files are sorted
	assert.Equal(t, filepath.Join(tempDir, "subdir/test3.json"), files[0].Path)
	assert.Equal(t, filepath.Join(tempDir, "test1.json"), files[1].Path)
	assert.Equal(t, filepath.Join(tempDir, "test2.json"), files[2].Path)
}

func TestNewFileRegistry_InvalidPath(t *testing.T) {
	registry, err := NewFileRegistry("/nonexistent/path")
	// Note: NewFileRegistry doesn't fail on non-existent paths, it just creates an empty registry
	// The error occurs during scanFiles, but we don't check that in the constructor
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "/nonexistent/path", registry.rootPath)
}

func TestFileRegistry_GetManifest(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{"file1.json", "file2.json", "subdir/file3.json"}
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(`{"test": true}`), 0644)
		require.NoError(t, err)
	}

	registry, err := NewFileRegistry(tempDir)
	require.NoError(t, err)

	manifest := registry.GetManifest()

	assert.Equal(t, 3, manifest["total_files"])
	assert.NotNil(t, manifest["files"])
	assert.NotNil(t, manifest["suggested_patterns"])

	files := manifest["files"]
	assert.NotNil(t, files)
	// The manifest structure is complex, just verify it exists and has content
	assert.NotNil(t, manifest["total_files"])
	assert.Equal(t, 3, manifest["total_files"])
}

func TestFileRegistry_StartWatching(t *testing.T) {
	tempDir := t.TempDir()

	registry, err := NewFileRegistry(tempDir)
	require.NoError(t, err)

	err = registry.StartWatching()
	require.NoError(t, err)

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a new file
	newFile := filepath.Join(tempDir, "newfile.json")
	err = os.WriteFile(newFile, []byte(`{"new": "file"}`), 0644)
	require.NoError(t, err)

	// Give debouncer time to trigger
	time.Sleep(600 * time.Millisecond)

	files := registry.GetFiles()
	assert.Len(t, files, 1)
	assert.Equal(t, newFile, files[0].Path)

	registry.Close()
}

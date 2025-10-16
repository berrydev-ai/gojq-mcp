package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// FileInfo stores metadata about discovered JSON files
type FileInfo struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// FileRegistry manages the list of discovered JSON files
type FileRegistry struct {
	mu        sync.RWMutex
	files     []FileInfo
	rootPath  string
	watcher   *fsnotify.Watcher
	debouncer *time.Timer
	mcpServer *server.MCPServer
}

// NewFileRegistry creates a new file registry
func NewFileRegistry(rootPath string) (*FileRegistry, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %w", err)
	}

	fr := &FileRegistry{
		rootPath: absPath,
		files:    make([]FileInfo, 0),
	}

	// Initial scan
	if err := fr.scanFiles(); err != nil {
		return nil, err
	}

	return fr, nil
}

// SetMCPServer sets the MCP server reference for sending notifications
func (fr *FileRegistry) SetMCPServer(s *server.MCPServer) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.mcpServer = s
}

// scanFiles discovers all JSON files in the root path
func (fr *FileRegistry) scanFiles() error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	var files []FileInfo
	err := filepath.Walk(fr.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error accessing %s: %v\n", path, err)
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		files = append(files, FileInfo{
			Path:     path,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("error scanning files: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	fr.files = files
	fmt.Fprintf(os.Stderr, "Discovered %d JSON files in %s\n", len(files), fr.rootPath)

	return nil
}

// notifyClients sends MCP notification to all connected clients
func (fr *FileRegistry) notifyClients() {
	fr.mu.RLock()
	mcpServer := fr.mcpServer
	fr.mu.RUnlock()

	if mcpServer == nil {
		return
	}

	mcpServer.SendNotificationToAllClients(
		string(mcp.MethodNotificationResourcesListChanged),
		nil,
	)

	fmt.Fprintf(os.Stderr, "📢 Sent notification to clients: resource list changed (%d files)\n", len(fr.files))
}

// StartWatching starts watching the directory for changes
func (fr *FileRegistry) StartWatching() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %w", err)
	}

	fr.watcher = watcher

	err = filepath.Walk(fr.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if err := watcher.Add(path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not watch directory %s: %v\n", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error adding watches: %w", err)
	}

	go fr.watch()

	fmt.Fprintf(os.Stderr, "👁 File watching enabled for %s\n", fr.rootPath)
	return nil
}

// watch monitors file system events
func (fr *FileRegistry) watch() {
	for {
		select {
		case event, ok := <-fr.watcher.Events:
			if !ok {
				return
			}

			isJSON := strings.HasSuffix(strings.ToLower(event.Name), ".json")
			isDir := false
			if info, err := os.Stat(event.Name); err == nil {
				isDir = info.IsDir()
			}

			if !isJSON && !isDir {
				continue
			}

			if fr.debouncer != nil {
				fr.debouncer.Stop()
			}

			fr.debouncer = time.AfterFunc(500*time.Millisecond, func() {
				fmt.Fprintf(os.Stderr, "📄 File system change detected (%s: %s), rescanning...\n", event.Op.String(), event.Name)

				if err := fr.scanFiles(); err != nil {
					fmt.Fprintf(os.Stderr, "Error rescanning files: %v\n", err)
					return
				}

				fr.notifyClients()

				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						if err := fr.watcher.Add(event.Name); err != nil {
							fmt.Fprintf(os.Stderr, "Warning: could not watch new directory %s: %v\n", event.Name, err)
						}
					}
				}
			})

		case err, ok := <-fr.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "File watcher error: %v\n", err)
		}
	}
}

// GetFiles returns a copy of the current file list
func (fr *FileRegistry) GetFiles() []FileInfo {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	files := make([]FileInfo, len(fr.files))
	copy(files, fr.files)
	return files
}

// GetManifest returns a structured manifest
func (fr *FileRegistry) GetManifest() map[string]interface{} {
	files := fr.GetFiles()

	type RelativeFileInfo struct {
		Path     string    `json:"path"`
		Modified time.Time `json:"modified"`
		Size     int64     `json:"size"`
	}

	var relativeFiles []RelativeFileInfo
	dirMap := make(map[string][]string)

	for _, file := range files {
		relPath, err := filepath.Rel(fr.rootPath, file.Path)
		if err != nil {
			relPath = file.Path
		}

		fileInfo := RelativeFileInfo{
			Path:     relPath,
			Modified: file.Modified,
			Size:     file.Size,
		}

		relativeFiles = append(relativeFiles, fileInfo)

		relDir := filepath.Dir(relPath)
		if relDir == "." {
			relDir = "base"
		}
		dirMap[relDir] = append(dirMap[relDir], relPath)
	}

	manifest := map[string]interface{}{
		"total_files": len(files),
		"files":       relativeFiles,
	}

	// Add glob pattern suggestions
	if len(files) > 0 {
		patterns := make(map[string]string)
		patterns["*.json"] = "All JSON files in base path"
		patterns["**/*.json"] = "All JSON files recursively"

		dirCounts := make(map[string]int)
		for dir, paths := range dirMap {
			dirCounts[dir] = len(paths)
		}

		for dir, count := range dirCounts {
			if count > 1 && dir != "base" {
				patterns[dir+"/*.json"] = fmt.Sprintf("All %d JSON files in %s", count, filepath.Base(dir))
			}
		}

		if len(relativeFiles) > 0 {
			specificCount := 3
			if len(relativeFiles) < specificCount {
				specificCount = len(relativeFiles)
			}
			var specificFiles []string
			for i := 0; i < specificCount; i++ {
				specificFiles = append(specificFiles, relativeFiles[i].Path)
			}
			patterns[strings.Join(specificFiles, " ")] = fmt.Sprintf("Example: Query %d specific files", specificCount)
		}

		manifest["suggested_patterns"] = patterns
	}

	return manifest
}

// Close stops the watcher
func (fr *FileRegistry) Close() error {
	if fr.watcher != nil {
		return fr.watcher.Close()
	}
	return nil
}

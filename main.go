package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/itchyny/gojq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

const version = "1.0.5"

// Config represents the YAML configuration file structure
type Config struct {
	DataPath     string         `yaml:"data_path"`
	Transport    string         `yaml:"transport"`
	Port         int            `yaml:"port"`
	Instructions string         `yaml:"instructions"`
	Prompts      []PromptConfig `yaml:"prompts"`
}

// PromptConfig defines a reusable prompt
type PromptConfig struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Arguments   []PromptArgumentConfig   `yaml:"arguments"`
}

// PromptArgumentConfig defines a prompt argument
type PromptArgumentConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Set defaults if not specified
	if config.Transport == "" {
		config.Transport = "stdio"
	}
	if config.Port == 0 {
		config.Port = 8080
	}

	return &config, nil
}

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

func printUsage() {
	fmt.Fprintf(os.Stderr, `gojq-mcp v%s - A dual-mode JSON query tool

USAGE:
  gojq-mcp [OPTIONS]

MODES:
  CLI Mode:         gojq-mcp -f <file> -q <query>
  Server Mode:      gojq-mcp -p <path> [-c <config>] [-i <instructions>]
  Generate Config:  gojq-mcp generate-config -p <path> [-o <output>]

OPTIONS:
  -f <file>       Path to JSON file (CLI mode)
  -q <query>      jq query to execute (CLI mode)
  -p <path>       Path to folder containing JSON files
  -c <config>     Path to YAML configuration file (Server mode)
  -i <instructions> Server instructions for LLM (overrides config)
  -o <output>     Output file for generated config (default: config.yaml)
  -t <transport>  Transport type: stdio, http, or sse (overrides config, default: stdio)
  -a <address>    Address to listen on for http/sse (overrides config, default: :8080)
  -token <token>  Bearer token required by http/sse transports
  -watch          Enable file system watching (default: true)
  --version       Display version information
  --help          Display this help message

FEATURES:
  • Generate config files with prompts and instructions
  • YAML configuration for transport, port, instructions, and prompts
  • Real-time file monitoring with automatic client notifications
  • HTTP streaming transport for push notifications

EXAMPLES:
  # Generate a config file
  gojq-mcp generate-config -p ./data -o config.yaml

  # CLI mode - query a JSON file
  gojq-mcp -f data.json -q '.users[] | select(.age > 30)'

  # Server mode with config file
  gojq-mcp -p ./data -c config.yaml

  # Server mode with CLI overrides
  gojq-mcp -p ./data -c config.yaml -t http -a :9000

DOCUMENTATION:
  https://github.com/berrydev-ai/gojq-mcp

`, version)
}

func expandGlobPatterns(patterns []string) ([]string, error) {
	var expandedPaths []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("error expanding glob pattern %q: %w", pattern, err)
		}
		expandedPaths = append(expandedPaths, matches...)
	}

	seen := make(map[string]bool)
	var uniquePaths []string
	for _, path := range expandedPaths {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}
	sort.Strings(uniquePaths)

	return uniquePaths, nil
}

func validateAndReadJSONFiles(filePaths []string) ([]interface{}, error) {
	var jsonData []interface{}

	for _, filePath := range filePaths {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file does not exist: %s", filePath)
			}
			return nil, fmt.Errorf("error accessing file %s: %w", filePath, err)
		}

		if fileInfo.IsDir() {
			return nil, fmt.Errorf("path is a directory, not a file: %s", filePath)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("file %s is not readable: %w", filePath, err)
		}

		var parsedData interface{}
		if err := json.Unmarshal(data, &parsedData); err != nil {
			return nil, fmt.Errorf("file %s does not contain valid JSON: %w", filePath, err)
		}

		jsonData = append(jsonData, parsedData)
	}

	return jsonData, nil
}

func executeJQ(jqFilter string, jsonData interface{}) (string, error) {
	query, err := gojq.Parse(jqFilter)
	if err != nil {
		return "", fmt.Errorf("invalid jq filter: %w", err)
	}

	iter := query.Run(jsonData)
	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return "", fmt.Errorf("jq execution error: %w", err)
		}
		results = append(results, v)
	}

	var output []byte
	if len(results) == 1 {
		output, err = json.MarshalIndent(results[0], "", "  ")
	} else {
		output, err = json.MarshalIndent(results, "", "  ")
	}

	if err != nil {
		return "", fmt.Errorf("error formatting results: %w", err)
	}

	return string(output), nil
}

func executeJQMultiFiles(jqFilter string, jsonData []interface{}) (string, error) {
	query, err := gojq.Parse(jqFilter)
	if err != nil {
		return "", fmt.Errorf("invalid jq filter: %w", err)
	}

	inputIter := gojq.NewIter(jsonData...)

	code, err := gojq.Compile(query, gojq.WithInputIter(inputIter))
	if err != nil {
		return "", fmt.Errorf("failed to compile jq query: %w", err)
	}

	iter := code.Run(nil)
	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return "", fmt.Errorf("jq execution error: %w", err)
		}
		results = append(results, v)
	}

	var output []byte
	if len(results) == 1 {
		output, err = json.MarshalIndent(results[0], "", "  ")
	} else {
		output, err = json.MarshalIndent(results, "", "  ")
	}

	if err != nil {
		return "", fmt.Errorf("error formatting results: %w", err)
	}

	return string(output), nil
}

func runCLIMode(filePath, query string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	result, err := executeJQ(query, jsonData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing jq query: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func extractBearerToken(header string) (string, bool) {
	const bearerPrefix = "Bearer "
	if header == "" {
		return "", false
	}
	if len(header) < len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(bearerPrefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

func tokensMatch(expected, candidate string) bool {
	if expected == "" {
		return true
	}
	if candidate == "" || len(candidate) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(candidate)) == 1
}

func authorizeHTTPBearer(expected string, r *http.Request) bool {
	if expected == "" {
		return true
	}
	candidate, ok := extractBearerToken(r.Header.Get("Authorization"))
	if !ok {
		return false
	}
	return tokensMatch(expected, candidate)
}

func authorizeSSEToken(expected string, r *http.Request) bool {
	if expected == "" {
		return true
	}
	if tokensMatch(expected, r.URL.Query().Get("token")) {
		return true
	}
	if candidate, ok := extractBearerToken(r.Header.Get("Authorization")); ok {
		return tokensMatch(expected, candidate)
	}
	return false
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer realm=\"gojq-mcp\"")
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func main() {
	flag.Usage = printUsage

	filePath := flag.String("f", "", "Path to JSON file")
	query := flag.String("q", "", "jq query to execute")
	dataPath := flag.String("p", "", "Path to folder containing JSON files")
	configPath := flag.String("c", "", "Path to YAML configuration file")
	instructions := flag.String("i", "", "Server instructions for LLM (overrides config)")
	transport := flag.String("t", "", "Transport type: stdio, http, or sse (overrides config)")
	address := flag.String("a", "", "Address to listen on (overrides config)")
	tokenFlag := flag.String("token", "", "Bearer token required by http/sse transports")
	enableWatch := flag.Bool("watch", true, "Enable file system watching")
	showVersion := flag.Bool("version", false, "Display version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gojq-mcp version %s\n", version)
		return
	}

	authToken := strings.TrimSpace(*tokenFlag)

	// CLI mode
	if *filePath != "" && *query != "" {
		runCLIMode(*filePath, *query)
		return
	}

	// Server mode - load config or use defaults
	var config *Config
	if *configPath != "" {
		var err error
		config, err = LoadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "✅ Loaded configuration from %s\n", *configPath)
	} else {
		// Default config
		config = &Config{
			Transport: "stdio",
			Port:      8080,
		}
	}

	// Override config with CLI flags if provided
	if *dataPath != "" {
		config.DataPath = *dataPath
	}
	if *transport != "" {
		config.Transport = *transport
	}
	if *address != "" {
		// Parse port from address
		parts := strings.Split(*address, ":")
		if len(parts) == 2 {
			fmt.Sscanf(parts[1], "%d", &config.Port)
		}
	}
	if *instructions != "" {
		config.Instructions = *instructions
	}

	// Verify data path is set
	if config.DataPath == "" {
		fmt.Fprintf(os.Stderr, "Error: data path is required. Use -p flag or set data_path in config\n\n")
		printUsage()
		os.Exit(1)
	}

	// Verify data path exists
	if info, err := os.Stat(config.DataPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: data path does not exist: %s\n", config.DataPath)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: data path is not a directory: %s\n", config.DataPath)
		os.Exit(1)
	}

	// Create file registry
	registry, err := NewFileRegistry(config.DataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing file registry: %v\n", err)
		os.Exit(1)
	}
	defer registry.Close()

	// Create MCP server with instructions
	serverOpts := []server.ServerOption{
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	}

	if config.Instructions != "" {
		serverOpts = append(serverOpts, server.WithInstructions(config.Instructions))
	}

	s := server.NewMCPServer("GoJQ MCP Server", version, serverOpts...)

	registry.SetMCPServer(s)

	// Start file watching if enabled
	if *enableWatch {
		if err := registry.StartWatching(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not enable file watching: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing without file watching...\n")
		} else {
			fmt.Fprintf(os.Stderr, "✅ Push notifications enabled - clients will be notified of file changes\n")
		}
	}

	// Register prompts
	for _, promptConfig := range config.Prompts {
		prompt := mcp.NewPrompt(
			promptConfig.Name,
			mcp.WithPromptDescription(promptConfig.Description),
		)

		s.AddPrompt(prompt, func(pc PromptConfig) server.PromptHandlerFunc {
			return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				// Build example based on prompt
				exampleQuery := fmt.Sprintf("Prompt: %s\n\nUse list_data_files to discover available files, then run_jq to query them.", pc.Description)

				// Include argument values if provided
				if len(request.Params.Arguments) > 0 {
					exampleQuery += "\n\nProvided arguments:"
					for k, v := range request.Params.Arguments {
						exampleQuery += fmt.Sprintf("\n- %s: %v", k, v)
					}
				}

				return mcp.NewGetPromptResult(
					pc.Description,
					[]mcp.PromptMessage{
						mcp.NewPromptMessage(
							mcp.RoleAssistant,
							mcp.NewTextContent(exampleQuery),
						),
					},
				), nil
			}
		}(promptConfig))
	}

	// Add run_jq tool
	runJqTool := mcp.NewTool("run_jq",
		mcp.WithDescription(`Queries JSON data using jq syntax. Supports single files, multiple files, and glob patterns.

FILE SPECIFICATION (relative to data directory):
- Single file: "file.json"
- Multiple files: "file1.json file2.json file3.json"
- Glob patterns: "subdir/*.json"
- Mixed: "schema.json segments/*.json"

JQ FILTER EXAMPLES:
- Identity: '.'
- Extract field: '.name' or '.users[0].email'
- Array iteration: '.users[]'
- Filter: '.users[] | select(.age > 30)'
- Multi-file collection: '[inputs]' (collects all input files into an array)
- Multi-file processing: 'inputs | .name' (processes each file separately)

REAL-TIME UPDATES:
When file watching is enabled, this server automatically notifies clients when files change.

TIP: Use 'list_data_files' first to discover available files.`),
		mcp.WithString("jq_filter",
			mcp.Required(),
			mcp.Description("The jq filter to execute. Use 'inputs' function for multi-file queries."),
		),
		mcp.WithString("json_file_path",
			mcp.Required(),
			mcp.Description("Space-separated string of file paths (relative to data directory) or glob patterns."),
		),
	)

	s.AddTool(runJqTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jqFilter, err := request.RequireString("jq_filter")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonFilePath, err := request.RequireString("json_file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		patterns := strings.Fields(jsonFilePath)
		if len(patterns) == 0 {
			return mcp.NewToolResultError("json_file_path cannot be empty"), nil
		}

		// Convert relative paths to absolute paths
		absolutePatterns := make([]string, len(patterns))
		for i, pattern := range patterns {
			if !filepath.IsAbs(pattern) {
				absolutePatterns[i] = filepath.Join(config.DataPath, pattern)
			} else {
				// Security check: verify absolute path is within data directory
				absDataPath, _ := filepath.Abs(config.DataPath)
				absPattern, _ := filepath.Abs(pattern)
				if !strings.HasPrefix(absPattern, absDataPath) {
					return mcp.NewToolResultError(fmt.Sprintf("access denied: path %s is outside data directory", pattern)), nil
				}
				absolutePatterns[i] = pattern
			}
		}

		expandedPaths, err := expandGlobPatterns(absolutePatterns)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error expanding glob patterns: %v", err)), nil
		}

		if len(expandedPaths) == 0 {
			return mcp.NewToolResultError("no files found matching the provided patterns"), nil
		}

		jsonDataList, err := validateAndReadJSONFiles(expandedPaths)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var results string
		if len(jsonDataList) == 1 {
			results, err = executeJQ(jqFilter, jsonDataList[0])
		} else {
			results, err = executeJQMultiFiles(jqFilter, jsonDataList)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(results), nil
	})

	// Add list_data_files tool
	listFilesTool := mcp.NewTool("list_data_files",
		mcp.WithDescription(`Lists all available JSON data files with metadata.

Returns file paths (relative to data directory), modification times, sizes, and suggested query patterns.

REAL-TIME UPDATES:
This server monitors the file system. When files change, clients receive
'notifications/resources/list_changed'. Call this tool again for updated information.`),
	)

	s.AddTool(listFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		manifest := registry.GetManifest()
		output, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error formatting manifest: %v", err)), nil
		}
		return mcp.NewToolResultText(string(output)), nil
	})

	// Build address string
	addressStr := fmt.Sprintf(":%d", config.Port)
	if *address != "" {
		addressStr = *address
	}

	// Start the server
	switch config.Transport {
	case "stdio":
		fmt.Fprintln(os.Stderr, "Starting MCP server with stdio transport...")
		if len(config.Prompts) > 0 {
			fmt.Fprintf(os.Stderr, "💬 Loaded %d prompt(s)\n", len(config.Prompts))
		}
		fmt.Fprintln(os.Stderr, "Note: stdio transport does not support push notifications")
		err = server.ServeStdio(s)
	case "http":
		fmt.Fprintf(os.Stderr, "Starting MCP server with HTTP streaming transport on %s...\n", addressStr)
		if len(config.Prompts) > 0 {
			fmt.Fprintf(os.Stderr, "💬 Loaded %d prompt(s)\n", len(config.Prompts))
		}
		fmt.Fprintln(os.Stderr, "✅ Push notifications ENABLED via HTTP streaming")
		if authToken != "" {
			var httpServer *server.StreamableHTTPServer
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !authorizeHTTPBearer(authToken, r) {
					writeUnauthorized(w)
					return
				}
				httpServer.ServeHTTP(w, r)
			})
			srv := &http.Server{Addr: addressStr, Handler: handler}
			httpServer = server.NewStreamableHTTPServer(s, server.WithStreamableHTTPServer(srv))
			err = httpServer.Start(addressStr)
		} else {
			httpServer := server.NewStreamableHTTPServer(s)
			err = httpServer.Start(addressStr)
		}
	case "sse":
		fmt.Fprintf(os.Stderr, "Starting MCP server with SSE transport on %s...\n", addressStr)
		if len(config.Prompts) > 0 {
			fmt.Fprintf(os.Stderr, "💬 Loaded %d prompt(s)\n", len(config.Prompts))
		}
		fmt.Fprintln(os.Stderr, "Note: SSE is deprecated, consider using 'http' transport instead")
		fmt.Fprintln(os.Stderr, "✅ Push notifications enabled via SSE")
		if authToken != "" {
			var sseServerInstance *server.SSEServer
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !authorizeSSEToken(authToken, r) {
					writeUnauthorized(w)
					return
				}
				sseServerInstance.ServeHTTP(w, r)
			})
			srv := &http.Server{Addr: addressStr, Handler: handler}
			opts := []server.SSEOption{server.WithHTTPServer(srv), server.WithAppendQueryToMessageEndpoint()}
			sseServerInstance = server.NewSSEServer(s, opts...)
			err = sseServerInstance.Start(addressStr)
		} else {
			sseServer := server.NewSSEServer(s)
			err = sseServer.Start(addressStr)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid transport type '%s'. Must be 'stdio', 'http', or 'sse'\n", config.Transport)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

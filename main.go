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

	"github.com/itchyny/gojq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const version = "1.0.3"

func printUsage() {
	fmt.Fprintf(os.Stderr, `gojq-mcp v%s - A dual-mode JSON query tool

USAGE:
  gojq-mcp [OPTIONS]

MODES:
  CLI Mode:    gojq-mcp -f <file> -q <query>
  Server Mode: gojq-mcp [-t <transport>] [-a <address>]

OPTIONS:
  -f <file>       Path to JSON file (CLI mode)
  -q <query>      jq query to execute (CLI mode)
  -t <transport>  Transport type: stdio, http, or sse (default: stdio)
  -a <address>    Address to listen on for http/sse (default: :8080)
  -token <token>  Bearer token required by http/sse transports
  --default-json-file <file>
                  Default JSON file path for http/sse transports
  --version       Display version information
  --help          Display this help message

EXAMPLES:
  # CLI mode - query a JSON file
  gojq-mcp -f data.json -q '.users[] | select(.age > 30)'

  # Server mode - stdio transport (default)
  gojq-mcp

  # Server mode - HTTP transport
  gojq-mcp -t http
  gojq-mcp -t http -a :9000

  # Server mode - SSE transport
  gojq-mcp -t sse
  gojq-mcp -t sse -a :8080

DOCUMENTATION:
  https://github.com/berrydev-ai/gojq-mcp

`, version)
}

// expandGlobPatterns expands glob patterns to actual file paths
func expandGlobPatterns(patterns []string) ([]string, error) {
	var expandedPaths []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("error expanding glob pattern %q: %w", pattern, err)
		}
		expandedPaths = append(expandedPaths, matches...)
	}

	// Remove duplicates and sort for consistent ordering
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

// validateAndReadJSONFiles validates and reads multiple JSON files
func validateAndReadJSONFiles(filePaths []string) ([]interface{}, error) {
	var jsonData []interface{}

	for _, filePath := range filePaths {
		// Check if file exists
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file does not exist: %s", filePath)
			}
			return nil, fmt.Errorf("error accessing file %s: %w", filePath, err)
		}

		// Ensure it's a file, not a directory
		if fileInfo.IsDir() {
			return nil, fmt.Errorf("path is a directory, not a file: %s", filePath)
		}

		// Read file contents
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("file %s is not readable: %w", filePath, err)
		}

		// Parse JSON
		var parsedData interface{}
		if err := json.Unmarshal(data, &parsedData); err != nil {
			return nil, fmt.Errorf("file %s does not contain valid JSON: %w", filePath, err)
		}

		jsonData = append(jsonData, parsedData)
	}

	return jsonData, nil
}

// executeJQ runs a jq query on JSON data and returns the results as a JSON string
func executeJQ(jqFilter string, jsonData interface{}) (string, error) {
	// Parse the jq filter
	query, err := gojq.Parse(jqFilter)
	if err != nil {
		return "", fmt.Errorf("invalid jq filter: %w", err)
	}

	// Execute the query on the JSON data
	iter := query.Run(jsonData)
	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			// Check for halt errors (normal termination)
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return "", fmt.Errorf("jq execution error: %w", err)
		}
		results = append(results, v)
	}

	// If only one result, return it directly; otherwise return as array
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

// executeJQMultiFiles runs a jq query on multiple JSON files using the 'inputs' function
func executeJQMultiFiles(jqFilter string, jsonData []interface{}) (string, error) {
	// Parse the jq filter
	query, err := gojq.Parse(jqFilter)
	if err != nil {
		return "", fmt.Errorf("invalid jq filter: %w", err)
	}

	// Create an iterator for the input data
	inputIter := gojq.NewIter(jsonData...)

	// Compile the query with input iterator
	code, err := gojq.Compile(query, gojq.WithInputIter(inputIter))
	if err != nil {
		return "", fmt.Errorf("failed to compile jq query: %w", err)
	}

	// Execute the query - use nil as the main input since we're using inputs
	iter := code.Run(nil)
	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			// Check for halt errors (normal termination)
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return "", fmt.Errorf("jq execution error: %w", err)
		}
		results = append(results, v)
	}

	// If only one result, return it directly; otherwise return as array
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

// runCLIMode executes jq query on a file and prints the result
func runCLIMode(filePath, query string) {
	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Execute jq query
	result, err := executeJQ(query, jsonData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing jq query: %v\n", err)
		os.Exit(1)
	}

	// Print result
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
	// Custom usage function
	flag.Usage = printUsage

	// Parse CLI flags
	filePath := flag.String("f", "", "Path to JSON file")
	query := flag.String("q", "", "jq query to execute")
	transport := flag.String("t", "stdio", "Transport type: stdio, http, or sse")
	address := flag.String("a", ":8080", "Address to listen on (for http/sse transports)")
	defaultJSON := flag.String("default-json-file", "", "Default JSON file path for http/sse transports")
	tokenFlag := flag.String("token", "", "Bearer token required by http/sse transports")
	showVersion := flag.Bool("version", false, "Display version information")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("gojq-mcp version %s\n", version)
		return
	}

	authToken := strings.TrimSpace(*tokenFlag)

	resolvedDefaultJSONPath := ""
	if *defaultJSON != "" {
		absPath, err := filepath.Abs(*defaultJSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving default JSON file path: %v\n", err)
			os.Exit(1)
		}
		resolvedDefaultJSONPath = absPath
	}

	// If CLI flags are provided, run in CLI mode
	if *filePath != "" && *query != "" {
		runCLIMode(*filePath, *query)
		return
	}

	// Otherwise, run in MCP server mode
	// Create a new MCP server
	s := server.NewMCPServer(
		"GoJQ MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	// Add a jq query tool
	jsonFileDescription := "Absolute path to the JSON file to process"
	if resolvedDefaultJSONPath != "" && *transport != "stdio" {
		jsonFileDescription = "Absolute path to the JSON file to process (optional when a default is configured)"
	}
	jsonFileArgOptions := []mcp.PropertyOption{mcp.Description(jsonFileDescription)}
	if *transport == "stdio" || resolvedDefaultJSONPath == "" {
		jsonFileArgOptions = append(jsonFileArgOptions, mcp.Required())
	}

	runJqTool := mcp.NewTool("run_jq",
		mcp.WithDescription("Queries the JSON data using a jq query. Supports single files or multiple files with glob patterns."),
		mcp.WithString("jq_filter",
			mcp.Required(),
			mcp.Description("The jq filter to execute. Use 'inputs' for multi-file queries"),
		),
		mcp.WithString("json_file_path",
			jsonFileArgOptions...,
		),
	)


	// Add the run_jq handler
	s.AddTool(runJqTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get arguments
		jqFilter, err := request.RequireString("jq_filter")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Handle json_file_path parameter - can be string or []string
		var filePaths []string
		if args := request.GetArguments(); args != nil {
			if rawPath, ok := args["json_file_path"]; ok {
				switch pathValue := rawPath.(type) {
				case string:
					// Single file path
					filePaths = []string{pathValue}
				case []interface{}:
					// Array of file paths
					for _, item := range pathValue {
						if str, ok := item.(string); ok {
							filePaths = append(filePaths, str)
						} else {
							return mcp.NewToolResultError("all elements in json_file_path array must be strings"), nil
						}
					}
				default:
					return mcp.NewToolResultError("json_file_path must be a string or array of strings"), nil
				}
			}
		}

		// For backward compatibility, also check for json_file_paths (plural)
		if len(filePaths) == 0 {
			if args := request.GetArguments(); args != nil {
				if rawPaths, ok := args["json_file_paths"]; ok {
					if pathsArray, ok := rawPaths.([]interface{}); ok {
						for _, item := range pathsArray {
							if str, ok := item.(string); ok {
								filePaths = append(filePaths, str)
							} else {
								return mcp.NewToolResultError("all elements in json_file_paths array must be strings"), nil
							}
						}
					}
				}
			}
		}

		// If no paths provided, use default for non-stdio transports
		if len(filePaths) == 0 {
			if *transport != "stdio" && resolvedDefaultJSONPath != "" {
				filePaths = []string{resolvedDefaultJSONPath}
			} else {
				return mcp.NewToolResultError("required argument \"json_file_path\" not found"), nil
			}
		}

		// Expand glob patterns
		expandedPaths, err := expandGlobPatterns(filePaths)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error expanding glob patterns: %v", err)), nil
		}

		if len(expandedPaths) == 0 {
			return mcp.NewToolResultError("no files found matching the provided patterns"), nil
		}

		// Validate and read all JSON files
		jsonDataList, err := validateAndReadJSONFiles(expandedPaths)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Execute jq query - use multi-file execution if multiple files
		var results string
		if len(jsonDataList) == 1 {
			// Single file - use original executeJQ
			results, err = executeJQ(jqFilter, jsonDataList[0])
		} else {
			// Multiple files - use executeJQMultiFiles with inputs
			results, err = executeJQMultiFiles(jqFilter, jsonDataList)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(results), nil
	})

	// Start the server with the specified transport
	var err error
	switch *transport {
	case "stdio":
		fmt.Fprintln(os.Stderr, "Starting MCP server with stdio transport...")
		err = server.ServeStdio(s)
	case "http":
		fmt.Fprintf(os.Stderr, "Starting MCP server with HTTP transport on %s...\n", *address)
		if authToken != "" {
			var httpServer *server.StreamableHTTPServer
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !authorizeHTTPBearer(authToken, r) {
					writeUnauthorized(w)
					return
				}
				httpServer.ServeHTTP(w, r)
			})
			srv := &http.Server{Addr: *address, Handler: handler}
			httpServer = server.NewStreamableHTTPServer(s, server.WithStreamableHTTPServer(srv))
			err = httpServer.Start(*address)
		} else {
			httpServer := server.NewStreamableHTTPServer(s)
			err = httpServer.Start(*address)
		}
	case "sse":
		fmt.Fprintf(os.Stderr, "Starting MCP server with SSE transport on %s...\n", *address)
		if authToken != "" {
			var sseServerInstance *server.SSEServer
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !authorizeSSEToken(authToken, r) {
					writeUnauthorized(w)
					return
				}
				sseServerInstance.ServeHTTP(w, r)
			})
			srv := &http.Server{Addr: *address, Handler: handler}
			opts := []server.SSEOption{server.WithHTTPServer(srv), server.WithAppendQueryToMessageEndpoint()}
			sseServerInstance = server.NewSSEServer(s, opts...)
			err = sseServerInstance.Start(*address)
		} else {
			sseServer := server.NewSSEServer(s)
			err = sseServer.Start(*address)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid transport type '%s'. Must be 'stdio', 'http', or 'sse'\n", *transport)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

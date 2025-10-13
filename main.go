package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/itchyny/gojq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const version = "1.0.2"

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

func main() {
	// Custom usage function
	flag.Usage = printUsage

	// Parse CLI flags
	filePath := flag.String("f", "", "Path to JSON file")
	query := flag.String("q", "", "jq query to execute")
	transport := flag.String("t", "stdio", "Transport type: stdio, http, or sse")
	address := flag.String("a", ":8080", "Address to listen on (for http/sse transports)")
	defaultJSON := flag.String("default-json-file", "", "Default JSON file path for http/sse transports")
	showVersion := flag.Bool("version", false, "Display version information")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("gojq-mcp version %s\n", version)
		return
	}

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
		mcp.WithDescription("Queries the JSON data using a jq query."),
		mcp.WithString("jq_filter",
			mcp.Required(),
			mcp.Description("The jq filter to execute"),
		),
		mcp.WithString("json_file_path", jsonFileArgOptions...),
	)

	// Add the run_jq handler
	s.AddTool(runJqTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get arguments
		jqFilter, err := request.RequireString("jq_filter")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonFilePath := ""
		if args := request.GetArguments(); args != nil {
			if rawPath, ok := args["json_file_path"]; ok {
				pathStr, ok := rawPath.(string)
				if !ok {
					return mcp.NewToolResultError("argument \"json_file_path\" is not a string"), nil
				}
				jsonFilePath = pathStr
			}
		}

		if jsonFilePath == "" {
			if *transport != "stdio" && resolvedDefaultJSONPath != "" {
				jsonFilePath = resolvedDefaultJSONPath
			} else {
				return mcp.NewToolResultError("required argument \"json_file_path\" not found"), nil
			}
		}

		// 1. Check if file exists
		fileInfo, err := os.Stat(jsonFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultError(fmt.Sprintf("file does not exist: %s", jsonFilePath)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("error accessing file: %v", err)), nil
		}

		// Ensure it's a file, not a directory
		if fileInfo.IsDir() {
			return mcp.NewToolResultError(fmt.Sprintf("path is a directory, not a file: %s", jsonFilePath)), nil
		}

		// 2. Check if file is readable and read contents
		data, err := os.ReadFile(jsonFilePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("file is not readable: %v", err)), nil
		}

		// 3. Check if contents are valid JSON
		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("file does not contain valid JSON: %v", err)), nil
		}

		// 4. Execute jq filter on the parsed JSON data
		results, err := executeJQ(jqFilter, jsonData)
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
		httpServer := server.NewStreamableHTTPServer(s)
		err = httpServer.Start(*address)
	case "sse":
		fmt.Fprintf(os.Stderr, "Starting MCP server with SSE transport on %s...\n", *address)
		sseServer := server.NewSSEServer(s)
		err = sseServer.Start(*address)
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid transport type '%s'. Must be 'stdio', 'http', or 'sse'\n", *transport)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

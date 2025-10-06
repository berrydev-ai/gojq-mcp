package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/itchyny/gojq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

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

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"GoJQ MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	// Add a jq query tool
	runJqTool := mcp.NewTool("run_jq",
		mcp.WithDescription("Queries the JSON data using a jq query."),
		mcp.WithString("jq_filter",
			mcp.Required(),
			mcp.Description("The jq filter to execute"),
		),
		mcp.WithString("json_file_path",
			mcp.Required(),
			mcp.Description("Absolute path to the JSON file to process"),
		),
	)

	// Add the run_jq handler
	s.AddTool(runJqTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get arguments
		jqFilter, err := request.RequireString("jq_filter")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonFilePath, err := request.RequireString("json_file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

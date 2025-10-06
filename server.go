package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

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
		mcp.WithString("json_file_path", // Changed from Number to String
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

		jsonFilePath, err := request.RequireString("json_file_path") // Changed from RequireFloat
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

		// TODO: Execute jq filter on jsonData
		// You'll need to integrate a jq library here
		// For now, returning a placeholder
		return mcp.NewToolResultText(fmt.Sprintf("Would execute jq filter '%s' on %s", jqFilter, jsonFilePath)), nil
	})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

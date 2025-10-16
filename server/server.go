package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/berrydev-ai/gojq-mcp/auth"
	"github.com/berrydev-ai/gojq-mcp/config"
	"github.com/berrydev-ai/gojq-mcp/jq"
	"github.com/berrydev-ai/gojq-mcp/registry"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SetupMCPServer creates and configures the MCP server with tools and prompts
func SetupMCPServer(cfg *config.Config, fileRegistry *registry.FileRegistry) (*server.MCPServer, error) {
	serverOpts := []server.ServerOption{
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	}

	if cfg.Instructions != "" {
		serverOpts = append(serverOpts, server.WithInstructions(cfg.Instructions))
	}

	s := server.NewMCPServer("GoJQ MCP Server", "1.0.5", serverOpts...)

	// Register prompts
	for _, promptConfig := range cfg.Prompts {
		prompt := mcp.NewPrompt(
			promptConfig.Name,
			mcp.WithPromptDescription(promptConfig.Description),
		)

		s.AddPrompt(prompt, func(pc config.PromptConfig) server.PromptHandlerFunc {
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

		results, err := jq.ProcessJQQuery(jqFilter, patterns, cfg.DataPath)
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
		manifest := fileRegistry.GetManifest()
		output, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error formatting manifest: %v", err)), nil
		}
		return mcp.NewToolResultText(string(output)), nil
	})

	return s, nil
}

// StartServer starts the MCP server with the specified transport
func StartServer(s *server.MCPServer, cfg *config.Config, authToken string) error {
	addressStr := fmt.Sprintf(":%d", cfg.Port)

	switch cfg.Transport {
	case "stdio":
		fmt.Fprintln(os.Stderr, "Starting MCP server with stdio transport...")
		if len(cfg.Prompts) > 0 {
			fmt.Fprintf(os.Stderr, "💬 Loaded %d prompt(s)\n", len(cfg.Prompts))
		}
		fmt.Fprintln(os.Stderr, "Note: stdio transport does not support push notifications")
		return server.ServeStdio(s)
	case "http":
		fmt.Fprintf(os.Stderr, "Starting MCP server with HTTP streaming transport on %s...\n", addressStr)
		if len(cfg.Prompts) > 0 {
			fmt.Fprintf(os.Stderr, "💬 Loaded %d prompt(s)\n", len(cfg.Prompts))
		}
		fmt.Fprintln(os.Stderr, "✅ Push notifications ENABLED via HTTP streaming")
		if authToken != "" {
			var httpServer *server.StreamableHTTPServer
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !auth.AuthorizeHTTPBearer(authToken, r) {
					auth.WriteUnauthorized(w)
					return
				}
				httpServer.ServeHTTP(w, r)
			})
			srv := &http.Server{Addr: addressStr, Handler: handler}
			httpServer = server.NewStreamableHTTPServer(s, server.WithStreamableHTTPServer(srv))
			return httpServer.Start(addressStr)
		} else {
			httpServer := server.NewStreamableHTTPServer(s)
			return httpServer.Start(addressStr)
		}
	case "sse":
		fmt.Fprintf(os.Stderr, "Starting MCP server with SSE transport on %s...\n", addressStr)
		if len(cfg.Prompts) > 0 {
			fmt.Fprintf(os.Stderr, "💬 Loaded %d prompt(s)\n", len(cfg.Prompts))
		}
		fmt.Fprintln(os.Stderr, "Note: SSE is deprecated, consider using 'http' transport instead")
		fmt.Fprintln(os.Stderr, "✅ Push notifications enabled via SSE")
		if authToken != "" {
			var sseServerInstance *server.SSEServer
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !auth.AuthorizeSSEToken(authToken, r) {
					auth.WriteUnauthorized(w)
					return
				}
				sseServerInstance.ServeHTTP(w, r)
			})
			srv := &http.Server{Addr: addressStr, Handler: handler}
			opts := []server.SSEOption{server.WithHTTPServer(srv), server.WithAppendQueryToMessageEndpoint()}
			sseServerInstance = server.NewSSEServer(s, opts...)
			return sseServerInstance.Start(addressStr)
		} else {
			sseServer := server.NewSSEServer(s)
			return sseServer.Start(addressStr)
		}
	default:
		return fmt.Errorf("invalid transport type '%s'. Must be 'stdio', 'http', or 'sse'", cfg.Transport)
	}
}

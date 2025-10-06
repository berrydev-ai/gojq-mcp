# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A dual-mode JSON query tool that operates both as an MCP (Model Context Protocol) server and a standalone CLI executable using jq syntax via the gojq library.

## Architecture

**Execution Modes:**
- **MCP Server Mode** (default): Runs as stdio-based MCP server exposing `run_jq` tool
- **CLI Mode**: Direct command-line execution with `-file` and `-query` flags (not yet implemented in main.go)

**Core Components:**
- `executeJQ()` (main.go:15-54): Parses jq filter using gojq, executes on JSON data, returns formatted results. Single results returned directly; multiple results returned as array.
- `main()` (main.go:56-130): Initializes MCP server with `run_jq` tool. Tool validates file existence/readability/validity before executing jq filter.

**Tool Interface:**
- Tool: `run_jq`
- Parameters: `jq_filter` (string, required), `json_file_path` (string, required - absolute path)
- Validation sequence: file existence → file readability → JSON validity → jq execution

## Build & Test Commands

```bash
# Build
make build              # Builds to dist/gojq-mcp
go build -o dist/gojq-mcp .

# Test
make test              # Run all tests
go test -v ./...

# Run
make run-server        # Start MCP server mode
make run-cli           # Example CLI execution (uses examples/sample.json)

# Clean
make clean             # Remove build artifacts
```

## Dependencies

- `github.com/itchyny/gojq`: jq implementation in Go
- `github.com/mark3labs/mcp-go`: MCP protocol server framework

## Testing

Test files located in `tests/` with testdata in `tests/testdata/`:
- `valid.json`: Valid JSON test case
- `invalid.json`: Invalid JSON test case
- `nested.json`: Complex nested structure test case

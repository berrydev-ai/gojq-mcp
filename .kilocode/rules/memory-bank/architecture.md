## System Architecture

gojq-mcp is a dual-mode application that operates as both an MCP (Model Context Protocol) server and a standalone CLI tool. The architecture is designed for simplicity, reliability, and extensibility.

### Core Components

#### `executeJQ(jqFilter string, jsonData interface{}) (string, error)`
- **Location**: `main.go:64-103`
- **Purpose**: Core jq execution engine
- **Features**:
  - Parses and executes jq filters using gojq library
  - Returns single results directly, multiple results as JSON arrays
  - Handles jq execution errors gracefully
  - Formats output as indented JSON strings

#### `runCLIMode(filePath, query string)`
- **Location**: `main.go:106-130`
- **Purpose**: Direct command-line execution mode
- **Features**:
  - Reads and validates JSON files
  - Executes jq queries
  - Prints results to stdout
  - Handles errors with exit codes

#### MCP Server Components
- **Location**: `main.go:186-366`
- **Purpose**: MCP protocol implementation
- **Features**:
  - Supports stdio, HTTP, and SSE transports
  - Implements `run_jq` tool with comprehensive validation
  - Bearer token authentication for HTTP/SSE transports
  - Default JSON file support for HTTP/SSE modes

### Transport Layer

The application supports three transport protocols:

1. **Stdio Transport** (default)
   - Used for MCP client integration (e.g., Claude Desktop)
   - Requires explicit `json_file_path` in requests

2. **HTTP Transport** (`-t http`)
   - RESTful API for web-based integrations
   - Supports bearer token authentication
   - Optional default JSON file configuration

3. **SSE Transport** (`-t sse`)
   - Server-sent events for real-time integrations
   - Supports both bearer token and query parameter authentication
   - Optional default JSON file configuration

### Validation Pipeline

All JSON file processing follows a strict validation sequence:

1. **File Existence Check**: `os.Stat()` to verify file exists
2. **Directory Validation**: Ensure path points to file, not directory
3. **Readability Check**: `os.ReadFile()` to verify file permissions
4. **JSON Validity**: `json.Unmarshal()` to parse and validate JSON structure
5. **jq Execution**: Parse and execute jq filter on validated data

### Key Technical Decisions

#### Single Binary Design
- All functionality (CLI + MCP server) in one executable
- Zero external runtime dependencies
- Cross-platform compatibility via Go compilation

#### Gojq Library Choice
- Native Go implementation of jq
- No external jq binary dependency
- Better performance and integration than shelling out

#### MCP Protocol Implementation
- Uses mark3labs/mcp-go library for protocol compliance
- Supports multiple transport layers (stdio, HTTP, SSE)
- Extensible tool architecture with comprehensive documentation in docs/mcp-go/
- Follows MCP-Go patterns for server/client implementation

#### Error Handling Strategy
- Early validation with descriptive error messages
- Consistent error formatting across all modes
- Appropriate exit codes for CLI usage

### Component Relationships

```
┌─────────────────┐
│   main()        │
│                 │
│ • Parse flags   │
│ • Route to mode │
└─────────┬───────┘
          │
    ┌─────┴─────┐
    │          │
┌───▼──┐   ┌───▼──┐
│ CLI  │   │ MCP  │
│ Mode │   │Server│
└───┬──┘   └───┬──┘
    │          │
    └────┬─────┘
         │
    ┌────▼────┐
    │executeJQ│
    │  Core   │
    └─────────┘
```

### Source Code Paths

- `main.go`: Application entry point, CLI parsing, MCP server setup
- `main_test.go`: Comprehensive test suite for executeJQ function
- `go.mod`: Go module definition and dependencies
- `Makefile`: Build automation and development commands
- `examples/`: Sample JSON files and usage examples
- `tests/testdata/`: Test fixtures for validation scenarios
- `dist/`: Build output directory (gitignored)

### Critical Implementation Paths

1. **CLI Execution Path**: `main()` → flag parsing → `runCLIMode()` → file validation → `executeJQ()`
2. **MCP Server Path**: `main()` → server creation → tool registration → transport startup → request handling → validation → `executeJQ()`
3. **HTTP/SSE Path**: Additional authentication layer before request processing

### Design Patterns

- **Command Pattern**: MCP tools encapsulate jq execution logic
- **Strategy Pattern**: Different transport implementations
- **Template Method**: Validation pipeline applied consistently
- **Factory Pattern**: Server creation with transport-specific configuration

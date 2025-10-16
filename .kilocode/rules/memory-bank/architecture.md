## System Architecture

gojq-mcp is a dual-mode application that operates as both an MCP (Model Context Protocol) server and a standalone CLI tool. The architecture is designed for simplicity, reliability, and extensibility with a modular package structure.

### Core Components

#### `jq.ExecuteJQ(jqFilter string, jsonData interface{}) (string, error)`
- **Location**: `jq/jq.go:72-108`
- **Purpose**: Core jq execution engine
- **Features**:
  - Parses and executes jq filters using gojq library
  - Returns single results directly, multiple results as JSON arrays
  - Handles jq execution errors gracefully
  - Formats output as indented JSON strings

#### `jq.ExecuteJQMultiFiles(jqFilter string, jsonData []interface{}) (string, error)`
- **Location**: `jq/jq.go:110-153`
- **Purpose**: Multi-file jq execution
- **Features**:
  - Processes multiple JSON files in a single query
  - Uses gojq's input iterator for efficient processing
  - Supports 'inputs' function for cross-file operations

#### `jq.ProcessJQQuery(jqFilter string, patterns []string, dataPath string) (string, error)`
- **Location**: `jq/jq.go:155-203`
- **Purpose**: High-level query processing with glob pattern support
- **Features**:
  - Expands glob patterns into file paths
  - Validates file access and JSON structure
  - Security checks for path traversal
  - Routes to single or multi-file execution

#### `cli.RunCLIMode(filePath, query string)`
- **Location**: `cli/cli.go:12-32`
- **Purpose**: Direct command-line execution mode
- **Features**:
  - Reads and validates single JSON files
  - Executes jq queries
  - Prints results to stdout
  - Handles errors with exit codes

#### MCP Server Components
- **Location**: `server/server.go:19-142`
- **Purpose**: MCP protocol implementation
- **Features**:
  - Supports stdio, HTTP, and SSE transports
  - Implements `run_jq` tool with comprehensive validation
  - Implements `list_data_files` tool for file discovery
  - Bearer token authentication for HTTP/SSE transports
  - Configurable prompts for LLM guidance

### Transport Layer

The application supports three transport protocols:

1. **Stdio Transport** (default)
   - Used for MCP client integration (e.g., Claude Desktop)
   - Requires explicit `json_file_path` in requests

2. **HTTP Transport** (`-t http`)
   - RESTful API for web-based integrations
   - Supports bearer token authentication
   - HTTP streaming for push notifications

3. **SSE Transport** (`-t sse`)
   - Server-sent events for real-time integrations
   - Supports both bearer token and query parameter authentication
   - Deprecated in favor of HTTP transport

### File Registry System

#### `registry.FileRegistry`
- **Location**: `registry/registry.go:25-290`
- **Purpose**: Manages JSON file discovery and monitoring
- **Features**:
  - Recursive directory scanning for JSON files
  - Real-time file system watching with fsnotify
  - Debounced change notifications to MCP clients
  - Thread-safe operations with mutex protection
  - Manifest generation with glob pattern suggestions

### Configuration System

#### `config.Config`
- **Location**: `config/config.go:11-54`
- **Purpose**: YAML-based configuration management
- **Features**:
  - Data path specification
  - Transport settings (type, port)
  - Server instructions and prompts
  - Default value handling

### Authentication System

#### `auth` package
- **Location**: `auth/auth.go:1-66`
- **Purpose**: HTTP/SSE transport security
- **Features**:
  - Bearer token authentication
  - Constant-time token comparison
  - Support for Authorization header and query parameters

### Validation Pipeline

All JSON file processing follows a strict validation sequence:

1. **File Existence Check**: `os.Stat()` to verify file exists
2. **Directory Validation**: Ensure path points to file, not directory
3. **Readability Check**: `os.ReadFile()` to verify file permissions
4. **JSON Validity**: `json.Unmarshal()` to parse and validate JSON structure
5. **jq Execution**: Parse and execute jq filter on validated data

### Key Technical Decisions

#### Modular Package Structure
- Separate packages for CLI, server, jq processing, config, registry, and auth
- Clear separation of concerns
- Easy testing and maintenance

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
│ • Load config  │
│ • Route to mode │
└─────────┬───────┘
          │
    ┌─────┴─────┐
    │          │
┌───▼──┐   ┌───▼──┐
│ CLI  │   │ MCP  │
│ Mode │   │Server│
│      │   │      │
│ • jq │   │ • jq │
│ • cli│   │ • server│
└───┬──┘   │ • config│
    │      │ • registry│
    │      │ • auth│
    └─────┬─┘
         │
    ┌────▼────┐
    │   jq    │
    │  Core   │
    │         │
    │ • ExecuteJQ│
    │ • ProcessJQ│
    └─────────┘
```

### Source Code Paths

- `main.go`: Application entry point, CLI parsing, MCP server setup
- `main_test.go`: Comprehensive test suite for jq functions
- `cli/cli.go`: CLI mode implementation
- `server/server.go`: MCP server setup and tools
- `jq/jq.go`: Core jq execution and file processing
- `config/config.go`: YAML configuration management
- `registry/registry.go`: File discovery and monitoring
- `auth/auth.go`: Authentication utilities
- `go.mod`: Go module definition and dependencies
- `Makefile`: Build automation and development commands
- `examples/`: Sample JSON files and usage examples
- `tests/testdata/`: Test fixtures for validation scenarios
- `dist/`: Build output directory (gitignored)

### Critical Implementation Paths

1. **CLI Execution Path**: `main()` → flag parsing → `cli.RunCLIMode()` → file validation → `jq.ExecuteJQ()`
2. **MCP Server Path**: `main()` → config loading → `registry.NewFileRegistry()` → `server.SetupMCPServer()` → transport startup → request handling → validation → `jq.ProcessJQQuery()`
3. **HTTP/SSE Path**: Additional authentication layer (`auth.AuthorizeHTTPBearer()`) before request processing

### Design Patterns

- **Command Pattern**: MCP tools encapsulate jq execution logic
- **Strategy Pattern**: Different transport implementations
- **Template Method**: Validation pipeline applied consistently
- **Factory Pattern**: Server creation with transport-specific configuration
- **Observer Pattern**: File registry notifies MCP clients of changes
- **Repository Pattern**: File registry manages data file access

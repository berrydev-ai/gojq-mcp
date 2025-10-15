## Technologies Used

- **Go 1.24.4**: Primary programming language for the application
- **gojq**: Native Go implementation of jq for JSON processing
- **mcp-go**: Model Context Protocol server framework for MCP integration
- **Standard Library**: HTTP server, JSON parsing, file I/O, command-line parsing

## Development Setup

### Prerequisites
- Go 1.21 or higher
- Make (optional, for using Makefile commands)

### Building
```bash
# Using Makefile
make build

# Using go directly
go build -o dist/gojq-mcp .
```

### Running
```bash
# MCP server mode (default)
make run-server
# or
./dist/gojq-mcp

# CLI mode (when implemented)
make run-cli

# HTTP/SSE transport server with default data
go run . -t http --default-json-file $(pwd)/examples/sample.json
go run . -t sse --default-json-file $(pwd)/examples/sample.json
```

### Testing
```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestExecuteJQ_SimpleQuery
```

## Technical Constraints

### Single Binary Design
- All functionality (CLI + MCP server) in one executable
- Zero external runtime dependencies
- Cross-platform compatibility via Go compilation

### Transport Limitations
- Stdio transport requires explicit `json_file_path` in requests
- HTTP/SSE transports support optional default JSON file configuration
- Authentication required for HTTP/SSE transports when token is configured

### jq Compatibility
- Full jq syntax support through gojq library
- No external jq binary dependency
- Native Go performance benefits

## Code Style

- **Imports**: Standard library first, then third-party, then local packages (goimports handles this)
- **Naming**: Use Go conventions - exported names (PascalCase), unexported names (camelCase), acronyms uppercase (HTTP, JSON, MCP)
- **Error handling**: Return sentinel errors (e.g., `ErrMethodNotFound`), wrap with `fmt.Errorf("context: %w", err)`, use `errors.Is/As` for checking
- **Types**: Use explicit types; avoid `any` except for protocol flexibility (e.g., `Arguments any`); prefer strongly-typed structs
- **Comments**: All exported types/functions MUST have godoc comments starting with the name; no inline comments unless necessary
- **Testing**: Use `testify/assert` and `testify/require`; table-driven tests with `tests := []struct{ name, ... }`; test files end in `_test.go`
- **Context**: Always accept `context.Context` as first parameter in handlers and long-running functions
- **Thread safety**: Use `sync.Mutex` for shared state; document thread-safety requirements in comments
- **JSON**: Use json tags with `omitempty` for optional fields; use `json.RawMessage` for flexible/deferred parsing

## Dependencies

### Direct Dependencies
- `github.com/itchyny/gojq v0.12.17`: jq implementation in Go
- `github.com/mark3labs/mcp-go v0.41.1`: MCP protocol framework

### Indirect Dependencies
- `github.com/bahlo/generic-list-go v0.2.0`: Generic list utilities
- `github.com/buger/jsonparser v1.1.1`: Fast JSON parsing
- `github.com/google/uuid v1.6.0`: UUID generation
- `github.com/invopop/jsonschema v0.13.0`: JSON schema validation
- `github.com/itchyny/timefmt-go v0.1.6`: Time formatting
- `github.com/mailru/easyjson v0.7.7`: JSON marshaling
- `github.com/spf13/cast v1.7.1`: Type conversion
- `github.com/wk8/go-ordered-map/v2 v2.1.8`: Ordered map implementation
- `github.com/yosida95/uritemplate/v3 v3.0.2`: URI template processing
- `gopkg.in/yaml.v3 v3.0.1`: YAML processing

## Tool Usage Patterns

### Build Automation
- Makefile provides standardized build, test, and run commands
- `make build`: Compile the binary
- `make test`: Run test suite
- `make clean`: Remove build artifacts

### Development Commands
- `make run-server`: Start MCP server with stdio transport
- `make run-cli`: Execute CLI mode with sample data
- `make run-inspector`: Test MCP integration with inspector

### Release Process
- GitHub Actions automates cross-platform binary builds
- Semantic versioning with git tags
- Automated changelog generation from commit history

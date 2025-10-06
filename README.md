# gojq-mcp

A dual-mode JSON query tool that operates both as an **MCP (Model Context Protocol) server** and a **standalone CLI executable** using jq syntax via the [gojq](https://github.com/itchyny/gojq) library.

## Features

- 🔍 **Execute jq queries** on JSON files with full jq syntax support
- ✅ **Comprehensive validation**: file existence, readability, and JSON validity
- 🔄 **Dual mode operation**: Run as MCP server or CLI tool
- 🧪 **Well-tested**: 14 comprehensive test cases for the core query engine
- 📦 **Zero configuration**: Works out of the box

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [MCP Server Mode](#mcp-server-mode)
  - [CLI Mode](#cli-mode-coming-soon)
- [Architecture](#architecture)
- [MCP Tool Interface](#mcp-tool-interface)
- [Development](#development)
- [Testing](#testing)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [License](#license)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/berrydev-ai/gojq-mcp.git
cd gojq-mcp

# Build the binary
make build
# or
go build -o dist/gojq-mcp .
```

### Using Go Install

```bash
go install github.com/berrydev-ai/gojq-mcp@latest
```

## Usage

### MCP Server Mode

The default mode runs as an MCP server using stdio transport, perfect for integration with MCP clients like Claude Desktop.

**Start the server:**

```bash
./dist/gojq-mcp
```

**Configure in Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "gojq": {
      "command": "/absolute/path/to/dist/gojq-mcp",
      "args": []
    }
  }
}
```

**Using with other MCP clients:**

```json
{
  "mcpServers": {
    "gojq": {
      "command": "/absolute/path/to/dist/gojq-mcp",
      "transport": "stdio"
    }
  }
}
```

### CLI Mode (Coming Soon)

Direct command-line execution is planned for future releases:

```bash
# Proposed usage
./gojq-mcp -file data.json -query '.users[] | select(.age > 30)'

# Short flags
./gojq-mcp -f data.json -q '.[] | .name'
```

## Architecture

### Project Structure

```
gojq-mcp/
├── main.go                 # Application entry point & MCP server setup
├── main_test.go           # Comprehensive tests for executeJQ
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── README.md              # This file
├── LICENSE                # MIT License
├── CLAUDE.md              # AI assistant guidance
├── Makefile               # Build automation
├── examples/              # Example JSON files
│   └── sample.json
├── tests/                 # Test infrastructure
│   ├── main_test.go       # (stub for future tests)
│   └── testdata/          # Test fixtures
│       ├── valid.json
│       ├── invalid.json
│       └── nested.json
└── dist/                  # Build output (gitignored)
    └── gojq-mcp
```

### Core Components

#### `executeJQ(jqFilter string, jsonData interface{}) (string, error)`

The heart of the application. Parses and executes jq filters using gojq.

**Features:**
- Parses jq filter syntax
- Executes queries on parsed JSON data
- Returns single results directly, multiple results as arrays
- Handles errors gracefully with detailed messages

**Location:** `main.go:15-54`

#### `main()`

Initializes the MCP server and registers the `run_jq` tool.

**Validation sequence:**
1. File existence check
2. File readability verification
3. JSON validity validation
4. jq filter execution

**Location:** `main.go:56-130`

## MCP Tool Interface

### Tool: `run_jq`

Query JSON files using jq filter syntax.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `jq_filter` | string | ✅ Yes | The jq filter to execute (e.g., `.users[] \| .name`) |
| `json_file_path` | string | ✅ Yes | Absolute path to the JSON file to process |

**Return Value:**
- Success: JSON-formatted string containing query results
- Error: Descriptive error message

**Example Request:**

```json
{
  "jq_filter": ".users[] | select(.age > 30)",
  "json_file_path": "/absolute/path/to/data.json"
}
```

**Example Response:**

```json
{
  "name": "Bob",
  "age": 35,
  "email": "bob@example.com"
}
```

### Error Handling

The tool provides detailed error messages for:

- **File not found**: `"file does not exist: /path/to/file.json"`
- **Directory instead of file**: `"path is a directory, not a file: /path/to/dir"`
- **Permission denied**: `"file is not readable: permission denied"`
- **Invalid JSON**: `"file does not contain valid JSON: invalid character..."`
- **Invalid jq filter**: `"invalid jq filter: unexpected token..."`
- **Query execution error**: `"jq execution error: ..."`

## Development

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
```

### Cleaning

```bash
make clean
```

## Testing

The project includes comprehensive tests for the core `executeJQ` function.

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestExecuteJQ_SimpleQuery
```

### Test Coverage

14 test cases covering:

✅ **Basic queries**: `.name`, `.age`, nested access
✅ **Array operations**: Access, mapping, filtering
✅ **Advanced filters**: `select()`, pipe operations
✅ **Built-in functions**: `keys`, `length`, `type`
✅ **Error handling**: Invalid filters, non-existent keys
✅ **Edge cases**: Empty arrays, identity filter, null values
✅ **Complex scenarios**: Nested data structures, multiple results

**Test files:**
- `main_test.go`: Core executeJQ function tests
- `tests/testdata/`: Test fixtures (valid, invalid, nested JSON)

## Examples

### Basic Queries

```jq
# Get a single field
.name

# Access nested field
.user.address.city

# Array access
.users[0].name
```

### Array Operations

```jq
# Map over array
.users[] | .name

# Filter array
.users[] | select(.age > 30)

# Transform to new array
[.users[] | .email]
```

### Advanced Queries

```jq
# Multiple filters
.users[] | select(.age > 25) | select(.active == true) | .name

# Using built-in functions
.users | length

# Get object keys
keys

# Type checking
.age | type
```

### Sample Data

See `examples/sample.json` for a complete example dataset.

## Best Practices

### For MCP Server Development

1. **Always use absolute paths** when specifying `json_file_path`
2. **Validate input early**: The tool validates files before processing
3. **Handle errors gracefully**: Check error responses from the tool
4. **Test queries incrementally**: Start with simple queries and build complexity
5. **Use the identity filter** (`.`) to inspect data structure first

### For jq Queries

1. **Start simple**: Test with `.` to see full structure
2. **Use `keys`**: Discover available fields with `keys` function
3. **Pipe operations**: Build complex queries by chaining with `|`
4. **Select carefully**: Use `select()` for filtering arrays
5. **Check types**: Use `type` function to verify data types

### For Integration

1. **Restart MCP client** after configuration changes
2. **Use absolute paths** in MCP server configuration
3. **Check logs** if the server doesn't appear in your MCP client
4. **Test manually** with `go run .` before deploying
5. **Keep dependencies updated** with `go get -u` and `go mod tidy`

## Dependencies

- [gojq](https://github.com/itchyny/gojq) - Pure Go implementation of jq
- [mcp-go](https://github.com/mark3labs/mcp-go) - Model Context Protocol server framework

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new features
- Follow Go conventions and run `go fmt`
- Update documentation for API changes
- Keep commits focused and atomic

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Resources

- [jq Manual](https://jqlang.github.io/jq/manual/) - Complete jq syntax reference
- [MCP Documentation](https://modelcontextprotocol.io/) - Model Context Protocol specification
- [gojq Documentation](https://github.com/itchyny/gojq#readme) - gojq library details


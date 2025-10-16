# gojq-mcp

A dual-mode JSON query tool that operates both as an **MCP (Model Context Protocol) server** and a **standalone CLI executable** using jq syntax via the [gojq](https://github.com/itchyny/gojq) library.

## Features

- 🔍 **Execute jq queries** on JSON files with full jq syntax support
- 📁 **Multi-file support**: Query multiple files with glob patterns using `inputs`
- ✅ **Comprehensive validation**: file existence, readability, and JSON validity
- 🔄 **Dual mode operation**: Run as MCP server or CLI tool
- 🔐 **Bearer token authentication**: Secure HTTP and SSE transports
- ⚙️ **YAML configuration**: Configure transport, prompts, and instructions
- 📡 **Real-time file monitoring**: Automatic client notifications on file changes
- 🌐 **Multiple transports**: stdio, HTTP streaming, and SSE
- 🧪 **Well-tested**: 30+ comprehensive test cases across all packages
- 📦 **Zero configuration**: Works out of the box with sensible defaults

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [MCP Server Mode](#mcp-server-mode)
  - [CLI Mode](#cli-mode)
  - [HTTP and SSE Transports](#http-and-sse-transports)
  - [Configuration Files](#configuration-files)
- [MCP Tool Interface](#mcp-tool-interface)
- [Development](#development)
- [Releases and Deployment](#releases-and-deployment)
- [Testing](#testing)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Download binary

See [releases](https://github.com/berrydev-ai/gojq-mcp/releases/latest) for the latest binary builds.

The current release is [v1.0.4](https://github.com/berrydev-ai/gojq-mcp/releases/tag/v1.0.4).

- [Apple Silicon](https://github.com/berrydev-ai/gojq-mcp/releases/download/v1.0.4/gojq-mcp-darwin-amd64)
- [Apple Intel](https://github.com/berrydev-ai/gojq-mcp/releases/download/v1.0.4/gojq-mcp-darwin-arm64)
- [Linux AMD](https://github.com/berrydev-ai/gojq-mcp/releases/download/v1.0.4/gojq-mcp-linux-amd64)
- [Linux ARM](https://github.com/berrydev-ai/gojq-mcp/releases/download/v1.0.4/gojq-mcp-linux-arm64)
- [Windows](https://github.com/berrydev-ai/gojq-mcp/releases/download/v1.0.4/gojq-mcp-windows-amd64.exe)

```bash
# Download the latest binary
wget https://github.com/berrydev-ai/gojq-mcp/releases/download/v1.0.4/gojq-mcp-darwin-arm64

# Move the binary to your PATH
mv ./gojq-mcp-darwin-arm64 /usr/loca/bin/gojq-mcp

# Try it out
gojq-mcp -h
```

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

**stdio**

```json
{
  "mcpServers": {
    "gojq-mcp": {
      "command": "gojq-mcp",
    }
  }
}
```

**streaming http**

```json
{
  "mcpServers": {
    "command": "npx",
    "args": [
      "-y",
      "mcp-remote",
      "http://my-server.com:8080/mcp"
    ]
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

### CLI Mode

The CLI mode lets you run jq queries directly on local JSON files, ideal for one-off filtering and exploration.
Pass file paths and query using `-f` and `-q` flags:

**Query a single file:**

```bash
gojq-mcp -f ./examples/data/sample.json -q '.users[] | .name'
```

**Query multiple specific files:**

```bash
gojq-mcp -f ./examples/data/multiple-files/2025-01/01.json \
         -f ./examples/data/multiple-files/2025-01/02.json \
         -q '[inputs.transactions[]] | map(.amount) | add'
```

**Query files using glob patterns:**

```bash
# Query all JSON files in a directory
gojq-mcp -f './examples/data/multiple-files/2025-01/*.json' \
         -q '[inputs.transactions[]] | map(.amount) | add'

# Query across multiple months
gojq-mcp -f './examples/data/multiple-files/*/*.json' \
         -q '[inputs.transactions[] | select(.category == "services")] | length'
```

**Features:**

- Supports glob patterns for matching multiple files
- Uses `inputs` function for multi-file queries
- Automatic file validation (existence, readability, JSON validity)
- Output printed to stdout

See [examples/data/sample.json](examples/data/sample.json) and [examples/data/multiple-files/](examples/data/multiple-files/) for sample data.

### HTTP and SSE Transports

Start the server with HTTP or SSE transport for web-based integrations:

```bash
# HTTP streaming mode
gojq-mcp -t http -p ./examples/data -a :8080

# SSE mode with authentication
gojq-mcp -t sse -p ./examples/data -a :8080 -token your-secret-token

# With configuration file
gojq-mcp -p ./examples/data -c examples/config.http.yaml
```

### Configuration Files

Create a YAML configuration file to customize server behavior:

```yaml
# config.yaml
server:
  name: My Data Server
  version: 1.0.0
  transport: http
  port: 8080
  data_path: ./examples/data
  instructions: "Custom instructions for the LLM client"

prompts:
  - name: analyze_transactions
    description: "Analyze transaction data"
    arguments:
      - name: month
        description: "Month to analyze (e.g., 2025-01)"
        required: true
```

Start the server with your config:

```bash
gojq-mcp -c config.yaml
```


## MCP Tool Interface

### Tool: `run_jq`

Query JSON files using jq filter syntax with support for single and multiple files.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `jq_filter` | string | ✅ Yes | The jq filter to execute (e.g., `.users[] \| .name`) |
| `file_patterns` | array[string] | ✅ Yes | Array of file patterns (relative to data path, supports globs) |

**Return Value:**

- Success: JSON-formatted string containing query results
- Error: Descriptive error message

**Single File Example:**

```json
{
  "jq_filter": ".users[] | select(.age > 30)",
  "file_patterns": ["sample.json"]
}
```

**Response:**

```json
[
  {
    "name": "Bob",
    "age": 35,
    "email": "bob@example.com"
  }
]
```

**Multi-File Example with Glob:**

```json
{
  "jq_filter": "[inputs.transactions[]] | map(.amount) | add",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Response:**

```json
225.50
```

**Multi-File Example with Specific Files:**

```json
{
  "jq_filter": "[inputs.transactions[] | select(.category == \"services\")] | length",
  "file_patterns": [
    "multiple-files/2025-01/01.json",
    "multiple-files/2025-01/02.json"
  ]
}
```

**Key Features:**

- 🌐 **Glob pattern support**: Use wildcards to match multiple files
- 🔗 **Multi-file queries**: Use `inputs` to process multiple files
- 🔒 **Path security**: All paths are restricted to the configured data directory
- ✅ **Automatic validation**: Files are validated before processing

### Error Handling

The tool provides detailed error messages for:

- **File not found**: `"file does not exist: filename.json"`
- **Directory instead of file**: `"path is a directory, not a file: dirname"`
- **Permission denied**: `"file filename.json is not readable: permission denied"`
- **Invalid JSON**: `"file filename.json does not contain valid JSON: invalid character..."`
- **Invalid jq filter**: `"invalid jq filter: unexpected token..."`
- **Query execution error**: `"jq execution error: ..."`
- **No matching files**: `"no files found matching the provided patterns"`
- **Path outside data directory**: `"access denied: path X is outside data directory"`

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
# MCP server mode (default - stdio)
make run-server
# or
./dist/gojq-mcp -p ./examples/data

# CLI mode - single file
./dist/gojq-mcp -f ./examples/data/sample.json -q '.users[].name'

# CLI mode - multiple files with glob
./dist/gojq-mcp -f './examples/data/multiple-files/*/*.json' \
                -q '[inputs.transactions[]] | length'

# HTTP transport server
./dist/gojq-mcp -t http -p ./examples/data -a :8080

# SSE transport server with authentication
./dist/gojq-mcp -t sse -p ./examples/data -a :8080 -token mysecret

# Server with config file
./dist/gojq-mcp -c examples/config.yaml

# Disable file watching
./dist/gojq-mcp -p ./examples/data -watch=false
```

### Cleaning

```bash
make clean
```

## Releases and Deployment

### Creating a Release

The project uses GitHub Actions to automatically build and publish releases for multiple platforms when you push a version tag.

**Steps to create a release:**

1. **Commit your changes:**

   ```bash
   git add .
   git commit -m "Prepare release v1.0.0"
   git push origin main
   ```

2. **Create and push a version tag:**

   ```bash
   # Create a tag following semantic versioning
   git tag v1.0.0

   # Push the tag to trigger the build workflow
   git push origin v1.0.0
   ```

3. **Wait for the build**: GitHub Actions will automatically:
   - Build binaries for all supported platforms
   - Create a GitHub Release with the tag
   - Attach all binaries to the release
   - Generate release notes from commit history

### Supported Platforms

Each release includes binaries for:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Binary Naming Convention

Binaries are named: `gojq-mcp-{version}-{os}-{arch}`

For example, version `v1.0.0` produces:

- `gojq-mcp-v1.0.0-linux-amd64`
- `gojq-mcp-v1.0.0-linux-arm64`
- `gojq-mcp-v1.0.0-darwin-amd64`
- `gojq-mcp-v1.0.0-darwin-arm64`
- `gojq-mcp-v1.0.0-windows-amd64.exe`

### Downloading Releases

1. Go to the [Releases page](https://github.com/berrydev-ai/gojq-mcp/releases)
2. Find your desired version
3. Download the appropriate binary for your platform
4. Make it executable (Linux/macOS): `chmod +x gojq-mcp-*`

### Version Naming

Follow [Semantic Versioning](https://semver.org/):

- **Major version** (`v2.0.0`): Breaking changes
- **Minor version** (`v1.1.0`): New features, backwards compatible
- **Patch version** (`v1.0.1`): Bug fixes, backwards compatible
- **Pre-release** (`v1.0.0-beta`, `v1.0.0-alpha.1`): Testing versions

### Continuous Integration

The project includes two GitHub Actions workflows:

**Test Workflow** (`.github/workflows/test.yml`):

- Runs on pushes and pull requests to `main`/`master`
- Tests against Go 1.24 and 1.25 (all currently supported versions)
- Ensures compatibility across supported Go versions

**Build Workflow** (`.github/workflows/build.yml`):

- Triggers only on version tags (e.g., `v*`)
- Builds for all supported platforms
- Creates GitHub Release with binaries attached
- Uses Go 1.25 for all builds

## Testing

The project includes comprehensive tests across all packages with 30+ test cases.

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./jq
go test -v ./auth
go test -v ./cli
go test -v ./config

# Run specific test
go test -v -run TestExecuteJQ_SimpleQuery

# Run with coverage
go test -cover ./...
```

### Test Coverage

**Package: `jq` (14 test cases in `main_test.go`)**

- ✅ Basic queries: `.name`, `.age`, nested access
- ✅ Array operations: Access, mapping, filtering
- ✅ Advanced filters: `select()`, pipe operations
- ✅ Built-in functions: `keys`, `length`, `type`
- ✅ Error handling: Invalid filters, non-existent keys
- ✅ Edge cases: Empty arrays, identity filter
- ✅ Complex scenarios: Nested data structures, multiple results
- ✅ Multi-file operations: `inputs`, glob expansion
- ✅ File validation: Existence, readability, JSON validity

**Package: `auth` (10 test cases)**

- ✅ Bearer token extraction
- ✅ Token matching logic
- ✅ HTTP bearer authorization
- ✅ SSE token authorization (query param and header)
- ✅ Case-insensitive bearer scheme
- ✅ Empty token handling
- ✅ Unauthorized response generation

**Package: `cli` (3 test cases)**

- ✅ Single file query execution
- ✅ Multiple file query execution
- ✅ Glob pattern expansion and execution

**Package: `config` (4 test cases)**

- ✅ Valid YAML config loading
- ✅ Default value application
- ✅ Invalid YAML handling
- ✅ Nonexistent file error handling

**Test Infrastructure:**

- `main_test.go`: Integration tests for core functionality
- `*/.*_test.go`: Package-specific unit tests
- `tests/testdata/`: Test fixtures (valid, invalid, nested JSON)
- `examples/data/`: Real-world example data for testing

## Examples

### Basic Queries (Single File)

```jq
# Get a single field
.name

# Access nested field
.user.address.city

# Array access
.users[0].name

# Map over array
.users[] | .name

# Filter array
.users[] | select(.age > 30)

# Transform to new array
[.users[] | .email]
```

### Advanced Queries (Single File)

```jq
# Multiple filters
.users[] | select(.age > 25) | select(.active == true) | .name

# Using built-in functions
.users | length

# Get object keys
keys

# Type checking
.age | type

# Arithmetic operations
.transactions[] | .amount | add
```

### Multi-File Queries

When working with multiple files, use the `inputs` function to access all files:

```jq
# Combine all transactions from multiple files
[inputs.transactions[]]

# Sum amounts across all files
[inputs.transactions[]] | map(.amount) | add

# Filter across all files
[inputs.transactions[] | select(.amount > 100)]

# Count specific items across all files
[inputs.transactions[] | select(.category == "services")] | length

# Get unique categories across all files
[inputs.transactions[].category] | unique

# Average amount across all files
[inputs.transactions[]] | map(.amount) | add / length
```

### Real-World Examples

**CLI Mode:**

```bash
# Find all users over 30
gojq-mcp -f examples/data/sample.json -q '.users[] | select(.age > 30)'

# Total revenue for January 2025
gojq-mcp -f 'examples/data/multiple-files/2025-01/*.json' \
         -q '[inputs.transactions[]] | map(.amount) | add'

# Count service transactions across all months
gojq-mcp -f 'examples/data/multiple-files/*/*.json' \
         -q '[inputs.transactions[] | select(.category == "services")] | length'
```

**MCP Tool Mode:**

```json
{
  "tool": "run_jq",
  "arguments": {
    "jq_filter": "[inputs.transactions[]] | group_by(.category) | map({category: .[0].category, total: map(.amount) | add})",
    "file_patterns": ["multiple-files/2025-01/*.json"]
  }
}
```

### Sample Data

- `examples/data/sample.json` - Users dataset with names, ages, and emails
- `examples/data/multiple-files/2025-01/` - January transaction data
- `examples/data/multiple-files/2025-02/` - February transaction data

## Best Practices

### For MCP Server Development

1. **Use relative paths**: File patterns are relative to the configured data directory
2. **Leverage glob patterns**: Use wildcards to match multiple files efficiently
3. **Handle errors gracefully**: Check error responses from the tool
4. **Test queries incrementally**: Start with simple queries and build complexity
5. **Use the identity filter** (`.`) to inspect data structure first
6. **Enable file watching**: Get real-time notifications when data files change
7. **Secure HTTP/SSE**: Always use bearer tokens for HTTP and SSE transports

### For jq Queries

**Single-File Queries:**

1. **Start simple**: Test with `.` to see full structure
2. **Use `keys`**: Discover available fields with `keys` function
3. **Pipe operations**: Build complex queries by chaining with `|`
4. **Select carefully**: Use `select()` for filtering arrays
5. **Check types**: Use `type` function to verify data types

**Multi-File Queries:**

1. **Use `inputs` function**: Access all files in multi-file queries
2. **Wrap in arrays**: Use `[inputs.field[]]` to collect results from all files
3. **Aggregate with built-ins**: Use `add`, `unique`, `group_by` for cross-file analysis
4. **Filter before collecting**: `[inputs.items[] | select(.condition)]` is more efficient
5. **Test with small sets**: Try queries on a few files before expanding to all

### For Configuration

1. **Start with YAML config**: Use configuration files for complex setups
2. **Set custom instructions**: Guide LLM clients with relevant context
3. **Define prompts**: Create reusable prompt templates for common tasks
4. **Override with CLI**: Use CLI flags to test different configurations
5. **Version control configs**: Store configuration files in your repository

### For Integration

1. **Restart MCP client** after configuration changes
2. **Use absolute binary paths** in MCP server configuration
3. **Verify data directory**: Ensure the data path exists and contains JSON files
4. **Check logs** if the server doesn't appear in your MCP client
5. **Test with Inspector**: Use `@modelcontextprotocol/inspector` to debug MCP integration
6. **Monitor file changes**: Watch stderr for file change notifications
7. **Keep dependencies updated** with `go get -u` and `go mod tidy`

### For Security

1. **Restrict data directory**: Only expose necessary JSON files
2. **Use authentication**: Enable bearer tokens for HTTP/SSE transports
3. **Validate queries**: Be aware that jq queries can be resource-intensive
4. **Monitor access**: Log queries in production environments
5. **Keep tokens secret**: Store bearer tokens in environment variables or secure vaults

## Dependencies

- [gojq](https://github.com/itchyny/gojq) - Pure Go implementation of jq
- [mcp-go](https://github.com/mark3labs/mcp-go) - Model Context Protocol server framework
- [yaml.v3](https://github.com/go-yaml/yaml) - YAML configuration file parsing
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system change notifications
- [testify](https://github.com/stretchr/testify) - Testing toolkit with assertions

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

# gojq-mcp

A dual-mode JSON query tool that works both as an MCP (Model Context Protocol) server and a standalone CLI executable using jq syntax.

## Features

- Execute jq queries on JSON files
- Validate JSON file existence, readability, and validity
- Dual mode operation: MCP server or CLI tool

## Installation
```bash
go build -o gojq-mcp

## Usage

### As CLI Tool

```bash
# Basic usage
./gojq-mcp -file data.json -query '.users[] | select(.age > 30)'

# Short flags
./gojq-mcp -f data.json -q '.[] | .name'
```

### As MCP Server

Run without arguments to start in MCP server mode:

```bash
./gojq-mcp
```

Configure in your MCP client (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "gojq": {
      "command": "/path/to/gojq-mcp"
    }
  }
}
```

## MCP Tool Interface
Here's a recommended project tree structure for your gojq-mcp project:

```
/Users/eberry/github.com/berrydev-ai/gojq-mcp/
├── main.go                 # Main application code (as provided above)
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies checksums
├── README.md               # Project documentation
├── LICENSE                 # License file (optional)
├── .gitignore             # Git ignore file
├── Makefile               # Build automation (optional)
├── examples/              # Example JSON files and queries
│   ├── sample.json
│   └── queries.txt
├── tests/                 # Test files
│   ├── main_test.go
│   └── testdata/
│       ├── valid.json
│       ├── invalid.json
│       └── nested.json
└── dist/                  # Build output directory (gitignored)
```

**Minimal files to create:**

**`.gitignore`**
```gitignore
# Binaries
gojq-mcp
dist/
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo
*~
.DS_Store
```

**`go.mod`** (initial)
```go
module github.com/berrydev-ai/gojq-mcp

go 1.21

require (
	github.com/itchyny/gojq v0.12.14
	github.com/mark3labs/mcp-go v0.6.0
)
```

**`README.md`**
```markdown
# gojq-mcp

A dual-mode JSON query tool that works both as an MCP (Model Context Protocol) server and a standalone CLI executable using jq syntax.

## Features

- Execute jq queries on JSON files
- Validate JSON file existence, readability, and validity
- Dual mode operation: MCP server or CLI tool

## Installation

```bash
go build -o gojq-mcp
```

## Usage

### As CLI Tool

```bash
# Basic usage
./gojq-mcp -file data.json -query '.users[] | select(.age > 30)'

# Short flags
./gojq-mcp -f data.json -q '.[] | .name'
```

### As MCP Server

Run without arguments to start in MCP server mode:

```bash
./gojq-mcp
```

Configure in your MCP client (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "gojq": {
      "command": "/path/to/gojq-mcp"
    }
  }
}
```

## MCP Tool Interface

**Tool Name:** `run_jq`

**Parameters:**
- `jq_filter` (string, required): The jq filter to execute
- `json_file_path` (string, required): Absolute path to the JSON file

## Examples

See the `examples/` directory for sample JSON files and common queries.

## License

[Your chosen license]
```

**`Makefile`** (optional but useful)
```makefile
.PHONY: build clean test install

BINARY_NAME=gojq-mcp
BUILD_DIR=dist

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	rm -rf $(BUILD_DIR)
	go clean

test:
	go test -v ./...

install:
	go install

run-cli:
	go run . -f examples/sample.json -q '.'

run-server:
	go run .
```

**`examples/sample.json`**
```json
{
  "users": [
    {
      "name": "Alice",
      "age": 30,
      "email": "alice@example.com"
    },
    {
      "name": "Bob",
      "age": 35,
      "email": "bob@example.com"
    },
    {
      "name": "Charlie",
      "age": 25,
      "email": "charlie@example.com"
    }
  ],
  "total": 3
}
```

**`examples/queries.txt`**
```
# Example jq queries for sample.json

# Get all user names
.users[] | .name

# Filter users over 30
.users[] | select(.age > 30)

# Get total count
.total

# Transform to array of emails
[.users[] | .email]
```

This structure keeps it clean and professional while providing everything needed for both development and usage.


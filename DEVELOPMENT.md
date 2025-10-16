# Development Guide

This guide is for developers who want to contribute to or modify the gojq-mcp project.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Building](#building)
- [Running](#running)
- [Cleaning](#cleaning)
- [Testing](#testing)
- [Releases and Deployment](#releases-and-deployment)
- [Contributing Guidelines](#contributing-guidelines)

## Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile commands)

## Building

```bash
# Using Makefile
make build

# Using go directly
go build -o dist/gojq-mcp .
```

## Running

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

## Cleaning

```bash
make clean
```

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

## Contributing Guidelines

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

## Dependencies

- [gojq](https://github.com/itchyny/gojq) - Pure Go implementation of jq
- [mcp-go](https://github.com/mark3labs/mcp-go) - Model Context Protocol server framework
- [yaml.v3](https://github.com/go-yaml/yaml) - YAML configuration file parsing
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system change notifications
- [testify](https://github.com/stretchr/testify) - Testing toolkit with assertions

## Resources

- [jq Manual](https://jqlang.github.io/jq/manual/) - Complete jq syntax reference
- [MCP Documentation](https://modelcontextprotocol.io/) - Model Context Protocol specification
- [gojq Documentation](https://github.com/itchyny/gojq#readme) - gojq library details


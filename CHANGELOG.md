# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.3] - 2025-10-13

### Added
- `--default-json-file` flag to configure a reusable JSON file path for HTTP/SSE transports (#3)
- Automatic fallback to the configured default JSON file when `run_jq` requests omit `json_file_path` over HTTP/SSE (#3)

### Changed
- Updated CLI help and docs to cover default JSON file behaviour in server transports (#3)

## [1.0.2] - 2025-10-06

### Added
- HTTP transport support via `-t http` flag for web-based MCP integrations
- SSE (Server-Sent Events) transport support via `-t sse` flag for event-driven clients
- Transport selection flag `-t` with support for stdio (default), http, and sse transports
- Address flag `-a` for configuring listening address for http/sse transports (default: :8080)
- `--help` and `-h` flags to display comprehensive usage information
- `--version` flag to display version information
- Makefile targets for running different transports: `run-http`, `run-http-custom`, `run-sse`, `run-sse-custom`

### Changed
- Enhanced CLI with better help output showing all available modes, options, and examples
- Server startup now displays informative messages indicating transport type and listening address

## [1.0.1] - 2025-10-06

### Fixed
- Fixed module path in go.mod to include full GitHub repository path, enabling `go install github.com/berrydev-ai/gojq-mcp@latest` to work correctly

## [1.0.0] - 2025-10-06

### Added
- MCP server implementation with `run_jq` tool for jq filter execution on JSON files
- JSON file validation and error handling
- Cross-platform binary builds (Linux, macOS, Windows for amd64/arm64)
- Automated release workflow with GitHub Actions
- Comprehensive test suite
- Release scripts with changelog integration

[Unreleased]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.0.3...HEAD
[1.0.3]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/berrydev-ai/gojq-mcp/releases/tag/v1.0.0

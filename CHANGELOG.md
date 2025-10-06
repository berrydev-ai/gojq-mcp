# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/berrydev-ai/gojq-mcp/releases/tag/v1.0.0

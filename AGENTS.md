# Repository Guidelines

## Project Structure & Module Organization
- `main.go` bootstraps the MCP server and CLI path; add new packages only when reusable and wire them here.
- `main_test.go` holds the jq engine suite; extend it with table-driven cases for every behaviour change.
- `examples/` (docs demos) and `tests/testdata/` (JSON fixtures) back new scenarios; update fixtures instead of embedding literals.
- `scripts/` contains release helpers, while `dist/` is generated output that should stay gitignored between builds.

## Build, Test, and Development Commands
- `make build` / `go build -o dist/gojq-mcp .` compile the binary.
- `make run-server` starts the stdio MCP server; `make run-http` or `make run-sse` exercise the transport variants.
- `make run-cli` runs the sample CLI pipeline; update `examples/sample.json` when showcasing new filters.
- `make test` or `go test -v ./...` execute unit tests; use `go test -run TestName` for targeted debugging.
- `make run-inspector` invokes `npx @modelcontextprotocol/inspector go run .` to validate the MCP contract end-to-end.

## Coding Style & Naming Conventions
- Format code with `gofmt` (tabs, no trailing spaces); never commit unformatted files.
- Exported APIs use UpperCamelCase, internals lowerCamelCase, and package names remain short lowercase nouns.
- Guard errors inline and return early; favour standard library helpers before adding dependencies.
- Run `go vet ./...` and address warnings whenever you touch transports, file IO, or concurrency.

## Testing Guidelines
- Stick to `testing`/table-driven patterns in `main_test.go`; share inputs via `tests/testdata/` to avoid duplication.
- Mirror new behaviours with regression cases, focusing on JSON validation, jq parsing, and transport negotiation.
- Maintain or increase the current 14-scenario breadth and include `go test -cover ./...` output in PR notes when coverage shifts.

## Commit & Pull Request Guidelines
- Follow history conventions: short, imperative subjects (e.g., `Add HTTP transport support`, `Bump version to 1.0.2`) with one concern per commit.
- Update the CHANGELOG and docs alongside user-visible changes.
- PRs should outline intent, list executed commands (`make test`, manual runs), and reference issues or release tags.
- Attach CLI transcripts or JSON diffs when altering sample data or protocol wiring to help reviewers reproduce.

## Release & Configuration Tips
- Tag `vX.Y.Z` to trigger CI packaging; `scripts/release.sh` wraps the release workflow for local dry runs.
- Sanity check `dist/` binaries and refresh `examples/` whenever flags or transports change.
- Configure MCP clients with absolute binary paths and verify stdio plus HTTP/SSE endpoints before merge.

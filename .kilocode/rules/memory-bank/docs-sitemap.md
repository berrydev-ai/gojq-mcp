# MCP-Go Documentation Sitemap

This sitemap provides a comprehensive reference to the MCP-Go documentation structure, organized by topic and difficulty level. Use this to quickly locate relevant content when working with MCP servers, clients, and transports.

## Documentation Structure Overview

```
docs/mcp-go/
├── index.mdx                    # Landing page with project overview
├── getting-started.mdx          # Installation and basic concepts
├── quick-start.mdx              # Hello world examples and testing
├── core-concepts.mdx            # MCP protocol fundamentals
├── clients/                     # Client-side documentation
├── servers/                     # Server-side documentation
└── transports/                  # Transport layer documentation
```

## Quick Reference by Topic

### Getting Started (Beginner)
- **[`index.mdx`](docs/mcp-go/index.mdx)** - Project overview, features, and getting started links
- **[`getting-started.mdx`](docs/mcp-go/getting-started.mdx)** - Installation, "why MCP-Go", key features, first server example
- **[`quick-start.mdx`](docs/mcp-go/quick-start.mdx)** - Complete hello world server/client examples, testing with Claude Desktop and MCP Inspector

### Core Concepts (Intermediate)
- **[`core-concepts.mdx`](docs/mcp-go/core-concepts.mdx)** - MCP protocol fundamentals (Resources, Tools, Prompts, Transports), SDK architecture, server vs client patterns

### Server Development
#### Basics
- **[`servers/index.mdx`](docs/mcp-go/servers/index.mdx)** - Server overview, complete example with tools/resources/prompts
- **[`servers/basics.mdx`](docs/mcp-go/servers/basics.mdx)** - Server creation, configuration, capabilities

#### Advanced Features
- **[`servers/tools.mdx`](docs/mcp-go/servers/tools.mdx)** - Tool definition, handlers, parameter validation
- **[`servers/resources.mdx`](docs/mcp-go/servers/resources.mdx)** - Resource creation, dynamic/static resources
- **[`servers/prompts.mdx`](docs/mcp-go/servers/prompts.mdx)** - Prompt templates, argument handling
- **[`servers/advanced.mdx`](docs/mcp-go/servers/advanced.mdx)** - Typed tools, middleware, hooks, session management
- **[`servers/advanced-sampling.mdx`](docs/mcp-go/servers/advanced-sampling.mdx)** - Sampling callbacks, progress reporting

### Client Development
#### Basics
- **[`clients/index.mdx`](docs/mcp-go/clients/index.mdx)** - Client overview, complete examples, transport types, common patterns
- **[`clients/basics.mdx`](docs/mcp-go/clients/basics.mdx)** - Client lifecycle, initialization, error handling

#### Operations
- **[`clients/operations.mdx`](docs/mcp-go/clients/operations.mdx)** - Using tools, resources, prompts from client side
- **[`clients/transports.mdx`](docs/mcp-go/clients/transports.mdx)** - Transport-specific client implementations

#### Advanced
- **[`clients/advanced-sampling.mdx`](docs/mcp-go/clients/advanced-sampling.mdx)** - Sampling in client applications

### Transport Layer
#### Overview
- **[`transports/index.mdx`](docs/mcp-go/transports/index.mdx)** - Transport comparison, selection guide, multi-transport examples

#### Specific Transports
- **[`transports/stdio.mdx`](docs/mcp-go/transports/stdio.mdx)** - STDIO transport for CLI/desktop apps
- **[`transports/sse.mdx`](docs/mcp-go/transports/sse.mdx)** - Server-Sent Events for real-time web apps
- **[`transports/http.mdx`](docs/mcp-go/transports/http.mdx)** - StreamableHTTP for REST-like interactions
- **[`transports/inprocess.mdx`](docs/mcp-go/transports/inprocess.mdx)** - In-process transport for embedded/testing scenarios

## Key Code Examples by Topic

### Server Examples
- **Hello World Tool**: `getting-started.mdx:39-84`, `quick-start.mdx:9-70`
- **Complete Server with Tools/Resources/Prompts**: `servers/index.mdx:37-143`
- **Multi-Transport Server**: `transports/index.mdx:177-223`

### Client Examples
- **Basic STDIO Client**: `quick-start.mdx:139-205`
- **StreamableHTTP Client**: `quick-start.mdx:212-273`
- **Multi-Server Client**: `clients/index.mdx:365-437`
- **LLM Integration**: `clients/index.mdx:307-362`

### Transport Examples
- **Transport-Agnostic Server**: `transports/index.mdx:25-76`
- **Environment-Based Selection**: `transports/index.mdx:147-173`

## Common Patterns & Recipes

### Tool Creation
- Basic tool with parameters: `core-concepts.mdx:59-69`
- Tool with validation: `servers/tools.mdx`
- Typed tools: `servers/advanced.mdx`

### Resource Management
- Static resources: `core-concepts.mdx:26-32`
- Dynamic resources: `core-concepts.mdx:34-40`
- Resource handlers: `servers/resources.mdx`

### Error Handling
- Client error patterns: `clients/index.mdx:234-265`
- Retry logic: `clients/index.mdx:268-302`
- Server error handling: `servers/advanced.mdx`

### Authentication & Security
- HTTP/SSE auth: `transports/http.mdx`, `transports/sse.mdx`
- Security middleware: `transports/index.mdx:256-262`

## Testing & Debugging

### Testing Tools
- **MCP Inspector**: `quick-start.mdx:107-117` - Web interface for testing tools interactively
- **Claude Desktop Integration**: `quick-start.mdx:83-105` - Real-world testing setup

### Debug Examples
- Server startup issues: `quick-start.mdx:286-290`
- Client connection problems: `quick-start.mdx:291-295`
- Tool call failures: `quick-start.mdx:296-300`

## Performance & Production Considerations

### Transport Performance
- Latency comparison: `transports/index.mdx:228-231`
- Throughput characteristics: `transports/index.mdx:232-236`
- Memory usage patterns: `transports/index.mdx:238-241`

### Production Features
- Session management: `core-concepts.mdx:246-263`
- Multi-client support: `core-concepts.mdx:221-244`
- Error recovery: `servers/advanced.mdx`

## File Dependencies & Relationships

### Core Dependencies
- All examples depend on: `github.com/mark3labs/mcp-go/mcp`, `github.com/mark3labs/mcp-go/server`
- Client examples additionally need: `github.com/mark3labs/mcp-go/client`

### Transport-Specific Code
- STDIO: `server.ServeStdio(s)`
- HTTP: `server.NewStreamableHTTPServer(s).Start(port)`
- SSE: `server.NewSSEServer(s).Start(port)`
- In-Process: `client.NewInProcessClient(server)`

## Quick Lookup: Most Frequently Referenced Files

1. **`quick-start.mdx`** - Most practical examples, testing setup
2. **`servers/index.mdx`** - Complete server implementation guide
3. **`clients/index.mdx`** - Client patterns and integration examples
4. **`core-concepts.mdx`** - Protocol understanding and architecture
5. **`transports/index.mdx`** - Transport selection and configuration

## Version Notes
- Documentation reflects MCP-Go library usage patterns
- Examples use `github.com/mark3labs/mcp-go` package
- Compatible with Go 1.21+ (as per tech.md requirements)
- Assumes basic Go and MCP protocol knowledge for advanced sections

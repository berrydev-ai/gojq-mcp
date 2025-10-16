gojq-mcp exists to bridge the gap between powerful jq query capabilities and modern AI assistant workflows. It solves the problem of AI assistants lacking native JSON processing tools by providing a standardized MCP server that enables seamless jq query execution on JSON files.

## Problems Solved

1. **AI Assistant JSON Processing Gap**: AI assistants like Claude Desktop lack built-in tools for complex JSON manipulation, forcing users to rely on external tools or manual processing.

2. **jq Learning Curve**: While jq is powerful, its syntax can be intimidating for non-technical users who need to perform JSON queries.

3. **Integration Complexity**: Existing jq tools require command-line usage, making them difficult to integrate into AI assistant workflows.

4. **Transport Limitations**: Traditional CLI tools don't support the various transport protocols needed for different MCP client integrations.

## How It Works

The tool operates in two modes:
- **MCP Server Mode**: Integrates with MCP clients (like Claude Desktop) via stdio, HTTP, or SSE transports
- **CLI Mode**: Direct command-line execution for traditional usage

Users can execute jq queries on JSON files with comprehensive validation and error handling, getting results formatted as JSON strings.

## User Experience Goals

1. **Zero Configuration**: Works out of the box with no setup required
2. **Comprehensive Validation**: Clear error messages for file access, JSON validity, and jq syntax issues
3. **Multiple Transport Support**: Flexible integration options for different client types
4. **Full jq Compatibility**: Support for the complete jq syntax and feature set
5. **Performance**: Fast execution using native Go implementation of jq
6. **Reliability**: Well-tested core with 14+ test cases covering edge cases
7. **Cross-Platform**: Consistent experience across Linux, macOS, and Windows

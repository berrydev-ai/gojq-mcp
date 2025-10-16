# GoJQ MCP Server - Usage Guide

## Quick Start

```bash
# Create a config.yaml file (see example below)

# Start server with config file
gojq-mcp -c config.yaml

# Or use CLI flags (no config required)
gojq-mcp -p ./data -t http -a :8080 -i "Your instructions here"
```

Output example:
```
✅ Loaded configuration from config.yaml
Discovered 27 JSON files in /path/to/data
💬 Loaded 2 prompt(s)
👁 File watching enabled for /path/to/data
✅ Push notifications enabled - clients will be notified of file changes
Starting MCP server with HTTP streaming transport on :8080...
```

## Configuration File

Create a `config.yaml` file with your settings:

```yaml
# GoJQ MCP Server Configuration

# Path to the data directory. Can be overridden by the -p flag.
data_path: ./data

# Transport type: stdio, http, or sse. Default: stdio.
# Recommended: http (for push notifications)
transport: http

# Port to listen on for http/sse transports. Default: 8080.
port: 8080

# Instructions for the MCP client. Can be overridden by the -i flag.
instructions: |
  You are a helpful assistant that can query and analyze data.

  ## Getting Started
  1. Use 'list_data_files' to discover available files
  2. Use 'run_jq' to query files with jq filters

  ## Tips
  - Use glob patterns for querying multiple files
  - Use '[inputs]' to collect all files into an array

# Reusable query templates.
prompts:
  - name: query_by_date
    description: "Query data for a specific date range"
    arguments:
      - name: start_date
        description: "Start date (YYYY-MM-DD)"
        required: true
      - name: end_date
        description: "End date (YYYY-MM-DD)"
        required: true
```

The config file defines:

1. **Data Path** - Location of JSON files
2. **Transport** - Communication method (stdio, http, sse)
3. **Port** - Port for http/sse transports
4. **Instructions** - Contextual guidance for LLMs
5. **Prompts** - Reusable query templates

### Instructions

Instructions are sent to the LLM during initialization and provide context about:
- What data is available
- How files are organized
- Common use cases
- Query tips and best practices

**Override with CLI flag:**
```bash
gojq-mcp -c config.yaml -i "Custom instructions here"
```

### Prompts

Prompts provide templates for common queries:

```yaml
prompts:
  - name: analyze_calls
    description: "Analyze call data for a specific campaign"
    arguments:
      - name: campaign_id
        description: "Campaign ID to analyze"
        required: true
      - name: date_range
        description: "Date range (optional)"
        required: false
```

Prompts appear in MCP clients and help users construct queries efficiently.

## CLI Flags

### Server Mode

```bash
# Basic usage with config
gojq-mcp -c config.yaml

# Override data path
gojq-mcp -p ./my-data -c config.yaml

# Override transport
gojq-mcp -c config.yaml -t http

# Override port
gojq-mcp -c config.yaml -t http -a :9000

# Override instructions
gojq-mcp -c config.yaml -i "Custom instructions"

# With authentication
gojq-mcp -c config.yaml -t http -token "your-secret-token"

# Disable file watching
gojq-mcp -c config.yaml -watch=false
```

### CLI Query Mode

```bash
# Query a single file
gojq-mcp -f data.json -q '.users[] | select(.age > 30)'
```

## MCP Features

### 1. Tools

**`list_data_files`** - Discover available files

Returns:
```json
{
  "total_files": 27,
  "files": [
    {
      "path": "calls/2025-01-15.json",
      "modified": "2025-10-15T13:19:13Z",
      "size": 1024000
    }
  ],
  "suggested_patterns": {
    "*.json": "All JSON files in base path",
    "calls/*.json": "All 27 JSON files in calls",
    "calls/2025-*.json": "Files matching pattern"
  }
}
```

**`run_jq`** - Query files with jq filters

Parameters:
- `jq_filter` - The jq query (e.g., `'.[0:5]'`, `'[inputs] | length'`)
- `json_file_path` - Space-separated file paths or glob patterns

Examples:
```json
{
  "jq_filter": ".[0:5]",
  "json_file_path": "calls/2025-01-15.json"
}
```

```json
{
  "jq_filter": "[inputs] | length",
  "json_file_path": "calls/2025-*.json"
}
```

### 2. Prompts

Configured prompts appear in MCP clients and provide quick access to common query patterns.

## Real-Time Updates

When files change:

1. **fsnotify** detects the change
2. **500ms debounce** to batch rapid changes
3. **Rescan** discovers new/modified/deleted files
4. **MCP notification** sent: `notifications/resources/list_changed`
5. **Client** receives notification automatically
6. **Client calls** `list_data_files` for updated list

**Note:** Real-time updates require `http` or `sse` transport. The `stdio` transport does not support push notifications.

## Transport Comparison

| Feature | stdio | http | sse |
|---------|-------|------|-----|
| Push Notifications | ❌ | ✅ | ✅ |
| File Watching | ✅ | ✅ | ✅ |
| Authentication | ❌ | ✅ | ✅ |
| Recommended | Local dev | ✅ Production | Deprecated |

**Recommendation:** Use `http` transport for production deployments and when you need push notifications.

## Example Queries

### Query single file
```json
{
  "jq_filter": ".[0:10]",
  "json_file_path": "calls/2025-01-15.json"
}
```

### Query multiple files (collect into array)
```json
{
  "jq_filter": "[inputs]",
  "json_file_path": "calls/2025-01-15.json calls/2025-01-16.json"
}
```

### Query with glob pattern
```json
{
  "jq_filter": "[inputs] | length",
  "json_file_path": "calls/2025-*.json"
}
```

### Process each file separately
```json
{
  "jq_filter": "inputs | {file: input_filename, count: length}",
  "json_file_path": "calls/2025-*.json"
}
```

### Filter and aggregate
```json
{
  "jq_filter": "[inputs | .[] | select(.status == \"completed\")] | length",
  "json_file_path": "calls/2025-*.json"
}
```

### Group by field
```json
{
  "jq_filter": "[inputs | .[]] | group_by(.campaign_id) | map({campaign: .[0].campaign_id, count: length})",
  "json_file_path": "calls/2025-*.json"
}
```

### Extract specific fields
```json
{
  "jq_filter": "[inputs | .[]] | map({id, status, duration})",
  "json_file_path": "calls/2025-*.json"
}
```

## Configuration Best Practices

### Organizing Your Data

Structure your JSON files logically:
```
data/
├── calls/
│   ├── 2025-01-15.json
│   ├── 2025-01-16.json
│   └── 2025-01-17.json
├── campaigns/
│   └── active.json
└── reports/
    └── monthly/
        └── 2025-01.json
```

### Writing Effective Instructions

Good instructions should:
- Explain what data is available
- Describe file organization patterns
- Provide common query examples
- Include tips for efficient querying

Example:
```yaml
instructions: |
  This server provides access to call tracking data.

  ## Data Organization
  - calls/*.json: Daily call logs (format: YYYY-MM-DD.json)
  - campaigns/active.json: Current campaigns
  - reports/monthly/*.json: Monthly aggregated reports

  ## Common Queries
  - Get today's calls: calls/2025-01-15.json with filter '.'
  - Count calls by status: Use 'group_by(.status) | map({status: .[0].status, count: length})'
  - Find calls over 5 minutes: '.[] | select(.duration > 300)'
```

### Creating Useful Prompts

Prompts should match your common workflows:

```yaml
prompts:
  - name: daily_summary
    description: "Get summary of calls for a specific date"
    arguments:
      - name: date
        description: "Date in YYYY-MM-DD format"
        required: true

  - name: campaign_analysis
    description: "Analyze calls for a specific campaign"
    arguments:
      - name: campaign_id
        description: "Campaign ID to analyze"
        required: true
      - name: days
        description: "Number of days to analyze (default: 7)"
        required: false
```

## Security Considerations

### Authentication

For production deployments with `http` or `sse` transport:

```bash
# Set a strong bearer token
gojq-mcp -c config.yaml -t http -token "your-secure-random-token-here"
```

Clients must include the token:
```
Authorization: Bearer your-secure-random-token-here
```

### Path Security

The server enforces strict path security:
- ✅ Only files within `data_path` are accessible
- ✅ Relative paths are resolved safely
- ✅ Path traversal attempts are blocked
- ✅ Absolute paths outside data directory are rejected

## Troubleshooting

### Config not loading?
- Check YAML syntax with a validator
- Verify file path with `-c`
- Look for error messages in stderr
- Ensure proper indentation (YAML is whitespace-sensitive)

### Files not found?
- Paths in queries are relative to `data_path`
- Use `list_data_files` to see available files
- Check glob pattern syntax (use `*` for wildcards)
- Verify file extensions are `.json`

### Push notifications not working?
- Use `http` or `sse` transport (not `stdio`)
- Enable file watching with `-watch=true` (default)
- Verify client supports MCP notifications
- Check that files are actually changing

### Port already in use?
```bash
# Use a different port
gojq-mcp -c config.yaml -a :9000
```

### Instructions not appearing?
- Verify `instructions` field in config
- Use `-i` flag to override: `gojq-mcp -c config.yaml -i "New instructions"`
- Check that client supports MCP instructions

## Advanced Usage

### Multiple Data Sources

Run multiple instances on different ports:
```bash
# Calls data
gojq-mcp -p ./calls -t http -a :8080 -i "Call tracking data"

# Campaign data
gojq-mcp -p ./campaigns -t http -a :8081 -i "Campaign configuration"
```

### Custom JQ Queries

Complex aggregations:
```json
{
  "jq_filter": "[inputs | .[]] | group_by(.campaign_id) | map({campaign: .[0].campaign_id, total_calls: length, completed: [.[] | select(.status == \"completed\")] | length, avg_duration: ([.[] | .duration] | add / length)})",
  "json_file_path": "calls/2025-*.json"
}
```

Date filtering:
```json
{
  "jq_filter": "[inputs | .[]] | map(select(.created_at >= \"2025-01-15T00:00:00Z\" and .created_at < \"2025-01-16T00:00:00Z\"))",
  "json_file_path": "calls/*.json"
}
```

### Performance Tips

1. **Use specific file patterns** instead of broad globs when possible
2. **Filter early** in your jq pipeline to reduce data processing
3. **Limit results** for exploratory queries: `.[0:100]`
4. **Use `list_data_files`** to understand data structure before querying

## Dependencies

Required Go modules:
```bash
go get gopkg.in/yaml.v3
go get github.com/fsnotify/fsnotify
go get github.com/itchyny/gojq
go get github.com/mark3labs/mcp-go
```

## Version Information

```bash
# Check version
gojq-mcp --version
```

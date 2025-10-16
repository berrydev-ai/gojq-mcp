# gojq-mcp Usage Guide

Complete guide for using gojq-mcp in both MCP server mode and CLI mode with real-world examples.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration File](#configuration-file)
- [CLI Flags](#cli-flags)
- [MCP Features](#mcp-features)
- [Transport Comparison](#transport-comparison)
- [Example Queries](#example-queries)
- [Configuration Best Practices](#configuration-best-practices)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

## Quick Start

```bash
# Start server with config file
gojq-mcp -c config.yaml

# Or use CLI flags (no config required)
gojq-mcp -p ./examples/data -t http -a :8080

# CLI query mode
gojq-mcp -f ./examples/data/sample.json -q '.users[] | .name'
```

Output example:

```
✅ Loaded configuration from config.yaml
Discovered 6 JSON files in ./examples/data
💬 Loaded 2 prompt(s)
👁 File watching enabled for ./examples/data
✅ Push notifications enabled - clients will be notified of file changes
Starting MCP server with HTTP streaming transport on :8080...
```

## Configuration File

Create a `config.yaml` file with your settings:

```yaml
# gojq-mcp Server Configuration

# Path to the data directory. Can be overridden by the -p flag.
data_path: ./examples/data

# Transport type: stdio, http, or sse. Default: stdio.
# Recommended: http (for push notifications)
transport: http

# Port to listen on for http/sse transports. Default: 8080.
port: 8080

# Instructions for the MCP client. Can be overridden by the -i flag.
instructions: |
  You are a helpful assistant that can query and analyze transaction data.

  ## Getting Started
  1. Use 'list_data_files' to discover available files
  2. Use 'run_jq' to query files with jq filters

  ## Data Organization
  - sample.json: User data with names, ages, and emails
  - multiple-files/2025-01/*.json: January transaction data
  - multiple-files/2025-02/*.json: February transaction data

  ## Tips
  - Use glob patterns for querying multiple files
  - Use 'inputs' function to process multiple files together
  - Files contain transactions with: id, created_at, amount, payout, description, category

# Reusable query templates.
prompts:
  - name: query_by_date_range
    description: "Query transaction data for a specific date range"
    arguments:
      - name: start_date
        description: "Start date (YYYY-MM-DD)"
        required: true
      - name: end_date
        description: "End date (YYYY-MM-DD)"
        required: true

  - name: analyze_category
    description: "Analyze transactions by category"
    arguments:
      - name: category
        description: "Transaction category (services, software, hosting, etc.)"
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
  - name: total_revenue
    description: "Calculate total revenue for a specific month"
    arguments:
      - name: month
        description: "Month in YYYY-MM format (e.g., 2025-01)"
        required: true

  - name: filter_by_category
    description: "Get all transactions for a specific category"
    arguments:
      - name: category
        description: "Category name (services, software, hosting, domains, security)"
        required: true
```

Prompts appear in MCP clients and help users construct queries efficiently.

## CLI Flags

### Server Mode

```bash
# Basic usage with config
gojq-mcp -c config.yaml

# Override data path
gojq-mcp -p ./examples/data -c config.yaml

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
gojq-mcp -f ./examples/data/sample.json -q '.users[] | select(.age > 30)'

# Query multiple files with glob pattern
gojq-mcp -f './examples/data/multiple-files/2025-01/*.json' \
         -q '[inputs.transactions[]] | map(.amount) | add'

# Query across multiple months
gojq-mcp -f './examples/data/multiple-files/*/*.json' \
         -q '[inputs.transactions[] | select(.category == "services")] | length'
```

## MCP Features

### 1. Tools

**`list_data_files`** - Discover available files

Returns:

```json
{
  "total_files": 6,
  "files": [
    {
      "path": "sample.json",
      "modified": "2025-10-15T13:19:13Z",
      "size": 245
    },
    {
      "path": "multiple-files/2025-01/01.json",
      "modified": "2025-10-15T13:19:13Z",
      "size": 468
    }
  ],
  "suggested_patterns": {
    "*.json": "All JSON files in base path",
    "multiple-files/2025-01/*.json": "All 3 JSON files in multiple-files/2025-01",
    "multiple-files/*/*.json": "Files matching pattern"
  }
}
```

**`run_jq`** - Query files with jq filters

Parameters:

- `jq_filter` - The jq query (e.g., `.users[] | .name`, `[inputs] | length`)
- `file_patterns` - Array of file paths or glob patterns (relative to data path)

Examples:

```json
{
  "jq_filter": ".users[] | .name",
  "file_patterns": ["sample.json"]
}
```

```json
{
  "jq_filter": "[inputs.transactions[]] | map(.amount) | add",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

### 2. Prompts

Configured prompts appear in MCP clients and provide quick access to common query patterns.

## Transport Comparison

| Feature | stdio | http | sse |
|---------|-------|------|-----|
| Push Notifications | ❌ | ✅ | ✅ |
| File Watching | ✅ | ✅ | ✅ |
| Authentication | ❌ | ✅ | ✅ |
| Recommended | Local dev | ✅ Production | Deprecated |

**Recommendation:** Use `http` transport for production deployments and when you need push notifications.

## Real-Time Updates

When files change:

1. **fsnotify** detects the change
2. **500ms debounce** to batch rapid changes
3. **Rescan** discovers new/modified/deleted files
4. **MCP notification** sent: `notifications/resources/list_changed`
5. **Client** receives notification automatically
6. **Client calls** `list_data_files` for updated list

**Note:** Real-time updates require `http` or `sse` transport. The `stdio` transport does not support push notifications.

## Example Queries

### Single File Queries

**Get all user names:**

```json
{
  "jq_filter": ".users[] | .name",
  "file_patterns": ["sample.json"]
}
```

**Filter users by age:**

```json
{
  "jq_filter": ".users[] | select(.age > 30)",
  "file_patterns": ["sample.json"]
}
```

**Get user emails:**

```json
{
  "jq_filter": "[.users[] | .email]",
  "file_patterns": ["sample.json"]
}
```

**Count users:**

```json
{
  "jq_filter": ".users | length",
  "file_patterns": ["sample.json"]
}
```

### Multi-File Queries

**Collect all transactions from multiple files:**

```json
{
  "jq_filter": "[inputs.transactions[]]",
  "file_patterns": ["multiple-files/2025-01/01.json", "multiple-files/2025-01/02.json"]
}
```

**Sum amounts across all January files:**

```json
{
  "jq_filter": "[inputs.transactions[]] | map(.amount) | add",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Count all transactions:**

```json
{
  "jq_filter": "[inputs.transactions[]] | length",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Filter by category across multiple files:**

```json
{
  "jq_filter": "[inputs.transactions[] | select(.category == \"services\")]",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Count transactions by category:**

```json
{
  "jq_filter": "[inputs.transactions[] | select(.category == \"services\")] | length",
  "file_patterns": ["multiple-files/*/*.json"]
}
```

**Group by category and sum:**

```json
{
  "jq_filter": "[inputs.transactions[]] | group_by(.category) | map({category: .[0].category, count: length, total: map(.amount) | add})",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Get unique categories:**

```json
{
  "jq_filter": "[inputs.transactions[].category] | unique",
  "file_patterns": ["multiple-files/*/*.json"]
}
```

**Calculate average transaction amount:**

```json
{
  "jq_filter": "[inputs.transactions[]] | map(.amount) | add / length",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Extract specific fields:**

```json
{
  "jq_filter": "[inputs.transactions[]] | map({id, amount, category})",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Filter by amount threshold:**

```json
{
  "jq_filter": "[inputs.transactions[] | select(.amount > 100)]",
  "file_patterns": ["multiple-files/*/*.json"]
}
```

### CLI Mode Examples

**Query single file:**

```bash
gojq-mcp -f ./examples/data/sample.json -q '.users[] | select(.age > 30)'
```

**Sum January transactions:**

```bash
gojq-mcp -f './examples/data/multiple-files/2025-01/*.json' \
         -q '[inputs.transactions[]] | map(.amount) | add'
```

**Count service transactions across all months:**

```bash
gojq-mcp -f './examples/data/multiple-files/*/*.json' \
         -q '[inputs.transactions[] | select(.category == "services")] | length'
```

**Get all transaction IDs:**

```bash
gojq-mcp -f './examples/data/multiple-files/2025-01/*.json' \
         -q '[inputs.transactions[].id]'
```

**Filter transactions by date:**

```bash
gojq-mcp -f './examples/data/multiple-files/2025-01/*.json' \
         -q '[inputs.transactions[] | select(.created_at >= "2025-01-02")]'
```

## Configuration Best Practices

### Organizing Your Data

Structure your JSON files logically:

```
examples/data/
├── sample.json                    # User data
└── multiple-files/
    ├── 2025-01/
    │   ├── 01.json               # Daily transactions
    │   ├── 02.json
    │   └── 03.json
    └── 2025-02/
        ├── 01.json
        ├── 02.json
        └── 03.json
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
  This server provides access to transaction and user data.

  ## Data Organization
  - sample.json: User data with names, ages, and emails
  - multiple-files/YYYY-MM/*.json: Daily transaction logs organized by month

  ## Transaction Fields
  Each transaction contains: id, created_at, amount, payout, description, category

  ## Common Queries
  - Sum all January transactions: Use 'multiple-files/2025-01/*.json' with filter '[inputs.transactions[]] | map(.amount) | add'
  - Filter by category: Use 'select(.category == "services")'
  - Get users over 30: Use 'sample.json' with filter '.users[] | select(.age > 30)'
```

### Creating Useful Prompts

Prompts should match your common workflows:

```yaml
prompts:
  - name: monthly_revenue
    description: "Calculate total revenue for a specific month"
    arguments:
      - name: month
        description: "Month in YYYY-MM format (e.g., 2025-01)"
        required: true

  - name: category_breakdown
    description: "Get transaction breakdown by category"
    arguments:
      - name: start_date
        description: "Start date (YYYY-MM-DD)"
        required: false
      - name: end_date
        description: "End date (YYYY-MM-DD)"
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

### Query returning unexpected results?

- Test query on single file first
- Verify data structure with identity filter: `.`
- Use `keys` to see available fields
- Check jq syntax at [jq playground](https://jqplay.org/)

## Advanced Usage

### Multiple Data Sources

Run multiple instances on different ports:

```bash
# User data
gojq-mcp -p ./examples/data -t http -a :8080 -i "User and transaction data"

# Another data source
gojq-mcp -p ./other-data -t http -a :8081 -i "Other data source"
```

### Complex Aggregations

**Transaction summary by category:**

```json
{
  "jq_filter": "[inputs.transactions[]] | group_by(.category) | map({category: .[0].category, total_amount: map(.amount) | add, total_payout: map(.payout) | add, count: length})",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Date range filtering:**

```json
{
  "jq_filter": "[inputs.transactions[]] | map(select(.created_at >= \"2025-01-02T00:00:00\" and .created_at < \"2025-01-03T00:00:00\"))",
  "file_patterns": ["multiple-files/2025-01/*.json"]
}
```

**Top transactions by amount:**

```json
{
  "jq_filter": "[inputs.transactions[]] | sort_by(.amount) | reverse | .[0:5]",
  "file_patterns": ["multiple-files/*/*.json"]
}
```

### Performance Tips

1. **Use specific file patterns** instead of broad globs when possible
2. **Filter early** in your jq pipeline to reduce data processing
3. **Limit results** for exploratory queries: `.[0:100]`
4. **Use `list_data_files`** to understand data structure before querying
5. **Test queries incrementally**: Start simple and add complexity

### jq Query Best Practices

**Single-File Queries:**

1. Start with `.` to see full structure
2. Use `keys` to discover available fields
3. Pipe operations to build complexity: `.users[] | select(.age > 30) | .name`
4. Use `select()` for filtering arrays
5. Check types with `type` function

**Multi-File Queries:**

1. Use `inputs` function to access all files
2. Wrap in arrays: `[inputs.transactions[]]` to collect results
3. Aggregate with built-ins: `add`, `unique`, `group_by`
4. Filter before collecting for efficiency
5. Test with small file sets first

## Version Information

```bash
# Check version
gojq-mcp --version

# Get help
gojq-mcp -h
```

## Related Documentation

- [README.md](README.md) - Project overview and installation
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide for contributors
- [jq Manual](https://jqlang.github.io/jq/manual/) - Complete jq syntax reference

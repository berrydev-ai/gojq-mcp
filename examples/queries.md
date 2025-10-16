# Example jq queries for multi-file support

## Single file queries (backward compatible)
```
./dist/gojq-mcp -f examples/data/sample.json -q '.users[] | select(.age > 30)'
```

## Multi-file queries using glob patterns
### Query all January files
```bash
./dist/gojq-mcp -f "examples/data/multiple-files/2025-01/*.json" -q '[inputs | .transactions[] | select(.created_at >= "2025-01-01T00:00:00.000-06:00" and .created_at < "2025-02-01T00:00:00.000-06:00")] | group_by(.created_at[:10]) | map({ date: .[0].created_at[:10], total_payout: map(.payout // 0) | add })'
```

### Query specific months
```bash
./dist/gojq-mcp -f "examples/data/multiple-files/2025-01/*.json" -q '[inputs | .transactions[]] | length'
./dist/gojq-mcp -f "examples/data/multiple-files/2025-02/*.json" -q '[inputs | .transactions[]] | length'
```

### Query all files
```bash
./dist/gojq-mcp -f "examples/data/multiple-files/*/*.json" -q '[inputs | .transactions[] | select(.amount > 200)] | length'
```

## Using the pipe separator
```bash
./dist/gojq-mcp \
  -f "examples/data/multiple-files/2025-02.json|examples/data/multiple-files/2025-03.json"
  -q '[inputs | .transactions[] | select(.amount > 200)] | length'

```

## MCP server usage examples

# In MCP client (like Claude Desktop), you can now use:
```json
{
  "jq_filter": "[inputs | .transactions[] | select(.created_at >= '2025-01-01T00:00:00.000-06:00' and .created_at < '2025-02-01T00:00:00.000-06:00')] | group_by(.created_at[:10]) | map({ date: .[0].created_at[:10], total_payout: map(.payout // 0) | add })",
  "json_file_path": ["examples/data/multiple-files/2025-01/01.json", "examples/data/multiple-files/2025-01/02.json", "examples/data/multiple-files/2025-01/03.json"]
}
```

### Or with glob patterns:
```json
{
  "jq_filter": "[inputs | .transactions[]] | length",
  "json_file_path": "examples/data/multiple-files/2025-*/01.json"
}
```

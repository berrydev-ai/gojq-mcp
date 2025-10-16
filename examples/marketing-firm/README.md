# Example - Marketing Analytics Dataset

This directory contains a realistic marketing analytics dataset spanning 3 months (January-March 2025) that demonstrates the capabilities of `gojq-mcp` for querying complex, multi-file JSON data structures.

## Dataset Overview

The dataset mimics what a marketing data analyst would work with in a real company, containing:

- **36 advertising campaigns** across 6 platforms (Facebook, Google Ads, Instagram, LinkedIn, TikTok, Twitter)
- **~114K impressions** with click-through tracking
- **~1,655 conversions** totaling ~$199K in attributed value
- **~5,000 revenue transactions** across 90 days
- **5 product SKUs** with varying price points

### Directory Structure

```
data/
├── ads.json                          # 36 ads across 6 campaigns
├── impressions/
│   ├── weekly/                       # 14 weekly impression files
│   │   ├── 2024-12-30.json           # ~5-10K impressions per week
│   │   ├── 2025-01-06.json
│   │   └── ...
│   └── monthly/
│       └── 2025/                     # 3 monthly aggregates
│           ├── 01.json               # ~38K impressions
│           ├── 02.json               # ~39K impressions
│           └── 03.json               # ~37K impressions
├── conversions/
│   ├── 2025-01.json                  # 368 conversions, $40K
│   ├── 2025-02.json                  # 492 conversions, $52K
│   └── 2025-03.json                  # 795 conversions, $105K
└── revenue/
    ├── 2025-01-01.json               # Daily revenue files
    ├── 2025-01-02.json               # 20-100 transactions/day
    └── ...                           # 90 daily files total
```

### Data Schema

#### ads.json

```json
{
  "ad_id": "AD-1000",
  "campaign_id": "spring_sale_2025",
  "campaign_budget": 14898,
  "platform": "linkedin",
  "ad_type": "carousel",
  "target_audience": {
    "segments": ["retargeting", "demographic"],
    "age_range": "35-44",
    "regions": ["US-WEST", "EU-CENTRAL"]
  },
  "status": "active",
  "daily_budget": 496.6
}
```

#### impressions/*.json

```json
{
  "impression_id": "IMP-123456",
  "ad_id": "AD-1000",
  "campaign_id": "spring_sale_2025",
  "timestamp": "2025-01-15T14:30:00",
  "platform": "linkedin",
  "device_type": "mobile",
  "region": "US-WEST",
  "clicked": true,
  "cost_per_impression": 0.85,
  "engagement_time_seconds": 12
}
```

#### conversions/*.json

```json
{
  "conversion_id": "CONV-123456",
  "ad_id": "AD-1000",
  "campaign_id": "spring_sale_2025",
  "timestamp": "2025-01-15T15:00:00",
  "conversion_type": "purchase",
  "value": 299.99,
  "attribution": {
    "model": "last_click",
    "touchpoints": 3
  },
  "device_type": "mobile",
  "region": "US-WEST"
}
```

#### revenue/*.json

```json
{
  "transaction_id": "TXN-123456",
  "timestamp": "2025-01-15T15:05:00",
  "attributed_ad_id": "AD-1000",
  "attributed_campaign_id": "spring_sale_2025",
  "product": {
    "id": "PROD-001",
    "name": "Premium Widget",
    "price": 299.99,
    "quantity": 1
  },
  "total_amount": 299.99,
  "discount_applied": 0,
  "region": "US-WEST",
  "customer_type": "new"
}
```

## Running the MCP Server

### Quick Start

1. **Build the server** (from repository root):

   ```bash
   make build
   # or
   go build -o ./dist/gojq-mcp .
   ```

   Ensure that `dist/gojq-mcp` is in your path to run tests below.

   Finally, cd into the `examples/marketing-firm` directory.

   ```bash
   cd examples/marketing-firm
   ```

2. **Test with CLI mode**:

   ```bash
   # Query all campaigns
   gojq-mcp -f ./data/ads.json -q '[.[].campaign_id] | unique'

   # Get January impression count
   gojq-mcp -f '.data/impressions/monthly/2025/01.json' -q 'length'

   # Total revenue for a specific day
   gojq-mcp -f ./data/revenue/2025-01-15.json \
     -q 'map(.total_amount) | add'
   ```

3. **Start as MCP server**:

   ```bash
   # stdio mode (for Claude Desktop)
   gojq-mcp -p .data

   # HTTP mode (for web clients)
   gojq-mcp -t http -p ./data -a :8080

   # HTTP mode with configuration file
   gojq-mcp -c ./config.yml
   ```

### Configuration for Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "marketing-analytics": {
      "command": "/absolute/path/to/dist/gojq-mcp",
      "args": ["-p", "/absolute/path/to/case-study-data"]
    }
  }
}
```

And if you are using the streaming-http server mode:

```json
{
  "mcpServers": {
    "marketing-analytics-http": {
      "command": "npx",
      "args": [
        "-y",
        "http://localhost:8080/mcp",
        "-c",
        "/absolute/path/to/examples/marketing-firm/config.yml",
        "--allow-http"
      ]
    }
  }
}
```

**Important:** Use absolute paths in the Claude Desktop configuration.

### Using a Configuration File

Create `config.yaml`:

```yaml
data_path: ./data
transport: http
port: 8080
instructions: |
  This server provides access to marketing analytics data spanning Q1 2025.

  Dataset includes:
  - ads.json: 36 ads across 6 campaigns and 6 platforms
  - impressions/: Weekly and monthly impression data with CTR tracking
  - conversions/: Monthly conversion data with attribution models
  - revenue/: Daily transaction data with product details

  Use 'inputs' for multi-file queries. Files support glob patterns.

  Common patterns:
  - impressions/weekly/*.json - All weekly impression files
  - impressions/monthly/2025/*.json - Monthly aggregates
  - conversions/2025-*.json - All conversion files
  - revenue/2025-01-*.json - January revenue files

prompts:
  - name: campaign_performance
    description: "Analyze campaign performance metrics"
    arguments:
      - name: campaign_id
        description: "Campaign to analyze (e.g., spring_sale_2025)"
        required: true

  - name: monthly_summary
    description: "Generate monthly marketing summary"
    arguments:
      - name: month
        description: "Month in YYYY-MM format (e.g., 2025-01)"
        required: true

  - name: roi_analysis
    description: "Calculate ROI for campaigns"
    arguments:
      - name: time_period
        description: "Time period (weekly, monthly, or date range)"
        required: true
```

Start with config:

```bash
gojq-mcp -c ./config.yml
```

## Example Queries

### Single-File Queries

#### Campaign Analysis

```bash
# Get all campaign names
gojq-mcp -f ./case-study-data/ads.json \
  -q '[.[].campaign_id] | unique'

# Active ads by platform
gojq-mcp -f ./case-study-data/ads.json \
  -q 'group_by(.platform) | map({platform: .[0].platform, count: length})'

# Total campaign budgets
gojq-mcp -f ./case-study-data/ads.json \
  -q 'group_by(.campaign_id) | map({campaign: .[0].campaign_id, budget: .[0].campaign_budget})'

# Facebook ads targeting specific demographics
gojq-mcp -f ./case-study-data/ads.json \
  -q '[.[] | select(.platform == "facebook" and .target_audience.age_range == "25-34")]'
```

#### Impression Analysis

```bash
# Click-through rate for a specific week
gojq-mcp -f ./case-study-data/impressions/weekly/2025-01-13.json \
  -q '([.[] | select(.clicked == true)] | length) as $clicks | length as $total | ($clicks / $total * 100)'

# Impressions by device type
gojq-mcp -f ./case-study-data/impressions/weekly/2025-01-13.json \
  -q 'group_by(.device_type) | map({device: .[0].device_type, count: length})'

# Average engagement time for engaged users
gojq-mcp -f ./case-study-data/impressions/weekly/2025-01-13.json \
  -q '[.[] | select(.engagement_time_seconds > 0)] | map(.engagement_time_seconds) | add / length'
```

#### Conversion Analysis

```bash
# Conversions by type
gojq-mcp -f ./case-study-data/conversions/2025-01.json \
  -q 'group_by(.conversion_type) | map({type: .[0].conversion_type, count: length})'

# Total conversion value
gojq-mcp -f ./case-study-data/conversions/2025-01.json \
  -q 'map(.value) | add'

# Attribution model breakdown
gojq-mcp -f ./case-study-data/conversions/2025-01.json \
  -q 'group_by(.attribution.model) | map({model: .[0].attribution.model, count: length, avg_touchpoints: (map(.attribution.touchpoints) | add / length)})'
```

#### Revenue Analysis

```bash
# Daily revenue total
gojq-mcp -f ./case-study-data/revenue/2025-01-15.json \
  -q 'map(.total_amount) | add'

# Revenue by product
gojq-mcp -f ./case-study-data/revenue/2025-01-15.json \
  -q 'group_by(.product.id) | map({product: .[0].product.name, revenue: (map(.total_amount) | add), units: (map(.product.quantity) | add)})'

# New customer revenue percentage
gojq-mcp -f ./case-study-data/revenue/2025-01-15.json \
  -q 'group_by(.customer_type) | map({type: .[0].customer_type, revenue: (map(.total_amount) | add)}) | (.[0].revenue / ([.[].revenue] | add) * 100)'
```

### Multi-File Queries

#### Cross-Week Analysis

```bash
# Total impressions for January
gojq-mcp -f 'case-study-data/impressions/weekly/2025-01-*.json' \
  -q '[inputs] | add | length'

# Average CTR across all weeks in January
gojq-mcp -f 'case-study-data/impressions/weekly/2025-01-*.json' \
  -q '[inputs] | add | ([.[] | select(.clicked == true)] | length) as $clicks | length as $total | ($clicks / $total * 100)'

# Total ad spend for January
gojq-mcp -f 'case-study-data/impressions/weekly/2025-01-*.json' \
  -q '[inputs] | add | map(.cost_per_impression) | add'
```

#### Monthly Aggregations

```bash
# Total conversions for Q1
gojq-mcp -f 'case-study-data/conversions/2025-*.json' \
  -q '[inputs] | add | length'

# Monthly conversion values
gojq-mcp -f 'case-study-data/conversions/2025-*.json' \
  -q '[inputs] | to_entries | map({month: (.key + 1), conversions: .value | length, total_value: (.value | map(.value) | add)})'

# Q1 revenue total
gojq-mcp -f 'case-study-data/revenue/2025-*.json' \
  -q '[inputs] | add | map(.total_amount) | add'
```

#### Cross-File Joins

```bash
# Ad performance: Combine ads with impressions
gojq-mcp \
  -f ./case-study-data/ads.json \
  -f 'case-study-data/impressions/weekly/2025-01-13.json' \
  -q 'first as $ads | last | group_by(.ad_id) | map({ad_id: .[0].ad_id, campaign: .[0].campaign_id, impressions: length, clicks: ([.[] | select(.clicked == true)] | length), ad_name: ($ads[] | select(.ad_id == .[0].ad_id) | .ad_name)})'

# Campaign ROI: Combine impressions with conversions
gojq-mcp \
  -f 'case-study-data/impressions/monthly/2025/01.json' \
  -f ./case-study-data/conversions/2025-01.json \
  -q 'first as $impr | last | group_by(.campaign_id) | map({campaign: .[0].campaign_id, conversions: length, conversion_value: (map(.value) | add), impressions: ($impr | [.[] | select(.campaign_id == .[0].campaign_id)] | length), ad_spend: ($impr | [.[] | select(.campaign_id == .[0].campaign_id)] | map(.cost_per_impression) | add)}) | map(. + {roi: ((.conversion_value / .ad_spend) - 1) * 100})'
```

### Advanced Analytics

#### Cohort Analysis

```bash
# Daily new vs returning customer revenue
gojq-mcp -f 'case-study-data/revenue/2025-01-*.json' \
  -q '[inputs] | group_by(.[0].timestamp | split("T")[0]) | map({date: .[0][0].timestamp | split("T")[0], new_revenue: ([.[] | .[] | select(.customer_type == "new")] | map(.total_amount) | add), returning_revenue: ([.[] | .[] | select(.customer_type == "returning")] | map(.total_amount) | add)})'
```

#### Funnel Metrics

```bash
# Simplified funnel: Impressions -> Clicks -> Conversions
gojq-mcp \
  -f 'case-study-data/impressions/monthly/2025/01.json' \
  -f ./case-study-data/conversions/2025-01.json \
  -q 'first as $impr | {impressions: ($impr | length), clicks: ($impr | [.[] | select(.clicked == true)] | length)} | . + {conversions: (last | length)} | . + {ctr: (.clicks / .impressions * 100), conversion_rate: (.conversions / .clicks * 100)}'
```

#### Time-Series Analysis

```bash
# Hourly revenue distribution
gojq-mcp -f ./case-study-data/revenue/2025-01-15.json \
  -q 'map(.timestamp | split("T")[1] | split(":")[0] | tonumber) | group_by(.) | map({hour: .[0], transactions: length}) | sort_by(.hour)'

# Week-over-week impression growth
gojq-mcp -f 'case-study-data/impressions/weekly/*.json' \
  -q '[inputs] | group_by(.[0].timestamp | split("T")[0] | split("-") | .[0:2] | join("-")) | map({week: (.[0][0].timestamp | split("T")[0]), impressions: ([.[] | .[]] | length)})'
```

## Usage in MCP Client (Claude)

Once configured in Claude Desktop, you can ask questions like:

**Campaign Performance:**
> "What are the top 3 campaigns by impression volume in January 2025?"

**ROI Analysis:**
> "Calculate the ROI for each campaign in February. Show me campaigns with ROI above 200%."

**Conversion Analysis:**
> "What's the average number of touchpoints before conversion for purchases vs signups?"

**Revenue Insights:**
> "Which product generated the most revenue in Q1? Show me daily trends."

**Cross-Platform:**
> "Compare click-through rates across different platforms for the spring_sale_2025 campaign."

**Audience Targeting:**
> "Which age range and device type combination has the highest conversion rate?"

The MCP server will automatically generate appropriate jq queries to answer these questions using the marketing dataset.

## Query Tips

1. **Start simple**: Use `.` to inspect data structure before building complex queries
2. **Use `keys`**: Discover available fields with `.[] | keys`
3. **Test incrementally**: Build queries step-by-step, testing each pipe operator
4. **Leverage `inputs`**: For multi-file queries, always use `inputs` function
5. **Aggregate efficiently**: Use `group_by()`, `unique`, `add`, and `length` for analytics
6. **Filter early**: Apply `select()` before collecting results for better performance
7. **Format output**: Use `-r` flag for raw output, or format with `@csv`, `@json`

## Generating Fresh Data

To regenerate the dataset with different parameters or time ranges:

```bash
cd ./case-study-data
python3 scripts/generate_example_data.py
```

Edit the script to customize:

- Date ranges (default: 3 months)
- Campaign names and budgets
- Impression volumes
- Conversion rates
- Product catalogs
- Geographic regions

## Additional Resources

- [jq Manual](https://jqlang.github.io/jq/manual/) - Complete jq syntax reference
- [gojq Documentation](https://github.com/itchyny/gojq) - Go implementation details
- [../USAGE_GUIDE.md](./USAGE_GUIDE.md) - Comprehensive gojq-mcp usage guide

## Troubleshooting

**"file does not exist" errors:**

- Ensure you're in the repository root when running commands
- Use absolute paths in Claude Desktop configuration
- Verify data files exist: `ls -la ./case-study-data/`

**"no files found matching the provided patterns":**

- Check glob pattern syntax: `'pattern/*.json'` (quotes required)
- Verify files exist: `ls ./case-study-data/impressions/weekly/`
- Ensure patterns are relative to data path

**Invalid jq syntax errors:**

- Test queries with plain `jq` first if available
- Check bracket/brace matching
- Verify pipe operators are used correctly
- Use `gojq` syntax (slight differences from standard jq)

**MCP server not appearing in Claude:**

- Restart Claude Desktop after config changes
- Check absolute paths are used in config
- Verify binary has execute permissions: `chmod +x dist/gojq-mcp`
- Check Claude Desktop logs for errors

## Contributing

Found an issue or have a suggestion for improving the example dataset? Please open an issue or submit a pull request to the [gojq-mcp repository](https://github.com/berrydev-ai/gojq-mcp).

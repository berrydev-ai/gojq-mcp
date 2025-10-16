# Makefile for examples

# Variables
GOJQ_MCP=../dist/gojq-mcp
DATA_DIR=./data
CONFIG_FILE=config.yaml

# ------------------------------------------------------------
# CLI mode
# ------------------------------------------------------------

# Run the sample CLI mode with the default config
.PHONY: run-cli-sample
run-cli-sample:
	$(GOJQ_MCP) -f $(DATA_DIR)/sample.json -q '.'

# Run the sample CLI mode with the custom config
.PHONY: run-cli-sample-with-config
run-cli-sample-with-config:
	$(GOJQ_MCP) -f $(DATA_DIR)/sample.json -q '.' -c $(CONFIG_FILE)

# ------------------------------------------------------------
# HTTP server mode
# ------------------------------------------------------------

# Run the streaming http server mode with the default config
.PHONY: run-http-server
run-http-server:
	$(GOJQ_MCP) -t http -p $(DATA_DIR)

# Run the SSE server mode with the custom config
.PHONY: run-http-server-with-config-and-auth
run-http-server-with-config-and-auth:
	$(GOJQ_MCP) -t http -c $(CONFIG_FILE)

# ------------------------------------------------------------
# SSE server mode - This transport is deprecated,
#                   Consider using 'http' transport instead
# ------------------------------------------------------------

# Run the SSE server mode with the default config
.PHONY: run-sse-server
run-sse-server:
	$(GOJQ_MCP) -t sse -p $(DATA_DIR)

# Run the SSE server mode with the custom config
.PHONY: run-sse-server-with-config
run-sse-server-with-config:
	$(GOJQ_MCP) -t sse -c $(CONFIG_FILE)

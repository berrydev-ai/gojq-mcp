# Makefile for examples

# Variables
DATA_DIR=./data
CONFIG_FILE=config.yaml

# ------------------------------------------------------------
# CLI mode
# ------------------------------------------------------------

# Run the sample CLI mode with the default config
.PHONY: run-cli-sample
run-cli-sample:
	go run ../main.go -f $(DATA_DIR)/sample.json -q '.'

# Run the sample CLI mode with the custom config
.PHONY: run-cli-sample-with-config
run-cli-sample-with-config:
	go run ../main.go -f $(DATA_DIR)/sample.json -q '.' -c $(CONFIG_FILE)

# ------------------------------------------------------------
# HTTP server mode
# ------------------------------------------------------------

# Run the streaming http server mode with the default config
.PHONY: run-http-server
run-http-server:
	go run ../main.go -t http -p $(DATA_DIR)

# Run the SSE server mode with the custom config
.PHONY: run-http-server-with-config
run-http-server-with-config:
	go run ../main.go -t http -c $(CONFIG_FILE)

# ------------------------------------------------------------
# SSE server mode - This transport is deprecated,
#                   Consider using 'http' transport instead
# ------------------------------------------------------------

# Run the SSE server mode with the default config
.PHONY: run-sse-server
run-sse-server:
	go run ../main.go -t sse -p $(DATA_DIR)

# Run the SSE server mode with the custom config
.PHONY: run-sse-server-with-config
run-sse-server-with-config:
	go run ../main.go -t sse -c $(CONFIG_FILE)

.PHONY: build clean test install run-cli run-build run-inspector run-server run-http run-http-custom run-sse run-sse-custom

BINARY_NAME=gojq-mcp
BUILD_DIR=dist

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	rm -rf $(BUILD_DIR)
	go clean

test:
	go test -v ./...

install:
	go install

run-cli:
	go run . -f examples/sample.json -q '.'

run-build: build
	./dist/gojq-mcp -f ./examples/sample.json -q '.'

run-inspector: build
	npx @modelcontextprotocol/inspector go run .

run-server:
	go run .

run-http:
	go run . -t http

run-http-custom:
	go run . -t http -a :9000

run-sse:
	go run . -t sse

run-sse-custom:
	go run . -t sse -a :9000
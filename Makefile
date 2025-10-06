.PHONY: build clean test install

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
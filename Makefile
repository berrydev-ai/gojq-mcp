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
	go run . -f ./examples/data/sample.json -q '.'

run-cli-build: build
	./dist/gojq-mcp -f ./examples/data/sample.json -q '.'

run-inspector: build
	DANGEROUSLY_OMIT_AUTH=true npx @modelcontextprotocol/inspector go run .

run-server:
	go run . -p ./examples/data

run-http:
	go run . -c ./config.sample.yml

run-http-custom:
	go run . -t http -a :9000

run-sse:
	go run . -t sse

run-sse-custom:
	go run . -t sse -a :9000

run-marketing-firm-example: build
	cd examples/marketing-firm && ../../dist/gojq-mcp -c ./config.yml

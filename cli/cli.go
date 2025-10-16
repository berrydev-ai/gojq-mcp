package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/berrydev-ai/gojq-mcp/jq"
)

// RunCLIMode executes jq query on a single file in CLI mode
func RunCLIMode(filePath, query string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	result, err := jq.ExecuteJQ(query, jsonData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing jq query: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

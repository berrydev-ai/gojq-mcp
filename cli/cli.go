package cli

import (
	"fmt"
	"os"

	"github.com/berrydev-ai/gojq-mcp/jq"
)

// RunCLIMode executes jq query on files in CLI mode
func RunCLIMode(filePaths []string, query string) {
	if len(filePaths) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no file paths provided\n")
		os.Exit(1)
	}

	// Expand glob patterns in file paths
	expandedPaths, err := jq.ExpandGlobPatterns(filePaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error expanding glob patterns: %v\n", err)
		os.Exit(1)
	}

	if len(expandedPaths) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no files found matching the provided patterns\n")
		os.Exit(1)
	}

	// Validate and read all JSON files
	jsonDataList, err := jq.ValidateAndReadJSONFiles(expandedPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading files: %v\n", err)
		os.Exit(1)
	}

	// Execute query based on number of files
	var result string
	if len(jsonDataList) == 1 {
		result, err = jq.ExecuteJQ(query, jsonDataList[0])
	} else {
		result, err = jq.ExecuteJQMultiFiles(query, jsonDataList)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing jq query: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

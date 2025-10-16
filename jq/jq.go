package jq

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
)

// ExpandGlobPatterns expands glob patterns into file paths
func ExpandGlobPatterns(patterns []string) ([]string, error) {
	var expandedPaths []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("error expanding glob pattern %q: %w", pattern, err)
		}
		expandedPaths = append(expandedPaths, matches...)
	}

	seen := make(map[string]bool)
	var uniquePaths []string
	for _, path := range expandedPaths {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}
	sort.Strings(uniquePaths)

	return uniquePaths, nil
}

// ValidateAndReadJSONFiles validates and reads JSON files
func ValidateAndReadJSONFiles(filePaths []string) ([]interface{}, error) {
	var jsonData []interface{}

	for _, filePath := range filePaths {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file does not exist: %s", filePath)
			}
			return nil, fmt.Errorf("error accessing file %s: %w", filePath, err)
		}

		if fileInfo.IsDir() {
			return nil, fmt.Errorf("path is a directory, not a file: %s", filePath)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("file %s is not readable: %w", filePath, err)
		}

		var parsedData interface{}
		if err := json.Unmarshal(data, &parsedData); err != nil {
			return nil, fmt.Errorf("file %s does not contain valid JSON: %w", filePath, err)
		}

		jsonData = append(jsonData, parsedData)
	}

	return jsonData, nil
}

// ExecuteJQ executes a jq filter on a single JSON data object
func ExecuteJQ(jqFilter string, jsonData interface{}) (string, error) {
	query, err := gojq.Parse(jqFilter)
	if err != nil {
		return "", fmt.Errorf("invalid jq filter: %w", err)
	}

	iter := query.Run(jsonData)
	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return "", fmt.Errorf("jq execution error: %w", err)
		}
		results = append(results, v)
	}

	var output []byte
	if len(results) == 1 {
		output, err = json.MarshalIndent(results[0], "", "  ")
	} else {
		output, err = json.MarshalIndent(results, "", "  ")
	}

	if err != nil {
		return "", fmt.Errorf("error formatting results: %w", err)
	}

	return string(output), nil
}

// ExecuteJQMultiFiles executes a jq filter on multiple JSON data objects
func ExecuteJQMultiFiles(jqFilter string, jsonData []interface{}) (string, error) {
	query, err := gojq.Parse(jqFilter)
	if err != nil {
		return "", fmt.Errorf("invalid jq filter: %w", err)
	}

	inputIter := gojq.NewIter(jsonData...)

	code, err := gojq.Compile(query, gojq.WithInputIter(inputIter))
	if err != nil {
		return "", fmt.Errorf("failed to compile jq query: %w", err)
	}

	iter := code.Run(nil)
	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return "", fmt.Errorf("jq execution error: %w", err)
		}
		results = append(results, v)
	}

	var output []byte
	if len(results) == 1 {
		output, err = json.MarshalIndent(results[0], "", "  ")
	} else {
		output, err = json.MarshalIndent(results, "", "  ")
	}

	if err != nil {
		return "", fmt.Errorf("error formatting results: %w", err)
	}

	return string(output), nil
}

// ProcessJQQuery processes a jq query on files specified by patterns
func ProcessJQQuery(jqFilter string, patterns []string, dataPath string) (string, error) {
	if len(patterns) == 0 {
		return "", fmt.Errorf("no file patterns provided")
	}

	// Convert relative paths to absolute paths
	absolutePatterns := make([]string, len(patterns))
	for i, pattern := range patterns {
		if !filepath.IsAbs(pattern) {
			absolutePatterns[i] = filepath.Join(dataPath, pattern)
		} else {
			// Security check: verify absolute path is within data directory
			absDataPath, _ := filepath.Abs(dataPath)
			absPattern, _ := filepath.Abs(pattern)
			if !strings.HasPrefix(absPattern, absDataPath) {
				return "", fmt.Errorf("access denied: path %s is outside data directory", pattern)
			}
			absolutePatterns[i] = pattern
		}
	}

	expandedPaths, err := ExpandGlobPatterns(absolutePatterns)
	if err != nil {
		return "", fmt.Errorf("error expanding glob patterns: %w", err)
	}

	if len(expandedPaths) == 0 {
		return "", fmt.Errorf("no files found matching the provided patterns")
	}

	jsonDataList, err := ValidateAndReadJSONFiles(expandedPaths)
	if err != nil {
		return "", err
	}

	var results string
	if len(jsonDataList) == 1 {
		results, err = ExecuteJQ(jqFilter, jsonDataList[0])
	} else {
		results, err = ExecuteJQMultiFiles(jqFilter, jsonDataList)
	}

	if err != nil {
		return "", err
	}

	return results, nil
}

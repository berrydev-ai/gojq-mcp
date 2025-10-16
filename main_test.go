package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/berrydev-ai/gojq-mcp/jq"
)

func TestExecuteJQ_SimpleQuery(t *testing.T) {
	jsonData := map[string]interface{}{
		"name": "John Doe",
		"age":  float64(30),
	}

	result, err := jq.ExecuteJQ(".name", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"John Doe"`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExecuteJQ_NumericQuery(t *testing.T) {
	jsonData := map[string]interface{}{
		"name": "John Doe",
		"age":  float64(30),
	}

	result, err := jq.ExecuteJQ(".age", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "30"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExecuteJQ_ArrayAccess(t *testing.T) {
	jsonData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		},
	}

	result, err := jq.ExecuteJQ(".users[0].name", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"Alice"`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExecuteJQ_ArrayMap(t *testing.T) {
	jsonData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": float64(25)},
			map[string]interface{}{"name": "Bob", "age": float64(30)},
		},
	}

	result, err := jq.ExecuteJQ(".users[].name", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Multiple results should be returned as an array
	var resultArray []interface{}
	if err := json.Unmarshal([]byte(result), &resultArray); err != nil {
		t.Fatalf("failed to parse result as array: %v", err)
	}

	if len(resultArray) != 2 {
		t.Errorf("expected 2 results, got %d", len(resultArray))
	}

	if resultArray[0] != "Alice" {
		t.Errorf("expected first result to be 'Alice', got %v", resultArray[0])
	}
	if resultArray[1] != "Bob" {
		t.Errorf("expected second result to be 'Bob', got %v", resultArray[1])
	}
}

func TestExecuteJQ_NestedAccess(t *testing.T) {
	jsonData := map[string]interface{}{
		"user": map[string]interface{}{
			"address": map[string]interface{}{
				"city": "Springfield",
			},
		},
	}

	result, err := jq.ExecuteJQ(".user.address.city", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"Springfield"`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExecuteJQ_FilterWithSelect(t *testing.T) {
	jsonData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": float64(25)},
			map[string]interface{}{"name": "Bob", "age": float64(30)},
			map[string]interface{}{"name": "Charlie", "age": float64(35)},
		},
	}

	result, err := jq.ExecuteJQ(".users[] | select(.age > 28) | .name", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resultArray []interface{}
	if err := json.Unmarshal([]byte(result), &resultArray); err != nil {
		t.Fatalf("failed to parse result as array: %v", err)
	}

	if len(resultArray) != 2 {
		t.Errorf("expected 2 results, got %d", len(resultArray))
	}
}

func TestExecuteJQ_KeysQuery(t *testing.T) {
	jsonData := map[string]interface{}{
		"name": "John",
		"age":  float64(30),
		"city": "NYC",
	}

	result, err := jq.ExecuteJQ("keys", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var keys []string
	if err := json.Unmarshal([]byte(result), &keys); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

func TestExecuteJQ_InvalidFilter(t *testing.T) {
	jsonData := map[string]interface{}{
		"name": "John Doe",
	}

	_, err := jq.ExecuteJQ(".[invalid", jsonData)
	if err == nil {
		t.Fatal("expected error for invalid filter, got nil")
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestExecuteJQ_NonExistentKey(t *testing.T) {
	jsonData := map[string]interface{}{
		"name": "John Doe",
	}

	result, err := jq.ExecuteJQ(".nonexistent", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "null"
	if result != expected {
		t.Errorf("expected %s for non-existent key, got %s", expected, result)
	}
}

func TestExecuteJQ_EmptyArray(t *testing.T) {
	jsonData := map[string]interface{}{
		"items": []interface{}{},
	}

	result, err := jq.ExecuteJQ(".items", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[]"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExtractBearerToken(t *testing.T) {
	testCases := []struct {
		name      string
		header    string
		expected  string
		expectsOK bool
	}{
		{
			name:      "valid header",
			header:    "Bearer secret",
			expected:  "secret",
			expectsOK: true,
		},
		{
			name:      "lowercase prefix",
			header:    "bearer secret",
			expected:  "secret",
			expectsOK: true,
		},
		{
			name:      "missing scheme",
			header:    "secret",
			expected:  "",
			expectsOK: false,
		},
		{
			name:      "wrong scheme",
			header:    "Basic secret",
			expected:  "",
			expectsOK: false,
		},
		{
			name:      "empty token",
			header:    "Bearer ",
			expected:  "",
			expectsOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, ok := auth.ExtractBearerToken(tc.header)
			if ok != tc.expectsOK {
				t.Fatalf("expected ok=%v, got %v", tc.expectsOK, ok)
			}
			if token != tc.expected {
				t.Fatalf("expected token %q, got %q", tc.expected, token)
			}
		})
	}
}

func TestAuthorizeHTTPBearer(t *testing.T) {
	const expectedToken = "secret-token"

	testCases := []struct {
		name   string
		header string
		wantOK bool
	}{
		{
			name:   "matching token",
			header: "Bearer secret-token",
			wantOK: true,
		},
		{
			name:   "case insensitive scheme",
			header: "bearer secret-token",
			wantOK: true,
		},
		{
			name:   "wrong token",
			header: "Bearer other",
			wantOK: false,
		},
		{
			name:   "missing header",
			header: "",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://example.com/mcp", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			if got := auth.AuthorizeHTTPBearer(expectedToken, req); got != tc.wantOK {
				t.Fatalf("expected %v, got %v", tc.wantOK, got)
			}
		})
	}
}

func TestAuthorizeSSEToken(t *testing.T) {
	const expectedToken = "secret-token"

	testCases := []struct {
		name       string
		method     string
		queryToken string
		header     string
		wantOK     bool
	}{
		{
			name:       "matching query token",
			method:     http.MethodGet,
			queryToken: expectedToken,
			wantOK:     true,
		},
		{
			name:       "wrong query token",
			method:     http.MethodGet,
			queryToken: "wrong",
			wantOK:     false,
		},
		{
			name:   "header fallback",
			method: http.MethodPost,
			header: "Bearer secret-token",
			wantOK: true,
		},
		{
			name:   "missing token",
			method: http.MethodPost,
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "http://example.com/mcp", nil)
			if tc.queryToken != "" {
				values := req.URL.Query()
				values.Set("token", tc.queryToken)
				req.URL.RawQuery = values.Encode()
			}
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			if got := auth.AuthorizeSSEToken(expectedToken, req); got != tc.wantOK {
				t.Fatalf("expected %v, got %v", tc.wantOK, got)
			}
		})
	}
}

func TestExecuteJQ_IdentityFilter(t *testing.T) {
	jsonData := map[string]interface{}{
		"name": "John Doe",
		"age":  float64(30),
	}

	result, err := jq.ExecuteJQ(".", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed["name"] != "John Doe" {
		t.Errorf("expected name to be 'John Doe', got %v", parsed["name"])
	}
	if parsed["age"] != float64(30) {
		t.Errorf("expected age to be 30, got %v", parsed["age"])
	}
}

func TestExecuteJQ_ComplexNestedData(t *testing.T) {
	jsonStr := `{
		"users": [
			{
				"id": 1,
				"name": "Alice",
				"address": {
					"street": "123 Main St",
					"city": "Springfield",
					"zip": "12345"
				}
			},
			{
				"id": 2,
				"name": "Bob",
				"address": {
					"street": "456 Oak Ave",
					"city": "Riverside",
					"zip": "67890"
				}
			}
		],
		"total": 2
	}`

	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		t.Fatalf("failed to parse test JSON: %v", err)
	}

	result, err := jq.ExecuteJQ(".users[].address.city", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cities []string
	if err := json.Unmarshal([]byte(result), &cities); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(cities) != 2 {
		t.Errorf("expected 2 cities, got %d", len(cities))
	}
	if cities[0] != "Springfield" {
		t.Errorf("expected first city to be 'Springfield', got %s", cities[0])
	}
	if cities[1] != "Riverside" {
		t.Errorf("expected second city to be 'Riverside', got %s", cities[1])
	}
}

func TestExecuteJQ_LengthQuery(t *testing.T) {
	jsonData := map[string]interface{}{
		"items": []interface{}{1, 2, 3, 4, 5},
	}

	result, err := jq.ExecuteJQ(".items | length", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "5"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExecuteJQ_TypeQuery(t *testing.T) {
	jsonData := map[string]interface{}{
		"name": "John",
		"age":  float64(30),
	}

	result, err := jq.ExecuteJQ(".age | type", jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"number"`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExpandGlobPatterns(t *testing.T) {
	testCases := []struct {
		name     string
		patterns []string
		expected []string
	}{
		{
			name:     "no patterns",
			patterns: []string{},
			expected: []string{},
		},
		{
			name:     "single file",
			patterns: []string{"examples/sample.json"},
			expected: []string{"examples/sample.json"},
		},
		{
			name:     "glob pattern",
			patterns: []string{"examples/multiple-files/2025-01/*.json"},
			expected: []string{
				"examples/multiple-files/2025-01/01.json",
				"examples/multiple-files/2025-01/02.json",
				"examples/multiple-files/2025-01/03.json",
			},
		},
		{
			name:     "multiple patterns",
			patterns: []string{"examples/multiple-files/2025-01/01.json", "examples/multiple-files/2025-02/01.json"},
			expected: []string{
				"examples/multiple-files/2025-01/01.json",
				"examples/multiple-files/2025-02/01.json",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := jq.ExpandGlobPatterns(tc.patterns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tc.expected) {
				t.Errorf("expected %d files, got %d: %v", len(tc.expected), len(result), result)
			}

			// Check that all expected files are present (order may vary)
			for _, expected := range tc.expected {
				found := false
				for _, actual := range result {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected file %s not found in result: %v", expected, result)
				}
			}
		})
	}
}

func TestExecuteJQMultiFiles(t *testing.T) {
	// Create test data
	jsonData1 := map[string]interface{}{
		"transactions": []interface{}{
			map[string]interface{}{"id": "txn_001", "amount": float64(100)},
			map[string]interface{}{"id": "txn_002", "amount": float64(200)},
		},
	}

	jsonData2 := map[string]interface{}{
		"transactions": []interface{}{
			map[string]interface{}{"id": "txn_003", "amount": float64(150)},
		},
	}

	jsonDataList := []interface{}{jsonData1, jsonData2}

	// Test query that uses inputs
	result, err := jq.ExecuteJQMultiFiles("[inputs | .transactions[] | select(.amount > 120)]", jsonDataList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result to verify
	var transactions []interface{}
	if err := json.Unmarshal([]byte(result), &transactions); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(transactions))
	}
}

func TestValidateAndReadJSONFiles(t *testing.T) {
	testCases := []struct {
		name        string
		filePaths   []string
		expectError bool
	}{
		{
			name:        "valid files",
			filePaths:   []string{"examples/sample.json"},
			expectError: false,
		},
		{
			name:        "multiple valid files",
			filePaths:   []string{"examples/multiple-files/2025-01/01.json", "examples/multiple-files/2025-01/02.json"},
			expectError: false,
		},
		{
			name:        "nonexistent file",
			filePaths:   []string{"examples/nonexistent.json"},
			expectError: true,
		},
		{
			name:        "directory instead of file",
			filePaths:   []string{"examples"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := jq.ValidateAndReadJSONFiles(tc.filePaths)

			if tc.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(result) != len(tc.filePaths) {
					t.Errorf("expected %d JSON objects, got %d", len(tc.filePaths), len(result))
				}
			}
		})
	}
}

func TestParseGenerateConfigArgs(t *testing.T) {
	testCases := []struct {
		name            string
		args            []string
		wantDataPath    string
		wantOutputPath  string
		wantErr         error
		wantErrContains string
	}{
		{
			name:           "with explicit output",
			args:           []string{"-p", "/tmp/data", "-o", "cfg.yaml"},
			wantDataPath:   "/tmp/data",
			wantOutputPath: "cfg.yaml",
		},
		{
			name:           "defaults output path",
			args:           []string{"-p", "/tmp/data"},
			wantDataPath:   "/tmp/data",
			wantOutputPath: "config.yaml",
		},
		{
			name:            "missing data path",
			args:            []string{},
			wantErrContains: "-p <path> is required",
		},
		{
			name:    "help flag exits early",
			args:    []string{"-h"},
			wantErr: flag.ErrHelp,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dataPath, outputPath, err := parseGenerateConfigArgs(tc.args)

			switch {
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			case tc.wantErrContains != "":
				if err == nil || !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErrContains, err)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if dataPath != tc.wantDataPath {
					t.Fatalf("expected data path %q, got %q", tc.wantDataPath, dataPath)
				}
				if outputPath != tc.wantOutputPath {
					t.Fatalf("expected output path %q, got %q", tc.wantOutputPath, outputPath)
				}
			}
		})
	}
}

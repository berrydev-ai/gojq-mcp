package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/berrydev-ai/gojq-mcp/cli"
	"github.com/berrydev-ai/gojq-mcp/config"
	"github.com/berrydev-ai/gojq-mcp/registry"
	"github.com/berrydev-ai/gojq-mcp/server"
)

const version = "1.0.5"

func printUsage() {
	fmt.Fprintf(os.Stderr, `gojq-mcp v%s - A dual-mode JSON query tool

USAGE:
   gojq-mcp [OPTIONS]

MODES:
   CLI Mode:         gojq-mcp -f <file> -q <query>
   Server Mode:      gojq-mcp -p <path> [-c <config>] [-i <instructions>]
   Generate Config:  gojq-mcp generate-config -p <path> [-o <output>]

OPTIONS:
    -f <file>       Path to JSON file (CLI mode, can be used multiple times)
    -q <query>      jq query to execute (CLI mode)
   -p <path>       Path to folder containing JSON files
   -c <config>     Path to YAML configuration file (Server mode)
   -i <instructions> Server instructions for LLM (overrides config)
   -o <output>     Output file for generated config (default: config.yaml)
   -t <transport>  Transport type: stdio, http, or sse (overrides config, default: stdio)
   -a <address>    Address to listen on for http/sse (overrides config, default: :8080)
   -token <token>  Bearer token required by http/sse transports
   -watch          Enable file system watching (default: true)
   --version       Display version information
   --help          Display this help message

FEATURES:
   • Generate config files with prompts and instructions
   • YAML configuration for transport, port, instructions, and prompts
   • Real-time file monitoring with automatic client notifications
   • HTTP streaming transport for push notifications

EXAMPLES:
    # Generate a config file
    gojq-mcp generate-config -p ./data -o config.yaml

    # CLI mode - query a single JSON file
    gojq-mcp -f data.json -q '.users[] | select(.age > 30)'

    # CLI mode - query multiple specific files
    gojq-mcp -f file1.json -f file2.json -q '.transactions[] | .amount | add'

    # CLI mode - query files using glob patterns
    gojq-mcp -f './data/*.json' -q '.transactions[] | .amount | add'

    # Server mode with config file
    gojq-mcp -p ./data -c config.yaml

    # Server mode with CLI overrides
    gojq-mcp -p ./data -c config.yaml -t http -a :9000

DOCUMENTATION:
   https://github.com/berrydev-ai/gojq-mcp

`, version)
}

func main() {
	flag.Usage = printUsage

	filePaths := make([]string, 0)
	query := flag.String("q", "", "jq query to execute")
	dataPath := flag.String("p", "", "Path to folder containing JSON files")
	configPath := flag.String("c", "", "Path to YAML configuration file")
	instructions := flag.String("i", "", "Server instructions for LLM (overrides config)")
	transport := flag.String("t", "", "Transport type: stdio, http, or sse (overrides config)")
	address := flag.String("a", "", "Address to listen on (overrides config)")
	tokenFlag := flag.String("token", "", "Bearer token required by http/sse transports")
	enableWatch := flag.Bool("watch", true, "Enable file system watching")
	showVersion := flag.Bool("version", false, "Display version information")

	// Custom flag parsing to support multiple -f flags
	// We need to parse -f flags manually since Go's flag package doesn't support repeated flags
	// Extract -f flags and remove them from args before calling flag.Parse()
	args := os.Args[1:]
	filteredArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "-f" && i+1 < len(args) {
			filePaths = append(filePaths, args[i+1])
			i++ // Skip the next argument since we consumed it
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	// Temporarily replace os.Args for flag.Parse()
	oldArgs := os.Args
	os.Args = append([]string{os.Args[0]}, filteredArgs...)
	flag.Parse()
	os.Args = oldArgs

	if *showVersion {
		fmt.Printf("gojq-mcp version %s\n", version)
		return
	}

	// Determine auth token: CLI flag overrides config
	authToken := strings.TrimSpace(*tokenFlag)

	// CLI mode
	if len(filePaths) > 0 && *query != "" {
		cli.RunCLIMode(filePaths, *query)
		return
	}

	// Server mode - load config or use defaults
	var cfg *config.Config
	if *configPath != "" {
		var err error
		cfg, err = config.LoadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "✅ Loaded configuration from %s\n", *configPath)
	} else {
		// Default config
		cfg = &config.Config{
			Transport: "stdio",
			Port:      8080,
		}
	}

	// Override config with CLI flags if provided
	if *dataPath != "" {
		cfg.DataPath = *dataPath
	}
	if *transport != "" {
		cfg.Transport = *transport
	}
	if *address != "" {
		// Parse port from address
		parts := strings.Split(*address, ":")
		if len(parts) == 2 {
			fmt.Sscanf(parts[1], "%d", &cfg.Port)
		}
	}
	if *instructions != "" {
		cfg.Instructions = *instructions
	}
	// Use config auth token if no flag was provided
	if authToken == "" && cfg.AuthToken != "" {
		authToken = cfg.AuthToken
	}

	// Verify data path is set
	if cfg.DataPath == "" {
		fmt.Fprintf(os.Stderr, "Error: data path is required. Use -p flag or set data_path in config\n\n")
		printUsage()
		os.Exit(1)
	}

	// Verify data path exists
	if info, err := os.Stat(cfg.DataPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: data path does not exist: %s\n", cfg.DataPath)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: data path is not a directory: %s\n", cfg.DataPath)
		os.Exit(1)
	}

	// Create file registry
	fileRegistry, err := registry.NewFileRegistry(cfg.DataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing file registry: %v\n", err)
		os.Exit(1)
	}
	defer fileRegistry.Close()

	// Create MCP server
	s, err := server.SetupMCPServer(cfg, fileRegistry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up MCP server: %v\n", err)
		os.Exit(1)
	}

	// Start file watching if enabled
	if *enableWatch {
		if err := fileRegistry.StartWatching(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not enable file watching: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing without file watching...\n")
		} else {
			fmt.Fprintf(os.Stderr, "✅ Push notifications enabled - clients will be notified of file changes\n")
		}
	}

	// Start the server
	err = server.StartServer(s, cfg, authToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

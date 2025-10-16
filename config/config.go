package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration file structure
type Config struct {
	DataPath     string         `yaml:"data_path"`
	Transport    string         `yaml:"transport"`
	Port         int            `yaml:"port"`
	Instructions string         `yaml:"instructions"`
	Prompts      []PromptConfig `yaml:"prompts"`
}

// PromptConfig defines a reusable prompt
type PromptConfig struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Arguments   []PromptArgumentConfig   `yaml:"arguments"`
}

// PromptArgumentConfig defines a prompt argument
type PromptArgumentConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Set defaults if not specified
	if config.Transport == "" {
		config.Transport = "stdio"
	}
	if config.Port == 0 {
		config.Port = 8080
	}

	return &config, nil
}

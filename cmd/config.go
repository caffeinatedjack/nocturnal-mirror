package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFileName = "nocturnal.yaml"

// Config represents the project-level configuration.
type Config struct {
	Validation ValidationConfig `yaml:"validation"`
	Context    ContextConfig    `yaml:"context"`
	Git        GitConfig        `yaml:"git"`
}

// ValidationConfig controls proposal validation behavior.
type ValidationConfig struct {
	Strict          bool     `yaml:"strict"`           // Treat warnings as errors
	RequireSections []string `yaml:"require_sections"` // Additional required sections
}

// ContextConfig controls MCP context tool behavior.
type ContextConfig struct {
	IncludeAffectedFiles bool `yaml:"include_affected_files"` // Include code from affected files
	MaxFileLines         int  `yaml:"max_file_lines"`         // Max lines to include per file
}

// GitConfig controls git integration behavior.
type GitConfig struct {
	AutoSnapshot bool `yaml:"auto_snapshot"` // Automatically create git snapshots before tasks
	AutoCommit   bool `yaml:"auto_commit"`   // Automatically commit changes when tasks complete
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Validation: ValidationConfig{
			Strict:          false,
			RequireSections: []string{},
		},
		Context: ContextConfig{
			IncludeAffectedFiles: false,
			MaxFileLines:         50,
		},
		Git: GitConfig{
			AutoSnapshot: true,
			AutoCommit:   true,
		},
	}
}

// getConfigPath returns the path to the config file.
func getConfigPath(specPath string) string {
	return filepath.Join(specPath, configFileName)
}

// loadConfig reads the config file. Returns default config if file doesn't exist.
func loadConfig(specPath string) (*Config, error) {
	configPath := getConfigPath(specPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// saveConfig writes the config file.
func saveConfig(specPath string, config *Config) error {
	configPath := getConfigPath(specPath)
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// loadConfigOrDefault loads config, returning default on any error.
func loadConfigOrDefault(specPath string) *Config {
	config, err := loadConfig(specPath)
	if err != nil {
		return DefaultConfig()
	}
	return config
}

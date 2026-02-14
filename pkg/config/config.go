// Package config provides configuration management for CoreMCP.
// It supports YAML configuration files and environment variables.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the complete CoreMCP configuration.
type Config struct {
	Server  ServerConfig   `mapstructure:"server"`
	Logging LoggingConfig  `mapstructure:"logging"`
	Sources []SourceConfig `mapstructure:"sources"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Name      string `mapstructure:"name"`
	Version   string `mapstructure:"version"`
	Transport string `mapstructure:"transport"`
	Port      int    `mapstructure:"port"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type SourceConfig struct {
	Name     string `mapstructure:"name"`
	Type     string `mapstructure:"type"` // mssql, firebird, postgres
	DSN      string `mapstructure:"dsn"`  // Data Source Name (Connection String)
	ReadOnly bool   `mapstructure:"readonly"`
}

// LoadConfig loads the CoreMCP configuration from a YAML file.
// It supports environment variables with COREMCP_ prefix and provides sensible defaults.
// The path parameter can be a directory or a full file path.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// File Settings
	v.SetConfigName("coremcp") // File Name (without extension)
	v.SetConfigType("yaml")    // File Type
	v.AddConfigPath(".")       // Current directory
	v.AddConfigPath(path)      // Specified path (e.g., /etc/coremcp/)

	// Environment Variables
	v.SetEnvPrefix("COREMCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Fallback
	v.SetDefault("server.transport", "stdio")
	v.SetDefault("server.port", 8080)
	v.SetDefault("logging.level", "info")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(os.Stderr, "Warning: Config file not found, using defaults.\n")
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// If no sources defined, add a default dummy source
	if len(cfg.Sources) == 0 {
		cfg.Sources = []SourceConfig{
			{
				Name:     "dummy",
				Type:     "dummy",
				DSN:      "dummy://default",
				ReadOnly: true,
			},
		}
	}

	return &cfg, nil
}

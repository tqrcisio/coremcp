package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectError bool
	}{
		{
			name: "valid config",
			configYAML: `
server:
  name: "test-server"
  version: "0.1.0"
  transport: "stdio"

sources:
  - name: "test_db"
    type: "dummy"
    dsn: "dummy://test"
    readonly: true
`,
			expectError: false,
		},
		{
			name: "minimal config",
			configYAML: `
server:
  name: "minimal"
  version: "0.1.0"
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "coremcp.yaml")

			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Test LoadConfig
			cfg, err := LoadConfig(configPath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && cfg == nil {
				t.Error("Expected config but got nil")
			}
		})
	}
}

func TestSourceReadOnlyDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "coremcp.yaml")

	// Source without readonly field — should default to true
	configYAML := `
server:
  name: "test"
  version: "0.1.0"
sources:
  - name: "test_db"
    type: "dummy"
    dsn: "dummy://test"
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Sources) == 0 {
		t.Fatal("Expected at least one source")
	}
	if !cfg.Sources[0].IsReadOnly() {
		t.Error("Expected source without readonly field to default to true")
	}

	// Source with readonly: false — should be respected
	configYAML2 := `
server:
  name: "test"
  version: "0.1.0"
sources:
  - name: "test_db"
    type: "dummy"
    dsn: "dummy://test"
    readonly: false
`
	if err := os.WriteFile(configPath, []byte(configYAML2), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	cfg2, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg2.Sources[0].IsReadOnly() {
		t.Error("Expected source with readonly: false to not be read-only")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "coremcp.yaml")

	// Minimal config to test defaults
	configYAML := `
server:
  name: "test"
  version: "0.1.0"
`

	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test defaults
	if cfg.Server.Transport != "stdio" {
		t.Errorf("Expected default transport 'stdio', got '%s'", cfg.Server.Transport)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.Logging.Level)
	}
}

package server

import (
	"testing"

	"github.com/corebasehq/coremcp/pkg/adapter/dummy"
)

func TestNewMCPServer(t *testing.T) {
	name := "test-server"
	version := "0.1.0"

	server := NewMCPServer(name, version)

	if server == nil {
		t.Fatal("Expected server, got nil")
	}

	if server.mcpServer == nil {
		t.Error("Expected mcpServer to be initialized")
	}

	if server.sources == nil {
		t.Error("Expected sources map to be initialized")
	}
}

func TestMCPServer_AddSource(t *testing.T) {
	server := NewMCPServer("test", "0.1.0")

	source, err := dummy.New("dummy://test")
	if err != nil {
		t.Fatalf("Failed to create dummy source: %v", err)
	}

	server.AddSource("test_db", source, false)

	if len(server.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(server.sources))
	}

	retrievedEntry, exists := server.sources["test_db"]
	if !exists {
		t.Error("Source 'test_db' not found")
	}

	if retrievedEntry.source != source {
		t.Error("Retrieved source doesn't match added source")
	}
}

func TestMCPServer_AddMultipleSources(t *testing.T) {
	server := NewMCPServer("test", "0.1.0")

	source1, _ := dummy.New("dummy://test1")
	source2, _ := dummy.New("dummy://test2")

	server.AddSource("db1", source1, false)
	server.AddSource("db2", source2, true)

	if len(server.sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(server.sources))
	}
}

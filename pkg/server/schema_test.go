package server

import (
	"context"
	"strings"
	"testing"

	"github.com/corebasehq/coremcp/pkg/adapter/dummy"
)

func TestLoadSchemas(t *testing.T) {
	mcpSrv := NewMCPServer("test-server", "1.0.0")

	// Create and add a dummy source
	src, err := dummy.New("dummy://test")
	if err != nil {
		t.Fatalf("Failed to create dummy source: %v", err)
	}

	if err := src.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect source: %v", err)
	}

	mcpSrv.AddSource("test_db", src, true)

	// Load schemas
	ctx := context.Background()
	if err := mcpSrv.LoadSchemas(ctx); err != nil {
		t.Fatalf("LoadSchemas failed: %v", err)
	}

	// Verify schema context was built
	schemaCtx := mcpSrv.GetSchemaContext()
	if schemaCtx == "" {
		t.Fatal("Schema context is empty")
	}

	// Check if context contains expected information
	expectedStrings := []string{
		"DATABASE SCHEMA CONTEXT",
		"Source: test_db",
		"Table: users",
		"Table: orders",
		"Primary Key",
		"Foreign Keys",
		"user_id → users.id",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(schemaCtx, expected) {
			t.Errorf("Schema context missing expected string: %q", expected)
		}
	}

	// Check for column descriptions
	if !strings.Contains(schemaCtx, "User ID (Primary Key)") {
		t.Error("Schema context missing column description")
	}
}

func TestGetSchemaContext(t *testing.T) {
	mcpSrv := NewMCPServer("test-server", "1.0.0")

	// Before loading schemas, context should be empty
	if ctx := mcpSrv.GetSchemaContext(); ctx != "" {
		t.Error("Schema context should be empty before LoadSchemas")
	}

	// Add a source
	src, _ := dummy.New("dummy://test")
	src.Connect(context.Background())
	mcpSrv.AddSource("test_db", src, false)

	// Load schemas
	mcpSrv.LoadSchemas(context.Background())

	// Now context should be populated
	if ctx := mcpSrv.GetSchemaContext(); ctx == "" {
		t.Error("Schema context should not be empty after LoadSchemas")
	}
}

func TestSchemaPromptHandler(t *testing.T) {
	mcpSrv := NewMCPServer("test-server", "1.0.0")

	// Create and add a dummy source
	src, _ := dummy.New("dummy://test")
	src.Connect(context.Background())
	mcpSrv.AddSource("test_db", src, true)
	mcpSrv.LoadSchemas(context.Background())

	// Test the prompt handler
	// We're just testing that the schema context contains expected information
	schemaCtx := mcpSrv.GetSchemaContext()
	if !strings.Contains(schemaCtx, "users") {
		t.Error("Schema context should contain 'users' table")
	}
}

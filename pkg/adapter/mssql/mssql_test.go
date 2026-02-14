package mssql

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	dsn := "sqlserver://user:pass@localhost:1433?database=test"
	adapter, err := New(dsn)

	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if adapter == nil {
		t.Fatal("Expected adapter, got nil")
	}
}

func TestMSSQLAdapter_Name(t *testing.T) {
	adapter, _ := New("sqlserver://user:pass@localhost:1433?database=test")

	name := adapter.Name()
	expected := "MSSQL"

	if name != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, name)
	}
}

// TestMSSQLAdapter_Connect tests connection handling
// Note: This test doesn't require a real database connection
// as we're testing the adapter's structure and error handling
func TestMSSQLAdapter_Connect_InvalidDSN(t *testing.T) {
	// Test with invalid DSN format
	adapter, _ := New("invalid-dsn")
	ctx := context.Background()

	err := adapter.Connect(ctx)
	if err == nil {
		t.Error("Expected error with invalid DSN, got nil")
	}
}

func TestMSSQLAdapter_Close_WithoutConnect(t *testing.T) {
	adapter, _ := New("sqlserver://user:pass@localhost:1433?database=test")
	ctx := context.Background()

	// Calling Close without Connect should not panic
	err := adapter.Close(ctx)
	if err != nil {
		t.Errorf("Close() should handle nil db gracefully, got error: %v", err)
	}
}

/*
// Integration tests - these require a real MSSQL instance
// Run with: go test -tags=integration ./pkg/adapter/mssql

func TestMSSQLAdapter_Connect_Integration(t *testing.T) {
	// This requires MSSQL_TEST_DSN environment variable
	dsn := os.Getenv("MSSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MSSQL_TEST_DSN not set, skipping integration test")
	}

	adapter, err := New(dsn)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	ctx := context.Background()

	err = adapter.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Close(ctx)

	// Test query execution
	result, err := adapter.ExecuteQuery(ctx, "SELECT 1 as test")
	if err != nil {
		t.Fatalf("ExecuteQuery() failed: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(result.Rows))
	}
}

func TestMSSQLAdapter_GetSchema_Integration(t *testing.T) {
	dsn := os.Getenv("MSSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MSSQL_TEST_DSN not set, skipping integration test")
	}

	adapter, _ := New(dsn)
	ctx := context.Background()

	err := adapter.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Close(ctx)

	schema, err := adapter.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema() failed: %v", err)
	}

	if schema == nil {
		t.Error("Expected schema, got nil")
	}
}
*/

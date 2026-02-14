package dummy

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	dsn := "dummy://test"
	adapter, err := New(dsn)

	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if adapter == nil {
		t.Fatal("Expected adapter, got nil")
	}
}

func TestDummyAdapter_Name(t *testing.T) {
	adapter, _ := New("dummy://test")

	name := adapter.Name()
	expected := "DummyDB"

	if name != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, name)
	}
}

func TestDummyAdapter_Connect(t *testing.T) {
	adapter, _ := New("dummy://test")
	ctx := context.Background()

	err := adapter.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() failed: %v", err)
	}
}

func TestDummyAdapter_Close(t *testing.T) {
	adapter, _ := New("dummy://test")
	ctx := context.Background()

	err := adapter.Close(ctx)
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestDummyAdapter_GetSchema(t *testing.T) {
	adapter, _ := New("dummy://test")
	ctx := context.Background()

	schema, err := adapter.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema() failed: %v", err)
	}

	if len(schema) == 0 {
		t.Error("Expected non-empty schema")
	}

	// Check for expected tables
	foundUsers := false
	foundOrders := false

	for _, table := range schema {
		if table.Name == "users" {
			foundUsers = true
		}
		if table.Name == "orders" {
			foundOrders = true
		}
	}

	if !foundUsers {
		t.Error("Expected 'users' table in schema")
	}

	if !foundOrders {
		t.Error("Expected 'orders' table in schema")
	}
}

func TestDummyAdapter_ExecuteQuery(t *testing.T) {
	adapter, _ := New("dummy://test")
	ctx := context.Background()

	query := "SELECT * FROM users"
	result, err := adapter.ExecuteQuery(ctx, query)

	if err != nil {
		t.Fatalf("ExecuteQuery() failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Columns) == 0 {
		t.Error("Expected non-empty columns")
	}

	if len(result.Rows) == 0 {
		t.Error("Expected non-empty rows")
	}
}

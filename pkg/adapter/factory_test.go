package adapter

import (
	"testing"
)

func TestNewSource_Dummy(t *testing.T) {
	src, err := NewSource("dummy", "dummy://test")

	if err != nil {
		t.Fatalf("NewSource() failed: %v", err)
	}

	if src == nil {
		t.Fatal("Expected source, got nil")
	}

	if src.Name() != "DummyDB" {
		t.Errorf("Expected 'DummyDB', got '%s'", src.Name())
	}
}

func TestNewSource_MSSQL(t *testing.T) {
	dsn := "sqlserver://user:pass@localhost:1433?database=test"
	src, err := NewSource("mssql", dsn)

	if err != nil {
		t.Fatalf("NewSource() failed: %v", err)
	}

	if src == nil {
		t.Fatal("Expected source, got nil")
	}

	if src.Name() != "MSSQL" {
		t.Errorf("Expected 'MSSQL', got '%s'", src.Name())
	}
}

func TestNewSource_Unsupported(t *testing.T) {
	_, err := NewSource("unsupported_db", "some://dsn")

	if err == nil {
		t.Error("Expected error for unsupported database type")
	}
}

func TestNewSource_Firebird(t *testing.T) {
	_, err := NewSource("firebird", "firebird://test")

	if err == nil {
		t.Error("Expected error for unimplemented Firebird adapter")
	}
}

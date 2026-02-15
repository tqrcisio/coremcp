// Package core defines the core interfaces and types for CoreMCP database adapters.
package core

import (
	"context"
)

// Source represents a database connection and query interface.
// All database adapters must implement this interface.
type Source interface {
	// Name returns the human-readable name of the database adapter.
	Name() string
	// Connect establishes a connection to the database.
	Connect(ctx context.Context) error
	// Close closes the database connection.
	Close(ctx context.Context) error
	// GetSchema retrieves the database schema information.
	GetSchema(ctx context.Context) ([]TableSchema, error)
	// ExecuteQuery runs a SQL query and returns the results.
	ExecuteQuery(ctx context.Context, query string, args ...any) (*QueryResult, error)
}

// TableSchema represents the structure of a database table.
type TableSchema struct {
	Name        string
	Columns     []ColumnInfo
	ForeignKeys []ForeignKey
	PrimaryKeys []string
}

// ColumnInfo represents detailed information about a table column.
type ColumnInfo struct {
	Name        string
	DataType    string
	IsNullable  bool
	Description string // Comment/Description from database
}

// ForeignKey represents a foreign key relationship.
type ForeignKey struct {
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	ConstraintName   string
}

// QueryResult contains the results of a database query execution.
type QueryResult struct {
	Columns []string
	Rows    []map[string]any // (column_name: value)
}

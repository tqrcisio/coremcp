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
	// GetSchema retrieves the database table schema information.
	GetSchema(ctx context.Context) ([]TableSchema, error)
	// GetViews retrieves all views in the database.
	GetViews(ctx context.Context) ([]ViewSchema, error)
	// GetProcedures retrieves all stored procedures in the database.
	GetProcedures(ctx context.Context) ([]StoredProcedure, error)
	// ExecuteQuery runs a SQL query and returns the results.
	ExecuteQuery(ctx context.Context, query string, args ...any) (*QueryResult, error)
	// ExecuteProcedure executes a stored procedure by name with the given named parameters.
	// The procedure name is validated before execution.
	// Implementations may reject execution when the source is read-only.
	ExecuteProcedure(ctx context.Context, name string, params map[string]string) (*QueryResult, error)
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

// ViewSchema represents a database view.
type ViewSchema struct {
	Name    string
	Columns []ColumnInfo
}

// StoredProcedure represents a stored procedure in the database.
type StoredProcedure struct {
	Name        string
	Parameters  []ProcParameter
	Description string
}

// ProcParameter represents a parameter of a stored procedure.
type ProcParameter struct {
	Name     string // e.g. "@StartDate"
	DataType string // e.g. "datetime"
	Mode     string // "IN", "OUT", "INOUT"
}

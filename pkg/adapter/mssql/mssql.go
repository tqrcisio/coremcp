// Package mssql provides a Microsoft SQL Server database adapter.
package mssql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/corebasehq/coremcp/pkg/core"
	_ "github.com/microsoft/go-mssqldb"
)

// MSSQLAdapter implements the core.Source interface for Microsoft SQL Server.
type MSSQLAdapter struct {
	dsn string
	db  *sql.DB
}

// New creates a new MSSQL adapter with the given DSN.
// DSN format: sqlserver://username:password@host:port?database=dbname
func New(dsn string) (core.Source, error) {
	return &MSSQLAdapter{dsn: dsn}, nil
}

func (m *MSSQLAdapter) Name() string {
	return "MSSQL"
}

func (m *MSSQLAdapter) Connect(ctx context.Context) error {
	// DSN format: sqlserver://username:password@host:port?database=dbname
	db, err := sql.Open("sqlserver", m.dsn)
	if err != nil {
		return err
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("MSSQL ping error: %w", err)
	}

	m.db = db
	return nil
}

func (m *MSSQLAdapter) Close(ctx context.Context) error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *MSSQLAdapter) GetSchema(ctx context.Context) ([]core.TableSchema, error) {
	query := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_TYPE = 'BASE TABLE'
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []core.TableSchema
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		tables = append(tables, core.TableSchema{
			Name:    tableName,
			Columns: []string{"(Run query for the details...)"},
		})
	}
	return tables, nil
}

func (m *MSSQLAdapter) ExecuteQuery(ctx context.Context, query string, args ...any) (*core.QueryResult, error) {
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	var finalRows []map[string]any

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Make result JSON friendly
		rowMap := make(map[string]any)
		for i, col := range columns {
			val := values[i]

			// If data is []byte, convert to string for better JSON compatibility
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		finalRows = append(finalRows, rowMap)
	}

	return &core.QueryResult{
		Columns: columns,
		Rows:    finalRows,
	}, nil
}

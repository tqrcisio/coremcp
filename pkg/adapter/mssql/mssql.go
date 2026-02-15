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
		ORDER BY TABLE_NAME
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}

	// Now get detailed info for each table
	var tables []core.TableSchema
	for _, tableName := range tableNames {
		schema, err := m.getTableDetail(ctx, tableName)
		if err != nil {
			// Log and continue with partial info
			schema = core.TableSchema{
				Name: tableName,
				Columns: []core.ColumnInfo{
					{Name: "(error loading columns)", DataType: "unknown"},
				},
			}
		}
		tables = append(tables, schema)
	}

	return tables, nil
}

func (m *MSSQLAdapter) getTableDetail(ctx context.Context, tableName string) (core.TableSchema, error) {
	schema := core.TableSchema{
		Name:        tableName,
		Columns:     []core.ColumnInfo{},
		ForeignKeys: []core.ForeignKey{},
		PrimaryKeys: []string{},
	}

	// Get columns with descriptions
	columnQuery := `
		SELECT 
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.IS_NULLABLE,
			ISNULL(ep.value, '') as DESCRIPTION
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN sys.extended_properties ep 
			ON ep.major_id = OBJECT_ID(c.TABLE_SCHEMA + '.' + c.TABLE_NAME)
			AND ep.minor_id = c.ORDINAL_POSITION
			AND ep.name = 'MS_Description'
		WHERE c.TABLE_NAME = @p1
		ORDER BY c.ORDINAL_POSITION
	`

	rows, err := m.db.QueryContext(ctx, columnQuery, tableName)
	if err != nil {
		return schema, err
	}
	defer rows.Close()

	for rows.Next() {
		var col core.ColumnInfo
		var isNullable string
		var description sql.NullString

		if err := rows.Scan(&col.Name, &col.DataType, &isNullable, &description); err != nil {
			continue
		}

		col.IsNullable = (isNullable == "YES")
		if description.Valid {
			col.Description = description.String
		}

		schema.Columns = append(schema.Columns, col)
	}

	// Get primary keys
	pkQuery := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE OBJECTPROPERTY(OBJECT_ID(CONSTRAINT_SCHEMA + '.' + CONSTRAINT_NAME), 'IsPrimaryKey') = 1
		AND TABLE_NAME = @p1
		ORDER BY ORDINAL_POSITION
	`

	pkRows, err := m.db.QueryContext(ctx, pkQuery, tableName)
	if err == nil {
		defer pkRows.Close()
		for pkRows.Next() {
			var pkCol string
			if err := pkRows.Scan(&pkCol); err == nil {
				schema.PrimaryKeys = append(schema.PrimaryKeys, pkCol)
			}
		}
	}

	// Get foreign keys
	fkQuery := `
		SELECT 
			fk.name AS FK_NAME,
			c1.name AS COLUMN_NAME,
			OBJECT_NAME(fk.referenced_object_id) AS REFERENCED_TABLE,
			c2.name AS REFERENCED_COLUMN
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.columns c1 ON fkc.parent_object_id = c1.object_id AND fkc.parent_column_id = c1.column_id
		INNER JOIN sys.columns c2 ON fkc.referenced_object_id = c2.object_id AND fkc.referenced_column_id = c2.column_id
		WHERE OBJECT_NAME(fk.parent_object_id) = @p1
	`

	fkRows, err := m.db.QueryContext(ctx, fkQuery, tableName)
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			var fk core.ForeignKey
			if err := fkRows.Scan(&fk.ConstraintName, &fk.ColumnName, &fk.ReferencedTable, &fk.ReferencedColumn); err == nil {
				schema.ForeignKeys = append(schema.ForeignKeys, fk)
			}
		}
	}

	return schema, nil
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

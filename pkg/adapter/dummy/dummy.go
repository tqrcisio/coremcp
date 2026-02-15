// Package dummy provides a dummy database adapter for testing purposes.
// It returns mock data and does not connect to any real database.
package dummy

import (
	"context"
	"fmt"
	"os"

	"github.com/corebasehq/coremcp/pkg/core"
)

// DummyAdapter is a mock database adapter that returns fake data.
// Useful for testing and development without a real database connection.
type DummyAdapter struct {
	dsn string
}

// New creates a new dummy database adapter.
func New(dsn string) (core.Source, error) {
	return &DummyAdapter{dsn: dsn}, nil
}

func (d *DummyAdapter) Name() string {
	return "DummyDB"
}

func (d *DummyAdapter) Connect(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "[DummyDB] Connecting... DSN: %s\n", d.dsn)
	return nil
}

func (d *DummyAdapter) Close(ctx context.Context) error {
	fmt.Fprintln(os.Stderr, "[DummyDB] Connection closed.")
	return nil
}

func (d *DummyAdapter) GetSchema(ctx context.Context) ([]core.TableSchema, error) {
	return []core.TableSchema{
		{
			Name: "users",
			Columns: []core.ColumnInfo{
				{Name: "id", DataType: "int", IsNullable: false, Description: "User ID (Primary Key)"},
				{Name: "username", DataType: "varchar(50)", IsNullable: false, Description: "User's unique username"},
				{Name: "email", DataType: "varchar(100)", IsNullable: false, Description: "User's email address"},
				{Name: "created_at", DataType: "datetime", IsNullable: false, Description: "Account creation timestamp"},
			},
			PrimaryKeys: []string{"id"},
			ForeignKeys: []core.ForeignKey{},
		},
		{
			Name: "orders",
			Columns: []core.ColumnInfo{
				{Name: "id", DataType: "int", IsNullable: false, Description: "Order ID (Primary Key)"},
				{Name: "user_id", DataType: "int", IsNullable: false, Description: "Reference to users table"},
				{Name: "total", DataType: "decimal(10,2)", IsNullable: false, Description: "Total order amount"},
				{Name: "status", DataType: "varchar(20)", IsNullable: false, Description: "Order status (pending, completed, cancelled)"},
				{Name: "created_at", DataType: "datetime", IsNullable: false, Description: "Order creation timestamp"},
			},
			PrimaryKeys: []string{"id"},
			ForeignKeys: []core.ForeignKey{
				{
					ColumnName:       "user_id",
					ReferencedTable:  "users",
					ReferencedColumn: "id",
					ConstraintName:   "FK_orders_users",
				},
			},
		},
	}, nil
}

func (d *DummyAdapter) ExecuteQuery(ctx context.Context, query string, args ...any) (*core.QueryResult, error) {
	fmt.Fprintf(os.Stderr, "[DummyDB] Running query: %s\n", query)
	return &core.QueryResult{
		Columns: []string{"result"},
		Rows: []map[string]any{
			{"result": "This is a dummy result"},
		},
	}, nil
}

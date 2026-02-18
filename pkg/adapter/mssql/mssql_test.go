package mssql

import (
	"context"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	dsn := "sqlserver://user:pass@localhost:1433?database=test"
	adapter, err := New(dsn, false, false)

	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if adapter == nil {
		t.Fatal("Expected adapter, got nil")
	}
}

func TestMSSQLAdapter_Name(t *testing.T) {
	adapter, _ := New("sqlserver://user:pass@localhost:1433?database=test", false, false)

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
	adapter, _ := New("invalid-dsn", false, false)
	ctx := context.Background()

	err := adapter.Connect(ctx)
	if err == nil {
		t.Error("Expected error with invalid DSN, got nil")
	}
}

func TestMSSQLAdapter_Name_AfterConnect(t *testing.T) {
	// After a successful Connect(), serverMajor is set and Name() includes the version.
	// We can't run a real connection here, but we can directly set serverMajor to
	// verify the label logic.
	a := &MSSQLAdapter{serverMajor: 8}
	if a.Name() != "MSSQL (SQL Server 2000)" {
		t.Errorf("unexpected name: %s", a.Name())
	}
	a.serverMajor = 10
	if a.Name() != "MSSQL (SQL Server 2008/2008 R2)" {
		t.Errorf("unexpected name: %s", a.Name())
	}
	a.serverMajor = 15
	if a.Name() != "MSSQL (SQL Server 2019)" {
		t.Errorf("unexpected name: %s", a.Name())
	}
	// serverMajor == 0 means not yet connected → plain name
	a.serverMajor = 0
	if a.Name() != "MSSQL" {
		t.Errorf("unexpected name before connect: %s", a.Name())
	}
}

func TestIsPreSQL2005(t *testing.T) {
	cases := []struct {
		major    int
		expected bool
	}{
		{0, false},  // not connected — unknown, default false
		{7, true},   // SQL Server 7.0
		{8, true},   // SQL Server 2000
		{9, false},  // SQL Server 2005
		{10, false}, // SQL Server 2008
		{15, false}, // SQL Server 2019
	}
	for _, tc := range cases {
		a := &MSSQLAdapter{serverMajor: tc.major}
		if got := a.isPreSQL2005(); got != tc.expected {
			t.Errorf("isPreSQL2005() with major=%d: got %v, want %v", tc.major, got, tc.expected)
		}
	}
}

func TestIsPreSQL2012(t *testing.T) {
	cases := []struct {
		major    int
		expected bool
	}{
		{0, false},  // not connected
		{8, true},   // SQL Server 2000
		{10, true},  // SQL Server 2008 R2
		{11, false}, // SQL Server 2012 — first with OFFSET FETCH
		{15, false}, // SQL Server 2019
	}
	for _, tc := range cases {
		a := &MSSQLAdapter{serverMajor: tc.major}
		if got := a.isPreSQL2012(); got != tc.expected {
			t.Errorf("isPreSQL2012() with major=%d: got %v, want %v", tc.major, got, tc.expected)
		}
	}
}

func TestConvertLimitToTop(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		// Standard LIMIT injection from security.QueryModifier
		{
			"SELECT * FROM users LIMIT 1000",
			"SELECT TOP 1000 * FROM users",
		},
		{
			"select id, name from orders limit 500",
			"select TOP 500 id, name from orders",
		},
		// Already has TOP — must not double-inject
		{
			"SELECT TOP 100 * FROM users",
			"SELECT TOP 100 * FROM users",
		},
		{
			"SELECT TOP(50) id FROM orders LIMIT 1000",
			"SELECT TOP(50) id FROM orders LIMIT 1000",
		},
		// No LIMIT clause — unchanged
		{
			"SELECT * FROM users",
			"SELECT * FROM users",
		},
		// Non-SELECT — unchanged
		{
			"INSERT INTO t VALUES (1) LIMIT 1",
			"INSERT INTO t VALUES (1) LIMIT 1",
		},
	}
	for _, tc := range cases {
		got := convertLimitToTop(tc.input)
		if got != tc.expected {
			t.Errorf("convertLimitToTop(%q)\n  got:  %q\n  want: %q", tc.input, got, tc.expected)
		}
	}
}

func TestAdaptQueryForVersion_LimitToTop(t *testing.T) {
	// On any version, LIMIT N should be rewritten to TOP N
	a := &MSSQLAdapter{serverMajor: 15}
	q := a.adaptQueryForVersion("SELECT * FROM sales LIMIT 1000")
	expected := "SELECT TOP 1000 * FROM sales"
	if q != expected {
		t.Errorf("adaptQueryForVersion LIMIT→TOP:\n  got:  %q\n  want: %q", q, expected)
	}
}

func TestAdaptQueryForVersion_OffsetFetchPreSQL2012(t *testing.T) {
	// Pre-2012 server: OFFSET FETCH should be stripped and replaced with TOP N
	a := &MSSQLAdapter{serverMajor: 10} // SQL Server 2008
	input := "SELECT id, name FROM customers ORDER BY id OFFSET 0 ROWS FETCH NEXT 50 ROWS ONLY"
	got := a.adaptQueryForVersion(input)
	// Should contain TOP 50 and not contain OFFSET or FETCH
	if !containsIgnoreCase(got, "TOP 50") {
		t.Errorf("expected TOP 50 in result, got: %q", got)
	}
	if containsIgnoreCase(got, "OFFSET") || containsIgnoreCase(got, "FETCH") {
		t.Errorf("OFFSET/FETCH should be stripped for pre-2012, got: %q", got)
	}
}

func TestAdaptQueryForVersion_OffsetFetchSQL2012Plus(t *testing.T) {
	// SQL 2012+: OFFSET FETCH should be left untouched
	a := &MSSQLAdapter{serverMajor: 11} // SQL Server 2012
	input := "SELECT id FROM orders ORDER BY id OFFSET 0 ROWS FETCH NEXT 100 ROWS ONLY"
	got := a.adaptQueryForVersion(input)
	if got != input {
		t.Errorf("adaptQueryForVersion should not modify OFFSET FETCH on SQL 2012+\n  got:  %q\n  want: %q", got, input)
	}
}

// containsIgnoreCase is a test helper for case-insensitive substring checks.
func containsIgnoreCase(s, sub string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(sub))
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

package server

import (
	"context"
	"strings"
	"testing"

	"github.com/corebasehq/coremcp/pkg/adapter/dummy"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleListTables(t *testing.T) {
	mcpSrv := NewMCPServer("test-server", "1.0.0")

	src, _ := dummy.New("dummy://test")
	if err := src.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect source: %v", err)
	}
	mcpSrv.AddSource("test_db", src, true)

	// Create request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"source_name": "test_db",
	}

	ctx := context.Background()
	result, err := mcpSrv.handleListTables(ctx, request)

	if err != nil {
		t.Fatalf("handleListTables failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check if result contains expected tables
	resultText := getResultText(result)
	if !strings.Contains(resultText, "users") || !strings.Contains(resultText, "orders") {
		t.Error("Result should contain 'users' and 'orders' tables")
	}
}

func TestHandleDescribeTable(t *testing.T) {
	mcpSrv := NewMCPServer("test-server", "1.0.0")

	src, _ := dummy.New("dummy://test")
	if err := src.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect source: %v", err)
	}
	mcpSrv.AddSource("test_db", src, true)

	// Create request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"source_name": "test_db",
		"table_name":  "users",
	}

	ctx := context.Background()
	result, err := mcpSrv.handleDescribeTable(ctx, request)

	if err != nil {
		t.Fatalf("handleDescribeTable failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	resultText := getResultText(result)

	// Check for expected content
	expectedStrings := []string{
		"Table: users",
		"Primary Key",
		"Columns",
		"id",
		"username",
		"email",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(resultText, expected) {
			t.Errorf("Result missing expected string: %q", expected)
		}
	}
}

func TestAddCustomTool(t *testing.T) {
	mcpSrv := NewMCPServer("test-server", "1.0.0")

	src, _ := dummy.New("dummy://test")
	if err := src.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect source: %v", err)
	}
	mcpSrv.AddSource("test_db", src, true)

	// Add a custom tool
	err := mcpSrv.AddCustomTool(
		"get_user_orders",
		"Get orders for a specific user",
		"test_db",
		"SELECT * FROM orders WHERE user_id = {{user_id}}",
		[]string{"user_id"},
	)

	if err != nil {
		t.Fatalf("AddCustomTool failed: %v", err)
	}

	// Check if tool was registered
	if _, exists := mcpSrv.customTools["get_user_orders"]; !exists {
		t.Error("Custom tool was not registered")
	}

	// Verify tool config
	toolEntry := mcpSrv.customTools["get_user_orders"]
	if toolEntry.sourceName != "test_db" {
		t.Error("Tool source name mismatch")
	}
	if !strings.Contains(toolEntry.query, "user_id") {
		t.Error("Tool query should contain user_id parameter")
	}
}

func TestDescribeTableNotFound(t *testing.T) {
	mcpSrv := NewMCPServer("test-server", "1.0.0")

	src, _ := dummy.New("dummy://test")
	if err := src.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect source: %v", err)
	}
	mcpSrv.AddSource("test_db", src, true)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"source_name": "test_db",
		"table_name":  "nonexistent_table",
	}

	ctx := context.Background()
	result, err := mcpSrv.handleDescribeTable(ctx, request)

	if err != nil {
		t.Fatalf("handleDescribeTable failed: %v", err)
	}

	resultText := getResultText(result)
	if !strings.Contains(resultText, "not found") && !strings.Contains(resultText, "Table not found") {
		t.Error("Should return error for nonexistent table")
	}
}

// Helper function to extract text from CallToolResult
func getResultText(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}

	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text
		}
	}

	return ""
}

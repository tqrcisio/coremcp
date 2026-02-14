// Package server implements the MCP (Model Context Protocol) server for database operations.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/corebasehq/coremcp/pkg/core"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/mark3labs/mcp-go/server"
)

// MCPServer represents the MCP server instance that handles database queries.
type MCPServer struct {
	mcpServer *server.MCPServer
	sources   map[string]sourceEntry
}

type sourceEntry struct {
	source   core.Source
	readOnly bool
}

// NewMCPServer creates a new MCP server instance with the specified name and version.
// It automatically registers the query_database tool for executing SQL queries.
func NewMCPServer(name, version string) *MCPServer {
	s := server.NewMCPServer(name, version)

	ms := &MCPServer{
		mcpServer: s,
		sources:   make(map[string]sourceEntry),
	}

	tool := mcp.NewTool("query_database",
		mcp.WithDescription("Executes a SQL query on a specified database source. Use this to retrieve data."),
		mcp.WithString("source_name", mcp.Required(), mcp.Description("The name of the database source defined in config (e.g., 'mydb')")),
		mcp.WithString("query", mcp.Required(), mcp.Description("The SQL query to execute (e.g., 'SELECT * FROM users')")),
	)

	s.AddTool(tool, ms.handleQueryDatabase)

	return ms
}

// AddSource registers a database source with the given name.
// This makes the source available for query execution via the MCP protocol.
func (ms *MCPServer) AddSource(name string, src core.Source, readOnly bool) {
	ms.sources[name] = sourceEntry{
		source:   src,
		readOnly: readOnly,
	}
}

// StartStdio starts the MCP server using stdio transport.
// This is the standard transport for Claude Desktop integration.
func (ms *MCPServer) StartStdio() error {
	return server.ServeStdio(ms.mcpServer)
}

func (ms *MCPServer) handleQueryDatabase(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourceName, err := request.RequireString("source_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	entry, exists := ms.sources[sourceName]
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("Source not found: %s", sourceName)), nil
	}

	if entry.readOnly && !isQuerySafe(query) {
		return mcp.NewToolResultError("Query rejected: Source is in read-only mode and query contains potentially unsafe keywords"), nil
	}

	result, err := entry.source.ExecuteQuery(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Query execution failed: %v", err)), nil
	}

	jsonResult, err := json.MarshalIndent(result.Rows, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal result"), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

// isQuerySafe checks if the query is safe for read-only mode.
// This is a basic heuristic check and not a full SQL parser substitution.
// Users should properly configure their database user permissions for true security.
func isQuerySafe(query string) bool {
	q := strings.TrimSpace(strings.ToUpper(query))

	// Allow only SELECT and WITH statements
	if !strings.HasPrefix(q, "SELECT") && !strings.HasPrefix(q, "WITH") {
		return false
	}

	// Reject if it contains generic destructive keywords as whole words
	// We check for " KEYWORD " or starting/ending
	forbidden := []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE", "EXEC", "CREATE", "GRANT", "REVOKE", "MERGE"}

	for _, word := range forbidden {
		// Check strictly for whole words to avoid blocking valid columns like 'inserted_at'
		// This is still primitive but covers the 99% usage by an LLM tool.
		if strings.Contains(q, " "+word+" ") ||
			strings.HasPrefix(q, word+" ") ||
			strings.HasSuffix(q, " "+word) ||
			strings.Contains(q, "\n"+word+" ") ||
			strings.Contains(q, "\t"+word+" ") {
			return false
		}
	}

	return true
}

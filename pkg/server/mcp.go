// Package server implements the MCP (Model Context Protocol) server for database operations.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/corebasehq/coremcp/pkg/core"
	"github.com/corebasehq/coremcp/pkg/security"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/mark3labs/mcp-go/server"
)

// MCPServer represents the MCP server instance that handles database queries.
type MCPServer struct {
	mcpServer      *server.MCPServer
	sources        map[string]sourceEntry
	schemaContext  string                     // Pre-built schema context for AI
	customTools    map[string]customToolEntry // Custom tools from config
	queryValidator *security.QueryValidator   // SQL query validator
	piiMasker      *security.PIIMasker        // PII data masker
	queryModifier  *security.QueryModifier    // Query modifier for row limits
}

type customToolEntry struct {
	sourceName string
	query      string
	parameters []string
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
		mcpServer:     s,
		sources:       make(map[string]sourceEntry),
		schemaContext: "",
		customTools:   make(map[string]customToolEntry),
		// Security components will be initialized via ConfigureSecurity()
		queryValidator: security.NewQueryValidator(nil, nil),
		queryModifier:  security.NewQueryModifier(1000),
	}

	tool := mcp.NewTool("query_database",
		mcp.WithDescription("Executes a SQL query on a specified database source. Use this to retrieve data."),
		mcp.WithString("source_name", mcp.Required(), mcp.Description("The name of the database source defined in config (e.g., 'mydb')")),
		mcp.WithString("query", mcp.Required(), mcp.Description("The SQL query to execute (e.g., 'SELECT * FROM users')")),
	)

	s.AddTool(tool, ms.handleQueryDatabase)

	// Add list_tables tool
	listTablesTool := mcp.NewTool("list_tables",
		mcp.WithDescription("Lists all available tables in a database source with their column counts"),
		mcp.WithString("source_name", mcp.Required(), mcp.Description("The name of the database source")),
	)
	s.AddTool(listTablesTool, ms.handleListTables)

	// Add describe_table tool
	describeTableTool := mcp.NewTool("describe_table",
		mcp.WithDescription("Shows detailed schema information for a specific table including columns, types, keys, and relationships"),
		mcp.WithString("source_name", mcp.Required(), mcp.Description("The name of the database source")),
		mcp.WithString("table_name", mcp.Required(), mcp.Description("The name of the table to describe")),
	)
	s.AddTool(describeTableTool, ms.handleDescribeTable)

	// Add prompt for database schema context
	prompt := mcp.NewPrompt("database_schema",
		mcp.WithPromptDescription("Shows the complete database schema with tables, columns, types, and relationships"),
	)

	s.AddPrompt(prompt, ms.handleSchemaPrompt)

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

// ConfigureSecurity initializes security components with config values.
func (ms *MCPServer) ConfigureSecurity(maxRowLimit int, enablePIIMasking bool, piiPatterns []security.MaskPattern, allowedKeywords, blockedKeywords []string) error {
	// Configure query validator
	ms.queryValidator = security.NewQueryValidator(allowedKeywords, blockedKeywords)

	// Configure query modifier for row limiting
	ms.queryModifier = security.NewQueryModifier(maxRowLimit)

	// Configure PII masker
	patterns := piiPatterns
	if len(patterns) == 0 && enablePIIMasking {
		// Use default patterns if none provided
		patterns = security.DefaultPIIPatterns()
	}

	masker, err := security.NewPIIMasker(patterns, enablePIIMasking)
	if err != nil {
		return fmt.Errorf("failed to initialize PII masker: %w", err)
	}
	ms.piiMasker = masker

	return nil
}

// LoadSchemas loads database schemas from all sources and builds the context prompt.
// This should be called after all sources are added and connected.
func (ms *MCPServer) LoadSchemas(ctx context.Context) error {
	var contextBuilder strings.Builder
	contextBuilder.WriteString("=== DATABASE SCHEMA CONTEXT ===\n\n")
	contextBuilder.WriteString("You have access to the following database sources and their schemas:\n\n")

	for name, entry := range ms.sources {
		contextBuilder.WriteString(fmt.Sprintf("## Source: %s (%s)\n", name, entry.source.Name()))
		if entry.readOnly {
			contextBuilder.WriteString("(Read-Only Mode)\n")
		}
		contextBuilder.WriteString("\n")

		schemas, err := entry.source.GetSchema(ctx)
		if err != nil {
			contextBuilder.WriteString(fmt.Sprintf("⚠️  Error loading schema: %v\n\n", err))
			continue
		}

		if len(schemas) == 0 {
			contextBuilder.WriteString("No tables found.\n\n")
			continue
		}

		for _, table := range schemas {
			contextBuilder.WriteString(fmt.Sprintf("### Table: %s\n", table.Name))

			// Primary Keys
			if len(table.PrimaryKeys) > 0 {
				contextBuilder.WriteString(fmt.Sprintf("Primary Key(s): %s\n", strings.Join(table.PrimaryKeys, ", ")))
			}

			// Columns
			contextBuilder.WriteString("Columns:\n")
			for _, col := range table.Columns {
				nullable := ""
				if col.IsNullable {
					nullable = " (nullable)"
				}
				contextBuilder.WriteString(fmt.Sprintf("  - %s: %s%s", col.Name, col.DataType, nullable))
				if col.Description != "" {
					contextBuilder.WriteString(fmt.Sprintf(" -- %s", col.Description))
				}
				contextBuilder.WriteString("\n")
			}

			// Foreign Keys
			if len(table.ForeignKeys) > 0 {
				contextBuilder.WriteString("Foreign Keys:\n")
				for _, fk := range table.ForeignKeys {
					contextBuilder.WriteString(fmt.Sprintf("  - %s → %s.%s (%s)\n",
						fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn, fk.ConstraintName))
				}
			}

			contextBuilder.WriteString("\n")
		}
	}

	contextBuilder.WriteString("\n=== END OF SCHEMA CONTEXT ===\n")
	contextBuilder.WriteString("\nUse the 'query_database' tool to execute SQL queries on these sources.\n")

	ms.schemaContext = contextBuilder.String()
	return nil
}

// GetSchemaContext returns the pre-built schema context for use in AI prompts.
func (ms *MCPServer) GetSchemaContext() string {
	return ms.schemaContext
}

// AddCustomTool registers a custom tool with a predefined query.
func (ms *MCPServer) AddCustomTool(name, description, sourceName, query string, parameters []string) error {
	// Build MCP tool with dynamic parameters
	toolOptions := []mcp.ToolOption{
		mcp.WithDescription(description),
		mcp.WithString("source_name", mcp.Required(), mcp.Description("The database source (auto-filled)")),
	}

	// Add custom parameters
	for _, param := range parameters {
		toolOptions = append(toolOptions,
			mcp.WithString(param, mcp.Required(), mcp.Description(fmt.Sprintf("Parameter: %s", param))))
	}

	tool := mcp.NewTool(name, toolOptions...)

	// Store tool info
	ms.customTools[name] = customToolEntry{
		sourceName: sourceName,
		query:      query,
		parameters: parameters,
	}

	// Register handler
	ms.mcpServer.AddTool(tool, ms.handleCustomTool)

	return nil
}

func (ms *MCPServer) handleCustomTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get tool name from request (this is a bit of a hack, but MCP doesn't expose tool name in request)
	// We'll need to find which custom tool this is

	// For now, we'll iterate through custom tools and try to match
	for _, toolEntry := range ms.customTools {
		// Check if all parameters for this tool are present
		allParamsPresent := true
		paramValues := make(map[string]string)

		for _, param := range toolEntry.parameters {
			val, err := request.RequireString(param)
			if err != nil {
				allParamsPresent = false
				break
			}
			paramValues[param] = val
		}

		if !allParamsPresent {
			continue
		}

		// This is our tool!
		entry, exists := ms.sources[toolEntry.sourceName]
		if !exists {
			return mcp.NewToolResultError(fmt.Sprintf("Source not found: %s", toolEntry.sourceName)), nil
		}

		// Replace parameters in query
		query := toolEntry.query
		for param, value := range paramValues {
			// Simple string replacement (in production, use proper SQL parameter binding)
			query = strings.ReplaceAll(query, fmt.Sprintf("{{%s}}", param), value)
		}

		// Execute query
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

	return mcp.NewToolResultError("Custom tool not found or parameters missing"), nil
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

	// Enhanced query validation using AST parser
	if err := ms.queryValidator.ValidateQuery(query); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Query validation failed: %v", err)), nil
	}

	// Add row limit to prevent database overload
	modifiedQuery, err := ms.queryModifier.AddRowLimit(query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add row limit: %v", err)), nil
	}

	// Execute query
	result, err := entry.source.ExecuteQuery(ctx, modifiedQuery)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Query execution failed: %v", err)), nil
	}

	// Apply PII masking to results
	if ms.piiMasker != nil {
		for i := range result.Rows {
			for key, value := range result.Rows[i] {
				result.Rows[i][key] = ms.piiMasker.MaskValue(value)
			}
		}
	}

	jsonResult, err := json.MarshalIndent(result.Rows, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal result"), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (ms *MCPServer) handleListTables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourceName, err := request.RequireString("source_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	entry, exists := ms.sources[sourceName]
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("Source not found: %s", sourceName)), nil
	}

	schemas, err := entry.source.GetSchema(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get schema: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("# Tables in %s\n\n", sourceName))

	for _, table := range schemas {
		result.WriteString(fmt.Sprintf("## %s\n", table.Name))
		result.WriteString(fmt.Sprintf("- Columns: %d\n", len(table.Columns)))
		if len(table.PrimaryKeys) > 0 {
			result.WriteString(fmt.Sprintf("- Primary Keys: %s\n", strings.Join(table.PrimaryKeys, ", ")))
		}
		if len(table.ForeignKeys) > 0 {
			result.WriteString(fmt.Sprintf("- Foreign Keys: %d\n", len(table.ForeignKeys)))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (ms *MCPServer) handleDescribeTable(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourceName, err := request.RequireString("source_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	tableName, err := request.RequireString("table_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	entry, exists := ms.sources[sourceName]
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("Source not found: %s", sourceName)), nil
	}

	schemas, err := entry.source.GetSchema(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get schema: %v", err)), nil
	}

	// Find the requested table
	var tableSchema *core.TableSchema
	for _, schema := range schemas {
		if strings.EqualFold(schema.Name, tableName) {
			tableSchema = &schema
			break
		}
	}

	if tableSchema == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Table not found: %s", tableName)), nil
	}

	// Build detailed description
	var result strings.Builder
	result.WriteString(fmt.Sprintf("# Table: %s\n\n", tableSchema.Name))

	// Primary Keys
	if len(tableSchema.PrimaryKeys) > 0 {
		result.WriteString(fmt.Sprintf("**Primary Key(s):** %s\n\n", strings.Join(tableSchema.PrimaryKeys, ", ")))
	}

	// Columns
	result.WriteString("## Columns\n\n")
	result.WriteString("| Column | Type | Nullable | Description |\n")
	result.WriteString("|--------|------|----------|-------------|\n")

	for _, col := range tableSchema.Columns {
		nullable := "NO"
		if col.IsNullable {
			nullable = "YES"
		}
		desc := col.Description
		if desc == "" {
			desc = "-"
		}
		result.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", col.Name, col.DataType, nullable, desc))
	}
	result.WriteString("\n")

	// Foreign Keys
	if len(tableSchema.ForeignKeys) > 0 {
		result.WriteString("## Foreign Keys\n\n")
		for _, fk := range tableSchema.ForeignKeys {
			result.WriteString(fmt.Sprintf("- **%s** → %s.%s (%s)\n",
				fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn, fk.ConstraintName))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (ms *MCPServer) handleSchemaPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	if ms.schemaContext == "" {
		return &mcp.GetPromptResult{
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: "No database schema information is available yet. Please ensure sources are connected and LoadSchemas() has been called.",
					},
				},
			},
		}, nil
	}

	return &mcp.GetPromptResult{
		Description: "Complete database schema with tables, columns, data types, primary keys, foreign keys, and column descriptions",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: ms.schemaContext,
				},
			},
		},
	}, nil
}

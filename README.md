# CoreMCP

[![CI](https://github.com/corebasehq/coremcp/workflows/CI/badge.svg)](https://github.com/corebasehq/coremcp/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/github/go-mod/go-version/corebasehq/coremcp)](https://golang.org/doc/devel/release.html)
[![Release](https://img.shields.io/github/v/release/corebasehq/coremcp)](https://github.com/corebasehq/coremcp/releases)

**Model Context Protocol (MCP) server for database operations**

CoreMCP by CoreBaseHQ provides a secure, extensible bridge between AI assistants (like Claude Desktop) and your databases through the Model Context Protocol.

## 🚀 Features
**⚠️ Safety First: CoreMCP is designed to be Read-Only by default. We strongly recommend creating a specific database user with SELECT permissions only.**
- 🔌 **Multiple Database Support**: MSSQL, Firebird (coming soon), and extensible adapter system
- 🧠 **Automatic Schema Discovery**: CoreMCP automatically scans your database tables, columns, foreign keys, and descriptions to provide AI context
- 📝 **Column Comments Support**: Extracts and presents database column comments/descriptions to the AI for better query understanding
- 🛠️ **Dynamic Tool Generation**: Built-in tools for common operations (list tables, describe schema) plus custom tool support
- 🎯 **Custom Query Tools**: Define reusable SQL queries as MCP tools in your config file
- 🛡️ **Secure**: Read-only mode support, connection string isolation
- 🎯 **MCP Native**: Built specifically for Model Context Protocol
- 🔧 **Easy Configuration**: Simple YAML-based setup
- 📦 **Lightweight**: Single binary, no runtime dependencies

## 📋 Requirements

- Go 1.23 or higher (for building from source)
- Database drivers are embedded in the binary

## 🔧 Installation

### From Source

```bash
git clone https://github.com/corebasehq/coremcp.git
cd coremcp
go build -o coremcp ./cmd/coremcp
```

### Binary Release

Download the latest release from the [Releases page](https://github.com/corebasehq/coremcp/releases).

## ⚙️ Configuration

Create a `coremcp.yaml` file in your working directory:

```yaml
server:
  name: "coremcp-agent"
  version: "0.1.0"
  transport: "stdio"
  port: 8080

logging:
  level: "info"
  format: "json"

sources:
  - name: "my_database"
    type: "mssql"
    dsn: "sqlserver://username:password@localhost:1433?database=mydb&encrypt=disable"
    readonly: true
```

See [coremcp.example.yaml](coremcp.example.yaml) for more examples.

### DSN Format

**Microsoft SQL Server:**
```
sqlserver://username:password@host:port?database=dbname&encrypt=disable
```

**Dummy (for testing):**
```
dummy://test
```

### Security Configuration

CoreMCP includes enterprise-grade security features:

```yaml
security:
  # Maximum rows to return (prevents DB overload)
  max_row_limit: 1000
  
  # Enable PII masking
  enable_pii_masking: true
  
  # PII patterns to mask
  pii_patterns:
    - name: "credit_card"
      pattern: '\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b'
      replacement: "****-****-****-****"
      enabled: true
    - name: "email"
      pattern: '\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b'
      replacement: "***@***.***"
      enabled: true
    - name: "turkish_id"
      pattern: '\b[1-9]\d{10}\b'
      replacement: "***********"
      enabled: true
```

**Security Features:**
- **AST-Based Query Validation**: Uses sqlparser to analyze SQL queries and block dangerous operations (DROP, ALTER, UPDATE, DELETE, TRUNCATE, EXEC, etc.)
- **Automatic Row Limiting**: Adds LIMIT clause to prevent accidentally returning millions of rows
- **PII Data Masking**: Automatically masks sensitive data like credit cards, emails, SSNs, Turkish IDs, IBANs
- **Configurable Patterns**: Define custom regex patterns for your specific PII requirements

## 🎯 Usage

### Start the MCP Server

```bash
coremcp serve --config coremcp.yaml
```

Or use stdio transport (default):

```bash
coremcp serve -t stdio
```

### Use with Claude Desktop

Add to your Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "coremcp": {
      "command": "/path/to/coremcp",
      "args": ["serve", "-c", "/path/to/coremcp.yaml"],
      "env": {}
    }
  }
}
```

## 🏗️ Architecture

```
coremcp/
├── cmd/coremcp/       # CLI application entry point
│   ├── main.go        # Main entry
│   ├── root.go        # Root command
│   └── serve.go       # Serve command
├── pkg/
│   ├── adapter/       # Database adapters
│   │   ├── factory.go # Adapter factory pattern
│   │   ├── dummy/     # Dummy adapter (for testing)
│   │   └── mssql/     # MSSQL adapter
│   ├── config/        # Configuration management
│   ├── core/          # Core type definitions
│   └── server/        # MCP server implementation
└── coremcp.yaml       # Configuration file
```

## 🔌 Available Tools & Prompts

### Built-in Tools

#### `query_database`

Executes arbitrary SQL queries on configured database sources.

**Parameters:**
- `source_name` (required): Name of the database source from config
- `query` (required): SQL query to execute

**Example:**
```sql
SELECT * FROM users WHERE id = 1
```

#### `list_tables`

Lists all tables in a database with summary information.

**Parameters:**
- `source_name` (required): Name of the database source

**Returns:** List of tables with column counts, primary keys, and foreign key counts.

#### `describe_table`

Shows detailed schema information for a specific table.

**Parameters:**
- `source_name` (required): Name of the database source
- `table_name` (required): Name of the table to describe

**Returns:** Complete table schema including:
- Column names and data types
- Nullable information
- Primary keys
- Foreign key relationships
- Column descriptions/comments

### Custom Tools

You can define reusable SQL queries as custom MCP tools in your `coremcp.yaml`:

```yaml
custom_tools:
  - name: "get_daily_sales"
    description: "Retrieves daily sales summary for a specific date"
    source: "production_db"
    query: "SELECT * FROM orders WHERE DATE(created_at) = '{{date}}'"
    parameters:
      - name: "date"
        description: "Date in YYYY-MM-DD format"
        required: true

  - name: "get_top_customers"
    description: "Lists top N customers by order count"
    source: "production_db"
    query: "SELECT user_id, COUNT(*) as order_count FROM orders GROUP BY user_id ORDER BY order_count DESC LIMIT {{limit}}"
    parameters:
      - name: "limit"
        description: "Number of customers to return"
        required: true
        default: "10"
```

**Benefits:**
- Encapsulate complex queries
- Provide simple interfaces for common operations
- Parameters are automatically validated
- AI can discover and use these tools automatically

### `database_schema` Prompt

Automatically provides complete database schema context to the AI, including:
- Table names
- Column names with data types
- Primary keys
- Foreign key relationships
- Column descriptions/comments from the database

When CoreMCP starts, it automatically:
1. Connects to all configured databases
2. Scans the schema (tables, columns, keys, relationships)
3. Extracts column comments/descriptions (e.g., `MS_Description` in MSSQL)
4. Creates a comprehensive context prompt for the AI

This allows Claude to understand your database structure and write accurate queries without you having to explain the schema manually.

**Example:**
When you ask Claude "Show me all sales", Claude can see that you have a `TBLSATIS` table with specific columns and automatically write the correct query.

## 🛠️ Adding Custom Adapters

1. Create a new package in `pkg/adapter/yourdb/`
2. Implement the `core.Source` interface
3. Register in `pkg/adapter/factory.go`

See [pkg/adapter/dummy/dummy.go](pkg/adapter/dummy/dummy.go) for a simple example.

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 🔒 Security

For security concerns, please see [SECURITY.md](SECURITY.md).

## 📄 License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## 🌟 Roadmap

- [x] **Automatic Schema Discovery** - Load database structure on startup
- [x] **Column Comments/Descriptions** - Extract and display database metadata
- [x] **Dynamic Tool Generation** - list_tables, describe_table tools
- [x] **Custom Query Tools** - Define reusable queries in config
- [x] **AST-Based Query Sanitization** - Block dangerous SQL operations
- [x] **PII Data Masking** - Mask sensitive information in results
- [x] **Automatic Row Limiting** - Prevent database overload
- [ ] PostgreSQL adapter
- [ ] MySQL adapter
- [ ] Firebird adapter (in progress)
- [ ] Query result caching
- [ ] HTTP transport support
- [ ] Write operation support (with strict safety guards)
- [ ] Audit logging

## 💬 Support

- 🐛 [Report a bug](https://github.com/corebasehq/coremcp/issues/new?template=bug_report.md)
- 💡 [Request a feature](https://github.com/corebasehq/coremcp/issues/new?template=feature_request.md)
- 📧 Email: support@corebasehq.com

---

Made with ❤️ by [CoreBaseHQ](https://github.com/corebasehq)


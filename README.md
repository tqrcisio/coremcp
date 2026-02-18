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
- 🛡️ **NOLOCK / Read Uncommitted**: Per-source option to run all SELECT queries under `READ UNCOMMITTED` isolation (MSSQL `WITH (NOLOCK)` equivalent) for zero-locking reads on busy OLTP databases
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
    no_lock: true            # Optional: READ UNCOMMITTED isolation (WITH (NOLOCK) equivalent)
    normalize_turkish: true  # Optional: Turkish character normalization for legacy ERP databases
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

### Source Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `name` | string | — | Unique source identifier |
| `type` | string | — | Adapter type: `mssql`, `dummy` |
| `dsn` | string | — | Connection string |
| `readonly` | bool | `false` | Restrict to SELECT-only at the config level |
| `no_lock` | bool | `false` | **(MSSQL only)** Run all SELECT queries under `READ UNCOMMITTED` transaction isolation level. Equivalent to adding `WITH (NOLOCK)` to every table reference. Eliminates shared lock acquisition, improving read throughput on busy OLTP databases. **Trade-off:** may return dirty (uncommitted) rows. |
| `normalize_turkish` | bool | `false` | **(MSSQL only)** Enable Turkish character normalization middleware. **Outgoing:** Turkish chars inside SQL string literals are converted to ASCII uppercase before the query is sent (`'Hüseyin'` → `'HUSEYIN'`, `'Şeker'` → `'SEKER'`). **Incoming:** Windows-1254 / Windows-1252 mojibake in result strings is auto-corrected. Use this for legacy Turkish ERP databases with `Turkish_CI_AS` collation. |

#### Example: MSSQL with NOLOCK enabled

```yaml
sources:
  - name: "oltp_db"
    type: "mssql"
    dsn: "sqlserver://user:pass@localhost:1433?database=production&encrypt=disable"
    readonly: true
    no_lock: true
```

#### Example: Legacy Turkish ERP Database

```yaml
sources:
  - name: "erp_db"
    type: "mssql"
    dsn: "sqlserver://user:pass@localhost:1433?database=LOGO&encrypt=disable"
    readonly: true
    no_lock: true           # Avoid locking on busy OLTP
    normalize_turkish: true # AI can now search 'Hüseyin' and it matches 'HUSEYIN'
```

**How Turkish normalization works:**

| AI sends | Normalized query (sent to DB) | Why |
|----------|-------------------------------|-----|
| `WHERE ADI = 'Hüseyin'` | `WHERE ADI = 'HUSEYIN'` | ERP stores names as uppercase ASCII |
| `WHERE SEHIR LIKE '%şeker%'` | `WHERE SEHIR LIKE '%SEKER%'` | `Ş` → `S` |
| `WHERE SEHIR = 'İstanbul'` | `WHERE SEHIR = 'ISTANBUL'` | `İ` → `I` |

**Mojibake correction (incoming results):**

| DB returns (garbled) | Fixed output | Cause |
|----------------------|--------------|-------|
| `GÐKHAN` | `GĞKHAN` | Win-1254 byte 0xD0 read as Win-1252 |
| `ÝSTANBUL` | `İSTANBUL` | Win-1254 byte 0xDD read as Win-1252 |
| `ÞEHİR` | `ŞEHİR` | Win-1254 byte 0xDE read as Win-1252 |

## 🎯 Usage

CoreMCP has two operation modes:

### 1. Local Mode (serve) - For Claude Desktop

Start the MCP Server locally:

```bash
coremcp serve --config coremcp.yaml
```

Or use stdio transport (default):

```bash
coremcp serve -t stdio
```

##### Use with Claude Desktop

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

### 2. Remote Mode (connect) - For SaaS & Factory Deployments 🚇

Connect to CoreBase Cloud Platform for remote management:

```bash
coremcp connect --server="wss://api.corebase.com/ws/agent" --token="sk_fabrika_123"
```

**Perfect for:**
- 🏭 **Factory Deployments**: No need to open inbound ports
- 🌐 **Remote Management**: Control databases from anywhere
- 🔐 **Secure**: Agent initiates connection from inside your network
- 🔄 **Auto-Reconnect**: Automatic reconnection on network failures
- ⚙️ **Remote Config**: Update database connections without redeployment

#### Connect Command Options

```bash
Flags:
  -s, --server string              CoreBase Cloud WebSocket URL (required)
  -t, --token string               Authentication token (required)
  -a, --agent-id string            Agent ID (optional, auto-generated if not provided)
  -r, --max-reconnect int          Maximum reconnection attempts (default: 10, 0 for infinite)
  -d, --reconnect-delay duration   Delay between reconnection attempts (default: 5s)
```

#### Example: Factory Deployment

```bash
# Factory IT admin runs this command
./coremcp connect \
  --server="wss://api.corebasehq.com/ws/agent" \
  --token="sk_fabrika_xyz" \
  --agent-id="factory-istanbul-001" \
  --max-reconnect=0  # Infinite reconnection
```

**How it works:**
1. 🔌 Agent connects to CoreBase Cloud via WebSocket (outbound only)
2. 🔐 Authenticates with your API token
3. 📡 Receives commands from your CoreBase dashboard
4. 🎯 Executes SQL queries on local databases
5. 📤 Sends results back through the secure tunnel
6. 🔄 Auto-reconnects on connection loss

**Remote Commands Supported:**
- `run_sql`: Execute SQL queries remotely
- `get_schema`: Retrieve database schema
- `list_sources`: List connected databases
- `health_check`: Check agent status
- `config_sync`: Update database configurations remotely

**No Port Forwarding Required!** 🎉

## 🏗️ Architecture

```
coremcp/
├── cmd/coremcp/       # CLI application entry point
│   ├── main.go        # Main entry
│   ├── root.go        # Root command
│   ├── serve.go       # Serve command (stdio mode for Claude Desktop)
│   └── connect.go     # Connect command (WebSocket mode for Cloud)
├── pkg/
│   ├── adapter/       # Database adapters
│   │   ├── factory.go # Adapter factory pattern
│   │   ├── dummy/     # Dummy adapter (for testing)
│   │   └── mssql/     # MSSQL adapter
│   ├── config/        # Configuration management
│   ├── core/          # Core type definitions
│   ├── security/      # Security features (PII masking, query validation)
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

#### `list_views`

Lists all views in a database with their column definitions.

**Parameters:**
- `source_name` (required): Name of the database source

**Returns:** Each view with its column names, types, and nullability.

#### `list_procedures`

Lists all stored procedures in a database with parameter details.

**Parameters:**
- `source_name` (required): Name of the database source

**Returns:** Each procedure with parameter names, types, modes (`IN`/`OUT`/`INOUT`) and a ready-to-copy example call.

#### `execute_procedure`

Executes a stored procedure by name with optional named parameters.
> ⚠️ Only available on sources where `readonly: false`.

**Parameters:**
- `source_name` (required): Name of the database source
- `procedure_name` (required): Stored procedure name (e.g. `sp_CiroHesapla`)
- `params` (optional): JSON string of parameter name/value pairs

**Security:**
- Procedure name validated against `^[a-zA-Z_][a-zA-Z0-9_#@.]*$`
- All values passed as named SQL parameters (`sql.Named`) — no string interpolation
- Parameter names also validated (alphanumeric + underscore only)
- Completely blocked when source is `readonly: true`

**Example:**
```json
{
  "source_name": "erp_db",
  "procedure_name": "sp_CiroHesapla",
  "params": "{\"StartDate\":\"2024-01-01\",\"EndDate\":\"2024-12-31\"}"
}
```

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
- [x] **WebSocket Connect Mode** - Remote management via CoreBase Cloud 🎉
- [x] **Auto-Reconnection Logic** - Resilient agent connections
- [x] **Remote Configuration Sync** - Update database configs remotely
- [x] **NOLOCK / Read Uncommitted** - Per-source zero-locking reads for MSSQL
- [x] **Turkish Character Normalization** - SQL literal normalization + mojibake fix for legacy Turkish ERP databases
- [x] **View & Stored Procedure Discovery** - `list_views`, `list_procedures`, `execute_procedure` tools; auto-included in schema context
- [ ] PostgreSQL adapter
- [ ] MySQL adapter
- [ ] Firebird adapter (in progress)
- [ ] Query result caching
- [ ] HTTP transport support
- [ ] Write operation support (with strict safety guards)
- [ ] Audit logging
- [ ] Multi-tenant agent management
- [ ] Real-time monitoring dashboard

## 💬 Support

- 🐛 [Report a bug](https://github.com/corebasehq/coremcp/issues/new?template=bug_report.md)
- 💡 [Request a feature](https://github.com/corebasehq/coremcp/issues/new?template=feature_request.md)
- 📧 Email: support@corebasehq.com

---

Made with ❤️ by [CoreBaseHQ](https://github.com/corebasehq)


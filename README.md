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

## 🔌 Available Tools

### `query_database`

Executes SQL queries on configured database sources.

**Parameters:**
- `source_name` (required): Name of the database source from config
- `query` (required): SQL query to execute

**Example:**
```sql
SELECT * FROM users WHERE id = 1
```

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

- [ ] PostgreSQL adapter
- [ ] MySQL adapter
- [ ] Firebird adapter (in progress)
- [ ] Schema introspection tools
- [ ] Query result caching
- [ ] HTTP transport support
- [ ] Write operation support (with strict safety guards)

## 💬 Support

- 🐛 [Report a bug](https://github.com/corebasehq/coremcp/issues/new?template=bug_report.md)
- 💡 [Request a feature](https://github.com/corebasehq/coremcp/issues/new?template=feature_request.md)
- 📧 Email: support@corebase.com

---

Made with ❤️ by [CoreBaseHQ](https://github.com/corebasehq)


# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2026-02-15

### Fixed
- **Code Quality**: Fixed golangci-lint errcheck warnings in test files
  - Added proper error handling for `src.Connect()` calls in all test functions
  - Added error checks for `mcpSrv.LoadSchemas()` calls in schema tests
  - Improved test reliability with explicit error validation

## [0.2.0] - 2026-02-15

### Added
- **Enterprise-Grade Security Features:**
  - **AST-Based Query Sanitization**: Uses sqlparser library for deep SQL analysis
    - Blocks dangerous operations: INSERT, UPDATE, DELETE, DROP, ALTER, TRUNCATE, EXEC, MERGE, etc.
    - Allows only SELECT and WITH (CTE) queries
    - Validates queries at AST level, not just regex
  - **PII Data Masking**: Automatically masks personally identifiable information
    - Built-in patterns: Credit cards, emails, SSNs, Turkish IDs, IBANs, phone numbers
    - Configurable custom patterns via regex
    - Enable/disable per pattern
    - Enterprise-ready for GDPR/KVKK compliance
  - **Automatic Row Limiting**: Prevents database overload
    - Configurable max row limit (default: 1000)
    - Automatically adds LIMIT clause to queries
    - Preserves existing LIMIT if present
- **Dynamic Tool Generation:**
  - New `list_tables` tool: Lists all tables in a database with summary information
  - New `describe_table` tool: Shows detailed schema for a specific table
  - Custom tool support: Define reusable SQL queries as MCP tools in config
  - Parameter substitution in custom queries using `{{parameter}}` syntax
- **Automatic Schema Discovery:** CoreMCP now automatically scans database tables, columns, primary keys, and foreign keys on startup
- **Column Comments/Descriptions Support:** Extracts and presents database column comments (e.g., MS_Description in MSSQL) to AI for better context
- **Schema Context Prompt:** New `database_schema` MCP prompt that provides complete database structure to Claude automatically
- **Enhanced Schema Types:** 
  - New `ColumnInfo` type with name, data type, nullable flag, and description
  - New `ForeignKey` type with full relationship information
  - Enhanced `TableSchema` with primary keys array
- **MSSQL Extended Support:**
  - Automatic extraction of column descriptions from MS_Description extended properties
  - Full primary key discovery
  - Complete foreign key relationship mapping

### Changed
- **Query Execution**: Now uses security validator and row limiter for all queries
- **Config Structure**: Added `security` section with PII masking and row limit configuration
- **MCP Server**: Added `list_tables` and `describe_table` handlers
- **Config Structure**: Added `custom_tools` section for defining custom query tools
- **MSSQL Adapter:** Enhanced `GetSchema()` to retrieve complete table metadata including types, constraints, and relationships
- **MCP Server:** Added automatic schema loading during startup via `LoadSchemas()` method
- **Core Types:** Updated `TableSchema` structure to support rich metadata

### Security
- **BREAKING**: All queries now validated with AST parser (blocks write operations)
- **BREAKING**: Automatic LIMIT clause added to all SELECT queries (configurable)
- PII masking can be enabled via `security.enable_pii_masking` in config

## [0.1.0] - 2026-02-14

### Added
- **Core Server:** Implemented Model Context Protocol (MCP) server over stdio transport.
- **Database Support:**
  - Added MSSQL (Microsoft SQL Server) adapter with full query execution and schema discovery.
  - Added Dummy adapter for testing and demonstration purposes.
- **Configuration:**
  - YAML-based configuration system (`coremcp.yaml`).
  - Environment variable support with COREMCP_ prefix.
  - Support for multiple database sources in single config.
- **CLI:**
  - `serve` command to start the MCP agent.
  - Version flag support.
  - Configurable transport (stdio/http).
- **DevOps & CI/CD:**
  - GitHub Actions pipeline for automated testing and linting.
  - `Makefile` for cross-platform builds (Linux, macOS, Windows, ARM64).
  - `Dockerfile` for containerized deployment.
  - Docker Compose configuration for easy local testing.
- **Documentation:**
  - Comprehensive `README.md` with architecture diagrams.
  - `CONTRIBUTING.md` guide for open-source contributors.
  - `SECURITY.md` policy for responsible disclosure.
  - `CODE_OF_CONDUCT.md` for community standards.
  - Example configuration files.

### Security
- Implemented strict read-only mode for database connections.
- SQL query validation to prevent destructive operations (INSERT, UPDATE, DELETE, DROP, etc.).
- Whole-word matching for SQL keyword detection to avoid false positives.
- Support for SELECT and WITH (CTE) queries only in read-only mode.

### Fixed
- **Critical:** Fixed JSON-RPC protocol corruption by redirecting all adapter debug output to stderr instead of stdout.
  - Dummy adapter now correctly uses `os.Stderr` for all log messages.
  - Ensures stdio transport compatibility with Claude Desktop and other MCP clients.

## [0.0.1] - 2026-02-13

### Added
- Initial repository initialization.
- Project structure and Go module setup.
- Basic license (Apache 2.0) and gitignore files.
- Core package interfaces and types.

[Unreleased]: https://github.com/corebasehq/coremcp/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/corebasehq/coremcp/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/corebasehq/coremcp/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/corebasehq/coremcp/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/corebasehq/coremcp/releases/tag/v0.0.1
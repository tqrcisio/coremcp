# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-02-18

### Added
- **NOLOCK / READ UNCOMMITTED Support (MSSQL):** New per-source `no_lock` config option
  - When `no_lock: true`, all SELECT queries are executed inside a `READ UNCOMMITTED` transaction — functionally identical to adding `WITH (NOLOCK)` to every table reference
  - Prevents shared lock acquisition on MSSQL, reducing contention on busy OLTP databases
  - Transaction is always rolled back after each read; no data is ever modified
  - Trade-off: may return dirty (uncommitted) data — use only on reporting / AI-assistant workloads
  - Supported in both `serve` (local) mode and `connect` (remote/WebSocket) mode
  - New field added to `SourceConfig` (`no_lock: bool`) and `RemoteSource` (`no_lock: bool` in JSON)
- **Turkish Character Normalization Middleware (MSSQL):** New per-source `normalize_turkish` config option
  - Solves the "collation war" problem in legacy Turkish ERP databases (`Turkish_CI_AS` / Windows-1254)
  - **Outgoing query normalization:** Turkish characters inside SQL single-quoted string literals are automatically converted to their ASCII uppercase equivalents before the query is sent
    - Example: `WHERE ADI = 'Hüseyin'` → `WHERE ADI = 'HUSEYIN'`
    - Example: `WHERE SEHIR LIKE '%şeker%'` → `WHERE SEHIR LIKE '%SEKER%'`
    - Only string literals are modified — SQL keywords, column names and query structure are untouched
    - Mapping: `İıŞşGğğÜüÖöÇç` → `I I S S G G U U O O C C` (plus circumflex variants `ÂâÎîÛû`)
  - **Incoming result fix (mojibake correction):** Common Windows-1254 bytes misread as Windows-1252 are automatically corrected in result strings
    - `Ð` → `Gğ`, `Ý` → `İ`, `Þ` → `Ş`, `ð` → `ğ`, `ý` → `ı`, `þ` → `ş`
  - New package `pkg/turkish` with exported functions: `NormalizeSQLLiterals`, `ToASCIIUpper`, `FixMojibake`, `FixResultValue`
  - Works together with `no_lock: true` for maximum compatibility on legacy OLTP databases
  - Supported in both `serve` and `connect` modes
- **View & Stored Procedure Support (MSSQL):** AI agents can now discover and execute views and stored procedures
  - New `Source` interface methods: `GetViews`, `GetProcedures`, `ExecuteProcedure`
  - `GetViews` — fetches all database views with their column definitions via `INFORMATION_SCHEMA.VIEWS`
  - `GetProcedures` — fetches all stored procedures with parameter names, types, and modes (`IN`/`OUT`/`INOUT`) via `INFORMATION_SCHEMA.ROUTINES` + `INFORMATION_SCHEMA.PARAMETERS`
  - `ExecuteProcedure` — executes a stored procedure safely:
    - Procedure name is validated against `^[a-zA-Z_][a-zA-Z0-9_#@.]*$` (SQL injection prevention)
    - All parameter values are passed as named SQL parameters (`sql.Named`) — no string interpolation
    - Individual parameter names are also validated against `^[a-zA-Z_][a-zA-Z0-9_]*$`
    - Blocked entirely when the source is `readonly: true`
  - **Three new MCP tools** exposed to AI agents:
    - `list_views` — lists views with columns in markdown table format
    - `list_procedures` — lists stored procedures with parameters and an inline example call hint
    - `execute_procedure` — runs a procedure; accepts `params` as a JSON string e.g. `{"StartDate":"2024-01-01","Limit":"10"}`
  - Views and stored procedures are also included in the schema context built by `LoadSchemas`, so agents see them automatically without calling a tool
  - Turkish mojibake correction (`normalize_turkish`) is applied to procedure result sets as well
  - Supported in both `serve` and `connect` modes
- **SQL Server Version Compatibility:** Automatic version detection and query adaptation for legacy SQL Server deployments
  - On `Connect()`, CoreMCP queries `SERVERPROPERTY('ProductVersion')` and stores the major version number
  - `Name()` now reports the detected release (e.g. `"MSSQL (SQL Server 2008/2008 R2)"`) in logs and schema context so the AI knows which SQL dialect to generate
  - **SQL Server 2000 (v8) compatibility:** all schema-discovery queries branch to legacy system tables:
    - `GetSchema` uses `sysobjects WHERE xtype = 'U'` instead of `INFORMATION_SCHEMA.TABLES`
    - Column query uses `INFORMATION_SCHEMA.COLUMNS` only (no `sys.extended_properties` join — column descriptions silently omitted)
    - FK query uses `sysforeignkeys` / `COL_NAME()` instead of `sys.foreign_keys` / `sys.columns`
  - **LIMIT → TOP rewrite (all versions):** `security.QueryModifier` uses MySQL-dialect sqlparser and may inject `LIMIT N`; the MSSQL adapter now rewrites this to `SELECT TOP N` before execution
  - **OFFSET FETCH → TOP rewrite (pre-2012):** AI-generated `OFFSET x ROWS FETCH NEXT N ROWS ONLY` clauses are automatically stripped and replaced with `SELECT TOP N` on SQL Server 2008/2008 R2 and older (major version < 11)
  - Version constants: 8=2000, 9=2005, 10=2008/R2, 11=2012, 12=2014, 13=2016, 14=2017, 15=2019, 16=2022

### Changed
- `adapter.NewSource` signature updated to accept `noLock bool` and `normalizeTurkish bool` parameters
- `mssql.New` signature updated to accept `noLock bool` and `normalizeTurkish bool` parameters
- Startup log now reports `[NoLock: true/false, NormalizeTurkish: true/false]` alongside source details

## [0.3.0] - 2026-02-16

### Added
- **Remote Connect Mode**: WebSocket-based cloud connectivity for SaaS deployments
  - New `connect` command for remote management
  - WebSocket client implementation with gorilla/websocket
  - Authentication with API tokens
  - Support for remote commands: `run_sql`, `get_schema`, `list_sources`, `health_check`
  - Auto-Reconnection Logic: Configurable retry mechanism with exponential backoff
  - Remote Configuration Sync: Update database connections without redeployment
  - Heartbeat/ping-pong for connection monitoring
  - Perfect for factory deployments (no inbound ports required!)
  - Agent-initiated outbound connections only
  - Graceful shutdown handling

### Changed
- Architecture updated to support both stdio (local) and WebSocket (remote) modes
- README updated with comprehensive `connect` mode documentation
- Removed emojis from log messages for professional enterprise-grade logging
- Log messages now use standard prefixes: [INFO], [ERROR], [WARN], [DEBUG]

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
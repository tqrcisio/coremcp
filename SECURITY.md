# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **security@corebasehq.com**

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

Please include the following information:

- Type of vulnerability
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

This information will help us triage your report more quickly.

## Disclosure Policy

When we receive a security bug report, we will:

1. Confirm the problem and determine affected versions
2. Audit code to find any similar problems
3. Prepare fixes for all supported releases
4. Release patched versions as soon as possible

## Security Best Practices

When using CoreMCP:

### Configuration Security

1. **Never commit sensitive credentials** to version control
   - Use `.gitignore` to exclude `coremcp.yaml`
   - Use environment variables for sensitive data
   
2. **Use read-only database accounts** whenever possible
   ```yaml
   sources:
     - name: "production_db"
       readonly: true  # Enforce read-only mode
   ```

3. **Restrict database permissions**
   - Grant only SELECT permissions to the database user
   - Use dedicated service accounts with minimal privileges

4. **Secure connection strings**
   - Use encrypted connections when possible
   - For MSSQL: Add `encrypt=true` to DSN
   - Store DSN strings in secure configuration management

### Network Security

1. **Use stdio transport** for local Claude Desktop integration
   - No network exposure
   - Recommended for most use cases

2. **If using HTTP transport** (future):
   - Use HTTPS only
   - Implement authentication
   - Use firewall rules to restrict access

### Operational Security

1. **Keep CoreMCP updated** to the latest version
2. **Monitor logs** for suspicious activity
3. **Audit database access** regularly
4. **Use principle of least privilege**

### Example Secure Configuration

```yaml
server:
  name: "coremcp-agent"
  version: "0.1.0"
  transport: "stdio"  # Secure local transport

logging:
  level: "info"
  format: "json"

sources:
  - name: "production_db"
    type: "mssql"
    # Use environment variable: ${COREMCP_SOURCES_0_DSN}
    dsn: "sqlserver://readonly_user:${DB_PASSWORD}@localhost:1433?database=mydb&encrypt=true"
    readonly: true  # Enforce read-only
```

### Environment Variables

Set sensitive values via environment:

```bash
export COREMCP_SOURCES_0_DSN="sqlserver://user:password@host:1433?database=db&encrypt=true"
./coremcp serve
```

## Known Security Considerations

### SQL Injection Prevention

- CoreMCP passes queries directly to the database
- The AI assistant constructs SQL queries
- **Use read-only database accounts** to limit potential damage
- Consider implementing query allowlists for production use

### Data Exposure

- Query results are returned to the AI assistant
- Be aware of what data is accessible
- Use database views to limit data exposure
- Apply row-level security in your database

## Security Updates

Security updates will be released as patch versions (e.g., 0.1.1, 0.1.2) and announced via:

- GitHub Security Advisories
- Release notes
- Email to registered users (if applicable)

## Comments on This Policy

If you have suggestions on how this policy could be improved, please submit a pull request or open an issue.

---

Last updated: February 14, 2026

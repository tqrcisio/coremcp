# Contributing to CoreMCP

Thank you for considering contributing to CoreMCP! We welcome contributions from the community.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/corebasehq/coremcp/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Your environment (OS, Go version, etc.)
   - Relevant logs or screenshots

### Suggesting Features

1. Open an issue with the `enhancement` label
2. Describe the feature and its use case
3. Explain why it would be valuable

### Pull Requests

1. **Fork** the repository
2. **Create a branch** for your feature (`git checkout -b feature/amazing-feature`)
3. **Make your changes**:
   - Write clean, readable code
   - Follow Go best practices and conventions
   - Add tests for new functionality
   - Update documentation as needed
4. **Test your changes**:
   ```bash
   go test ./...
   go vet ./...
   ```
5. **Commit** with clear messages:
   ```bash
   git commit -m "feat: add PostgreSQL adapter support"
   ```
6. **Push** to your fork:
   ```bash
   git push origin feature/amazing-feature
   ```
7. **Open a Pull Request** with:
   - Clear description of changes
   - Reference to related issues
   - Test results

## Development Setup

```bash
# Clone the repository
git clone https://github.com/corebasehq/coremcp.git
cd coremcp

# Install dependencies
go mod download

# Build
go build -o coremcp ./cmd/coremcp

# Run tests
go test ./...

# Run with example config
cp coremcp.example.yaml coremcp.yaml
# Edit coremcp.yaml with your settings
./coremcp serve
```

## Code Style

- Follow standard Go formatting (`gofmt`, `goimports`)
- Write GoDoc comments for public functions and types
- Keep functions small and focused
- Use meaningful variable names
- Handle errors explicitly

## Adding a New Database Adapter

1. Create a new package in `pkg/adapter/yourdb/`
2. Implement the `core.Source` interface:
   ```go
   type Source interface {
       Name() string
       Connect(ctx context.Context) error
       Close(ctx context.Context) error
       GetSchema(ctx context.Context) ([]TableSchema, error)
       ExecuteQuery(ctx context.Context, query string, args ...any) (*QueryResult, error)
   }
   ```
3. Register your adapter in `pkg/adapter/factory.go`
4. Add tests in `yourdb_test.go`
5. Update documentation

Example: See [pkg/adapter/dummy/dummy.go](pkg/adapter/dummy/dummy.go)

## Testing

- Write unit tests for all new code
- Aim for >70% code coverage
- Include integration tests where applicable
- Test edge cases and error handling

## Commit Message Format

Follow conventional commits:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions or changes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

Example: `feat: add PostgreSQL connection pooling`

## Code Review

All submissions require review. We'll:
- Check code quality and style
- Verify tests pass
- Review documentation
- Ensure backwards compatibility

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

## Questions?

Feel free to:
- Open an issue for discussion
- Email: support@corebasehq.com
- Check existing issues and PRs

Thank you for contributing! 🎉

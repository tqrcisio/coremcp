// Package adapter provides a factory for creating database adapters.
package adapter

import (
	"fmt"

	"github.com/corebasehq/coremcp/pkg/adapter/dummy"
	"github.com/corebasehq/coremcp/pkg/adapter/mssql" // Add our new package
	"github.com/corebasehq/coremcp/pkg/core"
)

// NewSource creates a new database adapter based on the specified type.
// Supported types: "dummy", "mssql", "firebird" (coming soon).
// Returns an error if the database type is unsupported or initialization fails.
func NewSource(dbType string, dsn string) (core.Source, error) {
	switch dbType {
	case "dummy":
		return dummy.New(dsn)
	case "mssql":
		return mssql.New(dsn)
	case "firebird":
		return nil, fmt.Errorf("Firebird is not implemented yet")
	default:
		return nil, fmt.Errorf("Unsupported DB type: %s", dbType)
	}
}

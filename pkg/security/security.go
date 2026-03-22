// Package security provides query sanitization and data masking functionality.
package security

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xwb1989/sqlparser"
)

// QueryValidator validates SQL queries for security purposes.
type QueryValidator struct {
	allowedKeywords []string
	blockedKeywords []string
}

// NewQueryValidator creates a new query validator with custom allowed/blocked keywords.
func NewQueryValidator(allowedKeywords, blockedKeywords []string) *QueryValidator {
	return &QueryValidator{
		allowedKeywords: allowedKeywords,
		blockedKeywords: blockedKeywords,
	}
}

// ValidateQuery validates if a SQL query is safe to execute.
// Uses sqlparser for AST-based analysis to detect dangerous operations.
func (qv *QueryValidator) ValidateQuery(query string) error {
	// First, try to parse the query with sqlparser
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		// If parsing fails, fall back to regex-based validation
		return qv.validateWithRegex(query)
	}

	// Check statement type
	switch s := stmt.(type) {
	case *sqlparser.Select:
		// SELECT is allowed
		return qv.validateSelectStatement(s)
	case *sqlparser.Union:
		// UNION is allowed (it's just multiple SELECTs)
		return nil
	case *sqlparser.Insert, *sqlparser.Update, *sqlparser.Delete:
		return fmt.Errorf("write operations (INSERT/UPDATE/DELETE) are not allowed")
	case *sqlparser.DDL:
		return fmt.Errorf("DDL operations (CREATE/ALTER/DROP/TRUNCATE) are not allowed")
	case *sqlparser.OtherAdmin:
		return fmt.Errorf("administrative operations are not allowed")
	default:
		return fmt.Errorf("unsupported query type: %T", stmt)
	}
}

// validateSelectStatement performs deeper validation on SELECT statements.
func (qv *QueryValidator) validateSelectStatement(_ *sqlparser.Select) error {
	// Check for subqueries that might contain dangerous operations
	// This is a simplified check - sqlparser already ensures SELECT-only in subqueries

	// Additional custom validations can be added here
	// For example, checking for specific function calls, etc.

	return nil
}

// validateWithRegex performs regex-based validation as a fallback.
func (qv *QueryValidator) validateWithRegex(query string) error {
	q := strings.TrimSpace(strings.ToUpper(query))

	// Allow only SELECT and WITH (CTE) statements
	if !strings.HasPrefix(q, "SELECT") && !strings.HasPrefix(q, "WITH") {
		return fmt.Errorf("only SELECT and WITH queries are allowed")
	}

	// Default blocked keywords (can be extended via config)
	defaultBlocked := []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE",
		"CREATE", "GRANT", "REVOKE", "EXEC", "EXECUTE", "MERGE",
		"REPLACE", "RENAME", "CALL", "LOAD", "COPY",
	}

	// Merge with custom blocked keywords
	allBlocked := append(defaultBlocked, qv.blockedKeywords...)

	for _, word := range allBlocked {
		// Check for whole word matches to avoid false positives
		pattern := fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(word))
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return fmt.Errorf("blocked keyword detected: %s", word)
		}
	}

	return nil
}

// PIIMasker masks personally identifiable information in query results.
type PIIMasker struct {
	patterns     []*regexp.Regexp
	replacements []string
	enabled      bool
}

// MaskPattern represents a PII masking pattern.
type MaskPattern struct {
	Name        string
	Pattern     string
	Replacement string
	Enabled     bool
}

// NewPIIMasker creates a new PII masker with configured patterns.
func NewPIIMasker(patterns []MaskPattern, enabled bool) (*PIIMasker, error) {
	if !enabled {
		return &PIIMasker{enabled: false}, nil
	}

	masker := &PIIMasker{
		patterns:     make([]*regexp.Regexp, 0),
		replacements: make([]string, 0),
		enabled:      true,
	}

	for _, p := range patterns {
		if !p.Enabled {
			continue
		}

		regex, err := regexp.Compile(p.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern for %s: %w", p.Name, err)
		}

		masker.patterns = append(masker.patterns, regex)

		replacement := p.Replacement
		if replacement == "" {
			replacement = "***MASKED***"
		}
		masker.replacements = append(masker.replacements, replacement)
	}

	return masker, nil
}

// MaskData masks PII data in a string.
func (pm *PIIMasker) MaskData(data string) string {
	if !pm.enabled {
		return data
	}

	result := data
	for i, pattern := range pm.patterns {
		result = pattern.ReplaceAllString(result, pm.replacements[i])
	}

	return result
}

// MaskValue masks PII in any value (handles different types).
func (pm *PIIMasker) MaskValue(value interface{}) interface{} {
	if !pm.enabled {
		return value
	}

	switch v := value.(type) {
	case string:
		return pm.MaskData(v)
	case []byte:
		return pm.MaskData(string(v))
	default:
		return value
	}
}

// DefaultPIIPatterns returns commonly used PII masking patterns.
func DefaultPIIPatterns() []MaskPattern {
	return []MaskPattern{
		{
			Name:        "credit_card",
			Pattern:     `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`,
			Replacement: "****-****-****-****",
			Enabled:     true,
		},
		{
			Name:        "ssn_us",
			Pattern:     `\b\d{3}-\d{2}-\d{4}\b`,
			Replacement: "***-**-****",
			Enabled:     true,
		},
		{
			Name:        "email",
			Pattern:     `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			Replacement: "***@***.***",
			Enabled:     true,
		},
		{
			Name:        "phone_us",
			Pattern:     `\b\d{3}[-.\s]?\d{3}[-.\s]?\d{4}\b`,
			Replacement: "***-***-****",
			Enabled:     true,
		},
		{
			Name:        "turkish_id",
			Pattern:     `\b[1-9]\d{10}\b`,
			Replacement: "***********",
			Enabled:     true,
		},
		{
			Name:        "iban",
			Pattern:     `\b[A-Z]{2}\d{2}[A-Z0-9]{1,30}\b`,
			Replacement: "********************",
			Enabled:     true,
		},
	}
}

// QueryModifier modifies queries to add safety constraints.
type QueryModifier struct {
	maxRowLimit int
}

// NewQueryModifier creates a new query modifier.
func NewQueryModifier(maxRowLimit int) *QueryModifier {
	if maxRowLimit <= 0 {
		maxRowLimit = 1000 // Default limit
	}
	return &QueryModifier{
		maxRowLimit: maxRowLimit,
	}
}

// AddRowLimit adds a LIMIT clause to a query if it doesn't already have one.
func (qm *QueryModifier) AddRowLimit(query string) (string, error) {
	// Parse the query
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		// Fallback to simple string append if parsing fails
		return qm.addRowLimitSimple(query), nil
	}

	switch s := stmt.(type) {
	case *sqlparser.Select:
		// If the query already has a LIMIT that is within the allowed maximum, keep it.
		// If the existing limit exceeds the maximum, override it to prevent overload.
		if s.Limit != nil && s.Limit.Rowcount != nil {
			if limVal, ok := s.Limit.Rowcount.(*sqlparser.SQLVal); ok && limVal.Type == sqlparser.IntVal {
				existing, err := strconv.Atoi(string(limVal.Val))
				if err == nil && existing <= qm.maxRowLimit {
					return query, nil
				}
			}
		}

		// Add or override LIMIT clause, preserving any existing OFFSET
		var existingOffset sqlparser.Expr
		if s.Limit != nil {
			existingOffset = s.Limit.Offset
		}

		s.Limit = &sqlparser.Limit{
			Offset:   existingOffset,
			Rowcount: sqlparser.NewIntVal([]byte(fmt.Sprintf("%d", qm.maxRowLimit))),
		}

		return sqlparser.String(s), nil

	case *sqlparser.Union:
		// Union queries - add limit at the end
		return query + fmt.Sprintf(" LIMIT %d", qm.maxRowLimit), nil

	default:
		return query, nil
	}
}

// addRowLimitSimple adds row limit using simple string manipulation.
func (qm *QueryModifier) addRowLimitSimple(query string) string {
	q := strings.TrimSpace(query)
	qUpper := strings.ToUpper(q)

	// Check if LIMIT already exists
	if strings.Contains(qUpper, "LIMIT") {
		return query
	}

	// Add LIMIT at the end
	return fmt.Sprintf("%s LIMIT %d", q, qm.maxRowLimit)
}

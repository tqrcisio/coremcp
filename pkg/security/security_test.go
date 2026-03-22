package security

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestQueryValidator_ValidateQuery(t *testing.T) {
	validator := NewQueryValidator(nil, nil)

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "Valid SELECT query",
			query:   "SELECT * FROM users WHERE id = 1",
			wantErr: false,
		},
		{
			name:    "Valid SELECT with JOIN",
			query:   "SELECT u.*, o.* FROM users u JOIN orders o ON u.id = o.user_id",
			wantErr: false,
		},
		{
			name:    "Valid WITH (CTE) query",
			query:   "WITH cte AS (SELECT * FROM users) SELECT * FROM cte",
			wantErr: false,
		},
		{
			name:    "Block INSERT",
			query:   "INSERT INTO users (name) VALUES ('hacker')",
			wantErr: true,
		},
		{
			name:    "Block UPDATE",
			query:   "UPDATE users SET password = 'hacked' WHERE id = 1",
			wantErr: true,
		},
		{
			name:    "Block DELETE",
			query:   "DELETE FROM users WHERE id = 1",
			wantErr: true,
		},
		{
			name:    "Block DROP",
			query:   "DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "Block ALTER",
			query:   "ALTER TABLE users ADD COLUMN hacked VARCHAR(255)",
			wantErr: true,
		},
		{
			name:    "Block TRUNCATE",
			query:   "TRUNCATE TABLE users",
			wantErr: true,
		},
		{
			name:    "Block EXEC",
			query:   "EXEC sp_executesql N'DROP TABLE users'",
			wantErr: true,
		},
		{
			name:    "Block EXECUTE",
			query:   "EXECUTE sp_malicious",
			wantErr: true,
		},
		{
			name:    "Allow column named 'updated_at'",
			query:   "SELECT id, updated_at FROM users",
			wantErr: false,
		},
		{
			name:    "Allow column named 'deleted'",
			query:   "SELECT id, deleted FROM users WHERE deleted = 0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPIIMasker_MaskData(t *testing.T) {
	patterns := []MaskPattern{
		{
			Name:        "credit_card",
			Pattern:     `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`,
			Replacement: "****-****-****-****",
			Enabled:     true,
		},
		{
			Name:        "email",
			Pattern:     `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			Replacement: "***@***.***",
			Enabled:     true,
		},
		{
			Name:        "turkish_id",
			Pattern:     `\b[1-9]\d{10}\b`,
			Replacement: "***********",
			Enabled:     true,
		},
	}

	masker, err := NewPIIMasker(patterns, true)
	if err != nil {
		t.Fatalf("Failed to create PIIMasker: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Mask credit card",
			input:    "My card is 1234 5678 9012 3456",
			expected: "My card is ****-****-****-****",
		},
		{
			name:     "Mask email",
			input:    "Contact: john.doe@example.com",
			expected: "Contact: ***@***.***",
		},
		{
			name:     "Mask Turkish ID",
			input:    "TC: 12345678901",
			expected: "TC: ***********",
		},
		{
			name:     "Mask multiple PIIs",
			input:    "Email: test@test.com, Card: 4111 1111 1111 1111",
			expected: "Email: ***@***.***, Card: ****-****-****-****",
		},
		{
			name:     "No PII to mask",
			input:    "This is a normal text without PII",
			expected: "This is a normal text without PII",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.MaskData(tt.input)
			if result != tt.expected {
				t.Errorf("MaskData() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPIIMasker_Disabled(t *testing.T) {
	masker, err := NewPIIMasker(DefaultPIIPatterns(), false)
	if err != nil {
		t.Fatalf("Failed to create PIIMasker: %v", err)
	}

	input := "Email: test@test.com, Card: 1234 5678 9012 3456"
	result := masker.MaskData(input)

	if result != input {
		t.Errorf("Disabled masker should return original data, got %v", result)
	}
}

func TestQueryModifier_AddRowLimit(t *testing.T) {
	modifier := NewQueryModifier(100)

	tests := []struct {
		name         string
		query        string
		expectLimit  bool
		maxRowsInSQL int
	}{
		{
			name:         "Add LIMIT to simple SELECT",
			query:        "SELECT * FROM users",
			expectLimit:  true,
			maxRowsInSQL: 100,
		},
		{
			name:         "Add LIMIT to SELECT with WHERE",
			query:        "SELECT * FROM users WHERE active = 1",
			expectLimit:  true,
			maxRowsInSQL: 100,
		},
		{
			name:         "Keep existing LIMIT",
			query:        "SELECT * FROM users LIMIT 10",
			expectLimit:  true,
			maxRowsInSQL: 10, // Should keep original limit
		},
		{
			name:         "Add LIMIT to JOIN query",
			query:        "SELECT u.*, o.* FROM users u JOIN orders o ON u.id = o.user_id",
			expectLimit:  true,
			maxRowsInSQL: 100,
		},
		{
			name:         "Add LIMIT to UNION",
			query:        "SELECT * FROM users UNION SELECT * FROM customers",
			expectLimit:  true,
			maxRowsInSQL: 100,
		},
		{
			name:         "Override excessive LIMIT",
			query:        "SELECT * FROM users LIMIT 9999999",
			expectLimit:  true,
			maxRowsInSQL: 100, // excessive limit should be overridden to 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := modifier.AddRowLimit(tt.query)
			if err != nil {
				t.Errorf("AddRowLimit() error = %v", err)
				return
			}

			resultUpper := strings.ToUpper(result)
			if tt.expectLimit && !strings.Contains(resultUpper, "LIMIT") {
				t.Errorf("Query should contain LIMIT: %v", result)
			}

			// Verify the numeric LIMIT value equals the expected cap.
			if tt.expectLimit {
				re := regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)`)
				m := re.FindStringSubmatch(result)
				if m == nil {
					t.Errorf("Could not parse LIMIT value from result: %s", result)
				} else {
					got, _ := strconv.Atoi(m[1])
					if got != tt.maxRowsInSQL {
						t.Errorf("LIMIT value = %d, want %d (query: %s)", got, tt.maxRowsInSQL, result)
					}
				}
			}

			t.Logf("Original: %s", tt.query)
			t.Logf("Modified: %s", result)
		})
	}
}

func TestQueryModifier_DefaultLimit(t *testing.T) {
	modifier := NewQueryModifier(0) // Should default to 1000

	query := "SELECT * FROM users"
	result, _ := modifier.AddRowLimit(query)

	if !strings.Contains(strings.ToUpper(result), "LIMIT") {
		t.Error("Should add default LIMIT when maxRowLimit is 0")
	}
}

func TestDefaultPIIPatterns(t *testing.T) {
	patterns := DefaultPIIPatterns()

	if len(patterns) == 0 {
		t.Error("DefaultPIIPatterns should return at least one pattern")
	}

	// Check that default patterns include common PII types
	expectedPatterns := []string{"credit_card", "email", "turkish_id"}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultPIIPatterns missing expected pattern: %s", expected)
		}
	}
}

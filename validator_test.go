package main

import (
	"strings"
	"testing"
)

func TestValidateCSP(t *testing.T) {
	tests := []struct {
		name           string
		csp            string
		expectValid    bool
		expectWarnings int
		expectError    bool
	}{
		{
			name:           "empty CSP",
			csp:            "",
			expectValid:    false,
			expectWarnings: 1,
			expectError:    true,
		},
		{
			name:           "valid minimal CSP",
			csp:            "default-src 'self'",
			expectValid:    true,
			expectWarnings: 0,
		},
		{
			name:           "unsafe-inline with hashes",
			csp:            "script-src 'self' 'unsafe-inline' 'sha256-abc123'",
			expectValid:    true,
			expectWarnings: 2, // unsafe-inline+hash + missing default-src
		},
		{
			name:           "unsafe-eval warning",
			csp:            "script-src 'self' 'unsafe-eval'",
			expectValid:    true,
			expectWarnings: 2, // missing default-src + unsafe-eval
		},
		{
			name:           "missing default-src",
			csp:            "script-src 'self'",
			expectValid:    true,
			expectWarnings: 1,
		},
		{
			name:           "wildcard source",
			csp:            "default-src *",
			expectValid:    true,
			expectWarnings: 1,
		},
		{
			name:           "deprecated directive",
			csp:            "default-src 'self'; block-all-mixed-content",
			expectValid:    true,
			expectWarnings: 1,
		},
		{
			name:           "style-src-attr without style-src",
			csp:            "default-src 'self'; style-src-attr 'unsafe-hashes'",
			expectValid:    true,
			expectWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCSP(tt.csp)

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
				for _, w := range result.Warnings {
					t.Logf("  Warning: %s", w.Message)
				}
			}

			hasError := false
			for _, w := range result.Warnings {
				if w.Severity == "error" {
					hasError = true
					break
				}
			}

			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got %v", tt.expectError, hasError)
			}
		})
	}
}

func TestCheckUnsafeInlineWithHashes(t *testing.T) {
	directives := map[string]string{
		"script-src": "'self' 'unsafe-inline' 'sha256-abc123'",
	}

	result := ValidationResult{Valid: true, Warnings: []ValidationWarning{}}
	checkUnsafeInlineWithHashes(&result, directives)

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if len(result.Warnings) > 0 && !strings.Contains(result.Warnings[0].Message, "unsafe-inline") {
		t.Errorf("Expected warning about unsafe-inline, got: %s", result.Warnings[0].Message)
	}
}

func TestCheckUnsafeEval(t *testing.T) {
	directives := map[string]string{
		"script-src": "'self' 'unsafe-eval'",
	}

	result := ValidationResult{Valid: true, Warnings: []ValidationWarning{}}
	checkUnsafeEval(&result, directives)

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if len(result.Warnings) > 0 && !strings.Contains(result.Warnings[0].Message, "unsafe-eval") {
		t.Errorf("Expected warning about unsafe-eval, got: %s", result.Warnings[0].Message)
	}
}

func TestCheckMissingDefaultSrc(t *testing.T) {
	directives := map[string]string{
		"script-src": "'self'",
	}

	result := ValidationResult{Valid: true, Warnings: []ValidationWarning{}}
	checkMissingDefaultSrc(&result, directives)

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if len(result.Warnings) > 0 && !strings.Contains(result.Warnings[0].Message, "default-src") {
		t.Errorf("Expected warning about default-src, got: %s", result.Warnings[0].Message)
	}
}

func TestCheckOverlyPermissive(t *testing.T) {
	tests := []struct {
		name       string
		directives map[string]string
		expectWarn bool
	}{
		{
			name:       "wildcard in default-src",
			directives: map[string]string{"default-src": "*"},
			expectWarn: true,
		},
		{
			name:       "https wildcard is ok",
			directives: map[string]string{"default-src": "https://*"},
			expectWarn: false,
		},
		{
			name:       "data URI in script-src",
			directives: map[string]string{"script-src": "'self' data:"},
			expectWarn: true,
		},
		{
			name:       "specific domains are ok",
			directives: map[string]string{"script-src": "'self' https://example.com"},
			expectWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidationResult{Valid: true, Warnings: []ValidationWarning{}}
			checkOverlyPermissive(&result, tt.directives)

			hasWarning := len(result.Warnings) > 0
			if hasWarning != tt.expectWarn {
				t.Errorf("Expected warning=%v, got %v (warnings: %d)", tt.expectWarn, hasWarning, len(result.Warnings))
			}
		})
	}
}

func TestCheckDeprecatedDirectives(t *testing.T) {
	directives := map[string]string{
		"default-src":             "'self'",
		"block-all-mixed-content": "",
	}

	result := ValidationResult{Valid: true, Warnings: []ValidationWarning{}}
	checkDeprecatedDirectives(&result, directives)

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning about deprecated directive, got %d", len(result.Warnings))
	}

	if len(result.Warnings) > 0 && !strings.Contains(result.Warnings[0].Message, "deprecated") {
		t.Errorf("Expected warning about deprecated directive, got: %s", result.Warnings[0].Message)
	}
}

func TestCheckConflictingDirectives(t *testing.T) {
	directives := map[string]string{
		"style-src-attr": "'unsafe-hashes'",
	}

	result := ValidationResult{Valid: true, Warnings: []ValidationWarning{}}
	checkConflictingDirectives(&result, directives)

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning about missing style-src, got %d", len(result.Warnings))
	}
}

func TestValidateCSPWithMultipleIssues(t *testing.T) {
	csp := "script-src 'self' 'unsafe-inline' 'unsafe-eval' 'sha256-test'; default-src *; block-all-mixed-content"

	result := ValidateCSP(csp)

	if !result.Valid {
		t.Error("Expected CSP to be syntactically valid despite warnings")
	}

	// Should have warnings for: unsafe-inline+hash, unsafe-eval, wildcard, deprecated directive
	if len(result.Warnings) < 3 {
		t.Errorf("Expected at least 3 warnings, got %d", len(result.Warnings))
		for _, w := range result.Warnings {
			t.Logf("  %s: %s", w.Severity, w.Message)
		}
	}
}

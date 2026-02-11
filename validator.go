package main

import (
	"fmt"
	"strings"
)

// ValidationWarning represents a CSP validation warning
type ValidationWarning struct {
	Severity string // "warning" or "error"
	Message  string
	Fix      string // suggested fix
}

// ValidationResult contains the results of CSP validation
type ValidationResult struct {
	Valid    bool
	Warnings []ValidationWarning
}

// ValidateCSP validates a CSP header and returns warnings about misconfigurations
func ValidateCSP(cspHeader string) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Warnings: []ValidationWarning{},
	}

	directives := parseCSPDirectives(cspHeader)

	// Check for empty or invalid CSP
	if cspHeader == "" {
		result.Valid = false
		result.Warnings = append(result.Warnings, ValidationWarning{
			Severity: "error",
			Message:  "CSP header is empty",
			Fix:      "Provide a valid CSP header string",
		})
		return result
	}

	// Check for 'unsafe-inline' with hashes
	checkUnsafeInlineWithHashes(&result, directives)

	// Check for 'unsafe-eval'
	checkUnsafeEval(&result, directives)

	// Check for missing default-src
	checkMissingDefaultSrc(&result, directives)

	// Check for overly permissive policies
	checkOverlyPermissive(&result, directives)

	// Check for deprecated directives
	checkDeprecatedDirectives(&result, directives)

	// Check for conflicting directives
	checkConflictingDirectives(&result, directives)

	return result
}

// checkUnsafeInlineWithHashes warns if unsafe-inline is used with hashes
func checkUnsafeInlineWithHashes(result *ValidationResult, directives map[string]string) {
	directivesToCheck := []string{"script-src", "style-src"}

	for _, directive := range directivesToCheck {
		if value, exists := directives[directive]; exists {
			hasUnsafeInline := strings.Contains(value, "'unsafe-inline'")
			hasHashes := strings.Contains(value, "'sha256-") ||
				strings.Contains(value, "'sha384-") ||
				strings.Contains(value, "'sha512-")

			if hasUnsafeInline && hasHashes {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Severity: "warning",
					Message:  fmt.Sprintf("%s contains both 'unsafe-inline' and hash values", directive),
					Fix:      fmt.Sprintf("Remove 'unsafe-inline' from %s - hashes are ignored when 'unsafe-inline' is present", directive),
				})
			}
		}
	}
}

// checkUnsafeEval warns about usage of unsafe-eval
func checkUnsafeEval(result *ValidationResult, directives map[string]string) {
	if scriptSrc, exists := directives["script-src"]; exists {
		if strings.Contains(scriptSrc, "'unsafe-eval'") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Severity: "warning",
				Message:  "script-src contains 'unsafe-eval' which allows dangerous eval() usage",
				Fix:      "Remove 'unsafe-eval' if possible and refactor code to avoid eval(), Function(), setTimeout(string), etc.",
			})
		}
	}
}

// checkMissingDefaultSrc warns if default-src is missing
func checkMissingDefaultSrc(result *ValidationResult, directives map[string]string) {
	if _, exists := directives["default-src"]; !exists {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Severity: "warning",
			Message:  "Missing 'default-src' directive",
			Fix:      "Add 'default-src' as a fallback for other directives (recommended: 'default-src 'self'')",
		})
	}
}

// checkOverlyPermissive warns about overly permissive policies
func checkOverlyPermissive(result *ValidationResult, directives map[string]string) {
	directivesToCheck := map[string]string{
		"default-src": "default-src",
		"script-src":  "script-src",
		"style-src":   "style-src",
		"img-src":     "img-src",
		"connect-src": "connect-src",
	}

	for directive, name := range directivesToCheck {
		if value, exists := directives[directive]; exists {
			// Check for wildcard
			if strings.Contains(value, "*") && !strings.Contains(value, "https://*") {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Severity: "warning",
					Message:  fmt.Sprintf("%s contains wildcard '*' which allows resources from any origin", name),
					Fix:      fmt.Sprintf("Restrict %s to specific domains or use 'self'", name),
				})
			}

			// Check for data: URIs in script-src
			if directive == "script-src" && strings.Contains(value, "data:") {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Severity: "warning",
					Message:  "script-src allows 'data:' URIs which can be exploited",
					Fix:      "Remove 'data:' from script-src if not absolutely necessary",
				})
			}
		}
	}
}

// checkDeprecatedDirectives warns about deprecated directives
func checkDeprecatedDirectives(result *ValidationResult, directives map[string]string) {
	deprecated := map[string]string{
		"block-all-mixed-content": "Use 'upgrade-insecure-requests' instead, or handle via HTTPS",
		"plugin-types":            "Deprecated - plugins are no longer supported in modern browsers",
		"referrer":                "Use the Referrer-Policy header instead",
	}

	for directive, suggestion := range deprecated {
		if _, exists := directives[directive]; exists {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Severity: "warning",
				Message:  fmt.Sprintf("'%s' is deprecated", directive),
				Fix:      suggestion,
			})
		}
	}
}

// checkConflictingDirectives checks for conflicting or redundant directives
func checkConflictingDirectives(result *ValidationResult, directives map[string]string) {
	// Check if style-src-attr exists without style-src
	if _, hasStyleSrcAttr := directives["style-src-attr"]; hasStyleSrcAttr {
		if _, hasStyleSrc := directives["style-src"]; !hasStyleSrc {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Severity: "warning",
				Message:  "'style-src-attr' is defined but 'style-src' is not",
				Fix:      "Consider adding 'style-src' as it acts as fallback for 'style-src-attr'",
			})
		}
	}

	// Check if script-src-attr exists without script-src
	if _, hasScriptSrcAttr := directives["script-src-attr"]; hasScriptSrcAttr {
		if _, hasScriptSrc := directives["script-src"]; !hasScriptSrc {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Severity: "warning",
				Message:  "'script-src-attr' is defined but 'script-src' is not",
				Fix:      "Consider adding 'script-src' as it acts as fallback for 'script-src-attr'",
			})
		}
	}
}

// PrintValidationResult prints validation results in a human-readable format
func PrintValidationResult(result ValidationResult, verbose bool) {
	if result.Valid && len(result.Warnings) == 0 {
		fmt.Println("✓ CSP validation passed with no warnings")
		return
	}

	if !result.Valid {
		fmt.Println("✗ CSP validation failed")
	} else {
		fmt.Printf("⚠ CSP validation passed with %d warning(s)\n", len(result.Warnings))
	}

	fmt.Println()

	for i, warning := range result.Warnings {
		symbol := "⚠"
		if warning.Severity == "error" {
			symbol = "✗"
		}

		fmt.Printf("%s %s\n", symbol, warning.Message)
		if verbose && warning.Fix != "" {
			fmt.Printf("  Fix: %s\n", warning.Fix)
		}
		if i < len(result.Warnings)-1 {
			fmt.Println()
		}
	}
}

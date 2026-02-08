package main

import (
	"fmt"
	"strings"
)

// UpdateCSP updates a CSP header string by adding script and style hashes to the appropriate directives
func UpdateCSP(cspHeader string, scriptHashes []string, styleTagHashes []string, styleAttrHashes []string, hasEventHandlers bool) (string, error) {
	// Parse CSP header into directives
	directives := parseCSPDirectives(cspHeader)

	// Update script-src directive
	if len(scriptHashes) > 0 {
		scriptSrc, exists := directives["script-src"]
		if exists {
			// Append hashes to existing directive
			directives["script-src"] = scriptSrc + " " + strings.Join(scriptHashes, " ")
		} else {
			// Create new directive with hashes
			directives["script-src"] = strings.Join(scriptHashes, " ")
		}

		// Add 'unsafe-hashes' if event handlers were found and it's not already present
		if hasEventHandlers && !strings.Contains(directives["script-src"], "'unsafe-hashes'") {
			directives["script-src"] = directives["script-src"] + " 'unsafe-hashes'"
		}
	}

	// Update style-src directive for <style> tags
	if len(styleTagHashes) > 0 {
		styleSrc, exists := directives["style-src"]
		if exists {
			// Append hashes to existing directive
			directives["style-src"] = styleSrc + " " + strings.Join(styleTagHashes, " ")
		} else {
			// Create new directive with hashes
			directives["style-src"] = strings.Join(styleTagHashes, " ")
		}
	}

	// Update style-src-attr or style-src directive for style attributes
	if len(styleAttrHashes) > 0 {
		directiveName := "style-src"
		if _, exists := directives["style-src-attr"]; exists {
			directiveName = "style-src-attr"
		}

		currentValue, exists := directives[directiveName]
		if exists {
			directives[directiveName] = currentValue + " " + strings.Join(styleAttrHashes, " ")
		} else {
			directives[directiveName] = strings.Join(styleAttrHashes, " ")
		}

		// Add 'unsafe-hashes' if it's not already present (required for style attributes)
		if !strings.Contains(directives[directiveName], "'unsafe-hashes'") {
			directives[directiveName] = directives[directiveName] + " 'unsafe-hashes'"
		}
	}

	// Reconstruct CSP header
	return reconstructCSP(directives), nil
}

// parseCSPDirectives parses a CSP header string into a map of directives
func parseCSPDirectives(cspHeader string) map[string]string {
	directives := make(map[string]string)

	// Split by semicolon to get individual directives
	parts := strings.Split(cspHeader, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split directive name from values
		spaceIdx := strings.IndexAny(part, " \t")
		if spaceIdx == -1 {
			// Directive with no values
			directives[part] = ""
		} else {
			directiveName := part[:spaceIdx]
			directiveValue := strings.TrimSpace(part[spaceIdx+1:])
			directives[directiveName] = directiveValue
		}
	}

	return directives
}

// reconstructCSP rebuilds a CSP header string from a map of directives
func reconstructCSP(directives map[string]string) string {
	var parts []string

	// Define a preferred order for common directives (optional, but makes output cleaner)
	orderedDirectives := []string{
		"default-src",
		"script-src",
		"style-src",
		"img-src",
		"font-src",
		"connect-src",
		"frame-src",
		"frame-ancestors",
		"object-src",
		"base-uri",
		"form-action",
	}

	// Add ordered directives first
	addedDirectives := make(map[string]bool)
	for _, name := range orderedDirectives {
		if value, exists := directives[name]; exists {
			if value == "" {
				parts = append(parts, name)
			} else {
				parts = append(parts, fmt.Sprintf("%s %s", name, value))
			}
			addedDirectives[name] = true
		}
	}

	// Add remaining directives
	for name, value := range directives {
		if !addedDirectives[name] {
			if value == "" {
				parts = append(parts, name)
			} else {
				parts = append(parts, fmt.Sprintf("%s %s", name, value))
			}
		}
	}

	return strings.Join(parts, "; ")
}

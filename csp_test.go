package main

import (
	"strings"
	"testing"
)

func TestUpdateCSP(t *testing.T) {
	result, err := UpdateCSP("default-src 'self'", []string{"'sha256-test'"}, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "script-src") || !strings.Contains(result, "'sha256-test'") {
		t.Errorf("Expected CSP to contain script-src with hash, got: %s", result)
	}
}

func TestParseCSPDirectives(t *testing.T) {
	result := parseCSPDirectives("default-src 'self'; script-src 'unsafe-inline'")
	if result["default-src"] != "'self'" {
		t.Errorf("Expected default-src='self', got: %s", result["default-src"])
	}
	if result["script-src"] != "'unsafe-inline'" {
		t.Errorf("Expected script-src='unsafe-inline', got: %s", result["script-src"])
	}
}

func TestReconstructCSP(t *testing.T) {
	directives := map[string]string{
		"default-src": "'self'",
		"script-src":  "'unsafe-inline'",
	}
	result := reconstructCSP(directives)
	if !strings.Contains(result, "default-src 'self'") {
		t.Errorf("Expected CSP to contain default-src 'self', got: %s", result)
	}
}

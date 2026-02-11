package main

import (
	"strings"
	"testing"
)

func TestApplyCSPModifications(t *testing.T) {
	tests := []struct {
		name              string
		initialCSP        string
		modifications     []CSPModification
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:       "add single value",
			initialCSP: "script-src 'self'",
			modifications: []CSPModification{
				{Action: "add", Directive: "script-src", Value: "https://example.com"},
			},
			expectContains: []string{"'self'", "https://example.com"},
		},
		{
			name:       "remove single value",
			initialCSP: "script-src 'self' https://example.com",
			modifications: []CSPModification{
				{Action: "remove", Directive: "script-src", Value: "'self'"},
			},
			expectContains:    []string{"https://example.com"},
			expectNotContains: []string{"'self'"},
		},
		{
			name:       "add then remove",
			initialCSP: "script-src 'self'",
			modifications: []CSPModification{
				{Action: "add", Directive: "script-src", Value: "https://example.com"},
				{Action: "remove", Directive: "script-src", Value: "'self'"},
			},
			expectContains:    []string{"https://example.com"},
			expectNotContains: []string{"'self'"},
		},
		{
			name:       "remove non-existent value",
			initialCSP: "script-src 'self'",
			modifications: []CSPModification{
				{Action: "remove", Directive: "script-src", Value: "https://example.com"},
			},
			expectContains: []string{"'self'"},
		},
		{
			name:       "add duplicate",
			initialCSP: "script-src 'self'",
			modifications: []CSPModification{
				{Action: "add", Directive: "script-src", Value: "'self'"},
			},
			expectContains: []string{"'self'"},
		},
		{
			name:       "multiple directives",
			initialCSP: "default-src 'self'; script-src 'self'; img-src 'self'",
			modifications: []CSPModification{
				{Action: "add", Directive: "script-src", Value: "https://example.com"},
				{Action: "remove", Directive: "img-src", Value: "'self'"},
			},
			expectContains: []string{"https://example.com"},
		},
		{
			name:       "remove all values",
			initialCSP: "script-src 'self'",
			modifications: []CSPModification{
				{Action: "remove", Directive: "script-src", Value: "'self'"},
			},
			expectNotContains: []string{"script-src"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyCSPModifications(tt.initialCSP, tt.modifications)

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %s", expected, result)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result NOT to contain %q, got: %s", notExpected, result)
				}
			}
		})
	}
}

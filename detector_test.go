package main

import (
	"strings"
	"testing"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL",
			url:      "https://example.com/script.js",
			expected: "https://example.com",
		},
		{
			name:     "HTTP URL",
			url:      "http://example.com/style.css",
			expected: "http://example.com",
		},
		{
			name:     "Protocol-relative URL",
			url:      "//cdn.example.com/file.js",
			expected: "https://cdn.example.com",
		},
		{
			name:     "Relative URL",
			url:      "/assets/script.js",
			expected: "",
		},
		{
			name:     "Data URL",
			url:      "data:image/png;base64,iVBORw0KG",
			expected: "",
		},
		{
			name:     "URL with port",
			url:      "https://example.com:8080/api",
			expected: "https://example.com:8080",
		},
		{
			name:     "URL with path and query",
			url:      "https://cdn.example.com/v1/file.js?version=1.2.3",
			expected: "https://cdn.example.com",
		},
		{
			name:     "Invalid URL",
			url:      "not a url",
			expected: "",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDomain(tt.url)
			if result != tt.expected {
				t.Errorf("ExtractDomain(%q) = %q, expected %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestGetUniqueDomains(t *testing.T) {
	resources := &ExternalResources{
		Scripts: []ExternalResource{
			{Type: "script", URL: "https://example.com/a.js", Domain: "https://example.com"},
			{Type: "script", URL: "https://example.com/b.js", Domain: "https://example.com"},
			{Type: "script", URL: "https://cdn.example.com/c.js", Domain: "https://cdn.example.com"},
		},
		Stylesheets: []ExternalResource{
			{Type: "stylesheet", URL: "https://fonts.googleapis.com/style.css", Domain: "https://fonts.googleapis.com"},
		},
		Images: []ExternalResource{
			{Type: "image", URL: "https://example.com/img.png", Domain: "https://example.com"},
		},
	}

	domains := resources.GetUniqueDomains()

	expected := []string{"https://cdn.example.com", "https://example.com", "https://fonts.googleapis.com"}
	if len(domains) != len(expected) {
		t.Errorf("GetUniqueDomains() returned %d domains, expected %d", len(domains), len(expected))
	}

	for i, domain := range domains {
		if domain != expected[i] {
			t.Errorf("GetUniqueDomains()[%d] = %q, expected %q", i, domain, expected[i])
		}
	}
}

func TestGetDomainsByType(t *testing.T) {
	resources := &ExternalResources{
		Scripts: []ExternalResource{
			{Type: "script", URL: "https://example.com/a.js", Domain: "https://example.com"},
			{Type: "script", URL: "https://cdn.example.com/b.js", Domain: "https://cdn.example.com"},
		},
		Stylesheets: []ExternalResource{
			{Type: "stylesheet", URL: "https://fonts.googleapis.com/style.css", Domain: "https://fonts.googleapis.com"},
		},
	}

	scriptDomains := resources.GetDomainsByType("script")
	expectedScripts := []string{"https://cdn.example.com", "https://example.com"}
	if len(scriptDomains) != len(expectedScripts) {
		t.Errorf("GetDomainsByType('script') returned %d domains, expected %d", len(scriptDomains), len(expectedScripts))
	}

	styleDomains := resources.GetDomainsByType("stylesheet")
	expectedStyles := []string{"https://fonts.googleapis.com"}
	if len(styleDomains) != len(expectedStyles) {
		t.Errorf("GetDomainsByType('stylesheet') returned %d domains, expected %d", len(styleDomains), len(expectedStyles))
	}
}

func TestAddExternalResourcesToCSP(t *testing.T) {
	resources := &ExternalResources{
		Scripts: []ExternalResource{
			{Type: "script", URL: "https://cdn.example.com/script.js", Domain: "https://cdn.example.com"},
		},
		Stylesheets: []ExternalResource{
			{Type: "stylesheet", URL: "https://fonts.googleapis.com/style.css", Domain: "https://fonts.googleapis.com"},
		},
		Images: []ExternalResource{
			{Type: "image", URL: "https://images.example.com/pic.jpg", Domain: "https://images.example.com"},
		},
	}

	csp := "default-src 'self'; script-src 'self'; style-src 'self'"
	updatedCSP := AddExternalResourcesToCSP(csp, resources)

	// Check if domains are added
	if !contains(updatedCSP, "https://cdn.example.com") {
		t.Error("Updated CSP should contain https://cdn.example.com for scripts")
	}
	if !contains(updatedCSP, "https://fonts.googleapis.com") {
		t.Error("Updated CSP should contain https://fonts.googleapis.com for styles")
	}
	if !contains(updatedCSP, "https://images.example.com") {
		t.Error("Updated CSP should contain https://images.example.com for images")
	}
}

func TestAppendUniqueDomainsToString(t *testing.T) {
	existing := "'self' https://example.com"
	newDomains := []string{"https://cdn.example.com", "https://example.com"}

	result := appendUniqueDomainsToString(existing, newDomains)

	// Check that all values are present without duplicates
	if !strings.Contains(result, "'self'") {
		t.Error("Result should contain 'self'")
	}
	if !strings.Contains(result, "https://example.com") {
		t.Error("Result should contain https://example.com")
	}
	if !strings.Contains(result, "https://cdn.example.com") {
		t.Error("Result should contain https://cdn.example.com")
	}

	// Count occurrences of https://example.com (should be 1)
	count := strings.Count(result, "https://example.com")
	if count != 1 {
		t.Errorf("https://example.com should appear exactly once, appeared %d times", count)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

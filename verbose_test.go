package main

import (
	"strings"
	"testing"
)

func TestNewVerboseOutput(t *testing.T) {
	vo := NewVerboseOutput(true)
	if vo == nil {
		t.Fatal("NewVerboseOutput returned nil")
	}
	if !vo.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if len(vo.Hashes) != 0 {
		t.Error("Expected Hashes to be empty initially")
	}
}

func TestAddHash(t *testing.T) {
	vo := NewVerboseOutput(true)
	hash := "'sha256-abc123'"
	contentType := "script"
	sourceFile := "test.html"
	content := "console.log('test');"

	vo.AddHash(hash, contentType, sourceFile, content)

	if len(vo.Hashes) != 1 {
		t.Fatalf("Expected 1 hash, got %d", len(vo.Hashes))
	}

	hi := vo.Hashes[0]
	if hi.Hash != hash {
		t.Errorf("Expected hash %q, got %q", hash, hi.Hash)
	}
	if hi.ContentType != contentType {
		t.Errorf("Expected contentType %q, got %q", contentType, hi.ContentType)
	}
	if hi.SourceFile != sourceFile {
		t.Errorf("Expected sourceFile %q, got %q", sourceFile, hi.SourceFile)
	}
	if hi.Content != content {
		t.Errorf("Expected content %q, got %q", content, hi.Content)
	}
}

func TestAddHashDisabled(t *testing.T) {
	vo := NewVerboseOutput(false)
	vo.AddHash("'sha256-abc123'", "script", "test.html", "console.log('test');")

	if len(vo.Hashes) != 0 {
		t.Error("Expected no hashes to be added when verbose is disabled")
	}
}

func TestCreateSnippet(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxLen   int
		expected string
	}{
		{
			name:     "short content",
			content:  "hello",
			maxLen:   60,
			expected: "hello",
		},
		{
			name:     "long content",
			content:  "This is a very long piece of content that should be truncated to fit within the maximum length",
			maxLen:   20,
			expected: "This is a very long ...",
		},
		{
			name:     "content with whitespace",
			content:  "  hello   world  \n  test  ",
			maxLen:   60,
			expected: "hello world test",
		},
		{
			name:     "multiline content",
			content:  "line1\nline2\nline3",
			maxLen:   60,
			expected: "line1 line2 line3",
		},
		{
			name:     "exact length",
			content:  "12345678901234567890",
			maxLen:   20,
			expected: "12345678901234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createSnippet(tt.content, tt.maxLen)
			if result != tt.expected {
				t.Errorf("createSnippet(%q, %d) = %q, expected %q", tt.content, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestFormatContentType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"script", "Inline Scripts"},
		{"style-tag", "Style Tags"},
		{"style-attr", "Style Attributes"},
		{"event-handler", "Event Handlers"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatContentType(tt.input)
			if result != tt.expected {
				t.Errorf("formatContentType(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHashInfoSnippet(t *testing.T) {
	vo := NewVerboseOutput(true)
	longContent := strings.Repeat("a", 100)
	vo.AddHash("'sha256-test'", "script", "test.html", longContent)

	if len(vo.Hashes) != 1 {
		t.Fatal("Expected 1 hash")
	}

	snippet := vo.Hashes[0].Snippet
	if len(snippet) > 63 { // 60 + "..."
		t.Errorf("Snippet is too long: %d characters", len(snippet))
	}
	if !strings.HasSuffix(snippet, "...") {
		t.Error("Long snippet should end with '...'")
	}
}

func TestSetExternalResources(t *testing.T) {
	vo := NewVerboseOutput(true)
	resources := &ExternalResources{
		Scripts: []ExternalResource{
			{Type: "script", URL: "https://example.com/script.js", Domain: "https://example.com"},
		},
	}

	vo.SetExternalResources(resources)

	if vo.ExternalResources == nil {
		t.Error("Expected ExternalResources to be set")
	}
	if len(vo.ExternalResources.Scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(vo.ExternalResources.Scripts))
	}
}

func TestSetExternalResourcesDisabled(t *testing.T) {
	vo := NewVerboseOutput(false)
	resources := &ExternalResources{
		Scripts: []ExternalResource{
			{Type: "script", URL: "https://example.com/script.js", Domain: "https://example.com"},
		},
	}

	vo.SetExternalResources(resources)

	// Should not be set when verbose is disabled
	if vo.ExternalResources != nil {
		t.Error("Expected ExternalResources to remain nil when verbose is disabled")
	}
}

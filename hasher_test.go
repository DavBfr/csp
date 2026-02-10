package main

import "testing"

func TestComputeSHA256Hash(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"empty", "", "'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='"},
		{"simple script", "console.log('hello');", "'sha256-uYeF7eHzVgKpiBg5fikv2NTctmJnxCfX1UhhlrizvNE='"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeSHA256Hash(tt.content)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

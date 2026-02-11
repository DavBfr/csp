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

func TestComputeHash(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		algorithm HashAlgorithm
		expected  string
	}{
		// SHA-256 tests
		{
			name:      "sha256 empty",
			content:   "",
			algorithm: SHA256,
			expected:  "'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='",
		},
		{
			name:      "sha256 simple script",
			content:   "console.log('hello');",
			algorithm: SHA256,
			expected:  "'sha256-uYeF7eHzVgKpiBg5fikv2NTctmJnxCfX1UhhlrizvNE='",
		},
		{
			name:      "sha256 multiline",
			content:   "function test() {\n  return true;\n}",
			algorithm: SHA256,
			expected:  "'sha256-QrFfnElA8UCB9pn+F/t+QOMJRvGlZ/qfbBkaqOOIe78='",
		},
		// SHA-384 tests
		{
			name:      "sha384 empty",
			content:   "",
			algorithm: SHA384,
			expected:  "'sha384-OLBgp1GsljhM2TJ+sbHjaiH9txEUvgdDTAzHv2P24donTt6/529l+9Ua0vFImLlb'",
		},
		{
			name:      "sha384 simple script",
			content:   "console.log('hello');",
			algorithm: SHA384,
			expected:  "'sha384-v393mDht/MNBowq0Z9UpetDvKE6u6EdCihklP1GZs66vL1YCFm1Z4Q4wJtb94rY9'",
		},
		// SHA-512 tests
		{
			name:      "sha512 empty",
			content:   "",
			algorithm: SHA512,
			expected:  "'sha512-z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg=='",
		},
		{
			name:      "sha512 simple script",
			content:   "console.log('hello');",
			algorithm: SHA512,
			expected:  "'sha512-/9wXJrT4fVC90Fxko/AY9VO6E6C1+atlV9CThcRFlmWODDqwRABAr/4EwtzU0W7yJy6PGyvNc9kZV66XEmkrKA=='",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeHash(tt.content, tt.algorithm)
			if result != tt.expected {
				t.Errorf("ComputeHash(%q, %s) = %v, want %v", tt.content, tt.algorithm, result, tt.expected)
			}
		})
	}
}

func TestComputeHashConsistency(t *testing.T) {
	content := "test content for consistency"
	algorithms := []HashAlgorithm{SHA256, SHA384, SHA512}

	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			hash1 := ComputeHash(content, algo)
			hash2 := ComputeHash(content, algo)

			if hash1 != hash2 {
				t.Errorf("Hash function is not consistent for %s: %v != %v", algo, hash1, hash2)
			}
		})
	}
}

func TestComputeHashFormat(t *testing.T) {
	content := "test"
	tests := []struct {
		algorithm HashAlgorithm
		prefix    string
	}{
		{SHA256, "'sha256-"},
		{SHA384, "'sha384-"},
		{SHA512, "'sha512-"},
	}

	for _, tt := range tests {
		t.Run(string(tt.algorithm), func(t *testing.T) {
			hash := ComputeHash(content, tt.algorithm)

			// Check that hash starts with the correct prefix
			if len(hash) < len(tt.prefix) || hash[:len(tt.prefix)] != tt.prefix {
				t.Errorf("Hash format incorrect: expected to start with %s, got %v", tt.prefix, hash)
			}

			// Check that hash ends with '
			if hash[len(hash)-1] != '\'' {
				t.Errorf("Hash format incorrect: expected to end with ', got %v", hash)
			}
		})
	}
}

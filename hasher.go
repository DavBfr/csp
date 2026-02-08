package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// ComputeSHA256Hash computes the SHA-256 hash of content and returns it in CSP format
func ComputeSHA256Hash(content string) string {
	// Compute SHA-256 hash
	hash := sha256.Sum256([]byte(content))

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(hash[:])

	// Return in CSP format with single quotes
	return fmt.Sprintf("'sha256-%s'", encoded)
}

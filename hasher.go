package main

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

// HashAlgorithm represents the supported hash algorithms
type HashAlgorithm string

const (
	SHA256 HashAlgorithm = "sha256"
	SHA384 HashAlgorithm = "sha384"
	SHA512 HashAlgorithm = "sha512"
)

// ComputeHash computes the hash of content using the specified algorithm and returns it in CSP format
func ComputeHash(content string, algo HashAlgorithm) string {
	var encoded string

	switch algo {
	case SHA384:
		hash := sha512.Sum384([]byte(content))
		encoded = base64.StdEncoding.EncodeToString(hash[:])
		return fmt.Sprintf("'sha384-%s'", encoded)
	case SHA512:
		hash := sha512.Sum512([]byte(content))
		encoded = base64.StdEncoding.EncodeToString(hash[:])
		return fmt.Sprintf("'sha512-%s'", encoded)
	default: // SHA256
		hash := sha256.Sum256([]byte(content))
		encoded = base64.StdEncoding.EncodeToString(hash[:])
		return fmt.Sprintf("'sha256-%s'", encoded)
	}
}

// ComputeSHA256Hash computes the SHA-256 hash of content and returns it in CSP format
// Deprecated: Use ComputeHash with SHA256 algorithm instead
func ComputeSHA256Hash(content string) string {
	return ComputeHash(content, SHA256)
}

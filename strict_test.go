package main

import (
	"strings"
	"testing"
)

func TestGetDefaultStrictTemplate(t *testing.T) {
	template := GetDefaultStrictTemplate()

	if len(template.DefaultSrc) == 0 {
		t.Error("Default template should have default-src")
	}
	if template.DefaultSrc[0] != "'none'" {
		t.Errorf("Default template default-src should be 'none', got %q", template.DefaultSrc[0])
	}
	if len(template.ScriptSrc) == 0 || template.ScriptSrc[0] != "'self'" {
		t.Error("Default template should have script-src 'self'")
	}
	if len(template.ImgSrc) == 0 || template.ImgSrc[0] != "'self'" || len(template.ImgSrc) != 1 {
		t.Error("Default template should have img-src 'self' only")
	}
	if len(template.ManifestSrc) == 0 || template.ManifestSrc[0] != "'self'" {
		t.Error("Default template should have manifest-src 'self'")
	}
	if len(template.WorkerSrc) == 0 || template.WorkerSrc[0] != "'self'" {
		t.Error("Default template should have worker-src 'self'")
	}
	if !template.UpgradeInsecure {
		t.Error("Default template should have upgrade-insecure-requests enabled")
	}
}

func TestGenerateStrictCSP(t *testing.T) {
	template := StrictCSPTemplate{
		DefaultSrc:  []string{"'none'"},
		ScriptSrc:   []string{"'self'"},
		StyleSrc:    []string{"'self'"},
		ImgSrc:      []string{"'self'", "data:"},
		ManifestSrc: []string{"'self'"},
		WorkerSrc:   []string{"'self'"},
		ObjectSrc:   []string{"'none'"},
	}

	csp := GenerateStrictCSP(template)

	// Check that all directives are present
	if !strings.Contains(csp, "default-src 'none'") {
		t.Error("Generated CSP should contain default-src 'none'")
	}
	if !strings.Contains(csp, "script-src 'self'") {
		t.Error("Generated CSP should contain script-src 'self'")
	}
	if !strings.Contains(csp, "style-src 'self'") {
		t.Error("Generated CSP should contain style-src 'self'")
	}
	if !strings.Contains(csp, "img-src 'self' data:") {
		t.Error("Generated CSP should contain img-src 'self' data:")
	}
	if !strings.Contains(csp, "manifest-src 'self'") {
		t.Error("Generated CSP should contain manifest-src 'self'")
	}
	if !strings.Contains(csp, "worker-src 'self'") {
		t.Error("Generated CSP should contain worker-src 'self'")
	}
	if !strings.Contains(csp, "object-src 'none'") {
		t.Error("Generated CSP should contain object-src 'none'")
	}

	// Check that directives are separated by semicolons
	if !strings.Contains(csp, "; ") {
		t.Error("Generated CSP should have directives separated by '; '")
	}
}

func TestGenerateStrictCSPWithUpgradeInsecure(t *testing.T) {
	template := StrictCSPTemplate{
		DefaultSrc:      []string{"'self'"},
		UpgradeInsecure: true,
	}

	csp := GenerateStrictCSP(template)

	if !strings.Contains(csp, "upgrade-insecure-requests") {
		t.Error("Generated CSP should contain upgrade-insecure-requests")
	}
}

func TestMergeStrictCSPWithHashes(t *testing.T) {
	strictCSP := "default-src 'none'; script-src 'self'; style-src 'self'"
	scriptHashes := []string{"'sha256-abc123'", "'sha256-def456'"}
	styleTagHashes := []string{"'sha256-xyz789'"}
	styleAttrHashes := []string{"'sha256-attr123'"}

	updatedCSP, err := MergeStrictCSPWithHashes(strictCSP, scriptHashes, styleTagHashes, styleAttrHashes, false)
	if err != nil {
		t.Fatalf("MergeStrictCSPWithHashes() error = %v", err)
	}

	// Check that script hashes are added
	if !strings.Contains(updatedCSP, "'sha256-abc123'") {
		t.Error("Updated CSP should contain script hash 'sha256-abc123'")
	}
	if !strings.Contains(updatedCSP, "'sha256-def456'") {
		t.Error("Updated CSP should contain script hash 'sha256-def456'")
	}

	// Check that style hashes are added
	if !strings.Contains(updatedCSP, "'sha256-xyz789'") {
		t.Error("Updated CSP should contain style tag hash 'sha256-xyz789'")
	}
	if !strings.Contains(updatedCSP, "'sha256-attr123'") {
		t.Error("Updated CSP should contain style attr hash 'sha256-attr123'")
	}
}

func TestMergeStrictCSPWithEventHandlers(t *testing.T) {
	strictCSP := "default-src 'none'; script-src 'self'"
	scriptHashes := []string{"'sha256-abc123'"}

	updatedCSP, err := MergeStrictCSPWithHashes(strictCSP, scriptHashes, nil, nil, true)
	if err != nil {
		t.Fatalf("MergeStrictCSPWithHashes() error = %v", err)
	}

	// Check that 'unsafe-hashes' is added for event handlers
	if !strings.Contains(updatedCSP, "'unsafe-hashes'") {
		t.Error("Updated CSP should contain 'unsafe-hashes' for event handlers")
	}
}

func TestMergeStrictCSPWithStyleAttrs(t *testing.T) {
	strictCSP := "default-src 'none'; style-src 'self'"
	styleAttrHashes := []string{"'sha256-attr123'"}

	updatedCSP, err := MergeStrictCSPWithHashes(strictCSP, nil, nil, styleAttrHashes, false)
	if err != nil {
		t.Fatalf("MergeStrictCSPWithHashes() error = %v", err)
	}

	// Check that 'unsafe-hashes' is added for style attributes
	if !strings.Contains(updatedCSP, "'unsafe-hashes'") {
		t.Error("Updated CSP should contain 'unsafe-hashes' for style attributes")
	}
}

func TestMergeStrictCSPPreservesExisting(t *testing.T) {
	strictCSP := "default-src 'none'; script-src 'self'; object-src 'none'; base-uri 'self'"
	scriptHashes := []string{"'sha256-abc123'"}

	updatedCSP, err := MergeStrictCSPWithHashes(strictCSP, scriptHashes, nil, nil, false)
	if err != nil {
		t.Fatalf("MergeStrictCSPWithHashes() error = %v", err)
	}

	// Check that existing directives are preserved
	if !strings.Contains(updatedCSP, "default-src 'none'") {
		t.Error("Updated CSP should preserve default-src 'none'")
	}
	if !strings.Contains(updatedCSP, "object-src 'none'") {
		t.Error("Updated CSP should preserve object-src 'none'")
	}
	if !strings.Contains(updatedCSP, "base-uri 'self'") {
		t.Error("Updated CSP should preserve base-uri 'self'")
	}
	if !strings.Contains(updatedCSP, "'self'") {
		t.Error("Updated CSP should preserve 'self' in script-src")
	}
}

func TestAddExternalResourcesToStrictCSP(t *testing.T) {
	strictCSP := "default-src 'none'; script-src 'self'; style-src 'self'"
	resources := &ExternalResources{
		Scripts: []ExternalResource{
			{Type: "script", URL: "https://cdn.example.com/script.js", Domain: "https://cdn.example.com"},
		},
	}

	updatedCSP := AddExternalResourcesToStrictCSP(strictCSP, resources)

	// Check that the external domain is added
	if !strings.Contains(updatedCSP, "https://cdn.example.com") {
		t.Error("Updated strict CSP should contain https://cdn.example.com")
	}
}

func TestGenerateStrictCSPWithRequireTrustedTypes(t *testing.T) {
	template := StrictCSPTemplate{
		DefaultSrc:             []string{"'self'"},
		ScriptSrc:              []string{"'self'"},
		RequireTrustedTypesFor: true,
	}

	csp := GenerateStrictCSP(template)

	if !strings.Contains(csp, "require-trusted-types-for 'script'") {
		t.Error("Generated CSP should contain require-trusted-types-for 'script'")
	}
}

func TestGenerateStrictCSPWithoutRequireTrustedTypes(t *testing.T) {
	template := StrictCSPTemplate{
		DefaultSrc:             []string{"'self'"},
		ScriptSrc:              []string{"'self'"},
		RequireTrustedTypesFor: false,
	}

	csp := GenerateStrictCSP(template)

	if strings.Contains(csp, "require-trusted-types-for") {
		t.Error("Generated CSP should not contain require-trusted-types-for when disabled")
	}
}

func TestDefaultTemplateDoesNotRequireTrustedTypes(t *testing.T) {
	template := GetDefaultStrictTemplate()

	if template.RequireTrustedTypesFor {
		t.Error("Default template should not enable require-trusted-types-for by default")
	}
}

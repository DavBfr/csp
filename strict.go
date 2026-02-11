package main

import (
	"strings"
)

// StrictCSPTemplate defines the structure of a strict CSP policy
type StrictCSPTemplate struct {
	DefaultSrc      []string
	ScriptSrc       []string
	StyleSrc        []string
	ImgSrc          []string
	FontSrc         []string
	ConnectSrc      []string
	ManifestSrc     []string
	WorkerSrc       []string
	FrameSrc        []string
	ObjectSrc       []string
	MediaSrc        []string
	BaseURI         []string
	FormAction      []string
	FrameAncestors  []string
	UpgradeInsecure bool
}

// GetDefaultStrictTemplate returns a recommended strict CSP template
func GetDefaultStrictTemplate() StrictCSPTemplate {
	return StrictCSPTemplate{
		DefaultSrc:      []string{"'none'"},
		ScriptSrc:       []string{"'self'"},
		StyleSrc:        []string{"'self'"},
		ImgSrc:          []string{"'self'"},
		FontSrc:         []string{"'self'"},
		ConnectSrc:      []string{"'self'"},
		ManifestSrc:     []string{"'self'"},
		WorkerSrc:       []string{"'self'"},
		FrameSrc:        []string{"'none'"},
		ObjectSrc:       []string{"'none'"},
		MediaSrc:        []string{"'self'"},
		BaseURI:         []string{"'self'"},
		FormAction:      []string{"'self'"},
		FrameAncestors:  []string{"'none'"},
		UpgradeInsecure: true,
	}
}

// GenerateStrictCSP generates a strict CSP from a template
func GenerateStrictCSP(template StrictCSPTemplate) string {
	var parts []string

	// Order matters for readability
	if len(template.DefaultSrc) > 0 {
		parts = append(parts, "default-src "+strings.Join(template.DefaultSrc, " "))
	}

	if len(template.ScriptSrc) > 0 {
		parts = append(parts, "script-src "+strings.Join(template.ScriptSrc, " "))
	}

	if len(template.StyleSrc) > 0 {
		parts = append(parts, "style-src "+strings.Join(template.StyleSrc, " "))
	}

	if len(template.ImgSrc) > 0 {
		parts = append(parts, "img-src "+strings.Join(template.ImgSrc, " "))
	}

	if len(template.FontSrc) > 0 {
		parts = append(parts, "font-src "+strings.Join(template.FontSrc, " "))
	}

	if len(template.ConnectSrc) > 0 {
		parts = append(parts, "connect-src "+strings.Join(template.ConnectSrc, " "))
	}

	if len(template.ManifestSrc) > 0 {
		parts = append(parts, "manifest-src "+strings.Join(template.ManifestSrc, " "))
	}

	if len(template.WorkerSrc) > 0 {
		parts = append(parts, "worker-src "+strings.Join(template.WorkerSrc, " "))
	}

	if len(template.FrameSrc) > 0 {
		parts = append(parts, "frame-src "+strings.Join(template.FrameSrc, " "))
	}

	if len(template.ObjectSrc) > 0 {
		parts = append(parts, "object-src "+strings.Join(template.ObjectSrc, " "))
	}

	if len(template.MediaSrc) > 0 {
		parts = append(parts, "media-src "+strings.Join(template.MediaSrc, " "))
	}

	if len(template.BaseURI) > 0 {
		parts = append(parts, "base-uri "+strings.Join(template.BaseURI, " "))
	}

	if len(template.FormAction) > 0 {
		parts = append(parts, "form-action "+strings.Join(template.FormAction, " "))
	}

	if len(template.FrameAncestors) > 0 {
		parts = append(parts, "frame-ancestors "+strings.Join(template.FrameAncestors, " "))
	}

	if template.UpgradeInsecure {
		parts = append(parts, "upgrade-insecure-requests")
	}

	return strings.Join(parts, "; ")
}

// MergeStrictCSPWithHashes takes a strict CSP and adds hashes to it
func MergeStrictCSPWithHashes(strictCSP string, scriptHashes, styleTagHashes, styleAttrHashes []string, hasEventHandlers bool) (string, error) {
	directives := parseCSPDirectives(strictCSP)

	// Add script hashes to script-src
	if len(scriptHashes) > 0 || hasEventHandlers {
		scriptSrc := directives["script-src"]

		// Add hashes
		if len(scriptHashes) > 0 {
			if scriptSrc != "" {
				scriptSrc = scriptSrc + " " + strings.Join(scriptHashes, " ")
			} else {
				scriptSrc = strings.Join(scriptHashes, " ")
			}
		}

		// Add 'unsafe-hashes' if there are event handlers
		if hasEventHandlers && !strings.Contains(scriptSrc, "'unsafe-hashes'") {
			if scriptSrc != "" {
				scriptSrc = scriptSrc + " 'unsafe-hashes'"
			} else {
				scriptSrc = "'unsafe-hashes'"
			}
		}

		directives["script-src"] = scriptSrc
	}

	// Add style hashes to style-src
	if len(styleTagHashes) > 0 || len(styleAttrHashes) > 0 {
		styleSrc := directives["style-src"]

		// Add style tag hashes
		if len(styleTagHashes) > 0 {
			if styleSrc != "" {
				styleSrc = styleSrc + " " + strings.Join(styleTagHashes, " ")
			} else {
				styleSrc = strings.Join(styleTagHashes, " ")
			}
		}

		// Add style attribute hashes
		if len(styleAttrHashes) > 0 {
			if styleSrc != "" {
				styleSrc = styleSrc + " " + strings.Join(styleAttrHashes, " ")
			} else {
				styleSrc = strings.Join(styleAttrHashes, " ")
			}
		}

		// Add 'unsafe-hashes' if there are style attributes
		if len(styleAttrHashes) > 0 && !strings.Contains(styleSrc, "'unsafe-hashes'") {
			if styleSrc != "" {
				styleSrc = styleSrc + " 'unsafe-hashes'"
			} else {
				styleSrc = "'unsafe-hashes'"
			}
		}

		directives["style-src"] = styleSrc
	}

	return reconstructCSP(directives), nil
}

// AddExternalResourcesToStrictCSP adds external resource domains to a strict CSP
func AddExternalResourcesToStrictCSP(strictCSP string, resources *ExternalResources) string {
	return AddExternalResourcesToCSP(strictCSP, resources)
}

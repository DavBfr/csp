package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command-line flags
	cspFlag := flag.String("csp", "", "Existing CSP header to update with hashes (optional with --generate-strict)")
	hashAlgo := flag.String("hash-algo", "sha256", "Hash algorithm to use: sha256, sha384, or sha512")
	validateOnly := flag.Bool("validate-only", false, "Only validate the CSP without processing HTML files")
	noValidate := flag.Bool("no-validate", false, "Skip CSP validation checks")
	noScripts := flag.Bool("no-scripts", false, "Skip processing inline <script> elements")
	noStyles := flag.Bool("no-styles", false, "Skip processing inline <style> tags")
	noInlineStyles := flag.Bool("no-inline-styles", false, "Skip processing inline style attributes")
	noEventHandlers := flag.Bool("no-event-handlers", false, "Skip processing inline event handlers (onclick, etc.)")
	includeExternal := flag.Bool("include-external", false, "Scan for external resources and add domains to CSP directives")
	generateStrict := flag.Bool("generate-strict", false, "Generate a complete strict CSP from scratch")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: csp [options] file1.html [file2.html ...]\n\n")
		fmt.Fprintf(os.Stderr, "Generate CSP hashes for inline content in HTML files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --no-scripts *.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --hash-algo sha384 index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --validate-only\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --no-event-handlers index.html about.html\n")
		fmt.Fprintf(os.Stderr, "  csp --generate-strict index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --include-external index.html\n")
	}

	flag.Parse()

	// Validate inputs
	if *cspFlag == "" && !*generateStrict {
		fmt.Fprintln(os.Stderr, "Error: --csp flag is required (or use --generate-strict)")
		fmt.Fprintln(os.Stderr, "Usage: csp --csp \"CSP_HEADER\" [options] file1.html file2.html ...")
		os.Exit(1)
	}

	// Validate hash algorithm
	var algorithm HashAlgorithm
	switch *hashAlgo {
	case "sha256":
		algorithm = SHA256
	case "sha384":
		algorithm = SHA384
	case "sha512":
		algorithm = SHA512
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid hash algorithm '%s'. Must be sha256, sha384, or sha512\n", *hashAlgo)
		os.Exit(1)
	}

	// Handle validate-only mode
	if *validateOnly {
		if *cspFlag == "" {
			fmt.Fprintln(os.Stderr, "Error: --csp flag is required for validation")
			os.Exit(1)
		}
		result := ValidateCSP(*cspFlag)
		PrintValidationResult(result, true)
		if !result.Valid {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Validate input CSP before processing (unless disabled or generating strict)
	if !*noValidate && *cspFlag != "" && !*generateStrict {
		result := ValidateCSP(*cspFlag)
		if !result.Valid {
			fmt.Fprintln(os.Stderr, "Input CSP validation failed:")
			PrintValidationResult(result, false)
			fmt.Fprintln(os.Stderr, "\nContinuing anyway...")
		} else if len(result.Warnings) > 0 {
			fmt.Fprintf(os.Stderr, "Input CSP has %d warning(s). Use --validate-only for details.\n\n", len(result.Warnings))
		}
	}

	htmlFiles := flag.Args()
	if len(htmlFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one HTML file is required")
		fmt.Fprintln(os.Stderr, "Usage: csp --csp \"CSP_HEADER\" [options] file1.html file2.html ...")
		os.Exit(1)
	}

	// Initialize or use provided CSP
	var baseCSP string
	if *generateStrict {
		// Generate a strict CSP from the default template
		template := GetDefaultStrictTemplate()
		baseCSP = GenerateStrictCSP(template)
	} else {
		baseCSP = *cspFlag
	}

	// Collect all script and style hashes from all HTML files
	var allScriptHashes []string
	var allStyleTagHashes []string
	var allStyleAttrHashes []string
	hasEventHandlers := false

	// Collect external resources if requested
	var allExternalResources *ExternalResources
	if *includeExternal {
		allExternalResources = &ExternalResources{
			Scripts:     []ExternalResource{},
			Stylesheets: []ExternalResource{},
			Images:      []ExternalResource{},
			Fonts:       []ExternalResource{},
			Frames:      []ExternalResource{},
			Other:       []ExternalResource{},
		}
	}

	for _, filePath := range htmlFiles {
		scripts, styleTags, styleAttrs, hasEvents, err := ExtractInlineContent(filePath, *noScripts, *noStyles, *noInlineStyles, *noEventHandlers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filePath, err)
			os.Exit(1)
		}
		if hasEvents {
			hasEventHandlers = true
		}

		// Extract external resources if requested
		if *includeExternal {
			externalRes, err := ExtractExternalResources(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to extract external resources from %s: %v\n", filePath, err)
			} else {
				// Merge resources
				allExternalResources.Scripts = append(allExternalResources.Scripts, externalRes.Scripts...)
				allExternalResources.Stylesheets = append(allExternalResources.Stylesheets, externalRes.Stylesheets...)
				allExternalResources.Images = append(allExternalResources.Images, externalRes.Images...)
				allExternalResources.Fonts = append(allExternalResources.Fonts, externalRes.Fonts...)
				allExternalResources.Frames = append(allExternalResources.Frames, externalRes.Frames...)
				allExternalResources.Other = append(allExternalResources.Other, externalRes.Other...)
			}
		}

		// Compute hashes for scripts (unless disabled)
		if !*noScripts {
			for _, script := range scripts {
				hash := ComputeHash(script, algorithm)
				allScriptHashes = append(allScriptHashes, hash)
			}
		}

		// Compute hashes for style tags (unless disabled)
		if !*noStyles {
			for _, style := range styleTags {
				hash := ComputeHash(style, algorithm)
				allStyleTagHashes = append(allStyleTagHashes, hash)
			}
		}

		// Compute hashes for style attributes (unless disabled)
		if !*noInlineStyles {
			for _, style := range styleAttrs {
				hash := ComputeHash(style, algorithm)
				allStyleAttrHashes = append(allStyleAttrHashes, hash)
			}
		}
	}

	// Remove duplicate hashes
	allScriptHashes = removeDuplicates(allScriptHashes)
	allStyleTagHashes = removeDuplicates(allStyleTagHashes)
	allStyleAttrHashes = removeDuplicates(allStyleAttrHashes)

	// Update CSP header with hashes
	var updatedCSP string
	var err error
	if *generateStrict {
		// Use strict CSP merge function
		updatedCSP, err = MergeStrictCSPWithHashes(baseCSP, allScriptHashes, allStyleTagHashes, allStyleAttrHashes, hasEventHandlers)
	} else {
		updatedCSP, err = UpdateCSP(baseCSP, allScriptHashes, allStyleTagHashes, allStyleAttrHashes, hasEventHandlers)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating CSP: %v\n", err)
		os.Exit(1)
	}

	// Add external resource domains if requested
	if *includeExternal && allExternalResources != nil {
		updatedCSP = AddExternalResourcesToCSP(updatedCSP, allExternalResources)
	}

	// Validate output CSP (unless disabled)
	if !*noValidate {
		result := ValidateCSP(updatedCSP)
		if len(result.Warnings) > 0 {
			fmt.Fprintf(os.Stderr, "Output CSP has %d warning(s). Use --validate-only to check.\n\n", len(result.Warnings))
		}
	}

	// Output the updated CSP header
	fmt.Println(updatedCSP)
}

// removeDuplicates removes duplicate strings from a slice while preserving order
func removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

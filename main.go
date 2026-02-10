package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command-line flags
	cspFlag := flag.String("csp", "", "Existing CSP header to update with hashes (required)")
	noScripts := flag.Bool("no-scripts", false, "Skip processing inline <script> elements")
	noStyles := flag.Bool("no-styles", false, "Skip processing inline <style> tags")
	noInlineStyles := flag.Bool("no-inline-styles", false, "Skip processing inline style attributes")
	noEventHandlers := flag.Bool("no-event-handlers", false, "Skip processing inline event handlers (onclick, etc.)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: csp --csp \"CSP_HEADER\" [options] file1.html [file2.html ...]\n\n")
		fmt.Fprintf(os.Stderr, "Generate CSP hashes for inline content in HTML files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --no-scripts *.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --no-event-handlers index.html about.html\n")
	}

	flag.Parse()

	// Validate inputs
	if *cspFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: --csp flag is required")
		fmt.Fprintln(os.Stderr, "Usage: csp --csp \"CSP_HEADER\" [options] file1.html file2.html ...")
		os.Exit(1)
	}

	htmlFiles := flag.Args()
	if len(htmlFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one HTML file is required")
		fmt.Fprintln(os.Stderr, "Usage: csp --csp \"CSP_HEADER\" [options] file1.html file2.html ...")
		os.Exit(1)
	}

	// Collect all script and style hashes from all HTML files
	var allScriptHashes []string
	var allStyleTagHashes []string
	var allStyleAttrHashes []string
	hasEventHandlers := false

	for _, filePath := range htmlFiles {
		scripts, styleTags, styleAttrs, hasEvents, err := ExtractInlineContent(filePath, *noScripts, *noStyles, *noInlineStyles, *noEventHandlers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filePath, err)
			os.Exit(1)
		}
		if hasEvents {
			hasEventHandlers = true
		}

		// Compute hashes for scripts (unless disabled)
		if !*noScripts {
			for _, script := range scripts {
				hash := ComputeSHA256Hash(script)
				allScriptHashes = append(allScriptHashes, hash)
			}
		}

		// Compute hashes for style tags (unless disabled)
		if !*noStyles {
			for _, style := range styleTags {
				hash := ComputeSHA256Hash(style)
				allStyleTagHashes = append(allStyleTagHashes, hash)
			}
		}

		// Compute hashes for style attributes (unless disabled)
		if !*noInlineStyles {
			for _, style := range styleAttrs {
				hash := ComputeSHA256Hash(style)
				allStyleAttrHashes = append(allStyleAttrHashes, hash)
			}
		}
	}

	// Remove duplicate hashes
	allScriptHashes = removeDuplicates(allScriptHashes)
	allStyleTagHashes = removeDuplicates(allStyleTagHashes)
	allStyleAttrHashes = removeDuplicates(allStyleAttrHashes)

	// Update CSP header with hashes
	updatedCSP, err := UpdateCSP(*cspFlag, allScriptHashes, allStyleTagHashes, allStyleAttrHashes, hasEventHandlers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating CSP: %v\n", err)
		os.Exit(1)
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

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// CSPModificationList implements flag.Value to collect CSP modifications in order
type CSPModificationList struct {
	modifications []CSPModification
}

func (cml *CSPModificationList) String() string {
	return fmt.Sprintf("%v", cml.modifications)
}

func (cml *CSPModificationList) Set(value string) error {
	// value is in format "directive:value", we extract the directive from the flag name
	cml.modifications = append(cml.modifications, CSPModification{Value: value})
	return nil
}

// DirectiveModification creates a flag type for a specific directive and action
func DirectiveModification(directive, action string) flag.Value {
	return &directiveFlag{directive: directive, action: action, modifications: &[]CSPModification{}}
}

type directiveFlag struct {
	directive     string
	action        string
	modifications *[]CSPModification
}

func (df *directiveFlag) String() string {
	return ""
}

func (df *directiveFlag) Set(value string) error {
	*df.modifications = append(*df.modifications, CSPModification{
		Action:    df.action,
		Directive: df.directive,
		Value:     value,
	})
	return nil
}

func main() {
	// Shared modifications list for all add/remove flags
	var modifications []CSPModification

	// Define command-line flags
	cspFlag := flag.String("csp", "", "Existing CSP header to update with hashes (optional, defaults to --generate-strict)")
	hashAlgo := flag.String("hash-algo", "sha256", "Hash algorithm to use: sha256, sha384, or sha512")
	validateOnly := flag.Bool("validate-only", false, "Only validate the CSP without processing HTML files")
	noValidate := flag.Bool("no-validate", false, "Skip CSP validation checks")
	noScripts := flag.Bool("no-scripts", false, "Skip processing inline <script> elements")
	noStyles := flag.Bool("no-styles", false, "Skip processing inline <style> tags")
	noInlineStyles := flag.Bool("no-inline-styles", false, "Skip processing inline style attributes")
	noEventHandlers := flag.Bool("no-event-handlers", false, "Skip processing inline event handlers (onclick, etc.)")
	includeExternal := flag.Bool("include-external", false, "Scan for external resources and add domains to CSP directives")
	useHeuristics := flag.Bool("heuristics", false, "Use heuristics to infer additional external resources (e.g., fonts loaded by stylesheets)")
	generateStrict := flag.Bool("generate-strict", false, "Generate a complete strict CSP from scratch")
	requireTrustedTypes := flag.Bool("require-trusted-types", false, "Add require-trusted-types-for 'script' directive (requires Trusted Types API support)")
	verbose := flag.Bool("verbose", false, "Show detailed information about hash generation")
	verboseShort := flag.Bool("v", false, "Show detailed information about hash generation (short)")

	// Create shared modifications list for add/remove directives
	addScriptSrc := &directiveFlag{directive: "script-src", action: "add", modifications: &modifications}
	removeScriptSrc := &directiveFlag{directive: "script-src", action: "remove", modifications: &modifications}
	addStyleSrc := &directiveFlag{directive: "style-src", action: "add", modifications: &modifications}
	removeStyleSrc := &directiveFlag{directive: "style-src", action: "remove", modifications: &modifications}
	addImgSrc := &directiveFlag{directive: "img-src", action: "add", modifications: &modifications}
	removeImgSrc := &directiveFlag{directive: "img-src", action: "remove", modifications: &modifications}
	addFontSrc := &directiveFlag{directive: "font-src", action: "add", modifications: &modifications}
	removeFontSrc := &directiveFlag{directive: "font-src", action: "remove", modifications: &modifications}
	addConnectSrc := &directiveFlag{directive: "connect-src", action: "add", modifications: &modifications}
	removeConnectSrc := &directiveFlag{directive: "connect-src", action: "remove", modifications: &modifications}
	addManifestSrc := &directiveFlag{directive: "manifest-src", action: "add", modifications: &modifications}
	removeManifestSrc := &directiveFlag{directive: "manifest-src", action: "remove", modifications: &modifications}
	addWorkerSrc := &directiveFlag{directive: "worker-src", action: "add", modifications: &modifications}
	removeWorkerSrc := &directiveFlag{directive: "worker-src", action: "remove", modifications: &modifications}
	addFrameSrc := &directiveFlag{directive: "frame-src", action: "add", modifications: &modifications}
	removeFrameSrc := &directiveFlag{directive: "frame-src", action: "remove", modifications: &modifications}
	addDefaultSrc := &directiveFlag{directive: "default-src", action: "add", modifications: &modifications}
	removeDefaultSrc := &directiveFlag{directive: "default-src", action: "remove", modifications: &modifications}
	addObjectSrc := &directiveFlag{directive: "object-src", action: "add", modifications: &modifications}
	removeObjectSrc := &directiveFlag{directive: "object-src", action: "remove", modifications: &modifications}
	addMediaSrc := &directiveFlag{directive: "media-src", action: "add", modifications: &modifications}
	removeMediaSrc := &directiveFlag{directive: "media-src", action: "remove", modifications: &modifications}
	addBaseURI := &directiveFlag{directive: "base-uri", action: "add", modifications: &modifications}
	removeBaseURI := &directiveFlag{directive: "base-uri", action: "remove", modifications: &modifications}
	addFormAction := &directiveFlag{directive: "form-action", action: "add", modifications: &modifications}
	removeFormAction := &directiveFlag{directive: "form-action", action: "remove", modifications: &modifications}
	addFrameAncestors := &directiveFlag{directive: "frame-ancestors", action: "add", modifications: &modifications}
	removeFrameAncestors := &directiveFlag{directive: "frame-ancestors", action: "remove", modifications: &modifications}

	// Register add/remove flags
	flag.Var(addScriptSrc, "add-script-src", "Add value to script-src directive (can be repeated, evaluated in order)")
	flag.Var(removeScriptSrc, "remove-script-src", "Remove value from script-src directive (can be repeated, evaluated in order)")
	flag.Var(addStyleSrc, "add-style-src", "Add value to style-src directive (can be repeated, evaluated in order)")
	flag.Var(removeStyleSrc, "remove-style-src", "Remove value from style-src directive (can be repeated, evaluated in order)")
	flag.Var(addImgSrc, "add-img-src", "Add value to img-src directive (can be repeated, evaluated in order)")
	flag.Var(removeImgSrc, "remove-img-src", "Remove value from img-src directive (can be repeated, evaluated in order)")
	flag.Var(addFontSrc, "add-font-src", "Add value to font-src directive (can be repeated, evaluated in order)")
	flag.Var(removeFontSrc, "remove-font-src", "Remove value from font-src directive (can be repeated, evaluated in order)")
	flag.Var(addConnectSrc, "add-connect-src", "Add value to connect-src directive (can be repeated, evaluated in order)")
	flag.Var(removeConnectSrc, "remove-connect-src", "Remove value from connect-src directive (can be repeated, evaluated in order)")
	flag.Var(addManifestSrc, "add-manifest-src", "Add value to manifest-src directive (can be repeated, evaluated in order)")
	flag.Var(removeManifestSrc, "remove-manifest-src", "Remove value from manifest-src directive (can be repeated, evaluated in order)")
	flag.Var(addWorkerSrc, "add-worker-src", "Add value to worker-src directive (can be repeated, evaluated in order)")
	flag.Var(removeWorkerSrc, "remove-worker-src", "Remove value from worker-src directive (can be repeated, evaluated in order)")
	flag.Var(addFrameSrc, "add-frame-src", "Add value to frame-src directive (can be repeated, evaluated in order)")
	flag.Var(removeFrameSrc, "remove-frame-src", "Remove value from frame-src directive (can be repeated, evaluated in order)")
	flag.Var(addDefaultSrc, "add-default-src", "Add value to default-src directive (can be repeated, evaluated in order)")
	flag.Var(removeDefaultSrc, "remove-default-src", "Remove value from default-src directive (can be repeated, evaluated in order)")
	flag.Var(addObjectSrc, "add-object-src", "Add value to object-src directive (can be repeated, evaluated in order)")
	flag.Var(removeObjectSrc, "remove-object-src", "Remove value from object-src directive (can be repeated, evaluated in order)")
	flag.Var(addMediaSrc, "add-media-src", "Add value to media-src directive (can be repeated, evaluated in order)")
	flag.Var(removeMediaSrc, "remove-media-src", "Remove value from media-src directive (can be repeated, evaluated in order)")
	flag.Var(addBaseURI, "add-base-uri", "Add value to base-uri directive (can be repeated, evaluated in order)")
	flag.Var(removeBaseURI, "remove-base-uri", "Remove value from base-uri directive (can be repeated, evaluated in order)")
	flag.Var(addFormAction, "add-form-action", "Add value to form-action directive (can be repeated, evaluated in order)")
	flag.Var(removeFormAction, "remove-form-action", "Remove value from form-action directive (can be repeated, evaluated in order)")
	flag.Var(addFrameAncestors, "add-frame-ancestors", "Add value to frame-ancestors directive (can be repeated, evaluated in order)")
	flag.Var(removeFrameAncestors, "remove-frame-ancestors", "Remove value from frame-ancestors directive (can be repeated, evaluated in order)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: csp [options] file1.html [file2.html ...]\n\n")
		fmt.Fprintf(os.Stderr, "Generate CSP hashes for inline content in HTML files.\n")
		fmt.Fprintf(os.Stderr, "If no CSP is provided, a strict CSP will be generated by default.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  csp index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --no-scripts *.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --hash-algo sha384 index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --validate-only\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --no-event-handlers index.html about.html\n")
		fmt.Fprintf(os.Stderr, "  csp --generate-strict index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" --include-external index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --include-external --heuristics index.html\n")
		fmt.Fprintf(os.Stderr, "  csp --csp \"default-src 'self'\" -v index.html\n")
	}

	flag.Parse()

	// Handle verbose flag (either -v or --verbose)
	verboseEnabled := *verbose || *verboseShort

	// Use safe default: generate strict CSP if neither --csp nor --generate-strict is specified
	if *cspFlag == "" && !*generateStrict {
		if verboseEnabled {
			fmt.Fprintln(os.Stderr, "No CSP provided, using --generate-strict as safe default")
		}
		*generateStrict = true
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
		template.RequireTrustedTypesFor = *requireTrustedTypes
		baseCSP = GenerateStrictCSP(template)
	} else {
		baseCSP = *cspFlag
	}

	// Initialize verbose output
	verboseOut := NewVerboseOutput(verboseEnabled)

	// Collect all script and style hashes from all HTML files
	var allScriptHashes []string
	var allStyleTagHashes []string
	var allStyleAttrHashes []string
	hasEventHandlers := false

	// Track counts for verbose output
	totalScripts := 0
	totalStyleTags := 0
	totalStyleAttrs := 0

	// Collect external resources if requested
	var allExternalResources *ExternalResources
	var allHeuristicResources []HeuristicResource
	if *includeExternal {
		allExternalResources = &ExternalResources{
			Scripts:      []ExternalResource{},
			Stylesheets:  []ExternalResource{},
			Images:       []ExternalResource{},
			Fonts:        []ExternalResource{},
			Frames:       []ExternalResource{},
			Other:        []ExternalResource{},
			UsesDataURLs: make(map[string]bool),
		}
	}

	for i, filePath := range htmlFiles {
		verboseOut.PrintProgress(filePath, i+1, len(htmlFiles))

		scripts, styleTags, styleAttrs, hasEvents, err := ExtractInlineContent(filePath, *noScripts, *noStyles, *noInlineStyles, *noEventHandlers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filePath, err)
			os.Exit(1)
		}
		if hasEvents {
			hasEventHandlers = true
		}

		// Count event handlers for verbose output
		eventHandlerCount := 0
		if hasEvents {
			for _, script := range scripts {
				// Simple heuristic: very short scripts are likely event handlers
				if len(script) < 200 && !strings.Contains(script, "\n") {
					eventHandlerCount++
				}
			}
		}

		verboseOut.PrintFileSummary(filePath, len(scripts), len(styleTags), len(styleAttrs), eventHandlerCount)

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
				// Merge data URL usage flags
				for resourceType, used := range externalRes.UsesDataURLs {
					if used {
						allExternalResources.UsesDataURLs[resourceType] = true
					}
				}
			}
		}

		// Compute hashes for scripts (unless disabled)
		if !*noScripts {
			for _, script := range scripts {
				hash := ComputeHash(script, algorithm)
				allScriptHashes = append(allScriptHashes, hash)
				totalScripts++

				// Determine if this is an event handler
				contentType := "script"
				if len(script) < 200 && !strings.Contains(script, "\n") && hasEvents {
					contentType = "event-handler"
				}
				verboseOut.AddHash(hash, contentType, filePath, script)
			}
		}

		// Compute hashes for style tags (unless disabled)
		if !*noStyles {
			for _, style := range styleTags {
				hash := ComputeHash(style, algorithm)
				allStyleTagHashes = append(allStyleTagHashes, hash)
				totalStyleTags++
				verboseOut.AddHash(hash, "style-tag", filePath, style)
			}
		}

		// Compute hashes for style attributes (unless disabled)
		if !*noInlineStyles {
			for _, style := range styleAttrs {
				hash := ComputeHash(style, algorithm)
				allStyleAttrHashes = append(allStyleAttrHashes, hash)
				totalStyleAttrs++
				verboseOut.AddHash(hash, "style-attr", filePath, style)
			}
		}
	}

	// Apply heuristics if requested
	if *includeExternal && *useHeuristics && allExternalResources != nil {
		// Collect all external resources into a flat list for heuristics
		var allResources []ExternalResource
		allResources = append(allResources, allExternalResources.Scripts...)
		allResources = append(allResources, allExternalResources.Stylesheets...)
		allResources = append(allResources, allExternalResources.Images...)
		allResources = append(allResources, allExternalResources.Fonts...)
		allResources = append(allResources, allExternalResources.Frames...)

		// Apply heuristics
		allHeuristicResources = ApplyHeuristics(allResources)

		// Convert heuristic resources back to external resources and merge
		for _, h := range allHeuristicResources {
			externalRes := ConvertHeuristicToExternalResource(h)
			switch h.Type {
			case "script":
				allExternalResources.Scripts = append(allExternalResources.Scripts, externalRes)
			case "stylesheet":
				allExternalResources.Stylesheets = append(allExternalResources.Stylesheets, externalRes)
			case "image":
				allExternalResources.Images = append(allExternalResources.Images, externalRes)
			case "font":
				allExternalResources.Fonts = append(allExternalResources.Fonts, externalRes)
			case "frame":
				allExternalResources.Frames = append(allExternalResources.Frames, externalRes)
			case "connect":
				allExternalResources.Other = append(allExternalResources.Other, externalRes)
			}
		}
	}

	// Remove duplicate hashes
	allScriptHashes = removeDuplicates(allScriptHashes)
	allStyleTagHashes = removeDuplicates(allStyleTagHashes)
	allStyleAttrHashes = removeDuplicates(allStyleAttrHashes)

	// Print verbose output
	if verboseEnabled {
		verboseOut.PrintHashDetails()

		// Print external resources if they were collected
		if *includeExternal && allExternalResources != nil {
			verboseOut.SetExternalResources(allExternalResources)
			verboseOut.PrintExternalResources()

			// Print heuristic inferences if they were used
			if *useHeuristics && len(allHeuristicResources) > 0 {
				fmt.Fprintln(os.Stderr, "\nInferred Resources (from heuristics):")
				fmt.Fprintln(os.Stderr, "================================================================================")
				for _, h := range allHeuristicResources {
					fmt.Fprintf(os.Stderr, "  [%s] %s\n", strings.ToUpper(h.Confidence), h.URL)
					fmt.Fprintf(os.Stderr, "      Type: %s\n", h.Type)
					fmt.Fprintf(os.Stderr, "      Reason: %s\n", h.Reason)
					fmt.Fprintf(os.Stderr, "      Source: %s (%s)\n", h.SourceURL, h.SourceType)
					fmt.Fprintln(os.Stderr)
				}
				summary := GetHeuristicsSummary(allHeuristicResources)
				fmt.Fprintf(os.Stderr, "Total inferred: %d resources\n", len(allHeuristicResources))
				for key, count := range summary {
					if !strings.HasPrefix(key, "confidence_") {
						fmt.Fprintf(os.Stderr, "  - %s: %d\n", key, count)
					}
				}
				fmt.Fprintf(os.Stderr, "Confidence levels: High=%d, Medium=%d, Low=%d\n\n",
					summary["confidence_high"], summary["confidence_medium"], summary["confidence_low"])
			}
		}

		verboseOut.PrintSummary(totalScripts, totalStyleTags, totalStyleAttrs,
			len(allScriptHashes), len(allStyleTagHashes), len(allStyleAttrHashes))
	}

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

	// Apply any add/remove modifications in order
	if len(modifications) > 0 {
		updatedCSP = ApplyCSPModifications(updatedCSP, modifications)
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

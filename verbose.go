package main

import (
	"fmt"
	"os"
	"strings"
)

// HashInfo stores information about a computed hash
type HashInfo struct {
	Hash        string
	ContentType string // "script", "style-tag", "style-attr", "event-handler"
	SourceFile  string
	Content     string
	Snippet     string // Truncated content for display
}

// VerboseOutput handles displaying detailed information about hash generation
type VerboseOutput struct {
	Enabled           bool
	Hashes            []HashInfo
	ExternalResources *ExternalResources
}

// NewVerboseOutput creates a new VerboseOutput instance
func NewVerboseOutput(enabled bool) *VerboseOutput {
	return &VerboseOutput{
		Enabled:           enabled,
		Hashes:            []HashInfo{},
		ExternalResources: nil,
	}
}

// AddHash records a hash with its metadata
func (vo *VerboseOutput) AddHash(hash, contentType, sourceFile, content string) {
	if !vo.Enabled {
		return
	}

	snippet := createSnippet(content, 60)
	vo.Hashes = append(vo.Hashes, HashInfo{
		Hash:        hash,
		ContentType: contentType,
		SourceFile:  sourceFile,
		Content:     content,
		Snippet:     snippet,
	})
}

// PrintProgress prints processing progress for a file
func (vo *VerboseOutput) PrintProgress(filePath string, fileNum, totalFiles int) {
	if !vo.Enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "[%d/%d] Processing %s\n", fileNum, totalFiles, filePath)
}

// PrintFileSummary prints a summary of what was found in a file
func (vo *VerboseOutput) PrintFileSummary(filePath string, scriptCount, styleTagCount, styleAttrCount, eventHandlerCount int) {
	if !vo.Enabled {
		return
	}

	items := []string{}
	if scriptCount > 0 {
		items = append(items, fmt.Sprintf("%d inline script(s)", scriptCount))
	}
	if styleTagCount > 0 {
		items = append(items, fmt.Sprintf("%d <style> tag(s)", styleTagCount))
	}
	if styleAttrCount > 0 {
		items = append(items, fmt.Sprintf("%d style attribute(s)", styleAttrCount))
	}
	if eventHandlerCount > 0 {
		items = append(items, fmt.Sprintf("%d event handler(s)", eventHandlerCount))
	}

	if len(items) > 0 {
		fmt.Fprintf(os.Stderr, "  Found: %s\n", strings.Join(items, ", "))
	} else {
		fmt.Fprintf(os.Stderr, "  Found: no inline content\n")
	}
}

// PrintHashDetails prints detailed information about all hashes
func (vo *VerboseOutput) PrintHashDetails() {
	if !vo.Enabled || len(vo.Hashes) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "\nHash Details:")
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))

	// Group hashes by content type
	byType := make(map[string][]HashInfo)
	for _, hi := range vo.Hashes {
		byType[hi.ContentType] = append(byType[hi.ContentType], hi)
	}

	// Print in order
	for _, contentType := range []string{"script", "style-tag", "style-attr", "event-handler"} {
		hashes := byType[contentType]
		if len(hashes) == 0 {
			continue
		}

		fmt.Fprintf(os.Stderr, "\n%s:\n", formatContentType(contentType))
		for i, hi := range hashes {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, hi.Hash)
			fmt.Fprintf(os.Stderr, "      File: %s\n", hi.SourceFile)
			fmt.Fprintf(os.Stderr, "      Content: %s\n", hi.Snippet)
		}
	}

	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))
}

// PrintSummary prints an overall summary
func (vo *VerboseOutput) PrintSummary(totalScripts, totalStyleTags, totalStyleAttrs, uniqueScripts, uniqueStyleTags, uniqueStyleAttrs int) {
	if !vo.Enabled {
		return
	}

	fmt.Fprintln(os.Stderr, "\nSummary:")
	fmt.Fprintf(os.Stderr, "  Total inline scripts: %d (unique: %d)\n", totalScripts, uniqueScripts)
	fmt.Fprintf(os.Stderr, "  Total <style> tags: %d (unique: %d)\n", totalStyleTags, uniqueStyleTags)
	fmt.Fprintf(os.Stderr, "  Total style attributes: %d (unique: %d)\n", totalStyleAttrs, uniqueStyleAttrs)
	fmt.Fprintln(os.Stderr, "")
}

// SetExternalResources stores external resources for verbose output
func (vo *VerboseOutput) SetExternalResources(resources *ExternalResources) {
	if !vo.Enabled {
		return
	}
	vo.ExternalResources = resources
}

// PrintExternalResources prints information about detected external resources
func (vo *VerboseOutput) PrintExternalResources() {
	if !vo.Enabled || vo.ExternalResources == nil {
		return
	}

	totalResources := len(vo.ExternalResources.Scripts) +
		len(vo.ExternalResources.Stylesheets) +
		len(vo.ExternalResources.Images) +
		len(vo.ExternalResources.Fonts) +
		len(vo.ExternalResources.Frames) +
		len(vo.ExternalResources.Other)

	if totalResources == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "\nExternal Resources:")
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))

	if len(vo.ExternalResources.Scripts) > 0 {
		fmt.Fprintln(os.Stderr, "\nExternal Scripts:")
		for i, res := range vo.ExternalResources.Scripts {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, res.URL)
			if res.Domain != "" {
				fmt.Fprintf(os.Stderr, "      Domain: %s\n", res.Domain)
			}
		}
	}

	if len(vo.ExternalResources.Stylesheets) > 0 {
		fmt.Fprintln(os.Stderr, "\nExternal Stylesheets:")
		for i, res := range vo.ExternalResources.Stylesheets {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, res.URL)
			if res.Domain != "" {
				fmt.Fprintf(os.Stderr, "      Domain: %s\n", res.Domain)
			}
		}
	}

	if len(vo.ExternalResources.Images) > 0 {
		fmt.Fprintln(os.Stderr, "\nExternal Images:")
		for i, res := range vo.ExternalResources.Images {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, res.URL)
			if res.Domain != "" {
				fmt.Fprintf(os.Stderr, "      Domain: %s\n", res.Domain)
			}
		}
	}

	if len(vo.ExternalResources.Fonts) > 0 {
		fmt.Fprintln(os.Stderr, "\nExternal Fonts:")
		for i, res := range vo.ExternalResources.Fonts {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, res.URL)
			if res.Domain != "" {
				fmt.Fprintf(os.Stderr, "      Domain: %s\n", res.Domain)
			}
		}
	}

	if len(vo.ExternalResources.Frames) > 0 {
		fmt.Fprintln(os.Stderr, "\nExternal Frames:")
		for i, res := range vo.ExternalResources.Frames {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, res.URL)
			if res.Domain != "" {
				fmt.Fprintf(os.Stderr, "      Domain: %s\n", res.Domain)
			}
		}
	}

	if len(vo.ExternalResources.Other) > 0 {
		fmt.Fprintln(os.Stderr, "\nOther External Resources:")
		for i, res := range vo.ExternalResources.Other {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, res.URL)
			if res.Domain != "" {
				fmt.Fprintf(os.Stderr, "      Domain: %s\n", res.Domain)
			}
		}
	}

	// Print unique domains summary
	domains := vo.ExternalResources.GetUniqueDomains()
	if len(domains) > 0 {
		fmt.Fprintf(os.Stderr, "\nUnique Domains (%d):\n", len(domains))
		for _, domain := range domains {
			fmt.Fprintf(os.Stderr, "  - %s\n", domain)
		}
	}

	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))
}

// createSnippet creates a truncated version of content for display
func createSnippet(content string, maxLen int) string {
	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Replace newlines and multiple spaces with single space
	content = strings.Join(strings.Fields(content), " ")

	if len(content) <= maxLen {
		return content
	}

	return content[:maxLen] + "..."
}

// formatContentType returns a human-readable content type name
func formatContentType(contentType string) string {
	switch contentType {
	case "script":
		return "Inline Scripts"
	case "style-tag":
		return "Style Tags"
	case "style-attr":
		return "Style Attributes"
	case "event-handler":
		return "Event Handlers"
	default:
		return contentType
	}
}

package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// ExtractInlineContent parses an HTML file and extracts inline script and style content
// Returns scripts, styleTags, styleAttributes, hasEventHandlers, error
func ExtractInlineContent(filePath string, noScripts, noStyles, noInlineStyles, noEventHandlers bool) (scripts []string, styleTags []string, styleAttributes []string, hasEventHandlers bool, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		return nil, nil, nil, false, fmt.Errorf("failed to parse HTML: %w", err)
	}

	scripts = []string{}
	styleTags = []string{}
	styleAttributes = []string{}
	hasEventHandlers = false

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" && !noScripts {
				// Check if it's an inline script (no src attribute)
				hasSource := false
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						hasSource = true
						break
					}
				}
				if !hasSource {
					// Extract text content
					content := extractTextContent(n)
					scripts = append(scripts, content)
				}
			} else if n.Data == "style" && !noStyles {
				// Extract inline style content
				content := extractTextContent(n)
				styleTags = append(styleTags, content)
			}

			// Extract inline event handler attributes and style attributes from any element
			for _, attr := range n.Attr {
				if isEventHandler(attr.Key) && !noEventHandlers {
					scripts = append(scripts, attr.Val)
					hasEventHandlers = true
					continue
				}
				if strings.EqualFold(attr.Key, "style") && !noInlineStyles {
					styleAttributes = append(styleAttributes, attr.Val)
				}
			}
		}

		// Traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return scripts, styleTags, styleAttributes, hasEventHandlers, nil
}

// isEventHandler checks if an attribute name is an event handler
func isEventHandler(attrName string) bool {
	// List of common event handler attributes
	eventHandlers := []string{
		"onclick", "ondblclick", "onmousedown", "onmouseup", "onmouseover",
		"onmousemove", "onmouseout", "onmouseenter", "onmouseleave",
		"onload", "onunload", "onbeforeunload",
		"onchange", "onsubmit", "onreset", "oninput", "oninvalid",
		"onfocus", "onblur", "onfocusin", "onfocusout",
		"onkeydown", "onkeyup", "onkeypress",
		"onerror", "onabort",
		"onscroll", "onresize",
		"oncontextmenu",
		"ondrag", "ondragstart", "ondragend", "ondragenter", "ondragleave", "ondragover", "ondrop",
		"onwheel",
		"ontouchstart", "ontouchmove", "ontouchend", "ontouchcancel",
		"onplay", "onpause", "onended", "onvolumechange",
		"oncanplay", "oncanplaythrough", "ondurationchange", "onloadeddata", "onloadedmetadata",
		"onprogress", "onseeked", "onseeking", "onstalled", "onsuspend", "ontimeupdate", "onwaiting",
		"onanimationstart", "onanimationend", "onanimationiteration",
		"ontransitionend",
	}

	attrLower := strings.ToLower(attrName)
	for _, handler := range eventHandlers {
		if attrLower == handler {
			return true
		}
	}
	return false
}

// extractTextContent extracts all text content from a node and its children
func extractTextContent(n *html.Node) string {
	var content strings.Builder
	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			content.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extract(c)
	}

	return content.String()
}

// ExtractExternalResources parses an HTML file and extracts external resource URLs
func ExtractExternalResources(filePath string) (*ExternalResources, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	resources := &ExternalResources{
		Scripts:      []ExternalResource{},
		Stylesheets:  []ExternalResource{},
		Images:       []ExternalResource{},
		Fonts:        []ExternalResource{},
		Frames:       []ExternalResource{},
		Other:        []ExternalResource{},
		UsesDataURLs: make(map[string]bool),
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script":
				// Look for external scripts (src attribute)
				for _, attr := range n.Attr {
					if attr.Key == "src" && attr.Val != "" {
						domain := ExtractDomain(attr.Val)
						resources.Scripts = append(resources.Scripts, ExternalResource{
							Type:   "script",
							URL:    attr.Val,
							Domain: domain,
						})
					}
				}
			case "link":
				// Look for stylesheets and fonts
				relType := ""
				href := ""
				for _, attr := range n.Attr {
					if attr.Key == "rel" {
						relType = strings.ToLower(attr.Val)
					}
					if attr.Key == "href" {
						href = attr.Val
					}
				}
				if href != "" {
					if strings.Contains(relType, "stylesheet") {
						domain := ExtractDomain(href)
						resources.Stylesheets = append(resources.Stylesheets, ExternalResource{
							Type:   "stylesheet",
							URL:    href,
							Domain: domain,
						})
					} else if strings.Contains(relType, "font") || strings.Contains(relType, "preload") {
						// Check if it's a font preload
						for _, attr := range n.Attr {
							if attr.Key == "as" && attr.Val == "font" {
								domain := ExtractDomain(href)
								resources.Fonts = append(resources.Fonts, ExternalResource{
									Type:   "font",
									URL:    href,
									Domain: domain,
								})
								break
							}
						}
					}
				}
			case "img":
				// Look for images
				for _, attr := range n.Attr {
					if attr.Key == "src" && attr.Val != "" {
						// Check for data: URLs
						if strings.HasPrefix(attr.Val, "data:") {
							resources.UsesDataURLs["image"] = true
						} else {
							domain := ExtractDomain(attr.Val)
							resources.Images = append(resources.Images, ExternalResource{
								Type:   "image",
								URL:    attr.Val,
								Domain: domain,
							})
						}
					}
				}
			case "iframe":
				// Look for frames
				for _, attr := range n.Attr {
					if attr.Key == "src" && attr.Val != "" {
						domain := ExtractDomain(attr.Val)
						resources.Frames = append(resources.Frames, ExternalResource{
							Type:   "frame",
							URL:    attr.Val,
							Domain: domain,
						})
					}
				}
			case "style":
				// Extract CSS content and parse for URLs
				content := extractTextContent(n)
				if content != "" {
					extractCSSURLs(content, resources)
				}
			}

			// Check for style attributes with @import or url()
			for _, attr := range n.Attr {
				if strings.EqualFold(attr.Key, "style") {
					// Parse CSS for external resources
					extractCSSURLs(attr.Val, resources)
				}
			}
		}

		// Traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return resources, nil
}

// extractCSSURLs extracts URLs from CSS content
func extractCSSURLs(cssContent string, resources *ExternalResources) {
	// Simple regex-like parsing for url() and @import
	// This is a basic implementation; a full CSS parser would be more robust

	// Look for url(...) patterns
	start := 0
	for {
		idx := strings.Index(cssContent[start:], "url(")
		if idx == -1 {
			break
		}
		idx += start + 4

		// Find closing parenthesis
		end := strings.Index(cssContent[idx:], ")")
		if end == -1 {
			break
		}
		end += idx

		urlStr := strings.TrimSpace(cssContent[idx:end])
		urlStr = strings.Trim(urlStr, "\"'")

		if urlStr != "" {
			// Check for data: URLs
			if strings.HasPrefix(urlStr, "data:") {
				// Determine type based on data URL mime type
				lowerURL := strings.ToLower(urlStr)
				if strings.HasPrefix(lowerURL, "data:font/") || strings.Contains(lowerURL, "data:application/font") ||
					strings.Contains(lowerURL, "data:application/x-font") {
					resources.UsesDataURLs["font"] = true
				} else if strings.HasPrefix(lowerURL, "data:image/") {
					resources.UsesDataURLs["image"] = true
				}
				start = end + 1
				continue
			}

			domain := ExtractDomain(urlStr)
			// Try to determine if it's a font based on extension
			lowerURL := strings.ToLower(urlStr)
			if strings.HasSuffix(lowerURL, ".woff") || strings.HasSuffix(lowerURL, ".woff2") ||
				strings.HasSuffix(lowerURL, ".ttf") || strings.HasSuffix(lowerURL, ".otf") ||
				strings.HasSuffix(lowerURL, ".eot") {
				resources.Fonts = append(resources.Fonts, ExternalResource{
					Type:   "font",
					URL:    urlStr,
					Domain: domain,
				})
			} else if strings.HasSuffix(lowerURL, ".jpg") || strings.HasSuffix(lowerURL, ".jpeg") ||
				strings.HasSuffix(lowerURL, ".png") || strings.HasSuffix(lowerURL, ".gif") ||
				strings.HasSuffix(lowerURL, ".svg") || strings.HasSuffix(lowerURL, ".webp") {
				resources.Images = append(resources.Images, ExternalResource{
					Type:   "image",
					URL:    urlStr,
					Domain: domain,
				})
			} else {
				resources.Other = append(resources.Other, ExternalResource{
					Type:   "other",
					URL:    urlStr,
					Domain: domain,
				})
			}
		}

		start = end + 1
	}
}

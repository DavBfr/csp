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

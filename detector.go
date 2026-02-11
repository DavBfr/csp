package main

import (
	"net/url"
	"sort"
	"strings"
)

// ExternalResource represents an external resource found in HTML
type ExternalResource struct {
	Type   string // script, stylesheet, image, font, frame, etc.
	URL    string
	Domain string
}

// ExternalResources contains all detected external resources
type ExternalResources struct {
	Scripts      []ExternalResource
	Stylesheets  []ExternalResource
	Images       []ExternalResource
	Fonts        []ExternalResource
	Frames       []ExternalResource
	Other        []ExternalResource
	UsesDataURLs map[string]bool // Tracks if data: URLs are used for each resource type ("image", "font", "style")
}

// GetUniqueDomains returns a sorted list of unique domains from all resources
func (er *ExternalResources) GetUniqueDomains() []string {
	domainSet := make(map[string]bool)

	resources := [][]ExternalResource{
		er.Scripts, er.Stylesheets, er.Images, er.Fonts, er.Frames, er.Other,
	}

	for _, resList := range resources {
		for _, res := range resList {
			if res.Domain != "" {
				domainSet[res.Domain] = true
			}
		}
	}

	domains := make([]string, 0, len(domainSet))
	for domain := range domainSet {
		domains = append(domains, domain)
	}

	sort.Strings(domains)
	return domains
}

// GetDomainsByType returns unique domains for a specific resource type
func (er *ExternalResources) GetDomainsByType(resourceType string) []string {
	domainSet := make(map[string]bool)

	var resources []ExternalResource
	switch resourceType {
	case "script":
		resources = er.Scripts
	case "stylesheet":
		resources = er.Stylesheets
	case "image":
		resources = er.Images
	case "font":
		resources = er.Fonts
	case "frame":
		resources = er.Frames
	case "other":
		resources = er.Other
	default:
		return []string{}
	}

	for _, res := range resources {
		if res.Domain != "" {
			domainSet[res.Domain] = true
		}
	}

	domains := make([]string, 0, len(domainSet))
	for domain := range domainSet {
		domains = append(domains, domain)
	}

	sort.Strings(domains)
	return domains
}

// ExtractDomain extracts the scheme and host from a URL
// Returns empty string if URL is relative or invalid
func ExtractDomain(rawURL string) string {
	// Skip data URLs
	if strings.HasPrefix(rawURL, "data:") {
		return ""
	}

	// Skip relative URLs (don't start with http/https/protocol)
	if !strings.HasPrefix(rawURL, "http://") &&
		!strings.HasPrefix(rawURL, "https://") &&
		!strings.HasPrefix(rawURL, "//") {
		return ""
	}

	// Handle protocol-relative URLs
	if strings.HasPrefix(rawURL, "//") {
		rawURL = "https:" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Return scheme + host (e.g., "https://example.com")
	if u.Host == "" {
		return ""
	}

	return u.Scheme + "://" + u.Host
}

// AddExternalResourcesToCSP adds external resource domains to appropriate CSP directives
func AddExternalResourcesToCSP(cspHeader string, resources *ExternalResources) string {
	directives := parseCSPDirectives(cspHeader)

	// Add data: to img-src if data URLs are used for images
	if resources.UsesDataURLs != nil && resources.UsesDataURLs["image"] {
		if existing, ok := directives["img-src"]; ok {
			if !strings.Contains(existing, "data:") {
				directives["img-src"] = appendUniqueDomainsToString(existing, []string{"data:"})
			}
		} else if defaultSrc, ok := directives["default-src"]; ok {
			directives["img-src"] = appendUniqueDomainsToString(defaultSrc, []string{"data:"})
		} else {
			directives["img-src"] = "data:"
		}
	}

	// Add data: to font-src if data URLs are used for fonts
	if resources.UsesDataURLs != nil && resources.UsesDataURLs["font"] {
		if existing, ok := directives["font-src"]; ok {
			if !strings.Contains(existing, "data:") {
				directives["font-src"] = appendUniqueDomainsToString(existing, []string{"data:"})
			}
		} else if defaultSrc, ok := directives["default-src"]; ok {
			directives["font-src"] = appendUniqueDomainsToString(defaultSrc, []string{"data:"})
		} else {
			directives["font-src"] = "data:"
		}
	}

	// Add script-src domains
	scriptDomains := resources.GetDomainsByType("script")
	if len(scriptDomains) > 0 {
		if existing, ok := directives["script-src"]; ok {
			directives["script-src"] = appendUniqueDomainsToString(existing, scriptDomains)
		} else if defaultSrc, ok := directives["default-src"]; ok {
			// Create script-src based on default-src
			directives["script-src"] = appendUniqueDomainsToString(defaultSrc, scriptDomains)
		} else {
			directives["script-src"] = strings.Join(scriptDomains, " ")
		}
	}

	// Add style-src domains
	styleDomains := resources.GetDomainsByType("stylesheet")
	if len(styleDomains) > 0 {
		if existing, ok := directives["style-src"]; ok {
			directives["style-src"] = appendUniqueDomainsToString(existing, styleDomains)
		} else if defaultSrc, ok := directives["default-src"]; ok {
			directives["style-src"] = appendUniqueDomainsToString(defaultSrc, styleDomains)
		} else {
			directives["style-src"] = strings.Join(styleDomains, " ")
		}
	}

	// Add img-src domains
	imgDomains := resources.GetDomainsByType("image")
	if len(imgDomains) > 0 {
		if existing, ok := directives["img-src"]; ok {
			directives["img-src"] = appendUniqueDomainsToString(existing, imgDomains)
		} else if defaultSrc, ok := directives["default-src"]; ok {
			directives["img-src"] = appendUniqueDomainsToString(defaultSrc, imgDomains)
		} else {
			directives["img-src"] = strings.Join(imgDomains, " ")
		}
	}

	// Add font-src domains
	fontDomains := resources.GetDomainsByType("font")
	if len(fontDomains) > 0 {
		if existing, ok := directives["font-src"]; ok {
			directives["font-src"] = appendUniqueDomainsToString(existing, fontDomains)
		} else if defaultSrc, ok := directives["default-src"]; ok {
			directives["font-src"] = appendUniqueDomainsToString(defaultSrc, fontDomains)
		} else {
			directives["font-src"] = strings.Join(fontDomains, " ")
		}
	}

	// Add frame-src domains
	frameDomains := resources.GetDomainsByType("frame")
	if len(frameDomains) > 0 {
		if existing, ok := directives["frame-src"]; ok {
			directives["frame-src"] = appendUniqueDomainsToString(existing, frameDomains)
		} else if defaultSrc, ok := directives["default-src"]; ok {
			directives["frame-src"] = appendUniqueDomainsToString(defaultSrc, frameDomains)
		} else {
			directives["frame-src"] = strings.Join(frameDomains, " ")
		}
	}

	// Add connect-src domains (from "other" type)
	connectDomains := resources.GetDomainsByType("other")
	if len(connectDomains) > 0 {
		if existing, ok := directives["connect-src"]; ok {
			directives["connect-src"] = appendUniqueDomainsToString(existing, connectDomains)
		} else if defaultSrc, ok := directives["default-src"]; ok {
			directives["connect-src"] = appendUniqueDomainsToString(defaultSrc, connectDomains)
		} else {
			directives["connect-src"] = strings.Join(connectDomains, " ")
		}
	}

	return reconstructCSP(directives)
}

// appendUniqueDomainsToString appends new domains to an existing space-separated string, removing duplicates
func appendUniqueDomainsToString(existing string, newDomains []string) string {
	seen := make(map[string]bool)
	result := []string{}

	// Add existing values
	if existing != "" {
		for _, val := range strings.Fields(existing) {
			if !seen[val] {
				seen[val] = true
				result = append(result, val)
			}
		}
	}

	// Add new domains
	for _, domain := range newDomains {
		if !seen[domain] {
			seen[domain] = true
			result = append(result, domain)
		}
	}

	return strings.Join(result, " ")
}

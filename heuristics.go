package main

import (
	"regexp"
	"strings"
)

// HeuristicResource represents an inferred external resource
type HeuristicResource struct {
	URL        string
	Type       string
	Confidence string // "high", "medium", "low"
	Reason     string
	SourceURL  string // The URL that triggered this inference
	SourceType string // The type of the source resource
}

// ApplyHeuristics analyzes existing external resources and infers additional ones
func ApplyHeuristics(resources []ExternalResource) []HeuristicResource {
	var inferred []HeuristicResource
	seen := make(map[string]bool)

	for _, resource := range resources {
		// Apply all heuristic rules
		inferred = append(inferred, inferFromStylesheet(resource, seen)...)
		inferred = append(inferred, inferFromScript(resource, seen)...)
		inferred = append(inferred, inferFromImage(resource, seen)...)
		inferred = append(inferred, inferFromHTML(resource, seen)...)
	}

	return inferred
}

// inferFromStylesheet applies heuristics for stylesheets
func inferFromStylesheet(resource ExternalResource, seen map[string]bool) []HeuristicResource {
	if resource.Type != "stylesheet" {
		return nil
	}

	var inferred []HeuristicResource
	urlStr := strings.ToLower(resource.URL)
	domain := ExtractDomain(resource.URL)

	// Heuristic 1: Stylesheets with "font" in name likely load fonts
	if strings.Contains(urlStr, "font") {
		// Just add the domain for font resources, not specific paths
		if !seen[domain+"-font-inference"] {
			inferred = append(inferred, HeuristicResource{
				URL:        domain,
				Type:       "font",
				Confidence: "high",
				Reason:     "Stylesheet name contains 'font' keyword",
				SourceURL:  resource.URL,
				SourceType: "stylesheet",
			})
			seen[domain+"-font-inference"] = true
		}
	}

	// Heuristic 2: Google Fonts CSS always loads font files
	if strings.Contains(domain, "fonts.googleapis.com") {
		fontDomain := "https://fonts.gstatic.com"
		if !seen[fontDomain] {
			inferred = append(inferred, HeuristicResource{
				URL:        fontDomain,
				Type:       "font",
				Confidence: "high",
				Reason:     "Google Fonts CSS always loads from fonts.gstatic.com",
				SourceURL:  resource.URL,
				SourceType: "stylesheet",
			})
			seen[fontDomain] = true
		}
	}

	// Heuristic 3: Icon fonts (fontawesome, material icons, etc.)
	iconFontPatterns := []string{"fontawesome", "font-awesome", "material-icons", "icomoon", "glyphicons"}
	for _, pattern := range iconFontPatterns {
		if strings.Contains(urlStr, pattern) {
			fontDomain := domain
			if !seen[fontDomain] {
				inferred = append(inferred, HeuristicResource{
					URL:        fontDomain,
					Type:       "font",
					Confidence: "high",
					Reason:     "Icon font library detected (" + pattern + ")",
					SourceURL:  resource.URL,
					SourceType: "stylesheet",
				})
				seen[fontDomain] = true
			}
			break
		}
	}

	// Heuristic 4: Bootstrap/framework CSS may load fonts
	frameworkPatterns := []string{"bootstrap", "foundation", "bulma", "tailwind"}
	for _, pattern := range frameworkPatterns {
		if strings.Contains(urlStr, pattern) {
			if !seen[domain+"-fonts"] {
				inferred = append(inferred, HeuristicResource{
					URL:        domain,
					Type:       "font",
					Confidence: "medium",
					Reason:     "CSS framework may include custom fonts (" + pattern + ")",
					SourceURL:  resource.URL,
					SourceType: "stylesheet",
				})
				seen[domain+"-fonts"] = true
			}
			break
		}
	}

	// Heuristic 5: CSS from CDNs often loads other resources
	cdnPatterns := []string{"cdn.jsdelivr.net", "unpkg.com", "cdnjs.cloudflare.com"}
	for _, pattern := range cdnPatterns {
		if strings.Contains(domain, pattern) {
			if !seen[domain+"-connect"] {
				inferred = append(inferred, HeuristicResource{
					URL:        domain,
					Type:       "connect",
					Confidence: "medium",
					Reason:     "CDN may dynamically load additional resources",
					SourceURL:  resource.URL,
					SourceType: "stylesheet",
				})
				seen[domain+"-connect"] = true
			}
			break
		}
	}

	return inferred
}

// inferFromScript applies heuristics for scripts
func inferFromScript(resource ExternalResource, seen map[string]bool) []HeuristicResource {
	if resource.Type != "script" {
		return nil
	}

	var inferred []HeuristicResource
	urlStr := strings.ToLower(resource.URL)
	domain := ExtractDomain(resource.URL)

	// Heuristic 1: Analytics scripts connect back to their domains
	analyticsPatterns := map[string]string{
		"google-analytics.com": "google-analytics.com",
		"googletagmanager.com": "google-analytics.com",
		"analytics.js":         domain,
		"gtag/js":              "google-analytics.com",
		"ga.js":                "google-analytics.com",
		"analytics":            domain,
	}

	for pattern, connectDomain := range analyticsPatterns {
		if strings.Contains(urlStr, pattern) {
			if !seen[connectDomain+"-connect"] {
				inferred = append(inferred, HeuristicResource{
					URL:        connectDomain,
					Type:       "connect",
					Confidence: "high",
					Reason:     "Analytics/tracking script needs to send data",
					SourceURL:  resource.URL,
					SourceType: "script",
				})
				seen[connectDomain+"-connect"] = true
			}
			break
		}
	}

	// Heuristic 2: React/Vue/Angular may load chunk files dynamically
	frameworkPatterns := []string{"react", "vue", "angular", "chunk", "bundle"}
	for _, pattern := range frameworkPatterns {
		if strings.Contains(urlStr, pattern) {
			if !seen[domain+"-script-chunks"] {
				inferred = append(inferred, HeuristicResource{
					URL:        domain,
					Type:       "script",
					Confidence: "high",
					Reason:     "JavaScript framework may lazy-load additional chunks",
					SourceURL:  resource.URL,
					SourceType: "script",
				})
				seen[domain+"-script-chunks"] = true
			}
			break
		}
	}

	// Heuristic 3: Payment processors (Stripe, PayPal, etc.)
	paymentPatterns := map[string]string{
		"stripe.com":    "stripe.com",
		"paypal.com":    "paypal.com",
		"square.com":    "square.com",
		"braintree.com": "braintreegateway.com",
	}

	for pattern, connectDomain := range paymentPatterns {
		if strings.Contains(domain, pattern) {
			if !seen[connectDomain+"-connect"] {
				inferred = append(inferred, HeuristicResource{
					URL:        connectDomain,
					Type:       "connect",
					Confidence: "high",
					Reason:     "Payment processor needs API connection",
					SourceURL:  resource.URL,
					SourceType: "script",
				})
				seen[connectDomain+"-connect"] = true
			}

			if !seen[connectDomain+"-frame"] {
				inferred = append(inferred, HeuristicResource{
					URL:        connectDomain,
					Type:       "frame",
					Confidence: "high",
					Reason:     "Payment processor may use iframes",
					SourceURL:  resource.URL,
					SourceType: "script",
				})
				seen[connectDomain+"-frame"] = true
			}
			break
		}
	}

	// Heuristic 4: Social media widgets
	socialPatterns := map[string][]string{
		"facebook":  {"connect.facebook.net", "facebook.com"},
		"twitter":   {"platform.twitter.com", "twitter.com"},
		"linkedin":  {"platform.linkedin.com", "linkedin.com"},
		"instagram": {"instagram.com"},
		"youtube":   {"youtube.com"},
	}

	for pattern, domains := range socialPatterns {
		if strings.Contains(domain, pattern) {
			for _, connectDomain := range domains {
				key := connectDomain + "-social"
				if !seen[key] {
					inferred = append(inferred, HeuristicResource{
						URL:        connectDomain,
						Type:       "connect",
						Confidence: "high",
						Reason:     "Social media widget needs API access",
						SourceURL:  resource.URL,
						SourceType: "script",
					})
					seen[key] = true
				}
			}
			break
		}
	}

	// Heuristic 5: Polyfill services
	if strings.Contains(urlStr, "polyfill") {
		if !seen[domain+"-polyfill"] {
			inferred = append(inferred, HeuristicResource{
				URL:        domain,
				Type:       "script",
				Confidence: "medium",
				Reason:     "Polyfill service may serve different files based on user agent",
				SourceURL:  resource.URL,
				SourceType: "script",
			})
			seen[domain+"-polyfill"] = true
		}
	}

	return inferred
}

// inferFromImage applies heuristics for images
func inferFromImage(resource ExternalResource, seen map[string]bool) []HeuristicResource {
	if resource.Type != "image" {
		return nil
	}

	var inferred []HeuristicResource
	urlStr := strings.ToLower(resource.URL)
	domain := ExtractDomain(resource.URL)

	// Heuristic 1: CDN images suggest more images from same CDN
	cdnPatterns := []string{"cloudinary", "imgix", "cloudflare", "fastly", "akamai", "cloudfront"}
	for _, pattern := range cdnPatterns {
		if strings.Contains(domain, pattern) {
			if !seen[domain+"-img"] {
				inferred = append(inferred, HeuristicResource{
					URL:        domain,
					Type:       "image",
					Confidence: "high",
					Reason:     "CDN domain likely serves multiple images",
					SourceURL:  resource.URL,
					SourceType: "image",
				})
				seen[domain+"-img"] = true
			}
			break
		}
	}

	// Heuristic 2: Responsive images (srcset patterns)
	responsivePatterns := regexp.MustCompile(`[-_@](xs|sm|md|lg|xl|[0-9]+x|2x|3x|retina)|@[0-9]x`)
	if responsivePatterns.MatchString(urlStr) {
		if !seen[domain+"-responsive"] {
			inferred = append(inferred, HeuristicResource{
				URL:        domain,
				Type:       "image",
				Confidence: "high",
				Reason:     "Responsive image pattern detected, likely has multiple variants",
				SourceURL:  resource.URL,
				SourceType: "image",
			})
			seen[domain+"-responsive"] = true
		}
	}

	// Heuristic 3: Avatar/profile images suggest user-generated content
	avatarPatterns := []string{"/avatar", "/profile", "/user", "/photo"}
	for _, pattern := range avatarPatterns {
		if strings.Contains(urlStr, pattern) {
			if !seen[domain+"-ugc"] {
				inferred = append(inferred, HeuristicResource{
					URL:        domain,
					Type:       "image",
					Confidence: "medium",
					Reason:     "User-generated content pattern detected",
					SourceURL:  resource.URL,
					SourceType: "image",
				})
				seen[domain+"-ugc"] = true
			}
			break
		}
	}

	return inferred
}

// inferFromHTML applies general heuristics
func inferFromHTML(resource ExternalResource, seen map[string]bool) []HeuristicResource {
	var inferred []HeuristicResource
	domain := ExtractDomain(resource.URL)

	// Heuristic 1: API domains (common patterns)
	apiPatterns := []string{"api.", "/api/", "graphql", "rest"}
	urlStr := strings.ToLower(resource.URL)

	for _, pattern := range apiPatterns {
		if strings.Contains(urlStr, pattern) || strings.Contains(domain, "api.") {
			if !seen[domain+"-api"] {
				inferred = append(inferred, HeuristicResource{
					URL:        domain,
					Type:       "connect",
					Confidence: "high",
					Reason:     "API endpoint detected",
					SourceURL:  resource.URL,
					SourceType: resource.Type,
				})
				seen[domain+"-api"] = true
			}
			break
		}
	}

	return inferred
}

// ConvertHeuristicToExternalResource converts heuristic resources to external resources
func ConvertHeuristicToExternalResource(heuristic HeuristicResource) ExternalResource {
	// Ensure URL has scheme
	url := heuristic.URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	return ExternalResource{
		URL:    url,
		Type:   heuristic.Type,
		Domain: ExtractDomain(url),
	}
}

// GetHeuristicsSummary returns a formatted summary of inferred resources
func GetHeuristicsSummary(heuristics []HeuristicResource) map[string]int {
	summary := make(map[string]int)

	for _, h := range heuristics {
		summary[h.Type]++
		summary["confidence_"+h.Confidence]++
	}

	return summary
}

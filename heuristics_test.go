package main

import (
	"strings"
	"testing"
)

func TestApplyHeuristics_Fonts(t *testing.T) {
	tests := []struct {
		name                string
		resources           []ExternalResource
		expectedInferences  int
		shouldContainType   string
		shouldContainReason string
	}{
		{
			name: "stylesheet with font in name",
			resources: []ExternalResource{
				{URL: "https://example.com/css/fonts-awesome.css", Type: "stylesheet"},
			},
			expectedInferences:  5, // woff2, woff, ttf, eot, otf
			shouldContainType:   "font",
			shouldContainReason: "Stylesheet name contains 'font' keyword",
		},
		{
			name: "Google Fonts stylesheet",
			resources: []ExternalResource{
				{URL: "https://fonts.googleapis.com/css?family=Roboto", Type: "stylesheet"},
			},
			expectedInferences:  1,
			shouldContainType:   "font",
			shouldContainReason: "Google Fonts CSS always loads from fonts.gstatic.com",
		},
		{
			name: "FontAwesome stylesheet",
			resources: []ExternalResource{
				{URL: "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css", Type: "stylesheet"},
			},
			expectedInferences:  1, // At least one inference
			shouldContainType:   "font",
			shouldContainReason: "font", // Just check if "font" is mentioned
		},
		{
			name: "Bootstrap CSS",
			resources: []ExternalResource{
				{URL: "https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css", Type: "stylesheet"},
			},
			expectedInferences:  1, // At least one inference
			shouldContainType:   "font",
			shouldContainReason: "bootstrap", // Match on framework name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inferred := ApplyHeuristics(tt.resources)

			if len(inferred) < 1 {
				t.Errorf("Expected at least 1 inference, got %d", len(inferred))
				return
			}

			// Check if expected type exists
			foundType := false
			foundReason := false
			for _, h := range inferred {
				if h.Type == tt.shouldContainType {
					foundType = true
				}
				if strings.Contains(strings.ToLower(h.Reason), strings.ToLower(tt.shouldContainReason)) {
					foundReason = true
				}
			}

			if !foundType {
				t.Errorf("Expected to find type %s in inferences", tt.shouldContainType)
			}

			if !foundReason {
				t.Errorf("Expected to find reason containing: %s", tt.shouldContainReason)
			}
		})
	}
}

func TestApplyHeuristics_Analytics(t *testing.T) {
	tests := []struct {
		name         string
		resources    []ExternalResource
		expectType   string
		expectDomain string
	}{
		{
			name: "Google Analytics script",
			resources: []ExternalResource{
				{URL: "https://www.google-analytics.com/analytics.js", Type: "script"},
			},
			expectType:   "connect",
			expectDomain: "google-analytics.com", // Just check for the main domain
		},
		{
			name: "Google Tag Manager",
			resources: []ExternalResource{
				{URL: "https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXX", Type: "script"},
			},
			expectType:   "connect",
			expectDomain: "google-analytics.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inferred := ApplyHeuristics(tt.resources)

			if len(inferred) == 0 {
				t.Error("Expected at least 1 inference")
				return
			}

			found := false
			for _, h := range inferred {
				if h.Type == tt.expectType && strings.Contains(h.URL, tt.expectDomain) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find type=%s with domain=%s", tt.expectType, tt.expectDomain)
			}
		})
	}
}

func TestApplyHeuristics_Frameworks(t *testing.T) {
	tests := []struct {
		name      string
		resources []ExternalResource
		wantType  string
	}{
		{
			name: "React script suggests chunks",
			resources: []ExternalResource{
				{URL: "https://unpkg.com/react@17/umd/react.production.min.js", Type: "script"},
			},
			wantType: "script",
		},
		{
			name: "Vue script suggests chunks",
			resources: []ExternalResource{
				{URL: "https://cdn.jsdelivr.net/npm/vue@3.2.31/dist/vue.global.js", Type: "script"},
			},
			wantType: "script",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inferred := ApplyHeuristics(tt.resources)

			if len(inferred) == 0 {
				t.Error("Expected at least 1 inference")
				return
			}

			found := false
			for _, h := range inferred {
				if h.Type == tt.wantType {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find type=%s in inferences", tt.wantType)
			}
		})
	}
}

func TestApplyHeuristics_PaymentProcessors(t *testing.T) {
	tests := []struct {
		name          string
		resources     []ExternalResource
		expectConnect bool
		expectFrame   bool
	}{
		{
			name: "Stripe script",
			resources: []ExternalResource{
				{URL: "https://js.stripe.com/v3/", Type: "script"},
			},
			expectConnect: true,
			expectFrame:   true,
		},
		{
			name: "PayPal script",
			resources: []ExternalResource{
				{URL: "https://www.paypal.com/sdk/js?client-id=xxx", Type: "script"},
			},
			expectConnect: true,
			expectFrame:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inferred := ApplyHeuristics(tt.resources)

			if len(inferred) == 0 {
				t.Error("Expected at least 1 inference")
				return
			}

			foundConnect := false
			foundFrame := false
			for _, h := range inferred {
				if h.Type == "connect" {
					foundConnect = true
				}
				if h.Type == "frame" {
					foundFrame = true
				}
			}

			if tt.expectConnect && !foundConnect {
				t.Error("Expected to find connect type")
			}

			if tt.expectFrame && !foundFrame {
				t.Error("Expected to find frame type")
			}
		})
	}
}

func TestApplyHeuristics_Images(t *testing.T) {
	tests := []struct {
		name       string
		resources  []ExternalResource
		wantReason string
	}{
		{
			name: "CDN image",
			resources: []ExternalResource{
				{URL: "https://res.cloudinary.com/demo/image/upload/sample.jpg", Type: "image"},
			},
			wantReason: "CDN domain likely serves multiple images",
		},
		{
			name: "responsive image",
			resources: []ExternalResource{
				{URL: "https://example.com/images/photo-1920x1080.jpg", Type: "image"},
			},
			wantReason: "Responsive image pattern detected, likely has multiple variants",
		},
		{
			name: "retina image",
			resources: []ExternalResource{
				{URL: "https://example.com/images/logo@2x.png", Type: "image"},
			},
			wantReason: "Responsive image pattern detected, likely has multiple variants",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inferred := ApplyHeuristics(tt.resources)

			if len(inferred) == 0 {
				t.Error("Expected at least 1 inference")
				return
			}

			found := false
			for _, h := range inferred {
				if h.Reason == tt.wantReason {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find reason: %s", tt.wantReason)
			}
		})
	}
}

func TestApplyHeuristics_SocialMedia(t *testing.T) {
	tests := []struct {
		name      string
		resources []ExternalResource
		wantType  string
	}{
		{
			name: "Facebook SDK",
			resources: []ExternalResource{
				{URL: "https://connect.facebook.net/en_US/sdk.js", Type: "script"},
			},
			wantType: "connect",
		},
		{
			name: "Twitter widget",
			resources: []ExternalResource{
				{URL: "https://platform.twitter.com/widgets.js", Type: "script"},
			},
			wantType: "connect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inferred := ApplyHeuristics(tt.resources)

			if len(inferred) == 0 {
				t.Error("Expected at least 1 inference")
				return
			}

			found := false
			for _, h := range inferred {
				if h.Type == tt.wantType {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find type=%s", tt.wantType)
			}
		})
	}
}

func TestConvertHeuristicToExternalResource(t *testing.T) {
	heuristic := HeuristicResource{
		URL:        "https://example.com/font.woff2",
		Type:       "font",
		Confidence: "high",
		Reason:     "Test reason",
	}

	result := ConvertHeuristicToExternalResource(heuristic)

	if result.URL != heuristic.URL {
		t.Errorf("Expected URL %s, got %s", heuristic.URL, result.URL)
	}

	if result.Type != heuristic.Type {
		t.Errorf("Expected Type %s, got %s", heuristic.Type, result.Type)
	}
}

func TestGetHeuristicsSummary(t *testing.T) {
	heuristics := []HeuristicResource{
		{Type: "font", Confidence: "high"},
		{Type: "font", Confidence: "high"},
		{Type: "script", Confidence: "medium"},
		{Type: "connect", Confidence: "high"},
	}

	summary := GetHeuristicsSummary(heuristics)

	if summary["font"] != 2 {
		t.Errorf("Expected 2 fonts, got %d", summary["font"])
	}

	if summary["script"] != 1 {
		t.Errorf("Expected 1 script, got %d", summary["script"])
	}

	if summary["confidence_high"] != 3 {
		t.Errorf("Expected 3 high confidence, got %d", summary["confidence_high"])
	}

	if summary["confidence_medium"] != 1 {
		t.Errorf("Expected 1 medium confidence, got %d", summary["confidence_medium"])
	}
}

func TestApplyHeuristics_NoDuplicates(t *testing.T) {
	resources := []ExternalResource{
		{URL: "https://fonts.googleapis.com/css?family=Roboto", Type: "stylesheet"},
		{URL: "https://fonts.googleapis.com/css?family=Open+Sans", Type: "stylesheet"},
	}

	inferred := ApplyHeuristics(resources)

	// Should only add fonts.gstatic.com once, not twice
	fontStaticCount := 0
	for _, h := range inferred {
		if h.URL == "https://fonts.gstatic.com" {
			fontStaticCount++
		}
	}

	if fontStaticCount != 1 {
		t.Errorf("Expected fonts.gstatic.com to appear once, got %d times", fontStaticCount)
	}
}

package main

import (
	"os"
	"testing"
)

func TestExtractInlineContent(t *testing.T) {
	html := `<html><head><script>console.log('test');</script></head></html>`
	tmpfile, err := os.CreateTemp("", "test*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write([]byte(html))
	tmpfile.Close()

	scripts, _, _, _, err := ExtractInlineContent(tmpfile.Name(), false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(scripts))
	}
}

func TestIsEventHandler(t *testing.T) {
	if !isEventHandler("onclick") {
		t.Error("onclick should be recognized as event handler")
	}
	if isEventHandler("class") {
		t.Error("class should not be recognized as event handler")
	}
}

func TestExtractExternalResourcesWithDataURLs(t *testing.T) {
	tests := []struct {
		name             string
		html             string
		expectImageData  bool
		expectFontData   bool
		expectImageCount int
	}{
		{
			name:            "img tag with data URL",
			html:            `<html><body><img src="data:image/png;base64,iVBORw0KGgo="></body></html>`,
			expectImageData: true,
		},
		{
			name:             "img tag with regular URL",
			html:             `<html><body><img src="https://example.com/image.png"></body></html>`,
			expectImageData:  false,
			expectImageCount: 1,
		},
		{
			name:            "style attribute with data URL image",
			html:            `<html><body><div style="background: url('data:image/svg+xml,<svg></svg>')"></div></body></html>`,
			expectImageData: true,
		},
		{
			name:           "style attribute with data URL font",
			html:           `<html><head><style>@font-face { src: url('data:font/woff2;base64,ABC123'); }</style></head></html>`,
			expectFontData: true,
		},
		{
			name:             "mixed data and regular URLs",
			html:             `<html><body><img src="data:image/png;base64,ABC"><img src="https://example.com/img.png"></body></html>`,
			expectImageData:  true,
			expectImageCount: 1,
		},
		{
			name:             "no data URLs",
			html:             `<html><body><img src="https://example.com/image.png"><script src="https://example.com/script.js"></script></body></html>`,
			expectImageData:  false,
			expectImageCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test*.html")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())
			tmpfile.Write([]byte(tt.html))
			tmpfile.Close()

			resources, err := ExtractExternalResources(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}

			if resources.UsesDataURLs["image"] != tt.expectImageData {
				t.Errorf("Expected image data URL usage %v, got %v", tt.expectImageData, resources.UsesDataURLs["image"])
			}

			if resources.UsesDataURLs["font"] != tt.expectFontData {
				t.Errorf("Expected font data URL usage %v, got %v", tt.expectFontData, resources.UsesDataURLs["font"])
			}

			if tt.expectImageCount > 0 && len(resources.Images) != tt.expectImageCount {
				t.Errorf("Expected %d regular images, got %d", tt.expectImageCount, len(resources.Images))
			}
		})
	}
}

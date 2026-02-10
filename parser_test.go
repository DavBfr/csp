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

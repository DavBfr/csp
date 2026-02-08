# CSP Hash Generator

A command-line tool for generating Content Security Policy (CSP) hashes from HTML files and updating CSP headers.

## Overview

This tool scans HTML files for inline `<script>` and `<style>` elements, inline event handlers (onclick, onload, etc.), computes SHA-256 hashes of their content, and injects those hashes into an existing CSP header string. This is useful for implementing strict CSP policies while still allowing specific inline scripts and styles.

When inline event handlers are detected, the tool automatically adds the `'unsafe-hashes'` directive to `script-src`, which is required by the CSP specification for event handlers to work with hashes.

## Installation

```bash
# Clone or navigate to this directory
cd csp

# Build the binary
go build -o csp

# Or install it globally
go install
```

## Usage

```bash
csp --csp "YOUR_CSP_HEADER" file1.html file2.html [file3.html ...]
```

### Arguments

- `--csp` (required): The existing CSP header string to update with hashes
- Additional arguments: One or more HTML file paths to scan

### Example

```bash
./csp --csp "default-src 'self'; script-src 'self'; style-src 'self'" index.html about.html
```

**Output:**

```text
default-src 'self'; script-src 'self' 'sha256-xyz123...'; style-src 'self' 'sha256-abc456...'
```

## How It Works

1. **Parses HTML files** to find:
   - Inline `<script>` tags (without `src` attribute)
   - Inline `<style>` tags
   - Inline event handler attributes (onclick, onload, onmouseover, etc.)
2. **Extracts the exact content** between tags and from attributes, preserving whitespace
3. **Computes SHA-256 hashes** of each inline script, style, and event handler
4. **Updates the CSP header** by adding hashes to `script-src` and `style-src` directives
5. **Adds `'unsafe-hashes'`** to `script-src` if event handlers were found (required by CSP spec)
6. **Outputs the updated CSP header** to stdout

## CSP Hash Format

Hashes are formatted according to the CSP specification:

- `'sha256-<base64-encoded-hash>'` (note the single quotes)
- Hashes are added to the appropriate directive (`script-src` for scripts, `style-src` for styles)
- If a directive doesn't exist in the input CSP, it will be created

## Example Workflow

Given an `index.html` file:

```html
<!DOCTYPE html>
<html>
  <head>
    <style>
      body {
        margin: 0;
      }
    </style>
  </head>
  <body>
    <script>
      console.log("Hello World");
    </script>
  </body>
</html>
```

Running:

```bash
./csp --csp "default-src 'self'; script-src 'self'" index.html
```

Produces:

```text
default-src 'self'; script-src 'self' 'sha256-jMeBDFyMMj3eH3XVRDI6d1kH0vcN/4mPrX8L0VVa+G0='; style-src 'sha256-fPMc5i1n0CrQXXE2yCpVdF0E5G0Y3wSGsKjQZHqKSvU='
```

## Notes

- Only inline scripts, styles, and event handlers are hashed (external resources via `src` or `href` are ignored)
- Event handlers include: onclick, onload, onmouseover, onsubmit, oninput, and many others
- When event handlers are detected, `'unsafe-hashes'` is automatically added to `script-src` (required by CSP)
- Duplicate hashes are automatically removed
- Multiple HTML files can be processed in one command
- The tool preserves the structure and order of your existing CSP directives
- Empty inline scripts/styles will still generate hashes

## Requirements

- Go 1.16 or later
- `golang.org/x/net/html` package (automatically installed via `go.mod`)

## License

BSD-3-Clause. See LICENSE.

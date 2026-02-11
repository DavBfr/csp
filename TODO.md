# TODO - Feature Roadmap

## High Priority Features

### Multiple Hash Algorithms

- [x] Add support for SHA-384 hashing
- [x] Add support for SHA-512 hashing
- [x] Add flag `--hash-algo sha256|sha384|sha512` (default: sha256)
- [x] Update tests to cover all hash algorithms

### Report/Dry-Run Mode

- [ ] Add `--report` or `--dry-run` flag
- [ ] Show what would be changed without modifying CSP
- [ ] Display detailed analysis per file:
  - Number of inline scripts found
  - Number of inline styles found
  - Number of style attributes found
  - Number of event handlers found
- [ ] Show before/after CSP comparison

### Verbose Mode

- [ ] Add `-v` or `--verbose` flag
- [ ] Show which file each hash came from
- [ ] Display content snippet for each hash
- [ ] Show processing progress for multiple files

## Medium Priority Features

### JSON Output

- [ ] Add `--output json` or `--format json` flag
- [ ] Output structured data for CI/CD integration
- [ ] Include file paths, hash values, and metadata
- [ ] Support both summary and detailed JSON formats

### External Resource Detection

- [ ] Scan for external scripts (src="...")
- [ ] Scan for external stylesheets (href="...")
- [ ] Extract domains from external resources
- [ ] Add flag `--include-external` to add domains to CSP directives
- [ ] Support for frame-src, img-src, font-src detection

### CSP Validation

- [x] Validate CSP syntax before modifications
- [x] Validate CSP syntax after modifications
- [x] Warn about common misconfigurations:
  - `'unsafe-inline'` used with hashes (hashes are ignored)
  - `'unsafe-eval'` without necessity
  - Missing required directives
- [x] Add `--validate-only` flag to just check CSP syntax

### Strict CSP Generator

- [ ] Add `--generate-strict` flag
- [ ] Generate a complete strict CSP from scratch
- [ ] Include recommended base directives (default-src, etc.)
- [ ] Option to start from a template rather than empty CSP

## Lower Priority Features

### Nonce Generation

- [ ] Add `--nonce` flag to generate random nonce
- [ ] Output nonce value for runtime injection
- [ ] Support template placeholders in HTML for nonce replacement
- [ ] Document nonce vs hash tradeoffs

### Watch Mode

- [ ] Add `--watch` flag
- [ ] Monitor HTML files for changes
- [ ] Auto-regenerate CSP on file modification
- [ ] Debounce rapid changes
- [ ] Display notifications on CSP updates

### Meta Tag Support

- [ ] Add `--meta-tag` flag
- [ ] Read CSP from HTML `<meta http-equiv="Content-Security-Policy">` tags
- [ ] Write updated CSP back to meta tags
- [ ] Support both header and meta tag output simultaneously

### Config File Support

- [ ] Support `.csprc.json` configuration file
- [ ] Support `csp.yaml` configuration file
- [ ] Allow defining:
  - Default hash algorithm
  - File patterns to include/exclude
  - Default flags
  - Custom CSP templates
- [ ] Add `--config` flag to specify custom config path

### Directive-specific Output

- [ ] Add `--directive script-src|style-src|img-src` flag
- [ ] Output only specified directive(s)
- [ ] Support multiple directive filtering
- [ ] Useful for modular CSP assembly

## Quality of Life Improvements

### Better Error Messages

- [ ] Improve error messages for invalid HTML
- [ ] Show line numbers for parsing errors
- [ ] Suggest fixes for common issues
- [ ] Color-coded error output

### Performance Optimizations

- [ ] Parallel processing of multiple HTML files
- [ ] Streaming parser for large files
- [ ] Cache hash computations for unchanged files
- [ ] Benchmark and optimize hot paths

### Documentation

- [ ] Add more examples to README
- [ ] Create troubleshooting guide
- [ ] Document CSP best practices
- [ ] Add comparison with other CSP tools
- [ ] Video tutorial or GIF demos

### Testing

- [ ] Increase test coverage to >80%
- [ ] Add integration tests with real HTML files
- [ ] Add benchmark tests
- [ ] Test edge cases (malformed HTML, huge files, etc.)

## Future Ideas

### Browser Extension

- [ ] Chrome/Firefox extension to generate CSP from current page
- [ ] Live CSP testing and validation
- [ ] Visual indicators for blocked resources

### CI/CD Integration

- [ ] GitHub Action for automatic CSP generation
- [ ] GitLab CI template
- [ ] Pre-commit hook integration
- [ ] Automated PR comments with CSP changes

### HTML Rewriting

- [ ] Option to inject hashes directly into HTML meta tags
- [ ] Option to move inline scripts to external files
- [ ] Auto-refactoring for CSP compliance

### CSP Policy Merging

- [ ] Merge multiple CSP headers intelligently
- [ ] Conflict resolution strategies
- [ ] Policy diff tool

## Completed Features

- [x] Basic SHA-256 hash generation for inline scripts
- [x] Basic SHA-256 hash generation for inline styles
- [x] Style attribute hash generation
- [x] Event handler detection and hashing
- [x] Multiple HTML file processing
- [x] Automatic `'unsafe-hashes'` injection
- [x] CLI flags to disable specific features
  - [x] `--no-scripts`
  - [x] `--no-styles`
  - [x] `--no-inline-styles`
  - [x] `--no-event-handlers`
- [x] Help text with examples
- [x] Unit tests for core functionality
- [x] GitHub Actions CI/CD workflow

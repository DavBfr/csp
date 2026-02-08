# Contributing

Thanks for helping improve CSP Hash Generator.

## Code of Conduct

By participating in this project, you agree to follow the Code of Conduct in
CODE_OF_CONDUCT.md.

## Getting Started

1. Fork the repo and create your branch from main.
2. Install Go (see go.mod for the version).
3. Run tests:

```bash
go test ./...
```

## Development Tips

- Format code before pushing:

```bash
go fmt ./...
```

- Run vet when changing logic:

```bash
go vet ./...
```

## Pull Requests

- Keep PRs focused and small when possible.
- Add tests for new behavior or bug fixes.
- Update README.md when behavior or usage changes.
- If your change should appear in release notes, add an appropriate label
  (bug, feature, docs, breaking, chore).

## Reporting Issues

Please include:
- Go version
- OS and architecture
- Exact command and input sample
- Actual vs expected output

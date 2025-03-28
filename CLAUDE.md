# CLAUDE.md - daiv-github Plugin Guide

## Build & Test Commands
- Build plugin: `make build`
- Install plugin: `make install`
- Clean build artifacts: `make clean`
- Run plugin: `../../daiv/out/daiv standup --prompt`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Test with coverage: `make test-cover`
- View coverage report: `make test-cover-html`

## Code Style Guidelines
- **Imports**: Group standard library, then external packages, then internal packages
- **Error handling**: Use structured errors via the custom error types in github/errors.go
- **Naming**: Use camelCase for vars/functions, PascalCase for exported symbols
- **Types**: Always define interface types for components, provide concrete implementations
- **File structure**: One package per directory, logical separation of concerns
- **Documentation**: All exported functions/types must have godoc comments
- **Organization**: Follow clean architecture (domain models → repos → services → formatters)
- **Concurrency**: Use goroutines for parallel processing where appropriate

## GitHub Integration
Uses the GitHub CLI for authentication. Uses go-github library for API interaction.

## Recent Fixes
- Added GetClient() method to GitHubClient to expose the underlying GitHub client
- Fixed parameter mismatch in NewActivityService call in plugin.go
- Added GetStandupContext implementation to GitHubClient

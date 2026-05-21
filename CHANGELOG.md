# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2026-05-22

### Added

- Integration tests for GitHub service functionality
- TestGetIssue_Integration to verify issue #1 exists with correct title and description
- TestCreateIssue_Integration to test creating and cleaning up issues
- CloseIssue method to GitHubService interface for closing GitHub issues
- Proper cleanup in integration tests by closing created issues

### Changed

- Added CloseIssue method to GitHubService interface and implementation
- Updated github_service.go to implement the CloseIssue method using PATCH request
- Normalized string comparison in integration test for issue body to handle whitespace differences

## [0.1.1] - 2026-05-21

### Added

- Test coverage for empty token edge case in `getToken()` method
- `empty token configured` test case for both `TestCreateIssue` and `TestGetIssue`

### Changed

- Fixed import ordering across all Go files (standard library → local project → external)
- Fixed struct field alignment in `Repository` and `githubService` structs
- Applied consistent formatting with `gofmt` to all source files
- Added trailing newlines to all Go files

## [0.1.0] - 2026-05-21

### Added

- `CreateIssue` method to `GitHubService` interface for creating GitHub issues
- `GetIssue` method to `GitHubService` interface for retrieving GitHub issues
- `Issue` model struct with `Number`, `HTMLURL`, `Title`, `Body`, `State` fields
- `CreateIssueRequest`, `CreateIssueResponse`, `GetIssueResponse` DTOs
- `NewGitHubServiceWithTokens()` constructor for token-based authentication
- Multi-repository token support via `GITHUB_TOKENS` environment variable (JSON map format)
- Sentinel errors: `ErrTokenNotConfigured`, `ErrUnauthorized`, `ErrIssueNotFound`
- `POST /api/v1/repositories/issues` endpoint for issue creation
- `GET /api/v1/repositories/issues/:number` endpoint for issue retrieval
- Table-driven unit tests for `CreateIssue` and `GetIssue` service methods
- Handler integration tests for both new endpoints

### Changed

- Extended `GitHubService` interface with new methods (backward compatible)

### Technical Details

- Token map parsed from `GITHUB_TOKENS` env var at startup in router setup
- Per-repository token lookup uses `owner/repo` key format
- HTTP 401 responses return `ErrUnauthorized`, HTTP 404 returns `ErrIssueNotFound`
- GET endpoint uses query parameters (`?owner=org&repo=repo`) for owner/repo
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
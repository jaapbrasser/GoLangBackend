# github-api

A secure Echo (Go) backend that exposes two GitHub REST API endpoints:

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/api/v1/repos/:owner/:repo/validate` | Check whether a repository exists |
| `POST` | `/api/v1/repos/:owner/:repo/issues`   | Create a new issue in a repository |

---

## Project layout

```
github-api/
├── main.go              # Server bootstrap, route registration
├── github/
│   ├── client.go        # Secure GitHub API client (token never exported)
│   └── models.go        # GitHub request/response types
├── handlers/
│   └── github.go        # Echo route handlers
├── middleware/
│   └── auth.go          # X-API-Key header validation
├── models/
│   └── models.go        # Internal API request/response DTOs
├── .env.example         # Template — copy to .env and fill in
└── .gitignore           # Excludes .env so secrets are never committed
```

---

## Quick start

### 1. Prerequisites

- Go 1.22+
- A [GitHub personal access token](https://github.com/settings/tokens)
  - Classic token scopes needed: **`repo`** (private) or **`public_repo`** (public only)

### 2. Clone & install dependencies

```bash
git clone <your-repo>
cd github-api
go mod tidy
```

### 3. Configure environment

```bash
cp .env.example .env
# Edit .env — set GITHUB_TOKEN and API_KEY
```

**Never commit `.env`** — it is in `.gitignore`.  
In production, inject these as real environment variables via your platform's secrets manager (AWS Secrets Manager, GCP Secret Manager, Vault, etc.).

### 4. Run

```bash
go run .
# Server starts on :8080 (override with PORT=xxxx)
```

---

## API Reference

All endpoints require the `X-API-Key` header matching `API_KEY` in your environment.

### Validate a repository

```
GET /api/v1/repos/{owner}/{repo}/validate
X-API-Key: <your-api-key>
```

**200 — repository found:**
```json
{
  "exists": true,
  "full_name": "octocat/Hello-World",
  "description": "My first repository on GitHub!",
  "private": false,
  "url": "https://github.com/octocat/Hello-World",
  "stars": 2000,
  "language": "C",
  "default_branch": "main"
}
```

**200 — repository not found / inaccessible:**
```json
{ "exists": false }
```

---

### Create an issue

```
POST /api/v1/repos/{owner}/{repo}/issues
X-API-Key: <your-api-key>
Content-Type: application/json

{
  "title": "Bug: login fails on Safari",
  "body": "Steps to reproduce…",
  "labels": ["bug", "high-priority"],
  "assignees": ["octocat"]
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `title` | string | ✅ | 1–256 characters |
| `body` | string | | Up to 65 536 characters |
| `labels` | string[] | | Must already exist in the repo |
| `assignees` | string[] | | GitHub usernames |
| `milestone` | int | | Milestone number |

**201 — issue created:**
```json
{
  "number": 42,
  "title": "Bug: login fails on Safari",
  "body": "Steps to reproduce…",
  "state": "open",
  "url": "https://github.com/owner/repo/issues/42",
  "created_at": "2025-05-10T12:00:00Z"
}
```

---

## Security design

| Concern | Approach |
|---------|----------|
| Token storage | Read from env var at startup only; never written to disk, logs, or response bodies |
| Inbound auth | `X-API-Key` header validated in middleware before any handler runs |
| Rate limiting | Echo's built-in in-memory rate limiter (20 req/s per IP) |
| Secure headers | `X-Frame-Options`, `X-XSS-Protection`, `HSTS`, `CSP` via Echo's Secure middleware |
| Request timeouts | 10-second HTTP client timeout on all upstream GitHub calls |
| Body size | Response bodies capped at 1 MB to prevent memory exhaustion |
| `.env` excluded | `.gitignore` ensures secrets are never committed |

---

## curl examples

```bash
BASE=http://localhost:8080
KEY=your_api_key_here

# Validate repo
curl -s -H "X-API-Key: $KEY" "$BASE/api/v1/repos/octocat/Hello-World/validate" | jq

# Create issue
curl -s -X POST "$BASE/api/v1/repos/octocat/Hello-World/issues" \
  -H "X-API-Key: $KEY" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test issue","body":"Created via API","labels":["bug"]}' | jq
```

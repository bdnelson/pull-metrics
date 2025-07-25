# Pull Metrics

A Golang utility that gathers comprehensive details about GitHub Pull Requests, including metrics, timestamps, and metadata.

## Overview

Pull Metrics retrieves detailed information about a specific GitHub Pull Request by organization, repository, and PR number. It provides insights into the PR lifecycle, including review activity, commit patterns, size metrics, and release tracking.

## Features

- **Basic PR Information**: Organization, repository, PR number, author, state
- **Review Metrics**: Number of commentors, approvers, requested reviewers, and their usernames
- **Size Analysis**: Lines of code changed and number of files modified
- **Timeline Tracking**: Timestamps for key events (creation, first review request, first comment, approvals, merge, close)
- **Development Activity**: Count of commits made after the first review request
- **Release Integration**: Identifies which release (if any) includes the merged PR code
- **Generation Metadata**: Timestamp indicating when the analysis was performed

## Prerequisites

- Go 1.19 or later
- GitHub Personal Access Token with repository read permissions
- Access to the target GitHub repository

## Installation

### Building from Source

```bash
# Clone or download the source code
cd pull-metrics

# Build the utility
make build

# Or build manually
go build -buildvcs=false -o pull-metrics .
```

### Dependencies

The utility uses only Go standard library packages and requires no external dependencies.

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | Yes | GitHub Personal Access Token for API authentication |

#### Setting up GitHub Token

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate a new token with `repo` permissions (or `public_repo` for public repositories only)
3. Export the token as an environment variable:

```bash
export GITHUB_TOKEN="your_github_token_here"
```

## Usage

### Command Line Syntax

```bash
./pull-metrics <organization> <repository> <pr_number>
```

### Parameters

- `organization`: GitHub organization or username
- `repository`: Repository name
- `pr_number`: Pull Request number (integer)

### Examples

```bash
# Analyze PR #123 in the microsoft/vscode repository
./pull-metrics microsoft vscode 123

# Analyze PR #456 in a personal repository
./pull-metrics username my-project 456
```

## Output

The utility outputs detailed PR information in JSON format to STDOUT. All errors are sent to STDERR.

### JSON Schema

```json
{
  "organization_name": "string",
  "repository_name": "string", 
  "pr_number": 123,
  "author_username": "string",
  "approver_usernames": ["string"],
  "state": "string",
  "num_commentors": 0,
  "num_approvers": 0,
  "num_requested_reviewers": 0,
  "lines_changed": 0,
  "files_changed": 0,
  "commits_after_first_review": 0,
  "release_name": "string",
  "created_at": "2023-01-01T10:00:00Z",
  "first_review_request": "2023-01-01T11:00:00Z",
  "first_comment": "2023-01-01T12:00:00Z",
  "first_approval": "2023-01-01T15:00:00Z",
  "second_approval": "2023-01-01T16:00:00Z",
  "merged_at": "2023-01-01T18:00:00Z",
  "closed_at": "2023-01-01T19:00:00Z",
  "generated_at": "2025-01-25T14:30:45Z"
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `organization_name` | string | GitHub organization or username |
| `repository_name` | string | Repository name |
| `pr_number` | integer | Pull Request number |
| `author_username` | string | Username of the PR author |
| `approver_usernames` | array | List of usernames who approved the PR |
| `state` | string | PR state: "draft", "open", "merged", or "closed" |
| `num_commentors` | integer | Number of unique commentors (excluding author) |
| `num_approvers` | integer | Number of users who approved the PR |
| `num_requested_reviewers` | integer | Number of requested reviewers |
| `lines_changed` | integer | Total lines of code impacted (additions + deletions) |
| `files_changed` | integer | Number of files modified in the PR |
| `commits_after_first_review` | integer | Number of commits made after the first review request |
| `release_name` | string | Name of the release containing the merged PR (optional) |
| `created_at` | string | UTC timestamp when the PR was created (optional) |
| `first_review_request` | string | UTC timestamp of the first review request (optional) |
| `first_comment` | string | UTC timestamp of the first comment (optional) |
| `first_approval` | string | UTC timestamp of the first approval (optional) |
| `second_approval` | string | UTC timestamp of the second approval (optional) |
| `merged_at` | string | UTC timestamp when the PR was merged (optional) |
| `closed_at` | string | UTC timestamp when the PR was closed (optional) |
| `generated_at` | string | UTC timestamp when this analysis was performed |

### Optional Fields

Fields marked as "optional" are only included in the output when applicable:
- Timestamps are excluded if the corresponding event never occurred
- `release_name` is only included for merged PRs where a matching release is found

### Timestamp Format

All timestamps are in RFC3339 format in UTC timezone (e.g., `2023-01-01T12:00:00Z`).

## Development

### Building and Testing

```bash
# Build the utility
make build

# Run all tests
make test

# Run a specific test
make test-single TEST=TestFunctionName

# Format code
make fmt

# Run linting (requires golint)
make lint

# Run all checks (format, lint, test)
make check

# Clean build artifacts
make clean

# Download dependencies
make deps

# Vendor dependencies
make vendor
```

### Project Structure

```
pull-metrics/
├── main.go           # Main application code
├── main_test.go      # Unit tests
├── Makefile          # Build automation
├── README.md         # This documentation
├── CLAUDE.md         # Development instructions
└── go.mod           # Go module definition
```

### Testing

The utility includes comprehensive unit tests covering:
- PR state determination
- Approver identification and counting
- Commentor counting with author exclusion
- UTC timestamp formatting
- Timestamp extraction from PR events
- PR size calculation
- Release identification for merged PRs
- Commit counting after review requests

## Error Handling

The utility handles various error conditions gracefully:

- **Missing GitHub Token**: Exits with error message to STDERR
- **Invalid Arguments**: Shows usage information and exits
- **GitHub API Errors**: Reports API status codes and error details
- **Network Issues**: Reports connection failures
- **Invalid PR Numbers**: Reports parsing errors
- **Rate Limiting**: Returns GitHub API rate limit responses

## Rate Limiting

The utility makes multiple GitHub API calls to gather comprehensive PR data:
- PR details
- Review information
- Comments
- Timeline events
- File changes
- Commits
- Releases (for merged PRs only)

Be aware of GitHub API rate limits (5,000 requests per hour for authenticated requests).

## Examples

### Successful Output

```bash
$ ./pull-metrics microsoft vscode 12345
{
  "organization_name": "microsoft",
  "repository_name": "vscode",
  "pr_number": 12345,
  "author_username": "contributor",
  "approver_usernames": ["maintainer1", "maintainer2"],
  "state": "merged",
  "num_commentors": 3,
  "num_approvers": 2,
  "num_requested_reviewers": 2,
  "lines_changed": 245,
  "files_changed": 7,
  "commits_after_first_review": 2,
  "release_name": "v1.75.0",
  "created_at": "2023-01-15T09:30:00Z",
  "first_review_request": "2023-01-15T10:00:00Z",
  "first_comment": "2023-01-15T11:30:00Z",
  "first_approval": "2023-01-16T14:00:00Z",
  "merged_at": "2023-01-16T15:30:00Z",
  "generated_at": "2025-01-25T14:30:45Z"
}
```

### Error Output

```bash
$ ./pull-metrics microsoft vscode abc
Invalid PR number: abc

$ ./pull-metrics microsoft vscode 99999
Error fetching PR details: GitHub API returned status 404
```

## License

This utility was generated as part of a development exercise and follows the project's existing license terms.

## Contributing

This utility was developed following the specifications in `CLAUDE.md`. Any contributions should adhere to those guidelines, including:
- Writing unit tests for new functionality
- Using Go as the primary language
- Following the existing code patterns and error handling
- Ensuring JSON output format consistency
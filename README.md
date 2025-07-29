# Pull Metrics

A Golang utility that gathers comprehensive details about GitHub Pull Requests, including metrics, timestamps, and metadata.

## Overview

Pull Metrics retrieves detailed information about a specific GitHub Pull Request by organization, repository, and PR number. It provides insights into the PR lifecycle, including review activity, commit patterns, size metrics, and release tracking.

## Features

- **Basic PR Information**: Organization, repository, PR number, title, web URL, node ID, author, state, bot detection
- **Review Metrics**: Number of commenters, approvers, requested reviewers, change requests, and their usernames
- **Size Analysis**: Lines of code changed and number of files modified
- **Timeline Tracking**: Timestamps for key events (first commit, creation, first review request, first comment, approvals, merge, close)
- **Development Activity**: Count of commits made after the first review request
- **Jira Integration**: Extracts Jira issue identifiers from PR title, body, or branch name; detects bot users for automated PRs
- **Performance Metrics**: Calculated metrics for PR review process efficiency and participation
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

The utility uses Go standard library packages with minimal external dependencies:
- `github.com/google/go-github/v66` - Official Go client library for GitHub API v3
- `golang.org/x/oauth2` - Official Go OAuth2 library for GitHub API authentication
- `github.com/joho/godotenv` - For loading environment variables from .env files
- `github.com/ardanlabs/conf/v3` - For configuration management with environment variables, command line arguments, and help/usage messages

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | Yes | GitHub Personal Access Token for API authentication |

#### Setting up GitHub Token

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate a new token with `repo` permissions (or `public_repo` for public repositories only)
3. Configure the token using one of these methods:

**Option 1: Environment Variable**
```bash
export GITHUB_TOKEN="your_github_token_here"
```

**Option 2: .env File (Recommended)**
```bash
# Copy the example file
cp .env.example .env

# Edit .env and add your token
echo 'GITHUB_TOKEN=your_github_token_here' > .env
```

The utility will automatically load environment variables from a `.env` file if present, making configuration more convenient. The `.env` file is ignored by git to prevent accidental token exposure.

## Usage

### Command Line Syntax

The utility supports both positional arguments and named options:

```bash
# Using positional arguments
./pull-metrics <organization> <repository> <pr_number>

# Using named options
./pull-metrics --organization <org> --repository <repo> --pr-number <number>

# Mixed usage (positional arguments preferred)
./pull-metrics microsoft vscode --pr-number 12345
```

### Parameters

- `organization`: GitHub organization or username (positional argument 1)
- `repository`: Repository name (positional argument 2)  
- `pr_number`: Pull Request number, integer (positional argument 3)

### Help and Usage

Display help information and available options:

```bash
./pull-metrics --help
```

This shows all configuration options, including environment variable names and descriptions.

### Examples

```bash
# Analyze PR #123 in the microsoft/vscode repository (positional arguments)
./pull-metrics microsoft vscode 123

# Analyze PR #456 in a personal repository (named options)
./pull-metrics --organization username --repository my-project --pr-number 456

# Mixed usage
./pull-metrics microsoft vscode --pr-number 123

# Get help information
./pull-metrics --help
```

## Output

The utility outputs detailed PR information in JSON format to STDOUT. All errors are sent to STDERR.

### JSON Schema

```json
{
  "organization_name": "string",
  "repository_name": "string", 
  "pr_number": 123,
  "pr_title": "string",
  "pr_web_url": "string",
  "pr_node_id": "string",
  "author_username": "string",
  "approver_usernames": ["string"],
  "commenter_usernames": ["string"],
  "state": "string",
  "num_comments": 0,
  "num_commenters": 0,
  "num_approvers": 0,
  "num_requested_reviewers": 0,
  "change_requests_count": 0,
  "lines_changed": 0,
  "files_changed": 0,
  "commits_after_first_review": 0,
  "jira_issue": "string",
  "is_bot": false,
  "metrics": {
    "draft_time_hours": 2.0,
    "time_to_first_review_request_hours": 2.0,
    "time_to_first_review_hours": 1.5,
    "review_cycle_time_hours": 24.0,
    "blocking_non_blocking_ratio": 0.33,
    "reviewer_participation_ratio": 0.75
  },
  "release_name": "string",
  "timestamps": {
    "first_commit": "2023-01-01T09:00:00Z",
    "created_at": "2023-01-01T10:00:00Z",
    "first_review_request": "2023-01-01T11:00:00Z",
    "first_comment": "2023-01-01T12:00:00Z",
    "first_approval": "2023-01-01T15:00:00Z",
    "second_approval": "2023-01-01T16:00:00Z",
    "merged_at": "2023-01-01T18:00:00Z",
    "closed_at": "2023-01-01T19:00:00Z"
  },
  "generated_at": "2025-01-25T14:30:45Z"
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `organization_name` | string | GitHub organization or username |
| `repository_name` | string | Repository name |
| `pr_number` | integer | Pull Request number |
| `pr_title` | string | Pull Request title |
| `pr_web_url` | string | GitHub web URL for the Pull Request |
| `pr_node_id` | string | GitHub GraphQL node ID for the Pull Request |
| `author_username` | string | Username of the PR author |
| `approver_usernames` | array | List of usernames who approved the PR |
| `commenter_usernames` | array | List of usernames who commented on the PR from both conversation comments and review comments (excluding author), sorted alphabetically |
| `state` | string | PR state: "draft", "open", "merged", or "closed" |
| `num_comments` | integer | Total number of comments on the PR (both conversation comments and review comments) |
| `num_commenters` | integer | Number of unique commenters from both conversation comments and review comments (excluding author) |
| `num_approvers` | integer | Number of users who approved the PR |
| `num_requested_reviewers` | integer | Total number of users who were requested to review the PR (includes both those who have reviewed and those who haven't) |
| `change_requests_count` | integer | Number of reviews that requested changes |
| `lines_changed` | integer | Total lines of code impacted (additions + deletions) |
| `files_changed` | integer | Number of files modified in the PR |
| `commits_after_first_review` | integer | Number of commits made after the first review request |
| `jira_issue` | string | Jira issue identifier associated with the PR (e.g., "ABC-123"), "BOT" for bot users with no Jira issue, or "UNKNOWN" if none found |
| `is_bot` | boolean | Indicates whether the PR was created by a bot (identified by "[bot]" in username) |
| `metrics` | object | Calculated performance metrics for the PR review process (optional) |
| `release_name` | string | Name of the release containing the merged PR (optional) |
| `timestamps` | object | Collection of all timestamp information for the PR lifecycle (optional) |
| `generated_at` | string | UTC timestamp when this analysis was performed |

### Timestamps Object

The `timestamps` object contains all timestamp information related to the PR lifecycle:

| Field | Type | Description |
|-------|------|-------------|
| `first_commit` | string | UTC timestamp of the first commit in the PR branch (optional) |
| `created_at` | string | UTC timestamp when the PR was created (optional) |
| `first_review_request` | string | UTC timestamp of the first review request (optional) |
| `first_comment` | string | UTC timestamp of the first comment from either conversation comments or review comments (optional) |
| `first_approval` | string | UTC timestamp of the first approval (optional) |
| `second_approval` | string | UTC timestamp of the second approval (optional) |
| `merged_at` | string | UTC timestamp when the PR was merged (optional) |
| `closed_at` | string | UTC timestamp when the PR was closed (optional) |

### Metrics Object

The `metrics` object contains calculated performance indicators for the PR review process:

| Field | Type | Description |
|-------|------|-------------|
| `draft_time_hours` | float | Hours from PR creation to first review request, minimum 0.0 |
| `time_to_first_review_request_hours` | float | Hours from PR creation to first review request (optional) |
| `time_to_first_review_hours` | float | Hours from first review request to first comment (conversation or review comment) or first approval, whichever comes first (optional) |
| `review_cycle_time_hours` | float | Hours from first review request to PR resolution (merge/close) (optional) |
| `blocking_non_blocking_ratio` | float | Ratio of blocking (CHANGES_REQUESTED) to non-blocking (APPROVED/COMMENTED) reviews (optional) |
| `reviewer_participation_ratio` | float | Ratio of actual reviewers to requested reviewers (optional) |

**Metrics Calculation Details:**
- **Draft Time**: Always included, minimum 0.0. Calculated as hours from PR creation to first review request when both timestamps are available and review request occurs after creation
- **Time to First Review Request**: Only calculated if first review request occurs after PR creation
- **Time to First Review**: Only calculated if first review activity (conversation comment, review comment, or approval) occurs after first review request
- **Review Cycle Time**: Uses merge time if available, otherwise close time
- **Blocking Ratio**: Only calculated if there are non-blocking reviews (avoids division by zero)
- **Participation Ratio**: Only calculated if reviewers were requested

### Optional Fields

Fields marked as "optional" are only included in the output when applicable:
- `timestamps` object is included when any timestamp information is available; individual timestamp fields within the object are excluded if the corresponding event never occurred
- `release_name` is only included for merged PRs where a matching release is found
- `metrics` object is excluded if no calculable metrics are available
- Individual metric fields are excluded if calculation requirements are not met

### Timestamp Format

All timestamps are in RFC3339 format in UTC timezone (e.g., `2023-01-01T12:00:00Z`).

### Jira Issue Extraction

The utility automatically extracts Jira issue identifiers from PRs using the following logic:

1. **Search Order**: Searches for Jira patterns in this priority order:
   - PR title
   - PR body (if available)
   - Branch name (head ref)

2. **Pattern Matching**: Matches standard Jira issue formats like `PROJECT-123`, `ABC-1234`, etc.
   - Project key: 2+ characters (uppercase letters or alphanumeric)
   - Followed by hyphen and number
   - **Excludes CVE identifiers**: Security vulnerability identifiers starting with `CVE-` (e.g., `CVE-2023-1234`) are excluded as they are not Jira issues

3. **Special Cases**:
   - **Bot Detection**: If no Jira issue is found and the PR author's username contains `[bot]`, returns `"BOT"` (the `is_bot` field is also set to `true`)
   - **Not Found**: If no Jira issue is found and the author is not a bot, returns `"UNKNOWN"` (the `is_bot` field is set to `false`)

**Examples**:
- `dependabot[bot]` creating a dependency update → `jira_issue: "BOT"`, `is_bot: true`
- `github-actions[bot]` creating an automated PR → `jira_issue: "BOT"`, `is_bot: true`
- Regular user with no Jira issue → `jira_issue: "UNKNOWN"`, `is_bot: false`
- Security fix with `CVE-2023-1234` but no Jira issue → `jira_issue: "UNKNOWN"`, `is_bot: false`
- PR with both `SECURITY-123` and `CVE-2023-1234` → `jira_issue: "SECURITY-123"`, `is_bot: false`

### Requested Reviewers Counting

The `num_requested_reviewers` field provides a comprehensive count of all users who were asked to review the PR:

1. **Includes Reviewers Who Have Reviewed**: Users who have submitted any type of review (approved, requested changes, or commented) are counted because they must have been requested to review
2. **Includes Pending Reviewers**: Users who are currently in the "requested reviewers" list but haven't reviewed yet
3. **Deduplication**: If a user appears in both categories (reviewed and still pending), they are counted only once
4. **Multiple Reviews**: Users who submitted multiple reviews are counted only once

**Rationale**: GitHub removes users from the `requested_reviewers` list once they submit a review, so the raw count would underestimate engagement. This comprehensive approach provides the true scope of review requests.

**Examples**:
- PR with 3 requested reviewers: 2 have reviewed, 1 is pending → `num_requested_reviewers: 3`
- PR where reviewer submitted multiple reviews → counted once in `num_requested_reviewers`
- PR where all requested reviewers have reviewed → `num_requested_reviewers` equals total unique reviewers

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
├── go.mod            # Go module definition
├── .env.example      # Example environment configuration
└── vendor/           # Vendored dependencies
```

### Testing

The utility includes comprehensive unit tests covering:
- PR state determination
- Approver identification and counting
- Commenter counting with author exclusion (including both conversation and review comments)
- Comment counting (total comment count from both sources)
- Comprehensive requested reviewers counting (includes both reviewed and pending reviewers)
- Commenter username extraction and sorting
- UTC timestamp formatting
- Timestamp extraction from PR events (including review comments for first comment detection)
- PR size calculation
- Release identification for merged PRs
- Commit counting after review requests
- Jira issue extraction and identification
- PR performance metrics calculations
- Review comments integration and processing
- JSON output structure validation for new fields

## Error Handling

The utility handles various error conditions gracefully:

- **Missing GitHub Token**: Exits with error message to STDERR
- **Invalid Arguments**: Shows detailed configuration error messages and exits
- **Invalid PR Numbers**: Reports type conversion errors with specific details  
- **Missing Required Arguments**: Automatic help display with `--help` flag
- **GitHub API Errors**: Reports API status codes and error details
- **Network Issues**: Reports connection failures
- **Rate Limiting**: Returns GitHub API rate limit responses

## Rate Limiting and Pagination

The utility makes multiple GitHub API calls to gather comprehensive PR data:
- PR details
- Review information
- Comments (conversation comments)
- Review comments (inline code comments)
- Timeline events
- File changes
- Commits
- Releases (for merged PRs only)

### Pagination Handling

All GitHub API endpoints that support pagination are handled automatically. The utility:
- Fetches up to 100 items per page (GitHub's maximum)
- Automatically follows pagination links to collect all data
- Ensures complete data collection for large PRs with many:
  - Reviews
  - Comments and review comments
  - Timeline events
  - Changed files
  - Commits
  - Repository releases

### Rate Limiting

Be aware of GitHub API rate limits (5,000 requests per hour for authenticated requests). Large PRs with extensive pagination may consume more API quota. The number of API calls depends on:
- Base calls: 7-8 API calls per PR (fixed)
- Pagination calls: Additional calls for every 100 items in paginated responses

For example, a PR with 250 comments would require 3 pagination calls for comments (100 + 100 + 50 items).

## Examples

### Successful Output

```bash
$ ./pull-metrics microsoft vscode 12345
{
  "organization_name": "microsoft",
  "repository_name": "vscode",
  "pr_number": 12345,
  "pr_title": "Fix authentication timeout issue",
  "pr_web_url": "https://github.com/microsoft/vscode/pull/12345",
  "pr_node_id": "PR_kwDOABCD123_node456",
  "author_username": "contributor",
  "approver_usernames": ["maintainer1", "maintainer2"],
  "commenter_usernames": ["reviewer1", "reviewer2", "user1"],
  "state": "merged",
  "num_comments": 12,
  "num_commenters": 3,
  "num_approvers": 2,
  "num_requested_reviewers": 2,
  "change_requests_count": 1,
  "lines_changed": 245,
  "files_changed": 7,
  "commits_after_first_review": 2,
  "jira_issue": "VSCODE-123",
  "is_bot": false,
  "metrics": {
    "draft_time_hours": 0.5,
    "time_to_first_review_request_hours": 0.5,
    "time_to_first_review_hours": 2.5,
    "review_cycle_time_hours": 25.5,
    "blocking_non_blocking_ratio": 0.5,
    "reviewer_participation_ratio": 1.0
  },
  "release_name": "v1.75.0",
  "timestamps": {
    "first_commit": "2023-01-15T09:00:00Z",
    "created_at": "2023-01-15T09:30:00Z",
    "first_review_request": "2023-01-15T10:00:00Z",
    "first_comment": "2023-01-15T11:30:00Z",
    "first_approval": "2023-01-16T14:00:00Z",
    "merged_at": "2023-01-16T15:30:00Z"
  },
  "generated_at": "2025-01-25T14:30:45Z"
}
```

### Error Output

```bash
$ ./pull-metrics --pr-number abc
Error parsing configuration: parsing config: conf: error assigning to field PRNumber: converting 'abc' to type int. details: strconv.ParseInt: parsing "abc": invalid syntax

$ ./pull-metrics microsoft vscode 99999
Error fetching PR details: GitHub API returned status 404

$ ./pull-metrics --help
Usage: pull-metrics [options...] [arguments...]

OPTIONS
      --git-hub-token  <string>    GitHub Personal Access Token
  -h, --help                       display this help message
      --organization   <string>    GitHub organization or username
      --pr-number      <int>       Pull Request number
      --repository     <string>    Repository name

ENVIRONMENT
  GITHUB_TOKEN  <string>    GitHub Personal Access Token
  ORGANIZATION  <string>    GitHub organization or username
  PR_NUMBER     <int>       Pull Request number
  REPOSITORY    <string>    Repository name
```

## License

See the [LICENSE](LICENSE.txt) file for license rights and limitations (MIT).

## Contributing

This utility was developed following the specifications in `CLAUDE.md`. Any contributions should adhere to those guidelines, including:
- Writing unit tests for new functionality
- Using Go as the primary language
- Following the existing code patterns and error handling
- Ensuring JSON output format consistency
- All tests must pass before merging

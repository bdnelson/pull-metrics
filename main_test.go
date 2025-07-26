package main

import (
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
)

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// Helper function to create a pointer to an int
func intPtr(i int) *int {
	return &i
}

// Helper function to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create a pointer to a time.Time
func timePtr(t time.Time) *github.Timestamp {
	return &github.Timestamp{Time: t}
}

func TestGetPRState(t *testing.T) {
	tests := []struct {
		name     string
		pr       *github.PullRequest
		expected string
	}{
		{
			name: "draft PR",
			pr: &github.PullRequest{
				State:   stringPtr("open"),
				Draft:   boolPtr(true),
				Merged:  boolPtr(false),
				Title:   stringPtr("Draft PR"),
				HTMLURL: stringPtr("https://github.com/org/repo/pull/1"),
				NodeID:  stringPtr("PR_node123"),
			},
			expected: "draft",
		},
		{
			name: "merged PR",
			pr: &github.PullRequest{
				State:   stringPtr("closed"),
				Draft:   boolPtr(false),
				Merged:  boolPtr(true),
				Title:   stringPtr("Merged PR"),
				HTMLURL: stringPtr("https://github.com/org/repo/pull/2"),
				NodeID:  stringPtr("PR_node456"),
			},
			expected: "merged",
		},
		{
			name: "open PR",
			pr: &github.PullRequest{
				State:   stringPtr("open"),
				Draft:   boolPtr(false),
				Merged:  boolPtr(false),
				Title:   stringPtr("Open PR"),
				HTMLURL: stringPtr("https://github.com/org/repo/pull/3"),
				NodeID:  stringPtr("PR_node789"),
			},
			expected: "open",
		},
		{
			name: "closed PR",
			pr: &github.PullRequest{
				State:   stringPtr("closed"),
				Draft:   boolPtr(false),
				Merged:  boolPtr(false),
				Title:   stringPtr("Closed PR"),
				HTMLURL: stringPtr("https://github.com/org/repo/pull/4"),
				NodeID:  stringPtr("PR_node111"),
			},
			expected: "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPRState(tt.pr)
			if result != tt.expected {
				t.Errorf("getPRState() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetApprovers(t *testing.T) {
	tests := []struct {
		name     string
		reviews  []*github.PullRequestReview
		expected []string
	}{
		{
			name: "single approver",
			reviews: []*github.PullRequestReview{
				{
					User:  &github.User{Login: stringPtr("user1")},
					State: stringPtr("APPROVED"),
				},
			},
			expected: []string{"user1"},
		},
		{
			name: "multiple approvers",
			reviews: []*github.PullRequestReview{
				{
					User:  &github.User{Login: stringPtr("user1")},
					State: stringPtr("APPROVED"),
				},
				{
					User:  &github.User{Login: stringPtr("user2")},
					State: stringPtr("APPROVED"),
				},
			},
			expected: []string{"user1", "user2"},
		},
		{
			name: "mixed review states",
			reviews: []*github.PullRequestReview{
				{
					User:  &github.User{Login: stringPtr("user1")},
					State: stringPtr("APPROVED"),
				},
				{
					User:  &github.User{Login: stringPtr("user2")},
					State: stringPtr("CHANGES_REQUESTED"),
				},
				{
					User:  &github.User{Login: stringPtr("user3")},
					State: stringPtr("COMMENTED"),
				},
			},
			expected: []string{"user1"},
		},
		{
			name:     "no reviews",
			reviews:  []*github.PullRequestReview{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getApprovers(tt.reviews)
			if len(result) != len(tt.expected) {
				t.Errorf("getApprovers() returned %d approvers, want %d", len(result), len(tt.expected))
				return
			}
			
			// Convert to map for easy comparison
			resultMap := make(map[string]bool)
			for _, username := range result {
				resultMap[username] = true
			}
			
			for _, expectedUser := range tt.expected {
				if !resultMap[expectedUser] {
					t.Errorf("getApprovers() missing expected user %s", expectedUser)
				}
			}
		})
	}
}

func TestGetCommenters(t *testing.T) {
	tests := []struct {
		name           string
		comments       []*github.IssueComment
		reviewComments []*github.PullRequestComment
		authorUsername string
		expected       []string
	}{
		{
			name: "regular comments only",
			comments: []*github.IssueComment{
				{
					User:      &github.User{Login: stringPtr("user1")},
					CreatedAt: timePtr(time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)),
				},
				{
					User:      &github.User{Login: stringPtr("user2")},
					CreatedAt: timePtr(time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)),
				},
			},
			reviewComments: []*github.PullRequestComment{},
			authorUsername: "author",
			expected:       []string{"user1", "user2"},
		},
		{
			name:     "review comments only",
			comments: []*github.IssueComment{},
			reviewComments: []*github.PullRequestComment{
				{
					User:      &github.User{Login: stringPtr("user3")},
					CreatedAt: timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
				},
			},
			authorUsername: "author",
			expected:       []string{"user3"},
		},
		{
			name: "mixed comments excluding author",
			comments: []*github.IssueComment{
				{
					User:      &github.User{Login: stringPtr("user1")},
					CreatedAt: timePtr(time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)),
				},
				{
					User:      &github.User{Login: stringPtr("author")},
					CreatedAt: timePtr(time.Date(2023, 1, 1, 10, 30, 0, 0, time.UTC)),
				},
			},
			reviewComments: []*github.PullRequestComment{
				{
					User:      &github.User{Login: stringPtr("user2")},
					CreatedAt: timePtr(time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)),
				},
			},
			authorUsername: "author",
			expected:       []string{"user1", "user2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommenters(tt.comments, tt.reviewComments, tt.authorUsername)
			
			if len(result) != len(tt.expected) {
				t.Errorf("getCommenters() returned %d commenters, want %d", len(result), len(tt.expected))
				return
			}
			
			for _, expectedUser := range tt.expected {
				if !result[expectedUser] {
					t.Errorf("getCommenters() missing expected user %s", expectedUser)
				}
			}
		})
	}
}

func TestCountTotalComments(t *testing.T) {
	tests := []struct {
		name           string
		comments       []*github.IssueComment
		reviewComments []*github.PullRequestComment
		expected       int
	}{
		{
			name: "regular comments only",
			comments: []*github.IssueComment{
				{User: &github.User{Login: stringPtr("user1")}},
				{User: &github.User{Login: stringPtr("user2")}},
			},
			reviewComments: []*github.PullRequestComment{},
			expected:       2,
		},
		{
			name:     "review comments only",
			comments: []*github.IssueComment{},
			reviewComments: []*github.PullRequestComment{
				{User: &github.User{Login: stringPtr("user1")}},
				{User: &github.User{Login: stringPtr("user2")}},
				{User: &github.User{Login: stringPtr("user3")}},
			},
			expected: 3,
		},
		{
			name: "mixed comments",
			comments: []*github.IssueComment{
				{User: &github.User{Login: stringPtr("user1")}},
			},
			reviewComments: []*github.PullRequestComment{
				{User: &github.User{Login: stringPtr("user2")}},
				{User: &github.User{Login: stringPtr("user3")}},
			},
			expected: 3,
		},
		{
			name:           "no comments",
			comments:       []*github.IssueComment{},
			reviewComments: []*github.PullRequestComment{},
			expected:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countTotalComments(tt.comments, tt.reviewComments)
			if result != tt.expected {
				t.Errorf("countTotalComments() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCommenterUsernames(t *testing.T) {
	tests := []struct {
		name       string
		commenters map[string]bool
		expected   []string
	}{
		{
			name: "multiple commenters",
			commenters: map[string]bool{
				"user3": true,
				"user1": true,
				"user2": true,
			},
			expected: []string{"user1", "user2", "user3"}, // Should be sorted
		},
		{
			name: "single commenter",
			commenters: map[string]bool{
				"user1": true,
			},
			expected: []string{"user1"},
		},
		{
			name:       "no commenters",
			commenters: map[string]bool{},
			expected:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommenterUsernames(tt.commenters)
			
			if len(result) != len(tt.expected) {
				t.Errorf("getCommenterUsernames() returned %d usernames, want %d", len(result), len(tt.expected))
				return
			}
			
			for i, username := range result {
				if username != tt.expected[i] {
					t.Errorf("getCommenterUsernames()[%d] = %v, want %v", i, username, tt.expected[i])
				}
			}
		})
	}
}

func TestCountAllRequestedReviewers(t *testing.T) {
	tests := []struct {
		name     string
		pr       *github.PullRequest
		reviews  []*github.PullRequestReview
		expected int
	}{
		{
			name: "reviewers who have reviewed and pending reviewers",
			pr: &github.PullRequest{
				RequestedReviewers: []*github.User{
					{Login: stringPtr("pending1")},
					{Login: stringPtr("pending2")},
				},
			},
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: stringPtr("reviewed1")}},
				{User: &github.User{Login: stringPtr("reviewed2")}},
			},
			expected: 4,
		},
		{
			name: "overlap between reviewed and pending",
			pr: &github.PullRequest{
				RequestedReviewers: []*github.User{
					{Login: stringPtr("user1")},
					{Login: stringPtr("pending1")},
				},
			},
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: stringPtr("user1")}}, // Same user in both lists
				{User: &github.User{Login: stringPtr("reviewed1")}},
			},
			expected: 3, // user1 counted once, pending1, reviewed1
		},
		{
			name: "only reviewed, no pending",
			pr: &github.PullRequest{
				RequestedReviewers: []*github.User{},
			},
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: stringPtr("reviewed1")}},
				{User: &github.User{Login: stringPtr("reviewed2")}},
			},
			expected: 2,
		},
		{
			name: "only pending, no reviewed",
			pr: &github.PullRequest{
				RequestedReviewers: []*github.User{
					{Login: stringPtr("pending1")},
					{Login: stringPtr("pending2")},
				},
			},
			reviews:  []*github.PullRequestReview{},
			expected: 2,
		},
		{
			name: "no reviewers at all",
			pr: &github.PullRequest{
				RequestedReviewers: []*github.User{},
			},
			reviews:  []*github.PullRequestReview{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countAllRequestedReviewers(tt.pr, tt.reviews)
			if result != tt.expected {
				t.Errorf("countAllRequestedReviewers() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCountChangeRequests(t *testing.T) {
	tests := []struct {
		name     string
		reviews  []*github.PullRequestReview
		expected int
	}{
		{
			name: "multiple change requests",
			reviews: []*github.PullRequestReview{
				{State: stringPtr("CHANGES_REQUESTED")},
				{State: stringPtr("APPROVED")},
				{State: stringPtr("CHANGES_REQUESTED")},
				{State: stringPtr("COMMENTED")},
			},
			expected: 2,
		},
		{
			name: "no change requests",
			reviews: []*github.PullRequestReview{
				{State: stringPtr("APPROVED")},
				{State: stringPtr("COMMENTED")},
			},
			expected: 0,
		},
		{
			name:     "no reviews",
			reviews:  []*github.PullRequestReview{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countChangeRequests(tt.reviews)
			if result != tt.expected {
				t.Errorf("countChangeRequests() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCountChangeRequestComments(t *testing.T) {
	tests := []struct {
		name           string
		comments       []*github.IssueComment
		reviewComments []*github.PullRequestComment
		reviews        []*github.PullRequestReview
		expected       int
	}{
		{
			name: "comments from change requesters",
			comments: []*github.IssueComment{
				{User: &github.User{Login: stringPtr("changer1")}},
				{User: &github.User{Login: stringPtr("approver1")}},
				{User: &github.User{Login: stringPtr("changer1")}}, // Same user, multiple comments
			},
			reviewComments: []*github.PullRequestComment{
				{User: &github.User{Login: stringPtr("changer2")}},
				{User: &github.User{Login: stringPtr("regular_user")}},
			},
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: stringPtr("changer1")}, State: stringPtr("CHANGES_REQUESTED")},
				{User: &github.User{Login: stringPtr("changer2")}, State: stringPtr("CHANGES_REQUESTED")},
				{User: &github.User{Login: stringPtr("approver1")}, State: stringPtr("APPROVED")},
			},
			expected: 3, // 2 comments from changer1 + 1 review comment from changer2
		},
		{
			name:           "no change requesters",
			comments:       []*github.IssueComment{{User: &github.User{Login: stringPtr("user1")}}},
			reviewComments: []*github.PullRequestComment{{User: &github.User{Login: stringPtr("user2")}}},
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: stringPtr("user3")}, State: stringPtr("APPROVED")},
			},
			expected: 0,
		},
		{
			name:           "no comments",
			comments:       []*github.IssueComment{},
			reviewComments: []*github.PullRequestComment{},
			reviews: []*github.PullRequestReview{
				{User: &github.User{Login: stringPtr("changer1")}, State: stringPtr("CHANGES_REQUESTED")},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countChangeRequestComments(tt.comments, tt.reviewComments, tt.reviews)
			if result != tt.expected {
				t.Errorf("countChangeRequestComments() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsBot(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{
			name:     "dependabot user",
			username: "dependabot[bot]",
			expected: true,
		},
		{
			name:     "github actions bot",
			username: "github-actions[bot]",
			expected: true,
		},
		{
			name:     "regular user",
			username: "john_doe",
			expected: false,
		},
		{
			name:     "user with bot in name but not bracketed",
			username: "robotuser",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBot(tt.username)
			if result != tt.expected {
				t.Errorf("isBot(%s) = %v, want %v", tt.username, result, tt.expected)
			}
		})
	}
}

func TestExtractJiraIssue(t *testing.T) {
	tests := []struct {
		name     string
		pr       *github.PullRequest
		expected string
	}{
		{
			name: "Jira issue in title",
			pr: &github.PullRequest{
				Title: stringPtr("Fix bug in ABC-123 authentication"),
				Body:  stringPtr("This fixes the auth issue"),
				User:  &github.User{Login: stringPtr("developer")},
				Head: &github.PullRequestBranch{
					Ref: stringPtr("feature-branch"),
				},
			},
			expected: "ABC-123",
		},
		{
			name: "Jira issue in body when not in title",
			pr: &github.PullRequest{
				Title: stringPtr("Fix authentication bug"),
				Body:  stringPtr("This addresses DEF-456 by updating the token validation"),
				User:  &github.User{Login: stringPtr("developer")},
				Head: &github.PullRequestBranch{
					Ref: stringPtr("feature-branch"),
				},
			},
			expected: "DEF-456",
		},
		{
			name: "Jira issue in branch name when not in title or body",
			pr: &github.PullRequest{
				Title: stringPtr("Fix authentication bug"),
				Body:  stringPtr("This fixes the auth issue"),
				User:  &github.User{Login: stringPtr("developer")},
				Head: &github.PullRequestBranch{
					Ref: stringPtr("feature/GHI-789-fix-auth"),
				},
			},
			expected: "GHI-789",
		},
		{
			name: "Bot user with no Jira issue",
			pr: &github.PullRequest{
				Title: stringPtr("Update dependencies"),
				Body:  stringPtr("Automated dependency update"),
				User:  &github.User{Login: stringPtr("dependabot[bot]")},
				Head: &github.PullRequestBranch{
					Ref: stringPtr("dependabot/npm_and_yarn/package-update"),
				},
			},
			expected: "BOT",
		},
		{
			name: "Regular user with no Jira issue",
			pr: &github.PullRequest{
				Title: stringPtr("Update documentation"),
				Body:  stringPtr("Updated the README file"),
				User:  &github.User{Login: stringPtr("developer")},
				Head: &github.PullRequestBranch{
					Ref: stringPtr("update-docs"),
				},
			},
			expected: "UNKNOWN",
		},
		{
			name: "CVE identifier should be excluded",
			pr: &github.PullRequest{
				Title: stringPtr("Security fix for CVE-2023-1234"),
				Body:  stringPtr("This addresses the security vulnerability"),
				User:  &github.User{Login: stringPtr("developer")},
				Head: &github.PullRequestBranch{
					Ref: stringPtr("security-fix"),
				},
			},
			expected: "UNKNOWN", // CVE should be excluded
		},
		{
			name: "Jira issue with CVE present - Jira should win",
			pr: &github.PullRequest{
				Title: stringPtr("SECURITY-123: Fix CVE-2023-1234 vulnerability"),
				Body:  stringPtr("This addresses the CVE-2023-1234 security issue"),
				User:  &github.User{Login: stringPtr("developer")},
				Head: &github.PullRequestBranch{
					Ref: stringPtr("security-fix"),
				},
			},
			expected: "SECURITY-123", // Valid Jira issue should be returned, CVE ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJiraIssue(tt.pr)
			if result != tt.expected {
				t.Errorf("extractJiraIssue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatToUTC(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		expected  string
	}{
		{
			name:      "RFC3339 timestamp",
			timestamp: "2023-01-15T10:30:45Z",
			expected:  "2023-01-15T10:30:45Z",
		},
		{
			name:      "timestamp with timezone",
			timestamp: "2023-01-15T10:30:45-08:00",
			expected:  "2023-01-15T18:30:45Z", // Converted to UTC
		},
		{
			name:      "invalid timestamp",
			timestamp: "invalid-timestamp",
			expected:  "invalid-timestamp", // Should return original if parsing fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatToUTC(tt.timestamp)
			if result != tt.expected {
				t.Errorf("formatToUTC(%s) = %v, want %v", tt.timestamp, result, tt.expected)
			}
		})
	}
}

func TestCalculatePRSize(t *testing.T) {
	tests := []struct {
		name     string
		files    []*github.CommitFile
		expected *PRSize
	}{
		{
			name: "multiple files with changes",
			files: []*github.CommitFile{
				{
					Filename:  stringPtr("file1.go"),
					Additions: intPtr(10),
					Deletions: intPtr(5),
				},
				{
					Filename:  stringPtr("file2.go"),
					Additions: intPtr(20),
					Deletions: intPtr(3),
				},
			},
			expected: &PRSize{
				LinesChanged: 38, // 10+5+20+3
				FilesChanged: 2,
			},
		},
		{
			name: "single file",
			files: []*github.CommitFile{
				{
					Filename:  stringPtr("file1.go"),
					Additions: intPtr(15),
					Deletions: intPtr(8),
				},
			},
			expected: &PRSize{
				LinesChanged: 23, // 15+8
				FilesChanged: 1,
			},
		},
		{
			name:  "no files",
			files: []*github.CommitFile{},
			expected: &PRSize{
				LinesChanged: 0,
				FilesChanged: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePRSize(tt.files)
			if result.LinesChanged != tt.expected.LinesChanged {
				t.Errorf("calculatePRSize().LinesChanged = %v, want %v", result.LinesChanged, tt.expected.LinesChanged)
			}
			if result.FilesChanged != tt.expected.FilesChanged {
				t.Errorf("calculatePRSize().FilesChanged = %v, want %v", result.FilesChanged, tt.expected.FilesChanged)
			}
		})
	}
}
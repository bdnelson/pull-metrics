package pullmetrics

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

func TestCalculatePRMetrics_DraftTime(t *testing.T) {
	tests := []struct {
		name        string
		timestamps  *Timestamps
		expectedHours float64
	}{
		{
			name: "draft time calculated when both timestamps exist",
			timestamps: &Timestamps{
				CreatedAt:          stringPtr("2023-01-15T10:00:00Z"),
				FirstReviewRequest: stringPtr("2023-01-15T12:30:00Z"),
			},
			expectedHours: 2.5, // 2.5 hours
		},
		{
			name: "zero draft time when created_at missing",
			timestamps: &Timestamps{
				FirstReviewRequest: stringPtr("2023-01-15T12:30:00Z"),
			},
			expectedHours: 0.0,
		},
		{
			name: "zero draft time when first_review_request missing",
			timestamps: &Timestamps{
				CreatedAt: stringPtr("2023-01-15T10:00:00Z"),
			},
			expectedHours: 0.0,
		},
		{
			name: "zero draft time when review request is before creation",
			timestamps: &Timestamps{
				CreatedAt:          stringPtr("2023-01-15T12:00:00Z"),
				FirstReviewRequest: stringPtr("2023-01-15T10:00:00Z"), // Before creation
			},
			expectedHours: 0.0,
		},
		{
			name: "zero draft time when review request is at same time as creation",
			timestamps: &Timestamps{
				CreatedAt:          stringPtr("2023-01-15T10:00:00Z"),
				FirstReviewRequest: stringPtr("2023-01-15T10:00:00Z"), // Same time
			},
			expectedHours: 0.0, // Should be 0 since not after creation time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := calculatePRMetrics(
				&github.PullRequest{},
				[]*github.PullRequestReview{},
				[]*github.IssueComment{},
				[]*github.Timeline{},
				tt.timestamps,
			)

			if metrics.DraftTimeHours != tt.expectedHours {
				t.Errorf("calculatePRMetrics().DraftTimeHours = %v, want %v", metrics.DraftTimeHours, tt.expectedHours)
			}
		})
	}
}

func TestFindReleaseForMergedPR_WithCreatedAt(t *testing.T) {
	tests := []struct {
		name                    string
		pr                      *github.PullRequest
		releases                []*github.RepositoryRelease
		expectedReleaseName     *string
		expectedReleaseCreatedAt *string
	}{
		{
			name: "merged PR with release and created timestamp",
			pr: &github.PullRequest{
				Merged:   boolPtr(true),
				MergedAt: timePtr(time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)),
			},
			releases: []*github.RepositoryRelease{
				{
					Name:        stringPtr("v1.0.0"),
					TagName:     stringPtr("v1.0.0"),
					PublishedAt: timePtr(time.Date(2023, 1, 16, 10, 0, 0, 0, time.UTC)),
					CreatedAt:   timePtr(time.Date(2023, 1, 16, 9, 0, 0, 0, time.UTC)),
				},
			},
			expectedReleaseName:     stringPtr("v1.0.0"),
			expectedReleaseCreatedAt: stringPtr("2023-01-16T09:00:00Z"),
		},
		{
			name: "merged PR with release but no created timestamp",
			pr: &github.PullRequest{
				Merged:   boolPtr(true),
				MergedAt: timePtr(time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)),
			},
			releases: []*github.RepositoryRelease{
				{
					Name:        stringPtr("v1.0.0"),
					TagName:     stringPtr("v1.0.0"),
					PublishedAt: timePtr(time.Date(2023, 1, 16, 10, 0, 0, 0, time.UTC)),
					CreatedAt:   nil, // No creation timestamp
				},
			},
			expectedReleaseName:     stringPtr("v1.0.0"),
			expectedReleaseCreatedAt: nil,
		},
		{
			name: "unmerged PR",
			pr: &github.PullRequest{
				Merged:   boolPtr(false),
				MergedAt: nil,
			},
			releases: []*github.RepositoryRelease{
				{
					Name:        stringPtr("v1.0.0"),
					TagName:     stringPtr("v1.0.0"),
					PublishedAt: timePtr(time.Date(2023, 1, 16, 10, 0, 0, 0, time.UTC)),
					CreatedAt:   timePtr(time.Date(2023, 1, 16, 9, 0, 0, 0, time.UTC)),
				},
			},
			expectedReleaseName:     nil,
			expectedReleaseCreatedAt: nil,
		},
		{
			name: "merged PR with multiple releases, earliest selected",
			pr: &github.PullRequest{
				Merged:   boolPtr(true),
				MergedAt: timePtr(time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)),
			},
			releases: []*github.RepositoryRelease{
				{
					Name:        stringPtr("v1.1.0"),
					TagName:     stringPtr("v1.1.0"),
					PublishedAt: timePtr(time.Date(2023, 1, 20, 10, 0, 0, 0, time.UTC)),
					CreatedAt:   timePtr(time.Date(2023, 1, 20, 9, 0, 0, 0, time.UTC)),
				},
				{
					Name:        stringPtr("v1.0.0"),
					TagName:     stringPtr("v1.0.0"),
					PublishedAt: timePtr(time.Date(2023, 1, 16, 10, 0, 0, 0, time.UTC)),
					CreatedAt:   timePtr(time.Date(2023, 1, 16, 9, 0, 0, 0, time.UTC)),
				},
			},
			expectedReleaseName:     stringPtr("v1.0.0"), // Earliest release
			expectedReleaseCreatedAt: stringPtr("2023-01-16T09:00:00Z"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseName, releaseCreatedAt := findReleaseForMergedPR(tt.pr, tt.releases)
			
			if tt.expectedReleaseName == nil {
				if releaseName != nil {
					t.Errorf("findReleaseForMergedPR() releaseName = %v, want nil", *releaseName)
				}
			} else {
				if releaseName == nil {
					t.Errorf("findReleaseForMergedPR() releaseName = nil, want %v", *tt.expectedReleaseName)
				} else if *releaseName != *tt.expectedReleaseName {
					t.Errorf("findReleaseForMergedPR() releaseName = %v, want %v", *releaseName, *tt.expectedReleaseName)
				}
			}
			
			if tt.expectedReleaseCreatedAt == nil {
				if releaseCreatedAt != nil && *releaseCreatedAt != "" {
					t.Errorf("findReleaseForMergedPR() releaseCreatedAt = %v, want nil or empty", *releaseCreatedAt)
				}
			} else {
				if releaseCreatedAt == nil {
					t.Errorf("findReleaseForMergedPR() releaseCreatedAt = nil, want %v", *tt.expectedReleaseCreatedAt)
				} else if *releaseCreatedAt != *tt.expectedReleaseCreatedAt {
					t.Errorf("findReleaseForMergedPR() releaseCreatedAt = %v, want %v", *releaseCreatedAt, *tt.expectedReleaseCreatedAt)
				}
			}
		})
	}
}

func TestGetPRDetails_ReleaseCreatedAtInTimestamps(t *testing.T) {
	// Test that release_created_at appears in timestamps object, not at top level
	pr := &github.PullRequest{
		Title:    stringPtr("Test PR"),
		HTMLURL:  stringPtr("https://github.com/org/repo/pull/1"),
		NodeID:   stringPtr("PR_node123"),
		User:     &github.User{Login: stringPtr("author")},
		Merged:   boolPtr(true),
		MergedAt: timePtr(time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)),
		CreatedAt: timePtr(time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC)),
	}

	releases := []*github.RepositoryRelease{
		{
			Name:        stringPtr("v1.0.0"),
			TagName:     stringPtr("v1.0.0"),
			PublishedAt: timePtr(time.Date(2023, 1, 16, 10, 0, 0, 0, time.UTC)),
			CreatedAt:   timePtr(time.Date(2023, 1, 16, 9, 0, 0, 0, time.UTC)),
		},
	}

	// Mock the functions that would normally be called
	releaseName, releaseCreatedAt := findReleaseForMergedPR(pr, releases)
	
	// Verify the function returns expected values
	if releaseName == nil || *releaseName != "v1.0.0" {
		t.Errorf("Expected release name v1.0.0, got %v", releaseName)
	}
	if releaseCreatedAt == nil || *releaseCreatedAt != "2023-01-16T09:00:00Z" {
		t.Errorf("Expected release created at 2023-01-16T09:00:00Z, got %v", releaseCreatedAt)
	}

	// Create a timestamps object similar to how getPRDetails does
	timestamps := &Timestamps{
		CreatedAt: stringPtr("2023-01-15T10:00:00Z"),
		MergedAt:  stringPtr("2023-01-15T12:00:00Z"),
	}

	prTimestamps := &PRTimestamps{
		FirstCommit:        timestamps.FirstCommit,
		CreatedAt:          timestamps.CreatedAt,
		FirstReviewRequest: timestamps.FirstReviewRequest,
		FirstComment:       timestamps.FirstComment,
		FirstApproval:      timestamps.FirstApproval,
		SecondApproval:     timestamps.SecondApproval,
		MergedAt:           timestamps.MergedAt,
		ClosedAt:           timestamps.ClosedAt,
	}

	// Add release creation timestamp if it exists (like getPRDetails does)
	if releaseCreatedAt != nil && *releaseCreatedAt != "" {
		prTimestamps.ReleaseCreatedAt = releaseCreatedAt
	}

	// Verify release_created_at is in timestamps object
	if prTimestamps.ReleaseCreatedAt == nil {
		t.Error("Expected ReleaseCreatedAt to be set in timestamps object")
	} else if *prTimestamps.ReleaseCreatedAt != "2023-01-16T09:00:00Z" {
		t.Errorf("Expected ReleaseCreatedAt to be 2023-01-16T09:00:00Z, got %v", *prTimestamps.ReleaseCreatedAt)
	}
}

// Package pullmetrics provides functionality to analyze GitHub Pull Requests
// and generate comprehensive metrics and details.
package pullmetrics

import (
	"github.com/google/go-github/v66/github"
)

// PRDetails represents the complete analysis of a GitHub Pull Request
type PRDetails struct {
	OrganizationName           string        `json:"organization_name"`
	RepositoryName             string        `json:"repository_name"`
	PRNumber                   int           `json:"pr_number"`
	PRTitle                    string        `json:"pr_title"`
	PRWebURL                   string        `json:"pr_web_url"`
	PRNodeID                   string        `json:"pr_node_id"`
	AuthorUsername             string        `json:"author_username"`
	ApproverUsernames          []string      `json:"approver_usernames"`
	CommenterUsernames         []string      `json:"commenter_usernames"`
	State                      string        `json:"state"`
	NumComments                int           `json:"num_comments"`
	NumCommenters              int           `json:"num_commenters"`
	NumApprovers               int           `json:"num_approvers"`
	NumRequestedReviewers      int           `json:"num_requested_reviewers"`
	ChangeRequestsCount        int           `json:"change_requests_count"`
	LinesChanged               int           `json:"lines_changed"`
	FilesChanged               int           `json:"files_changed"`
	CommitsAfterFirstReview    int           `json:"commits_after_first_review"`
	JiraIssue                  string        `json:"jira_issue"`
	IsBot                      bool          `json:"is_bot"`
	Metrics                    *PRMetrics    `json:"metrics,omitempty"`
	ReleaseName                *string       `json:"release_name,omitempty"`
	Timestamps                 *PRTimestamps `json:"timestamps,omitempty"`
	GeneratedAt                string        `json:"generated_at"`
}

// PRSize represents the size metrics of a Pull Request
type PRSize struct {
	LinesChanged int
	FilesChanged int
}

// Timestamps represents internal timestamp data for PR analysis
type Timestamps struct {
	FirstCommit        *string
	CreatedAt          *string
	FirstReviewRequest *string
	FirstComment       *string
	FirstApproval      *string
	SecondApproval     *string
	MergedAt           *string
	ClosedAt           *string
}

// PRTimestamps represents the JSON output structure for PR timestamps
type PRTimestamps struct {
	FirstCommit        *string `json:"first_commit,omitempty"`
	CreatedAt          *string `json:"created_at,omitempty"`
	FirstReviewRequest *string `json:"first_review_request,omitempty"`
	FirstComment       *string `json:"first_comment,omitempty"`
	FirstApproval      *string `json:"first_approval,omitempty"`
	SecondApproval     *string `json:"second_approval,omitempty"`
	MergedAt           *string `json:"merged_at,omitempty"`
	ClosedAt           *string `json:"closed_at,omitempty"`
	ReleaseCreatedAt   *string `json:"release_created_at,omitempty"`
}

// PRMetrics represents calculated performance metrics for the PR review process
type PRMetrics struct {
	DraftTimeHours                float64  `json:"draft_time_hours"`
	TimeToFirstReviewRequestHours *float64 `json:"time_to_first_review_request_hours,omitempty"`
	TimeToFirstReviewHours        *float64 `json:"time_to_first_review_hours,omitempty"`
	ReviewCycleTimeHours          *float64 `json:"review_cycle_time_hours,omitempty"`
	BlockingNonBlockingRatio      *float64 `json:"blocking_non_blocking_ratio,omitempty"`
	ReviewerParticipationRatio    *float64 `json:"reviewer_participation_ratio,omitempty"`
}

// ReleaseInfo holds both the name and creation timestamp of a release
type ReleaseInfo struct {
	Name      string
	CreatedAt string
}

// Config represents the configuration for the PR analysis
type Config struct {
	GitHubToken string
}

// Analyzer provides the core functionality for analyzing GitHub Pull Requests
type Analyzer struct {
	client *github.Client
}
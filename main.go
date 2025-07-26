package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/google/go-github/v66/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type PRDetails struct {
	OrganizationName  string   `json:"organization_name"`
	RepositoryName     string   `json:"repository_name"`
	PRNumber          int      `json:"pr_number"`
	PRTitle           string   `json:"pr_title"`
	PRWebURL          string   `json:"pr_web_url"`
	PRNodeID          string   `json:"pr_node_id"`
	AuthorUsername    string   `json:"author_username"`
	ApproverUsernames []string `json:"approver_usernames"`
	CommenterUsernames []string `json:"commenter_usernames"`
	State             string   `json:"state"`
	NumComments       int      `json:"num_comments"`
	NumCommenters     int      `json:"num_commenters"`
	NumApprovers      int      `json:"num_approvers"`
	NumRequestedReviewers int  `json:"num_requested_reviewers"`
	ChangeRequestsCount int    `json:"change_requests_count"`
	ChangeRequestCommentsCount int `json:"change_request_comments_count"`
	LinesChanged      int      `json:"lines_changed"`
	FilesChanged      int      `json:"files_changed"`
	CommitsAfterFirstReview int `json:"commits_after_first_review"`
	JiraIssue         string   `json:"jira_issue"`
	IsBot             bool     `json:"is_bot"`
	Metrics           *PRMetrics `json:"metrics,omitempty"`
	ReleaseName       *string  `json:"release_name,omitempty"`
	Timestamps        *PRTimestamps `json:"timestamps,omitempty"`
	GeneratedAt       string   `json:"generated_at"`
}


type PRSize struct {
	LinesChanged int
	FilesChanged int
}

type Timestamps struct {
	FirstCommit       *string
	CreatedAt         *string
	FirstReviewRequest *string
	FirstComment      *string
	FirstApproval     *string
	SecondApproval    *string
	MergedAt          *string
	ClosedAt          *string
}

type PRTimestamps struct {
	FirstCommit       *string `json:"first_commit,omitempty"`
	CreatedAt         *string `json:"created_at,omitempty"`
	FirstReviewRequest *string `json:"first_review_request,omitempty"`
	FirstComment      *string `json:"first_comment,omitempty"`
	FirstApproval     *string `json:"first_approval,omitempty"`
	SecondApproval    *string `json:"second_approval,omitempty"`
	MergedAt          *string `json:"merged_at,omitempty"`
	ClosedAt          *string `json:"closed_at,omitempty"`
}

type PRMetrics struct {
	TimeToFirstReviewRequestHours *float64 `json:"time_to_first_review_request_hours,omitempty"`
	TimeToFirstReviewHours     *float64 `json:"time_to_first_review_hours,omitempty"`
	ReviewCycleTimeHours       *float64 `json:"review_cycle_time_hours,omitempty"`
	BlockingNonBlockingRatio   *float64 `json:"blocking_non_blocking_ratio,omitempty"`
	ReviewerParticipationRatio *float64 `json:"reviewer_participation_ratio,omitempty"`
}

type Config struct {
	Organization string `conf:"pos:0,env:ORGANIZATION,help:GitHub organization or username"`
	Repository   string `conf:"pos:1,env:REPOSITORY,help:Repository name"`
	PRNumber     int    `conf:"pos:2,env:PR_NUMBER,help:Pull Request number"`
	GitHubToken  string `conf:"env:GITHUB_TOKEN,help:GitHub Personal Access Token"`
}

func main() {
	// Load environment variables from .env file if it exists
	// This is optional - if the file doesn't exist, it will just use system environment variables
	_ = godotenv.Load()

	cfg := Config{}
	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Fprintf(os.Stdout, "%s", help)
			return
		}
		fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
		os.Exit(1)
	}


	if cfg.GitHubToken == "" {
		fmt.Fprintf(os.Stderr, "GITHUB_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	// Create GitHub client with OAuth2 token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	prDetails, err := getPRDetails(ctx, client, cfg.Organization, cfg.Repository, cfg.PRNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching PR details: %v\n", err)
		os.Exit(1)
	}

	output, err := json.Marshal(prDetails)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func getPRDetails(ctx context.Context, client *github.Client, org, repo string, prNumber int) (*PRDetails, error) {
	pr, err := fetchPR(ctx, client, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	reviews, err := fetchReviews(ctx, client, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	comments, err := fetchComments(ctx, client, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	reviewComments, err := fetchReviewComments(ctx, client, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	timeline, err := fetchTimeline(ctx, client, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	files, err := fetchPRFiles(ctx, client, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	commits, err := fetchPRCommits(ctx, client, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	var releases []*github.RepositoryRelease
	if *pr.Merged {
		releases, err = fetchReleases(ctx, client, org, repo)
		if err != nil {
			return nil, err
		}
	}

	state := getPRState(pr)
	approvers := getApprovers(reviews)
	commenters := getCommenters(comments, reviewComments, *pr.User.Login)
	commenterUsernames := getCommenterUsernames(commenters)
	numComments := countTotalComments(comments, reviewComments)
	numRequestedReviewers := countAllRequestedReviewers(pr, reviews)
	timestamps := getTimestamps(pr, reviews, comments, reviewComments, timeline, commits)
	prSize := calculatePRSize(files)
	releaseName := findReleaseForMergedPR(pr, releases)
	commitsAfterFirstReview := countCommitsAfterFirstReview(commits, timeline)
	changeRequestsCount := countChangeRequests(reviews)
	changeRequestCommentsCount := countChangeRequestComments(comments, reviewComments, reviews)
	jiraIssue := extractJiraIssue(pr)
	metrics := calculatePRMetrics(pr, reviews, comments, timeline, timestamps)

	result := &PRDetails{
		OrganizationName:     org,
		RepositoryName:        repo,
		PRNumber:             prNumber,
		PRTitle:              *pr.Title,
		PRWebURL:             *pr.HTMLURL,
		PRNodeID:             *pr.NodeID,
		AuthorUsername:       *pr.User.Login,
		ApproverUsernames:    approvers,
		CommenterUsernames:   commenterUsernames,
		State:                state,
		NumComments:          numComments,
		NumCommenters:        len(commenters),
		NumApprovers:         len(approvers),
		NumRequestedReviewers: numRequestedReviewers,
		ChangeRequestsCount:  changeRequestsCount,
		ChangeRequestCommentsCount: changeRequestCommentsCount,
		LinesChanged:         prSize.LinesChanged,
		FilesChanged:         prSize.FilesChanged,
		CommitsAfterFirstReview: commitsAfterFirstReview,
		JiraIssue:            jiraIssue,
		IsBot:                isBot(*pr.User.Login),
		Metrics:              metrics,
		GeneratedAt:          time.Now().UTC().Format(time.RFC3339),
	}

	// Add release name if it exists
	if releaseName != nil {
		result.ReleaseName = releaseName
	}

	// Create timestamps object
	prTimestamps := &PRTimestamps{
		FirstCommit:       timestamps.FirstCommit,
		CreatedAt:         timestamps.CreatedAt,
		FirstReviewRequest: timestamps.FirstReviewRequest,
		FirstComment:      timestamps.FirstComment,
		FirstApproval:     timestamps.FirstApproval,
		SecondApproval:    timestamps.SecondApproval,
		MergedAt:          timestamps.MergedAt,
		ClosedAt:          timestamps.ClosedAt,
	}
	
	result.Timestamps = prTimestamps

	return result, nil
}

func fetchPR(ctx context.Context, client *github.Client, org, repo string, prNumber int) (*github.PullRequest, error) {
	pr, _, err := client.PullRequests.Get(ctx, org, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR: %w", err)
	}
	return pr, nil
}

func fetchReviews(ctx context.Context, client *github.Client, org, repo string, prNumber int) ([]*github.PullRequestReview, error) {
	var allReviews []*github.PullRequestReview
	opts := &github.ListOptions{PerPage: 100}
	
	for {
		reviews, resp, err := client.PullRequests.ListReviews(ctx, org, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch reviews: %w", err)
		}
		allReviews = append(allReviews, reviews...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	
	return allReviews, nil
}

func fetchComments(ctx context.Context, client *github.Client, org, repo string, prNumber int) ([]*github.IssueComment, error) {
	var allComments []*github.IssueComment
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	
	for {
		comments, resp, err := client.Issues.ListComments(ctx, org, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch comments: %w", err)
		}
		allComments = append(allComments, comments...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	
	return allComments, nil
}

func fetchReviewComments(ctx context.Context, client *github.Client, org, repo string, prNumber int) ([]*github.PullRequestComment, error) {
	var allReviewComments []*github.PullRequestComment
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	
	for {
		reviewComments, resp, err := client.PullRequests.ListComments(ctx, org, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch review comments: %w", err)
		}
		allReviewComments = append(allReviewComments, reviewComments...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	
	return allReviewComments, nil
}

func fetchTimeline(ctx context.Context, client *github.Client, org, repo string, prNumber int) ([]*github.Timeline, error) {
	var allTimeline []*github.Timeline
	opts := &github.ListOptions{PerPage: 100}
	
	for {
		timeline, resp, err := client.Issues.ListIssueTimeline(ctx, org, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch timeline: %w", err)
		}
		allTimeline = append(allTimeline, timeline...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	
	return allTimeline, nil
}

func fetchPRFiles(ctx context.Context, client *github.Client, org, repo string, prNumber int) ([]*github.CommitFile, error) {
	var allFiles []*github.CommitFile
	opts := &github.ListOptions{PerPage: 100}
	
	for {
		files, resp, err := client.PullRequests.ListFiles(ctx, org, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PR files: %w", err)
		}
		allFiles = append(allFiles, files...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	
	return allFiles, nil
}

func fetchReleases(ctx context.Context, client *github.Client, org, repo string) ([]*github.RepositoryRelease, error) {
	var allReleases []*github.RepositoryRelease
	opts := &github.ListOptions{PerPage: 100}
	
	for {
		releases, resp, err := client.Repositories.ListReleases(ctx, org, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch releases: %w", err)
		}
		allReleases = append(allReleases, releases...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	
	return allReleases, nil
}

func fetchPRCommits(ctx context.Context, client *github.Client, org, repo string, prNumber int) ([]*github.RepositoryCommit, error) {
	var allCommits []*github.RepositoryCommit
	opts := &github.ListOptions{PerPage: 100}
	
	for {
		commits, resp, err := client.PullRequests.ListCommits(ctx, org, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PR commits: %w", err)
		}
		allCommits = append(allCommits, commits...)
		
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	
	return allCommits, nil
}

func getPRState(pr *github.PullRequest) string {
	if pr.GetDraft() {
		return "draft"
	}
	if pr.GetMerged() {
		return "merged"
	}
	return pr.GetState()
}

func getApprovers(reviews []*github.PullRequestReview) []string {
	approvers := make(map[string]bool)
	for _, review := range reviews {
		if review.GetState() == "APPROVED" {
			approvers[review.GetUser().GetLogin()] = true
		}
	}

	result := make([]string, 0, len(approvers))
	for username := range approvers {
		result = append(result, username)
	}
	return result
}

func getCommenters(comments []*github.IssueComment, reviewComments []*github.PullRequestComment, authorUsername string) map[string]bool {
	commenters := make(map[string]bool)
	
	// Process regular comments
	for _, comment := range comments {
		if comment.GetUser().GetLogin() != authorUsername {
			commenters[comment.GetUser().GetLogin()] = true
		}
	}
	
	// Process review comments
	for _, reviewComment := range reviewComments {
		if reviewComment.GetUser().GetLogin() != authorUsername {
			commenters[reviewComment.GetUser().GetLogin()] = true
		}
	}
	
	return commenters
}

func countTotalComments(comments []*github.IssueComment, reviewComments []*github.PullRequestComment) int {
	return len(comments) + len(reviewComments)
}

func getCommenterUsernames(commenters map[string]bool) []string {
	usernames := make([]string, 0, len(commenters))
	for username := range commenters {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames) // Sort for consistent output
	return usernames
}

func countAllRequestedReviewers(pr *github.PullRequest, reviews []*github.PullRequestReview) int {
	// Count all reviewers who were requested to review (both those who reviewed and those who haven't)
	requestedReviewers := make(map[string]bool)
	
	// Add users who have submitted reviews (they must have been requested to review)
	for _, review := range reviews {
		requestedReviewers[review.GetUser().GetLogin()] = true
	}
	
	// Add current requested reviewers (those who haven't reviewed yet)
	for _, reviewer := range pr.RequestedReviewers {
		requestedReviewers[reviewer.GetLogin()] = true
	}
	
	return len(requestedReviewers)
}

func getTimestamps(pr *github.PullRequest, reviews []*github.PullRequestReview, comments []*github.IssueComment, reviewComments []*github.PullRequestComment, timeline []*github.Timeline, commits []*github.RepositoryCommit) *Timestamps {
	timestamps := &Timestamps{}

	// First commit timestamp (from commits)
	if len(commits) > 0 {
		// Sort commits by date to get the first one
		sort.Slice(commits, func(i, j int) bool {
			return commits[i].GetCommit().GetAuthor().GetDate().Before(commits[j].GetCommit().GetAuthor().GetDate().Time)
		})
		utcTime := formatToUTC(commits[0].GetCommit().GetAuthor().GetDate().Format(time.RFC3339))
		timestamps.FirstCommit = &utcTime
	}

	// Created timestamp (from PR)
	if !pr.GetCreatedAt().IsZero() {
		utcTime := formatToUTC(pr.GetCreatedAt().Format(time.RFC3339))
		timestamps.CreatedAt = &utcTime
	}

	// Merged and closed timestamps (from PR)
	if pr.MergedAt != nil && !pr.GetMergedAt().IsZero() {
		utcTime := formatToUTC(pr.GetMergedAt().Format(time.RFC3339))
		timestamps.MergedAt = &utcTime
	}
	if pr.ClosedAt != nil && !pr.GetClosedAt().IsZero() {
		utcTime := formatToUTC(pr.GetClosedAt().Format(time.RFC3339))
		timestamps.ClosedAt = &utcTime
	}

	// First review request (from timeline)
	for _, event := range timeline {
		if event.GetEvent() == "review_requested" && timestamps.FirstReviewRequest == nil {
			utcTime := formatToUTC(event.GetCreatedAt().Format(time.RFC3339))
			timestamps.FirstReviewRequest = &utcTime
			break
		}
	}

	// First comment (from both regular comments and review comments)
	var allComments []time.Time
	
	// Collect all comment timestamps
	for _, comment := range comments {
		allComments = append(allComments, comment.GetCreatedAt().Time)
	}
	for _, reviewComment := range reviewComments {
		allComments = append(allComments, reviewComment.GetCreatedAt().Time)
	}
	
	if len(allComments) > 0 {
		// Sort all comment timestamps to get the first one
		sort.Slice(allComments, func(i, j int) bool {
			return allComments[i].Before(allComments[j])
		})
		utcTime := formatToUTC(allComments[0].Format(time.RFC3339))
		timestamps.FirstComment = &utcTime
	}

	// First and second approvals (from reviews)
	var approvals []*github.PullRequestReview
	for _, review := range reviews {
		if review.GetState() == "APPROVED" {
			approvals = append(approvals, review)
		}
	}
	
	// Sort approvals by submission time
	sort.Slice(approvals, func(i, j int) bool {
		return approvals[i].GetSubmittedAt().Before(approvals[j].GetSubmittedAt().Time)
	})

	if len(approvals) > 0 {
		utcTime := formatToUTC(approvals[0].GetSubmittedAt().Format(time.RFC3339))
		timestamps.FirstApproval = &utcTime
	}
	if len(approvals) > 1 {
		utcTime := formatToUTC(approvals[1].GetSubmittedAt().Format(time.RFC3339))
		timestamps.SecondApproval = &utcTime
	}

	return timestamps
}

func formatToUTC(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp // Return original if parsing fails
	}
	return t.UTC().Format(time.RFC3339)
}

func calculatePRSize(files []*github.CommitFile) *PRSize {
	size := &PRSize{
		LinesChanged: 0,
		FilesChanged: len(files),
	}

	for _, file := range files {
		// Count total lines changed (additions + deletions)
		size.LinesChanged += file.GetAdditions() + file.GetDeletions()
	}

	return size
}

func findReleaseForMergedPR(pr *github.PullRequest, releases []*github.RepositoryRelease) *string {
	// Only check for releases if the PR was merged
	if !pr.GetMerged() || pr.MergedAt == nil {
		return nil
	}

	mergedTime := pr.GetMergedAt().Time

	// Find releases published after the PR was merged
	var validReleases []*github.RepositoryRelease
	for _, release := range releases {
		if release.PublishedAt == nil || release.GetPublishedAt().IsZero() {
			continue
		}
		
		publishedTime := release.GetPublishedAt().Time

		// If the release was published after the PR was merged, 
		// this PR is likely included in this release
		if publishedTime.After(mergedTime) {
			validReleases = append(validReleases, release)
		}
	}

	if len(validReleases) == 0 {
		return nil
	}

	// Sort valid releases by published date (oldest first) to get the first release after merge
	sort.Slice(validReleases, func(i, j int) bool {
		return validReleases[i].GetPublishedAt().Before(validReleases[j].GetPublishedAt().Time)
	})

	// Return the first (earliest) release after merge
	release := validReleases[0]
	releaseName := release.GetName()
	if releaseName == "" {
		releaseName = release.GetTagName()
	}
	return &releaseName
}

func countCommitsAfterFirstReview(commits []*github.RepositoryCommit, timeline []*github.Timeline) int {
	// Find the first review request timestamp
	var firstReviewRequestTime *time.Time
	for _, event := range timeline {
		if event.GetEvent() == "review_requested" {
			t := event.GetCreatedAt().Time
			firstReviewRequestTime = &t
			break
		}
	}

	// If no review request was made, return 0
	if firstReviewRequestTime == nil {
		return 0
	}

	// Count commits made after the first review request
	count := 0
	for _, commit := range commits {
		commitTime := commit.GetCommit().GetAuthor().GetDate().Time
		if commitTime.After(*firstReviewRequestTime) {
			count++
		}
	}

	return count
}

func countChangeRequests(reviews []*github.PullRequestReview) int {
	count := 0
	for _, review := range reviews {
		if review.GetState() == "CHANGES_REQUESTED" {
			count++
		}
	}
	return count
}

func countChangeRequestComments(comments []*github.IssueComment, reviewComments []*github.PullRequestComment, reviews []*github.PullRequestReview) int {
	// Create a set of user logins who submitted change requests
	changeRequesters := make(map[string]bool)
	for _, review := range reviews {
		if review.GetState() == "CHANGES_REQUESTED" {
			changeRequesters[review.GetUser().GetLogin()] = true
		}
	}
	
	count := 0
	
	// Count regular comments from users who submitted change requests
	for _, comment := range comments {
		if changeRequesters[comment.GetUser().GetLogin()] {
			count++
		}
	}
	
	// Count review comments from users who submitted change requests
	for _, reviewComment := range reviewComments {
		if changeRequesters[reviewComment.GetUser().GetLogin()] {
			count++
		}
	}
	
	return count
}

func isBot(username string) bool {
	return strings.Contains(username, "[bot]")
}

func findValidJiraIssue(pattern *regexp.Regexp, text string) string {
	// Find all matches in the text
	matches := pattern.FindAllString(text, -1)
	for _, match := range matches {
		upperMatch := strings.ToUpper(match)
		// Exclude CVE identifiers (security vulnerability IDs)
		if !strings.HasPrefix(upperMatch, "CVE-") {
			return upperMatch
		}
	}
	return ""
}

func extractJiraIssue(pr *github.PullRequest) string {
	// Jira issue pattern: PROJECT-123, ABC-1234, etc.
	// Matches project key (2+ uppercase letters or alphanumeric) followed by hyphen and number
	// Excludes CVE- identifiers which are security vulnerability IDs, not Jira issues
	jiraPattern := regexp.MustCompile(`\b[A-Z][A-Z0-9]+-\d+\b`)
	
	// Search in PR title first
	if issue := findValidJiraIssue(jiraPattern, pr.GetTitle()); issue != "" {
		return issue
	}
	
	// Search in PR body if available
	if pr.Body != nil && pr.GetBody() != "" {
		if issue := findValidJiraIssue(jiraPattern, pr.GetBody()); issue != "" {
			return issue
		}
	}
	
	// Search in branch name (head ref)
	if issue := findValidJiraIssue(jiraPattern, strings.ToUpper(pr.GetHead().GetRef())); issue != "" {
		return issue
	}
	
	// If not found, check if the user is a bot
	if isBot(pr.GetUser().GetLogin()) {
		return "BOT"
	}
	
	// If not a bot and no Jira issue found, return UNKNOWN
	return "UNKNOWN"
}

func calculatePRMetrics(pr *github.PullRequest, reviews []*github.PullRequestReview, comments []*github.IssueComment, timeline []*github.Timeline, timestamps *Timestamps) *PRMetrics {
	metrics := &PRMetrics{}
	
	// Time to First Review Request: time from PR creation to first review request
	if timestamps.CreatedAt != nil && timestamps.FirstReviewRequest != nil {
		if createdTime, err := time.Parse(time.RFC3339, *timestamps.CreatedAt); err == nil {
			if firstReviewRequestTime, err := time.Parse(time.RFC3339, *timestamps.FirstReviewRequest); err == nil {
				if firstReviewRequestTime.After(createdTime) {
					hours := firstReviewRequestTime.Sub(createdTime).Hours()
					metrics.TimeToFirstReviewRequestHours = &hours
				}
			}
		}
	}
	
	// Time to First Review: time from first review request to first comment or first approval
	if timestamps.FirstReviewRequest != nil {
		if firstReviewRequestTime, err := time.Parse(time.RFC3339, *timestamps.FirstReviewRequest); err == nil {
			var firstReviewActivityTime *time.Time
			
			// Find the earliest between first comment and first approval
			if timestamps.FirstComment != nil {
				if firstCommentTime, err := time.Parse(time.RFC3339, *timestamps.FirstComment); err == nil {
					firstReviewActivityTime = &firstCommentTime
				}
			}
			
			if timestamps.FirstApproval != nil {
				if firstApprovalTime, err := time.Parse(time.RFC3339, *timestamps.FirstApproval); err == nil {
					if firstReviewActivityTime == nil || firstApprovalTime.Before(*firstReviewActivityTime) {
						firstReviewActivityTime = &firstApprovalTime
					}
				}
			}
			
			// Calculate time to first review activity if we have one and it's after the review request
			if firstReviewActivityTime != nil && firstReviewActivityTime.After(firstReviewRequestTime) {
				hours := firstReviewActivityTime.Sub(firstReviewRequestTime).Hours()
				metrics.TimeToFirstReviewHours = &hours
			}
		}
	}
	
	// Review Cycle Time: time from first review request to PR resolution (merged or closed)
	if timestamps.FirstReviewRequest != nil {
		if firstReviewTime, err := time.Parse(time.RFC3339, *timestamps.FirstReviewRequest); err == nil {
			var resolutionTime *time.Time
			
			// Use merged time if available, otherwise closed time
			if timestamps.MergedAt != nil {
				if mergedTime, err := time.Parse(time.RFC3339, *timestamps.MergedAt); err == nil {
					resolutionTime = &mergedTime
				}
			} else if timestamps.ClosedAt != nil {
				if closedTime, err := time.Parse(time.RFC3339, *timestamps.ClosedAt); err == nil {
					resolutionTime = &closedTime
				}
			}
			
			if resolutionTime != nil && resolutionTime.After(firstReviewTime) {
				hours := resolutionTime.Sub(firstReviewTime).Hours()
				metrics.ReviewCycleTimeHours = &hours
			}
		}
	}
	
	// Blocking vs Non-Blocking comment ratio
	blockingCount := 0
	nonBlockingCount := 0
	
	for _, review := range reviews {
		if review.GetState() == "CHANGES_REQUESTED" {
			blockingCount++
		} else if review.GetState() == "COMMENTED" || review.GetState() == "APPROVED" {
			nonBlockingCount++
		}
	}
	
	if nonBlockingCount > 0 {
		ratio := float64(blockingCount) / float64(nonBlockingCount)
		metrics.BlockingNonBlockingRatio = &ratio
	}
	
	// Reviewer Participation Ratio: (actual reviewers) / (requested reviewers)
	actualReviewers := make(map[string]bool)
	for _, review := range reviews {
		actualReviewers[review.GetUser().GetLogin()] = true
	}
	
	requestedReviewers := countAllRequestedReviewers(pr, reviews)
	if requestedReviewers > 0 {
		ratio := float64(len(actualReviewers)) / float64(requestedReviewers)
		metrics.ReviewerParticipationRatio = &ratio
	}
	
	return metrics
}
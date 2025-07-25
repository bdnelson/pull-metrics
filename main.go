package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type PRDetails struct {
	OrganizationName  string   `json:"organization_name"`
	RepositoryName     string   `json:"repository_name"`
	PRNumber          int      `json:"pr_number"`
	AuthorUsername    string   `json:"author_username"`
	ApproverUsernames []string `json:"approver_usernames"`
	State             string   `json:"state"`
	NumCommentors     int      `json:"num_commentors"`
	NumApprovers      int      `json:"num_approvers"`
	NumRequestedReviewers int  `json:"num_requested_reviewers"`
	LinesChanged      int      `json:"lines_changed"`
	FilesChanged      int      `json:"files_changed"`
	ReleaseName       *string  `json:"release_name,omitempty"`
	CreatedAt         *string  `json:"created_at,omitempty"`
	FirstReviewRequest *string `json:"first_review_request,omitempty"`
	FirstComment      *string  `json:"first_comment,omitempty"`
	FirstApproval     *string  `json:"first_approval,omitempty"`
	SecondApproval    *string  `json:"second_approval,omitempty"`
	MergedAt          *string  `json:"merged_at,omitempty"`
	ClosedAt          *string  `json:"closed_at,omitempty"`
}

type GitHubPR struct {
	Number int `json:"number"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	State       string `json:"state"`
	Draft       bool   `json:"draft"`
	Merged      bool   `json:"merged"`
	CreatedAt   string `json:"created_at"`
	MergedAt    *string `json:"merged_at"`
	ClosedAt    *string `json:"closed_at"`
	RequestedReviewers []struct {
		Login string `json:"login"`
	} `json:"requested_reviewers"`
}

type GitHubReview struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	State       string `json:"state"`
	SubmittedAt string `json:"submitted_at"`
}

type GitHubComment struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
}

type GitHubTimelineEvent struct {
	Event     string `json:"event"`
	CreatedAt string `json:"created_at"`
	Actor     struct {
		Login string `json:"login"`
	} `json:"actor"`
}

type GitHubPRFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
}

type GitHubRelease struct {
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`
}

type PRSize struct {
	LinesChanged int
	FilesChanged int
}

type Timestamps struct {
	CreatedAt         *string
	FirstReviewRequest *string
	FirstComment      *string
	FirstApproval     *string
	SecondApproval    *string
	MergedAt          *string
	ClosedAt          *string
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <organization> <repository> <pr_number>\n", os.Args[0])
		os.Exit(1)
	}

	org := os.Args[1]
	repo := os.Args[2]
	prNumStr := os.Args[3]

	prNumber, err := strconv.Atoi(prNumStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid PR number: %s\n", prNumStr)
		os.Exit(1)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "GITHUB_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	client := &http.Client{}

	prDetails, err := getPRDetails(client, token, org, repo, prNumber)
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

func getPRDetails(client *http.Client, token, org, repo string, prNumber int) (*PRDetails, error) {
	pr, err := fetchPR(client, token, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	reviews, err := fetchReviews(client, token, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	comments, err := fetchComments(client, token, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	timeline, err := fetchTimeline(client, token, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	files, err := fetchPRFiles(client, token, org, repo, prNumber)
	if err != nil {
		return nil, err
	}

	var releases []GitHubRelease
	if pr.Merged {
		releases, err = fetchReleases(client, token, org, repo)
		if err != nil {
			return nil, err
		}
	}

	state := getPRState(pr)
	approvers := getApprovers(reviews)
	commentors := getCommentors(comments, pr.User.Login)
	timestamps := getTimestamps(pr, reviews, comments, timeline)
	prSize := calculatePRSize(files)
	releaseName := findReleaseForMergedPR(pr, releases)

	result := &PRDetails{
		OrganizationName:     org,
		RepositoryName:        repo,
		PRNumber:             prNumber,
		AuthorUsername:       pr.User.Login,
		ApproverUsernames:    approvers,
		State:                state,
		NumCommentors:        len(commentors),
		NumApprovers:         len(approvers),
		NumRequestedReviewers: len(pr.RequestedReviewers),
		LinesChanged:         prSize.LinesChanged,
		FilesChanged:         prSize.FilesChanged,
	}

	// Add release name if it exists
	if releaseName != nil {
		result.ReleaseName = releaseName
	}

	// Add timestamps if they exist
	if timestamps.CreatedAt != nil {
		result.CreatedAt = timestamps.CreatedAt
	}
	if timestamps.FirstReviewRequest != nil {
		result.FirstReviewRequest = timestamps.FirstReviewRequest
	}
	if timestamps.FirstComment != nil {
		result.FirstComment = timestamps.FirstComment
	}
	if timestamps.FirstApproval != nil {
		result.FirstApproval = timestamps.FirstApproval
	}
	if timestamps.SecondApproval != nil {
		result.SecondApproval = timestamps.SecondApproval
	}
	if timestamps.MergedAt != nil {
		result.MergedAt = timestamps.MergedAt
	}
	if timestamps.ClosedAt != nil {
		result.ClosedAt = timestamps.ClosedAt
	}

	return result, nil
}

func fetchPR(client *http.Client, token, org, repo string, prNumber int) (*GitHubPR, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", org, repo, prNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var pr GitHubPR
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

func fetchReviews(client *http.Client, token, org, repo string, prNumber int) ([]GitHubReview, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews", org, repo, prNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for reviews", resp.StatusCode)
	}

	var reviews []GitHubReview
	if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func fetchComments(client *http.Client, token, org, repo string, prNumber int) ([]GitHubComment, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", org, repo, prNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for comments", resp.StatusCode)
	}

	var comments []GitHubComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func fetchTimeline(client *http.Client, token, org, repo string, prNumber int) ([]GitHubTimelineEvent, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/timeline", org, repo, prNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.mockingbird-preview")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for timeline", resp.StatusCode)
	}

	var timeline []GitHubTimelineEvent
	if err := json.NewDecoder(resp.Body).Decode(&timeline); err != nil {
		return nil, err
	}

	return timeline, nil
}

func fetchPRFiles(client *http.Client, token, org, repo string, prNumber int) ([]GitHubPRFile, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files", org, repo, prNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for PR files", resp.StatusCode)
	}

	var files []GitHubPRFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	return files, nil
}

func fetchReleases(client *http.Client, token, org, repo string) ([]GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", org, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for releases", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

func getPRState(pr *GitHubPR) string {
	if pr.Draft {
		return "draft"
	}
	if pr.Merged {
		return "merged"
	}
	return pr.State
}

func getApprovers(reviews []GitHubReview) []string {
	approvers := make(map[string]bool)
	for _, review := range reviews {
		if review.State == "APPROVED" {
			approvers[review.User.Login] = true
		}
	}

	result := make([]string, 0, len(approvers))
	for username := range approvers {
		result = append(result, username)
	}
	return result
}

func getCommentors(comments []GitHubComment, authorUsername string) map[string]bool {
	commentors := make(map[string]bool)
	for _, comment := range comments {
		if comment.User.Login != authorUsername {
			commentors[comment.User.Login] = true
		}
	}
	return commentors
}

func getTimestamps(pr *GitHubPR, reviews []GitHubReview, comments []GitHubComment, timeline []GitHubTimelineEvent) *Timestamps {
	timestamps := &Timestamps{}

	// Created timestamp (from PR)
	if pr.CreatedAt != "" {
		utcTime := formatToUTC(pr.CreatedAt)
		timestamps.CreatedAt = &utcTime
	}

	// Merged and closed timestamps (from PR)
	if pr.MergedAt != nil && *pr.MergedAt != "" {
		utcTime := formatToUTC(*pr.MergedAt)
		timestamps.MergedAt = &utcTime
	}
	if pr.ClosedAt != nil && *pr.ClosedAt != "" {
		utcTime := formatToUTC(*pr.ClosedAt)
		timestamps.ClosedAt = &utcTime
	}

	// First review request (from timeline)
	for _, event := range timeline {
		if event.Event == "review_requested" && timestamps.FirstReviewRequest == nil {
			utcTime := formatToUTC(event.CreatedAt)
			timestamps.FirstReviewRequest = &utcTime
			break
		}
	}

	// First comment (from comments)
	if len(comments) > 0 {
		// Sort comments by creation time to get the first one
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].CreatedAt < comments[j].CreatedAt
		})
		utcTime := formatToUTC(comments[0].CreatedAt)
		timestamps.FirstComment = &utcTime
	}

	// First and second approvals (from reviews)
	var approvals []GitHubReview
	for _, review := range reviews {
		if review.State == "APPROVED" {
			approvals = append(approvals, review)
		}
	}
	
	// Sort approvals by submission time
	sort.Slice(approvals, func(i, j int) bool {
		return approvals[i].SubmittedAt < approvals[j].SubmittedAt
	})

	if len(approvals) > 0 {
		utcTime := formatToUTC(approvals[0].SubmittedAt)
		timestamps.FirstApproval = &utcTime
	}
	if len(approvals) > 1 {
		utcTime := formatToUTC(approvals[1].SubmittedAt)
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

func calculatePRSize(files []GitHubPRFile) *PRSize {
	size := &PRSize{
		LinesChanged: 0,
		FilesChanged: len(files),
	}

	for _, file := range files {
		// Count total lines changed (additions + deletions)
		size.LinesChanged += file.Additions + file.Deletions
	}

	return size
}

func findReleaseForMergedPR(pr *GitHubPR, releases []GitHubRelease) *string {
	// Only check for releases if the PR was merged
	if !pr.Merged || pr.MergedAt == nil {
		return nil
	}

	mergedTime, err := time.Parse(time.RFC3339, *pr.MergedAt)
	if err != nil {
		return nil
	}

	// Find releases published after the PR was merged
	var validReleases []GitHubRelease
	for _, release := range releases {
		if release.PublishedAt == "" {
			continue
		}
		
		publishedTime, err := time.Parse(time.RFC3339, release.PublishedAt)
		if err != nil {
			continue
		}

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
		timeI, errI := time.Parse(time.RFC3339, validReleases[i].PublishedAt)
		timeJ, errJ := time.Parse(time.RFC3339, validReleases[j].PublishedAt)
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.Before(timeJ)
	})

	// Return the first (earliest) release after merge
	release := validReleases[0]
	releaseName := release.Name
	if releaseName == "" {
		releaseName = release.TagName
	}
	return &releaseName
}
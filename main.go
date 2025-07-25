package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type PRDetails struct {
	RepositoryName     string   `json:"repository_name"`
	PRNumber          int      `json:"pr_number"`
	AuthorUsername    string   `json:"author_username"`
	ApproverUsernames []string `json:"approver_usernames"`
	State             string   `json:"state"`
	NumCommentors     int      `json:"num_commentors"`
	NumApprovers      int      `json:"num_approvers"`
	NumRequestedReviewers int  `json:"num_requested_reviewers"`
}

type GitHubPR struct {
	Number int `json:"number"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	State       string `json:"state"`
	Draft       bool   `json:"draft"`
	Merged      bool   `json:"merged"`
	RequestedReviewers []struct {
		Login string `json:"login"`
	} `json:"requested_reviewers"`
}

type GitHubReview struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	State string `json:"state"`
}

type GitHubComment struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
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

	state := getPRState(pr)
	approvers := getApprovers(reviews)
	commentors := getCommentors(comments, pr.User.Login)

	return &PRDetails{
		RepositoryName:        repo,
		PRNumber:             prNumber,
		AuthorUsername:       pr.User.Login,
		ApproverUsernames:    approvers,
		State:                state,
		NumCommentors:        len(commentors),
		NumApprovers:         len(approvers),
		NumRequestedReviewers: len(pr.RequestedReviewers),
	}, nil
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
package main

import (
	"testing"
)

func TestGetPRState(t *testing.T) {
	tests := []struct {
		name     string
		pr       *GitHubPR
		expected string
	}{
		{
			name: "draft PR",
			pr: &GitHubPR{
				State: "open",
				Draft: true,
				Merged: false,
			},
			expected: "draft",
		},
		{
			name: "merged PR",
			pr: &GitHubPR{
				State: "closed",
				Draft: false,
				Merged: true,
			},
			expected: "merged",
		},
		{
			name: "open PR",
			pr: &GitHubPR{
				State: "open",
				Draft: false,
				Merged: false,
			},
			expected: "open",
		},
		{
			name: "closed PR",
			pr: &GitHubPR{
				State: "closed",
				Draft: false,
				Merged: false,
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
		reviews  []GitHubReview
		expected []string
	}{
		{
			name: "no reviews",
			reviews: []GitHubReview{},
			expected: []string{},
		},
		{
			name: "single approver",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State: "APPROVED",
				},
			},
			expected: []string{"user1"},
		},
		{
			name: "multiple approvers",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State: "APPROVED",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user2"},
					State: "APPROVED",
				},
			},
			expected: []string{"user1", "user2"},
		},
		{
			name: "mixed review states",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State: "APPROVED",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user2"},
					State: "CHANGES_REQUESTED",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user3"},
					State: "COMMENTED",
				},
			},
			expected: []string{"user1"},
		},
		{
			name: "duplicate approver (same user approves multiple times)",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State: "APPROVED",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State: "APPROVED",
				},
			},
			expected: []string{"user1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getApprovers(tt.reviews)
			
			// Convert result to map for easier comparison since order doesn't matter
			resultMap := make(map[string]bool)
			for _, username := range result {
				resultMap[username] = true
			}
			
			expectedMap := make(map[string]bool)
			for _, username := range tt.expected {
				expectedMap[username] = true
			}
			
			if len(resultMap) != len(expectedMap) {
				t.Errorf("getApprovers() returned %d approvers, want %d", len(resultMap), len(expectedMap))
				return
			}
			
			for username := range expectedMap {
				if !resultMap[username] {
					t.Errorf("getApprovers() missing approver %s", username)
				}
			}
			
			for username := range resultMap {
				if !expectedMap[username] {
					t.Errorf("getApprovers() unexpected approver %s", username)
				}
			}
		})
	}
}

func TestGetCommentors(t *testing.T) {
	tests := []struct {
		name           string
		comments       []GitHubComment
		authorUsername string
		expectedCount  int
	}{
		{
			name:           "no comments",
			comments:       []GitHubComment{},
			authorUsername: "author",
			expectedCount:  0,
		},
		{
			name: "single commentor",
			comments: []GitHubComment{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter1"},
				},
			},
			authorUsername: "author",
			expectedCount:  1,
		},
		{
			name: "multiple commentors",
			comments: []GitHubComment{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter1"},
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter2"},
				},
			},
			authorUsername: "author",
			expectedCount:  2,
		},
		{
			name: "exclude author comments",
			comments: []GitHubComment{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter1"},
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "author"},
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter2"},
				},
			},
			authorUsername: "author",
			expectedCount:  2,
		},
		{
			name: "duplicate commentor",
			comments: []GitHubComment{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter1"},
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter1"},
				},
			},
			authorUsername: "author",
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommentors(tt.comments, tt.authorUsername)
			if len(result) != tt.expectedCount {
				t.Errorf("getCommentors() returned %d commentors, want %d", len(result), tt.expectedCount)
			}
		})
	}
}
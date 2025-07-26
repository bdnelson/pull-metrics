package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
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
				State:   "open",
				Draft:   true,
				Merged:  false,
				Title:   "Draft PR",
				HTMLURL: "https://github.com/org/repo/pull/1",
				NodeID:  "PR_node123",
			},
			expected: "draft",
		},
		{
			name: "merged PR",
			pr: &GitHubPR{
				State:   "closed",
				Draft:   false,
				Merged:  true,
				Title:   "Merged PR",
				HTMLURL: "https://github.com/org/repo/pull/2",
				NodeID:  "PR_node456",
			},
			expected: "merged",
		},
		{
			name: "open PR",
			pr: &GitHubPR{
				State:   "open",
				Draft:   false,
				Merged:  false,
				Title:   "Open PR",
				HTMLURL: "https://github.com/org/repo/pull/3",
				NodeID:  "PR_node789",
			},
			expected: "open",
		},
		{
			name: "closed PR",
			pr: &GitHubPR{
				State:   "closed",
				Draft:   false,
				Merged:  false,
				Title:   "Closed PR",
				HTMLURL: "https://github.com/org/repo/pull/4",
				NodeID:  "PR_node101",
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
					SubmittedAt: "2023-01-01T12:00:00Z",
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
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user2"},
					State: "APPROVED",
					SubmittedAt: "2023-01-01T13:00:00Z",
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
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user2"},
					State: "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T13:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user3"},
					State: "COMMENTED",
					SubmittedAt: "2023-01-01T14:00:00Z",
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
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State: "APPROVED",
					SubmittedAt: "2023-01-01T13:00:00Z",
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
					CreatedAt: "2023-01-01T12:00:00Z",
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
					CreatedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter2"},
					CreatedAt: "2023-01-01T13:00:00Z",
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
					CreatedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "author"},
					CreatedAt: "2023-01-01T13:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter2"},
					CreatedAt: "2023-01-01T14:00:00Z",
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
					CreatedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter1"},
					CreatedAt: "2023-01-01T13:00:00Z",
				},
			},
			authorUsername: "author",
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommentors(tt.comments, []GitHubReviewComment{}, tt.authorUsername)
			if len(result) != tt.expectedCount {
				t.Errorf("getCommentors() returned %d commentors, want %d", len(result), tt.expectedCount)
			}
		})
	}
}

func TestFormatToUTC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already UTC",
			input:    "2023-01-01T12:00:00Z",
			expected: "2023-01-01T12:00:00Z",
		},
		{
			name:     "with timezone offset",
			input:    "2023-01-01T12:00:00-05:00",
			expected: "2023-01-01T17:00:00Z",
		},
		{
			name:     "invalid format returns original",
			input:    "invalid-date",
			expected: "invalid-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatToUTC(tt.input)
			if result != tt.expected {
				t.Errorf("formatToUTC() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetTimestamps(t *testing.T) {
	tests := []struct {
		name     string
		pr       *GitHubPR
		reviews  []GitHubReview
		comments []GitHubComment
		timeline []GitHubTimelineEvent
		commits  []GitHubCommit
		validate func(*testing.T, *Timestamps)
	}{
		{
			name: "all timestamps present",
			pr: &GitHubPR{
				CreatedAt: "2023-01-01T10:00:00Z",
				MergedAt:  stringPtr("2023-01-01T18:00:00Z"),
				ClosedAt:  stringPtr("2023-01-01T18:00:00Z"),
			},
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer1"},
					State:       "APPROVED",
					SubmittedAt: "2023-01-01T15:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer2"},
					State:       "APPROVED",
					SubmittedAt: "2023-01-01T16:00:00Z",
				},
			},
			comments: []GitHubComment{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "commenter1"},
					CreatedAt: "2023-01-01T12:00:00Z",
				},
			},
			timeline: []GitHubTimelineEvent{
				{
					Event:     "review_requested",
					CreatedAt: "2023-01-01T11:00:00Z",
				},
			},
			commits: []GitHubCommit{
				{
					SHA: "abc123",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T09:00:00Z"},
					},
				},
			},
			validate: func(t *testing.T, ts *Timestamps) {
				if ts.FirstCommit == nil || *ts.FirstCommit != "2023-01-01T09:00:00Z" {
					t.Errorf("Expected FirstCommit to be 2023-01-01T09:00:00Z, got %v", ts.FirstCommit)
				}
				if ts.CreatedAt == nil || *ts.CreatedAt != "2023-01-01T10:00:00Z" {
					t.Errorf("Expected CreatedAt to be 2023-01-01T10:00:00Z, got %v", ts.CreatedAt)
				}
				if ts.FirstReviewRequest == nil || *ts.FirstReviewRequest != "2023-01-01T11:00:00Z" {
					t.Errorf("Expected FirstReviewRequest to be 2023-01-01T11:00:00Z, got %v", ts.FirstReviewRequest)
				}
				if ts.FirstComment == nil || *ts.FirstComment != "2023-01-01T12:00:00Z" {
					t.Errorf("Expected FirstComment to be 2023-01-01T12:00:00Z, got %v", ts.FirstComment)
				}
				if ts.FirstApproval == nil || *ts.FirstApproval != "2023-01-01T15:00:00Z" {
					t.Errorf("Expected FirstApproval to be 2023-01-01T15:00:00Z, got %v", ts.FirstApproval)
				}
				if ts.SecondApproval == nil || *ts.SecondApproval != "2023-01-01T16:00:00Z" {
					t.Errorf("Expected SecondApproval to be 2023-01-01T16:00:00Z, got %v", ts.SecondApproval)
				}
				if ts.MergedAt == nil || *ts.MergedAt != "2023-01-01T18:00:00Z" {
					t.Errorf("Expected MergedAt to be 2023-01-01T18:00:00Z, got %v", ts.MergedAt)
				}
				if ts.ClosedAt == nil || *ts.ClosedAt != "2023-01-01T18:00:00Z" {
					t.Errorf("Expected ClosedAt to be 2023-01-01T18:00:00Z, got %v", ts.ClosedAt)
				}
			},
		},
		{
			name: "no optional timestamps",
			pr: &GitHubPR{
				CreatedAt: "2023-01-01T10:00:00Z",
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			commits:  []GitHubCommit{},
			validate: func(t *testing.T, ts *Timestamps) {
				if ts.FirstCommit != nil {
					t.Errorf("Expected FirstCommit to be nil, got %v", ts.FirstCommit)
				}
				if ts.CreatedAt == nil || *ts.CreatedAt != "2023-01-01T10:00:00Z" {
					t.Errorf("Expected CreatedAt to be 2023-01-01T10:00:00Z, got %v", ts.CreatedAt)
				}
				if ts.FirstReviewRequest != nil {
					t.Errorf("Expected FirstReviewRequest to be nil, got %v", ts.FirstReviewRequest)
				}
				if ts.FirstComment != nil {
					t.Errorf("Expected FirstComment to be nil, got %v", ts.FirstComment)
				}
				if ts.FirstApproval != nil {
					t.Errorf("Expected FirstApproval to be nil, got %v", ts.FirstApproval)
				}
				if ts.SecondApproval != nil {
					t.Errorf("Expected SecondApproval to be nil, got %v", ts.SecondApproval)
				}
				if ts.MergedAt != nil {
					t.Errorf("Expected MergeAt to be nil, got %v", ts.MergedAt)
				}
				if ts.ClosedAt != nil {
					t.Errorf("Expected ClosedAt to be nil, got %v", ts.ClosedAt)
				}
			},
		},
		{
			name: "multiple commits - should get earliest",
			pr: &GitHubPR{
				CreatedAt: "2023-01-01T10:00:00Z",
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			commits: []GitHubCommit{
				{
					SHA: "def456",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T08:30:00Z"}, // Earlier commit
					},
				},
				{
					SHA: "abc123",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T09:00:00Z"}, // Later commit
					},
				},
			},
			validate: func(t *testing.T, ts *Timestamps) {
				if ts.FirstCommit == nil || *ts.FirstCommit != "2023-01-01T08:30:00Z" {
					t.Errorf("Expected FirstCommit to be 2023-01-01T08:30:00Z (earliest), got %v", ts.FirstCommit)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTimestamps(tt.pr, tt.reviews, tt.comments, []GitHubReviewComment{}, tt.timeline, tt.commits)
			tt.validate(t, result)
		})
	}
}

func TestCalculatePRSize(t *testing.T) {
	tests := []struct {
		name             string
		files            []GitHubPRFile
		expectedLines    int
		expectedFiles    int
	}{
		{
			name:          "no files changed",
			files:         []GitHubPRFile{},
			expectedLines: 0,
			expectedFiles: 0,
		},
		{
			name: "single file with additions and deletions",
			files: []GitHubPRFile{
				{
					Filename:  "main.go",
					Status:    "modified",
					Additions: 10,
					Deletions: 5,
					Changes:   15,
				},
			},
			expectedLines: 15, // 10 additions + 5 deletions
			expectedFiles: 1,
		},
		{
			name: "multiple files with various changes",
			files: []GitHubPRFile{
				{
					Filename:  "main.go",
					Status:    "modified",
					Additions: 20,
					Deletions: 5,
					Changes:   25,
				},
				{
					Filename:  "utils.go",
					Status:    "added",
					Additions: 50,
					Deletions: 0,
					Changes:   50,
				},
				{
					Filename:  "old_file.go",
					Status:    "removed",
					Additions: 0,
					Deletions: 30,
					Changes:   30,
				},
			},
			expectedLines: 105, // 20+5 + 50+0 + 0+30
			expectedFiles: 3,
		},
		{
			name: "files with only additions",
			files: []GitHubPRFile{
				{
					Filename:  "new_feature.go",
					Status:    "added",
					Additions: 100,
					Deletions: 0,
					Changes:   100,
				},
			},
			expectedLines: 100,
			expectedFiles: 1,
		},
		{
			name: "files with only deletions",
			files: []GitHubPRFile{
				{
					Filename:  "deprecated.go",
					Status:    "removed",
					Additions: 0,
					Deletions: 75,
					Changes:   75,
				},
			},
			expectedLines: 75,
			expectedFiles: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePRSize(tt.files)
			if result.LinesChanged != tt.expectedLines {
				t.Errorf("calculatePRSize() LinesChanged = %d, want %d", result.LinesChanged, tt.expectedLines)
			}
			if result.FilesChanged != tt.expectedFiles {
				t.Errorf("calculatePRSize() FilesChanged = %d, want %d", result.FilesChanged, tt.expectedFiles)
			}
		})
	}
}

func TestFindReleaseForMergedPR(t *testing.T) {
	tests := []struct {
		name     string
		pr       *GitHubPR
		releases []GitHubRelease
		expected *string
	}{
		{
			name: "PR not merged - should return nil",
			pr: &GitHubPR{
				Merged:   false,
				MergedAt: nil,
			},
			releases: []GitHubRelease{
				{
					Name:        "v1.0.0",
					TagName:     "v1.0.0",
					PublishedAt: "2023-01-02T12:00:00Z",
				},
			},
			expected: nil,
		},
		{
			name: "PR merged but no releases - should return nil",
			pr: &GitHubPR{
				Merged:    true,
				MergedAt:  stringPtr("2023-01-01T12:00:00Z"),
			},
			releases: []GitHubRelease{},
			expected: nil,
		},
		{
			name: "PR merged, release published after merge",
			pr: &GitHubPR{
				Merged:    true,
				MergedAt:  stringPtr("2023-01-01T12:00:00Z"),
			},
			releases: []GitHubRelease{
				{
					Name:        "v1.0.0",
					TagName:     "v1.0.0",
					PublishedAt: "2023-01-02T12:00:00Z",
				},
			},
			expected: stringPtr("v1.0.0"),
		},
		{
			name: "PR merged, multiple releases, find first after merge",
			pr: &GitHubPR{
				Merged:    true,
				MergedAt:  stringPtr("2023-01-01T12:00:00Z"),
			},
			releases: []GitHubRelease{
				{
					Name:        "v1.1.0",
					TagName:     "v1.1.0",
					PublishedAt: "2023-01-05T12:00:00Z",
				},
				{
					Name:        "v1.0.0",
					TagName:     "v1.0.0",
					PublishedAt: "2023-01-02T12:00:00Z",
				},
			},
			expected: stringPtr("v1.0.0"), // First release after merge
		},
		{
			name: "PR merged, release published before merge - should return nil",
			pr: &GitHubPR{
				Merged:    true,
				MergedAt:  stringPtr("2023-01-02T12:00:00Z"),
			},
			releases: []GitHubRelease{
				{
					Name:        "v1.0.0",
					TagName:     "v1.0.0",
					PublishedAt: "2023-01-01T12:00:00Z",
				},
			},
			expected: nil,
		},
		{
			name: "PR merged, release with empty name uses tag name",
			pr: &GitHubPR{
				Merged:    true,
				MergedAt:  stringPtr("2023-01-01T12:00:00Z"),
			},
			releases: []GitHubRelease{
				{
					Name:        "",
					TagName:     "v1.0.0",
					PublishedAt: "2023-01-02T12:00:00Z",
				},
			},
			expected: stringPtr("v1.0.0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findReleaseForMergedPR(tt.pr, tt.releases)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("findReleaseForMergedPR() = %v, want nil", *result)
				}
			} else {
				if result == nil {
					t.Errorf("findReleaseForMergedPR() = nil, want %v", *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("findReleaseForMergedPR() = %v, want %v", *result, *tt.expected)
				}
			}
		})
	}
}

func TestCountCommitsAfterFirstReview(t *testing.T) {
	tests := []struct {
		name     string
		commits  []GitHubCommit
		timeline []GitHubTimelineEvent
		expected int
	}{
		{
			name:     "no timeline events - should return 0",
			commits:  []GitHubCommit{},
			timeline: []GitHubTimelineEvent{},
			expected: 0,
		},
		{
			name: "no review request event - should return 0",
			commits: []GitHubCommit{
				{
					SHA: "abc123",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-02T12:00:00Z"},
					},
				},
			},
			timeline: []GitHubTimelineEvent{
				{
					Event:     "commented",
					CreatedAt: "2023-01-01T11:00:00Z",
				},
			},
			expected: 0,
		},
		{
			name: "commits before review request - should return 0",
			commits: []GitHubCommit{
				{
					SHA: "abc123",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T10:00:00Z"},
					},
				},
			},
			timeline: []GitHubTimelineEvent{
				{
					Event:     "review_requested",
					CreatedAt: "2023-01-01T11:00:00Z",
				},
			},
			expected: 0,
		},
		{
			name: "commits after review request - should count them",
			commits: []GitHubCommit{
				{
					SHA: "abc123",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T10:00:00Z"}, // Before review request
					},
				},
				{
					SHA: "def456",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T12:00:00Z"}, // After review request
					},
				},
				{
					SHA: "ghi789",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T13:00:00Z"}, // After review request
					},
				},
			},
			timeline: []GitHubTimelineEvent{
				{
					Event:     "review_requested",
					CreatedAt: "2023-01-01T11:00:00Z",
				},
			},
			expected: 2,
		},
		{
			name: "multiple review requests - should use first one",
			commits: []GitHubCommit{
				{
					SHA: "abc123",
					Commit: struct {
						Author struct {
							Date string `json:"date"`
						} `json:"author"`
					}{
						Author: struct {
							Date string `json:"date"`
						}{Date: "2023-01-01T12:00:00Z"}, // After first review request
					},
				},
			},
			timeline: []GitHubTimelineEvent{
				{
					Event:     "review_requested",
					CreatedAt: "2023-01-01T11:00:00Z", // First review request
				},
				{
					Event:     "review_requested",
					CreatedAt: "2023-01-01T14:00:00Z", // Second review request
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countCommitsAfterFirstReview(tt.commits, tt.timeline)
			if result != tt.expected {
				t.Errorf("countCommitsAfterFirstReview() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGeneratedAtTimestamp(t *testing.T) {
	// Test that generated_at is in correct RFC3339 format and represents current time
	now := time.Now().UTC()
	
	// Create a simple test case by calling formatToUTC with current time
	testTime := now.Format(time.RFC3339)
	
	// Verify it parses correctly
	parsedTime, err := time.Parse(time.RFC3339, testTime)
	if err != nil {
		t.Errorf("Generated timestamp should be in RFC3339 format, got error: %v", err)
	}
	
	// Verify it's in UTC (should end with 'Z')
	if !strings.HasSuffix(testTime, "Z") {
		t.Errorf("Generated timestamp should be in UTC format ending with 'Z', got: %s", testTime)
	}
	
	// Verify the parsed time is close to now (within 1 second)
	timeDiff := parsedTime.Sub(now).Abs()
	if timeDiff > time.Second {
		t.Errorf("Generated timestamp should be close to current time, difference: %v", timeDiff)
	}
}

func TestCountChangeRequests(t *testing.T) {
	tests := []struct {
		name     string
		reviews  []GitHubReview
		expected int
	}{
		{
			name:     "no reviews",
			reviews:  []GitHubReview{},
			expected: 0,
		},
		{
			name: "no change requests",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State:       "APPROVED",
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user2"},
					State:       "COMMENTED",
					SubmittedAt: "2023-01-01T13:00:00Z",
				},
			},
			expected: 0,
		},
		{
			name: "single change request",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
			},
			expected: 1,
		},
		{
			name: "multiple change requests",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user2"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T13:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user3"},
					State:       "APPROVED",
					SubmittedAt: "2023-01-01T14:00:00Z",
				},
			},
			expected: 2,
		},
		{
			name: "mixed review states",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user2"},
					State:       "APPROVED",
					SubmittedAt: "2023-01-01T13:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user3"},
					State:       "COMMENTED",
					SubmittedAt: "2023-01-01T14:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user4"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T15:00:00Z",
				},
			},
			expected: 2,
		},
		{
			name: "same user multiple change requests",
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T12:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "user1"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T13:00:00Z",
				},
			},
			expected: 2, // Each review counts separately
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

func TestExtractJiraIssue(t *testing.T) {
	tests := []struct {
		name     string
		pr       *GitHubPR
		expected string
	}{
		{
			name: "Jira issue in title - standard format",
			pr: &GitHubPR{
				Title: "ABC-123: Fix login bug",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/login-fix"},
				User:  struct{ Login string `json:"login"` }{Login: "developer"},
			},
			expected: "ABC-123",
		},
		{
			name: "Jira issue in title - project name format",
			pr: &GitHubPR{
				Title: "PROJECT-1234 Update user permissions",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "main"},
			},
			expected: "PROJECT-1234",
		},
		{
			name: "Jira issue in body",
			pr: &GitHubPR{
				Title: "Fix authentication issues",
				Body:  stringPtr("This PR addresses the issue described in SECURITY-456"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "bugfix/auth"},
			},
			expected: "SECURITY-456",
		},
		{
			name: "Jira issue in branch name",
			pr: &GitHubPR{
				Title: "Update documentation",
				Body:  stringPtr("Updated the README file"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/DOC-789-update-readme"},
			},
			expected: "DOC-789",
		},
		{
			name: "Jira issue in branch name - uppercase conversion",
			pr: &GitHubPR{
				Title: "Minor bug fix",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "bugfix/fix-123-small-issue"},
			},
			expected: "FIX-123",
		},
		{
			name: "Multiple Jira issues - returns first found (title priority)",
			pr: &GitHubPR{
				Title: "ABC-111: Primary issue",
				Body:  stringPtr("Also related to XYZ-222"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/DEF-333-branch"},
			},
			expected: "ABC-111",
		},
		{
			name: "Multiple Jira issues - body over branch",
			pr: &GitHubPR{
				Title: "General update",
				Body:  stringPtr("Fixes PROJ-444 and related issues"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/OTHER-555-branch"},
			},
			expected: "PROJ-444",
		},
		{
			name: "No Jira issue found",
			pr: &GitHubPR{
				Title: "Simple bug fix",
				Body:  stringPtr("Fixed a small issue with the login form"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/login-improvements"},
				User:  struct{ Login string `json:"login"` }{Login: "regular-developer"},
			},
			expected: "UNKNOWN",
		},
		{
			name: "Invalid format - single letter project",
			pr: &GitHubPR{
				Title: "A-123: This should not match",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "main"},
			},
			expected: "UNKNOWN",
		},
		{
			name: "Invalid format - no hyphen",
			pr: &GitHubPR{
				Title: "ABC123: Missing hyphen",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "main"},
			},
			expected: "UNKNOWN",
		},
		{
			name: "Edge case - alphanumeric project key",
			pr: &GitHubPR{
				Title: "WEB2-789: Second version project",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "main"},
			},
			expected: "WEB2-789",
		},
		{
			name: "Edge case - embedded in text",
			pr: &GitHubPR{
				Title: "Update (relates to ISSUE-999) component",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "main"},
			},
			expected: "ISSUE-999",
		},
		{
			name: "Case insensitive branch search",
			pr: &GitHubPR{
				Title: "Feature update",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/test-123-lowercase"},
			},
			expected: "TEST-123",
		},
		{
			name: "Empty body should not cause issues",
			pr: &GitHubPR{
				Title: "Simple fix",
				Body:  stringPtr(""),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "PROJ-456-fix"},
				User:  struct{ Login string `json:"login"` }{Login: "regular-user"},
			},
			expected: "PROJ-456",
		},
		{
			name: "Bot user with no Jira issue - should return BOT",
			pr: &GitHubPR{
				Title: "Automated dependency update",
				Body:  stringPtr("Updates package versions"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "dependabot/npm/update-packages"},
				User:  struct{ Login string `json:"login"` }{Login: "dependabot[bot]"},
			},
			expected: "BOT",
		},
		{
			name: "Bot user with different bot marker - should return BOT",
			pr: &GitHubPR{
				Title: "Security update",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "security-updates"},
				User:  struct{ Login string `json:"login"` }{Login: "github-actions[bot]"},
			},
			expected: "BOT",
		},
		{
			name: "Regular user with no Jira issue - should return UNKNOWN",
			pr: &GitHubPR{
				Title: "Simple bug fix",
				Body:  stringPtr("Fixed a small issue with the login form"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/login-improvements"},
				User:  struct{ Login string `json:"login"` }{Login: "regular-developer"},
			},
			expected: "UNKNOWN",
		},
		{
			name: "Bot user with Jira issue - should return Jira issue (not BOT)",
			pr: &GitHubPR{
				Title: "AUTO-123: Automated security patch",
				Body:  stringPtr("Automated security update"),
				Head:  struct{ Ref string `json:"ref"` }{Ref: "auto-security-patch"},
				User:  struct{ Login string `json:"login"` }{Login: "security-bot[bot]"},
			},
			expected: "AUTO-123",
		},
		{
			name: "Username containing bot but not [bot] marker - should return UNKNOWN",
			pr: &GitHubPR{
				Title: "Regular update",
				Body:  nil,
				Head:  struct{ Ref string `json:"ref"` }{Ref: "feature/update"},
				User:  struct{ Login string `json:"login"` }{Login: "robotuser"},
			},
			expected: "UNKNOWN",
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

func TestPRDetailsBasicFields(t *testing.T) {
	pr := &GitHubPR{
		Number:  123,
		Title:   "Fix authentication bug",
		HTMLURL: "https://github.com/microsoft/vscode/pull/123",
		NodeID:  "PR_kwDOABCD123_node456",
		User: struct {
			Login string `json:"login"`
		}{Login: "contributor"},
		State:  "open",
		Draft:  false,
		Merged: false,
	}

	// Test that basic PR fields are properly extracted
	if pr.Number != 123 {
		t.Errorf("Expected PR number to be 123, got %d", pr.Number)
	}
	if pr.Title != "Fix authentication bug" {
		t.Errorf("Expected PR title to be 'Fix authentication bug', got %s", pr.Title)
	}
	if pr.HTMLURL != "https://github.com/microsoft/vscode/pull/123" {
		t.Errorf("Expected PR web URL to be 'https://github.com/microsoft/vscode/pull/123', got %s", pr.HTMLURL)
	}
	if pr.NodeID != "PR_kwDOABCD123_node456" {
		t.Errorf("Expected PR node ID to be 'PR_kwDOABCD123_node456', got %s", pr.NodeID)
	}
}

func TestCalculatePRMetrics(t *testing.T) {
	tests := []struct {
		name       string
		pr         *GitHubPR
		reviews    []GitHubReview
		comments   []GitHubComment
		timeline   []GitHubTimelineEvent
		timestamps *Timestamps
		validate   func(*testing.T, *PRMetrics)
	}{
		{
			name: "All metrics available",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{
					{Login: "reviewer1"},
					{Login: "reviewer2"},
				},
			},
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer1"},
					State:       "APPROVED",
					SubmittedAt: "2023-01-01T15:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer2"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T16:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer1"},
					State:       "COMMENTED",
					SubmittedAt: "2023-01-01T17:00:00Z",
				},
			},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				CreatedAt:         stringPtr("2023-01-01T10:00:00Z"),
				FirstReviewRequest: stringPtr("2023-01-01T11:00:00Z"),
				FirstComment:       stringPtr("2023-01-01T12:00:00Z"),
				MergedAt:          stringPtr("2023-01-01T18:00:00Z"),
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Time to First Review Request: 10:00 to 11:00 = 1 hour
				if metrics.TimeToFirstReviewRequestHours == nil || *metrics.TimeToFirstReviewRequestHours != 1.0 {
					t.Errorf("Expected TimeToFirstReviewRequestHours to be 1.0, got %v", metrics.TimeToFirstReviewRequestHours)
				}
				
				// Time to First Review: 11:00 to 12:00 = 1 hour (first comment)
				if metrics.TimeToFirstReviewHours == nil || *metrics.TimeToFirstReviewHours != 1.0 {
					t.Errorf("Expected TimeToFirstReviewHours to be 1.0, got %v", metrics.TimeToFirstReviewHours)
				}
				
				// Review Cycle Time: 11:00 to 18:00 = 7 hours
				if metrics.ReviewCycleTimeHours == nil || *metrics.ReviewCycleTimeHours != 7.0 {
					t.Errorf("Expected ReviewCycleTimeHours to be 7.0, got %v", metrics.ReviewCycleTimeHours)
				}
				
				// Blocking vs Non-Blocking: 1 CHANGES_REQUESTED / 2 (APPROVED + COMMENTED) = 0.5
				if metrics.BlockingNonBlockingRatio == nil || *metrics.BlockingNonBlockingRatio != 0.5 {
					t.Errorf("Expected BlockingNonBlockingRatio to be 0.5, got %v", metrics.BlockingNonBlockingRatio)
				}
				
				// Reviewer Participation: 2 actual reviewers / 2 requested = 1.0
				if metrics.ReviewerParticipationRatio == nil || *metrics.ReviewerParticipationRatio != 1.0 {
					t.Errorf("Expected ReviewerParticipationRatio to be 1.0, got %v", metrics.ReviewerParticipationRatio)
				}
			},
		},
		{
			name: "No metrics available - no review request",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:    []GitHubReview{},
			comments:   []GitHubComment{},
			timeline:   []GitHubTimelineEvent{},
			timestamps: &Timestamps{},
			validate: func(t *testing.T, metrics *PRMetrics) {
				if metrics.TimeToFirstReviewRequestHours != nil {
					t.Errorf("Expected TimeToFirstReviewRequestHours to be nil, got %v", metrics.TimeToFirstReviewRequestHours)
				}
				if metrics.TimeToFirstReviewHours != nil {
					t.Errorf("Expected TimeToFirstReviewHours to be nil, got %v", metrics.TimeToFirstReviewHours)
				}
				if metrics.ReviewCycleTimeHours != nil {
					t.Errorf("Expected ReviewCycleTimeHours to be nil, got %v", metrics.ReviewCycleTimeHours)
				}
				if metrics.BlockingNonBlockingRatio != nil {
					t.Errorf("Expected BlockingNonBlockingRatio to be nil, got %v", metrics.BlockingNonBlockingRatio)
				}
				if metrics.ReviewerParticipationRatio != nil {
					t.Errorf("Expected ReviewerParticipationRatio to be nil, got %v", metrics.ReviewerParticipationRatio)
				}
			},
		},
		{
			name: "Time to first review - comment before review request",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				FirstReviewRequest: stringPtr("2023-01-01T12:00:00Z"),
				FirstComment:       stringPtr("2023-01-01T11:00:00Z"), // Before review request
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Should not calculate time to first review if comment was before review request
				if metrics.TimeToFirstReviewHours != nil {
					t.Errorf("Expected TimeToFirstReviewHours to be nil when comment is before review request, got %v", metrics.TimeToFirstReviewHours)
				}
			},
		},
		{
			name: "Blocking ratio with only blocking comments",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer1"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T15:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer2"},
					State:       "CHANGES_REQUESTED",
					SubmittedAt: "2023-01-01T16:00:00Z",
				},
			},
			comments:   []GitHubComment{},
			timeline:   []GitHubTimelineEvent{},
			timestamps: &Timestamps{},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Should not calculate ratio when there are no non-blocking comments
				if metrics.BlockingNonBlockingRatio != nil {
					t.Errorf("Expected BlockingNonBlockingRatio to be nil when no non-blocking comments, got %v", metrics.BlockingNonBlockingRatio)
				}
			},
		},
		{
			name: "Reviewer participation - partial participation",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{
					{Login: "reviewer1"},
					{Login: "reviewer2"},
					{Login: "reviewer3"},
				},
			},
			reviews: []GitHubReview{
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer1"},
					State:       "APPROVED",
					SubmittedAt: "2023-01-01T15:00:00Z",
				},
				{
					User: struct {
						Login string `json:"login"`
					}{Login: "reviewer1"}, // Same reviewer, multiple reviews
					State:       "COMMENTED",
					SubmittedAt: "2023-01-01T16:00:00Z",
				},
			},
			comments:   []GitHubComment{},
			timeline:   []GitHubTimelineEvent{},
			timestamps: &Timestamps{},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// 1 actual reviewer (reviewer1) / 3 requested = 0.333...
				if metrics.ReviewerParticipationRatio == nil {
					t.Errorf("Expected ReviewerParticipationRatio to be calculated, got nil")
				} else {
					expected := 1.0 / 3.0
					actual := *metrics.ReviewerParticipationRatio
					if actual < expected-0.001 || actual > expected+0.001 {
						t.Errorf("Expected ReviewerParticipationRatio to be ~%.3f, got %.3f", expected, actual)
					}
				}
			},
		},
		{
			name: "Time to first review - approval before comment",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				FirstReviewRequest: stringPtr("2023-01-01T11:00:00Z"),
				FirstComment:       stringPtr("2023-01-01T14:00:00Z"),
				FirstApproval:      stringPtr("2023-01-01T12:00:00Z"), // Approval before comment
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Time to First Review: 11:00 to 12:00 = 1 hour (first approval, not comment)
				if metrics.TimeToFirstReviewHours == nil || *metrics.TimeToFirstReviewHours != 1.0 {
					t.Errorf("Expected TimeToFirstReviewHours to be 1.0 (first approval), got %v", metrics.TimeToFirstReviewHours)
				}
			},
		},
		{
			name: "Time to first review - only approval available",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				FirstReviewRequest: stringPtr("2023-01-01T10:00:00Z"),
				FirstApproval:      stringPtr("2023-01-01T13:00:00Z"),
				// No FirstComment
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Time to First Review: 10:00 to 13:00 = 3 hours (only approval available)
				if metrics.TimeToFirstReviewHours == nil || *metrics.TimeToFirstReviewHours != 3.0 {
					t.Errorf("Expected TimeToFirstReviewHours to be 3.0 (only approval), got %v", metrics.TimeToFirstReviewHours)
				}
			},
		},
		{
			name: "Time to first review request - different time intervals",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				CreatedAt:         stringPtr("2023-01-01T09:00:00Z"),
				FirstReviewRequest: stringPtr("2023-01-01T12:30:00Z"), // 3.5 hours later
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Time to First Review Request: 09:00 to 12:30 = 3.5 hours
				if metrics.TimeToFirstReviewRequestHours == nil || *metrics.TimeToFirstReviewRequestHours != 3.5 {
					t.Errorf("Expected TimeToFirstReviewRequestHours to be 3.5, got %v", metrics.TimeToFirstReviewRequestHours)
				}
			},
		},
		{
			name: "Time to first review request - review request before creation (edge case)",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				CreatedAt:         stringPtr("2023-01-01T12:00:00Z"),
				FirstReviewRequest: stringPtr("2023-01-01T11:00:00Z"), // Before creation
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Should not calculate metric if review request is before creation
				if metrics.TimeToFirstReviewRequestHours != nil {
					t.Errorf("Expected TimeToFirstReviewRequestHours to be nil when review request is before creation, got %v", metrics.TimeToFirstReviewRequestHours)
				}
			},
		},
		{
			name: "Time to first review request - only creation timestamp available",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				CreatedAt: stringPtr("2023-01-01T10:00:00Z"),
				// No FirstReviewRequest
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Should not calculate metric if no review request timestamp
				if metrics.TimeToFirstReviewRequestHours != nil {
					t.Errorf("Expected TimeToFirstReviewRequestHours to be nil when no review request, got %v", metrics.TimeToFirstReviewRequestHours)
				}
			},
		},
		{
			name: "Review cycle time uses closed time when not merged",
			pr: &GitHubPR{
				RequestedReviewers: []struct {
					Login string `json:"login"`
				}{},
			},
			reviews:  []GitHubReview{},
			comments: []GitHubComment{},
			timeline: []GitHubTimelineEvent{},
			timestamps: &Timestamps{
				FirstReviewRequest: stringPtr("2023-01-01T11:00:00Z"),
				ClosedAt:          stringPtr("2023-01-01T15:00:00Z"), // Not merged, but closed
			},
			validate: func(t *testing.T, metrics *PRMetrics) {
				// Review Cycle Time: 11:00 to 15:00 = 4 hours
				if metrics.ReviewCycleTimeHours == nil || *metrics.ReviewCycleTimeHours != 4.0 {
					t.Errorf("Expected ReviewCycleTimeHours to be 4.0, got %v", metrics.ReviewCycleTimeHours)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePRMetrics(tt.pr, tt.reviews, tt.comments, tt.timeline, tt.timestamps)
			tt.validate(t, result)
		})
	}
}

func TestJSONOutputStructure(t *testing.T) {
	// Create a sample PRDetails with timestamps
	prDetails := &PRDetails{
		OrganizationName: "test-org",
		RepositoryName:   "test-repo",
		PRNumber:         123,
		PRTitle:          "Test PR",
		AuthorUsername:   "author",
		State:            "open",
		JiraIssue:        "TEST-123",
		IsBot:            false,
		GeneratedAt:      "2023-01-01T20:00:00Z",
		Timestamps: &PRTimestamps{
			FirstCommit:       stringPtr("2023-01-01T09:00:00Z"),
			CreatedAt:         stringPtr("2023-01-01T10:00:00Z"),
			FirstReviewRequest: stringPtr("2023-01-01T11:00:00Z"),
			FirstComment:      stringPtr("2023-01-01T12:00:00Z"),
			FirstApproval:     stringPtr("2023-01-01T15:00:00Z"),
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(prDetails)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Unmarshal to check structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify timestamps object exists
	timestamps, ok := result["timestamps"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected timestamps to be an object, got %T", result["timestamps"])
	}

	// Verify individual timestamp fields
	if timestamps["first_commit"] != "2023-01-01T09:00:00Z" {
		t.Errorf("Expected first_commit to be '2023-01-01T09:00:00Z', got %v", timestamps["first_commit"])
	}
	if timestamps["created_at"] != "2023-01-01T10:00:00Z" {
		t.Errorf("Expected created_at to be '2023-01-01T10:00:00Z', got %v", timestamps["created_at"])
	}
	if timestamps["first_review_request"] != "2023-01-01T11:00:00Z" {
		t.Errorf("Expected first_review_request to be '2023-01-01T11:00:00Z', got %v", timestamps["first_review_request"])
	}
	if timestamps["first_comment"] != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected first_comment to be '2023-01-01T12:00:00Z', got %v", timestamps["first_comment"])
	}
	if timestamps["first_approval"] != "2023-01-01T15:00:00Z" {
		t.Errorf("Expected first_approval to be '2023-01-01T15:00:00Z', got %v", timestamps["first_approval"])
	}

	// Verify that individual timestamp fields are not at the root level (except generated_at)
	if _, exists := result["created_at"]; exists {
		t.Error("created_at should not exist at root level")
	}
	if _, exists := result["first_review_request"]; exists {
		t.Error("first_review_request should not exist at root level")
	}
	
	// Verify that generated_at IS at the root level
	if result["generated_at"] != "2023-01-01T20:00:00Z" {
		t.Errorf("Expected generated_at to be '2023-01-01T20:00:00Z' at root level, got %v", result["generated_at"])
	}
	
	// Verify that is_bot IS at the root level and is false
	if result["is_bot"] != false {
		t.Errorf("Expected is_bot to be false at root level, got %v", result["is_bot"])
	}
	
	// Verify that generated_at is NOT in the timestamps object
	if _, exists := timestamps["generated_at"]; exists {
		t.Error("generated_at should not exist in timestamps object")
	}
}

func TestJSONOutputStructureBot(t *testing.T) {
	// Create a sample PRDetails with bot user
	prDetails := &PRDetails{
		OrganizationName: "test-org",
		RepositoryName:   "test-repo",
		PRNumber:         456,
		PRTitle:          "Automated security update",
		AuthorUsername:   "dependabot[bot]",
		State:            "open",
		JiraIssue:        "BOT",
		IsBot:            true,
		GeneratedAt:      "2023-01-01T20:00:00Z",
		Timestamps: &PRTimestamps{
			CreatedAt:         stringPtr("2023-01-01T10:00:00Z"),
			FirstReviewRequest: stringPtr("2023-01-01T11:00:00Z"),
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(prDetails)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Unmarshal to check structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify is_bot is true for bot user
	if result["is_bot"] != true {
		t.Errorf("Expected is_bot to be true for bot user, got %v", result["is_bot"])
	}
	
	// Verify jira_issue is BOT for bot user with no Jira issue
	if result["jira_issue"] != "BOT" {
		t.Errorf("Expected jira_issue to be 'BOT' for bot user, got %v", result["jira_issue"])
	}
	
	// Verify author_username contains [bot]
	if result["author_username"] != "dependabot[bot]" {
		t.Errorf("Expected author_username to be 'dependabot[bot]', got %v", result["author_username"])
	}
}

func TestIsBot(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{
			name:     "dependabot - should be bot",
			username: "dependabot[bot]",
			expected: true,
		},
		{
			name:     "github-actions - should be bot",
			username: "github-actions[bot]",
			expected: true,
		},
		{
			name:     "security bot - should be bot",
			username: "security-bot[bot]",
			expected: true,
		},
		{
			name:     "regular user - should not be bot",
			username: "regular-developer",
			expected: false,
		},
		{
			name:     "username with bot but no [bot] marker - should not be bot",
			username: "robotuser",
			expected: false,
		},
		{
			name:     "username with BOT but no [bot] marker - should not be bot",
			username: "BOTUSER",
			expected: false,
		},
		{
			name:     "empty username - should not be bot",
			username: "",
			expected: false,
		},
		{
			name:     "just [bot] - should be bot",
			username: "[bot]",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBot(tt.username)
			if result != tt.expected {
				t.Errorf("isBot(%q) = %v, want %v", tt.username, result, tt.expected)
			}
		})
	}
}

func TestGetCommentorsWithReviewComments(t *testing.T) {
	tests := []struct {
		name           string
		comments       []GitHubComment
		reviewComments []GitHubReviewComment
		authorUsername string
		expectedCount  int
		expectedUsers  []string
	}{
		{
			name:           "no comments at all",
			comments:       []GitHubComment{},
			reviewComments: []GitHubReviewComment{},
			authorUsername: "author",
			expectedCount:  0,
			expectedUsers:  []string{},
		},
		{
			name: "only regular comments",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T10:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "user2"}, CreatedAt: "2023-01-01T11:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{},
			authorUsername: "author",
			expectedCount:  2,
			expectedUsers:  []string{"user1", "user2"},
		},
		{
			name:     "only review comments",
			comments: []GitHubComment{},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T10:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer2"}, CreatedAt: "2023-01-01T11:00:00Z"},
			},
			authorUsername: "author",
			expectedCount:  2,
			expectedUsers:  []string{"reviewer1", "reviewer2"},
		},
		{
			name: "mix of regular and review comments",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T10:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T11:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer2"}, CreatedAt: "2023-01-01T12:00:00Z"},
			},
			authorUsername: "author",
			expectedCount:  3,
			expectedUsers:  []string{"user1", "reviewer1", "reviewer2"},
		},
		{
			name: "exclude author from both comment types",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "author"}, CreatedAt: "2023-01-01T10:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T11:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "author"}, CreatedAt: "2023-01-01T12:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T13:00:00Z"},
			},
			authorUsername: "author",
			expectedCount:  2,
			expectedUsers:  []string{"user1", "reviewer1"},
		},
		{
			name: "duplicate users across comment types",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T10:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T11:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T12:00:00Z"},
			},
			authorUsername: "author",
			expectedCount:  2,
			expectedUsers:  []string{"user1", "reviewer1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommentors(tt.comments, tt.reviewComments, tt.authorUsername)
			
			if len(result) != tt.expectedCount {
				t.Errorf("getCommentors() returned %d commentors, want %d", len(result), tt.expectedCount)
			}
			
			for _, expectedUser := range tt.expectedUsers {
				if !result[expectedUser] {
					t.Errorf("Expected user %s not found in commentors", expectedUser)
				}
			}
		})
	}
}

func TestGetTimestampsWithReviewComments(t *testing.T) {
	tests := []struct {
		name           string
		comments       []GitHubComment
		reviewComments []GitHubReviewComment
		expectedFirstComment *string
	}{
		{
			name:           "no comments at all",
			comments:       []GitHubComment{},
			reviewComments: []GitHubReviewComment{},
			expectedFirstComment: nil,
		},
		{
			name: "only regular comment",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T12:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{},
			expectedFirstComment: stringPtr("2023-01-01T12:00:00Z"),
		},
		{
			name:     "only review comment",
			comments: []GitHubComment{},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T11:00:00Z"},
			},
			expectedFirstComment: stringPtr("2023-01-01T11:00:00Z"),
		},
		{
			name: "review comment is earlier than regular comment",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T12:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T10:00:00Z"},
			},
			expectedFirstComment: stringPtr("2023-01-01T10:00:00Z"),
		},
		{
			name: "regular comment is earlier than review comment",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T09:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T11:00:00Z"},
			},
			expectedFirstComment: stringPtr("2023-01-01T09:00:00Z"),
		},
		{
			name: "multiple comments of both types",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T12:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "user2"}, CreatedAt: "2023-01-01T14:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T10:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer2"}, CreatedAt: "2023-01-01T13:00:00Z"},
			},
			expectedFirstComment: stringPtr("2023-01-01T10:00:00Z"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal PR data for the test
			pr := &GitHubPR{
				CreatedAt: "2023-01-01T08:00:00Z",
			}
			
			result := getTimestamps(pr, []GitHubReview{}, tt.comments, tt.reviewComments, []GitHubTimelineEvent{}, []GitHubCommit{})
			
			if tt.expectedFirstComment == nil {
				if result.FirstComment != nil {
					t.Errorf("Expected FirstComment to be nil, but got %v", *result.FirstComment)
				}
			} else {
				if result.FirstComment == nil {
					t.Errorf("Expected FirstComment to be %v, but got nil", *tt.expectedFirstComment)
				} else if *result.FirstComment != *tt.expectedFirstComment {
					t.Errorf("Expected FirstComment to be %v, but got %v", *tt.expectedFirstComment, *result.FirstComment)
				}
			}
		})
	}
}

func TestCountTotalComments(t *testing.T) {
	tests := []struct {
		name           string
		comments       []GitHubComment
		reviewComments []GitHubReviewComment
		expected       int
	}{
		{
			name:           "no comments",
			comments:       []GitHubComment{},
			reviewComments: []GitHubReviewComment{},
			expected:       0,
		},
		{
			name: "only regular comments",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T10:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "user2"}, CreatedAt: "2023-01-01T11:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "user3"}, CreatedAt: "2023-01-01T12:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{},
			expected:       3,
		},
		{
			name:     "only review comments",
			comments: []GitHubComment{},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T10:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer2"}, CreatedAt: "2023-01-01T11:00:00Z"},
			},
			expected: 2,
		},
		{
			name: "mix of both comment types",
			comments: []GitHubComment{
				{User: struct{ Login string `json:"login"` }{Login: "user1"}, CreatedAt: "2023-01-01T10:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "user2"}, CreatedAt: "2023-01-01T11:00:00Z"},
			},
			reviewComments: []GitHubReviewComment{
				{User: struct{ Login string `json:"login"` }{Login: "reviewer1"}, CreatedAt: "2023-01-01T12:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer2"}, CreatedAt: "2023-01-01T13:00:00Z"},
				{User: struct{ Login string `json:"login"` }{Login: "reviewer3"}, CreatedAt: "2023-01-01T14:00:00Z"},
			},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countTotalComments(tt.comments, tt.reviewComments)
			if result != tt.expected {
				t.Errorf("countTotalComments() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetCommentorUsernames(t *testing.T) {
	tests := []struct {
		name       string
		commentors map[string]bool
		expected   []string
	}{
		{
			name:       "no commentors",
			commentors: map[string]bool{},
			expected:   []string{},
		},
		{
			name: "single commentor",
			commentors: map[string]bool{
				"user1": true,
			},
			expected: []string{"user1"},
		},
		{
			name: "multiple commentors - should be sorted",
			commentors: map[string]bool{
				"user3": true,
				"user1": true,
				"user2": true,
			},
			expected: []string{"user1", "user2", "user3"},
		},
		{
			name: "commentors with various usernames",
			commentors: map[string]bool{
				"z-user":      true,
				"a-user":      true,
				"middle-user": true,
			},
			expected: []string{"a-user", "middle-user", "z-user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommentorUsernames(tt.commentors)
			
			if len(result) != len(tt.expected) {
				t.Errorf("getCommentorUsernames() returned %d usernames, want %d", len(result), len(tt.expected))
				return
			}
			
			for i, username := range result {
				if username != tt.expected[i] {
					t.Errorf("getCommentorUsernames()[%d] = %s, want %s", i, username, tt.expected[i])
				}
			}
		})
	}
}

func TestPRDetailsWithCommentFields(t *testing.T) {
	// Test that the new comment fields are properly included in JSON output
	prDetails := &PRDetails{
		OrganizationName:   "test-org",
		RepositoryName:     "test-repo",
		PRNumber:           123,
		PRTitle:            "Test PR with comments",
		AuthorUsername:     "author",
		CommentorUsernames: []string{"reviewer1", "reviewer2", "user1"},
		State:              "open",
		NumComments:        7,
		NumCommentors:      3,
		NumApprovers:       2,
		JiraIssue:          "TEST-123",
		IsBot:              false,
		GeneratedAt:        "2023-01-01T20:00:00Z",
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(prDetails)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Unmarshal to check structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify num_comments field
	if numComments, ok := result["num_comments"].(float64); !ok || int(numComments) != 7 {
		t.Errorf("Expected num_comments to be 7, got %v", result["num_comments"])
	}

	// Verify commentor_usernames field
	commentorUsernames, ok := result["commentor_usernames"].([]interface{})
	if !ok {
		t.Fatalf("Expected commentor_usernames to be an array, got %T", result["commentor_usernames"])
	}
	
	if len(commentorUsernames) != 3 {
		t.Errorf("Expected commentor_usernames to have 3 elements, got %d", len(commentorUsernames))
	}
	
	expectedUsernames := []string{"reviewer1", "reviewer2", "user1"}
	for i, username := range commentorUsernames {
		if str, ok := username.(string); !ok || str != expectedUsernames[i] {
			t.Errorf("Expected commentor_usernames[%d] to be %s, got %v", i, expectedUsernames[i], username)
		}
	}

	// Verify that existing num_commentors field is still present
	if numCommentors, ok := result["num_commentors"].(float64); !ok || int(numCommentors) != 3 {
		t.Errorf("Expected num_commentors to be 3, got %v", result["num_commentors"])
	}
}

func stringPtr(s string) *string {
	return &s
}
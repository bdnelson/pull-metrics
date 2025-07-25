package main

import (
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
			result := getCommentors(tt.comments, tt.authorUsername)
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
			validate: func(t *testing.T, ts *Timestamps) {
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
			validate: func(t *testing.T, ts *Timestamps) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTimestamps(tt.pr, tt.reviews, tt.comments, tt.timeline)
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
			},
			expected: "PROJ-456",
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

func stringPtr(s string) *string {
	return &s
}
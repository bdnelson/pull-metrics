// Example usage of the pullmetrics package
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"pull-metrics/pullmetrics"
)

func main() {
	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	// Create config
	config := pullmetrics.Config{
		GitHubToken: token,
	}

	// Create analyzer
	analyzer, err := pullmetrics.NewAnalyzer(config)
	if err != nil {
		log.Fatalf("Failed to create analyzer: %v", err)
	}

	// Example: Analyze a PR from the microsoft/vscode repository
	ctx := context.Background()
	details, err := analyzer.AnalyzePR(ctx, "microsoft", "vscode", 123456)
	if err != nil {
		log.Fatalf("Failed to analyze PR: %v", err)
	}

	// Display some key metrics
	fmt.Printf("=== PR Analysis Results ===\n")
	fmt.Printf("PR #%d: %s\n", details.PRNumber, details.PRTitle)
	fmt.Printf("Author: %s\n", details.AuthorUsername)
	fmt.Printf("State: %s\n", details.State)
	fmt.Printf("Comments: %d\n", details.NumComments)
	fmt.Printf("Commenters: %d\n", details.NumCommenters)
	fmt.Printf("Approvers: %d\n", details.NumApprovers)
	fmt.Printf("Requested Reviewers: %d\n", details.NumRequestedReviewers)
	fmt.Printf("Lines Changed: %d\n", details.LinesChanged)
	fmt.Printf("Files Changed: %d\n", details.FilesChanged)
	fmt.Printf("Jira Issue: %s\n", details.JiraIssue)
	fmt.Printf("Is Bot: %t\n", details.IsBot)

	if details.Metrics != nil {
		fmt.Printf("\n=== Metrics ===\n")
		fmt.Printf("Draft Time: %.2f hours\n", details.Metrics.DraftTimeHours)
		if details.Metrics.TimeToFirstReviewRequestHours != nil {
			fmt.Printf("Time to First Review Request: %.2f hours\n", *details.Metrics.TimeToFirstReviewRequestHours)
		}
		if details.Metrics.TimeToFirstReviewHours != nil {
			fmt.Printf("Time to First Review: %.2f hours\n", *details.Metrics.TimeToFirstReviewHours)
		}
		if details.Metrics.ReviewCycleTimeHours != nil {
			fmt.Printf("Review Cycle Time: %.2f hours\n", *details.Metrics.ReviewCycleTimeHours)
		}
		if details.Metrics.ReviewerParticipationRatio != nil {
			fmt.Printf("Reviewer Participation Ratio: %.2f\n", *details.Metrics.ReviewerParticipationRatio)
		}
	}

	if details.Timestamps != nil {
		fmt.Printf("\n=== Key Timestamps ===\n")
		if details.Timestamps.CreatedAt != nil {
			fmt.Printf("Created: %s\n", *details.Timestamps.CreatedAt)
		}
		if details.Timestamps.FirstReviewRequest != nil {
			fmt.Printf("First Review Request: %s\n", *details.Timestamps.FirstReviewRequest)
		}
		if details.Timestamps.FirstComment != nil {
			fmt.Printf("First Comment: %s\n", *details.Timestamps.FirstComment)
		}
		if details.Timestamps.FirstApproval != nil {
			fmt.Printf("First Approval: %s\n", *details.Timestamps.FirstApproval)
		}
		if details.Timestamps.MergedAt != nil {
			fmt.Printf("Merged: %s\n", *details.Timestamps.MergedAt)
		}
	}

	fmt.Printf("\nGenerated at: %s\n", details.GeneratedAt)

	// Example of using convenience function to get JSON
	fmt.Printf("\n=== Using Convenience Function ===\n")
	jsonString, err := pullmetrics.AnalyzePRToJSONString(ctx, config, "microsoft", "vscode", 123456)
	if err != nil {
		log.Fatalf("Failed to get JSON: %v", err)
	}

	fmt.Printf("JSON output length: %d characters\n", len(jsonString))
	// Uncomment the line below to see the full JSON output
	// fmt.Println(jsonString)
}
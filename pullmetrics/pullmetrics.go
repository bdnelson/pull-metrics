// Package pullmetrics provides functionality to analyze GitHub Pull Requests
// and generate comprehensive metrics and details.
//
// This package allows you to analyze GitHub Pull Requests programmatically,
// extracting detailed information including review metrics, timestamps,
// and performance indicators.
//
// Example usage:
//
//	config := pullmetrics.Config{
//		GitHubToken: "your_github_token",
//	}
//
//	analyzer, err := pullmetrics.NewAnalyzer(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	details, err := analyzer.AnalyzePR(context.Background(), "microsoft", "vscode", 12345)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("PR %d has %d comments and %d approvers\n",
//		details.PRNumber, details.NumComments, details.NumApprovers)
package pullmetrics

import (
	"context"
	"encoding/json"
	"fmt"
)

// AnalyzePRToJSON is a convenience function that analyzes a PR and returns JSON output
func AnalyzePRToJSON(ctx context.Context, config Config, org, repo string, prNumber int) ([]byte, error) {
	analyzer, err := NewAnalyzer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}

	details, err := analyzer.AnalyzePR(ctx, org, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze PR: %w", err)
	}

	jsonOutput, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonOutput, nil
}

// AnalyzePRToJSONString is a convenience function that analyzes a PR and returns JSON as a string
func AnalyzePRToJSONString(ctx context.Context, config Config, org, repo string, prNumber int) (string, error) {
	jsonOutput, err := AnalyzePRToJSON(ctx, config, org, repo, prNumber)
	if err != nil {
		return "", err
	}
	
	return string(jsonOutput), nil
}
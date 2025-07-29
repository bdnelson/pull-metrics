package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ardanlabs/conf/v3"
	"github.com/joho/godotenv"
	
	"pull-metrics/pullmetrics"
)

// Config represents the application configuration from command line arguments and environment variables
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

	// Create pullmetrics config
	pmConfig := pullmetrics.Config{
		GitHubToken: cfg.GitHubToken,
	}

	// Use the convenience function to get JSON output
	ctx := context.Background()
	jsonOutput, err := pullmetrics.AnalyzePRToJSONString(ctx, pmConfig, cfg.Organization, cfg.Repository, cfg.PRNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing PR: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(jsonOutput)
}
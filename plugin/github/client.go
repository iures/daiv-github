package github

import (
	"fmt"
	"strings"
	"time"

	externalGithub "github.com/google/go-github/v68/github"
	plug "github.com/iures/daivplug"
)

// GitHubConfig represents the configuration for the GitHub client
type GitHubConfig struct {
	Username     string
	Token        string
	Organization string
	Repositories []string
	QueryOptions QueryOptions
	Format       string
}

// GitHubClient provides a client for interacting with GitHub
type GitHubClient struct {
	client     *externalGithub.Client
	config     *GitHubConfig
	repository GitHubRepository
}

// NewGitHubClient creates a new GitHubClient
func NewGitHubClient(config *GitHubConfig) (*GitHubClient, error) {
	if config.Username == "" {
		return nil, NewValidationError("username is required", nil)
	}
	
	if config.Token == "" {
		return nil, NewValidationError("token is required", nil)
	}
	
	authToken := externalGithub.BasicAuthTransport{
		Username: config.Username,
		Password: config.Token,
	}
	
	client := externalGithub.NewClient(authToken.Client())
	
	githubClient := &GitHubClient{
		client: client,
		config: config,
	}
	
	// Create the repository
	repository := NewGitHubAPIRepository(client, config.Username)
	githubClient.repository = repository
	
	return githubClient, nil
}

// GetRepository returns the GitHub repository
func (g *GitHubClient) GetRepository() GitHubRepository {
	return g.repository
}

// GetConfig returns the client configuration
func (g *GitHubClient) GetConfig() *GitHubConfig {
	return g.config
}

// GetStandupContext generates a standup context report for the given time range
func (g *GitHubClient) GetStandupContext(timeRange plug.TimeRange) (string, error) {
	// Create a context with timeout
	ctx, cancel := NewContextWithTimeout(30 * time.Second)
	defer cancel()

	// Add context values
	ctx = WithUsername(ctx, g.config.Username)
	ctx = WithOrganization(ctx, g.config.Organization)
	ctx = WithTimeRange(ctx, TimeRange{Start: timeRange.Start, End: timeRange.End})

	var report strings.Builder

	for _, repo := range g.config.Repositories {
		repoHasContent := false
		repoSection := &strings.Builder{}
		fmt.Fprintf(repoSection, "\n# Repository: %s\n", repo)

		// Create a repository-specific context
		repoCtx := WithRepository(ctx, repo)

		// Get pull requests for the repository
		pullRequests, err := g.repository.GetPullRequests(
			repoCtx,
			g.config.Organization, 
			repo, 
			TimeRange{Start: timeRange.Start, End: timeRange.End}, 
			g.config.QueryOptions,
		)
		if err != nil {
			return "", NewAPIError(
				fmt.Sprintf("error getting pull requests for %s/%s", g.config.Organization, repo),
				err,
			)
		}

		// Process authored pull requests
		if g.config.QueryOptions.IncludeAuthored {
			authoredPRs := filterAuthoredPullRequests(pullRequests, g.config.Username)
			if len(authoredPRs) > 0 {
				repoHasContent = true
				fmt.Fprintf(repoSection, "\n## Authored Pull Requests\n\n")
				for _, pr := range authoredPRs {
					fmt.Fprintf(repoSection, "- [%s](%s) - %s\n", pr.Title, pr.URL, pr.State)
					if len(pr.Commits) > 0 {
						fmt.Fprintf(repoSection, "  Commits:\n")
						for _, commit := range pr.Commits {
							fmt.Fprintf(repoSection, "  - %s: %s\n", commit.SHA[:7], commit.Message)
						}
					}
				}
			}
		}

		// Process reviewed pull requests
		if g.config.QueryOptions.IncludeReviewed {
			reviewedPRs := filterReviewedPullRequests(pullRequests)
			if len(reviewedPRs) > 0 {
				repoHasContent = true
				fmt.Fprintf(repoSection, "\n## Reviewed Pull Requests\n\n")
				for _, pr := range reviewedPRs {
					fmt.Fprintf(repoSection, "- [%s](%s) - %s\n", pr.Title, pr.URL, pr.State)
					if len(pr.Reviews) > 0 {
						fmt.Fprintf(repoSection, "  Reviews:\n")
						for _, review := range pr.Reviews {
							fmt.Fprintf(repoSection, "  - %s: %s\n", review.State, review.Body)
						}
					}
				}
			}
		}

		if repoHasContent {
			report.WriteString(repoSection.String())
		}
	}

	if report.Len() == 0 {
		return "No GitHub activity found for the specified time range.", nil
	}

	return report.String(), nil
}

// Helper functions for filtering pull requests
func filterAuthoredPullRequests(prs []PullRequest, username string) []PullRequest {
	var result []PullRequest
	for _, pr := range prs {
		if pr.Author == username {
			result = append(result, pr)
		}
	}
	return result
}

func filterReviewedPullRequests(prs []PullRequest) []PullRequest {
	var result []PullRequest
	for _, pr := range prs {
		if pr.IsReviewed {
			result = append(result, pr)
		}
	}
	return result
}

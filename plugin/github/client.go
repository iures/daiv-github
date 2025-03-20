package github

import (
	externalGithub "github.com/google/go-github/v68/github"
)

// GitHubConfig represents the configuration for the GitHub client
type GitHubConfig struct {
	Username     string
	Token        string
	Organization string
	Repositories []string
	Format       string
}

// GitHubClient provides a client for interacting with GitHub
type GitHubClient struct {
	client     *externalGithub.Client
	config     *GitHubConfig
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
	return githubClient, nil
}

// GetConfig returns the client configuration
func (g *GitHubClient) GetConfig() *GitHubConfig {
	return g.config
}

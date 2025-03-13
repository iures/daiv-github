package internal

import (
	"github.com/google/go-github/v68/github"
)

type GithubClientSettings struct {
	Username string
	Token string
	Org string
	Repos []string
}

type GithubClient struct {
	Client *github.Client
	Settings GithubClientSettings
}

// NewGithubClient creates a new GithubClient instance
func NewGithubClient() *GithubClient {
	return &GithubClient{}
}

func (gc *GithubClient) Init(settings GithubClientSettings) {
	authToken := github.BasicAuthTransport{
		Username: settings.Username,
		Password: settings.Token,
	}

	gc.Client = github.NewClient(authToken.Client())
	gc.Settings = settings
}

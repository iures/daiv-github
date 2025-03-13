package plugin

import (
	"fmt"
	"os/exec"
	"strings"

	"daiv-github/plugin/contexts"
	"daiv-github/plugin/github"

	plug "github.com/iures/daivplug"
)

type GitHubPlugin struct {
	client *github.GithubClient
}

func New() *GitHubPlugin {
	return &GitHubPlugin{
		client: &github.GithubClient{},
	}
}

func (g *GitHubPlugin) Name() string {
	return "github"
}

func (g *GitHubPlugin) Manifest() *plug.PluginManifest {
	return &plug.PluginManifest{
		ConfigKeys: []plug.ConfigKey{
			{
				Type:        plug.ConfigTypeString,
				Key:         "github.username",
				Name:        "GitHub Username",
				Description: "Your GitHub username",
				Required:    true,
			},
			{
				Type:        plug.ConfigTypeString,
				Key:         "github.organization",
				Name:        "GitHub Organization",
				Description: "The GitHub organization to monitor",
				Required:    true,
			},
			{
				Type:        plug.ConfigTypeMultiline,
				Key:         "github.repositories",
				Name:        "GitHub Repositories",
				Description: "List of repositories to monitor",
				Required:    true,
			},
		},
	}
}

func (g *GitHubPlugin) Initialize(settings map[string]any) error {
	token, err := getGhCliToken()
	if err != nil {
		return fmt.Errorf("failed to get gh cli token: %w", err)
	}

	repos := settings["github.repositories"].(string)
	reposStr := strings.Split(repos, ",")

	username, ok := settings["github.username"].(string)
	if !ok {
		return fmt.Errorf("username is required")
	}
	org, ok := settings["github.organization"].(string)
	if !ok {
		return fmt.Errorf("organization is required")
	}

	g.client.Init(github.GithubClientSettings{
		Username: username,
		Token:    token,
		Org:      org,
		Repos:    reposStr,
	})

	return nil
}

func (g *GitHubPlugin) Shutdown() error {
	return nil
}

func (g *GitHubPlugin) GetStandupContext(timeRange plug.TimeRange) (plug.StandupContext, error) {
	standupContext, err := contexts.Standup(g.client, timeRange)
	if err != nil {
		return plug.StandupContext{}, fmt.Errorf("failed to get standup context: %w", err)
	}

	return plug.StandupContext{
		PluginName: g.Name(),
		Content:    standupContext,
	}, nil
}

func getGhCliToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh cli error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute gh cli: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

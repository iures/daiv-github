package plugin

import (
	"fmt"
	"os/exec"
	"strings"

	"daiv-github/plugin/github"

	plug "github.com/iures/daivplug"
)

// GitHubPlugin represents the GitHub plugin for the daiv platform
type GitHubPlugin struct {
	client    *github.GitHubClient
	config    *github.GitHubConfig
	service   *github.ActivityService
	formatter github.ReportFormatter
}

// New creates a new instance of the GitHub plugin
func New() *GitHubPlugin {
	return &GitHubPlugin{}
}

// Name returns the name of the plugin
func (g *GitHubPlugin) Name() string {
	return "github"
}

// Manifest returns the plugin manifest with configuration options
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
		},
	}
}

// Initialize sets up the plugin with the provided settings
func (g *GitHubPlugin) Initialize(settings map[string]any) error {
	// Get GitHub token from the CLI
	token, err := getGhCliToken()
	if err != nil {
		return github.NewAuthenticationError("failed to get GitHub CLI token", err)
	}

  username, err := getGhUsername()
  if err != nil {
    return github.NewAuthenticationError("Failed to get GitHub CLI token", err)
  }

	// Create a map with the token
	configMap := make(map[string]any)
	for k, v := range settings {
		configMap[k] = v
	}
	configMap["github.token"] = token
  configMap["github.username"] = username

	// Create config provider and loader
	provider := github.NewMapConfigProvider(configMap)
	loader := github.NewConfigLoader(provider)

	// Load configuration
	config, err := loader.LoadConfig()
	if err != nil {
		return err
	}
	g.config = config

	// Create client
	client, err := github.NewGitHubClient(config)
	if err != nil {
		return github.NewConfigurationError("failed to create GitHub client", err)
	}
	g.client = client

	// Create service
	g.service = github.NewActivityService(client.GetClient(), g.config)

	// Create formatter based on format
	switch config.Format {
	case "json":
		g.formatter = github.NewJSONFormatter()
	default:
		g.formatter = github.NewJSONFormatter()
	}

	return nil
}

// Shutdown cleans up resources when the plugin is being shut down
func (g *GitHubPlugin) Shutdown() error {
	// No resources to clean up in this plugin
	return nil
}

func (g *GitHubPlugin) GetStandupContext(timeRange plug.TimeRange) (plug.StandupContext, error) {
	report, err := g.service.GetGithubActivityReport(timeRange)
	if err != nil {
		return plug.StandupContext{}, github.NewInternalError("failed to get github activity", err)
	}

	content, err := g.formatter.Format(report)
	if err != nil {
		return plug.StandupContext{}, github.NewInternalError("failed to format github activity", err)
	}

	return plug.StandupContext{
		PluginName: g.Name(),
		Content:    content.Content,
	}, nil
}

// getGhCliToken retrieves the GitHub token from the GitHub CLI
func getGhCliToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("GitHub CLI error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute GitHub CLI: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func getGhUsername() (string, error) {
  cmd := exec.Command("gh", "api", "user", "--jq", ".login")

  output, err := cmd.Output()
  if err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
      return "", fmt.Errorf("GitHub CLI error: %s", string(exitErr.Stderr))
    }
    return "", fmt.Errorf("failed to execute GitHub CLI: %v", err)
  }
  return strings.TrimSpace(string(output)), nil
}

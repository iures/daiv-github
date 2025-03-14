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
				Description: "List of repositories to monitor (comma-separated)",
				Required:    true,
			},
			{
				Type:        plug.ConfigTypeString,
				Key:         "github.format",
				Name:        "Report Format",
				Description: "The format for the activity report (json, markdown, or html)",
				Required:    false,
			},
			{
				Type:        plug.ConfigTypeString,
				Key:         "github.query.base_branch",
				Name:        "Base Branch",
				Description: "The base branch to filter pull requests by (default: master)",
				Required:    false,
			},
			{
				Type:        plug.ConfigTypeString,
				Key:         "github.query.include_authored",
				Name:        "Include Authored PRs",
				Description: "Whether to include authored pull requests (true/false)",
				Required:    false,
			},
			{
				Type:        plug.ConfigTypeString,
				Key:         "github.query.include_reviewed",
				Name:        "Include Reviewed PRs",
				Description: "Whether to include reviewed pull requests (true/false)",
				Required:    false,
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

	// Create a map with the token
	configMap := make(map[string]any)
	for k, v := range settings {
		configMap[k] = v
	}
	configMap["github.token"] = token

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
	g.service = github.NewActivityService(client.GetRepository(), config)

	// Create formatter based on format
	switch config.Format {
	case "json":
		g.formatter = github.NewJSONFormatter()
	case "html":
		g.formatter = github.NewHTMLFormatter()
	default:
		g.formatter = github.NewMarkdownFormatter()
	}

	return nil
}

// Shutdown cleans up resources when the plugin is being shut down
func (g *GitHubPlugin) Shutdown() error {
	// No resources to clean up in this plugin
	return nil
}

// GetContext retrieves the GitHub activity context for the given time range
func (g *GitHubPlugin) GetStandupContext(timeRange plug.TimeRange) (plug.StandupContext, error) {
	// Get activity report from service
	report, err := g.service.GetActivityReport(timeRange)
	if err != nil {
		return plug.StandupContext{}, github.NewInternalError("failed to get activity report", err)
	}

	// Format the report
	formattedContent, err := g.formatter.Format(report)
	if err != nil {
		return plug.StandupContext{}, github.NewInternalError("failed to format report", err)
	}

	return plug.StandupContext{
		PluginName: g.Name(),
		Content: formattedContent.Content,
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

package plugin

import (
	"fmt"
	"os/exec"
	"strings"

	"daiv-github/plugin/github"

	plug "github.com/iures/daivplug"
)

type GitHubPlugin struct {
	client    *github.GitHubClient
	config    *github.GitHubConfig
	service   *github.ActivityService
	formatter github.ReportFormatter
}

func New() *GitHubPlugin {
	return &GitHubPlugin{}
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

func (g *GitHubPlugin) Initialize(settings map[string]any) error {
	token, err := getGhCliToken()
	if err != nil {
		return fmt.Errorf("failed to get gh cli token: %w", err)
	}

	reposStr, ok := settings["github.repositories"].(string)
	if !ok {
		return fmt.Errorf("repositories are required")
	}
	repos := strings.Split(reposStr, ",")
	// Trim whitespace from each repository
	for i, repo := range repos {
		repos[i] = strings.TrimSpace(repo)
	}

	username, ok := settings["github.username"].(string)
	if !ok {
		return fmt.Errorf("username is required")
	}
	
	org, ok := settings["github.organization"].(string)
	if !ok {
		return fmt.Errorf("organization is required")
	}

	// Create default query options
	queryOptions := github.DefaultQueryOptions()

	// Override with user-provided options if available
	if baseBranch, ok := settings["github.query.base_branch"].(string); ok && baseBranch != "" {
		queryOptions.BaseBranch = baseBranch
	}

	if includeAuthored, ok := settings["github.query.include_authored"].(string); ok && includeAuthored != "" {
		queryOptions.IncludeAuthored = includeAuthored == "true"
	}

	if includeReviewed, ok := settings["github.query.include_reviewed"].(string); ok && includeReviewed != "" {
		queryOptions.IncludeReviewed = includeReviewed == "true"
	}

	// Create the config
	config := &github.GitHubConfig{
		Username:     username,
		Token:        token,
		Organization: org,
		Repositories: repos,
		QueryOptions: queryOptions,
	}

	// Create the client
	client, err := github.NewGitHubClient(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	g.client = client
	g.config = config
	
	// Create the service
	g.service = github.NewActivityService(client.GetRepository(), config)

	// Set the formatter based on configuration
	format, ok := settings["github.format"].(string)
	if !ok || format == "" {
		format = "markdown" // Default to markdown if not specified
	}

	switch format {
	case "json":
		g.formatter = github.NewJSONFormatter()
	case "html":
		g.formatter = github.NewHTMLFormatter()
	case "markdown":
		g.formatter = github.NewMarkdownFormatter()
	default:
		g.formatter = github.NewMarkdownFormatter()
	}

	return nil
}

func (g *GitHubPlugin) Shutdown() error {
	return nil
}

func (g *GitHubPlugin) GetStandupContext(timeRange plug.TimeRange) (plug.StandupContext, error) {
	// Get activity report from service
	report, err := g.service.GetActivityReport(timeRange)
	if err != nil {
		return plug.StandupContext{}, fmt.Errorf("failed to get activity report: %w", err)
	}
	
	// Format the report using the configured formatter
	formattedContent, err := g.formatter.Format(report)
	if err != nil {
		return plug.StandupContext{}, fmt.Errorf("failed to format activity report: %w", err)
	}

	return plug.StandupContext{
		PluginName: g.Name(),
		Content:    formattedContent.Content,
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

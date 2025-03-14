package contexts

import (
	"daiv-github/plugin/github"

	plug "github.com/iures/daivplug"
)

// Standup generates a standup context for GitHub activity
// This is kept for backward compatibility
func Standup(client *github.GithubClient, timeRange plug.TimeRange) (string, error) {
	return client.GetStandupContext(timeRange)
}

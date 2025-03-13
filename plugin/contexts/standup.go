package contexts

import (
	"daiv-github/plugin/github"

	plug "github.com/iures/daivplug"
)

func Standup(client *github.GithubClient, timeRange plug.TimeRange) (string, error) {
	return client.GetStandupContext(timeRange)
}

package github

import (
	"context"
	"time"

	externalGithub "github.com/google/go-github/v68/github"
	plug "github.com/iures/daivplug"
)

// ActivityService handles the processing of GitHub data into domain models
type ActivityService struct {
	client *externalGithub.Client
	config *GitHubConfig
}

// NewActivityService creates a new activity service
func NewActivityService(client *externalGithub.Client, config *GitHubConfig) *ActivityService {
	return &ActivityService{
		client: client,
		config: config,
	}
}

func (s *ActivityService) GetGithubActivityReport(timeRange plug.TimeRange) (*ActivityReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch all events - GitHub API doesn't support time range filtering directly
	public_only := false
	events, _, err := s.client.Activity.ListEventsPerformedByUser(ctx, s.config.Username, public_only, &externalGithub.ListOptions{PerPage: 100})
	if err != nil {
		return nil, NewAPIError("failed to list events", err)
	}

	// Filter events by time range
	filteredEvents := []*externalGithub.Event{}
	for _, event := range events {
		eventTime := event.GetCreatedAt().Time

		if timeRange.IsInRange(eventTime) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return &ActivityReport{
		TimeRange: timeRange,
		Events:    filteredEvents,
		User: User{
			Username: s.config.Username,
		},
	}, nil
}

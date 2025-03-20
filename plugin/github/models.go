package github

import (
	externalGithub "github.com/google/go-github/v68/github"
	plug "github.com/iures/daivplug"
)

// ActivityReport represents processed GitHub activity data for a specific time range
type ActivityReport struct {
	Events       []*externalGithub.Event
	TimeRange    plug.TimeRange
	User         User
}

// User represents a GitHub user
type User struct {
	Username string
	Email    string
	Name     string
}

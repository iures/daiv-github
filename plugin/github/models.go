package github

import (
	"time"

	plug "github.com/iures/daivplug"
)

// ActivityReport represents processed GitHub activity data for a specific time range
type ActivityReport struct {
	TimeRange    TimeRange
	User         User
	Repositories []Repository
}

// TimeRange represents a time period with start and end times
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// IsInRange checks if a given time is within the time range
func (tr TimeRange) IsInRange(t time.Time) bool {
	return (t.Equal(tr.Start) || t.After(tr.Start)) && 
	       (t.Equal(tr.End) || t.Before(tr.End))
}

// FromPlugTimeRange converts a plug.TimeRange to a github.TimeRange
func FromPlugTimeRange(tr plug.TimeRange) TimeRange {
	return TimeRange{
		Start: tr.Start,
		End:   tr.End,
	}
}

// User represents a GitHub user
type User struct {
	Username string
	Email    string
	Name     string
}

// Repository represents a GitHub repository with activity
type Repository struct {
	Name         string
	Organization string
	PullRequests []PullRequest
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number      int
	Title       string
	State       string
	Author      string
	URL         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Commits     []Commit
	Comments    []Comment
	Reviews     []Review
	IsReviewed  bool
	IsAuthored  bool
	BaseBranch  string
	HeadBranch  string
	Repository  string
	Organization string
}

// Commit represents a Git commit
type Commit struct {
	SHA       string
	Message   string
	Author    string
	Timestamp time.Time
}

// Review represents a review on a pull request
type Review struct {
	ID        int64
	State     string
	Body      string
	Author    string
	Timestamp time.Time
}

// Comment represents a comment on a pull request
type Comment struct {
	ID        int64
	Body      string
	Author    string
	Timestamp time.Time
	URL       string
}

// QueryOptions represents options for querying GitHub data
type QueryOptions struct {
	IncludeAuthored  bool
	IncludeReviewed  bool
	IncludeCommits   bool
	IncludeComments  bool
	BaseBranch       string
}

// DefaultQueryOptions returns the default query options
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		IncludeAuthored:  true,
		IncludeReviewed:  true,
		IncludeCommits:   true,
		IncludeComments:  true,
		BaseBranch:       "master",
	}
} 

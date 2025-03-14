package github

import "time"

// ActivityReport represents processed GitHub activity data for a specific time range
type ActivityReport struct {
	TimeRange    TimeRange
	User         User
	Repositories []Repository
}

// TimeRange represents a time period for the report
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// IsInRange checks if a given time is within the time range
func (tr TimeRange) IsInRange(t time.Time) bool {
	return (t.Equal(tr.Start) || t.After(tr.Start)) && t.Before(tr.End)
}

// User represents a GitHub user
type User struct {
	Username string
	Email    string
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
	URL         string
	State       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Author      string
	Commits     []Commit
	Reviews     []Review
	Comments    []Comment
	IsAuthored  bool
	IsReviewed  bool
}

// Commit represents a commit in a pull request
type Commit struct {
	SHA       string
	Message   string
	Author    string
	Timestamp time.Time
}

// Review represents a review on a pull request
type Review struct {
	ID        int64
	Author    string
	State     string
	Body      string
	Timestamp time.Time
}

// Comment represents a comment on a pull request
type Comment struct {
	ID        int64
	Author    string
	Body      string
	Timestamp time.Time
	Path      string
	Position  int
}

// QueryOptions represents configurable options for GitHub queries
type QueryOptions struct {
	// Base branch to filter pull requests by
	BaseBranch string
	
	// Maximum number of results to return
	MaxResults int
	
	// Whether to include authored pull requests
	IncludeAuthored bool
	
	// Whether to include reviewed pull requests
	IncludeReviewed bool
	
	// Whether to include comments
	IncludeComments bool
	
	// Whether to include commits
	IncludeCommits bool
}

// DefaultQueryOptions returns the default query options
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		BaseBranch:      "master",
		MaxResults:      100,
		IncludeAuthored: true,
		IncludeReviewed: true,
		IncludeComments: true,
		IncludeCommits:  true,
	}
} 

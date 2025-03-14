package github

import "context"

// MockGitHubRepository is a mock implementation of GitHubRepository for testing
type MockGitHubRepository struct {
	MockGetUser        func(ctx context.Context) (*User, error)
	MockGetPullRequests func(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error)
}

// GetUser implements the GitHubRepository interface
func (m *MockGitHubRepository) GetUser(ctx context.Context) (*User, error) {
	return m.MockGetUser(ctx)
}

// GetPullRequests implements the GitHubRepository interface
func (m *MockGitHubRepository) GetPullRequests(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
	return m.MockGetPullRequests(ctx, org, repo, timeRange, options)
} 

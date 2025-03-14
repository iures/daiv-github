package github

// MockGitHubRepository is a mock implementation of GitHubRepository for testing
type MockGitHubRepository struct {
	MockGetUser        func() (*User, error)
	MockGetPullRequests func(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error)
}

// GetUser implements the GitHubRepository interface
func (m *MockGitHubRepository) GetUser() (*User, error) {
	return m.MockGetUser()
}

// GetPullRequests implements the GitHubRepository interface
func (m *MockGitHubRepository) GetPullRequests(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
	return m.MockGetPullRequests(org, repo, timeRange, options)
} 

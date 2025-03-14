package github

import (
	"context"
	"errors"
	"testing"
	"time"

	plug "github.com/iures/daivplug"
)

// We're using the MockGitHubRepository from mock_repository.go

func TestNewActivityService(t *testing.T) {
	// Create a mock repository
	mockRepo := &MockGitHubRepository{
		MockGetUser: func(ctx context.Context) (*User, error) {
			return &User{Username: "testuser"}, nil
		},
		MockGetPullRequests: func(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
			return []PullRequest{}, nil
		},
	}
	
	// Create a config
	config := &GitHubConfig{
		Username:     "testuser",
		Token:        "testtoken",
		Organization: "testorg",
		Repositories: []string{"repo1", "repo2"},
		QueryOptions: DefaultQueryOptions(),
	}
	
	// Create the service
	service := NewActivityService(mockRepo, config)
	
	// Check that the service was created correctly
	if service.repository == nil {
		t.Errorf("Expected repository to be set, got nil")
	}
	
	if service.config != config {
		t.Errorf("Expected config to be %v, got %v", config, service.config)
	}
}

func TestActivityService_GetActivityReport(t *testing.T) {
	// Setup test cases
	testCases := []struct {
		name          string
		mockRepo      *MockGitHubRepository
		config        *GitHubConfig
		timeRange     plug.TimeRange
		expectError   bool
		expectedRepos int
	}{
		{
			name: "Successful report generation",
			mockRepo: &MockGitHubRepository{
				MockGetUser: func(ctx context.Context) (*User, error) {
					return &User{
						Username: "testuser",
						Email:    "test@example.com",
					}, nil
				},
				MockGetPullRequests: func(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
					return []PullRequest{
						{
							Number:    1,
							Title:     "Test PR",
							State:     "open",
							Author:    "testuser",
							URL:       "https://github.com/testorg/repo1/pull/1",
							CreatedAt: time.Now().Add(-24 * time.Hour),
							UpdatedAt: time.Now(),
						},
					}, nil
				},
			},
			config: &GitHubConfig{
				Username:     "testuser",
				Token:        "testtoken",
				Organization: "testorg",
				Repositories: []string{"repo1"},
				QueryOptions: DefaultQueryOptions(),
			},
			timeRange: plug.TimeRange{
				Start: time.Now().Add(-48 * time.Hour),
				End:   time.Now(),
			},
			expectError:   false,
			expectedRepos: 1,
		},
		{
			name: "Error getting user",
			mockRepo: &MockGitHubRepository{
				MockGetUser: func(ctx context.Context) (*User, error) {
					return nil, errors.New("failed to get user")
				},
				MockGetPullRequests: func(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
					return []PullRequest{}, nil
				},
			},
			config: &GitHubConfig{
				Username:     "testuser",
				Token:        "testtoken",
				Organization: "testorg",
				Repositories: []string{"repo1"},
				QueryOptions: DefaultQueryOptions(),
			},
			timeRange: plug.TimeRange{
				Start: time.Now().Add(-48 * time.Hour),
				End:   time.Now(),
			},
			expectError:   true,
			expectedRepos: 0,
		},
		{
			name: "Error getting pull requests",
			mockRepo: &MockGitHubRepository{
				MockGetUser: func(ctx context.Context) (*User, error) {
					return &User{
						Username: "testuser",
						Email:    "test@example.com",
					}, nil
				},
				MockGetPullRequests: func(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
					return nil, errors.New("failed to get pull requests")
				},
			},
			config: &GitHubConfig{
				Username:     "testuser",
				Token:        "testtoken",
				Organization: "testorg",
				Repositories: []string{"repo1"},
				QueryOptions: DefaultQueryOptions(),
			},
			timeRange: plug.TimeRange{
				Start: time.Now().Add(-48 * time.Hour),
				End:   time.Now(),
			},
			expectError:   false, // We don't expect an error because we continue on repository errors
			expectedRepos: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the service
			service := NewActivityService(tc.mockRepo, tc.config)
			
			// Get the activity report
			report, err := service.GetActivityReport(tc.timeRange)
			
			// Check for errors
			if tc.expectError && err == nil {
				t.Errorf("Expected an error, got nil")
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			
			// If we don't expect an error, check the report
			if !tc.expectError && err == nil {
				if report == nil {
					t.Errorf("Expected a report, got nil")
				} else {
					if len(report.Repositories) != tc.expectedRepos {
						t.Errorf("Expected %d repositories, got %d", tc.expectedRepos, len(report.Repositories))
					}
				}
			}
		})
	}
}

func TestActivityService_ProcessRepository(t *testing.T) {
	// Create a mock repository
	mockRepo := &MockGitHubRepository{
		MockGetUser: func(ctx context.Context) (*User, error) {
			return &User{
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
		MockGetPullRequests: func(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
			if repo == "error-repo" {
				return nil, errors.New("failed to get pull requests")
			}
			return []PullRequest{
				{
					Number:    1,
					Title:     "Test PR",
					State:     "open",
					Author:    "testuser",
					URL:       "https://github.com/testorg/repo1/pull/1",
					CreatedAt: time.Now().Add(-24 * time.Hour),
					UpdatedAt: time.Now(),
				},
			}, nil
		},
	}
	
	// Create a config
	config := &GitHubConfig{
		Username:     "testuser",
		Token:        "testtoken",
		Organization: "testorg",
		Repositories: []string{"repo1", "error-repo"},
		QueryOptions: DefaultQueryOptions(),
	}
	
	// Create the service
	service := NewActivityService(mockRepo, config)
	
	// Create a context
	ctx := context.Background()
	
	// Test successful repository processing
	repo, err := service.processRepository(ctx, "testorg", "repo1", TimeRange{
		Start: time.Now().Add(-48 * time.Hour),
		End:   time.Now(),
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if repo.Name != "repo1" {
		t.Errorf("Expected repository name to be 'repo1', got '%s'", repo.Name)
	}
	
	if repo.Organization != "testorg" {
		t.Errorf("Expected organization to be 'testorg', got '%s'", repo.Organization)
	}
	
	if len(repo.PullRequests) != 1 {
		t.Errorf("Expected 1 pull request, got %d", len(repo.PullRequests))
	}
	
	// Test error case
	_, err = service.processRepository(ctx, "testorg", "error-repo", TimeRange{
		Start: time.Now().Add(-48 * time.Hour),
		End:   time.Now(),
	})
	
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
} 

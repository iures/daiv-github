package github

import (
	"errors"
	"testing"
	"time"

	plug "github.com/iures/daivplug"
)

// We're using the MockGitHubRepository from repository_test.go

func TestNewActivityService(t *testing.T) {
	// Create a mock repository
	mockRepo := &MockGitHubRepository{}
	
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
	if service.repository != mockRepo {
		t.Errorf("Expected repository to be %v, got %v", mockRepo, service.repository)
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
				MockGetUser: func() (*User, error) {
					return &User{
						Username: "testuser",
						Email:    "test@example.com",
					}, nil
				},
				MockGetPullRequests: func(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
					return []PullRequest{
						{
							Number:     1,
							Title:      "Test PR",
							URL:        "https://github.com/testorg/repo1/pull/1",
							State:      "open",
							CreatedAt:  time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
							UpdatedAt:  time.Date(2023, 1, 1, 16, 0, 0, 0, time.UTC),
							Author:     "testuser",
							IsAuthored: true,
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
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expectError:   false,
			expectedRepos: 1,
		},
		{
			name: "Error getting user",
			mockRepo: &MockGitHubRepository{
				MockGetUser: func() (*User, error) {
					return nil, errors.New("failed to get user")
				},
				MockGetPullRequests: func(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
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
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expectError:   true,
			expectedRepos: 0,
		},
		{
			name: "Error getting pull requests",
			mockRepo: &MockGitHubRepository{
				MockGetUser: func() (*User, error) {
					return &User{
						Username: "testuser",
						Email:    "test@example.com",
					}, nil
				},
				MockGetPullRequests: func(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
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
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expectError:   false, // We don't expect an error because we continue with other repositories
			expectedRepos: 0,
		},
		{
			name: "Multiple repositories",
			mockRepo: &MockGitHubRepository{
				MockGetUser: func() (*User, error) {
					return &User{
						Username: "testuser",
						Email:    "test@example.com",
					}, nil
				},
				MockGetPullRequests: func(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
					return []PullRequest{
						{
							Number:     1,
							Title:      "Test PR",
							URL:        "https://github.com/testorg/" + repo + "/pull/1",
							State:      "open",
							CreatedAt:  time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
							UpdatedAt:  time.Date(2023, 1, 1, 16, 0, 0, 0, time.UTC),
							Author:     "testuser",
							IsAuthored: true,
						},
					}, nil
				},
			},
			config: &GitHubConfig{
				Username:     "testuser",
				Token:        "testtoken",
				Organization: "testorg",
				Repositories: []string{"repo1", "repo2", "repo3"},
				QueryOptions: DefaultQueryOptions(),
			},
			timeRange: plug.TimeRange{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expectError:   false,
			expectedRepos: 3,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with mock repository
			service := NewActivityService(tc.mockRepo, tc.config)

			// Call the method being tested
			report, err := service.GetActivityReport(tc.timeRange)

			// Check error
			if tc.expectError && err == nil {
				t.Errorf("Expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// If no error is expected, check the report
			if !tc.expectError && err == nil {
				// Check time range
				if !report.TimeRange.Start.Equal(tc.timeRange.Start) {
					t.Errorf("Expected start time %v, got %v", tc.timeRange.Start, report.TimeRange.Start)
				}
				if !report.TimeRange.End.Equal(tc.timeRange.End) {
					t.Errorf("Expected end time %v, got %v", tc.timeRange.End, report.TimeRange.End)
				}

				// Check repositories count
				if len(report.Repositories) != tc.expectedRepos {
					t.Errorf("Expected %d repositories, got %d", tc.expectedRepos, len(report.Repositories))
				}

				// Check user info if repositories were returned
				if tc.expectedRepos > 0 {
					expectedUser, _ := tc.mockRepo.GetUser()
					if report.User.Username != expectedUser.Username {
						t.Errorf("Expected username %s, got %s", expectedUser.Username, report.User.Username)
					}
					if report.User.Email != expectedUser.Email {
						t.Errorf("Expected email %s, got %s", expectedUser.Email, report.User.Email)
					}
				}
			}
		})
	}
}

func TestActivityService_ProcessRepository(t *testing.T) {
	// Create a mock repository
	mockRepo := &MockGitHubRepository{
		MockGetPullRequests: func(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
			return []PullRequest{
				{
					Number:     1,
					Title:      "Test PR",
					URL:        "https://github.com/testorg/repo1/pull/1",
					State:      "open",
					CreatedAt:  time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2023, 1, 1, 16, 0, 0, 0, time.UTC),
					Author:     "testuser",
					IsAuthored: true,
				},
			}, nil
		},
	}
	
	// Create a config
	config := &GitHubConfig{
		Username:     "testuser",
		Token:        "testtoken",
		Organization: "testorg",
		Repositories: []string{"repo1"},
		QueryOptions: DefaultQueryOptions(),
	}
	
	// Create the service
	service := NewActivityService(mockRepo, config)
	
	// Create a time range
	timeRange := TimeRange{
		Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	
	// Call the method being tested
	repo, err := service.processRepository("testorg", "repo1", timeRange)
	
	// Check error
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	
	// Check repository
	if repo.Name != "repo1" {
		t.Errorf("Expected repository name to be 'repo1', got '%s'", repo.Name)
	}
	
	if repo.Organization != "testorg" {
		t.Errorf("Expected repository organization to be 'testorg', got '%s'", repo.Organization)
	}
	
	if len(repo.PullRequests) != 1 {
		t.Errorf("Expected 1 pull request, got %d", len(repo.PullRequests))
	}
	
	// Test error case
	mockRepo.MockGetPullRequests = func(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
		return nil, errors.New("failed to get pull requests")
	}
	
	// Call the method being tested
	_, err = service.processRepository("testorg", "repo1", timeRange)
	
	// Check error
	if err == nil {
		t.Errorf("Expected an error but got nil")
	}
} 

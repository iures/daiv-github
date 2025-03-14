package github

import (
	"context"
	"fmt"
	"sync"
	"time"

	plug "github.com/iures/daivplug"
)

// ActivityService handles the processing of GitHub data into domain models
type ActivityService struct {
	repository GitHubRepository
	config     *GitHubConfig
}

// NewActivityService creates a new activity service
func NewActivityService(repository GitHubRepository, config *GitHubConfig) *ActivityService {
	return &ActivityService{
		repository: repository,
		config:     config,
	}
}

// GetActivityReport retrieves and processes GitHub activity data for the given time range
func (s *ActivityService) GetActivityReport(pluginTimeRange plug.TimeRange) (*ActivityReport, error) {
	// Create a context with timeout
	ctx, cancel := NewContextWithTimeout(30 * time.Second)
	defer cancel()

	// Convert plugin.TimeRange to our domain TimeRange
	timeRange := TimeRange{
		Start: pluginTimeRange.Start,
		End:   pluginTimeRange.End,
	}

	// Add context values
	ctx = WithUsername(ctx, s.config.Username)
	ctx = WithOrganization(ctx, s.config.Organization)
	ctx = WithTimeRange(ctx, timeRange)

	// Get the current user
	user, err := s.repository.GetUser(ctx)
	if err != nil {
		return nil, NewInternalError("failed to get user", err)
	}

	// Create the activity report
	report := &ActivityReport{
		TimeRange: timeRange,
		User:      *user,
		Repositories: make([]Repository, 0, len(s.config.Repositories)),
	}

	// Process repositories concurrently
	if len(s.config.Repositories) > 1 {
		report.Repositories = s.processRepositoriesConcurrently(ctx, timeRange)
	} else {
		report.Repositories = s.processRepositoriesSequentially(ctx, timeRange)
	}

	return report, nil
}

// processRepositoriesConcurrently processes repositories in parallel
func (s *ActivityService) processRepositoriesConcurrently(ctx context.Context, timeRange TimeRange) []Repository {
	var wg sync.WaitGroup
	resultChan := make(chan Repository, len(s.config.Repositories))

	for _, repoName := range s.config.Repositories {
		wg.Add(1)
		go func(repoName string) {
			defer wg.Done()
			
			// Create a repository-specific context
			repoCtx := WithRepository(ctx, repoName)
			
			repo, err := s.processRepository(repoCtx, s.config.Organization, repoName, timeRange)
			if err != nil {
				// Log error but continue with other repositories
				return
			}
			resultChan <- repo
		}(repoName)
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from the channel
	repositories := make([]Repository, 0, len(s.config.Repositories))
	for repo := range resultChan {
		repositories = append(repositories, repo)
	}

	return repositories
}

// processRepositoriesSequentially processes repositories sequentially
func (s *ActivityService) processRepositoriesSequentially(ctx context.Context, timeRange TimeRange) []Repository {
	repositories := make([]Repository, 0, len(s.config.Repositories))

	for _, repoName := range s.config.Repositories {
		// Create a repository-specific context
		repoCtx := WithRepository(ctx, repoName)
		
		repo, err := s.processRepository(repoCtx, s.config.Organization, repoName, timeRange)
		if err != nil {
			// Log error but continue with other repositories
			continue
		}
		repositories = append(repositories, repo)
	}

	return repositories
}

// processRepository processes a single repository
func (s *ActivityService) processRepository(ctx context.Context, org string, repoName string, timeRange TimeRange) (Repository, error) {
	// Get pull requests for the repository
	pullRequests, err := s.repository.GetPullRequests(ctx, org, repoName, timeRange, s.config.QueryOptions)
	if err != nil {
		return Repository{}, NewAPIError(fmt.Sprintf("failed to get pull requests for %s/%s", org, repoName), err)
	}

	// Create the repository
	repository := Repository{
		Name:         repoName,
		Organization: org,
		PullRequests: pullRequests,
	}

	return repository, nil
} 

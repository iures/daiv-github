package github

import (
	"fmt"
	"sync"

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
	// Convert plugin.TimeRange to our domain TimeRange
	timeRange := TimeRange{
		Start: pluginTimeRange.Start,
		End:   pluginTimeRange.End,
	}

	// Get the current user
	user, err := s.repository.GetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create the activity report
	report := &ActivityReport{
		TimeRange: timeRange,
		User:      *user,
		Repositories: make([]Repository, 0, len(s.config.Repositories)),
	}

	// Process repositories concurrently
	if len(s.config.Repositories) > 1 {
		report.Repositories = s.processRepositoriesConcurrently(timeRange)
	} else {
		report.Repositories = s.processRepositoriesSequentially(timeRange)
	}

	return report, nil
}

// processRepositoriesConcurrently processes repositories in parallel
func (s *ActivityService) processRepositoriesConcurrently(timeRange TimeRange) []Repository {
	var wg sync.WaitGroup
	resultChan := make(chan Repository, len(s.config.Repositories))

	for _, repoName := range s.config.Repositories {
		wg.Add(1)
		go func(repoName string) {
			defer wg.Done()
			repo, err := s.processRepository(s.config.Organization, repoName, timeRange)
			if err != nil {
				// Log error but continue with other repositories
				fmt.Printf("Error processing repository %s: %v\n", repoName, err)
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
func (s *ActivityService) processRepositoriesSequentially(timeRange TimeRange) []Repository {
	repositories := make([]Repository, 0, len(s.config.Repositories))

	for _, repoName := range s.config.Repositories {
		repo, err := s.processRepository(s.config.Organization, repoName, timeRange)
		if err != nil {
			// Log error but continue with other repositories
			fmt.Printf("Error processing repository %s: %v\n", repoName, err)
			continue
		}
		repositories = append(repositories, repo)
	}

	return repositories
}

// processRepository processes a single repository
func (s *ActivityService) processRepository(org string, repoName string, timeRange TimeRange) (Repository, error) {
	repository := Repository{
		Name:         repoName,
		Organization: org,
	}

	// Get pull requests for the repository
	pullRequests, err := s.repository.GetPullRequests(org, repoName, timeRange, s.config.QueryOptions)
	if err != nil {
		return repository, fmt.Errorf("failed to get pull requests for %s/%s: %w", org, repoName, err)
	}

	// Only include repositories with activity
	if len(pullRequests) > 0 {
		repository.PullRequests = pullRequests
	}

	return repository, nil
} 

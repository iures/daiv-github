package github

import (
	"context"
	"fmt"

	externalGithub "github.com/google/go-github/v68/github"
)

// GitHubRepository defines the interface for accessing GitHub data
type GitHubRepository interface {
	GetUser() (*User, error)
	GetPullRequests(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error)
}

// GitHubAPIRepository implements GitHubRepository using the GitHub API
type GitHubAPIRepository struct {
	client   *externalGithub.Client
	username string
}

// NewGitHubAPIRepository creates a new GitHubAPIRepository
func NewGitHubAPIRepository(client *externalGithub.Client, username string) *GitHubAPIRepository {
	return &GitHubAPIRepository{
		client:   client,
		username: username,
	}
}

// GetUser retrieves the current user from GitHub
func (r *GitHubAPIRepository) GetUser() (*User, error) {
	ctx := context.Background()
	
	user, _, err := r.client.Users.Get(ctx, r.username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from GitHub: %w", err)
	}
	
	return &User{
		Username: user.GetLogin(),
		Email:    user.GetEmail(),
	}, nil
}

// GetPullRequests retrieves pull requests from GitHub based on the given parameters
func (r *GitHubAPIRepository) GetPullRequests(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
	var allPRs []PullRequest

	// Get authored PRs if enabled
	if options.IncludeAuthored {
		authoredPRs, err := r.searchAuthoredPullRequests(org, repo, timeRange, options)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, authoredPRs...)
	}
	
	// Get reviewed PRs if enabled
	if options.IncludeReviewed {
		reviewedPRs, err := r.searchReviewedPullRequests(org, repo, timeRange, options)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, reviewedPRs...)
	}
	
	// Enrich pull requests with commits, reviews, and comments
	for i := range allPRs {
		if options.IncludeCommits {
			commits, err := r.getCommits(org, repo, allPRs[i].Number, timeRange)
			if err != nil {
				return nil, err
			}
			allPRs[i].Commits = commits
		}
		
		if options.IncludeComments {
			comments, err := r.getComments(org, repo, allPRs[i].Number, timeRange)
			if err != nil {
				return nil, err
			}
			allPRs[i].Comments = comments
		}
		
		if allPRs[i].IsReviewed {
			reviews, err := r.getReviews(org, repo, allPRs[i].Number, timeRange)
			if err != nil {
				return nil, err
			}
			allPRs[i].Reviews = reviews
		}
	}
	
	return allPRs, nil
}

// searchAuthoredPullRequests searches for pull requests authored by the user
func (r *GitHubAPIRepository) searchAuthoredPullRequests(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
	ctx := context.Background()
	
	query := fmt.Sprintf(
		"is:pr author:%s repo:%s/%s base:%s updated:%s..%s",
		r.username,
		org,
		repo,
		options.BaseBranch,
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
	)
	
	searchOptions := &externalGithub.SearchOptions{
		ListOptions: externalGithub.ListOptions{PerPage: options.MaxResults},
	}
	
	result, _, err := r.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to search authored pull requests: %w", err)
	}
	
	prs := make([]PullRequest, 0, len(result.Issues))
	for _, issue := range result.Issues {
		prs = append(prs, PullRequest{
			Number:     issue.GetNumber(),
			Title:      issue.GetTitle(),
			URL:        issue.GetHTMLURL(),
			State:      issue.GetState(),
			CreatedAt:  issue.GetCreatedAt().Time,
			UpdatedAt:  issue.GetUpdatedAt().Time,
			Author:     issue.GetUser().GetLogin(),
			IsAuthored: true,
		})
	}
	
	return prs, nil
}

// searchReviewedPullRequests searches for pull requests reviewed by the user
func (r *GitHubAPIRepository) searchReviewedPullRequests(org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
	ctx := context.Background()
	
	query := fmt.Sprintf(
		"is:pr -author:%s reviewed-by:%s repo:%s/%s base:%s updated:%s..%s",
		r.username,
		r.username,
		org,
		repo,
		options.BaseBranch,
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
	)
	
	searchOptions := &externalGithub.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: externalGithub.ListOptions{PerPage: options.MaxResults},
	}
	
	result, _, err := r.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to search reviewed pull requests: %w", err)
	}
	
	prs := make([]PullRequest, 0, len(result.Issues))
	for _, issue := range result.Issues {
		prs = append(prs, PullRequest{
			Number:     issue.GetNumber(),
			Title:      issue.GetTitle(),
			URL:        issue.GetHTMLURL(),
			State:      issue.GetState(),
			CreatedAt:  issue.GetCreatedAt().Time,
			UpdatedAt:  issue.GetUpdatedAt().Time,
			Author:     issue.GetUser().GetLogin(),
			IsReviewed: true,
		})
	}
	
	return prs, nil
}

// getCommits retrieves commits for a pull request
func (r *GitHubAPIRepository) getCommits(org string, repo string, prNumber int, timeRange TimeRange) ([]Commit, error) {
	ctx := context.Background()
	
	prCommits, _, err := r.client.PullRequests.ListCommits(ctx, org, repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for PR #%d: %w", prNumber, err)
	}
	
	commits := make([]Commit, 0)
	for _, prCommit := range prCommits {
		commitTime := prCommit.GetCommit().GetCommitter().GetDate().Time
		
		// Only include commits within the time range
		if timeRange.IsInRange(commitTime) {
			commits = append(commits, Commit{
				SHA:       prCommit.GetSHA(),
				Message:   prCommit.GetCommit().GetMessage(),
				Author:    prCommit.GetCommit().GetAuthor().GetName(),
				Timestamp: commitTime,
			})
		}
	}
	
	return commits, nil
}

// getComments retrieves comments for a pull request
func (r *GitHubAPIRepository) getComments(org string, repo string, prNumber int, timeRange TimeRange) ([]Comment, error) {
	ctx := context.Background()
	
	prComments, _, err := r.client.PullRequests.ListComments(ctx, org, repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments for PR #%d: %w", prNumber, err)
	}
	
	comments := make([]Comment, 0)
	for _, prComment := range prComments {
		commentTime := prComment.GetCreatedAt().Time
		
		// Only include comments within the time range and by the current user
		if timeRange.IsInRange(commentTime) && prComment.GetUser().GetLogin() == r.username {
			comments = append(comments, Comment{
				ID:        prComment.GetID(),
				Author:    prComment.GetUser().GetLogin(),
				Body:      prComment.GetBody(),
				Timestamp: commentTime,
				Path:      prComment.GetPath(),
				Position:  prComment.GetPosition(),
			})
		}
	}
	
	return comments, nil
}

// getReviews retrieves reviews for a pull request
func (r *GitHubAPIRepository) getReviews(org string, repo string, prNumber int, timeRange TimeRange) ([]Review, error) {
	ctx := context.Background()
	
	prReviews, _, err := r.client.PullRequests.ListReviews(ctx, org, repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews for PR #%d: %w", prNumber, err)
	}
	
	reviews := make([]Review, 0)
	for _, prReview := range prReviews {
		reviewTime := prReview.GetSubmittedAt().Time
		
		// Only include reviews within the time range and by the current user
		if timeRange.IsInRange(reviewTime) && prReview.GetUser().GetLogin() == r.username {
			reviews = append(reviews, Review{
				ID:        prReview.GetID(),
				Author:    prReview.GetUser().GetLogin(),
				State:     prReview.GetState(),
				Body:      prReview.GetBody(),
				Timestamp: reviewTime,
			})
		}
	}
	
	return reviews, nil
} 

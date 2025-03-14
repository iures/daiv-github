package github

import (
	"context"
	"fmt"

	externalGithub "github.com/google/go-github/v68/github"
)

// GitHubRepository defines the interface for accessing GitHub data
type GitHubRepository interface {
	GetUser(ctx context.Context) (*User, error)
	GetPullRequests(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error)
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
func (r *GitHubAPIRepository) GetUser(ctx context.Context) (*User, error) {
	user, _, err := r.client.Users.Get(ctx, r.username)
	if err != nil {
		return nil, NewAPIError("failed to get user from GitHub", err)
	}
	
	return &User{
		Username: user.GetLogin(),
		Email:    user.GetEmail(),
		Name:     user.GetName(),
	}, nil
}

// GetPullRequests retrieves pull requests from GitHub based on the given parameters
func (r *GitHubAPIRepository) GetPullRequests(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
	var allPRs []PullRequest

	// Get authored PRs if enabled
	if options.IncludeAuthored {
		authoredPRs, err := r.searchAuthoredPullRequests(ctx, org, repo, timeRange, options)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, authoredPRs...)
	}
	
	// Get reviewed PRs if enabled
	if options.IncludeReviewed {
		reviewedPRs, err := r.searchReviewedPullRequests(ctx, org, repo, timeRange, options)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, reviewedPRs...)
	}
	
	// Enrich pull requests with commits, reviews, and comments
	for i := range allPRs {
		if options.IncludeCommits {
			commits, err := r.getCommits(ctx, org, repo, allPRs[i].Number, timeRange)
			if err != nil {
				return nil, err
			}
			allPRs[i].Commits = commits
		}
		
		if options.IncludeComments {
			comments, err := r.getComments(ctx, org, repo, allPRs[i].Number, timeRange)
			if err != nil {
				return nil, err
			}
			allPRs[i].Comments = comments
		}
		
		if allPRs[i].IsReviewed {
			reviews, err := r.getReviews(ctx, org, repo, allPRs[i].Number, timeRange)
			if err != nil {
				return nil, err
			}
			allPRs[i].Reviews = reviews
		}
	}
	
	return allPRs, nil
}

// searchAuthoredPullRequests searches for pull requests authored by the user
func (r *GitHubAPIRepository) searchAuthoredPullRequests(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
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
		ListOptions: externalGithub.ListOptions{PerPage: 100},
	}
	
	result, _, err := r.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, NewAPIError("failed to search for authored pull requests", err)
	}
	
	pullRequests := make([]PullRequest, 0, len(result.Issues))
	for _, issue := range result.Issues {
		pr := PullRequest{
			Number:       issue.GetNumber(),
			Title:        issue.GetTitle(),
			State:        issue.GetState(),
			Author:       issue.GetUser().GetLogin(),
			URL:          issue.GetHTMLURL(),
			CreatedAt:    issue.GetCreatedAt().Time,
			UpdatedAt:    issue.GetUpdatedAt().Time,
			IsAuthored:   true,
			BaseBranch:   options.BaseBranch,
			Repository:   repo,
			Organization: org,
		}
		pullRequests = append(pullRequests, pr)
	}
	
	return pullRequests, nil
}

// searchReviewedPullRequests searches for pull requests reviewed by the user
func (r *GitHubAPIRepository) searchReviewedPullRequests(ctx context.Context, org string, repo string, timeRange TimeRange, options QueryOptions) ([]PullRequest, error) {
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
		ListOptions: externalGithub.ListOptions{PerPage: 100},
	}
	
	result, _, err := r.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, NewAPIError("failed to search for reviewed pull requests", err)
	}
	
	pullRequests := make([]PullRequest, 0, len(result.Issues))
	for _, issue := range result.Issues {
		pr := PullRequest{
			Number:       issue.GetNumber(),
			Title:        issue.GetTitle(),
			State:        issue.GetState(),
			Author:       issue.GetUser().GetLogin(),
			URL:          issue.GetHTMLURL(),
			CreatedAt:    issue.GetCreatedAt().Time,
			UpdatedAt:    issue.GetUpdatedAt().Time,
			IsReviewed:   true,
			BaseBranch:   options.BaseBranch,
			Repository:   repo,
			Organization: org,
		}
		pullRequests = append(pullRequests, pr)
	}
	
	return pullRequests, nil
}

// getCommits retrieves commits for a pull request
func (r *GitHubAPIRepository) getCommits(ctx context.Context, org string, repo string, prNumber int, timeRange TimeRange) ([]Commit, error) {
	prCommits, _, err := r.client.PullRequests.ListCommits(ctx, org, repo, prNumber, nil)
	if err != nil {
		return nil, NewAPIError("failed to list commits for pull request", err)
	}
	
	commits := make([]Commit, 0, len(prCommits))
	for _, commit := range prCommits {
		if commit.GetCommit().GetCommitter() == nil {
			continue
		}
		
		commitTime := commit.GetCommit().GetCommitter().GetDate().Time
		if !timeRange.IsInRange(commitTime) {
			continue
		}
		
		authorLogin := ""
		if commit.GetAuthor() != nil {
			authorLogin = commit.GetAuthor().GetLogin()
		}
		
		commits = append(commits, Commit{
			SHA:       commit.GetSHA(),
			Message:   commit.GetCommit().GetMessage(),
			Author:    authorLogin,
			Timestamp: commitTime,
		})
	}
	
	return commits, nil
}

// getComments retrieves comments for a pull request
func (r *GitHubAPIRepository) getComments(ctx context.Context, org string, repo string, prNumber int, timeRange TimeRange) ([]Comment, error) {
	prComments, _, err := r.client.PullRequests.ListComments(ctx, org, repo, prNumber, nil)
	if err != nil {
		return nil, NewAPIError("failed to list comments for pull request", err)
	}
	
	comments := make([]Comment, 0, len(prComments))
	for _, comment := range prComments {
		if comment.GetUser() == nil {
			continue
		}
		
		commentTime := comment.GetCreatedAt().Time
		if !timeRange.IsInRange(commentTime) {
			continue
		}
		
		comments = append(comments, Comment{
			ID:        comment.GetID(),
			Body:      comment.GetBody(),
			Author:    comment.GetUser().GetLogin(),
			Timestamp: commentTime,
			URL:       comment.GetHTMLURL(),
		})
	}
	
	return comments, nil
}

// getReviews retrieves reviews for a pull request
func (r *GitHubAPIRepository) getReviews(ctx context.Context, org string, repo string, prNumber int, timeRange TimeRange) ([]Review, error) {
	prReviews, _, err := r.client.PullRequests.ListReviews(ctx, org, repo, prNumber, nil)
	if err != nil {
		return nil, NewAPIError("failed to list reviews for pull request", err)
	}
	
	reviews := make([]Review, 0, len(prReviews))
	for _, review := range prReviews {
		if review.GetUser() == nil {
			continue
		}
		
		reviewTime := review.GetSubmittedAt().Time
		if !timeRange.IsInRange(reviewTime) {
			continue
		}
		
		reviews = append(reviews, Review{
			ID:        review.GetID(),
			State:     review.GetState(),
			Body:      review.GetBody(),
			Author:    review.GetUser().GetLogin(),
			Timestamp: reviewTime,
		})
	}
	
	return reviews, nil
} 

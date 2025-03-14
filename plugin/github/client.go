package github

import (
	"context"
	"fmt"
	"slices"
	"strings"

	externalGithub "github.com/google/go-github/v68/github"
	plug "github.com/iures/daivplug"
)

// GitHubConfig represents the configuration for the GitHub client
type GitHubConfig struct {
	Username     string
	Token        string
	Organization string
	Repositories []string
	QueryOptions QueryOptions
}

// GitHubClient provides a client for interacting with GitHub
type GitHubClient struct {
	client     *externalGithub.Client
	config     *GitHubConfig
	repository GitHubRepository
}

// NewGitHubClient creates a new GitHubClient
func NewGitHubClient(config *GitHubConfig) (*GitHubClient, error) {
	authToken := externalGithub.BasicAuthTransport{
		Username: config.Username,
		Password: config.Token,
	}
	
	client := externalGithub.NewClient(authToken.Client())
	
	githubClient := &GitHubClient{
		client: client,
		config: config,
	}
	
	// Create the repository
	repository := NewGitHubAPIRepository(client, config.Username)
	githubClient.repository = repository
	
	return githubClient, nil
}

// GetRepository returns the GitHub repository
func (g *GitHubClient) GetRepository() GitHubRepository {
	return g.repository
}

type GithubClientSettings struct {
	Username string
	Token string
	Org string
	Repos []string
}

type GithubClient struct {
	Client *externalGithub.Client
	Settings GithubClientSettings
}

// NewGithubClient creates a new GithubClient instance
func NewGithubClient() *GithubClient {
	return &GithubClient{}
}

func (gc *GithubClient) Init(settings GithubClientSettings) {
	authToken := externalGithub.BasicAuthTransport{
		Username: settings.Username,
		Password: settings.Token,
	}

	gc.Client = externalGithub.NewClient(authToken.Client())
	gc.Settings = settings
}

func (gc *GithubClient) GetStandupContext(timeRange plug.TimeRange) (string, error) {
	var report strings.Builder

	for _, repo := range gc.Settings.Repos {
		repoHasContent := false
		repoSection := &strings.Builder{}
		fmt.Fprintf(repoSection, "\n# Repository: %s\n", repo)

		authoredPRs, err := gc.renderAuthoredPullRequestCommits(repo, timeRange)
		if err != nil {
			return "", fmt.Errorf("error rendering authored pull request commits for %s/%s: %v", gc.Settings.Org, repo, err)
		}
		if authoredPRs != "" {
			repoHasContent = true
			repoSection.WriteString(authoredPRs)
		}

		issuesReviewed, err := gc.searchReviewedPullRequests(repo, timeRange)
		if err != nil {
			return "", fmt.Errorf("error searching reviewed PRs for %s/%s: %v", gc.Settings.Org, repo, err)
		}

		if len(issuesReviewed) > 0 {
			repoHasContent = true
			repoSection.WriteString("\n## Reviewed Pull Requests\n")
			
			var hasReviewsInPeriod bool
			for _, issue := range issuesReviewed {
				reviewReport, err := gc.renderReviews(repo, issue, timeRange)
				if err != nil {
					return "", fmt.Errorf("error fetching reviews for PR #%d in %s/%s: %v", issue.GetNumber(), gc.Settings.Org, repo, err)
				}
				if reviewReport != "" {
					hasReviewsInPeriod = true
					fmt.Fprintln(repoSection, formatPullRequestFromIssue(issue))
					repoSection.WriteString(reviewReport)

					reviewCommentReport, err := gc.renderPrComments(repo, issue.GetNumber(), timeRange)
					if err != nil {
						return "", fmt.Errorf("error fetching comments for PR #%d in %s/%s: %v", issue.GetNumber(), gc.Settings.Org, repo, err)
					}
					repoSection.WriteString(reviewCommentReport)
				}
			}

			if !hasReviewsInPeriod {
				repoSection.WriteString("No reviews found in the specified time period.\n")
			}
		}

		if repoHasContent {
			report.WriteString(repoSection.String())
		}
	}

	if report.Len() == 0 {
		report.WriteString("\nNo GitHub activity found in the specified time period.\n")
	}

	return report.String(), nil
}

func (gc *GithubClient) renderAuthoredPullRequestCommits(repo string, timeRange plug.TimeRange) (string, error) {
	issues, err := gc.searchPullRequests(repo, timeRange)
	if err != nil {
		return "", err
	}

	var report strings.Builder

	if len(issues) > 0 {
		report.WriteString("\n## Authored Pull Requests\n")
		for _, issue := range issues {
			report.WriteString(formatPullRequestFromIssue(issue))

			commitsReport, err := gc.renderCommits(repo, issue.GetNumber(), timeRange)
			if err != nil {
				return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", issue.GetNumber(), gc.Settings.Org, repo, err)
			}
			report.WriteString(commitsReport)
		}
	}

	return report.String(), nil
}

func (gc *GithubClient) renderReviewedPullRequestCommits(repo string, timeRange plug.TimeRange) (string, error) {
	issues, err := gc.searchPullRequests(repo, timeRange)
	if err != nil {
		return "", err
	}

	var report strings.Builder

	for _, issue := range issues {
		report.WriteString(formatPullRequestFromIssue(issue))

		commitsReport, err := gc.renderCommits(repo, issue.GetNumber(), timeRange)
		if err != nil {
			return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", issue.GetNumber(), gc.Settings.Org, repo, err)
		}
		report.WriteString(commitsReport)
	}

	return report.String(), nil
}

func (gc *GithubClient) searchPullRequests(repo string, timeRange plug.TimeRange) ([]*externalGithub.Issue, error) {
	ctx := context.Background()

	query := fmt.Sprintf(
		"is:pr author:%s repo:%s/%s base:%s updated:%s..%s",
		gc.Settings.Username,
		gc.Settings.Org,
		repo,
		"master",
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
	)

	searchOptions := &externalGithub.SearchOptions{
		ListOptions: externalGithub.ListOptions{PerPage: 100},
	}
	result, _, err := gc.Client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, err
	}
	return result.Issues, nil
}

func (gc *GithubClient) searchReviewedPullRequests(repo string, timeRange plug.TimeRange) ([]*externalGithub.Issue, error) {
	ctx := context.Background()

	query := fmt.Sprintf(
		"is:pr -author:%s reviewed-by:%s repo:%s/%s base:%s updated:%s..%s",
		gc.Settings.Username,
		gc.Settings.Username,
		gc.Settings.Org,
		repo,
		"master",
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
	)

	searchOptions := &externalGithub.SearchOptions{
		Sort: "updated",
		Order: "desc",
		ListOptions: externalGithub.ListOptions{PerPage: 100},
	}

	result, _, err := gc.Client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, err
	}

	return result.Issues, nil
}


func (gc *GithubClient) renderCommits(repo string, prNumber int, timeRange plug.TimeRange) (string, error) {
	ctx := context.Background()

	prCommits, _, err := gc.Client.PullRequests.ListCommits(ctx, gc.Settings.Org, repo, prNumber, nil)
	if err != nil {
		return "", err
	}

	slices.SortFunc(prCommits, func(a, b *externalGithub.RepositoryCommit) int {
		return a.GetCommit().GetCommitter().GetDate().Compare(b.GetCommit().GetCommitter().GetDate().Time)
	})

	var commitReport strings.Builder
	relevantCommits := filterRelevantCommits(prCommits, gc.Settings.Username, timeRange)
	if len(relevantCommits) > 0 {
		commitReport.WriteString("#### Commits:\n")
		for _, commit := range relevantCommits {
			commitReport.WriteString(formatCommit(commit))
		}
	}

	return commitReport.String(), nil
}

func (gc *GithubClient) renderPrComments(repo string, prNumber int, timeRange plug.TimeRange) (string, error) {
	ctx := context.Background()

	comments, _, err := gc.Client.PullRequests.ListComments(ctx, gc.Settings.Org, repo, prNumber, nil)
	if err != nil {
		return "", err
	}

	var commentReport strings.Builder
	relevantComments := filterRelevantPRComments(comments, gc.Settings.Username, timeRange)
	if len(relevantComments) > 0 {
		commentReport.WriteString("### Comments:\n")
		for _, comment := range relevantComments {
			commentReport.WriteString(formatComment(comment))
		}
	}

	return commentReport.String(), nil
}

func filterRelevantPRComments(comments []*externalGithub.PullRequestComment, username string, timeRange plug.TimeRange) []*externalGithub.PullRequestComment {
	var relevant []*externalGithub.PullRequestComment
	for _, comment := range comments {
		if comment.User != nil && comment.User.GetLogin() == username &&
			timeRange.IsInRange(comment.GetCreatedAt().Time) {
			relevant = append(relevant, comment)
		}
	}
	return relevant
}

func filterRelevantCommits(commits []*externalGithub.RepositoryCommit, username string, timeRange plug.TimeRange) []*externalGithub.RepositoryCommit {
	var relevant []*externalGithub.RepositoryCommit
	for _, commit := range commits {
		if commit.Author != nil && commit.Author.GetLogin() == username &&
			timeRange.IsInRange(commit.GetCommit().GetCommitter().GetDate().Time) {
			relevant = append(relevant, commit)
		}
	}
	return relevant
}

func formatPullRequestFromIssue(issue *externalGithub.Issue) string {
	return fmt.Sprintf( "### PR (%s) #%d: %s\n\n", 
		strings.ToUpper(issue.GetState()),
		issue.GetNumber(), 
		issue.GetTitle(),
	)
}

func formatCommit(commit *externalGithub.RepositoryCommit) string {
	return fmt.Sprintf(
		"##### %s\n\n",
		commit.GetCommit().GetMessage(),
	)
}

func formatComment(comment *externalGithub.PullRequestComment) string {
	return fmt.Sprintf(
		"**%s** - @%s:\n```\n%s\n```\n\n",
		comment.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		comment.User.GetLogin(),
		*comment.Body,
	)
}

func (gc *GithubClient) renderReviews(repo string, issue *externalGithub.Issue, timeRange plug.TimeRange) (string, error) {
	ctx := context.Background()

	reviews, _, err := gc.Client.PullRequests.ListReviews(ctx, gc.Settings.Org, repo, issue.GetNumber(), nil)
	if err != nil {
		return "", err
	}

	var reviewReport strings.Builder
	var relevantReviews []*externalGithub.PullRequestReview

	// First collect all relevant reviews
	for _, review := range reviews {
		if review.User != nil && review.User.GetLogin() == gc.Settings.Username  {
			if review.GetSubmittedAt().IsZero() || !timeRange.IsInRange(review.GetSubmittedAt().Time) {
				continue
			}
			relevantReviews = append(relevantReviews, review)
		}
	}

	if len(relevantReviews) > 0 {
		slices.SortFunc(relevantReviews, func(a, b *externalGithub.PullRequestReview) int {
			return a.GetSubmittedAt().Compare(b.GetSubmittedAt().Time)
		})

		for _, review := range relevantReviews {
			reviewReport.WriteString(formatPullRequestReview(review))
		}
	}

	if reviewReport.Len() > 0 {
		return reviewReport.String(), nil
	}
	return "", nil
}

func formatPullRequestReview(review *externalGithub.PullRequestReview) string {
	report := fmt.Sprintf("**Review %s** - %s\n",
		strings.ToUpper(review.GetState()),
		review.GetSubmittedAt().Format("2006-01-02 15:04:05"))

	if body := review.GetBody(); body != "" {
		report += fmt.Sprintf("```\n%s\n```\n\n", body)
	}
	return report
}

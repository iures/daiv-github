package github

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormattedContent represents formatted content with its content type
type FormattedContent struct {
	ContentType string // MIME type of the content
	Content     string // The formatted content
}

// ReportFormatter is an interface for formatting activity reports
type ReportFormatter interface {
	Format(report *ActivityReport) (*FormattedContent, error)
	Name() string // Returns the name of the formatter
}

// JSONFormatter formats activity reports as JSON
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Name returns the name of the formatter
func (f *JSONFormatter) Name() string {
	return "json"
}

// Format formats an activity report as JSON
func (f *JSONFormatter) Format(report *ActivityReport) (*FormattedContent, error) {
	if len(report.Repositories) == 0 || allRepositoriesEmpty(report.Repositories) {
		return &FormattedContent{
			ContentType: "application/json",
			Content:     "{}",
		}, nil
	}

	// Marshal to JSON with proper indentation
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return &FormattedContent{
		ContentType: "application/json",
		Content:     string(output),
	}, nil
}

// MarkdownFormatter formats activity reports as Markdown
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new Markdown formatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// Name returns the name of the formatter
func (f *MarkdownFormatter) Name() string {
	return "markdown"
}

// Format formats an activity report as Markdown
func (f *MarkdownFormatter) Format(report *ActivityReport) (*FormattedContent, error) {
	if len(report.Repositories) == 0 || allRepositoriesEmpty(report.Repositories) {
		return &FormattedContent{
			ContentType: "text/markdown",
			Content:     "No GitHub activity found for the specified time range.",
		}, nil
	}

	var sb strings.Builder

	// Add report header
	sb.WriteString(fmt.Sprintf("# GitHub Activity Report\n\n"))
	sb.WriteString(fmt.Sprintf("**Time Range:** %s to %s\n\n", 
		report.TimeRange.Start.Format("2006-01-02"),
		report.TimeRange.End.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**User:** %s\n\n", report.User.Username))
	
	// Process each repository
	for _, repo := range report.Repositories {
		if len(repo.PullRequests) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## Repository: %s/%s\n\n", repo.Organization, repo.Name))
		
		// Group PRs by authored/reviewed
		var authoredPRs, reviewedPRs []PullRequest
		for _, pr := range repo.PullRequests {
			if pr.IsAuthored {
				authoredPRs = append(authoredPRs, pr)
			}
			if pr.IsReviewed {
				reviewedPRs = append(reviewedPRs, pr)
			}
		}
		
		// Add authored PRs section
		if len(authoredPRs) > 0 {
			sb.WriteString("### Authored Pull Requests\n\n")
			for _, pr := range authoredPRs {
				sb.WriteString(fmt.Sprintf("#### [#%d] %s (%s)\n\n", 
					pr.Number, pr.Title, pr.State))
				sb.WriteString(fmt.Sprintf("URL: %s\n\n", pr.URL))
				
				// Add commits
				if len(pr.Commits) > 0 {
					sb.WriteString("**Commits:**\n\n")
					for _, commit := range pr.Commits {
						sb.WriteString(fmt.Sprintf("- %s: %s\n", 
							commit.Timestamp.Format("2006-01-02 15:04"),
							commit.Message))
					}
					sb.WriteString("\n")
				}
				
				// Add comments
				if len(pr.Comments) > 0 {
					sb.WriteString("**Comments:**\n\n")
					for _, comment := range pr.Comments {
						sb.WriteString(fmt.Sprintf("- %s: %s\n", 
							comment.Timestamp.Format("2006-01-02 15:04"),
							comment.Body))
					}
					sb.WriteString("\n")
				}
				
				sb.WriteString("---\n\n")
			}
		}
		
		// Add reviewed PRs section
		if len(reviewedPRs) > 0 {
			sb.WriteString("### Reviewed Pull Requests\n\n")
			for _, pr := range reviewedPRs {
				sb.WriteString(fmt.Sprintf("#### [#%d] %s (%s)\n\n", 
					pr.Number, pr.Title, pr.State))
				sb.WriteString(fmt.Sprintf("URL: %s\n\n", pr.URL))
				
				// Add reviews
				if len(pr.Reviews) > 0 {
					sb.WriteString("**Reviews:**\n\n")
					for _, review := range pr.Reviews {
						sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", 
							review.Timestamp.Format("2006-01-02 15:04"),
							review.State,
							review.Body))
					}
					sb.WriteString("\n")
				}
				
				// Add comments
				if len(pr.Comments) > 0 {
					sb.WriteString("**Comments:**\n\n")
					for _, comment := range pr.Comments {
						sb.WriteString(fmt.Sprintf("- %s: %s\n", 
							comment.Timestamp.Format("2006-01-02 15:04"),
							comment.Body))
					}
					sb.WriteString("\n")
				}
				
				sb.WriteString("---\n\n")
			}
		}
	}

	return &FormattedContent{
		ContentType: "text/markdown",
		Content:     sb.String(),
	}, nil
}

// HTMLFormatter formats activity reports as HTML
type HTMLFormatter struct{}

// NewHTMLFormatter creates a new HTML formatter
func NewHTMLFormatter() *HTMLFormatter {
	return &HTMLFormatter{}
}

// Name returns the name of the formatter
func (f *HTMLFormatter) Name() string {
	return "html"
}

// Format formats an activity report as HTML
func (f *HTMLFormatter) Format(report *ActivityReport) (*FormattedContent, error) {
	if len(report.Repositories) == 0 || allRepositoriesEmpty(report.Repositories) {
		return &FormattedContent{
			ContentType: "text/html",
			Content:     "<html><body><h1>GitHub Activity Report</h1><p>No activity found for the specified time range.</p></body></html>",
		}, nil
	}

	var sb strings.Builder

	// Start HTML document
	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>GitHub Activity Report</title>\n")
	sb.WriteString("<style>\n")
	sb.WriteString("body { font-family: Arial, sans-serif; margin: 20px; }\n")
	sb.WriteString("h1 { color: #24292e; }\n") // GitHub dark
	sb.WriteString("h2 { color: #24292e; border-bottom: 1px solid #e1e4e8; padding-bottom: 8px; }\n")
	sb.WriteString("h3 { margin-top: 20px; color: #0366d6; }\n") // GitHub blue
	sb.WriteString(".pr { background-color: #f6f8fa; border-radius: 3px; padding: 15px; margin-bottom: 15px; }\n")
	sb.WriteString(".pr-title { font-size: 16px; margin-bottom: 10px; }\n")
	sb.WriteString(".pr-number { color: #0366d6; font-weight: bold; }\n")
	sb.WriteString(".pr-state-open { color: #28a745; }\n") // GitHub green
	sb.WriteString(".pr-state-closed { color: #d73a49; }\n") // GitHub red
	sb.WriteString(".pr-state-merged { color: #6f42c1; }\n") // GitHub purple
	sb.WriteString(".metadata { color: #586069; font-size: 14px; margin-bottom: 15px; }\n")
	sb.WriteString(".commits, .reviews, .comments { margin-top: 10px; }\n")
	sb.WriteString(".commit, .review, .comment { background-color: white; border: 1px solid #e1e4e8; padding: 10px; margin-bottom: 8px; }\n")
	sb.WriteString(".timestamp { color: #586069; font-size: 12px; }\n")
	sb.WriteString("</style>\n")
	sb.WriteString("</head>\n<body>\n")

	// Add report header
	sb.WriteString("<h1>GitHub Activity Report</h1>\n")
	sb.WriteString("<div class=\"metadata\">\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Time Range:</strong> %s to %s</p>\n", 
		report.TimeRange.Start.Format("2006-01-02"),
		report.TimeRange.End.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("<p><strong>User:</strong> %s</p>\n", report.User.Username))
	sb.WriteString("</div>\n")
	
	// Process each repository
	for _, repo := range report.Repositories {
		if len(repo.PullRequests) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("<h2>Repository: %s/%s</h2>\n", repo.Organization, repo.Name))
		
		// Group PRs by authored/reviewed
		var authoredPRs, reviewedPRs []PullRequest
		for _, pr := range repo.PullRequests {
			if pr.IsAuthored {
				authoredPRs = append(authoredPRs, pr)
			}
			if pr.IsReviewed {
				reviewedPRs = append(reviewedPRs, pr)
			}
		}
		
		// Add authored PRs section
		if len(authoredPRs) > 0 {
			sb.WriteString("<h3>Authored Pull Requests</h3>\n")
			for _, pr := range authoredPRs {
				sb.WriteString("<div class=\"pr\">\n")
				
				// Add PR state class
				stateClass := "pr-state-open"
				if pr.State == "closed" {
					stateClass = "pr-state-closed"
				} else if pr.State == "merged" {
					stateClass = "pr-state-merged"
				}
				
				sb.WriteString(fmt.Sprintf("<h4><span class=\"pr-number\">#%d</span> <span class=\"pr-title\">%s</span> <span class=\"%s\">(%s)</span></h4>\n", 
					pr.Number, pr.Title, stateClass, pr.State))
				sb.WriteString(fmt.Sprintf("<p><a href=\"%s\">%s</a></p>\n", pr.URL, pr.URL))
				
				// Add commits
				if len(pr.Commits) > 0 {
					sb.WriteString("<div class=\"commits\">\n")
					sb.WriteString("<h5>Commits</h5>\n")
					for _, commit := range pr.Commits {
						sb.WriteString("<div class=\"commit\">\n")
						sb.WriteString(fmt.Sprintf("<p>%s</p>\n", commit.Message))
						sb.WriteString(fmt.Sprintf("<p class=\"timestamp\">%s</p>\n", 
							commit.Timestamp.Format("2006-01-02 15:04:05")))
						sb.WriteString("</div>\n")
					}
					sb.WriteString("</div>\n")
				}
				
				// Add comments
				if len(pr.Comments) > 0 {
					sb.WriteString("<div class=\"comments\">\n")
					sb.WriteString("<h5>Comments</h5>\n")
					for _, comment := range pr.Comments {
						sb.WriteString("<div class=\"comment\">\n")
						sb.WriteString(fmt.Sprintf("<p>%s</p>\n", comment.Body))
						sb.WriteString(fmt.Sprintf("<p class=\"timestamp\">%s</p>\n", 
							comment.Timestamp.Format("2006-01-02 15:04:05")))
						sb.WriteString("</div>\n")
					}
					sb.WriteString("</div>\n")
				}
				
				sb.WriteString("</div>\n")
			}
		}
		
		// Add reviewed PRs section
		if len(reviewedPRs) > 0 {
			sb.WriteString("<h3>Reviewed Pull Requests</h3>\n")
			for _, pr := range reviewedPRs {
				sb.WriteString("<div class=\"pr\">\n")
				
				// Add PR state class
				stateClass := "pr-state-open"
				if pr.State == "closed" {
					stateClass = "pr-state-closed"
				} else if pr.State == "merged" {
					stateClass = "pr-state-merged"
				}
				
				sb.WriteString(fmt.Sprintf("<h4><span class=\"pr-number\">#%d</span> <span class=\"pr-title\">%s</span> <span class=\"%s\">(%s)</span></h4>\n", 
					pr.Number, pr.Title, stateClass, pr.State))
				sb.WriteString(fmt.Sprintf("<p><a href=\"%s\">%s</a></p>\n", pr.URL, pr.URL))
				
				// Add reviews
				if len(pr.Reviews) > 0 {
					sb.WriteString("<div class=\"reviews\">\n")
					sb.WriteString("<h5>Reviews</h5>\n")
					for _, review := range pr.Reviews {
						sb.WriteString("<div class=\"review\">\n")
						sb.WriteString(fmt.Sprintf("<p><strong>%s</strong></p>\n", review.State))
						if review.Body != "" {
							sb.WriteString(fmt.Sprintf("<p>%s</p>\n", review.Body))
						}
						sb.WriteString(fmt.Sprintf("<p class=\"timestamp\">%s</p>\n", 
							review.Timestamp.Format("2006-01-02 15:04:05")))
						sb.WriteString("</div>\n")
					}
					sb.WriteString("</div>\n")
				}
				
				// Add comments
				if len(pr.Comments) > 0 {
					sb.WriteString("<div class=\"comments\">\n")
					sb.WriteString("<h5>Comments</h5>\n")
					for _, comment := range pr.Comments {
						sb.WriteString("<div class=\"comment\">\n")
						sb.WriteString(fmt.Sprintf("<p>%s</p>\n", comment.Body))
						sb.WriteString(fmt.Sprintf("<p class=\"timestamp\">%s</p>\n", 
							comment.Timestamp.Format("2006-01-02 15:04:05")))
						sb.WriteString("</div>\n")
					}
					sb.WriteString("</div>\n")
				}
				
				sb.WriteString("</div>\n")
			}
		}
	}
	
	// Close HTML document
	sb.WriteString("</body>\n</html>")

	return &FormattedContent{
		ContentType: "text/html",
		Content:     sb.String(),
	}, nil
}

// Helper function to check if all repositories are empty
func allRepositoriesEmpty(repositories []Repository) bool {
	for _, repo := range repositories {
		if len(repo.PullRequests) > 0 {
			return false
		}
	}
	return true
} 

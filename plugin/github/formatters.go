package github

import (
	"encoding/json"
	"fmt"
	"strings"
	
	externalGithub "github.com/google/go-github/v68/github"
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
	if len(report.Events) == 0 {
		return &FormattedContent{
			ContentType: "text/plain",
			Content:     "No GitHub activity found in specified time range for user " + report.User.Username,
		}, nil
	}

	content := "GitHub Activity Report for " + report.User.Username + ":\n\n"
	
	for _, event := range report.Events {
		if event.GetType() == "" {
			continue
		}

		eventType := event.GetType()
		repo := "unknown repository"
		if event.GetRepo() != nil {
			repo = event.GetRepo().GetName()
		}
		timestamp := event.GetCreatedAt().Format("2006-01-02 15:04:05")

		// Base event info
		eventLine := "- " + eventType + " on " + repo + " at " + timestamp
		
		// Add specific details for PushEvent
		if eventType == "PushEvent" {
			var pushEvent externalGithub.PushEvent
			err := json.Unmarshal(event.GetRawPayload(), &pushEvent)
			if err == nil {
				// Extract branch/ref information
				ref := pushEvent.GetRef()
				if ref != "" {
					// Convert refs/heads/main to just "main"
					branchName := ref
					if len(ref) > 11 && ref[:11] == "refs/heads/" {
						branchName = ref[11:]
					}
					eventLine += "\n  Branch: " + branchName
				}
				
				// Extract commit count
				commits := pushEvent.Commits
				numCommits := len(commits)
				if numCommits > 0 {
					commitWord := "commit"
					if numCommits > 1 {
						commitWord = "commits"
					}
					eventLine += fmt.Sprintf("\n  %d %s", numCommits, commitWord)
					
					// Show commit messages
					for i, commit := range commits {
						if i >= 3 { // Only show up to 3 commits to avoid too much text
							eventLine += fmt.Sprintf("\n  ... and %d more commits", numCommits-3)
							break
						}
						
						message := commit.GetMessage()
						if message != "" {
							// Truncate long messages and show only first line
							if len(message) > 60 {
								message = message[:57] + "..."
							}
							if idx := strings.Index(message, "\n"); idx != -1 {
								message = message[:idx]
							}
							eventLine += "\n    - " + message
						}
					}
				}
			}
		}
		
		content += eventLine + "\n\n"
	}

	return &FormattedContent{
		ContentType: "text/plain",
		Content:     content,
	}, nil
}
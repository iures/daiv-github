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
		
		// Add specific details for different event types
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
							// Only show the first line of the commit message
							if idx := strings.Index(message, "\n"); idx != -1 {
								message = message[:idx]
							}
							eventLine += "\n    - " + message
						}
					}
				}
			}
		} else if eventType == "PullRequestEvent" {
			var prEvent externalGithub.PullRequestEvent
			err := json.Unmarshal(event.GetRawPayload(), &prEvent)
			if err == nil {
				// Get action (opened, closed, merged, etc.)
				action := prEvent.GetAction()
				
				// Get PR details
				pr := prEvent.GetPullRequest()
				if pr != nil {
					prNumber := pr.GetNumber()
					prTitle := pr.GetTitle()
					
					// Add PR information
					eventLine += fmt.Sprintf("\n  PR #%d: %s", prNumber, prTitle)
					eventLine += fmt.Sprintf("\n  Action: %s", action)
					
					// Show merged status if PR is closed
					if action == "closed" && pr.GetMerged() {
						eventLine += "\n  Merged: Yes"
					} else if action == "closed" {
						eventLine += "\n  Merged: No (closed without merging)"
					}
					
					// Show base and head branches
					if pr.GetBase() != nil && pr.GetBase().GetRef() != "" {
						eventLine += fmt.Sprintf("\n  Base: %s", pr.GetBase().GetRef())
					}
					
					if pr.GetHead() != nil && pr.GetHead().GetRef() != "" {
						eventLine += fmt.Sprintf("\n  Head: %s", pr.GetHead().GetRef())
					}
					
					// Show number of commits, additions, deletions if available
					if pr.GetCommits() > 0 || pr.GetAdditions() > 0 || pr.GetDeletions() > 0 {
						eventLine += fmt.Sprintf("\n  Changes: %d commits, +%d -%d", 
							pr.GetCommits(), pr.GetAdditions(), pr.GetDeletions())
					}
				}
			}
		} else if eventType == "IssueCommentEvent" {
			var commentEvent externalGithub.IssueCommentEvent
			err := json.Unmarshal(event.GetRawPayload(), &commentEvent)
			if err == nil {
				issue := commentEvent.GetIssue()
				comment := commentEvent.GetComment()
				
				if issue != nil {
					issueTitle := issue.GetTitle()
					issueNumber := issue.GetNumber()
					isPR := issue.IsPullRequest()
					
					// Show type (PR or Issue)
					itemType := "Issue"
					if isPR {
						itemType = "PR"
					}
					
					// Show issue/PR title if available
					if issueTitle != "" {
						eventLine += fmt.Sprintf("\n  %s #%d: %s", itemType, issueNumber, issueTitle)
					} else {
						eventLine += fmt.Sprintf("\n  %s #%d", itemType, issueNumber)
					}
					
					// Show comment preview
					if comment != nil && comment.GetBody() != "" {
						commentBody := comment.GetBody()
						if len(commentBody) > 100 {
							// Truncate long comments and show first line
							if idx := strings.Index(commentBody, "\n"); idx != -1 && idx < 100 {
								commentBody = commentBody[:idx] + "..."
							} else {
								commentBody = commentBody[:97] + "..."
							}
						}
						eventLine += fmt.Sprintf("\n  Comment: %s", commentBody)
					}
				}
			}
		} else if eventType == "PullRequestReviewEvent" {
			var reviewEvent externalGithub.PullRequestReviewEvent
			err := json.Unmarshal(event.GetRawPayload(), &reviewEvent)
			if err == nil {
				// Get review details
				review := reviewEvent.GetReview()
				pr := reviewEvent.GetPullRequest()
				
				if pr != nil {
					prNumber := pr.GetNumber()
					prTitle := pr.GetTitle()
					
					// Show PR title and number
					eventLine += fmt.Sprintf("\n  PR #%d: %s", prNumber, prTitle)
					
					// Show review state (approved, commented, changes_requested)
					if review != nil {
						state := review.GetState()
						stateStr := "Commented"
						
						if state == "APPROVED" {
							stateStr = "Approved"
						} else if state == "CHANGES_REQUESTED" {
							stateStr = "Requested changes"
						}
						
						eventLine += fmt.Sprintf("\n  Review: %s", stateStr)
						
						// Show review body
						if review.GetBody() != "" {
							reviewBody := review.GetBody()
							if len(reviewBody) > 100 {
								// Truncate long reviews and show first line
								if idx := strings.Index(reviewBody, "\n"); idx != -1 && idx < 100 {
									reviewBody = reviewBody[:idx] + "..."
								} else {
									reviewBody = reviewBody[:97] + "..."
								}
							}
							eventLine += fmt.Sprintf("\n  Comment: %s", reviewBody)
						}
					}
				}
			}
		} else if eventType == "PullRequestReviewCommentEvent" {
			var commentEvent externalGithub.PullRequestReviewCommentEvent
			err := json.Unmarshal(event.GetRawPayload(), &commentEvent)
			if err == nil {
				// Get comment details
				comment := commentEvent.GetComment()
				pr := commentEvent.GetPullRequest()
				
				if pr != nil {
					prNumber := pr.GetNumber()
					prTitle := pr.GetTitle()
					
					// Show PR title and number
					eventLine += fmt.Sprintf("\n  PR #%d: %s", prNumber, prTitle)
					
					// Show comment details
					if comment != nil {
						// Show file path and line if available
						if comment.GetPath() != "" {
							path := comment.GetPath()
							line := comment.GetLine()
							eventLine += fmt.Sprintf("\n  File: %s:%d", path, line)
						}
						
						// Show comment body
						if comment.GetBody() != "" {
							commentBody := comment.GetBody()
							if len(commentBody) > 100 {
								// Truncate long comments and show first line
								if idx := strings.Index(commentBody, "\n"); idx != -1 && idx < 100 {
									commentBody = commentBody[:idx] + "..."
								} else {
									commentBody = commentBody[:97] + "..."
								}
							}
							eventLine += fmt.Sprintf("\n  Comment: %s", commentBody)
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
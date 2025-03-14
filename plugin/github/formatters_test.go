package github

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// createTestActivityReport creates a sample activity report for testing
func createTestActivityReport() *ActivityReport {
	timeRange := TimeRange{
		Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	return &ActivityReport{
		TimeRange: timeRange,
		User: User{
			Username: "testuser",
		},
		Repositories: []Repository{
			{
				Name: "testrepo",
				Organization: "testorg",
				PullRequests: []PullRequest{
					{
						Number:    123,
						Title:     "Test PR",
						URL:       "https://github.com/testorg/testrepo/pull/123",
						State:     "open",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC),
						Author:    "testuser",
						IsAuthored: true,
					},
				},
			},
		},
	}
}

// createEmptyActivityReport creates an empty activity report for testing
func createEmptyActivityReport() *ActivityReport {
	timeRange := TimeRange{
		Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	return &ActivityReport{
		TimeRange: timeRange,
		User: User{
			Username: "testuser",
		},
		Repositories: []Repository{},
	}
}

// TestJSONFormatter tests the JSON formatter
func TestJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()

	// Test formatter name
	if formatter.Name() != "json" {
		t.Errorf("Expected formatter name to be 'json', got '%s'", formatter.Name())
	}

	// Test formatting a non-empty report
	report := createTestActivityReport()
	content, err := formatter.Format(report)
	if err != nil {
		t.Fatalf("Error formatting report: %v", err)
	}

	// Check content type
	if content.ContentType != "application/json" {
		t.Errorf("Expected content type to be 'application/json', got '%s'", content.ContentType)
	}

	// Verify JSON can be parsed
	var parsedReport ActivityReport
	err = json.Unmarshal([]byte(content.Content), &parsedReport)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	// Check some values
	if parsedReport.User.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", parsedReport.User.Username)
	}

	// Test formatting an empty report
	emptyReport := createEmptyActivityReport()
	emptyContent, err := formatter.Format(emptyReport)
	if err != nil {
		t.Fatalf("Error formatting empty report: %v", err)
	}

	if emptyContent.Content != "{}" {
		t.Errorf("Expected empty JSON content to be '{}', got '%s'", emptyContent.Content)
	}
}

// TestMarkdownFormatter tests the Markdown formatter
func TestMarkdownFormatter(t *testing.T) {
	formatter := NewMarkdownFormatter()

	// Test formatter name
	if formatter.Name() != "markdown" {
		t.Errorf("Expected formatter name to be 'markdown', got '%s'", formatter.Name())
	}

	// Test formatting a non-empty report
	report := createTestActivityReport()
	content, err := formatter.Format(report)
	if err != nil {
		t.Fatalf("Error formatting report: %v", err)
	}

	// Check content type
	if content.ContentType != "text/markdown" {
		t.Errorf("Expected content type to be 'text/markdown', got '%s'", content.ContentType)
	}

	// Check for expected markdown elements
	expectedElements := []string{
		"# GitHub Activity Report",
		"**Time Range:**",
		"**User:** testuser",
		"## Repository: testorg/testrepo",
		"### Authored Pull Requests",
		"Test PR",
	}

	for _, element := range expectedElements {
		if !strings.Contains(content.Content, element) {
			t.Errorf("Expected markdown to contain '%s', but it doesn't", element)
		}
	}

	// Test formatting an empty report
	emptyReport := createEmptyActivityReport()
	emptyContent, err := formatter.Format(emptyReport)
	if err != nil {
		t.Fatalf("Error formatting empty report: %v", err)
	}

	if !strings.Contains(emptyContent.Content, "No GitHub activity found") {
		t.Errorf("Expected empty markdown content to mention 'No GitHub activity found', got '%s'", emptyContent.Content)
	}
}

// TestHTMLFormatter tests the HTML formatter
func TestHTMLFormatter(t *testing.T) {
	formatter := NewHTMLFormatter()

	// Test formatter name
	if formatter.Name() != "html" {
		t.Errorf("Expected formatter name to be 'html', got '%s'", formatter.Name())
	}

	// Test formatting a non-empty report
	report := createTestActivityReport()
	content, err := formatter.Format(report)
	if err != nil {
		t.Fatalf("Error formatting report: %v", err)
	}

	// Check content type
	if content.ContentType != "text/html" {
		t.Errorf("Expected content type to be 'text/html', got '%s'", content.ContentType)
	}

	// Check for expected HTML elements
	expectedElements := []string{
		"<!DOCTYPE html>",
		"<html>",
		"<head>",
		"<title>GitHub Activity Report</title>",
		"<body>",
		"<h1>GitHub Activity Report</h1>",
		"testuser",
		"testorg/testrepo",
		"Test PR",
	}

	for _, element := range expectedElements {
		if !strings.Contains(content.Content, element) {
			t.Errorf("Expected HTML to contain '%s', but it doesn't", element)
		}
	}

	// Test formatting an empty report
	emptyReport := createEmptyActivityReport()
	emptyContent, err := formatter.Format(emptyReport)
	if err != nil {
		t.Fatalf("Error formatting empty report: %v", err)
	}

	if !strings.Contains(emptyContent.Content, "No activity found") {
		t.Errorf("Expected empty HTML content to mention 'No activity found', got '%s'", emptyContent.Content)
	}
}

// TestAllRepositoriesEmpty tests the allRepositoriesEmpty helper function
func TestAllRepositoriesEmpty(t *testing.T) {
	// Test cases
	testCases := []struct {
		name         string
		repositories []Repository
		expected     bool
	}{
		{
			name:         "Empty repositories slice",
			repositories: []Repository{},
			expected:     true,
		},
		{
			name: "Repository with no PRs",
			repositories: []Repository{
				{
					Name:         "repo1",
					Organization: "org1",
					PullRequests: []PullRequest{},
				},
			},
			expected: true,
		},
		{
			name: "Repository with PRs",
			repositories: []Repository{
				{
					Name:         "repo1",
					Organization: "org1",
					PullRequests: []PullRequest{
						{
							Number: 123,
							Title:  "Test PR",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Mixed repositories",
			repositories: []Repository{
				{
					Name:         "repo1",
					Organization: "org1",
					PullRequests: []PullRequest{},
				},
				{
					Name:         "repo2",
					Organization: "org1",
					PullRequests: []PullRequest{
						{
							Number: 123,
							Title:  "Test PR",
						},
					},
				},
			},
			expected: false,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := allRepositoriesEmpty(tc.repositories)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
} 

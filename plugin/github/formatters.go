package github

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
		if event.GetType() != "" {
			content += "- " + event.GetType()
			if event.GetRepo() != nil {
				content += " on " + event.GetRepo().GetName()
			}
			content += " at " + event.GetCreatedAt().Format("2006-01-02 15:04:05") + "\n"
		}
	}

	return &FormattedContent{
		ContentType: "text/plain",
		Content:     content,
	}, nil
}

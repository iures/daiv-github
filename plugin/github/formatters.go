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
	return nil, nil
}

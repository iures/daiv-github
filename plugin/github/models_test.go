package github

import (
	"testing"
	"time"
)

func TestTimeRange_IsInRange(t *testing.T) {
	// Create a time range from 2023-01-01 to 2023-01-31
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)
	timeRange := TimeRange{
		Start: start,
		End:   end,
	}

	// Test cases
	testCases := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "Time before range",
			time:     time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: false,
		},
		{
			name:     "Time at start of range",
			time:     start,
			expected: true,
		},
		{
			name:     "Time in middle of range",
			time:     time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "Time at end of range",
			time:     end,
			expected: true,
		},
		{
			name:     "Time after range",
			time:     time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := timeRange.IsInRange(tc.time)
			if result != tc.expected {
				t.Errorf("Expected IsInRange(%v) to be %v, got %v", tc.time, tc.expected, result)
			}
		})
	}
}

func TestDefaultQueryOptions(t *testing.T) {
	options := DefaultQueryOptions()

	if options.BaseBranch != "master" {
		t.Errorf("Expected default BaseBranch to be 'master', got '%s'", options.BaseBranch)
	}

	if !options.IncludeAuthored {
		t.Errorf("Expected default IncludeAuthored to be true, got false")
	}

	if !options.IncludeReviewed {
		t.Errorf("Expected default IncludeReviewed to be true, got false")
	}

	if !options.IncludeComments {
		t.Errorf("Expected default IncludeComments to be true, got false")
	}

	if !options.IncludeCommits {
		t.Errorf("Expected default IncludeCommits to be true, got false")
	}
} 

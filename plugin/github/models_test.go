package github

import (
	"testing"
	"time"
)

func TestTimeRange_IsInRange(t *testing.T) {
	// Setup test cases
	testCases := []struct {
		name      string
		timeRange TimeRange
		testTime  time.Time
		expected  bool
	}{
		{
			name: "Time is in range",
			timeRange: TimeRange{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			},
			testTime:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name: "Time is equal to start (inclusive)",
			timeRange: TimeRange{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			},
			testTime:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:  true,
		},
		{
			name: "Time is equal to end (exclusive)",
			timeRange: TimeRange{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			},
			testTime:  time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name: "Time is before range",
			timeRange: TimeRange{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			},
			testTime:  time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
			expected:  false,
		},
		{
			name: "Time is after range",
			timeRange: TimeRange{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			},
			testTime:  time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC),
			expected:  false,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.timeRange.IsInRange(tc.testTime)
			if result != tc.expected {
				t.Errorf("Expected IsInRange to return %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestDefaultQueryOptions(t *testing.T) {
	// Get default options
	options := DefaultQueryOptions()

	// Test default values
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

	if options.MaxResults != 100 {
		t.Errorf("Expected default MaxResults to be 100, got %d", options.MaxResults)
	}
} 

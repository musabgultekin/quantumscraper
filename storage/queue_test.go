package storage

import (
	"testing"
)

func TestSelectAppropriateWorkerId(t *testing.T) {
	// Initialize a new Queue with 100 workers
	queue, err := NewQueue("/tmp/badger", 1000)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	testCases := []struct {
		name     string
		host     string
		expected int
	}{
		{
			name:     "Test case 1: Normal host",
			host:     "www.google.com",
			expected: 354, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
		{
			name:     "Test case 2: IP address",
			host:     "192.168.1.1",
			expected: 932, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
		{
			name:     "Test case 3: Empty host",
			host:     "",
			expected: 37, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
		{
			name:     "Test case 3: Long hostname",
			host:     "verylongsubdomain.imaverylongdomain.net",
			expected: 459, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
		{
			name:     "Test case 3: Long hostname 2",
			host:     "verylongsubdomain.imaverylongdomain.com",
			expected: 685, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
		{
			name:     "Test case 3: Long hostname 3",
			host:     "verylongsubdomain.imaverylongdomain.org",
			expected: 38, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
		{
			name:     "Test case 3: Long hostname 4",
			host:     "verylongsubdomain.imaverylongdomain.me",
			expected: 286, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
		{
			name:     "Test case 3: Long hostname 4",
			host:     "excessivelylongsubdomain.imaverylongdomain.me",
			expected: 612, // This is a placeholder. Replace it with the expected worker ID for this host.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workerId := queue.selectAppropriateWorkerId(tc.host)
			if workerId != tc.expected {
				t.Errorf("Expected worker ID %d, but got %d", tc.expected, workerId)
			}
		})
	}
}

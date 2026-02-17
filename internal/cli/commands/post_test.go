package commands

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{
			name:     "ShortString",
			input:    "hello",
			max:      10,
			expected: "hello",
		},
		{
			name:     "ExactLength",
			input:    "hello",
			max:      5,
			expected: "hello",
		},
		{
			name:     "LongString",
			input:    "hello world, this is a long string",
			max:      10,
			expected: "hello w...",
		},
		{
			name:     "EmptyString",
			input:    "",
			max:      10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expected)
			}
		})
	}
}

func TestTruncate_ZeroMax(t *testing.T) {
	// truncate panics when max < 3 and the string is longer than max,
	// because it tries to slice s[:max-3] which yields a negative index.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected truncate to panic with max=0 and non-empty string, but it did not")
		}
	}()
	truncate("hello", 0)
}

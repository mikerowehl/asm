package buf

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBufferStartChar(t *testing.T) {
	tests := []struct {
		input    string
		check    byte
		expected bool
	}{
		{
			input:    "abcde",
			check:    'a',
			expected: true,
		}, {
			input:    "abcde",
			check:    'b',
			expected: false,
		},
	}
	for _, tc := range tests {
		b := Buffer{tc.input}
		require.Equal(t, tc.expected, b.StartsWith(Char(tc.check)))
	}
}

func TestBufferStartString(t *testing.T) {
	tests := []struct {
		input    string
		check    string
		expected bool
	}{
		{
			input:    "abcde",
			check:    "abc",
			expected: true,
		}, {
			input:    "abcde",
			check:    "abd",
			expected: false,
		},
	}
	for _, tc := range tests {
		b := Buffer{tc.input}
		require.Equal(t, tc.expected, b.StartsWith(Str(tc.check)))
	}
}

func TestBufferScanChar(t *testing.T) {
	tests := []struct {
		input    string
		check    byte
		expected int
	}{
		{
			input:    "aaade",
			check:    'a',
			expected: 3,
		}, {
			input:    "abcde",
			check:    'a',
			expected: 1,
		}, {
			input:    "abcde",
			check:    'b',
			expected: 0,
		},
	}
	for _, tc := range tests {
		b := Buffer{tc.input}
		require.Equal(t, tc.expected, b.Scan(Char(tc.check)))
	}
}

func TestBufferTakeWhile(t *testing.T) {
	tests := []struct {
		input         string
		compareFn     Compare
		expectedTaken string
		expectedLeft  string
	}{
		{
			input:         "aaaabcd",
			compareFn:     Char('a'),
			expectedTaken: "aaaa",
			expectedLeft:  "bcd",
		}, {
			input:         "abcd efgh",
			compareFn:     Letter,
			expectedTaken: "abcd",
			expectedLeft:  " efgh",
		}, {
			input:         " \t  abcd",
			compareFn:     Whitespace,
			expectedTaken: " \t  ",
			expectedLeft:  "abcd",
		}, {
			input:         "a][;b cd",
			compareFn:     Word,
			expectedTaken: "a][;b",
			expectedLeft:  " cd",
		},
	}
	for _, tc := range tests {
		b := Buffer{s: tc.input}
		taken, left := b.TakeWhile(tc.compareFn)
		require.Equal(t, tc.expectedTaken, taken.String())
		require.Equal(t, tc.expectedLeft, left.String())
	}
}

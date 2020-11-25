package expr

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
		b := buffer{tc.input}
		require.Equal(t, tc.expected, b.startsWith(char(tc.check)))
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
		b := buffer{tc.input}
		require.Equal(t, tc.expected, b.startsWith(str(tc.check)))
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
		b := buffer{tc.input}
		require.Equal(t, tc.expected, b.scan(char(tc.check)))
	}
}

func TestBufferTakeWhile(t *testing.T) {
	tests := []struct {
		input         string
		compareFn     compare
		expectedTaken string
		expectedLeft  string
	}{
		{
			input:         "aaaabcd",
			compareFn:     char('a'),
			expectedTaken: "aaaa",
			expectedLeft:  "bcd",
		}, {
			input:         "abcd efgh",
			compareFn:     letter,
			expectedTaken: "abcd",
			expectedLeft:  " efgh",
		}, {
			input:         " \t  abcd",
			compareFn:     whitespace,
			expectedTaken: " \t  ",
			expectedLeft:  "abcd",
		}, {
			input:         "a][;b cd",
			compareFn:     word,
			expectedTaken: "a][;b",
			expectedLeft:  " cd",
		},
	}
	for _, tc := range tests {
		b := buffer{s: tc.input}
		taken, left := b.takeWhile(tc.compareFn)
		require.Equal(t, tc.expectedTaken, taken.String())
		require.Equal(t, tc.expectedLeft, left.String())
	}
}

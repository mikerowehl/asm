package expr

import (
	"github.com/mikerowehl/asm/buf"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseNumber(t *testing.T) {
	tests := []struct {
		input          string
		expectedVal    int
		expectedRemain string
		expectedErr    bool
	}{
		{
			input:          "1234",
			expectedVal:    1234,
			expectedRemain: "",
			expectedErr:    false,
		}, {
			input:          "0x1234",
			expectedVal:    0x1234,
			expectedRemain: "",
			expectedErr:    false,
		}, {
			input:          "0xw",
			expectedVal:    0,
			expectedRemain: "w",
			expectedErr:    true,
		},
	}
	for _, tc := range tests {
		p := Parser{}
		b := buf.NewBuffer(tc.input)
		v, r, e := p.parseNumber(b)
		require.Equal(t, tc.expectedVal, v)
		require.Equal(t, tc.expectedRemain, r.String())
		require.Equal(t, tc.expectedErr, e != nil)
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		input          string
		expectedVal    string
		expectedRemain string
		expectedErr    bool
	}{
		{
			input:          "\"abcd\"",
			expectedVal:    "abcd",
			expectedRemain: "",
			expectedErr:    false,
		}, {
			input:          "\"0x1234\"",
			expectedVal:    "0x1234",
			expectedRemain: "",
			expectedErr:    false,
		}, {
			input:          "\"abc",
			expectedVal:    "abc",
			expectedRemain: "",
			expectedErr:    true,
		},
	}
	for _, tc := range tests {
		p := Parser{}
		b := buf.NewBuffer(tc.input)
		v, r, e := p.parseString(b)
		require.Equal(t, tc.expectedVal, v)
		require.Equal(t, tc.expectedRemain, r.String())
		require.Equal(t, tc.expectedErr, e != nil)
	}
}

func TestParseToken(t *testing.T) {
	tests := []struct {
		input          string
		expectedToken  Token
		expectedRemain string
	}{
		{
			input:         "1234",
			expectedToken: Token{typ: tokenNumber, value: 1234},
		}, {
			input:          "1234 and some other stuff",
			expectedToken:  Token{typ: tokenNumber, value: 1234},
			expectedRemain: "and some other stuff",
		}, {
			input:          "\"string\" and some other stuff",
			expectedToken:  Token{typ: tokenString, stringValue: "string"},
			expectedRemain: "and some other stuff",
		}, {
			input:          "- and some other stuff",
			expectedToken:  Token{typ: tokenOp, op: opUnaryNeg},
			expectedRemain: "and some other stuff",
		},
	}
	for _, tc := range tests {
		p := Parser{}
		b := buf.NewBuffer(tc.input)
		tok, r, e := p.parseToken(b)
		require.Nil(t, e)
		require.Equal(t, tc.expectedToken, tok)
		require.Equal(t, tc.expectedRemain, r.String())
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input          string
		expectedNode   Node
		expectedRemain string
	}{
		{
			input:          "1+2",
			expectedNode:   Node{op: opAdd},
			expectedRemain: "",
		},
	}
	for _, tc := range tests {
		p := Parser{}
		b := buf.NewBuffer(tc.input)
		n, r, e := p.Parse(b)
		require.Nil(t, e)
		require.Equal(t, tc.expectedNode.op, n.op)
		require.Equal(t, tc.expectedRemain, r.String())
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			input:    "1+2",
			expected: 3,
		}, {
			input:    "3-1",
			expected: 2,
		}, {
			input:    "1+2+3",
			expected: 6,
		}, {
			input:    "-4",
			expected: -4,
		}, {
			input:    "4-(2-1)",
			expected: 3,
		},
	}
	for _, tc := range tests {
		p := Parser{}
		b := buf.NewBuffer(tc.input)
		n, _, e := p.Parse(b)
		require.Nil(t, e)
		eval := n.Eval(map[string]int{})
		require.True(t, eval)
		require.Equal(t, tc.expected, n.value)
	}
}

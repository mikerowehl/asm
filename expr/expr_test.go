package expr

import (
	"github.com/stretchr/testify/require"
	"testing"
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
		b := buffer{s: tc.input}
		v, r, e := p.parseNumber(b)
		require.Equal(t, tc.expectedVal, v)
		require.Equal(t, tc.expectedRemain, r.s)
		require.Equal(t, tc.expectedErr, e != nil)
	}
}

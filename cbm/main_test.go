package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstructionString(t *testing.T) {
	t1, e1 := ToInstruction("ADC")
	require.Nil(t, e1)
	require.Equal(t, t1.String(), "ADC")
	_, e2 := ToInstruction("XXX")
	require.NotNil(t, e2)
}

func TestStripComment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "one two ; comment four",
			expected: "one two",
		}, {
			input:    "one two \t ; comment four",
			expected: "one two",
		}, {
			input:    "; one two",
			expected: "",
		}, {
			input:    "one two three",
			expected: "one two three",
		}, {
			input:    "one two \t ",
			expected: "one two",
		},
	}
	for _, tc := range tests {
		require.Equal(t, tc.expected, stripComment(tc.input))
	}
}

func TestAssembleInstruction(t *testing.T) {
	tests := []struct {
		instruction inst
		forms       []MachineCode
		expected    []uint8
	}{
		{
			instruction: inst{
				op:   ADC,
				args: args{imm: nil, addr: nil},
			},
			forms: []MachineCode{
				{opcode: 0x01, mode: Implied},
			},
			expected: []uint8{0x01},
		},
		{
			instruction: inst{
				op:   LDA,
				args: args{imm: &val{imm: 0xab}, addr: nil},
			},
			forms: []MachineCode{
				{opcode: 0x02, mode: Immediate},
			},
			expected: []uint8{0x02, 0xab},
		},
	}

	for i, tc := range tests {
		err := assembleInstruction(&tc.instruction, tc.forms)
		require.Nil(t, err)
		require.Equal(t, tc.expected, tc.instruction.chunk.mem, "Failed test %d", i)
	}
}

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
		{
			instruction: inst{
				op:   ASL,
				args: args{reg: 1},
			},
			forms: []MachineCode{
				{opcode: 0x03, mode: Accumulator},
			},
			expected: []uint8{0x03},
		},
		{
			instruction: inst{
				op:   AND,
				args: args{addr: &val{imm: 0x1234}},
			},
			forms: []MachineCode{
				{opcode: 0x04, mode: Absolute},
			},
			expected: []uint8{0x04, 0x12, 0x34},
		},
		{
			instruction: inst{
				op:   AND,
				args: args{addr: &val{imm: 0x1234}, reg: RegX},
			},
			forms: []MachineCode{
				{opcode: 0x05, mode: AbsoluteXIndex},
			},
			expected: []uint8{0x05, 0x12, 0x34},
		},
		{
			instruction: inst{
				op:   AND,
				args: args{addr: &val{imm: 0x1234}, reg: RegY},
			},
			forms: []MachineCode{
				{opcode: 0x06, mode: AbsoluteYIndex},
			},
			expected: []uint8{0x06, 0x12, 0x34},
		},
		{
			instruction: inst{
				op:   ADC,
				args: args{addr: &val{imm: 0x34}},
			},
			forms: []MachineCode{
				{opcode: 0x07, mode: Zeropage},
			},
			expected: []uint8{0x07, 0x34},
		},
		{
			instruction: inst{
				op:   ADC,
				args: args{addr: &val{imm: 0x54}, reg: RegX},
			},
			forms: []MachineCode{
				{opcode: 0x08, mode: ZeropageXIndexed},
			},
			expected: []uint8{0x08, 0x54},
		},
		{
			instruction: inst{
				op:   ADC,
				args: args{addr: &val{imm: 0x56}, reg: RegY},
			},
			forms: []MachineCode{
				{opcode: 0x09, mode: ZeropageYIndexed},
			},
			expected: []uint8{0x09, 0x56},
		},
		{
			instruction: inst{
				op:   ADC,
				args: args{addr: &val{imm: 0x2345}, ind: true},
			},
			forms: []MachineCode{
				{opcode: 0x0a, mode: Indirect},
			},
			expected: []uint8{0x0a, 0x23, 0x45},
		},
		{
			instruction: inst{
				op:   ADC,
				args: args{addr: &val{imm: 0x3456}, ind: true, reg: RegX},
			},
			forms: []MachineCode{
				{opcode: 0x0b, mode: XIndexedIndirect},
			},
			expected: []uint8{0x0b, 0x34, 0x56},
		},
		{
			instruction: inst{
				op:   ADC,
				args: args{addr: &val{imm: 0x4567}, ind: true, reg: RegY},
			},
			forms: []MachineCode{
				{opcode: 0x0c, mode: IndirectYIndexed},
			},
			expected: []uint8{0x0c, 0x45, 0x67},
		},
		{
			instruction: inst{
				op:   JMP,
				args: args{addr: &val{imm: -2}},
			},
			forms: []MachineCode{
				{opcode: 0x0d, mode: Relative},
			},
			expected: []uint8{0x0d, 0xfe},
		},
	}

	for i, tc := range tests {
		err := assembleInstruction(&tc.instruction, tc.forms)
		require.Nil(t, err)
		require.Equal(t, tc.expected, tc.instruction.chunk.mem, "Failed test %d", i)
	}
}

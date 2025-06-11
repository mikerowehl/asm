package main

import (
	"strings"
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

func TestAssemble(t *testing.T) {
	a := assembler{}
	err := a.parseReader(strings.NewReader(" LDA #4"))
	if err != nil {
		t.Fatal("Error from parseReader")
	}
	i := a.prg[0]
	require.Equal(t, i.op, LDA)
	e := i.operands.e.Eval(map[string]int{})
	if !e {
		t.Fatal("Error evaluating expression")
	}
	val, err := i.operands.e.Value()
	if err != nil {
		t.Fatal("Error getting value of expression")
	}
	require.Equal(t, val, 4)
}

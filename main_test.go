package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstructionString(t *testing.T) {
	t1, e1 := ToInstruction("ADC")
	require.Nil(t, e1)
	require.Equal(t, t1, ADC)
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
	require.Equal(t, LDA, i.op)
	eval, err := i.operands.e.Eval(map[string]int{})
	if !eval {
		t.Fatal("Error evaluating expression")
	}
	if err != nil {
		t.Fatal("Error evaluating expression")
	}
	val, err := i.operands.e.Value()
	if err != nil {
		t.Fatal("Error getting value of expression")
	}
	require.Equal(t, 4, val)
	require.Equal(t, uint8(0xa9), i.chunk.mem[0])
}

func TestConst(t *testing.T) {
	a := assembler{
		constants: make(map[string]int),
	}
	err := a.parseReader(strings.NewReader("TESTVAL=1234"))
	if err != nil {
		t.Fatal("Error from parseReader")
	}
	require.Equal(t, a.constants["TESTVAL"], 1234)
}

func TestMultiConst(t *testing.T) {
	a := assembler{
		constants: make(map[string]int),
	}
	err := a.parseReader(strings.NewReader("TESTONE\nTESTVAL=345"))
	if err != nil {
		t.Fatal("Error from parseReader")
	}
	require.Equal(t, a.constants["TESTVAL"], 345)
	require.Equal(t, a.constants["TESTONE"], 345)
}

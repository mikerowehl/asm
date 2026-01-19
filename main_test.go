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

func TestParseInstruction(t *testing.T) {
	a := assembler{}
	err := a.parseReader(strings.NewReader(" LDA #4"))
	require.Nil(t, err)
	node := a.prg[0]
	require.IsType(t, &InstructionNode{}, node)
	in := node.(*InstructionNode)
	require.NotNil(t, in.inst)
	require.Equal(t, LDA, in.inst.op)
	eval, err := in.inst.operands.e.Eval(map[string]int{})
	if !eval {
		t.Fatal("Error evaluating expression")
	}
	if err != nil {
		t.Fatal("Error evaluating expression")
	}
	val, err := in.inst.operands.e.Value()
	if err != nil {
		t.Fatal("Error getting value of expression")
	}
	require.Equal(t, 4, val)
}

func TestParseLabel(t *testing.T) {
	a := assembler{}
	err := a.parseReader(strings.NewReader("TESTLABEL:"))
	require.Nil(t, err)
	node := a.prg[0]
	require.IsType(t, &LabelNode{}, node)
	ln := node.(*LabelNode)
	require.Equal(t, "TESTLABEL", ln.Name)
}

func TestParsePseudo(t *testing.T) {
	a := assembler{}
	err := a.parseReader(strings.NewReader(" .org"))
	require.Nil(t, err)
	node := a.prg[0]
	require.IsType(t, &PseudoNode{}, node)
	pn := node.(*PseudoNode)
	require.Equal(t, PseudoOrg, pn.Pseudo.Kind)
}

func TestParsePseudoMultiarg(t *testing.T) {
	a := assembler{}
	err := a.parseReader(strings.NewReader(" .BYTE 1,2,3"))
	require.Nil(t, err)
	node := a.prg[0]
	require.IsType(t, &PseudoNode{}, node)
	pn := node.(*PseudoNode)
	require.Equal(t, PseudoByte, pn.Pseudo.Kind)
	require.Len(t, pn.Pseudo.Args, 3)
}

/*
func TestImmediateExpr(t *testing.T) {
	a := assembler{}
	err := a.parseReader(strings.NewReader(" LDA #(2+4)"))
	if err != nil {
		t.Fatal("Error from parseReader")
	}
	bytes := a.binaryImage()
	require.Equal(t, bytes[0], uint8(0xa9))
	require.Equal(t, bytes[1], uint8(6))
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
*/

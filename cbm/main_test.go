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
	require.Equal(t, "one two", stripComment("one two ; comment four"))
	require.Equal(t, "one two", stripComment("one two \t ; comment four"))
	require.Equal(t, "", stripComment("; one two comment four"))
	require.Equal(t, "one two three", stripComment("one two three"))
	require.Equal(t, "one two", stripComment("one two \t "))
}
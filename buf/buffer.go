// Package buf provides convenient operations for lexing line oriented data
package buf

import "strings"

type Buffer struct {
	s string
}

func NewBuffer(s string) Buffer {
	return Buffer{s: s}
}

func (b Buffer) String() string {
	return b.s
}

// Advance the buffer by i characters. Returns the new buffer with the initial
// characters dropped.
func (b Buffer) Advance(i int) Buffer {
	return Buffer{s: b.s[i:]}
}

// Truncates a buffer to i characters long. Returns the new truncated buffer.
func (b Buffer) Trunc(i int) Buffer {
	return Buffer{s: b.s[:i]}
}

func (b Buffer) IsEmpty() bool {
	return len(b.s) == 0
}

// The interface used by buffer operations that require checking the current
// value of a buffer against some expected format. The Compare function
// passed to those calls is invoked with the whole string content of the
// buffer so that the compare function can check against it. Some Compare
// functions are provided directly to be used as is:
//
//	if line.StartsWith(buf.Whitespace) {
//
// While others are calls that return compare functions:
//
//	label, remain := line.TakeUntil(buf.Char(':'))
type Compare func(s string) bool

func (b Buffer) StartsWith(fn Compare) bool {
	return len(b.s) > 0 && fn(b.s)
}

// Creates a Compare function that checks to see if the current first byte of
// the buffer matches a single bype passes as in.
func Char(in byte) Compare {
	return func(s string) bool {
		return s[0] == in
	}
}

// Creates a Compare function that checks if the current start of the buffer
// matches the string in.
func Str(in string) Compare {
	return func(s string) bool {
		return strings.HasPrefix(s, in)
	}
}

// A Compare function that matches whitespace
func Whitespace(s string) bool {
	return s[0] == ' ' || s[0] == '\t'
}

// A Compare function that matches anything that isn't whitespace
func Word(s string) bool {
	return !Whitespace(s)
}

// A Compare function that checks for any lower or uppercase letter
func Letter(s string) bool {
	return (s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z')
}

// A Compare function that looks for any digit 0-9
func Digit(s string) bool {
	return s[0] >= '0' && s[0] <= '9'
}

// A Compare function that checks for digits 0-9 or upper or lowercase hex
// digits a-f
func HexDigit(s string) bool {
	return Digit(s) || (s[0] >= 'A' && s[0] <= 'F') || (s[0] >= 'a' && s[0] <= 'f')
}

// Return the first index into the buffer where the Compare function fn
// doesn't match the content.
func (b Buffer) Scan(fn Compare) (i int) {
	for i = 0; i < len(b.s) && fn(b.s[i:]); i++ {
	}
	return
}

// Return the first index into buffere where the Compare function matches
func (b Buffer) ScanUntil(fn Compare) (i int) {
	for i = 0; i < len(b.s) && !fn(b.s[i:]); i++ {
	}
	return
}

// Take the start of the buffer up to the point where the Compare function fn
// no longer matches. Returns the buffer split into two parts, the taken part
// is the intial section that matched and left is everything after.
func (b Buffer) TakeWhile(fn Compare) (taken Buffer, left Buffer) {
	i := b.Scan(fn)
	taken = b.Trunc(i)
	left = b.Advance(i)
	return
}

// Split the buffer into two sections, determined by the first place where the
// Compare function matches. taken is the initial section before the first
// match and left is everything after.
func (b Buffer) TakeUntil(fn Compare) (taken Buffer, left Buffer) {
	i := b.ScanUntil(fn)
	taken = b.Trunc(i)
	left = b.Advance(i)
	return
}

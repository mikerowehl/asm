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

func (b Buffer) Advance(i int) Buffer {
	return Buffer{s: b.s[i:]}
}

func (b Buffer) Trunc(i int) Buffer {
	return Buffer{s: b.s[:i]}
}

func (b Buffer) IsEmpty() bool {
	return len(b.s) == 0
}

type Compare func(s string) bool

func (b Buffer) StartsWith(fn Compare) bool {
	return len(b.s) > 0 && fn(b.s)
}

// Some of these functions return a function that satisfies the compare
// signature, so for example scanning the buffer while the current character
// is 'a' would be:
//   b.scan(char('a'))
// while scanning for a class of characters has a provided function, like
// scanning while the current character is whitespace:
//   b.scan(whitespace)

func Char(in byte) Compare {
	return func(s string) bool {
		return s[0] == in
	}
}

func Str(in string) Compare {
	return func(s string) bool {
		return strings.HasPrefix(s, in)
	}
}

func Whitespace(s string) bool {
	return s[0] == ' ' || s[0] == '\t'
}

func Word(s string) bool {
	return !Whitespace(s)
}

func Letter(s string) bool {
	return (s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z')
}

func Digit(s string) bool {
	return s[0] >= '0' && s[0] <= '9'
}

func HexDigit(s string) bool {
	return Digit(s) || (s[0] >= 'A' && s[0] <= 'F') || (s[0] >= 'a' && s[0] <= 'f')
}

func (b Buffer) Scan(fn Compare) (i int) {
	for i = 0; i < len(b.s) && fn(b.s[i:]); i++ {
	}
	return
}

func (b Buffer) ScanUntil(fn Compare) (i int) {
	for i = 0; i < len(b.s) && !fn(b.s[i:]); i++ {
	}
	return
}

func (b Buffer) TakeWhile(fn Compare) (taken Buffer, left Buffer) {
	i := b.Scan(fn)
	taken = b.Trunc(i)
	left = b.Advance(i)
	return
}

func (b Buffer) TakeUntil(fn Compare) (taken Buffer, left Buffer) {
	i := b.ScanUntil(fn)
	taken = b.Trunc(i)
	left = b.Advance(i)
	return
}

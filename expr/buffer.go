package expr

import "strings"

type buffer struct {
	s string
}

func (b buffer) String() string {
	return b.s
}

func (b buffer) advance(i int) buffer {
	return buffer{s: b.s[i:]}
}

func (b buffer) trunc(i int) buffer {
	return buffer{s: b.s[:i]}
}

func (b buffer) isEmpty() bool {
	return len(b.s) == 0
}

type compare func(s string) bool

func (b buffer) startsWith(fn compare) bool {
	return len(b.s) > 0 && fn(b.s)
}

// Some of these functions return a function that satisfies the compare
// signature, so for example scanning the buffer while the current character
// is 'a' would be:
//   b.scan(char('a'))
// while scanning for a class of characters has a provided function, like
// scanning while the current character is whitespace:
//   b.scan(whitespace)

func char(in byte) compare {
	return func(s string) bool {
		return s[0] == in
	}
}

func str(in string) compare {
	return func(s string) bool {
		return strings.HasPrefix(s, in)
	}
}

func whitespace(s string) bool {
	return s[0] == ' ' || s[0] == '\t'
}

func word(s string) bool {
	return !whitespace(s)
}

func letter(s string) bool {
	return (s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z')
}

func (b buffer) scan(fn compare) (i int) {
	for i = 0; i < len(b.s) && fn(b.s[i:]); i++ {
	}
	return
}

func (b buffer) scanUntil(fn compare) (i int) {
	for i = 0; i < len(b.s) && !fn(b.s[i:]); i++ {
	}
	return
}

func (b buffer) takeWhile(fn compare) (taken buffer, left buffer) {
	i := b.scan(fn)
	taken = b.trunc(i)
	left = b.advance(i)
	return
}

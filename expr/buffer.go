package expr

import "strings"

type buffer struct {
	s string
}

func (b buffer) String() string {
	return b.s
}

func (b buffer) advance(n int) buffer {
	return buffer{b.s[n:]}
}

func (b buffer) isEmpty() bool {
	return len(b.s) == 0
}

type startCheck func(s string) bool

func (b buffer) startsWith(fn startCheck) bool {
	return len(b.s) > 0 && fn(b.s)
}

func char(in byte) startCheck {
	return func(s string) bool {
		return s[0] == in
	}
}

func str(in string) startCheck {
	return func(s string) bool {
		return strings.HasPrefix(s, in)
	}
}

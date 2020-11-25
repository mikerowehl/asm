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

type compare func(s string) bool

func (b buffer) startsWith(fn compare) bool {
	return len(b.s) > 0 && fn(b.s)
}

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

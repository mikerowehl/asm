package expr

import "fmt"

type nodeStack struct {
	data []*node
}

func (s *nodeStack) isEmpty() bool {
	return len(s.data) == 0
}

func (s *nodeStack) push(n *node) {
	s.data = append(s.data, n)
}

func (s *nodeStack) pop() (n *node, err error) {
	l := len(s.data)
	if l == 0 {
		err = fmt.Errorf("Attempt to pop from empty stack: %v", s)
		return
	}
	n = s.data[l-1]
	s.data = s.data[:l-1]
	return
}

func (s *nodeStack) peek() (n *node) {
	l := len(s.data)
	if l == 0 {
		return nil
	}
	return s.data[l-1]
}

func (s *nodeStack) tree(op Op) (err error) {
	if len(s.data) < 2 {
		err = fmt.Errorf("Attempt to tree with too few nodes: %v", s)
	}
	rc, err := s.pop()
	if err != nil {
		return
	}
	lc, err := s.pop()
	if err != nil {
		return
	}
	n := &node{
		op:     op,
		rChild: rc,
		lChild: lc,
	}
	s.push(n)
	return
}

type opStack struct {
	data []Op
}

func (s *opStack) isEmpty() bool {
	return len(s.data) == 0
}

func (s *opStack) push(op Op) {
	s.data = append(s.data, op)
}

func (s *opStack) pop() (op Op, err error) {
	l := len(s.data)
	if l == 0 {
		err = fmt.Errorf("Attempt to pop from empty op stack: %v", s)
		return
	}
	op = s.data[l-1]
	s.data = s.data[:l-1]
	return
}

func (s *opStack) peek() (op Op) {
	l := len(s.data)
	if l == 0 {
		op = 0
		return
	}
	op = s.data[l-1]
	return
}
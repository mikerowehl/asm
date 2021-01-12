package expr

import (
	"fmt"
	"github.com/mikerowehl/asm/buf"
	"strconv"
)

type Op int

type TokenType int

const (
	tokenNil TokenType = iota
	tokenNumber
	tokenString
	tokenIdentifier
	tokenLeftParen
	tokenRightParen
	tokenOp
)

func (t TokenType) canPrecedeUnary() bool {
	return t == tokenOp || t == tokenLeftParen || t == tokenNil
}

type Token struct {
	typ         TokenType
	value       int
	stringValue string
	identifier  string
	op          Op
}

type Parser struct {
	nodeStack     nodeStack
	opStack       opStack
	prevTokenType TokenType
}

type opEval func(a int, b int) int

type opEntry struct {
	precedence int
	childCount int
	leftAssoc  bool
	sym        string
	eval       opEval
}

func (e opEntry) isBinary() bool {
	return e.childCount == 2
}

func (e opEntry) isUnary() bool {
	return e.childCount == 1
}

func (e opEntry) parseable() bool {
	return len(e.sym) > 0
}

const (
	opUnaryNeg Op = iota
	opUnaryPlus
	opLowByte
	opHighByte

	opDivide
	opMultiply
	opAdd
	opSub
	opLeftShift
	opRightShift
	opAnd
	opOr

	opNumber
	opString
	opLeftParen
	opRightParen
)

var opTable = []opEntry{
	{6, 1, false, "-", func(a int, b int) int { return -a }},
	{6, 1, false, "+", func(a int, b int) int { return a }},
	{6, 1, false, "<", func(a int, b int) int { return a & 0xff }},
	{6, 1, false, ">", func(a int, b int) int { return (a >> 8) & 0xff }},

	{5, 2, true, "/", func(a int, b int) int { return a / b }},
	{5, 2, true, "*", func(a int, b int) int { return a * b }},
	{4, 2, true, "+", func(a int, b int) int { return a + b }},
	{4, 2, true, "-", func(a int, b int) int { return a - b }},
	{3, 2, true, "<<", func(a int, b int) int { return a << b }},
	{3, 2, true, ">>", func(a int, b int) int { return a >> b }},
	{2, 2, true, "&", func(a int, b int) int { return a & b }},
	{1, 2, true, "|", func(a int, b int) int { return a | b }},

	{0, 0, false, "", nil}, // num
	{0, 0, false, "", nil}, // string
	{0, 0, false, "", nil}, // left paren
	{0, 0, false, "", nil}, // righ paren
}

func (op Op) isBinary() bool {
	return opTable[op].childCount == 2
}

func (op Op) isUnary() bool {
	return opTable[op].childCount == 1
}

func (op Op) eval(a int, b int) int {
	return opTable[op].eval(a, b)
}

func (op Op) sym() string {
	return opTable[op].sym
}

func (op Op) isTreeable() bool {
	return opTable[op].precedence > 0
}

func (op Op) canTree(other Op) bool {
	if opTable[op].leftAssoc {
		return opTable[op].precedence <= opTable[other].precedence
	}
	return opTable[op].precedence < opTable[other].precedence
}

type Node struct {
	op        Op
	value     int
	evaluated bool
	lChild    *Node
	rChild    *Node
}

func (n *Node) String() string {
	switch {
	case n.op == opNumber:
		return fmt.Sprintf("%d", n.value)
	case n.op.isBinary():
		return fmt.Sprintf("%s %s %s", n.lChild.String(), n.rChild.String(), n.op.sym())
	case n.op.isUnary():
		return fmt.Sprintf("%s %s", n.lChild.String(), n.op.sym())
	default:
		return "unknown"
	}
}

func (n *Node) Eval(sym map[string]int) bool {
	if !n.evaluated {
		switch {
		case n.op == opNumber:
			n.evaluated = true
		case n.op.isBinary():
			n.lChild.Eval(sym)
			n.rChild.Eval(sym)
			n.value = opTable[n.op].eval(n.lChild.value, n.rChild.value)
			n.evaluated = true
		case n.op.isUnary():
			n.lChild.Eval(sym)
			n.value = opTable[n.op].eval(n.lChild.value, 0)
			n.evaluated = true
		}
	}
	return n.evaluated
}

func (n *Node) Value() (int, error) {
	if !n.evaluated {
		return 0, fmt.Errorf("Attempt to take value of unevaluated expression")
	}
	return n.value, nil
}

func (p *Parser) Parse(line buf.Buffer) (n *Node, remain buf.Buffer, err error) {
	p.prevTokenType = tokenNil
	for err == nil {
		var token Token
		token, remain, err = p.parseToken(line)
		if err != nil {
			break
		}

		if token.typ == tokenNil {
			break
		}

		switch token.typ {
		case tokenNumber:
			cur := &Node{
				op:        opNumber,
				value:     token.value,
				evaluated: true,
			}
			p.nodeStack.push(cur)

		case tokenOp:
			for err == nil && !p.opStack.isEmpty() && token.op.canTree(p.opStack.peek()) {
				var treeOp Op
				treeOp, err = p.opStack.pop()
				if err != nil {
					return
				}
				err = p.nodeStack.tree(treeOp)
				if err != nil {
					return
				}
			}
			p.opStack.push(token.op)
		case tokenLeftParen:
			p.opStack.push(opLeftParen)
		case tokenRightParen:
			for err == nil {
				if p.opStack.isEmpty() {
					err = fmt.Errorf("Mismatched parens")
					return
				}
				var treeOp Op
				treeOp, err = p.opStack.pop()
				if err != nil {
					return
				}
				if treeOp == opLeftParen {
					break
				}
				err = p.nodeStack.tree(treeOp)
				if err != nil {
					return
				}
			}
		}
		line = remain
	}

	for err == nil && !p.opStack.isEmpty() {
		op, err := p.opStack.pop()
		if err != nil {
			return nil, buf.Buffer{}, err
		}
		err = p.nodeStack.tree(op)
		if err != nil {
			return nil, buf.Buffer{}, err
		}
	}

	n, err = p.nodeStack.pop()
	return
}

func (p *Parser) parseToken(line buf.Buffer) (t Token, remain buf.Buffer, err error) {
	if line.IsEmpty() {
		t.typ = tokenNil
		return
	}

	switch {
	case line.StartsWith(buf.Digit) || line.StartsWith(buf.Char('$')):
		t.value, remain, err = p.parseNumber(line)
		t.typ = tokenNumber
	case line.StartsWith(buf.Char('"')):
		t.stringValue, remain, err = p.parseString(line)
		t.typ = tokenString
	case line.StartsWith(buf.Char('(')):
		t.typ = tokenLeftParen
		t.op = opLeftParen
		remain = line.Advance(1)
	case line.StartsWith(buf.Char(')')):
		t.typ = tokenRightParen
		t.op = opRightParen
		remain = line.Advance(1)
	default:
		for i, o := range opTable {
			if o.parseable() && line.StartsWith(buf.Str(o.sym)) {
				if o.isBinary() || (o.isUnary() && p.prevTokenType.canPrecedeUnary()) {
					t.typ = tokenOp
					t.op = Op(i)
					remain = line.Advance(len(o.sym))
					break
				}
			}
		}
		if t.typ != tokenOp {
			err = fmt.Errorf("Invalid operation: %s", line)
			return
		}
	}

	p.prevTokenType = t.typ
	remain = remain.Advance(remain.Scan(buf.Whitespace))
	return
}

func (p *Parser) identifyNumber(line buf.Buffer) (remain buf.Buffer, base int,
	digitFn buf.Compare) {
	remain = line
	if remain.StartsWith(buf.Char('$')) {
		remain = remain.Advance(1)
		base = 16
		digitFn = buf.HexDigit
		return
	} else if remain.StartsWith(buf.Str("0x")) {
		remain = remain.Advance(2)
		base = 16
		digitFn = buf.HexDigit
		return
	}
	base = 10
	digitFn = buf.Digit
	return
}

func (p *Parser) parseNumber(line buf.Buffer) (value int, remain buf.Buffer,
	err error) {
	line, base, digitFn := p.identifyNumber(line)

	str, remain := line.TakeWhile(digitFn)
	num, err := strconv.ParseInt(str.String(), base, 32)
	if err != nil {
		return
	}

	value = int(num)
	return
}

func (p *Parser) parseString(line buf.Buffer) (value string, remain buf.Buffer,
	err error) {
	remain = line.Advance(1)
	valueBuf, remain := remain.TakeUntil(buf.Char('"'))
	value = valueBuf.String()
	if remain.IsEmpty() {
		err = fmt.Errorf("Unterminated string in: %s", line)
		return
	}
	remain = remain.Advance(1)
	return
}

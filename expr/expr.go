package expr

import (
	"fmt"
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

type node struct {
	op        Op
	value     int
	evaluated bool
	lChild    *node
	rChild    *node
}

func (n *node) eval(sym map[string]int) bool {
	if !n.evaluated {
		switch {
		case n.op == opNumber:
			n.evaluated = true
		default:
			n.lChild.eval(sym)
			n.rChild.eval(sym)
			n.value = opTable[n.op].eval(n.lChild.value, n.rChild.value)
			n.evaluated = true
		}
	}
	return n.evaluated
}

func (p *Parser) parse(line buffer) (n *node, remain buffer, err error) {
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
			cur := &node{
				op:        opNumber,
				value:     token.value,
				evaluated: true,
			}
			p.nodeStack.push(cur)

		case tokenOp:
			p.opStack.push(token.op)
		}
		line = remain
	}

	for err == nil && !p.opStack.isEmpty() {
		op, err := p.opStack.pop()
		if err != nil {
			return nil, buffer{}, err
		}
		err = p.nodeStack.tree(op)
		if err != nil {
			return nil, buffer{}, err
		}
	}

	n, err = p.nodeStack.pop()
	return
}

func (p *Parser) parseToken(line buffer) (t Token, remain buffer, err error) {
	if line.isEmpty() {
		t.typ = tokenNil
		return
	}

	switch {
	case line.startsWith(digit) || line.startsWith(char('$')):
		t.value, remain, err = p.parseNumber(line)
		t.typ = tokenNumber
	case line.startsWith(char('"')):
		t.stringValue, remain, err = p.parseString(line)
		t.typ = tokenString
	default:
		for i, o := range opTable {
			if o.parseable() && line.startsWith(str(o.sym)) {
				if o.isBinary() || (o.isUnary() && p.prevTokenType.canPrecedeUnary()) {
					t.typ = tokenOp
					t.op = Op(i)
					remain = line.advance(len(o.sym))
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
	remain = remain.advance(remain.scan(whitespace))
	return
}

func (p *Parser) identifyNumber(line buffer) (remain buffer, base int,
	digitFn compare) {
	remain = line
	if remain.startsWith(char('$')) {
		remain = remain.advance(1)
		base = 16
		digitFn = hexDigit
		return
	} else if remain.startsWith(str("0x")) {
		remain = remain.advance(2)
		base = 16
		digitFn = hexDigit
		return
	}
	base = 10
	digitFn = digit
	return
}

func (p *Parser) parseNumber(line buffer) (value int, remain buffer,
	err error) {
	line, base, digitFn := p.identifyNumber(line)

	str, remain := line.takeWhile(digitFn)
	num, err := strconv.ParseInt(str.s, base, 32)
	if err != nil {
		return
	}

	value = int(num)
	return
}

func (p *Parser) parseString(line buffer) (value string, remain buffer,
	err error) {
	remain = line.advance(1)
	valueBuf, remain := remain.takeUntil(char('"'))
	value = valueBuf.s
	if remain.isEmpty() {
		err = fmt.Errorf("Unterminated string in: %s", line)
		return
	}
	remain = remain.advance(1)
	return
}

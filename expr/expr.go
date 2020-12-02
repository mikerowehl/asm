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

type Token struct {
	typ         TokenType
	value       int
	stringValue string
	identifier  string
	op          Op
}

type Parser struct {
	prevTokenType TokenType
}

type opEval func(a int, b int) int

type opEntry struct {
	precedence int
	sym        string
	eval       opEval
}

const (
	opDivide   Op = 0
	opMultiply    = 1
	opAdd         = 2
	opSub         = 3
)

var opTable = []opEntry{
	{2, "/", func(a int, b int) int { return a / b }},
	{2, "*", func(a int, b int) int { return a * b }},
	{1, "+", func(a int, b int) int { return a + b }},
	{1, "-", func(a int, b int) int { return a - b }},
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
			if line.startsWith(str(o.sym)) {
				t.typ = tokenOp
				t.op = Op(i)
				remain = line.advance(len(o.sym))
				break
			}
		}
		if t.typ != tokenOp {
			err = fmt.Errorf("Invalid operation: %s", line)
			return
		}
	}

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

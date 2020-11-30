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

func (p *Parser) parseToken(line buffer) (t Token, remain buffer, err error) {
	if line.isEmpty() {
		t.typ = tokenNil
		return
	}

	switch {
	case line.startsWith(digit) || line.startsWith(char('$')) ||
		line.startsWith(char('-')):
		t.value, remain, err = p.parseNumber(line)
		t.typ = tokenNumber
	case line.startsWith(char('"')):
		t.stringValue, remain, err = p.parseString(line)
		t.typ = tokenString
	}

	return
}

func (p *Parser) identifyNumber(line buffer) (remain buffer, base int,
	digitFn compare, negative bool) {
	remain = line
	if remain.startsWith(char('-')) {
		remain = remain.advance(1)
		negative = true
	}

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
	line, base, digitFn, negative := p.identifyNumber(line)

	str, remain := line.takeWhile(digitFn)
	num, err := strconv.ParseInt(str.s, base, 32)
	if err != nil {
		return
	}

	value = int(num)
	if negative {
		value = -value
	}

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

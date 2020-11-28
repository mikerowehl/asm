package expr

import "strconv"

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
	typ           TokenType
	value         int
	stringLiteral string
	identifier    string
	op            Op
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
	case line.startsWith(digit) || line.startsWith(char('$')):
		t.value, remain, err = p.parseNumber(line)
		t.typ = tokenNumber
	}
	return
}

func (p *Parser) identifyNumber(line buffer) (remain buffer, base int, digitFn compare, negative bool) {
	return line, 10, digit, false
}

func (p *Parser) parseNumber(line buffer) (value int, remain buffer, err error) {
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

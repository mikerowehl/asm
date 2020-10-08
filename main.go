package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType
	val string
}

type itemType int

const (
	itemEOF        itemType = iota
	itemLabel               // line label must end with a colon
	itemColon               // colon for a label
	itemIdentifier          // must not be an opcode
	itemEqual               // used for definitions
	itemOpcode              // machine instruction
	itemPseudo              // must begin with dot, also explicit list
	itemSemicolon           // start of comment
	itemComment             // text of the comment
)

func (t itemType) String() string {
	switch t {
	case itemEOF:
		return "EOF"
	case itemLabel:
		return "LABEL"
	case itemColon:
		return "COLON"
	case itemIdentifier:
		return "IDENTIFIER"
	case itemEqual:
		return "EQUAL"
	case itemOpcode:
		return "OPCODE"
	case itemPseudo:
		return "PSEUDO"
	case itemSemicolon:
		return "SEMICOLON"
	case itemComment:
		return "COMMENT"
	default:
		return fmt.Sprintf("%d", t)
	}
}

type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan item
}

const eof = -1

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexLine; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func lexLine(l *lexer) stateFn {
escape:
	for {
		switch r := l.next(); {
		case r == eof:
			break escape
		case r == ' ':
			l.ignore()
		case r == ';':
			l.emit(itemSemicolon)
			return lexComment
		case isAlphaNumeric(r):
			l.backup()
			return lexInitialIdentifier
		}
	}
	l.emit(itemEOF)
	return nil
}

func lexLabelColon(l *lexer) stateFn {
	l.pos += len(":")
	l.emit(itemColon)
	return lexComment
}

func lexComment(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(itemComment)
			l.emit(itemEOF)
			return nil
		case r == '\n':
			l.backup()
			l.emit(itemComment)
			l.ignore()
			return lexLine
		}
	}
}

func lexInitialIdentifier(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == ':':
			l.backup()
			l.emit(itemLabel)
			return lexLabelColon
		case r == '=':
			l.backup()
			l.emit(itemIdentifier)
			return lexComment
		case !isAlphaNumeric(r):
			l.emit(itemOpcode)
			return lexLine
		}
	}
}

func isAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r)
}

// program = line... eof
// line = definition | statement [;comment] eol
// definition = identifier = operands
// statement = [label:] opcode operands

var testInput = `testlabel: with stuff
someident=value`

func main() {
	lex := &lexer{
		input: testInput,
		items: make(chan item),
	}
	go lex.run()
	for i := range lex.items {
		fmt.Println("Got item from lexer")
		fmt.Println("  type: ", i.typ)
		fmt.Println("  val:  ", i.val)
	}
	fmt.Println("Done lexing")
}

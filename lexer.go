package main

import (
	"log"
	"unicode"
)

/*Token type definitions*/
type TokenType string

const SYMBOL TokenType = "SYMBOL"
const NUMBER TokenType = "NUMBER"
const ATOM TokenType = "ATOM"
const LIST TokenType = "LIST"
const EXP TokenType = "EXP"
const ENV TokenType = "ENV"
const EOF TokenType = "EOF"

const STRING TokenType = "STRING"
const ID TokenType = "ID"
const LPAREN TokenType = "LPAREN"
const RPAREN TokenType = "RPAREN"
const PLUS TokenType = "PLUS"
const INTEGER TokenType = "INTEGER"
const MINUS TokenType = "MINUS"
const MULTIPLY TokenType = "MULTIPLY"
const DIVIDE TokenType = "DIVIDE"
const IF TokenType = "IF"
const TRUE TokenType = "TRUE"
const FALSE TokenType = "FALSE" //nil as false, everything else is treated as true

//some user defined token
const IDEN TokenType = "IDEN"

type Token struct {
	Token   TokenType
	Literal string
}

/**********
Lexer
************/
type Lexer struct {
	Input        string
	Position     int
	ReadPosition int
	Char         byte
}

func New(input string) *Lexer {
	return &Lexer{Input: input, Position: 0, ReadPosition: 0, Char: 0}
}

func (l *Lexer) advance() {
	if l.ReadPosition >= len(l.Input) {
		//Not sure about this bit
		l.Char = 0
	} else {
		l.Char = l.Input[l.ReadPosition]
	}
	l.Position = l.ReadPosition
	l.ReadPosition += 1
}

func (l *Lexer) peek() byte {
	if l.ReadPosition >= len(l.Input) {
		return 0
	}
	return l.Input[l.ReadPosition]
}

func (l *Lexer) skipWhiteSpace() {
	for unicode.IsSpace(rune(l.Char)) {
		l.advance()
	}
}

func (l *Lexer) getInteger() Token {
	old := l.Position
	//peek and not advance since advance is called at the end of scanToken, and this could cause us to jump and skip a step
	for unicode.IsDigit(rune(l.peek())) {
		l.advance()
	}
	return newToken(INTEGER, l.Input[old:l.ReadPosition])
}

func newToken(token TokenType, literal string) Token {
	return Token{Token: token, Literal: literal}
}

//function to get entire string or symbol token
func (l *Lexer) getUntil(until byte, token TokenType) Token {
	old := l.Position
	for l.peek() != until && l.Char != 0 {
		l.advance()
	}
	//if nothing after symbol
	if l.Char == 0 {
		return newToken(token, l.Input[old:l.Position])
	}
	return newToken(token, l.Input[old:l.ReadPosition])
}

func (l *Lexer) scanToken() Token {
	l.skipWhiteSpace()
	var token Token
	switch l.Char {
	case '(':
		token = newToken(LPAREN, "(")
	case ')':
		token = newToken(RPAREN, ")")
	case '`':
		l.advance()
		token = l.getUntil(' ', SYMBOL)
	case '"':
		//skip the first "
		l.advance()
		token = l.getUntil('"', STRING)
	case '+':
		token = newToken(PLUS, "+")
	case '-':
		token = newToken(MINUS, "-")
	case '/':
		token = newToken(DIVIDE, "/")
	case '*':
		token = newToken(MULTIPLY, "*")
	case 0:
		token = newToken(EOF, "EOF")
	default:
		if unicode.IsDigit(rune(l.Char)) {
			token = l.getInteger()
		} else {
			//more things potentially here
			log.Fatal("whoah, you sure that's a valid character mate")
		}
	}
	l.advance()
	return token
}

func (l *Lexer) tokenize(source string) []Token {
	var tokens []Token
	//set the first character
	l.advance()
	for l.Position < len(l.Input) {
		next := l.scanToken()
		tokens = append(tokens, next)
	}
	tokens = append(tokens, Token{Token: EOF, Literal: "EOF"})
	return tokens
}

//Takes as input the source code as a string and returns a list of tokens
func readStr(source string) []Token {
	l := New(source)
	tokens := l.tokenize(source)
	return tokens
}

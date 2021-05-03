package main

import (
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
const EOF TokenTpe = "EOF"

const LPAREN TokenType = "LPAREN"
const RPAREN TokenType = "RPAREN"
const PLUS TokenType = "PLUS"
const INTEGER TokenType = "INTEGER"
const MINUS TokenType = "MINUS"
const MULTIPLY TokenType = "MULTIPLY"
const DIVIDE TokenType = "DIVIDE"
const IF TokenType = "IF"

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
	return &Lexer{Input: input, Position: 0, ReadPosition: 0, ""}
}

func (l *Lexer) advance() {
	l.ReadPosition += 1
	if l.ReadPosition >= len(l.Input) {
		l.Char = nil
	} else {
		l.Char = l.Input[l.ReadPosition]
	}
}

func (l *Lexer) peek() byte {
	return r[l.ReadPosition+1]
}

func (l *Lexer) skipWhiteSpace() {
	for unicode.IsSpace(l.Char) {
		l.advance()
	}
}

func (l *Lexer) getInteger() Token {
	var number string
	for unicode.IsDigit(l.Char) {
		number += string(l.Char)
		l.advance()
	}
	return Token{Token: INTEGER, Literal: number}
}

func (l *Lexer) scanToken() Token {
	for true {
		switch curr := string(l.Char) {
		case " ":
			l.skipWhiteSpace()
			continue
		case "(":
			l.advance()
			return Token{Token: LPAREN, Literal: "("}
		case ")":
			l.advance()
			return Token{Token: RPAREN, Literal: ")"}
		case "+":
			l.advance()
			return Token{Token: PLUS, Literal: "+"}
		case "-":
			l.advance()
			return Token{Token: MINUS, Literal: "-"}
		case "/":
			l.advance()
			return Token{Token: DIVIDE, Literal: "/"}
		case "*":
			l.advance()
			return Token{Token: MULTIPLY, Literal: "*"}
		case unicode.IsDigit([]rune(curr)):
			return l.getInteger()
		}
	}
}

//Takes as input the source code as a string and returns a list of tokens
func ReadStr(source string) {
	l := New(source)
	var currentToken = l.getNextToken()
	for currentToken.Token != EOF {

	}
}





/*
Grammar

expr:   ID | STR | NUM | list
list:   ( seq )  
seq:       | expr seq
(note ^ is an empty list)


*/
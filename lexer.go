package main

import (
	"fmt"
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
	return &Lexer{Input: input, Position: -1, ReadPosition: -1, Char: 0}
}

func (l *Lexer) advance() {
	l.ReadPosition += 1
	if l.ReadPosition >= len(l.Input) {
		//Not sure about this bit
		l.Char = 0
	} else {
		l.Char = l.Input[l.ReadPosition]
		l.Position = l.ReadPosition - 1
	}
}

func (l *Lexer) peek() byte {
	return l.Input[l.ReadPosition+1]
}

func (l *Lexer) skipWhiteSpace() {
	for unicode.IsSpace(rune(l.Char)) {
		l.advance()
	}
}

func (l *Lexer) getInteger() Token {
	var number string
	for unicode.IsDigit(rune(l.Char)) {
		number += string(l.Char)
		l.advance()
	}
	return Token{Token: INTEGER, Literal: number}
}

func (l *Lexer) scanToken() Token {
	for true {
		char := string(l.Char)
		switch char {
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
		default:
			if unicode.IsDigit(rune(l.Char)) {
				return l.getInteger()
			}
			//more things potentially here
			fmt.Println("whoah, you sure that's a valid character mate")
			return Token{}
		}
	}
	return Token{}
}

func (l *Lexer) tokenize(source string) []Token {
	var tokens []Token
	l.advance()
	for l.Position < len(l.Input) {
		next := l.scanToken()
		tokens = append(tokens, next)
	}
	tokens = append(tokens, Token{Token: EOF, Literal: "EOF"})
	return tokens
}

//Takes as input the source code as a string and returns a list of tokens
func readStr(source string) {
	l := New(source)
	tokens := l.tokenize(source)
	//parse the tokens
	for _, Token := range tokens {
		fmt.Println(Token)
	}

}

/*
Grammar

expr:   ID | STR | NUM | list
list:   ( seq )
seq:       | expr seq
(note ^ is an empty list)


*/

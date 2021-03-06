package lispy

import (
	"io"
	"io/ioutil"
	"log"
	"unicode"
)

/*Token type definitions*/
type TokenType string

const SYMBOL TokenType = "SYMBOL"

//not used
// const ATOM TokenType = "ATOM"
// const EXP TokenType = "EXP"
// const ENV TokenType = "ENV"
const EOF TokenType = "EOF"

//eventually refactor to hashmap?

const LPAREN TokenType = "LPAREN"
const RPAREN TokenType = "RPAREN"
const LSQUARE TokenType = "LSQUARE"
const RSQUARE TokenType = "RSQUARE"

const INTEGER TokenType = "INTEGER"
const FLOAT TokenType = "FLOAT"

//Symbols
const STRING TokenType = "STRING"
const COMMENT TokenType = "COMMENT"

const ID TokenType = "ID"
const IF TokenType = "IF"
const DEFINE TokenType = "DEFINE"
const TRUE TokenType = "TRUE"
const FALSE TokenType = "FALSE"
const QUOTE TokenType = "QUOTE"
const UNQUOTE TokenType = "UNQUOTE"
const DO TokenType = "DO"
const ARRAY TokenType = "ARRAY"
const MACRO TokenType = "MACRO"

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
	for unicode.IsSpace(rune(l.Char)) || l.Char == '\n' {
		l.advance()
	}

}

func (l *Lexer) getFloat(start int) Token {
	//advance to skip .
	l.advance()
	for unicode.IsDigit(rune(l.peek())) {
		l.advance()
	}
	return newToken(FLOAT, l.Input[start:l.ReadPosition])
}

func (l *Lexer) getInteger() Token {
	old := l.Position
	if l.Char == '-' {
		//skip first char if minus
		l.advance()
	}
	//peek and not advance since advance is called at the end of scanToken, and this could cause us to jump and skip a step
	for unicode.IsDigit(rune(l.peek())) {
		l.advance()
	}
	if l.peek() == '.' {
		return l.getFloat(old)
	}
	return newToken(INTEGER, l.Input[old:l.ReadPosition])
}

func (l *Lexer) getSymbol() Token {
	old := l.Position
	for !unicode.IsSpace(rune(l.peek())) && l.peek() != 0 && l.peek() != ')' && l.peek() != ']' && l.peek() != '(' {
		l.advance()
	}
	//use position because when l.Char is at a space, l.ReadPosition will be one ahead
	val := l.Input[old:l.ReadPosition]
	var token Token
	switch val {
	case "define":
		token = newToken(DEFINE, "define")
	case "if":
		token = newToken(IF, "if")
	case "true":
		token = newToken(TRUE, "true")
	case "false", "nil":
		token = newToken(FALSE, "false")
	case "do":
		token = newToken(DO, "do")
	case "macro":
		token = newToken(MACRO, "macro")
	//will add others later
	default:
		token = newToken(SYMBOL, val)
	}
	return token
}

func newToken(token TokenType, literal string) Token {
	return Token{Token: token, Literal: literal}
}

//function to get entire string or symbol token
func (l *Lexer) getUntil(until byte, token TokenType, after bool) Token {
	old := l.Position
	//get until assumes we eat the last token, which is why we don't use peek
	for l.Char != until && l.Char != 0 {
		l.advance()
	}
	if after && l.Char != 0 {
		l.advance()
	}
	return newToken(token, l.Input[old:l.Position])
}

func (l *Lexer) scanToken() Token {
	//skips white space and new lines
	l.skipWhiteSpace()
	var token Token
	switch l.Char {
	case '(':
		token = newToken(LPAREN, "(")
	case ')':
		token = newToken(RPAREN, ")")
	case '[':
		token = newToken(LSQUARE, "[")
	case ']':
		token = newToken(RSQUARE, "]")
	case '\'':
		token = newToken(QUOTE, "'")
	case '-':
		if unicode.IsDigit(rune(l.peek())) {
			token = l.getInteger()
		} else {
			token = l.getSymbol()
		}

	case ';':
		if l.peek() == ';' {
			//current char is ; and next char is ; so advance twice
			l.advance()
			l.advance()
			token = l.getUntil(';', COMMENT, false)
		} else {
			token = l.getUntil('\n', COMMENT, false)
		}

	case '"':
		//skip the first "
		l.advance()
		token = l.getUntil('"', STRING, false)
	case 0:
		token = newToken(EOF, "EOF")
	default:
		if unicode.IsDigit(rune(l.Char)) {
			token = l.getInteger()
		} else {
			//more things potentially here
			token = l.getSymbol()
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
		if next.Token != COMMENT {
			tokens = append(tokens, next)
		}
	}
	return tokens
}

//Takes as input the source code as a string and returns a list of tokens
func Read(reader io.Reader) []Token {
	source := loadReader(reader)
	l := New(source)
	tokens := l.tokenize(source)
	return tokens
}

func loadReader(reader io.Reader) string {
	//todo: ReadAll puts everything in memory, very inefficient for large files
	//files will remain small for lispy but potentially adapt to buffered approach (reads in buffers)
	ltxtb, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal("Error trying to read source file: ", err)
	}
	return string(ltxtb)
}

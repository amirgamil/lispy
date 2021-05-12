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
const PLUS TokenType = "PLUS"
const MINUS TokenType = "MINUS"
const STRING TokenType = "STRING"
const MULTIPLY TokenType = "MULTIPLY"
const DIVIDE TokenType = "DIVIDE"
const EQUAL TokenType = "EQUAL"
const GEQUAL TokenType = "GEQUAL"
const LEQUAL TokenType = "LEQUAL"
const GTHAN TokenType = "GTHAN"
const LTHAN TokenType = "LTHAN"
const COMMENT TokenType = "COMMENT"

const ID TokenType = "ID"
const IF TokenType = "IF"
const DEFINE TokenType = "DEFINE"
const PRINT TokenType = "PRINT"
const TRUE TokenType = "TRUE"
const FALSE TokenType = "FALSE"
const AND TokenType = "AND"
const OR TokenType = "OR"
const NOT TokenType = "NOT"
const QUOTE TokenType = "QUOTE"
const DO TokenType = "DO"
const LIST TokenType = "LIST"
const ARRAY TokenType = "ARRAY"

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
	case "println":
		token = newToken(PRINT, "println")
	case "true":
		token = newToken(TRUE, "true")
	case "false", "nil":
		token = newToken(FALSE, "false")
	case "and":
		token = newToken(AND, "and")
	case "or":
		token = newToken(OR, "or")
	case "not":
		token = newToken(NOT, "not")
	case "do":
		token = newToken(DO, "do")

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
	if after {
		l.advance()
	}
	return newToken(token, l.Input[old:l.Position])
}

// func (l *Lexer) getQuote() Token {
// 	old := l.Position
// 	left := 0
// 	end := false
// 	if l.Char == '(' {
// 		left = 1
// 	}
// 	for left >= 0 && !end {
// 		l.advance()
// 		if l.Char == '(' {
// 			left += 1
// 		} else if l.Char == ')' || (left == 0 && l.Char == ' ') {
// 			left -= 1
// 			if left == 0 {
// 				//case where initially (, so set end to prevent infinite loop
// 				end = true
// 				//skip )
// 				l.advance()
// 			}
// 		}
// 	}
// 	fmt.Println("here => ", l.Input[old:l.Position])
// 	return newToken(QUOTE, l.Input[old:l.Position])
// }

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
	case ';':
		token = newToken(FALSE, "nil")
		l.getUntil('\n', COMMENT, true)
	case '"':
		//skip the first "
		l.advance()
		token = l.getUntil('"', STRING, false)
		//skip final quote
	case '+':
		token = newToken(PLUS, "+")
	case '.':
		token = l.getFloat(l.Position)
	case '-':
		token = newToken(MINUS, "-")
	case '/':
		token = newToken(DIVIDE, "/")
	case '*':
		token = newToken(MULTIPLY, "*")
	case '=':
		token = newToken(EQUAL, "=")
	case '>':
		if l.peek() == '=' {
			token = newToken(GTHAN, ">=")
			//skip equal
			l.advance()
		} else {
			token = newToken(GEQUAL, ">")
		}
	case '<':
		if l.peek() == '=' {
			token = newToken(LTHAN, "<=")
			//skip equal
			l.advance()
		} else {
			token = newToken(LEQUAL, "<")
		}
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
		tokens = append(tokens, next)
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

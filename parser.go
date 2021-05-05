package main

import (
	"log"
	"strconv"
)

// const (
// 	Num = iota
// 	List
// 	Identifier
// 	String
// )

//Generic interface for an Sexp (any node in our AST must implement this interface)
type Sexp interface {
	String() string
}

//Symbol
type SexpSymbol struct {
	//TODO: clean this up? don't know how yet
	ofType TokenType
	value  string
}

func (s SexpSymbol) String() string {
	return s.value
}

//Number
type Number int

func (n Number) String() string {
	return strconv.Itoa(int(n))
}

//List node in our AST
//Implement a list trivially as this for now
//change to linkedlist late (since in Lisp an Sexp is defined inductively)
type List []Sexp

func (l List) String() string {
	if len(l) == 0 {
		return "[]"
	}

	strBuilder := "["
	for i := 0; i < len(l); i++ {
		strBuilder += " " + l[i].String()
	}
	strBuilder += "]"
	return strBuilder
}

func parse(tokens []Token) ([]Sexp, error) {
	idx, length := 0, len(tokens)
	nodes := make([]Sexp, 0)

	for idx < length && tokens[idx].Token != EOF {
		expr, add, err := parseExpr(tokens[idx:])
		if err != nil {
			log.Fatal("Error parsing tokens: ", err)
		}
		idx += add
		nodes = append(nodes, expr)
	}
	return nodes, nil
}

//Implement a list trivially as this for now
func parseList(tokens []Token) (Sexp, int, error) {
	idx, length := 0, len(tokens)
	arr := make([]Sexp, 0)
	for idx < length && tokens[idx].Token != RPAREN {
		currExpr, add, err := parseExpr(tokens[idx:])
		if err != nil {
			return nil, 0, err
		}
		idx += add
		arr = append(arr, currExpr)
	}
	return List(arr), idx + 1, nil

}

func parseExpr(tokens []Token) (Sexp, int, error) {
	idx := 0
	var expr Sexp
	var err error
	var add int

	switch tokens[idx].Token {
	case LPAREN:
		idx++
		expr, add, err = parseList(tokens[idx:])
		if err != nil {
			return nil, 0, err
		}
		idx += add
	case INTEGER:
		i, err := strconv.Atoi(tokens[idx].Literal)
		if err != nil {
			return nil, 0, err
		}
		idx++
		expr = Number(i)
	//eventually refactor to handle other symbols like identifiers
	//create a map with all of these operators pre-stored and just get, or default, passing in tokentype to check if it exists
	case PLUS, MULTIPLY, DIVIDE, MINUS:
		expr = SexpSymbol{ofType: tokens[idx].Token, value: tokens[idx].Literal}
		idx++
	default:
		log.Fatal("you screwed it up my dude")
	}
	return expr, idx, nil
}

/*
Grammar

number : /-?[0-9]+/ ;                    \
symbol : '+' | '-' | '*' | '/' ;         \
list  : ( <expr>* ) ;               \
expr   : <number> | <symbol> | <list> ; \
	- symbol = operator, variable, or function
lispy  : /^/ <expr>* /$/ ;               \
	- /^/ means start of the input is required (n/a atm, don't have a start token)
	- /$/ means end of the input is required (EOF tag)
-----------------------------------
expr:  ID | NUM | list
	ID = identifier like variable or binding
	expr: ID | STR | NUM | list (but we'll ignore this for now to simplify things)
list:   "(" seq ")"
	"" here just mean that we see a token i.e. LPAREN seq RPAREN
seq:       | expr seq
(note ^ is an empty list)



ATOMS:
- Strings
- Symbols
- Numbers
*/

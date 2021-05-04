package main

import (
	"fmt"
	"log"
)

const (
	Num = iota
	List
	Identifier
	String
)

type Atom struct {
	value interface{}
}

type List struct {
	list []interface{}
}

type Expression struct {
	operator  TokenType
	arguments List
}

func parse(tokens []Token) ([]Expression, error) {
	idx, length := 0, len(tokens)
	nodes := make([]Expression, 0)
	for idx < length && tokens[idx].Token != EOF {
		token := tokens[idx]
		switch token.Token {
		case LPAREN:
			//skip paren
			idx++
			node, add, err := parseList(tokens[idx:])
			if err != nil {
				log.Fatal("error parsing the list")
			}
			fmt.Println(node)
			nodes = append(nodes, node)
			idx += add

		default:
			node, add, err := parseExpression(tokens[idx:])
			if err != nil {
				log.Fatal("error parsing the list")
			}
			nodes = append(nodes, node)
			idx += add
		}
	}
	return nodes, nil
}

func parseList(tokens []Token) (Expression, int, error) {
	idx, length := 0, len(tokens)
	expr := Expression{}
	for idx < length && tokens[idx].Token != RPAREN {
		switch tokens[idx].Token {
		case LPAREN:
			//nest call again
			idx++
			nestExpr, add, err := parseList(tokens[idx:])
			if err != nil {
				log.Fatal("Error parsing a nested list: ", err)
			}
			idx += add
			expr.arguments = append(expr.arguments, nestExpr)
		case PLUS, MINUS, DIVIDE, MULTIPLY:
			expr.operator = tokens[idx].Token
		default:
			//This will be a call to expr
			expr.arguments = append(expr.arguments, tokens[idx])
		}
		idx++
	}
	return expr, idx + 1, nil

}

func parseSeq(tokens []Token) (Expression, int, error) {
	idx, length := 0, len(tokens)
	expr := Expression{}

}

func parseExpr(token Token) (interface{}, idx, error) {
	expr := Expression{}
	switch token.Token {
	case LPAREN:
		//something
	case NUMBER:
		idx++
		return token, idx, nil
	case ID:
		//do some stuff
		fmt.Println("deal with this")
	}
	return expr, idx, nil
}

/*
Grammar

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

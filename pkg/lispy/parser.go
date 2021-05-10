package lispy

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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
//have dedicated types for strings and bools or this suffices?
type SexpSymbol struct {
	ofType TokenType
	value  string
}

func (s SexpSymbol) String() string {
	return s.value
}

// SexpInt
type SexpInt int

func (n SexpInt) String() string {
	return strconv.Itoa(int(n))
}

//SexpFloat
type SexpFloat float64

func (n SexpFloat) String() string {
	return fmt.Sprintf("%f", n)
}

//List node in our AST
//Implement a list trivially as this for now
//change to linkedlist later (since in Lisp an Sexp is defined inductively)
type SexpList struct {
	ofType TokenType
	value  []Sexp
}

func (l SexpList) String() string {
	if len(l.value) == 0 {
		return "[]"
	}

	strBuilder := make([]string, 0)
	strBuilder = append(strBuilder, "[")
	for i := 0; i < len(l.value); i++ {
		strBuilder = append(strBuilder, l.value[i].String())
	}
	strBuilder = append(strBuilder, "]")
	return strings.Join(strBuilder, ",")
}

//SexpFunction Literal to store the functions when parsing them
type SexpFunctionLiteral struct {
	name string
	//when we store the arguments, will call arg.String() for each arg - may need to be fixed for some edge cases
	arguments SexpList
	body      Sexp
}

func (f SexpFunctionLiteral) String() string {
	args := make([]string, 0)
	for _, node := range f.arguments.value {
		args = append(args, node.String())
	}
	return fmt.Sprintf("Define (%s) on (%s)",
		f.name,
		strings.Join(args, ", "))
}

type SexpFunctionCall struct {
	//for now keep arguments as string, in future potentially refacto wrap in SexpIdentifierNode
	name      string
	arguments SexpList
}

func (f SexpFunctionCall) String() string {
	args := make([]string, 0)
	for _, node := range f.arguments.value {
		args = append(args, node.String())
	}
	return fmt.Sprintf("Function call (%s) on (%s)",
		f.name,
		strings.Join(args, ", "))
}

/********** PARSING CODE ****************/
func Parse(tokens []Token) ([]Sexp, error) {
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
	return SexpList{ofType: LIST, value: arr}, idx + 1, nil
}

//parses array e.g. in arguments of a function
func parseArray(tokens []Token) (SexpList, int, error) {
	idx, length := 0, len(tokens)
	arr := make([]Sexp, 0)

	for idx < length && tokens[idx].Token != RSQUARE {
		expr, add, err := parseExpr(tokens[idx:])
		if err != nil {
			return SexpList{}, 0, err
		}
		idx += add
		arr = append(arr, expr)
	}
	return SexpList{ofType: ARRAY, value: arr}, idx + 1, nil
}

//parses a function literal
func parseFunctionLiteral(tokens []Token) (Sexp, int, error) {
	idx := 0
	if tokens[0].Token != SYMBOL {
		log.Fatal("Unexpected syntax trying to define a function")
	}
	//function name will be at index 0
	name := tokens[0].Literal
	//skip the [ and go to the next character
	idx += 2
	//parse arguments first
	args, add, err := parseArray(tokens[idx:])
	if err != nil {
		return nil, 0, err
	}
	idx += add
	//parse body of the function which which will be an Sexpr
	body, addBlock, err := parseExpr(tokens[idx:])
	if err != nil {
		return nil, 0, err
	}
	idx += addBlock
	//entire function include define was enclosed in (), note DON'T SKIP 1 otherwise may read code outside function
	return SexpFunctionLiteral{name: name, arguments: args, body: body}, idx, nil
}

//parses a function call
func parseFunctionCall(tokens []Token) (Sexp, int, error) {
	idx := 0
	name := tokens[0].Literal
	idx += 1
	args, add, err := parseList(tokens[idx:])
	//is there a better way than to type assert?
	origArgs, isArgs := args.(SexpList)
	if !isArgs {
		log.Fatal("Error parsing function parameters")
	}
	if err != nil {
		return nil, 0, err
	}
	idx += add
	return SexpFunctionCall{name: name, arguments: origArgs}, idx, nil
}

//parses a single expression (list or non-list)
func parseExpr(tokens []Token) (Sexp, int, error) {
	idx := 0
	var expr Sexp
	var err error
	var add int

	switch tokens[idx].Token {
	case DEFINE:
		//look ahead one to check if it's a function or just data-binding
		if idx+2 < len(tokens) && tokens[idx+2].Token == LSQUARE {
			//skip define token
			idx++
			expr, add, err = parseFunctionLiteral(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}
			idx += add
		} else {
			expr = SexpSymbol{ofType: tokens[idx].Token, value: tokens[idx].Literal}
			//POSSIBLE FEATURE AMMENDMENT: If I add local binding via let similar to Clojure, will be added here
			idx++
		}
	case LPAREN:
		idx++
		//check if this is a function call i.e. the next token is a symbol
		if idx < len(tokens) && tokens[idx].Token == SYMBOL && tokens[idx].Literal != "define" {
			expr, add, err = parseFunctionCall(tokens[idx:])
		} else {
			expr, add, err = parseList(tokens[idx:])
		}
		idx += add
		if err != nil {
			return nil, 0, err
		}
	// case QUOTE:
	// 	idx++
	// 	if idx < len(tokens) && tokens[idx].Token == LPAREN {
	// 		expr, add, err = parseList
	// 	} else {
	// 		expr = SexpSymbol{ofType: QUOTE, value: tokens[idx].Literal}
	// 		idx++
	// 	}
	case INTEGER:
		i, err := strconv.Atoi(tokens[idx].Literal)
		if err != nil {
			return nil, 0, err
		}
		idx++
		expr = SexpInt(i)
	case FLOAT:
		i, err := strconv.ParseFloat(tokens[idx].Literal, 64)
		if err != nil {
			return nil, 0, err
		}
		expr = SexpFloat(i)
		idx++
	//eventually refactor to handle other symbols like identifiers
	//create a map with all of these operators pre-stored and just get, or default, passing in tokentype to check if it exists
	case PLUS, MULTIPLY, DIVIDE, MINUS, STRING, TRUE, FALSE, GEQUAL, LEQUAL, GTHAN, LTHAN, EQUAL, QUOTE, AND, OR, NOT, IF, PRINT, DO, SYMBOL:
		expr = SexpSymbol{ofType: tokens[idx].Token, value: tokens[idx].Literal}
		idx++
	default:
		log.Fatal("error parsing")
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

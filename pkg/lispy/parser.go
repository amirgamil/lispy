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

//SexpPair is an implementation of a cons cell with a head (car) and tail (cdr)
//Lists in lispy are defined as linked lists of cons cells
type SexpPair struct {
	head Sexp
	tail Sexp
}

func (l SexpPair) String() string {
	if l.head == nil {
		return "()"
	}

	strBuilder := make([]string, 0)
	pair := l
	count := 0
	if l.tail == nil {
		count += 1
		//not great way, is there a better way?
		switch pair.head.(type) {
		case SexpPair:
			//hacky way of not doing anything
			count = count
		default:
			//subtract 1 if this is just block data not in a list
			count -= 1
		}
	}
	for {
		strBuilder = append(strBuilder, pair.head.String())
		switch pair.tail.(type) {
		case SexpPair:
			pair = pair.tail.(SexpPair)
			continue
		}
		break
	}
	return strings.Repeat("(", count) + strings.Join(strBuilder, " ") + strings.Repeat(")", count)
}

type SexpArray struct {
	ofType TokenType
	value  []Sexp
}

func (s SexpArray) String() string {
	args := make([]string, 0)
	for _, node := range s.value {
		args = append(args, node.String())
	}
	return "[" + strings.Join(args, " ") + "]"
}

//SexpFunction Literal to store the functions when parsing them
type SexpFunctionLiteral struct {
	name string
	//when we store the arguments, will call arg.String() for each arg - may need to be fixed for some edge cases
	arguments SexpArray
	body      Sexp
	//userfunc represents a native built-in implementation (which can be overrided e.g. with macros through the body argument)
	userfunc LispyUserFunction
}

func (f SexpFunctionLiteral) String() string {
	if f.userfunc == nil {
		return fmt.Sprintf("Define (%s) on (%s)",
			f.name,
			f.arguments.String())
	} else {
		return "Built-in native implementation function"
	}
}

type SexpFunctionCall struct {
	//for now keep arguments as string, in future potentially refacto wrap in SexpIdentifierNode
	name      string
	arguments SexpPair
}

func (f SexpFunctionCall) String() string {
	return fmt.Sprintf("Function call (%s) on (%s)",
		f.name,
		f.arguments.String())
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
	idx := 0
	curr := SexpPair{head: nil, tail: nil}
	if tokens[idx].Token == RPAREN {
		//return idx of 1 so we skip the RPAREN
		return nil, 1, nil
	}
	currExpr, add, err := parseExpr(tokens[idx:])
	if err != nil {
		return nil, 0, err
	}
	idx += add
	curr.head = currExpr
	//recursively build out list of cons cells
	tailExpr, addTail, err := parseList(tokens[idx:])
	if err != nil {
		return nil, 0, err
	}
	idx += addTail
	curr.tail = tailExpr
	return curr, idx, nil
}

//parses array e.g. in arguments of a function
func parseArray(tokens []Token) (SexpArray, int, error) {
	idx, length := 0, len(tokens)
	arr := make([]Sexp, 0)

	for idx < length && tokens[idx].Token != RSQUARE {
		expr, add, err := parseExpr(tokens[idx:])
		if err != nil {
			return SexpArray{}, 0, err
		}
		idx += add
		arr = append(arr, expr)
	}
	return SexpArray{ofType: ARRAY, value: arr}, idx + 1, nil
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
	return SexpFunctionLiteral{name: name, arguments: args, body: body, userfunc: nil}, idx, nil
}

//parses a function call
func parseFunctionCall(tokens []Token) (Sexp, int, error) {
	idx := 0
	name := tokens[0].Literal
	idx += 1
	args, add, err := parseList(tokens[idx:])
	//is there a better way than to type assert?
	origArgs, isArgs := args.(SexpPair)
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
		} else if tokens[idx].Token == RPAREN {
			//check for empty list
			expr = SexpPair{head: nil, tail: nil}
			add = 1
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
	case QUOTE:
		idx++
		nextExpr, toAdd, errorL := parseExpr(tokens[idx:])
		if errorL != nil {
			log.Fatal("Error parsing quote!")
		}
		expr = MakeList([]Sexp{SexpSymbol{ofType: QUOTE, value: "'"}, nextExpr})
		idx += toAdd
	//eventually refactor to handle other symbols like identifiers
	//create a map with all of these operators pre-stored and just get, or default, passing in tokentype to check if it exists
	case PLUS, MULTIPLY, DIVIDE, MINUS, STRING, TRUE, FALSE, GEQUAL, LEQUAL, GTHAN, LTHAN, EQUAL, AND, OR, NOT, IF, PRINT, DO, SYMBOL:
		expr = SexpSymbol{ofType: tokens[idx].Token, value: tokens[idx].Literal}
		idx++
	default:
		log.Fatal("error parsing")
	}
	return expr, idx, nil
}

func MakeList(expressions []Sexp) Sexp {
	if len(expressions) == 0 {
		return nil
	}

	return consHelper(expressions[0], MakeList(expressions[1:]))
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

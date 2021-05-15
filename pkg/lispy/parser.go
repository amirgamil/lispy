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
	str := "("
	pair := l
	for {
		switch pair.tail.(type) {
		case SexpPair:
			str += pair.head.String() + " "
			pair = pair.tail.(SexpPair)
			continue
		}
		break
	}
	//add last item
	str += pair.head.String()

	if pair.tail == nil {
		str += ")"
	} else {
		str += pair.tail.String() + ")"
	}
	return str
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
	//used for annonymous function calls
	body Sexp
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

func getName(token Token) string {
	if token.Token != SYMBOL {
		log.Fatal("Unexpected syntax trying to define a function")
	}
	//function name will be at index 0
	name := token.Literal
	return name
}

//parses a function literal
func parseFunctionLiteral(tokens []Token, name string) (Sexp, int, error) {
	idx := 0
	var args SexpArray
	var add int
	var err error
	if tokens[idx].Token == LSQUARE {
		//skip the [ and go to the next character
		idx += 1
		//parse arguments first
		args, add, err = parseArray(tokens[idx:])
		if err != nil {
			return nil, 0, err
		}
		idx += add
	} else {
		//means we have a lambda expression here
		args = SexpArray{}
	}

	//parse body of the function which which will be an Sexpr
	body, addBlock, err := parseExpr(tokens[idx:])
	if err != nil {
		return nil, 0, err
	}
	idx += addBlock
	//entire function include define was enclosed in (), note DON'T SKIP 1 otherwise may read code outside function
	return SexpFunctionLiteral{name: name, arguments: args, body: body, userfunc: nil}, idx + 1, nil
}

//parses a function call
func parseFunctionCall(tokens []Token, body Sexp) (Sexp, int, error) {
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
	return SexpFunctionCall{name: name, arguments: origArgs, body: nil}, idx, nil
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
		if idx+2 < len(tokens) && (tokens[idx+2].Token == LSQUARE || tokens[idx+2].Token == LPAREN) {
			//skip define token
			idx++
			name := getName(tokens[idx])
			expr, add, err = parseFunctionLiteral(tokens[idx+1:], name)
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
		//check if anonymous function
		if idx+1 < len(tokens) && tokens[idx].Token == FN && tokens[idx+1].Token == LSQUARE {
			//skip fn
			idx++
			//give anonymous functions the same name because by definition, should not be able to refer
			//to them after they have been defined (designed to execute there and then)
			expr, add, err = parseFunctionLiteral(tokens[idx:], "fn")
		} else if idx < len(tokens) && tokens[idx].Token == SYMBOL && tokens[idx].Token != DEFINE {
			//check if this is a function call i.e. the next token is a symbol
			expr, add, err = parseFunctionCall(tokens[idx:], nil)
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
	case STRING, TRUE, FALSE, IF, DO, SYMBOL:
		expr = SexpSymbol{ofType: tokens[idx].Token, value: tokens[idx].Literal}
		idx++
	default:
		fmt.Println(tokens[idx])
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

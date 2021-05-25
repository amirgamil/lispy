package lispy

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type Env struct {
	//pointer to the environment with globals
	global *Env
	store  map[string]Value
}

//Value is a reference to any Value in a Lispy program
type Value interface {
	String() string
}

//Value referencing any functions
type FunctionValue struct {
	defn *SexpFunctionLiteral
}

//allow functionvalue to implement value
func (env Env) String() string {
	data := make([]string, 0)
	for key, val := range env.store {
		data = append(data, key+":"+val.String())
	}
	return strings.Join(data, " ")
}

func (f FunctionValue) String() string {
	//TODO: clean this up later
	return f.defn.String()
}

func returnDefinedFunctions() map[string]LispyUserFunction {
	functions := make(map[string]LispyUserFunction)
	functions["car"] = car
	functions["cdr"] = cdr
	functions["cons"] = cons
	functions["+"] = add
	functions["-"] = minus
	functions["/"] = divide
	functions["*"] = multiply
	functions["#"] = expo
	functions["%"] = modulo
	functions["="] = equal
	functions[">="] = gequal
	functions["<="] = lequal
	functions[">"] = gthan
	functions["<"] = lthan
	functions["and"] = and
	functions["or"] = or
	functions["not"] = not
	functions["println"] = printlnStatement
	functions["list"] = createList
	functions["type"] = typeOf
	functions["quote"] = quote
	return functions
}

func InitState() *Env {
	//add more ops as need for function bodies, assignments etc
	env := new(Env)
	env.store = make(map[string]Value)
	for key, function := range returnDefinedFunctions() {
		env.store[key] = makeUserFunction(key, function)
	}
	//this is the global reference, so set the pointer to nil
	env.global = nil
	return env
}

func (env *Env) evalSymbol(s SexpSymbol, args []Sexp) Sexp {
	switch s.ofType {
	case TRUE, FALSE, STRING:
		return s
	case IF:
		return conditionalStatement(env, s.value, args)
	case DEFINE:
		return varDefinition(env, args[0].String(), args[1:])
	case QUOTE:
		return s
	case SYMBOL:
		//if no argument then it's a variable
		if len(args) == 0 {
			return getVarBinding(env, s.value, args)
		}
		//otherwise assume this is a function call - this is MACROEXPANSION CODE!!
		argList, isList := args[0].(SexpPair)
		if !isList {
			log.Fatal("Error trying to parse arguments for function call")
		}
		//check if this is an anonymous function the macro called
		if s.value == "fn" {
			params, isArray := argList.head.(SexpArray)
			if !isArray {
				log.Fatal("Error parsing anonymous function in macro expansion!")
			}
			newParams := params
			anonFunc := SexpFunctionLiteral{name: "fn", arguments: newParams, body: argList.tail, userfunc: nil, macro: false}
			return anonFunc
		}
		funcCall := SexpFunctionCall{name: s.value, arguments: argList}
		return env.evalFunctionCall(&funcCall)
	default:
		fmt.Println(s.ofType, " ", s.value, " args: ", args)
		log.Fatal("Uh oh, weird symbol my dude")
		return nil
	}
}

func (env *Env) evalFunctionLiteral(s *SexpFunctionLiteral) Sexp {
	//if it's an anonymous function, just return it the way a function would be stored in the environment
	if (s.name) == "fn" {
		return FunctionValue{defn: s}
	}
	return funcDefinition(env, s)
}

func (env *Env) evalFunctionCall(s *SexpFunctionCall) Sexp {
	//each call should get its own environment for recursion to work
	functionCallEnv := new(Env)
	functionCallEnv.store = make(map[string]Value)
	//copy globals
	for key, element := range env.store {
		functionCallEnv.store[key] = element
	}
	return getFuncBinding(functionCallEnv, s)
}

func (env *Env) evalList(n SexpPair) Sexp {
	var toReturn Sexp
	//empty string
	if n.head == nil {
		return SexpPair{}
	}
	tail, isTail := n.tail.(SexpPair)
	switch head := n.head.(type) {
	case SexpSymbol:
		symbol, ok := n.head.(SexpSymbol)
		if !ok {
			log.Fatal("error trying to interpret symbol")
		}
		arguments := make([]Sexp, 0)
		//process all arguments here for ease?
		switch symbol.ofType {
		case DEFINE:
			if !isTail {
				log.Fatal("Unexpected definition, missing value!")
			}
			//binding to a variable
			toReturn = env.evalSymbol(symbol, makeList(tail))
		case QUOTE:
			if !isTail {
				log.Fatal("Error trying to interpret quote")
			}
			//don't evaluate the expression
			toReturn = tail.head
		case IF:
			if !isTail {
				fmt.Println("Error interpreting condition for the if statement")
			}
			arguments = append(arguments, env.evalNode(tail.head))
			statements, isValid := tail.tail.(SexpPair)
			if !isValid {
				log.Fatal("Error please provide valid responses to the if condition!")
			}
			res := makeList(statements)
			arguments = append(arguments, res...)
			toReturn = env.evalSymbol(symbol, arguments)
		case DO:
			//if symbol is do, we just evaluate the nodes and return the (result of the) last node
			//note do's second element will be a list of lists so we need to unwrap it
			if !isTail {
				log.Fatal("Error trying to interpret do statements")
			}
			for {
				toReturn = env.evalNode(tail.head)
				switch tail.tail.(type) {
				case SexpPair:
					tail = tail.tail.(SexpPair)
					continue
				}
				break
			}
		default:
			// fmt.Println("default symbol in list -> ", n.head)
			toReturn = env.evalSymbol(symbol, []Sexp{tail})
		}
	case SexpFunctionLiteral:
		//anonymous function, so handle differently
		if head.name == "fn" {
			//save body of function to the env then call
			funcDefinition(env, &head)
			//check tail != nil for anon function with no parameters
			if !isTail && n.tail != nil {
				log.Fatal("Error interpreting anonymous function parameters")
			}
			funcCall := SexpFunctionCall{name: "fn", arguments: tail, body: nil}
			toReturn = env.evalFunctionCall(&funcCall)
		} else {
			toReturn = env.evalNode(n.head)
			//in a function literal, body should only be on Sexp, if there is more, throw an error
			//in a function call, arguments will be pased into SexpFunctionCall so similar idea
			if n.tail != nil {
				log.Fatal("Error interpreting function declaration or literal - ensure only one Sexp in body of function literal!")
			}
		}
	case SexpFunctionCall:
		toReturn = env.evalNode(n.head)
	case SexpPair:
		original, ok := n.head.(SexpPair)
		if ok {
			toReturn = env.evalList(original)
			//if this is an anon function from a macro, need to set it up as such
			funcLiteral, isFuncLiteral := toReturn.(SexpFunctionLiteral)
			if isFuncLiteral && funcLiteral.name == "fn" {
				//this is a function call so we can use the code above under case SexpFunctionLiteral
				//by artificially constructing a list as such
				toReturn = env.evalList(SexpPair{head: funcLiteral, tail: n.tail})
			}
		} else {
			//TODO: might need to be fixed
			toReturn = SexpSymbol{FALSE, "false"}
		}
	//if it's just a list without a symbol at the front, treat it as data and return it
	default:
		toReturn = n
	}
	return toReturn
}

//wrapper for evaluating an individual Sexp node in our AST
func (env *Env) evalNode(node Sexp) Sexp {
	var toReturn Sexp
	switch i := node.(type) {
	case SexpPair:
		//Assert type since ast is composed of generic Sexp interface
		original, ok := node.(SexpPair)
		if ok {
			toReturn = env.evalList(original)
		}
		// fmt.Println("return of list => ", toReturn)
	case SexpInt, SexpFloat:
		toReturn = node
	case SexpSymbol:
		original, ok := node.(SexpSymbol)
		if ok {
			toReturn = env.evalSymbol(original, []Sexp{})
		}
	case SexpFunctionLiteral:
		original, ok := node.(SexpFunctionLiteral)
		if ok {
			toReturn = env.evalFunctionLiteral(&original)
		} else {
			log.Fatal("Error evaluating function literal!")
		}
	case SexpFunctionCall:
		original, ok := node.(SexpFunctionCall)
		if ok {
			toReturn = env.evalFunctionCall(&original)
		} else {
			log.Fatal("Error evaluating function call")
		}
	case SexpArray:
		//necessary for handling parameters in anon functions in macros
		//note we do this because we CAN'T CHANGE ORIGINAL FUNCTION DEFINITION
		//would cause problems with recursion
		new := make([]Sexp, 0)
		for index := range i.value {
			new = append(new, env.evalNode(i.value[index]))
		}
		toReturn = SexpArray{ofType: ARRAY, value: new}
	default:
		//TODO: fix this later
		fmt.Println(node)
		fmt.Println(reflect.TypeOf(node))
		log.Fatal("error unexpected node")
	}
	return toReturn
}

//evaluates and interprets our AST
func Eval(nodes []Sexp, env *Env) []string {
	res := make([]string, 0)
	for _, node := range nodes {
		curr := env.evalNode(node)
		if curr != nil {
			res = append(res, curr.String())
		}
	}
	return res
}

//method which exposes eval to other packages which call this as an API to get a result
func EvalSource(source string) ([]string, error) {
	tokens := Read(strings.NewReader(source))
	ast, err := Parse(tokens)
	if err != nil {
		return nil, errors.New("Error parsing!")
	}
	env := InitState()
	return Eval(ast, env), nil
}

// func listLen(expr Sexp) int {
// 	sz := 0
// 	var list *SexpPair
// 	ok := false
// 	for expr != nil {
// 		list, ok = expr.(*SexpPair)
// 		if !ok {
// 			log.Fatal("ListLen() called on non-list")
// 		}
// 		sz++
// 		expr = list.tail
// 	}
// 	return sz
// }

//helper function to return a list of Sexp nodes from a linked list of cons cell
func makeList(s SexpPair) []Sexp {
	toReturn := make([]Sexp, 0)
	for {
		toReturn = append(toReturn, s.head)
		switch s.tail.(type) {
		case SexpPair:
			s = s.tail.(SexpPair)
			continue
		}
		break
	}
	return toReturn
}

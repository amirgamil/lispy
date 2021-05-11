package lispy

import (
	"log"
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

func InitState() *Env {
	//add more ops as need for function bodies, assignments etc
	env := new(Env)
	env.store = make(map[string]Value)
	//this is the global reference, so set the pointer to nil
	env.global = nil
	return env
}

func (env *Env) evalSymbol(s SexpSymbol, args []Sexp) Sexp {
	switch s.ofType {
	case SYMBOL:
		return getVarBinding(env, s.value, args)
	case PLUS, MINUS, MULTIPLY, DIVIDE:
		return binaryOperation(env, s.value, args)
	case EQUAL, GEQUAL, LEQUAL, GTHAN, LTHAN:
		return relationalOperator(env, s.value, args)
	case AND, OR, NOT:
		return logicalOperator(env, s.value, args)
	case TRUE, FALSE, QUOTE, STRING:
		return s
	case IF:
		return conditionalStatement(env, s.value, args)
	case DEFINE:
		return varDefinition(env, args[0].String(), args[1:])
	case PRINT:
		return printlnStatement(env, s.String(), args)
	}
	return nil
}

func (env *Env) evalFunctionLiteral(s *SexpFunctionLiteral) Sexp {
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
	switch n.head.(type) {
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
		case PRINT:
			if n.tail == nil {
				log.Fatal("Error trying to print nothing!")
			}
			arguments = makeList(tail)
			toReturn = env.evalSymbol(symbol, arguments)
		case IF:
			if !isTail {
				log.Fatal("Error - please provide a valid condition for the if statement")
			}
			arguments = append(arguments, env.evalList(tail))
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
		case PLUS, MINUS, MULTIPLY, DIVIDE, GEQUAL, LEQUAL, GTHAN, LTHAN, AND, OR, NOT, EQUAL:
			//loop through elements in the list and carry out operation, will need to be adapted as we add more functionality
			if !isTail {
				log.Fatal("Error, incorrect arguments passed into binary/relational operator")
			}
			for {
				arguments = append(arguments, env.evalNode(tail.head))
				switch tail.tail.(type) {
				case SexpPair:
					tail = tail.tail.(SexpPair)
					continue
				}
				break
			}

			toReturn = env.evalSymbol(symbol, arguments)
		default:
			toReturn = env.evalSymbol(symbol, []Sexp{})
		}
	case SexpFunctionCall, SexpFunctionLiteral:
		toReturn = env.evalNode(n.head)
		//in a function literal, body should only be on Sexp, if there is more, throw an error
		//in a function call, arguments will be pased into SexpFunctionCall so similar idea
		if n.tail != nil {
			log.Fatal("Error interpreting function declaration or literal - ensure only one Sexp in body of function literal!")
		}
	case SexpInt, SexpFloat:
		toReturn = n.head
	case SexpPair:
		original, ok := n.head.(SexpPair)
		if ok {
			toReturn = env.evalList(original)
		} else {
			log.Fatal("error interpreting nested list")
		}
	//if it's just a list without a symbol at the front, treat it as data and return it
	default:
		toReturn = n.head
	}
	return toReturn
}

//wrapper for evaluating an individual Sexp node in our AST
func (env *Env) evalNode(node Sexp) Sexp {
	var toReturn Sexp
	switch node.(type) {
	case SexpPair:
		//Assert type since ast is composed of generic Sexp interface
		original, ok := node.(SexpPair)
		if ok {
			toReturn = env.evalList(original)
		}
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
	default:
		//TODO: fix this later
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

func listLen(expr Sexp) int {
	sz := 0
	var list *SexpPair
	ok := false
	for expr != nil {
		list, ok = expr.(*SexpPair)
		if !ok {
			log.Fatal("ListLen() called on non-list")
		}
		sz++
		expr = list.tail
	}
	return sz
}

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

package lispy

import (
	"fmt"
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
	//fix this bit
	parent *Env
}

//allow functionvalue to implement value
func (env Env) String() string {
	data := make([]string, 0)
	for _, val := range env.store {
		data = append(data, val.String())
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

func (env *Env) evalList(n SexpList) Sexp {
	var toReturn Sexp
	//empty string
	if len(n.value) == 0 {
		return SexpList{}
	}

	switch n.value[0].(type) {
	case SexpSymbol:
		symbol, ok := n.value[0].(SexpSymbol)
		if !ok {
			log.Fatal("error trying to interpret symbol")
		}
		arguments := make([]Sexp, 0)
		switch symbol.ofType {
		case DEFINE:
			//binding to a variable
			toReturn = env.evalSymbol(symbol, []Sexp{n.value[1], n.value[2]})
		case PRINT:
			if len(n.value) <= 1 {
				log.Fatal("Error trying to print nothing!")
			}
			toReturn = env.evalSymbol(symbol, n.value[1:])
		case IF:
			if len(n.value) < 3 {
				log.Fatal("Syntax error, too few arguments to if")
			}
			//condition for the if statement will be a list
			condition, ok := n.value[1].(SexpList)
			if !ok {
				log.Fatal("Error - please provide a valid condition for the if statement")
			}
			arguments = append(arguments, env.evalList(condition))
			//note we don't want to evaluate the nodes before we check the result of that condition - delegate that to later
			arguments = append(arguments, n.value[2])
			if len(n.value) == 4 {
				arguments = append(arguments, n.value[3])
			}
			toReturn = env.evalSymbol(symbol, arguments)
		case DO:
			//if symbol is do, we just evaluate the nodes and return the (result of the) last node
			//note do's second element will be a list of lists so we need to unwrap it
			doList, isDoList := n.value[1].(SexpList)
			if !isDoList {
				fmt.Println("Error parsing body of do statement")
			}
			for i := 1; i < len(doList.value); i++ {
				toReturn = env.evalNode(doList.value[i])
			}
		case PLUS, MINUS, MULTIPLY, DIVIDE, GEQUAL, LEQUAL, GTHAN, LTHAN, AND, OR, NOT, EQUAL:
			//loop through elements in the list and carry out operation, will need to be adapted as we add more functionality
			for i := 1; i < len(n.value); i++ {
				arguments = append(arguments, env.evalNode(n.value[i]))
			}
			toReturn = env.evalSymbol(symbol, arguments)
		default:
			toReturn = env.evalSymbol(symbol, []Sexp{})
		}
	case SexpFunctionCall, SexpFunctionLiteral:
		toReturn = env.evalNode(n.value[0])
	case SexpInt, SexpFloat:
		toReturn = n.value[0]
	case SexpList:
		original, ok := n.value[0].(SexpList)
		if ok {
			toReturn = env.evalList(original)
		} else {
			log.Fatal("error interpreting nested list")
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
	switch node.(type) {
	case SexpList:
		//Assert type since ast is composed of generic Sexp interface
		original, ok := node.(SexpList)
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
		res = append(res, env.evalNode(node).String())
	}
	return res
}

package lispy

import (
	"log"
)

//not convenient to have binOps as separate but for now, this is what we'll work with
type Env struct {
	store map[string]interface{}
}

func InitState() *Env {
	//add more ops as need for function bodies, assignments etc
	env := new(Env)
	env.store = make(map[string]interface{})
	return env
}

func (env *Env) evalSymbol(s SexpSymbol, args []Sexp) Sexp {
	switch s.ofType {
	case SYMBOL:
		return getBinding(env, s.value, args)
	case PLUS, MINUS, MULTIPLY, DIVIDE:
		return binaryOperation(env, s.value, args)
	case GEQUAL, LEQUAL, GTHAN, LTHAN:
		return relationalOperator(env, s.value, args)
	case AND, OR, NOT:
		return logicalOperator(env, s.value, args)
	case TRUE, FALSE:
		return s
	case QUOTE:
		return s
	case IF:
		return conditionalStatement(env, s.value, args)
	case DEFINE:
		return definition(env, args[0].String(), args[1:])
	case PRINT:
		return printlnStatement(env, s.String(), args)
	}
	//TODO: fix this bit
	return SexpSymbol{} //env.getBinding(s.value)
}

func (env *Env) evalNumber(n SexpInt) int {
	return int(n)
}

func (env *Env) evalList(n List) Sexp {
	var toReturn Sexp
	//empty string
	if len(n) == 0 {
		return List{}
	}

	switch n[0].(type) {
	case SexpSymbol:
		symbol, ok := n[0].(SexpSymbol)
		if !ok {
			log.Fatal("error trying to interpret symbol")
		}
		arguments := make([]Sexp, 0)
		switch symbol.ofType {
		case DEFINE:
			toReturn = env.evalSymbol(symbol, []Sexp{n[1], n[2]})
		case PRINT:
			if len(n) <= 1 {
				log.Fatal("Error trying to print nothing!")
			}
			toReturn = env.evalSymbol(symbol, n[1:])
		case IF:
			if len(n) < 3 {
				log.Fatal("Syntax error, too few arguments to if")
			}
			//condition for the if statement will be a list
			condition, ok := n[1].(List)
			if !ok {
				log.Fatal("Error - please provide a valid condition for the if statement")
			}
			arguments = append(arguments, env.evalList(condition))
			arguments = append(arguments, env.evalNode(n[2]))
			if len(n) == 4 {
				arguments = append(arguments, env.evalNode(n[3]))
			}
			toReturn = env.evalSymbol(symbol, arguments)
		case PLUS, MINUS, MULTIPLY, DIVIDE, GEQUAL, LEQUAL, GTHAN, LTHAN, AND, OR, NOT:
			//loop through elements in the list and carry out operation, will need to be adapted as we add more functionality
			for i := 1; i < len(n); i++ {
				arguments = append(arguments, env.evalNode(n[i]))
			}
			toReturn = env.evalSymbol(symbol, arguments)
		default:
			toReturn = env.evalSymbol(symbol, []Sexp{})
		}
	//if it's just a list without a symbol at the front, treat it as data and return it
	case List:
		original, ok := n[0].(List)
		if ok {
			toReturn = env.evalList(original)
		}
		log.Fatal("error interpreting nested list")
	default:
		toReturn = n
	}
	return toReturn
}

//wrapper for evaluating an individual Sexp node in our AST
func (env *Env) evalNode(node Sexp) Sexp {
	var toReturn Sexp
	switch node.(type) {
	case List:
		//Assert type since ast is composed of generic Sexp interface
		original, ok := node.(List)
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

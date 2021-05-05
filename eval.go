package main

import (
	"log"
)

//not convenient to have binOps as separate but for now, this is what we'll work with
type Env struct {
	store map[string]interface{}
}

// func (e *Env) addDefinition(key string, value interface{}) {
// 	e.store[key] = value
// }

// func (e *Env) getBinding(key string) interface{} {
// 	if v, found := e.store[key]; found {
// 		return v
// 	}
// 	log.Fatal("Error, ", key, " has not previously been defined!")
// 	return nil
// }

func initState() *Env {
	//add more ops as need for function bodies, assignments etc
	env := new(Env)
	env.store = make(map[string]interface{})
	return env
}

func (env *Env) evalSymbol(s SexpSymbol, args []Sexp) Sexp {
	switch s.ofType {
	case PLUS, MINUS, MULTIPLY, DIVIDE:
		return binaryOperation(env, s.value, args)
	}
	//TODO: fix this bit
	return SexpSymbol{} //env.getBinding(s.value)
}

func (env *Env) evalNumber(n Number) int {
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
		//loop through elements in the list and carry out operation, will need to be adapted as we add more functionality
		for i := 1; i < len(n); i++ {
			arguments = append(arguments, env.evalNode(n[i]))
		}
		toReturn = env.evalSymbol(symbol, arguments)
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
	case Number:
		toReturn = node
	default:
		//TODO: fix this later
		log.Fatal("error unexpected onode")
	}
	return toReturn
}

//evaluates and interprets our AST
func Eval(nodes []Sexp) []string {
	env := initState()
	res := make([]string, 0)
	for _, node := range nodes {
		res = append(res, env.evalNode(node).String())
	}
	return res
}

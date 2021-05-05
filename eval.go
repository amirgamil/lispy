package main

import (
	"log"
	"strconv"
)

func add(x int, y int) int {
	return x + y
}

func subtract(x int, y int) int {
	return x - y
}

func multiply(x int, y int) int {
	return x * y
}

func divide(x int, y int) int {
	return x / y
}

type fn func(int, int) int

//eventually change struct and also to capture environment
var ops map[string]fn

func initState() {
	//add more ops as need for function bodies, assignments etc
	ops = make(map[string]fn)
	ops["+"] = add
	ops["-"] = subtract
	ops["*"] = multiply
	ops["/"] = divide
}

func evalSymbol(s Symbol) fn {
	return ops[s.value]
}

func evalNumber(n Number) int {
	return int(n)
}

func evalList(n List) string {
	var toReturn string
	switch n[0].(type) {
	case Symbol:
		res := 0
		original, ok := n[0].(Symbol)
		var operation fn
		if ok {
			operation = evalSymbol(original)
		} else {
			log.Fatal("error trying to interpret symbol")
		}
		for i := 1; i < len(n); i++ {
			test := evalNode(n[i])
			intNode, err := strconv.Atoi(test)
			if err != nil {
				log.Fatal("error casting to int: ", err)
			}
			res = operation(res, intNode)
		}
		toReturn = strconv.Itoa(res)
	case List:
		original, ok := n[0].(List)
		if ok {
			toReturn = evalList(original)
		}
		log.Fatal("error interpreting nested list")
		toReturn = "error!"
	default:
		toReturn = n.String()
	}
	return toReturn
}

func evalNode(node Sexp) string {
	switch node.(type) {
	case List:
		//Assert type since ast is composed of generic Sexp interface
		original, ok := node.(List)
		if ok {
			test := evalList(original)
			return test
		}
	case Number:
		original, ok := node.(Number)
		if ok {
			num := evalNumber(original)
			return strconv.Itoa(num)
		}
	default:
		return "error!"
	}
	return "error!"
}

func Eval(nodes []Sexp) []string {
	initState()
	res := make([]string, 0)
	for _, node := range nodes {
		res = append(res, evalNode(node))
	}
	return res
}

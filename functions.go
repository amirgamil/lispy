package main

import "log"

type LispyUserFunction func(env *Env, name string, args []Sexp) Sexp

func binaryOperation(env *Env, name string, args []Sexp) Sexp {
	res := Number(0)
	for _, arg := range args {
		//check types that we're dealing with numbers first
		x, ok1 := arg.(Number)
		if !ok1 {
			log.Fatal("Can't pass a non-Number type to an arithmetic calculation")
		}
		res = handleOperation(name, res, x)
	}
	return res
}

func handleOperation(name string, x Number, y Number) Number {
	var result Number
	switch name {
	case "+":
		result = add(x, y)
	case "-":
		result = subtract(x, y)
	case "*":
		result = multiply(x, y)
	case "/":
		result = divide(x, y)
	}
	return result
}

func add(x Number, y Number) Number {
	return x + y
}

func subtract(x Number, y Number) Number {
	return x - y
}

func multiply(x Number, y Number) Number {
	return x * y
}

func divide(x Number, y Number) Number {
	return x / y
}

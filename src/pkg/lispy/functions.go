package lispy

import (
	"fmt"
	"log"
	"strconv"
)

type LispyUserFunction func(env *Env, name string, args []Sexp) Sexp

/******* handle definitions *********/
//create new variable binding
func definition(env *Env, key string, args []Sexp) Sexp {
	var value Sexp
	for _, arg := range args {
		value = arg
		env.store[key] = value
	}
	return value
}

//retrieve existing binding
func getBinding(env *Env, key string, args []Sexp) Sexp {
	if v, found := env.store[key]; found {
		value, ok := v.(Sexp)
		if !ok {
			log.Fatal("error converting stored data to a string")
		}
		//TODO: fix ofType
		return SexpSymbol{ofType: INTEGER, value: value.String()}
	}
	log.Fatal("Error, ", key, " has not previously been defined!")
	return nil
}

/******* handle conditional statements *********/
func conditionalStatement(env *Env, name string, args []Sexp) Sexp {
	condition, okC := args[0].(SexpSymbol)
	var toReturn Sexp
	if !okC {
		log.Fatal("Error interpreting condition for the if statement!")
	}
	//TODO: adapt this for expressions or functions specifically? Not sure how do that
	if condition.ofType == TRUE {
		toReturn = args[1]
	} else {
		if len(args) > 2 {
			toReturn = args[2]
		} else {
			//no provided else block despite the condition evaluating to such
			toReturn = SexpSymbol{ofType: FALSE, value: "nil"}
		}
	}
	return toReturn
}

/******* handle println statements *********/
func printlnStatement(env *Env, name string, args []Sexp) Sexp {
	for _, arg := range args {
		fmt.Println(arg)
		fmt.Println(arg.String())
	}
	return SexpSymbol{}
}

/******* handle logical (and or not) operations *********/
func logicalOperator(env *Env, name string, args []Sexp) Sexp {
	result := true
	tokenType := TRUE
	if len(args) == 0 {
		log.Fatal("Invalid syntax, pass in more than logical operator!")
	}
	//not can only take one parameter so check that first
	if name == "not" {
		if len(args) > 1 {
			log.Fatal("Error, cannot pass more than one logical operator to not!")
		}
		result = handleLogicalOp(name, getBoolFromTokenType(args[0]))
	} else {
		if len(args) < 2 {
			log.Fatal("Error, cannot carry out an ", name, " operator with only 1 condition!")
		}
		//for and, or, loop through the arguments and aggregate
		for _, arg := range args {
			result = handleLogicalOp(name, result, getBoolFromTokenType(arg))
			if result == false {
				//note we can't break early beacuse of the or operator
				tokenType = FALSE
			}
		}
	}

	return SexpSymbol{ofType: tokenType, value: strconv.FormatBool(result)}
}

//helper code to keep code DRY
func getBoolFromTokenType(symbol Sexp) bool {
	value, ok := symbol.(SexpSymbol)
	if !ok {
		log.Fatal("Error, invalid input to logical operation!")
	}
	if value.ofType == TRUE {
		return true
	}
	return false
}

func handleLogicalOp(name string, log ...bool) bool {
	var result bool
	switch name {
	case "not":
		result = !log[0]
	case "or":
		result = log[0] || log[1]
	case "and":
		result = log[0] && log[1]
	}
	return result
}

/******* handle relational operations *********/
func relationalOperator(env *Env, name string, args []Sexp) Sexp {
	result := true
	tokenType := TRUE
	initial, ok := args[0].(Number)

	if !ok {
		log.Fatal("Error, must provide only a number with a relational operator")
	}
	for i := 1; i < len(args); i++ {
		curr, ok := args[i].(Number)
		if !ok {
			log.Fatal("Error, must provide only a number with a relational operator")
		}
		result := handleRelOperator(name, initial, curr)
		if result == false {
			tokenType = FALSE
		}
	}
	return SexpSymbol{ofType: tokenType, value: strconv.FormatBool(result)}
}

func handleRelOperator(name string, x Number, y Number) bool {
	var result bool
	switch name {
	case ">":
		result = x > y
	case ">=":
		result = x >= y
	case "<":
		result = x < y
	case "<=":
		result = x < y
	}
	return result
}

/******* handle binary arithmetic operations *********/
func binaryOperation(env *Env, name string, args []Sexp) Sexp {
	res := Number(0)
	for _, arg := range args {
		//check types that we're dealing with numbers first
		x, ok1 := arg.(Number)
		if !ok1 {
			log.Fatal("Can't pass a non-Number type to an arithmetic calculation")
		}
		res = handleBinOperation(name, res, x)
	}
	return res
}

func handleBinOperation(name string, x Number, y Number) Number {
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

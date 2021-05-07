package lispy

import (
	"fmt"
	"log"
	"strconv"
)

type LispyUserFunction func(env *Env, name string, args []Sexp) Sexp

/******* handle definitions *********/
//create new variable binding
func varDefinition(env *Env, key string, args []Sexp) Sexp {
	var value Sexp
	for _, arg := range args {
		value = arg
		env.store[key] = value
	}
	return value
}

//retrieve existing variable binding
func getVarBinding(env *Env, key string, args []Sexp) Sexp {
	if v, found := env.store[key]; found {
		value, ok := v.(Sexp)
		if !ok {
			log.Fatal("error converting stored data to a string")
		}
		return value
	}
	log.Fatal("Error, ", key, " has not previously been defined!")
	return nil
}

//create new function binding
func funcDefinition(env *Env, s *SexpFunctionLiteral) Sexp {
	env.store[s.name] = FunctionValue{defn: s, parent: env}
	//fix this
	return SexpSymbol{ofType: STRING, value: (*s).name}
}

func getFuncBinding(env *Env, s *SexpFunctionCall) Sexp {
	node, isFuncLiteral := env.store[s.name].(FunctionValue)
	if !isFuncLiteral {
		log.Fatal("Error, badly defined function")
	}
	//set the passed in data to the arguments of the function
	for i, arg := range node.defn.arguments.value {
		env.store[arg.String()] = s.arguments.value[i]
	}
	return evalLispyFunction(env, node)
}

func evalLispyFunction(env *Env, fn FunctionValue) Sexp {
	return env.evalNode(fn.defn.body)
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
		toReturn = env.evalNode(args[1])
	} else {
		if len(args) > 2 {
			toReturn = env.evalNode(args[2])
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
	if value.ofType != FALSE {
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
	initial, ok := args[0].(SexpInt)

	if !ok {
		log.Fatal("Error, must provide only a  SexpInt with a relational operator")
	}
	for i := 1; i < len(args); i++ {
		curr, ok := args[i].(SexpInt)
		if !ok {
			log.Fatal("Error, must provide only a SexpInter with a relational operator")
		}
		result := handleRelOperator(name, initial, curr)
		if result == false {
			tokenType = FALSE
		}
	}
	return SexpSymbol{ofType: tokenType, value: strconv.FormatBool(result)}
}

func handleRelOperator(name string, x SexpInt, y SexpInt) bool {
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
	res := args[0]
	for i := 1; i < len(args); i++ {
		//pass in new argument under consideration first for compare operation
		switch term := res.(type) {
		case SexpFloat:
			res = numericMatchFloat(name, term, args[i])
		case SexpInt:
			res = numericMatchInt(name, term, args[i])

		}
	}
	return res
}

func numericMatchInt(name string, x SexpInt, y Sexp) Sexp {
	var res Sexp
	switch i := y.(type) {
	case SexpInt:
		res = numericOpInt(name, x, i)
	case SexpFloat:
		res = numericOpFloat(name, SexpFloat(x), i)
	default:
		log.Fatal("Error adding two numbers!")
	}
	return res
}

func numericMatchFloat(name string, x SexpFloat, y Sexp) Sexp {
	var res Sexp
	switch i := y.(type) {
	case SexpInt:
		res = numericOpFloat(name, x, SexpFloat(i))
	case SexpFloat:
		res = numericOpFloat(name, x, i)
	default:
		log.Fatal("Error adding two numbers!")
	}
	return res
}

func numericOpInt(name string, x SexpInt, y SexpInt) Sexp {
	var res Sexp
	switch name {
	case "+":
		res = x + y
	case "-":
		res = x - y
	case "/":
		if x%y == 0 {
			res = x / y
		} else {
			res = SexpFloat(x) / SexpFloat(y)
		}
	case "*":
		res = x * y
	default:
		log.Fatal("Error invalid operation")
	}
	return res

}

func numericOpFloat(name string, x SexpFloat, y SexpFloat) Sexp {
	var res Sexp
	switch name {
	case "+":
		res = x + y
	case "-":
		res = x - y
	case "/":
		res = x / y
	case "*":
		res = x * y
	}
	return res
}

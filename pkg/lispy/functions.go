package lispy

import (
	"fmt"
	"log"
	"strconv"
)

type LispyUserFunction func(env *Env, name string, args []Sexp) Sexp

//turns this into something that can take an annonymous inner function and return LispyUserFunction?
//or FunctionValue to store in the environment?
// func makeLispyFunc()

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
	//FunctionValue is a compile-time representation of a function
	env.store[s.name] = FunctionValue{defn: s}
	//fix this
	return SexpSymbol{ofType: STRING, value: "#user/" + s.name}
}

func getFuncBinding(env *Env, s *SexpFunctionCall) Sexp {
	//note quite critically, we need to evaluate the result of any expression arguments BEFORE we set them
	//(before any old values get overwritten)
	newExprs := make([]Sexp, 0)
	for _, toEvaluate := range makeList(s.arguments) {
		newExprs = append(newExprs, env.evalNode(toEvaluate))
	}

	node, isFuncLiteral := env.store[s.name].(FunctionValue)
	if !isFuncLiteral {
		log.Fatal("Error, badly defined function")
	}
	//Call LispyUserFunction if this is a builtin function
	//note if user-defined version exists, then it takes precedence (to ensure idea of macro functions correctly)
	if node.defn.userfunc != nil && node.defn.body == nil {
		return node.defn.userfunc(env, s.name, newExprs)
	}
	//check we have the correct number of parameters
	if len(node.defn.arguments.value) != len(newExprs) {
		log.Fatal("Incorrect number of arguments passed in!")
	}
	//load the passed in data to the arguments of the function in the environment
	for i, arg := range node.defn.arguments.value {
		env.store[arg.String()] = newExprs[i]
	}
	//evaluate user-defined function
	return env.evalNode(node.defn.body)
}

//helper function to take a function and return a function literal which can be saved to the environment
func makeUserFunction(name string, function LispyUserFunction) FunctionValue {
	var sfunc SexpFunctionLiteral
	sfunc.name = name
	sfunc.userfunc = function
	return FunctionValue{defn: &sfunc}
}

/******* cars, cons, cdr **********/
//helper function to unwrap quote data
func unwrap(arg Sexp) SexpPair {
	pair1, isPair1 := arg.(SexpPair)
	if !isPair1 {
		log.Fatal("Error unwrapping for built in functions")
	}
	if pair1.tail == nil {
		return pair1
	} else {
		//to allow for recursive calls from cons (which remove the double wrap), so return the correct depth
		return SexpPair{head: pair1, tail: nil}
	}

}

func car(env *Env, name string, args []Sexp) Sexp {
	//need to unwrap twice since function call arguments wrap inner arguments in a SexpPair
	//so we have SexpPair{head: SexpPair{...}}
	pair1 := unwrap(args[0])
	switch i := pair1.head.(type) {
	case SexpPair:
		return i.head
	case SexpInt, SexpFloat:
		return i
	}
	return nil
}

func cdr(env *Env, name string, args []Sexp) Sexp {
	pair1 := unwrap(args[0])
	switch i := pair1.head.(type) {
	case SexpPair:
		return i.tail
	default:
		log.Fatal("argument 0 of cdr has wrong type!")
	}
	return nil
}

func cons(env *Env, name string, args []Sexp) Sexp {
	if len(args) < 2 {
		log.Fatal("Incorrect number of arguments!")
	}

	//unwrap the list in the block quote (need to evaluate first to allow for recursive calls)
	list := unwrap(args[1])
	newHead := consHelper(args[0], list.head)
	return newHead
}

func consHelper(a Sexp, b Sexp) SexpPair {
	return SexpPair{a, b}
}

/******* return quote *********/
func returnQuote(args []Sexp) Sexp {
	list, isList := args[0].(SexpPair)
	if !isList {
		return args[0]
	}
	l := makeList(list)
	stringQuote := ""
	for _, el := range l {
		stringQuote += el.String()
	}
	return SexpSymbol{ofType: STRING, value: stringQuote}
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
		res := env.evalNode(arg)
		fmt.Print(res.String(), " ")
		// return res
	}
	fmt.Println()
	return nil
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
		result = handleLogicalOp(name, getBoolFromTokenType(env.evalNode(args[0])))
	} else {
		if len(args) < 2 {
			log.Fatal("Error, cannot carry out an ", name, " operator with only 1 condition!")
		}
		//for and, or, loop through the arguments and aggregate
		for _, arg := range args {
			result = handleLogicalOp(name, result, getBoolFromTokenType(env.evalNode(arg)))
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
	switch i := symbol.(type) {
	case SexpSymbol:
		if i.ofType == FALSE {
			return false
		}
		return true
	default:
		return true
	}
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
	orig := env.evalNode(args[0])

	for i := 1; i < len(args); i++ {
		curr := env.evalNode(args[i])
		switch i := orig.(type) {
		case SexpFloat:
			result = relationalOperatorMatchFloat(name, i, curr)
		case SexpInt:
			result = relationalOperatorMatchInt(name, i, curr)
		default:
			log.Fatal("Error, must provide only a SexpInter with a relational operator")
		}
		if result == false {
			tokenType = FALSE
			break
		}
	}
	return SexpSymbol{ofType: tokenType, value: getBoolFromString(result)}
}

func relationalOperatorMatchFloat(name string, x SexpFloat, y Sexp) bool {
	var res bool
	switch i := y.(type) {
	case SexpFloat:
		res = handleRelOperator(name, x, i)
	case SexpInt:
		res = handleRelOperator(name, x, SexpFloat(i))
	default:
		log.Fatal("Invalid expression for relational operator!")
	}
	return res
}

func relationalOperatorMatchInt(name string, x SexpInt, y Sexp) bool {
	var res bool
	switch i := y.(type) {
	case SexpFloat:
		res = handleRelOperator(name, SexpFloat(x), i)
	case SexpInt:
		res = handleRelOperator(name, SexpFloat(x), SexpFloat(i))
	default:
		log.Fatal("Invalid expression for relational operator!")
	}
	return res
}

func getBoolFromString(boolean bool) string {
	var res string
	switch boolean {
	case true:
		res = "true"
	case false:
		res = "false"
	default:
		log.Fatal("Error with passed in bool")
	}
	return res
}

func handleRelOperator(name string, x SexpFloat, y SexpFloat) bool {
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
	case "=":
		result = x == y
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
		if y == 0 {
			log.Fatal("Error attempted division by 0")
		}
		res = x / y
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
		if y == 0 {
			log.Fatal("Error attempted division by 0")
		}
		res = x / y
	case "*":
		res = x * y
	}
	return res
}

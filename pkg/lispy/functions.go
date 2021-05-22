package lispy

import (
	"fmt"
	"log"
	"math"
	"reflect"
)

type LispyUserFunction func(env *Env, name string, args []Sexp) Sexp

/******* handle definitions *********/
//create new variable binding
func varDefinition(env *Env, key string, args []Sexp) Sexp {
	var value Sexp
	for _, arg := range args {
		value = env.evalNode(arg)
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
	name := s.name
	node, isFuncLiteral := env.store[name].(FunctionValue)
	if !isFuncLiteral {
		//check if this is a reference to another function / variable
		funcName, isVar := env.store[s.name].(Value)
		if !isVar {
			log.Fatal("Error, badly defined function trying to be called: ", s.name)
		}
		s.name = funcName.String()
		//don't know how deep the reference so need to recurse
		return getFuncBinding(env, s)
	}

	//note quite critically, we need to evaluate the result of any expression arguments BEFORE we set them
	//(before any old values get overwritten)
	newExprs := make([]Sexp, 0)

	if node.defn.macro {
		//pass the args directly, macro takes in one input so we can do this directly
		env.store[node.defn.arguments.value[0].String()] = s.arguments.head
		macroRes := env.evalNode(node.defn.body)
		fmt.Println("macro => ", macroRes)
		//evaluate the result of the macro transformed input
		return env.evalNode(macroRes)
	}
	//otherwise not a macro, so evaluate all of the arguments before calling the function
	if s.arguments.head != nil {
		fmt.Println("args: ", s.arguments)
		for _, toEvaluate := range makeList(s.arguments) {
			//TODO: figure why adding this in makeList is causing problems
			if toEvaluate != nil {
				newExprs = append(newExprs, env.evalNode(toEvaluate))
			}
		}
	}

	//load the passed in data to the arguments of the function in the environment
	for i, arg := range node.defn.arguments.value {
		env.store[arg.String()] = newExprs[i]
	}

	//Call LispyUserFunction if this is a builtin function
	//note if user-defined version exists, then it takes precedence (to ensure idea of macro functions correctly)
	if node.defn.userfunc != nil && node.defn.body == nil {
		return node.defn.userfunc(env, name, newExprs)
	}

	//check we have the correct number of parameters
	//only do this if not a macro
	if len(node.defn.arguments.value) != len(newExprs) {
		log.Fatal("Incorrect number of arguments passed in to ", node.defn.name)
	}

	//evaluate function
	return env.evalNode(node.defn.body)
}

//helper function to take a function and return a function literal which can be saved to the environment
func makeUserFunction(name string, function LispyUserFunction) FunctionValue {
	var sfunc SexpFunctionLiteral
	sfunc.name = name
	sfunc.userfunc = function
	return FunctionValue{defn: &sfunc}
}

/******* create list *********/
func createList(env *Env, name string, args []Sexp) Sexp {
	i := unwrapSList(args)
	return i
}

/******** quote **********/
func quote(env *Env, name string, args []Sexp) Sexp {
	return args[0]
}

//helper function to convert list of args into list of cons cells
//this method is very similar to makeSList, the only difference is that it unwraps quotes
//so that they are no longer stored as SexpPair{head: quote, tail: actual symbol}
func unwrapSList(expressions []Sexp) Sexp {
	if len(expressions) == 0 {
		return nil
	}
	switch i := expressions[0].(type) {
	case SexpPair:
		if i.tail == nil {
			return consHelper(i.head, unwrapSList(expressions[1:]))
		}
	}
	return consHelper(expressions[0], unwrapSList(expressions[1:]))

}

/******* cars, cons, cdr **********/

//helper function to unwrap quote data
func unwrap(arg Sexp) SexpPair {
	pair1, isPair1 := arg.(SexpPair)
	if !isPair1 {
		//check if we only have one item
		//some weird behavior with (car 1 2 3) not sure if needs to change
		switch i := arg.(type) {
		case SexpInt, SexpFloat, SexpSymbol:
			return SexpPair{head: SexpPair{head: i, tail: nil}, tail: nil}
		default:
			log.Fatal("Error unwrapping for built in functions")
		}
	}
	return SexpPair{head: pair1, tail: nil}
}

func car(env *Env, name string, args []Sexp) Sexp {
	if len(args) == 0 {
		log.Fatal("Uh oh, you need to pass an argument to car")
	}
	//need to unwrap twice since function call arguments wrap inner arguments in a SexpPair
	//so we have SexpPair{head: SexpPair{...}}
	pair1 := unwrap(args[0])
	switch i := pair1.head.(type) {
	case SexpPair:
		return i.head
	case SexpInt, SexpFloat:
		return i
	default:
		return nil
	}
}

func cdr(env *Env, name string, args []Sexp) Sexp {
	if len(args) == 0 {
		log.Fatal("Uh oh, you need to pass an argument to car")
	}
	pair1 := unwrap(args[0])
	switch i := pair1.head.(type) {
	case SexpPair:
		//if we cdr a one-item list, we should return an empty list
		if i.tail == nil {
			return SexpPair{}
		}
		return i.tail
	case SexpInt, SexpFloat, SexpSymbol:
		return SexpPair{}
	default:
		fmt.Println(reflect.TypeOf(i))
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

/******* handle conditional statements *********/
func conditionalStatement(env *Env, name string, args []Sexp) Sexp {
	var condition bool
	var toReturn Sexp
	switch i := args[0].(type) {
	case SexpFloat, SexpInt:
		condition = true
	case SexpPair:
		//empty list
		if i.head == nil {
			condition = false
		} else {
			condition = true
		}
	case SexpSymbol:
		if i.ofType == FALSE {
			condition = false
		} else {
			condition = true
		}
	default:
		log.Fatal("Error trying to interpret condition for if statement")
	}
	//TODO: adapt this for expressions or functions specifically? Not sure how do that
	if condition {
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
//These wrappers are necessary to map unique functions to the built-in symbols in the store
//This becomes important when passing (built-in) functions as parameters without knowing ahead of time which
//one will be used
func and(env *Env, name string, args []Sexp) Sexp {
	return logicalOperator(env, "and", args)
}

func or(env *Env, name string, args []Sexp) Sexp {
	return logicalOperator(env, "or", args)
}

func not(env *Env, name string, args []Sexp) Sexp {
	return logicalOperator(env, "not", args)
}

func logicalOperator(env *Env, name string, args []Sexp) Sexp {
	result := getBoolFromTokenType(env.evalNode(args[0]))
	if len(args) == 0 {
		log.Fatal("Invalid syntax, pass in more than logical operator!")
	}
	//not can only take one parameter so check that first
	if name == "not" {
		if len(args) > 1 {
			log.Fatal("Error, cannot pass more than one logical operator to not!")
		}
		result = handleLogicalOp(name, result)
	} else {
		if len(args) < 2 {
			log.Fatal("Error, cannot carry out an ", name, " operator with only 1 condition!")
		}
		//for and, or, loop through the arguments and aggregate
		for i := 1; i < len(args); i++ {
			result = handleLogicalOp(name, result, getBoolFromTokenType(env.evalNode(args[i])))
			if result == false {
				//note we can't break early beacuse of the or operator
				break
			}
		}
	}

	return getSexpSymbolFromBool(result)
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

func getSexpSymbolFromBool(truthy bool) SexpSymbol {
	switch truthy {
	case true:
		return SexpSymbol{ofType: TRUE, value: "true"}
	default:
		return SexpSymbol{ofType: FALSE, value: "false"}
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

/******* handle typeOf *********/
func typeOf(env *Env, name string, args []Sexp) Sexp {
	var typeCurr SexpSymbol
	if len(args) < 1 {
		log.Fatal("require a parameter to check type of")
	}
	switch i := args[0].(type) {
	case SexpInt:
		typeCurr = SexpSymbol{ofType: STRING, value: "int"}
	case SexpFloat:
		typeCurr = SexpSymbol{ofType: STRING, value: "float"}
	case SexpPair:
		if i.tail == nil {
			return typeOf(env, name, []Sexp{i.head})
		} else {
			typeCurr = SexpSymbol{ofType: STRING, value: "list"}
		}
	case SexpSymbol:
		typeCurr = SexpSymbol{ofType: STRING, value: "symbol"}
	case SexpFunctionLiteral:
		typeCurr = SexpSymbol{ofType: STRING, value: "funcLiteral"}
	case SexpFunctionCall:
		typeCurr = SexpSymbol{ofType: STRING, value: "funcCall"}
	default:
		log.Fatal("unexpected type!")
	}
	return typeCurr
}

/******* handle relational operations *********/
//These wrappers are necessary to map unique functions to the built-in symbols in the store
//This becomes important when passing (built-in) functions as parameters without knowing ahead of time which
//one will be used
func equal(env *Env, name string, args []Sexp) Sexp {
	return relationalOperator(env, "=", args)
}

func gequal(env *Env, name string, args []Sexp) Sexp {
	return relationalOperator(env, ">=", args)
}

func lequal(env *Env, name string, args []Sexp) Sexp {
	return relationalOperator(env, "<=", args)
}

func gthan(env *Env, name string, args []Sexp) Sexp {
	return relationalOperator(env, ">", args)
}

func lthan(env *Env, name string, args []Sexp) Sexp {
	return relationalOperator(env, "<", args)
}

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
		case SexpSymbol:
			result = relationalOperatorMatchSymbol(name, i, curr)
		case SexpPair:
			result = relationalOperatorMatchList(name, i, curr)
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

func relationalOperatorMatchList(name string, x SexpPair, y Sexp) bool {
	var res bool
	switch i := y.(type) {
	case SexpPair:
		list1 := makeList(x)
		list2 := makeList(i)
		if len(list1) != len(list2) {
			res = false
		} else {
			for i := 0; i < len(list1); i++ {
				res = list1[i] == list2[i]
			}
		}
	default:
		res = false
	}
	return res
}

func relationalOperatorMatchSymbol(name string, x SexpSymbol, y Sexp) bool {
	var res bool
	switch i := y.(type) {
	case SexpSymbol:
		res = handleRelOperatorSymbols(name, x, i)
	default:
		res = false
	}
	return res
}

func relationalOperatorMatchFloat(name string, x SexpFloat, y Sexp) bool {
	var res bool
	switch i := y.(type) {
	case SexpFloat:
		res = handleRelOperator(name, x, i)
	case SexpInt:
		res = handleRelOperator(name, x, SexpFloat(i))
	default:
		res = false
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
		res = false
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

func handleRelOperatorSymbols(name string, x SexpSymbol, y SexpSymbol) bool {
	var result bool
	switch name {
	case ">":
		result = (x.value > y.value)
	case ">=":
		result = (x.value >= y.value)
	case "<":
		result = (x.value < y.value)
	case "<=":
		result = (x.value <= y.value)
	case "=":
		result = (x.value == y.value)
	}
	return result
}

func handleRelOperator(name string, x SexpFloat, y SexpFloat) bool {
	var result bool
	switch name {
	case ">":
		result = (x > y)
	case ">=":
		result = (x >= y)
	case "<":
		result = (x < y)
	case "<=":
		result = (x <= y)
	case "=":
		result = (x == y)
	}
	return result
}

/******* handle binary arithmetic operations *********/
//These wrappers are necessary to map unique functions to the built-in symbols in the store
//This becomes important when passing (built-in) functions as parameters without knowing ahead of time which
//one will be used
func add(env *Env, name string, args []Sexp) Sexp {
	return binaryOperation(env, "+", args)
}

func minus(env *Env, name string, args []Sexp) Sexp {
	return binaryOperation(env, "-", args)
}

func multiply(env *Env, name string, args []Sexp) Sexp {
	return binaryOperation(env, "*", args)
}

func expo(env *Env, name string, args []Sexp) Sexp {
	return binaryOperation(env, "#", args)
}

func divide(env *Env, name string, args []Sexp) Sexp {
	return binaryOperation(env, "/", args)
}

func modulo(env *Env, name string, args []Sexp) Sexp {
	return binaryOperation(env, "%", args)
}

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
	case "#":
		res = SexpInt(math.Pow(float64(x), float64(y)))
	case "%":
		res = x % y
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
	case "#":
		res = SexpFloat(math.Pow(float64(x), float64(y)))
	case "%":
		res = SexpInt(int(x) % int(y))
	default:
		log.Fatal("Error invalid operation")

	}
	return res
}

package lispy

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"time"
)

type LispyUserFunction func(env *Env, name string, args []Sexp) Sexp

type FunctionThunkValue struct {
	env      *Env
	function FunctionValue
}

func (thunk FunctionThunkValue) String() string {
	return fmt.Sprintf("function thunk: %s", thunk.function)
}

//not ideal
func (thunk FunctionThunkValue) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	return nil
}

/******* handle definitions *********/
//create new variable binding
func varDefinition(env *Env, key string, args []Sexp) Sexp {
	value := args[0].Eval(env, &StackFrame{}, false)
	env.store[key] = value
	return value
}

//retrieve existing variable binding
func getVarBinding(env *Env, key string, args []Sexp) Sexp {
	if v, found := env.store[key]; found {
		value, ok := v.(Sexp)
		if !ok {
			//means empty list passed in
			return SexpPair{}
		}
		// newVal := value
		// if next, foundDeeper := env.store[newVal.String()]; foundDeeper {
		// 	_, isFunc := next.(FunctionValue)
		// 	//make sure it's not a function value and so we don't recurse into an error
		// 	if !isFunc {
		// 		return getVarBinding(env, newVal.String(), args)
		// 	}
		// }
		return value
	} else if env.parent != nil {
		return getVarBinding(env.parent, key, args)
	}
	log.Fatal("Error, ", key, " has not previously been defined!")
	return nil
}

//create new function binding
func (funcVal FunctionValue) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	name := frame.args[len(frame.args)-1].String()
	//FunctionValue is a compile-time representation of a function
	env.store[name] = funcVal
	dec(env)
	list := []Sexp{SexpSymbol{ofType: STRING, value: funcVal.defn.name}, funcVal.defn.arguments, funcVal.defn.body}
	return makeSList(list)
}

func evalFunc(env *Env, s *SexpFunctionCall, allowThunk bool) Sexp {
	name := s.name
	node, isFuncLiteral := env.store[name].(FunctionValue)
	if !isFuncLiteral {
		//check if this is a reference to another function / variable
		node, isFuncLiteral = getVarBinding(env, name, []Sexp{}).(FunctionValue)
		if !isFuncLiteral {
			log.Fatal("Error, badly defined function trying to be called: ", s.name)
		}
	}
	//note quite critically, we need to evaluate the result of any expression arguments BEFORE we set them
	//(before any old values get overwritten)
	newExprs := make([]Sexp, 0)
	if node.defn.macro {
		macroArgs := s.arguments
		switch i := s.arguments.head.(type) {
		case SexpPair:
			quote, isQuote := i.head.(SexpSymbol)
			//skip quote so it doesn't interfere with list manipulation in the macro
			if isQuote && quote.value == "" {
				pair, isPair := i.tail.(SexpPair)
				if isPair {
					macroArgs = pair
				}
			}
		}
		//pass the args directly, macro takes in one input so we can do this directly
		env.store[node.defn.arguments.value[0].String()] = macroArgs
		// fmt.Println("macro args => ", node.defn.body)
		macroRes := node.defn.body.Eval(env, &StackFrame{}, false)
		//uncomment line below to see macro-expansion
		// fmt.Println("macro => ", macroRes)
		finalRes := macroRes.Eval(env, &StackFrame{}, allowThunk)
		// fmt.Println(name, " res => ", finalRes)
		//evaluate the result of the macro transformed input
		return finalRes
	}
	//otherwise not a macro, so evaluate all of the arguments before calling the function
	if s.arguments.head != nil {
		// fmt.Println("args: ", s.arguments)
		//quote is a special form where we don't want to evaluate the args
		if name == "quote" {
			newExprs = makeList(s.arguments)
		} else {
			for _, toEvaluate := range makeList(s.arguments) {
				//TODO: figure why adding this in makeList is causing problems
				if toEvaluate != nil {
					//note pass false in case this is function call
					evaluatedArg := toEvaluate.Eval(env, &StackFrame{}, false)
					newExprs = append(newExprs, evaluatedArg)
					// fmt.Println("arg: ", toEvaluate, " res: ", evaluatedArg)
				}
			}
			// fmt.Println("evaluated params: ", newExprs, "\n for ", s.name, "\n\n\n")
		}

	}
	variableNumberOfArgs := false
	//load the passed in data to the arguments of the function in the environment
	for i, arg := range node.defn.arguments.value {
		//if arg has &, means it takes variable number of arguments, so create list of cons cells and set it to name pointing to variable arg
		if arg.String() == "&" {
			env.store[node.defn.arguments.value[i+1].String()] = makeSList(newExprs[i:])
			variableNumberOfArgs = true
			break
		} else {
			env.store[arg.String()] = newExprs[i]
		}
	}

	//check we have the correct number of parameters
	//only do this if not a macro or a built-in function (most of which take a variable number of args and handle invalid ones
	//internally)
	if len(node.defn.arguments.value) != len(newExprs) && node.defn.userfunc == nil && !variableNumberOfArgs {
		fmt.Println(len(node.defn.arguments.value), newExprs)
		log.Fatal("Incorrect number of arguments passed in to ", node.defn.name)
	}

	//Call LispyUserFunction if this is a builtin function
	//note if user-defined version exists, then it takes precedence (to ensure idea of macro functions correctly)
	if node.defn.userfunc != nil && node.defn.body == nil {
		return node.defn.userfunc(env, name, newExprs)
	}
	// if name == "set-new-env" {
	// 	fmt.Println("here")
	// }
	// fmt.Println(name)
	functionThunk := FunctionThunkValue{env: env, function: node}
	//if we're at a tail position inside a function body, return the thunk directly for tail call optimization
	if allowThunk {
		return functionThunk
	}
	//evaluate function
	return unwrapThunks(functionThunk)
}

//unwrap nested function calls into flat for loop structure for tail call optimization
func unwrapThunks(functionThunk FunctionThunkValue) Sexp {
	isTail := true
	var funcResult Sexp
	for isTail {
		funcResult = functionThunk.function.defn.body.Eval(functionThunk.env, &StackFrame{}, true)
		functionThunk, isTail = funcResult.(FunctionThunkValue)
		//fmt.Println("cheeky -> ", isTail, " ", funcResult)
	}
	return funcResult
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
	if i == nil {
		//return empty list ()
		return SexpPair{}
	}
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
		switch i := arg.(type) {
		case SexpInt, SexpFloat, SexpArray, SexpSymbol, FunctionValue:
			return SexpPair{head: SexpPair{head: i, tail: nil}, tail: nil}
		case SexpFunctionLiteral:
			argList := makeSList(i.arguments.value)
			//set up in list format
			list := makeSList([]Sexp{SexpSymbol{ofType: STRING, value: i.name}, argList, i.body})
			listPair, _ := list.(SexpPair)
			return listPair
		default:
			fmt.Println(reflect.TypeOf(arg), arg)
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
	case SexpInt, SexpFloat, SexpSymbol, SexpArray:
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
		if pair1.tail == nil {
			return SexpPair{}
		}
		return pair1.tail
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

//since quote is not stored as a special form, we need an internal function to check
/******* quote *********/
func isQuote(env *Env, name string, args []Sexp) Sexp {
	if len(args) == 0 {
		log.Fatal("Error checking quote type")
	}
	switch i := args[0].(type) {
	case SexpSymbol:
		if i.ofType == QUOTE || i.value == "quote" {
			return SexpSymbol{ofType: TRUE, value: "true"}
		}
	}
	return SexpSymbol{ofType: FALSE, value: "false"}
}

/******* swap *************/
//note swap only works for lists!
func swap(env *Env, name string, args []Sexp) Sexp {
	if len(args) == 0 {
		log.Fatal("Error trying to swap element")
	}
	//enforce swap only for lists
	list, isList := args[0].(SexpPair)
	if !isList {
		log.Fatal("Error trying to parse arguments of swap")
	}
	newList, isNewList := list.tail.(SexpPair)
	if !isNewList {
		log.Fatal("Error swapping non-list!")
	}
	fmt.Println(list.head.String())
	newVal := newList.head.Eval(env, &StackFrame{}, false)
	setValWhileKeyExists(env, list.head.String(), newVal)
	return newVal
}

//helper method for set
func setValWhileKeyExists(env *Env, key string, val Value) {
	if _, found := env.store[key]; found {
		env.store[key] = val
		if env.parent != nil {
			setValWhileKeyExists(env.parent, key, val)
		}
	}
}

/******* readstring *******/
//reads one object from a string
func readstring(env *Env, name string, args []Sexp) Sexp {
	if len(args) < 1 {
		log.Fatal("Error trying to read object from tring!")
	}
	stringObj, isString := args[0].(SexpSymbol)
	if !isString || stringObj.ofType != STRING {
		log.Fatal("Error trying to read an object from a non-string!")
	}
	res, err := evalHelper(stringObj.value)
	if err != nil {
		log.Fatal("Error trying to parse an object from a string !")
	}
	//readstring only reads first object
	return res[0]
}

/******* readline *********/
func readline(env *Env, name string, args []Sexp) Sexp {
	if len(args) > 0 {
		fmt.Print(args[0].String())
	}
	scanner := bufio.NewScanner(os.Stdin)
	var val string
	if scanner.Scan() {
		val = scanner.Text()
	}

	return SexpSymbol{ofType: STRING, value: val}
}

/******* string join *********/
func str(env *Env, name string, args []Sexp) Sexp {
	val := ""
	for _, arg := range args {
		val += arg.String()
	}
	return SexpSymbol{ofType: STRING, value: val}
}

/******* handle conditional statements *********/
func conditionalStatement(env *Env, name string, args []Sexp) Sexp {
	//fmt.Println(args)
	//put thunk as last argument
	thunk, isThunk := args[len(args)-1].(SexpSymbol)
	//TODO: improve this
	args = args[:len(args)-1]
	if !isThunk {
		log.Fatal("Error passing thunk into conditional statement")
	}
	allowThunk := thunk.ofType == TRUE
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
		fmt.Println(i)
		log.Fatal("Error trying to interpret condition for if statement")
	}
	if condition {
		toReturn = args[1].Eval(env, &StackFrame{}, allowThunk)
	} else {
		if len(args) > 2 && args[2] != nil {
			toReturn = args[2].Eval(env, &StackFrame{}, allowThunk)
		} else {
			//no provided else block despite the condition evaluating to such
			toReturn = SexpSymbol{ofType: FALSE, value: "nil"}
		}
	}
	return toReturn
}

/******* handle random numbers *********/
func random(env *Env, name string, args []Sexp) Sexp {
	if len(args) != 0 {
		log.Fatal("Error generating random number")
	}
	//generate a random seed, otherwise the same random number will be generated
	rand.Seed(time.Now().UnixNano())
	return SexpFloat(rand.Float64())
}

/******* applies function to list of args similar to function applyTo in Clojure *********/
func applyTo(env *Env, name string, args []Sexp) Sexp {
	if len(args) < 2 {
		log.Fatal("Error applying function to args")
	}
	functionLiteral, isFuncLiteral := args[0].(FunctionValue)
	if !isFuncLiteral {
		if env.store[args[0].String()] != nil {
			functionLiteral, isFuncLiteral = env.store[args[0].String()].(FunctionValue)
		}
		if !isFuncLiteral {
			log.Fatal("Error trying to apply a value that is not a function")
		}

	}
	arguments, isArgs := args[1].(SexpPair)
	if !isArgs {
		log.Fatal("Error applyTo only operates on lists!")
	}
	return SexpFunctionCall{name: functionLiteral.defn.name, arguments: arguments}.Eval(env, &StackFrame{}, false)
}

/******* handle type conversions for non-list *********/
func number(env *Env, name string, args []Sexp) Sexp {
	if len(args) > 1 {
		log.Fatal("Error casting to number")
	}
	switch i := args[0].(type) {
	case SexpSymbol:
		num, err := strconv.ParseFloat(i.value, 64)
		if err != nil {
			log.Fatal(err)
		}
		return SexpFloat(num)
	case SexpInt:
		return SexpFloat(i)
	case SexpFloat:
		return i
	default:
		log.Fatal("Error casting list to number")
		return nil
	}
}

func symbol(env *Env, name string, args []Sexp) Sexp {
	if len(args) > 1 {
		log.Fatal("Error casting to number")
	}
	return SexpSymbol{ofType: SYMBOL, value: args[0].String()}
}

/******* handle println statements *********/
func printlnStatement(env *Env, name string, args []Sexp) Sexp {
	for _, arg := range args {
		//fmt.Print(arg.String(), " ")
		return arg
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
	result := getBoolFromTokenType(args[0].Eval(env, &StackFrame{}, false))
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
			result = handleLogicalOp(name, result, getBoolFromTokenType(args[i].Eval(env, &StackFrame{}, false)))
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
	case SexpFunctionLiteral, SexpFunctionCall:
		typeCurr = SexpSymbol{ofType: STRING, value: "list"}
	default:
		fmt.Println(i)
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
	//recall we evaluated params before function call already
	orig := args[0]

	for i := 1; i < len(args); i++ {
		curr := args[i]
		switch i := orig.(type) {
		case SexpFloat:
			result = relationalOperatorMatchFloat(name, i, curr)
		case SexpInt:
			result = relationalOperatorMatchInt(name, i, curr)
		case SexpSymbol:
			result = relationalOperatorMatchSymbol(name, i, curr)
		case SexpPair:
			result = relationalOperatorMatchList(name, i, curr)
		case SexpFunctionLiteral:
			result = relationalOperatorMatchLiteral(name, i, curr)
		default:
			fmt.Println(args)
			log.Fatal("Error, unexpected type in relational operator")
		}
		if !result {
			tokenType = FALSE
			break
		}
	}
	return SexpSymbol{ofType: tokenType, value: getBoolFromString(result)}
}

func relationalOperatorMatchLiteral(name string, x SexpFunctionLiteral, y Sexp) bool {
	res := true
	switch i := y.(type) {
	case SexpFunctionLiteral:
		res = i.name == x.name
	default:
		res = false
	}
	return res
}

func relationalOperatorMatchList(name string, x SexpPair, y Sexp) bool {
	res := true
	switch i := y.(type) {
	case SexpPair:
		list1 := makeList(x)
		list2 := makeList(i)
		if len(list1) != len(list2) {
			res = false
		} else {
			for i := 0; i < len(list1); i++ {
				res = (res && list1[i] == list2[i])
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
	switch i := res.(type) {
	case SexpArray, SexpPair, SexpFunctionCall, SexpFunctionLiteral:
		log.Fatal("Invalid type passed to a binary operation!")
	case SexpSymbol:
		if i.value == "" {
			return binaryOperation(env, name, args[1:])
		}
	}
	for i := 1; i < len(args); i++ {
		//pass in new argument under consideration first for compare operation
		switch term := res.(type) {
		case SexpFloat:
			res = numericMatchFloat(name, term, args[i])
		case SexpInt:
			res = numericMatchInt(name, term, args[i])
		default:
			log.Fatal("Invalid type passed to a binary operation!")

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
	case SexpSymbol:
		if i.value == "" {
			return x
		}
	default:
		fmt.Println(y)
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
	case SexpSymbol:
		if i.value == "" {
			return x
		}
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

package lispy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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
}

//struct to store function arguments for now
type StackFrame struct {
	args []Sexp
}

//allow functionvalue to implement value
func (env Env) String() string {
	data := make([]string, 0)
	for key, val := range env.store {
		data = append(data, key+":"+val.String())
	}
	return strings.Join(data, " ")
}

func (f FunctionValue) String() string {
	//TODO: clean this up later
	return fmt.Sprintf("function value: %s", f.defn.String())
}

func returnDefinedFunctions() map[string]LispyUserFunction {
	functions := make(map[string]LispyUserFunction)
	functions["car"] = car
	functions["cdr"] = cdr
	functions["cons"] = cons
	functions["+"] = add
	functions["-"] = minus
	functions["/"] = divide
	functions["*"] = multiply
	functions["#"] = expo
	functions["%"] = modulo
	functions["="] = equal
	functions[">="] = gequal
	functions["<="] = lequal
	functions[">"] = gthan
	functions["<"] = lthan
	functions["and"] = and
	functions["or"] = or
	functions["not"] = not
	functions["println"] = printlnStatement
	functions["list"] = createList
	functions["type"] = typeOf
	functions["quote"] = quote
	functions["rand"] = random
	functions["number"] = number
	functions["symbol"] = symbol
	return functions
}

func InitState() *Env {
	//add more ops as need for function bodies, assignments etc
	env := new(Env)
	env.store = make(map[string]Value)
	for key, function := range returnDefinedFunctions() {
		env.store[key] = makeUserFunction(key, function)
	}
	//load library functions
	file, err := os.Open("lib/library.lpy")
	if err != nil {
		log.Fatal("Error opening file to read! ", err)
	}
	defer file.Close()
	errLib := EvalSourceIO(file, env)
	if errLib != nil {
		log.Fatal("Error loading library packages of lispy")
	}
	//this is the global reference, so set the pointer to nil
	env.global = nil
	return env
}

func (s SexpSymbol) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	switch s.ofType {
	case TRUE, FALSE, STRING:
		return s
	case IF:
		frame.args = append(frame.args, getSexpSymbolFromBool(allowThunk))
		return conditionalStatement(env, s.value, frame.args)
	case DEFINE:
		return varDefinition(env, frame.args[0].String(), frame.args[1:])
	case QUOTE:
		return s
	case SYMBOL:
		//if no argument then it's a variable
		if len(frame.args) == 0 {
			return getVarBinding(env, s.value, frame.args)
		}
		//otherwise assume this is a function call - this is MACROEXPANSION CODE!!
		argList, isList := frame.args[0].(SexpPair)
		if !isList {
			log.Fatal("Error trying to parse arguments for function call")
		}
		//check if this is an anonymous function the macro called
		if s.value == "fn" {
			params, isArray := argList.head.(SexpArray)
			if !isArray {
				log.Fatal("Error parsing anonymous function in macro expansion!")
			}
			bodyFunc, isValid := argList.tail.(SexpPair)
			if !isValid {
				log.Fatal("Error macroexpanding anon function!")
			}
			anonFunc := SexpFunctionLiteral{name: "fn", arguments: params, body: bodyFunc.head, userfunc: nil, macro: false}
			return anonFunc
		}
		funcCall := SexpFunctionCall{name: s.value, arguments: argList}
		return funcCall.Eval(env, frame, allowThunk)
	default:
		fmt.Println(s.ofType, " ", s.value, " args: ", frame.args)
		log.Fatal("Uh oh, weird symbol my dude")
		return nil
	}
}

func (s SexpFunctionLiteral) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	funcDefinition := FunctionValue{defn: &s}
	//append name of function to end of args
	frame.args = append(frame.args, SexpSymbol{ofType: STRING, value: s.name})
	funcDefinition.Eval(env, frame, allowThunk)
	return funcDefinition
}

func (s SexpFunctionCall) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	//each call should get its own environment for recursion to work
	functionCallEnv := new(Env)
	functionCallEnv.store = make(map[string]Value)
	//copy globals
	for key, element := range env.store {
		functionCallEnv.store[key] = element
	}
	return evalFunc(functionCallEnv, &s, allowThunk)
}

func (n SexpPair) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	var toReturn Sexp
	//empty string
	if n.head == nil {
		return SexpPair{}
	}
	tail, isTail := n.tail.(SexpPair)
	switch head := n.head.(type) {
	case SexpSymbol:
		symbol, ok := n.head.(SexpSymbol)
		if !ok {
			log.Fatal("error trying to interpret symbol")
		}
		arguments := make([]Sexp, 0)
		//process all arguments here for ease?
		switch symbol.ofType {
		case DEFINE:
			if !isTail {
				log.Fatal("Unexpected definition, missing value!")
			}
			newFrame := StackFrame{args: makeList(tail)}
			//binding to a variable
			toReturn = symbol.Eval(env, &newFrame, allowThunk)
		case QUOTE:
			if !isTail {
				log.Fatal("Error trying to interpret quote")
			}
			//don't evaluate the expression
			toReturn = tail.head
		case IF:
			if !isTail {
				fmt.Println("Error interpreting condition for the if statement")
			}
			//evaluating arguments so pass thunk as false
			arguments = append(arguments, tail.head.Eval(env, frame, false))
			statements, isValid := tail.tail.(SexpPair)
			if !isValid {
				log.Fatal("Error please provide valid responses to the if condition!")
			}
			res := makeList(statements)
			arguments = append(arguments, res...)
			newFrame := StackFrame{args: arguments}
			toReturn = symbol.Eval(env, &newFrame, allowThunk)
		case DO:
			//if symbol is do, we just evaluate the nodes and return the (result of the) last node
			//note do's second element will be a list of lists so we need to unwrap it
			if !isTail {
				log.Fatal("Error trying to interpret do statements")
			}
			for {
				toReturn = tail.head.Eval(env, frame, allowThunk)
				switch tail.tail.(type) {
				case SexpPair:
					tail = tail.tail.(SexpPair)
					continue
				}
				break
			}
		default:
			// fmt.Println("default symbol ", symbol, " in list -> ", tail)
			toReturn = symbol.Eval(env, &StackFrame{args: []Sexp{tail}}, allowThunk)
		}
	case SexpFunctionLiteral:
		//anonymous function, so handle differently
		if head.name == "fn" {
			//save body of function to the env then call
			head.Eval(env, frame, allowThunk)
			//check tail != nil for anon function with no parameters
			if !isTail && n.tail != nil {
				log.Fatal("Error interpreting anonymous function parameters")
			}
			funcCall := SexpFunctionCall{name: "fn", arguments: tail, body: nil}
			toReturn = funcCall.Eval(env, frame, allowThunk)
		} else {
			toReturn = head.Eval(env, frame, allowThunk)
			//in a function literal, body should only be on Sexp, if there is more, throw an error
			//in a function call, arguments will be pased into SexpFunctionCall so similar idea
			if n.tail != nil {
				log.Fatal("Error interpreting function declaration or literal - ensure only one Sexp in body of function literal!")
			}
		}
	case SexpFunctionCall:
		toReturn = head.Eval(env, frame, allowThunk)
	case SexpPair:
		original, ok := n.head.(SexpPair)
		if ok {
			toReturn = original.Eval(env, frame, allowThunk)
			//if this is an anon function from a macro, need to set it up as such
			funcLiteral, isFuncLiteral := toReturn.(SexpFunctionLiteral)
			if isFuncLiteral && funcLiteral.name == "fn" {
				//this is a function call so we can use the code above under case SexpFunctionLiteral
				//by artificially constructing a list as such
				toReturn = (SexpPair{head: funcLiteral, tail: n.tail}).Eval(env, frame, allowThunk)
			} else {
				//TODO: add a check for quote, otherwise invalid!
				//just a nested list so return entire list
				toReturn = n
			}
		} else {
			//TODO: might need to be fixed
			toReturn = SexpSymbol{FALSE, "false"}
		}
	//if it's just a list without a symbol at the front, treat it as data and return it
	default:
		toReturn = n
	}
	return toReturn
}

func (arr SexpArray) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	new := make([]Sexp, 0)
	for index := range arr.value {
		new = append(new, arr.value[index].Eval(env, frame, allowThunk))
	}
	return SexpArray{ofType: ARRAY, value: new}
}

func (s SexpFloat) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	return s
}

func (s SexpInt) Eval(env *Env, frame *StackFrame, allowThunk bool) Sexp {
	return s
}

//evaluates and interprets our AST
func (env *Env) Eval(nodes []Sexp) []string {
	res := make([]string, 0)
	frame := StackFrame{}
	for _, node := range nodes {
		curr := node.Eval(env, &frame, false)
		if curr != nil {
			res = append(res, curr.String())
		}
	}
	return res
}

//method which exposes eval to other packages which call this as an API to get a result
func EvalSource(source string) ([]string, error) {
	tokens := Read(strings.NewReader(source))
	ast, err := Parse(tokens)
	if err != nil {
		return nil, errors.New("Error parsing!")
	}
	env := InitState()
	return env.Eval(ast), nil
}

//used to load library packages into the env
func EvalSourceIO(source io.Reader, env *Env) error {
	tokens := Read(source)
	ast, err := Parse(tokens)
	if err != nil {
		return errors.New("Error parsing!")
	}
	env.Eval(ast)
	return nil
}

//helper function to return a list of Sexp nodes from a linked list of cons cell
func makeList(s SexpPair) []Sexp {
	toReturn := make([]Sexp, 0)
	for {
		toReturn = append(toReturn, s.head)
		switch s.tail.(type) {
		case SexpPair:
			s = s.tail.(SexpPair)
			continue
		}
		break
	}
	return toReturn
}

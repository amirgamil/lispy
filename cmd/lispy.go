package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/amirgamil/lispy/pkg/lispy"
)

// read
func read(str string) []Sexp {
	tokens := lispy.readStr(str)
	exprs, err := lispy.parse(tokens)
	if err != nil {
		log.Fatal("Error parsing")
	}
	return exprs
}

// eval
func eval(ast []lispy.Sexp, env *lispy.Env) []string {
	return Eval(ast, env)
}

// print
func print(res []string) {
	for _, result := range res {
		fmt.Println(result)
	}
}

// repl
func repl(str string, env *lispy.Env) {
	print(eval(read(str), env))
}

func main() {
	// repl loop
	scanner := bufio.NewScanner(os.Stdin)
	env := initState()
	for {
		fmt.Print("user> ")
		// reads user input until \n by default
		scanner.Scan()
		// Holds the string that was scanned
		text := scanner.Text()
		repl(text, env)

	}
}

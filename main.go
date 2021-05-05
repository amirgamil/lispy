package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// read
func READ(str string) []Sexp {
	tokens := readStr(str)
	exprs, err := parse(tokens)
	if err != nil {
		log.Fatal("Error parsing")
	}
	return exprs
}

// eval
func EVAL(ast []Sexp, env string) []string {
	return Eval(ast)
}

// print
func PRINT(res []string) {
	for _, result := range res {
		fmt.Println(result)
	}
}

// repl
func repl(str string) {
	PRINT(EVAL(READ(str), ""))
}

func main() {
	// repl loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("user> ")
		// reads user input until \n by default
		scanner.Scan()
		// Holds the string that was scanned
		text := scanner.Text()
		repl(text)

	}
}

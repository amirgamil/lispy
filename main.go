package main

import (
	"bufio"
	"fmt"
	"os"
)

// read
func READ(str string) string {
	return str
}

// eval
func EVAL(ast string, env string) string {
	return ast
}

// print
func PRINT(exp string) string {
	return exp
}

// repl
func repl(str string) string {
	readStr(str)
	return ""
	// return PRINT(EVAL(READ(str), ""))
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
		fmt.Println(repl(text))

	}
}

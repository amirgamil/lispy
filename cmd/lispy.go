package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/amirgamil/lispy/pkg/lispy"
)

// read
func read(str string) []lispy.Sexp {
	tokens := lispy.Read(str)
	exprs, err := lispy.Parse(tokens)
	if err != nil {
		log.Fatal("Error parsing")
	}
	return exprs
}

// eval
func eval(ast []lispy.Sexp, env *lispy.Env) []string {
	return lispy.Eval(ast, env)
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

const cliVersion = "0.1.0"
const helpMessage = `
Welcome to Lispy! Hack away
`

func main() {

	flag.Usage = func() {
		fmt.Printf(helpMessage, cliVersion)
		flag.PrintDefaults()
	}

	isRepl := flag.Bool("repl", false, "Run as an interactive repl")
	flag.Parse()
	args := flag.Args()
	fmt.Println(args)
	//default to repl if no files given
	if *isRepl || len(args) == 0 {
		// repl loop
		scanner := bufio.NewScanner(os.Stdin)
		env := lispy.InitState()
		for {
			fmt.Print("lispy> ")
			// reads user input until \n by default
			scanner.Scan()
			// Holds the string that was scanned
			text := scanner.Text()
			repl(text, env)

		}
	} else {
		// filePath := args[0]
		// file, err := os.Open(filePath)

		// toke := make(chan lispy.Token)
		// nodes := make(chan lispy.Sexp)
	}
}

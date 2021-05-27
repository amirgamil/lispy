package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/amirgamil/lispy/pkg/lispy"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

// read
func read(reader io.Reader) []lispy.Sexp {
	tokens := lispy.Read(reader)
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
func repl(str io.Reader, env *lispy.Env) {
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
	//default to repl if no files given
	if *isRepl || len(args) == 0 {
		// repl loop
		reader := bufio.NewReader(os.Stdin)
		env := lispy.InitState()
		for {
			fmt.Print(Green + "lispy> " + Reset)
			// reads user input until \n by default
			text, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal("Error reading input from the console")
			} else if err == io.EOF {
				break
			}
			repl(strings.NewReader(text), env)

		}
	} else {
		filePath := args[0]
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal("Error opening file to read!")
		}
		defer file.Close()
		env := lispy.InitState()
		repl(file, env)
	}
}
